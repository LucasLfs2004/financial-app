begin;

create extension if not exists pgtap with schema extensions;

select plan(18);

select has_type('public', 'plan_status', 'plan_status enum exists');
select has_type('public', 'snapshot_kind', 'snapshot_kind enum exists');
select has_table('public', 'plans', 'plans table exists');
select has_table('public', 'plan_snapshots', 'plan_snapshots table exists');

insert into auth.users (
  id,
  instance_id,
  aud,
  role,
  email,
  encrypted_password,
  email_confirmed_at,
  raw_app_meta_data,
  raw_user_meta_data,
  created_at,
  updated_at
)
values
  (
    '10000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'release1-owner-1@example.com',
    '',
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now()
  ),
  (
    '20000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'release1-owner-2@example.com',
    '',
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now()
  );

insert into public.plans (
  id,
  user_id,
  name,
  start_month,
  end_month,
  currency_code
)
values
  (
    '11000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000001',
    'Planejamento principal',
    '2026-01-01',
    '2026-12-01',
    'BRL'
  ),
  (
    '22000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000002',
    'Outro planejamento',
    '2026-01-01',
    '2026-12-01',
    'BRL'
  );

select throws_ok(
  $$
    insert into public.plans (user_id, name, start_month, end_month)
    values (
      '10000000-0000-0000-0000-000000000001',
      'Duplicado',
      '2027-01-01',
      '2027-12-01'
    )
  $$,
  '23505',
  null,
  'a user cannot have two non-archived plans'
);

select lives_ok(
  $$
    insert into public.plans (
      user_id,
      name,
      status,
      start_month,
      end_month,
      archived_at
    )
    values (
      '10000000-0000-0000-0000-000000000001',
      'Histórico arquivado',
      'archived',
      '2025-01-01',
      '2025-12-01',
      now()
    )
  $$,
  'archived plans do not violate current-plan uniqueness'
);

select throws_ok(
  $$
    insert into public.plans (user_id, name, start_month, end_month)
    values (
      '10000000-0000-0000-0000-000000000001',
      'Mês não normalizado',
      '2027-01-02',
      '2027-12-01'
    )
  $$,
  '23514',
  null,
  'plan months must be normalized to their first day'
);

select throws_ok(
  $$
    insert into public.plans (user_id, name, start_month, end_month)
    values (
      '10000000-0000-0000-0000-000000000001',
      'Horizonte inválido',
      '2027-12-01',
      '2027-01-01'
    )
  $$,
  '23514',
  null,
  'plan horizon cannot end before it starts'
);

insert into public.plan_snapshots (
  id,
  plan_id,
  user_id,
  kind,
  schema_version,
  document
)
values (
  '11100000-0000-0000-0000-000000000001',
  '11000000-0000-0000-0000-000000000001',
  '10000000-0000-0000-0000-000000000001',
  'original',
  1,
  '{"name":"Planejamento principal"}'
);

select throws_ok(
  $$
    update public.plan_snapshots
    set document = '{"changed":true}'
    where id = '11100000-0000-0000-0000-000000000001'
  $$,
  '55000',
  'plan snapshots are immutable',
  'snapshot payload cannot be updated even by a privileged connection'
);

select throws_ok(
  $$
    insert into public.plan_snapshots (
      plan_id,
      user_id,
      kind,
      schema_version,
      document
    )
    values (
      '22000000-0000-0000-0000-000000000002',
      '10000000-0000-0000-0000-000000000001',
      'original',
      1,
      '{}'
    )
  $$,
  '23503',
  null,
  'snapshot ownership must match plan ownership'
);

select throws_ok(
  $$
    insert into public.plan_snapshots (
      plan_id,
      user_id,
      kind,
      schema_version,
      document
    )
    values (
      '11000000-0000-0000-0000-000000000001',
      '10000000-0000-0000-0000-000000000001',
      'original',
      1,
      '{}'
    )
  $$,
  '23505',
  null,
  'a plan can have only one original snapshot'
);

select throws_ok(
  $$
    insert into public.plan_snapshots (
      plan_id,
      user_id,
      kind,
      schema_version,
      document
    )
    values (
      '22000000-0000-0000-0000-000000000002',
      '20000000-0000-0000-0000-000000000002',
      'original',
      0,
      '{}'
    )
  $$,
  '23514',
  null,
  'snapshot schema version must be positive'
);

insert into public.plan_snapshots (
  id,
  plan_id,
  user_id,
  kind,
  schema_version,
  document
)
values (
  '22200000-0000-0000-0000-000000000002',
  '22000000-0000-0000-0000-000000000002',
  '20000000-0000-0000-0000-000000000002',
  'original',
  1,
  '{"name":"Outro planejamento"}'
);

set local role authenticated;
select set_config(
  'request.jwt.claim.sub',
  '10000000-0000-0000-0000-000000000001',
  true
);

select is(
  (select count(*) from public.plans),
  2::bigint,
  'authenticated users see all of their own plans'
);

select is_empty(
  $$
    select id from public.plans
    where user_id = '20000000-0000-0000-0000-000000000002'
  $$,
  'authenticated users cannot see another user plans'
);

select is(
  (select count(*) from public.plan_snapshots),
  1::bigint,
  'authenticated users see their own snapshots'
);

select is_empty(
  $$
    select id from public.plan_snapshots
    where user_id = '20000000-0000-0000-0000-000000000002'
  $$,
  'authenticated users cannot see another user snapshots'
);

select throws_ok(
  $$
    insert into public.plans (user_id, name, start_month, end_month)
    values (
      '10000000-0000-0000-0000-000000000001',
      'Escrita direta',
      '2028-01-01',
      '2028-12-01'
    )
  $$,
  '42501',
  null,
  'authenticated clients cannot write plans directly'
);

select throws_ok(
  $$
    delete from public.plan_snapshots
    where id = '11100000-0000-0000-0000-000000000001'
  $$,
  '42501',
  null,
  'authenticated clients cannot delete snapshots'
);

reset role;

select * from finish();
rollback;
