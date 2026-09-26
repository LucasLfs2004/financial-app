begin;

create extension if not exists pgtap with schema extensions;
select plan(16);

select has_table('public', 'financial_item_payment_periods', 'payment periods table exists');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_item_start_idx', 'payment periods have an item lookup index');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_user_interval_idx', 'payment periods have a user interval index');
select has_index('public', 'financial_item_payment_periods', 'financial_item_payment_periods_card_interval_idx', 'payment periods have a card interval index');
select hasnt_column('public', 'financial_item_payment_periods', 'plan_id', 'payment periods are not owned by plans');

insert into auth.users (id, instance_id, aud, role, email, encrypted_password, email_confirmed_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at) values
  ('b1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'payment-owner@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('b2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'payment-other@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());

insert into public.financial_institutions (id, user_id, name) values
  ('b1110000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Banco'),
  ('b2220000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'Outro banco');
insert into public.credit_cards (id, user_id, institution_id, name) values
  ('b1111000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'b1110000-0000-0000-0000-000000000001', 'Principal'),
  ('b2222000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'b2220000-0000-0000-0000-000000000002', 'Outro');
insert into public.financial_items (id, user_id, currency_code, name, kind) values
  ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'BRL', 'Seguro', 'fixed_expense'),
  ('b1111200-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'BRL', 'Gasolina', 'projected_variable_expense'),
  ('b1111300-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'BRL', 'Salário', 'recurring_income');

select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-03-01', 'direct') $$,
  'direct payment period is accepted'
);
select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111200-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-06-01', 'credit_card', 'b1111000-0000-0000-0000-000000000001') $$,
  'credit card payment period is accepted'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2026-04-01', 'direct', 'b1111000-0000-0000-0000-000000000001') $$,
  '23514', null, 'direct payment cannot reference a card'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2026-04-01', 'credit_card') $$,
  '23514', null, 'credit card payment requires a card'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111300-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-02-01', 'direct') $$,
  '23514', null, 'income cannot have a payment method period'
);
select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2026-04-01', '2027-09-01', 'direct') $$,
  'payment configuration can extend beyond a plan horizon'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2027-01-01', '2027-10-01', 'direct') $$,
  '23P01', null, 'payment periods cannot overlap'
);
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method, credit_card_id) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2028-01-01', '2028-01-01', 'credit_card', 'b2222000-0000-0000-0000-000000000002') $$,
  '23503', null, 'payment card must belong to the item owner'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'b1000000-0000-0000-0000-000000000001', true);
select is((select count(*) from public.financial_item_payment_periods), 3::bigint, 'users see their own payment periods');
select is_empty($$ select id from public.financial_item_payment_periods where user_id = 'b2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user payment periods');
select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('b1111100-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '2027-10-01', '2027-10-01', 'direct') $$,
  '42501', null, 'authenticated clients cannot write payment periods directly'
);
reset role;

select * from finish();
rollback;
