begin;

create extension if not exists pgtap with schema extensions;

select plan(30);

select has_type('public', 'financial_item_kind', 'financial_item_kind enum exists');
select has_type('public', 'financial_item_status', 'financial_item_status enum exists');
select has_type('public', 'recurrence_kind', 'recurrence_kind enum exists');
select has_table('public', 'financial_items', 'financial_items table exists');
select has_table('public', 'financial_item_periods', 'financial_item_periods table exists');
select has_index(
  'public',
  'financial_items',
  'financial_items_user_id_idx',
  'financial items have an ownership index'
);
select has_index(
  'public',
  'financial_items',
  'financial_items_plan_status_idx',
  'financial items have a plan and status index'
);
select has_index(
  'public',
  'financial_item_periods',
  'financial_item_periods_item_start_idx',
  'financial periods have an item and start index'
);
select has_index(
  'public',
  'financial_item_periods',
  'financial_item_periods_plan_interval_idx',
  'financial periods have a plan interval index'
);
select has_index(
  'public',
  'financial_item_periods',
  'financial_item_periods_user_id_idx',
  'financial periods have an ownership index'
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
    '30000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'items-owner-1@example.com',
    '',
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now()
  ),
  (
    '40000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000000',
    'authenticated',
    'authenticated',
    'items-owner-2@example.com',
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
    '33000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'Plano de itens 1',
    '2026-01-01',
    '2027-12-01'
  ),
  (
    '44000000-0000-0000-0000-000000000004',
    '40000000-0000-0000-0000-000000000004',
    'Plano de itens 2',
    '2026-01-01',
    '2027-12-01'
  );

insert into public.financial_items (
  id,
  plan_id,
  user_id,
  name,
  kind
)
values
  (
    '33100000-0000-0000-0000-000000000003',
    '33000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'Salário',
    'recurring_income'
  ),
  (
    '33200000-0000-0000-0000-000000000003',
    '33000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'Aluguel',
    'fixed_expense'
  ),
  (
    '44100000-0000-0000-0000-000000000004',
    '44000000-0000-0000-0000-000000000004',
    '40000000-0000-0000-0000-000000000004',
    'Outro salário',
    'recurring_income'
  );

select throws_ok(
  $$
    insert into public.financial_items (plan_id, user_id, name, kind)
    values (
      '33000000-0000-0000-0000-000000000003',
      '40000000-0000-0000-0000-000000000004',
      'Item com dono incorreto',
      'fixed_expense'
    )
  $$,
  '23503',
  null,
  'financial item ownership must match plan ownership'
);

select throws_ok(
  $$
    insert into public.financial_items (
      plan_id,
      user_id,
      name,
      kind,
      status
    )
    values (
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      'Arquivado inconsistente',
      'fixed_expense',
      'archived'
    )
  $$,
  '23514',
  null,
  'archived financial items require archived_at'
);

insert into public.financial_item_periods (
  id,
  financial_item_id,
  plan_id,
  user_id,
  start_month,
  end_month,
  amount_cents,
  recurrence,
  cash_month_offset
)
values (
  '33110000-0000-0000-0000-000000000003',
  '33100000-0000-0000-0000-000000000003',
  '33000000-0000-0000-0000-000000000003',
  '30000000-0000-0000-0000-000000000003',
  '2026-01-01',
  '2026-03-01',
  600000,
  'monthly',
  1
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id,
      plan_id,
      user_id,
      start_month,
      end_month,
      amount_cents,
      recurrence
    )
    values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '40000000-0000-0000-0000-000000000004',
      '2026-04-01',
      '2026-04-01',
      600000,
      'monthly'
    )
  $$,
  '23503',
  null,
  'financial period ownership must match item ownership'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-02', '2026-04-01', 600000, 'monthly'
    )
  $$,
  '23514',
  null,
  'period months must be normalized'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-05-01', '2026-04-01', 600000, 'monthly'
    )
  $$,
  '23514',
  null,
  'period cannot end before it starts'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-01', '2026-04-01', -1, 'monthly'
    )
  $$,
  '23514',
  null,
  'period amount cannot be negative'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence, cash_month_offset
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-01', '2026-04-01', 600000, 'monthly', 13
    )
  $$,
  '23514',
  null,
  'cash month offset cannot exceed twelve months'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-01', null, 600000, 'once'
    )
  $$,
  '23514',
  null,
  'one-time period cannot remain open'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-01', '2026-05-01', 600000, 'once'
    )
  $$,
  '23514',
  null,
  'one-time period must start and end in the same month'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-03-01', '2026-04-01', 650000, 'monthly'
    )
  $$,
  '23P01',
  null,
  'inclusive periods cannot share their boundary month'
);

select lives_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-04-01', '2026-06-01', 650000, 'monthly'
    )
  $$,
  'a period can begin in the month after the previous inclusive end'
);

select lives_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-07-01', null, 700000, 'monthly'
    )
  $$,
  'an open period can follow a closed period'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33100000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2027-01-01', '2027-01-01', 750000, 'monthly'
    )
  $$,
  '23P01',
  null,
  'an open period conflicts with every later period of the same item'
);

select lives_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33200000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2026-01-01', null, 150000, 'monthly'
    )
  $$,
  'different items may have periods covering the same months'
);

insert into public.financial_item_periods (
  financial_item_id, plan_id, user_id, start_month, end_month,
  amount_cents, recurrence
) values (
  '44100000-0000-0000-0000-000000000004',
  '44000000-0000-0000-0000-000000000004',
  '40000000-0000-0000-0000-000000000004',
  '2026-01-01', null, 500000, 'monthly'
);

set local role authenticated;
select set_config(
  'request.jwt.claim.sub',
  '30000000-0000-0000-0000-000000000003',
  true
);

select is(
  (select count(*) from public.financial_items),
  2::bigint,
  'authenticated users see only their financial items'
);

select is_empty(
  $$
    select id from public.financial_items
    where user_id = '40000000-0000-0000-0000-000000000004'
  $$,
  'RLS hides another user financial items'
);

select is(
  (select count(*) from public.financial_item_periods),
  4::bigint,
  'authenticated users see only their financial periods'
);

select is_empty(
  $$
    select id from public.financial_item_periods
    where user_id = '40000000-0000-0000-0000-000000000004'
  $$,
  'RLS hides another user financial periods'
);

select throws_ok(
  $$
    insert into public.financial_items (plan_id, user_id, name, kind)
    values (
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      'Escrita direta',
      'fixed_expense'
    )
  $$,
  '42501',
  null,
  'authenticated clients cannot write financial items directly'
);

select throws_ok(
  $$
    insert into public.financial_item_periods (
      financial_item_id, plan_id, user_id, start_month, end_month,
      amount_cents, recurrence
    ) values (
      '33200000-0000-0000-0000-000000000003',
      '33000000-0000-0000-0000-000000000003',
      '30000000-0000-0000-0000-000000000003',
      '2027-01-01', '2027-01-01', 150000, 'monthly'
    )
  $$,
  '42501',
  null,
  'authenticated clients cannot write financial periods directly'
);

reset role;

select * from finish();
rollback;
