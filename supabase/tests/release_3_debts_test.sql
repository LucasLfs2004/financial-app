begin;

create extension if not exists pgtap with schema extensions;
select plan(20);

select has_table('public', 'debts', 'debts table exists');
select has_index('public', 'debts', 'debts_user_scheduled_end_idx', 'debts have a user and end month index');
select has_column('public', 'debts', 'financial_item_id', 'debts share the financial item identity');
select ok(
  'debt_installment' = any(enum_range(null::public.financial_item_kind)::text[]),
  'financial item kind accepts debt installments'
);

insert into auth.users (
  id, instance_id, aud, role, email, encrypted_password, email_confirmed_at,
  raw_app_meta_data, raw_user_meta_data, created_at, updated_at
) values
  ('d1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'debt-owner@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('d2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'debt-other@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());

insert into public.financial_items (
  id, user_id, currency_code, name, kind
) values
  ('d1110000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000001', 'BRL', 'Transplante', 'debt_installment'),
  ('d2220000-0000-0000-0000-000000000002', 'd2000000-0000-0000-0000-000000000002', 'BRL', 'Outra dívida', 'debt_installment'),
  ('d1110000-0000-0000-0000-000000000003', 'd1000000-0000-0000-0000-000000000001', 'BRL', 'Seguro', 'fixed_expense'),
  ('d2220000-0000-0000-0000-000000000004', 'd2000000-0000-0000-0000-000000000002', 'BRL', 'Outro seguro', 'fixed_expense');

insert into public.financial_item_periods (
  id, financial_item_id, user_id, start_month, end_month,
  amount_cents, recurrence
) values
  ('d1111000-0000-0000-0000-000000000001', 'd1110000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000001', '2026-09-01', '2027-04-01', 60000, 'monthly'),
  ('d2222000-0000-0000-0000-000000000002', 'd2220000-0000-0000-0000-000000000002', 'd2000000-0000-0000-0000-000000000002', '2026-01-01', '2026-02-01', 25000, 'monthly');

select lives_ok(
  $$ insert into public.debts (financial_item_id, user_id, original_total_cents, total_installments, first_projected_installment, scheduled_start_month, scheduled_end_month) values ('d1110000-0000-0000-0000-000000000001', 'd1000000-0000-0000-0000-000000000001', 720000, 12, 5, '2026-09-01', '2027-04-01') $$,
  'a debt can begin in the middle of its original installment structure'
);
select lives_ok(
  $$ insert into public.debts (financial_item_id, user_id, total_installments, first_projected_installment, scheduled_start_month, scheduled_end_month) values ('d2220000-0000-0000-0000-000000000002', 'd2000000-0000-0000-0000-000000000002', 2, 1, '2026-01-01', '2026-02-01') $$,
  'another user can keep an independent debt schedule'
);

set constraints all immediate;

select throws_ok(
  $$ insert into public.debts (financial_item_id, user_id, total_installments, first_projected_installment, scheduled_start_month, scheduled_end_month) values ('d1110000-0000-0000-0000-000000000003', 'd1000000-0000-0000-0000-000000000001', 1, 1, '2026-09-01', '2026-09-01') $$,
  '23514', 'debt metadata requires a debt installment item', 'debt metadata requires the specialized item kind'
);
select throws_ok(
  $$ insert into public.debts (financial_item_id, user_id, total_installments, first_projected_installment, scheduled_start_month, scheduled_end_month) values ('d2220000-0000-0000-0000-000000000004', 'd1000000-0000-0000-0000-000000000001', 1, 1, '2026-09-01', '2026-09-01') $$,
  '23503', null, 'debt ownership must match financial item ownership'
);
select throws_ok(
  $$ update public.debts set scheduled_end_month = '2027-05-01' where financial_item_id = 'd1110000-0000-0000-0000-000000000001' $$,
  '23514', null, 'scheduled end must match the installment structure'
);
select throws_ok(
  $$ update public.debts set total_installments = 125, scheduled_end_month = '2036-08-01' where financial_item_id = 'd1110000-0000-0000-0000-000000000001' $$,
  '23514', null, 'remaining debt schedule cannot exceed 120 months'
);
select throws_ok(
  $$ update public.debts set original_total_cents = -1 where financial_item_id = 'd1110000-0000-0000-0000-000000000001' $$,
  '23514', null, 'original total cannot be negative'
);
select throws_ok(
  $$ update public.financial_item_periods set recurrence = 'once', end_month = start_month where id = 'd1111000-0000-0000-0000-000000000001' $$,
  '23514', 'debt periods must be positive monthly intervals inside the schedule', 'debt periods must be monthly'
);
select throws_ok(
  $$ update public.financial_item_periods set start_month = '2026-10-01' where id = 'd1111000-0000-0000-0000-000000000001' $$,
  '23514', 'debt periods must continuously cover the schedule', 'debt periods cannot leave a schedule gap'
);
select throws_ok(
  $$ update public.financial_item_periods set end_month = null where id = 'd1111000-0000-0000-0000-000000000001' $$,
  '23514', 'debt periods must be positive monthly intervals inside the schedule', 'debt periods must have a bounded end'
);
select throws_ok(
  $$ update public.financial_items set kind = 'fixed_expense' where id = 'd1110000-0000-0000-0000-000000000001' $$,
  '23514', 'debt metadata requires a debt installment item', 'a debt item cannot lose its specialized kind'
);
select throws_ok(
  $$ delete from public.debts where financial_item_id = 'd1110000-0000-0000-0000-000000000001' $$,
  '23514', 'debt installment items require debt metadata', 'a debt item cannot lose its metadata'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'd1000000-0000-0000-0000-000000000001', true);

select is((select count(*) from public.debts), 1::bigint, 'authenticated users see only their debts');
select is_empty(
  $$ select financial_item_id from public.debts where user_id = 'd2000000-0000-0000-0000-000000000002' $$,
  'RLS hides another user debts'
);
select throws_ok(
  $$ update public.debts set original_total_cents = 1 where financial_item_id = 'd1110000-0000-0000-0000-000000000001' $$,
  '42501', null, 'authenticated clients cannot update debts directly'
);
select throws_ok(
  $$ insert into public.debts (financial_item_id, user_id, total_installments, first_projected_installment, scheduled_start_month, scheduled_end_month) values ('d1110000-0000-0000-0000-000000000003', 'd1000000-0000-0000-0000-000000000001', 1, 1, '2026-09-01', '2026-09-01') $$,
  '42501', null, 'authenticated clients cannot insert debts directly'
);

reset role;

select * from finish();
rollback;
