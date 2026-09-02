begin;

create extension if not exists pgtap with schema extensions;
select plan(18);

select has_type('public', 'payment_method_kind', 'payment method enum exists');
select has_table('public', 'financial_item_payment_periods', 'payment periods table exists');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_item_start_idx', 'payment periods have an item lookup index');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_plan_interval_idx', 'payment periods have a plan interval index');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_card_interval_idx', 'payment periods have a card interval index');

insert into auth.users (id, instance_id, aud, role, email, encrypted_password, email_confirmed_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at)
values
  ('b1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'payments-owner-1@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('b2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'payments-owner-2@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());
insert into public.plans (id, user_id, name, start_month, end_month) values
  ('b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Plano A', '2026-01-01', '2026-12-01'),
  ('b2200000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'Plano B', '2026-01-01', '2026-12-01');
insert into public.financial_institutions (id, user_id, name) values
  ('b1110000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Banco A'),
  ('b2220000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'Banco B');
insert into public.credit_cards (id, user_id, institution_id, name) values
  ('b1111000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'b1110000-0000-0000-0000-000000000001', 'Cartão A'),
  ('b2222000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'b2220000-0000-0000-0000-000000000002', 'Cartão B');
insert into public.financial_items (id, plan_id, user_id, name, kind) values
  ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Aluguel', 'fixed_expense'),
  ('b1111200-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Mercado', 'projected_variable_expense'),
  ('b1111300-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Salário', 'recurring_income'),
  ('b2222100-0000-0000-0000-000000000002', 'b2200000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'Despesa B', 'fixed_expense');

select is_empty(
  $$ select id from public.financial_item_payment_periods where financial_item_id = 'b1111200-0000-0000-0000-000000000001' $$,
  'absence of an explicit payment period represents the direct fallback'
);
select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-03-01', 'direct') $$,
  'an expense can have a direct payment period'
);
select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111200-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-06-01', 'credit_card', 'b1111000-0000-0000-0000-000000000001') $$,
  'an expense can have a credit card payment period'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2026-04-01', 'direct', 'b1111000-0000-0000-0000-000000000001') $$,
  '23514', null, 'direct payment cannot reference a card'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2026-04-01', 'credit_card') $$,
  '23514', null, 'credit card payment requires a card'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111300-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-02-01', 'direct') $$,
  '23514', 'payment periods can only be linked to expense items', 'income items cannot have payment periods'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2025-12-01', '2025-12-01', 'direct') $$,
  '23514', 'payment period must be contained in the plan horizon', 'payment periods cannot start before the plan'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-10-01', '2027-01-01', 'direct') $$,
  '23514', 'payment period must be contained in the plan horizon', 'payment periods cannot end after the plan'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-03-01', '2026-05-01', 'direct') $$,
  '23P01', null, 'payment periods cannot overlap inclusively'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2026-04-01', 'credit_card', 'b2222000-0000-0000-0000-000000000002') $$,
  '23503', null, 'payment periods cannot reference another user card'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'b1000000-0000-0000-0000-000000000001', true);
select is((select count(*) from public.financial_item_payment_periods), 2::bigint, 'users see their own payment periods');
select is_empty($$ select id from public.financial_item_payment_periods where user_id = 'b2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user payment periods');
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, plan_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1100000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-07-01', '2026-07-01', 'direct') $$,
  '42501', null, 'authenticated clients cannot write payment periods directly'
);

select * from finish();
rollback;
