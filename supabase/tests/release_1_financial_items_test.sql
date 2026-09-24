begin;

create extension if not exists pgtap with schema extensions;

select plan(27);

select has_type('public', 'financial_item_kind', 'financial_item_kind enum exists');
select has_type('public', 'financial_item_status', 'financial_item_status enum exists');
select has_type('public', 'recurrence_kind', 'recurrence_kind enum exists');
select has_table('public', 'financial_items', 'financial_items table exists');
select has_table('public', 'financial_item_periods', 'financial_item_periods table exists');
select has_index('public', 'financial_items', 'financial_items_user_currency_status_idx', 'items have a user, currency and status index');
select has_index('public', 'financial_item_periods', 'financial_item_periods_user_interval_idx', 'periods have a user interval index');
select has_column('public', 'financial_items', 'currency_code', 'items store their currency');
select hasnt_column('public', 'financial_items', 'plan_id', 'items are not owned by plans');
select hasnt_column('public', 'financial_item_periods', 'plan_id', 'periods are not owned by plans');

insert into auth.users (
  id, instance_id, aud, role, email, encrypted_password, email_confirmed_at,
  raw_app_meta_data, raw_user_meta_data, created_at, updated_at
) values
  ('30000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'items-owner-1@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('40000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'items-owner-2@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());

insert into public.plans (id, user_id, name, start_month, end_month) values
  ('33000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', 'Plano curto', '2026-09-01', '2026-12-01');

insert into public.financial_items (id, user_id, currency_code, name, kind) values
  ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', 'BRL', 'Salário', 'recurring_income'),
  ('33200000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', 'BRL', 'Seguro', 'fixed_expense'),
  ('44100000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000004', 'BRL', 'Outro salário', 'recurring_income');

select throws_ok(
  $$ insert into public.financial_items (user_id, currency_code, name, kind) values ('30000000-0000-0000-0000-000000000003', 'real', 'Moeda inválida', 'fixed_expense') $$,
  '23514', null, 'item currency must use three uppercase letters'
);

select throws_ok(
  $$ insert into public.financial_items (user_id, currency_code, name, kind, status) values ('30000000-0000-0000-0000-000000000003', 'BRL', 'Arquivado inconsistente', 'fixed_expense', 'archived') $$,
  '23514', null, 'archived financial items require archived_at'
);

insert into public.financial_item_periods (
  id, financial_item_id, user_id, start_month, end_month,
  amount_cents, recurrence, cash_month_offset
) values (
  '33110000-0000-0000-0000-000000000003',
  '33100000-0000-0000-0000-000000000003',
  '30000000-0000-0000-0000-000000000003',
  '2026-01-01', '2026-03-01', 600000, 'monthly', 1
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000004', '2026-04-01', '2026-04-01', 600000, 'monthly') $$,
  '23503', null, 'financial period ownership must match item ownership'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-04-02', '2026-04-01', 600000, 'monthly') $$,
  '23514', null, 'period months must be normalized'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-05-01', '2026-04-01', 600000, 'monthly') $$,
  '23514', null, 'period cannot end before it starts'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-04-01', '2026-04-01', -1, 'monthly') $$,
  '23514', null, 'period amount cannot be negative'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence, cash_month_offset) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-04-01', '2026-04-01', 600000, 'monthly', 13) $$,
  '23514', null, 'cash month offset cannot exceed twelve months'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-04-01', null, 600000, 'once') $$,
  '23514', null, 'one-time period cannot remain open'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-03-01', '2026-04-01', 650000, 'monthly') $$,
  '23P01', null, 'inclusive periods cannot share their boundary month'
);

select lives_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33100000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-04-01', '2026-06-01', 650000, 'monthly') $$,
  'a period can begin after the previous inclusive end'
);

select lives_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33200000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2026-09-01', '2027-09-01', 18000, 'monthly') $$,
  'a financial period can extend beyond the current plan horizon'
);

select lives_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('44100000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000004', '2026-01-01', null, 500000, 'monthly') $$,
  'different users can keep independent periods'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', '30000000-0000-0000-0000-000000000003', true);

select is((select count(*) from public.financial_items), 2::bigint, 'authenticated users see only their financial items');
select is((select count(*) from public.financial_item_periods), 3::bigint, 'authenticated users see only their financial periods');
select is_empty($$ select id from public.financial_items where user_id = '40000000-0000-0000-0000-000000000004' $$, 'RLS hides another user financial items');

select throws_ok(
  $$ insert into public.financial_items (user_id, currency_code, name, kind) values ('30000000-0000-0000-0000-000000000003', 'BRL', 'Escrita direta', 'fixed_expense') $$,
  '42501', null, 'authenticated clients cannot write financial items directly'
);

select throws_ok(
  $$ insert into public.financial_item_periods (financial_item_id, user_id, start_month, end_month, amount_cents, recurrence) values ('33200000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '2027-10-01', '2027-10-01', 150000, 'monthly') $$,
  '42501', null, 'authenticated clients cannot write financial periods directly'
);

reset role;

select * from finish();
rollback;
