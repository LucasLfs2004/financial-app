begin;

create extension if not exists pgtap with schema extensions;

select plan(22);

select has_table('public', 'saving_periods', 'saving_periods table exists');
select has_index(
  'public',
  'saving_periods',
  'saving_periods_plan_start_idx',
  'saving periods have a plan and start-month index'
);
select has_index(
  'public',
  'saving_periods',
  'saving_periods_user_id_idx',
  'saving periods have an ownership index'
);

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
    '50000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'savings-owner-1@example.com',
    '',
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now()
  ),
  (
    '60000000-0000-0000-0000-000000000006',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'savings-owner-2@example.com',
    '',
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now()
  );

insert into public.plans (id, user_id, name, start_month, end_month)
values
  (
    '55000000-0000-0000-0000-000000000005',
    '50000000-0000-0000-0000-000000000005',
    'Plano com economia explícita',
    '2026-01-01',
    '2027-12-01'
  ),
  (
    '66000000-0000-0000-0000-000000000006',
    '60000000-0000-0000-0000-000000000006',
    'Plano sem economia',
    '2026-01-01',
    '2027-12-01'
  );

insert into public.saving_periods (
  id,
  plan_id,
  user_id,
  start_month,
  end_month,
  amount_cents,
  context
)
values (
  '55500000-0000-0000-0000-000000000005',
  '55000000-0000-0000-0000-000000000005',
  '50000000-0000-0000-0000-000000000005',
  '2026-01-01',
  '2026-03-01',
  0,
  'Pausa temporária da reserva'
);

select is(
  (
    select amount_cents
    from public.saving_periods
    where id = '55500000-0000-0000-0000-000000000005'
  ),
  0::bigint,
  'zero is stored as an explicit saving configuration'
);

select is(
  (
    select count(*)
    from public.saving_periods
    where plan_id = '66000000-0000-0000-0000-000000000006'
  ),
  0::bigint,
  'a plan without rows has no saving configuration'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '60000000-0000-0000-0000-000000000006',
      '2026-04-01', '2026-04-01', 100000
    )
  $$,
  '23503',
  null,
  'saving period ownership must match plan ownership'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-04-02', '2026-04-01', 100000
    )
  $$,
  '23514',
  null,
  'saving periods require normalized months'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-04-01', '2026-04-02', 100000
    )
  $$,
  '23514',
  null,
  'saving period end month must also be normalized'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-05-01', '2026-04-01', 100000
    )
  $$,
  '23514',
  null,
  'saving periods cannot end before they start'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-04-01', '2026-04-01', -1
    )
  $$,
  '23514',
  null,
  'saving amount cannot be negative'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents, context
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-04-01', '2026-04-01', 100000, repeat('a', 501)
    )
  $$,
  '23514',
  null,
  'saving context cannot exceed five hundred characters'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-03-01', '2026-04-01', 100000
    )
  $$,
  '23P01',
  null,
  'saving periods cannot share an inclusive boundary month'
);

select lives_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-04-01', '2026-06-01', 120000
    )
  $$,
  'saving period can begin after the previous inclusive end'
);

select lives_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2026-07-01', null, 150000
    )
  $$,
  'open saving period can follow a closed period'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2027-01-01', '2027-01-01', 200000
    )
  $$,
  '23P01',
  null,
  'open saving period conflicts with every later period'
);

select lives_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '66000000-0000-0000-0000-000000000006',
      '60000000-0000-0000-0000-000000000006',
      '2026-01-01', null, 50000
    )
  $$,
  'saving periods from different plans may cover the same months'
);

select is(
  (
    select amount_cents
    from public.saving_periods
    where plan_id = '66000000-0000-0000-0000-000000000006'
  ),
  50000::bigint,
  'a positive saving amount is retained independently per plan'
);

set local role authenticated;
select set_config(
  'request.jwt.claim.sub',
  '50000000-0000-0000-0000-000000000005',
  true
);

select is(
  (select count(*) from public.saving_periods),
  3::bigint,
  'authenticated users see only their own saving periods'
);

select is_empty(
  $$
    select id from public.saving_periods
    where user_id = '60000000-0000-0000-0000-000000000006'
  $$,
  'RLS hides another user saving periods'
);

select is_empty(
  $$
    select id from public.saving_periods
    where plan_id = '66000000-0000-0000-0000-000000000006'
  $$,
  'RLS hides another user plan saving periods'
);

select throws_ok(
  $$
    insert into public.saving_periods (
      plan_id, user_id, start_month, end_month, amount_cents
    ) values (
      '55000000-0000-0000-0000-000000000005',
      '50000000-0000-0000-0000-000000000005',
      '2028-01-01', '2028-12-01', 100000
    )
  $$,
  '42501',
  null,
  'authenticated clients cannot write saving periods directly'
);

select throws_ok(
  $$
    delete from public.saving_periods
    where id = '55500000-0000-0000-0000-000000000005'
  $$,
  '42501',
  null,
  'authenticated clients cannot delete saving periods directly'
);

reset role;

select * from finish();
rollback;
