begin;

create extension if not exists pgtap with schema extensions;
select plan(3);

insert into auth.users (
  id, instance_id, aud, role, email, encrypted_password, email_confirmed_at,
  raw_app_meta_data, raw_user_meta_data, created_at, updated_at
) values (
  'f1000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000000',
  'authenticated', 'authenticated', 'debt-payment@example.com', '', now(),
  '{"provider":"email","providers":["email"]}', '{}', now(), now()
);

insert into public.financial_institutions (id, user_id, name) values
  ('f1110000-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', 'Banco');
insert into public.credit_cards (id, user_id, institution_id, name) values
  ('f1111000-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', 'f1110000-0000-0000-0000-000000000001', 'Principal');
insert into public.financial_items (id, user_id, currency_code, name, kind) values
  ('f1111100-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', 'BRL', 'Dívida', 'debt_installment'),
  ('f1111300-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', 'BRL', 'Salário', 'recurring_income');
insert into public.financial_item_periods (
  financial_item_id, user_id, start_month, end_month,
  amount_cents, recurrence
) values (
  'f1111100-0000-0000-0000-000000000001',
  'f1000000-0000-0000-0000-000000000001',
  '2026-09-01', '2027-04-01', 60000, 'monthly'
);
insert into public.debts (
  financial_item_id, user_id, total_installments,
  first_projected_installment, scheduled_start_month, scheduled_end_month
) values (
  'f1111100-0000-0000-0000-000000000001',
  'f1000000-0000-0000-0000-000000000001',
  12, 5, '2026-09-01', '2027-04-01'
);

set constraints all immediate;

select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('f1111100-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', '2026-09-01', '2026-12-01', 'direct') $$,
  'debt accepts direct payment periods'
);

select lives_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method, credit_card_id) values ('f1111100-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', '2027-01-01', '2027-04-01', 'credit_card', 'f1111000-0000-0000-0000-000000000001') $$,
  'debt accepts credit card periods by effective month'
);

select throws_ok(
  $$ insert into public.financial_item_payment_periods (financial_item_id, user_id, start_month, end_month, method) values ('f1111300-0000-0000-0000-000000000001', 'f1000000-0000-0000-0000-000000000001', '2026-03-01', '2026-03-01', 'direct') $$,
  '23514', null, 'income continues rejecting payment periods'
);

select * from finish();
rollback;
