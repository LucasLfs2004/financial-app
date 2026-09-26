begin;

create extension if not exists pgtap with schema extensions;
select plan(18);

select has_table('public', 'debt_early_settlements', 'early settlements table exists');
select has_index(
  'public',
  'debt_early_settlements',
  'debt_early_settlements_user_reference_idx',
  'early settlements have a user and reference month index'
);
select has_column('public', 'debt_early_settlements', 'recorded_by', 'early settlements record their author');

insert into auth.users (
  id, instance_id, aud, role, email, encrypted_password, email_confirmed_at,
  raw_app_meta_data, raw_user_meta_data, created_at, updated_at
) values
  ('e1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'settlement-owner@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('e2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'settlement-other@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());

insert into public.financial_items (
  id, user_id, currency_code, name, kind
) values
  ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', 'BRL', 'Transplante', 'debt_installment'),
  ('e2220000-0000-0000-0000-000000000002', 'e2000000-0000-0000-0000-000000000002', 'BRL', 'Outra dívida', 'debt_installment'),
  ('e2220000-0000-0000-0000-000000000003', 'e2000000-0000-0000-0000-000000000002', 'BRL', 'Dívida sem quitação', 'debt_installment');

insert into public.financial_item_periods (
  id, financial_item_id, user_id, start_month, end_month,
  amount_cents, recurrence
) values
  ('e1111000-0000-0000-0000-000000000001', 'e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-09-01', '2027-04-01', 60000, 'monthly'),
  ('e2222000-0000-0000-0000-000000000002', 'e2220000-0000-0000-0000-000000000002', 'e2000000-0000-0000-0000-000000000002', '2026-01-01', '2026-03-01', 25000, 'monthly'),
  ('e2222000-0000-0000-0000-000000000003', 'e2220000-0000-0000-0000-000000000003', 'e2000000-0000-0000-0000-000000000002', '2026-04-01', '2026-05-01', 30000, 'monthly');

insert into public.debts (
  financial_item_id, user_id, total_installments,
  first_projected_installment, scheduled_start_month, scheduled_end_month
) values
  ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', 12, 5, '2026-09-01', '2027-04-01'),
  ('e2220000-0000-0000-0000-000000000002', 'e2000000-0000-0000-0000-000000000002', 3, 1, '2026-01-01', '2026-03-01'),
  ('e2220000-0000-0000-0000-000000000003', 'e2000000-0000-0000-0000-000000000002', 2, 1, '2026-04-01', '2026-05-01');

set constraints all immediate;

select lives_ok(
  $$ insert into public.debt_early_settlements (id, financial_item_id, user_id, reference_month, amount_cents, reason, recorded_by) values ('e1111100-0000-0000-0000-000000000001', 'e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-12-01', 210000, 'Desconto para quitação', 'e1000000-0000-0000-0000-000000000001') $$,
  'an early settlement inside the schedule is accepted'
);
select lives_ok(
  $$ insert into public.debt_early_settlements (id, financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e2222200-0000-0000-0000-000000000002', 'e2220000-0000-0000-0000-000000000002', 'e2000000-0000-0000-0000-000000000002', '2026-01-01', 40000, 'e2000000-0000-0000-0000-000000000002') $$,
  'the first projected installment can be replaced by settlement'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2027-01-01', 180000, 'e1000000-0000-0000-0000-000000000001') $$,
  '23505', null, 'the unique constraint serializes competing settlements for one debt'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e2220000-0000-0000-0000-000000000003', 'e1000000-0000-0000-0000-000000000001', '2026-04-01', 40000, 'e1000000-0000-0000-0000-000000000001') $$,
  '23503', null, 'settlement ownership must match debt ownership'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-08-01', 210000, 'e1000000-0000-0000-0000-000000000001') $$,
  '23514', 'early settlement month must be inside the debt schedule and before its end', 'settlement cannot precede the projected schedule'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e2220000-0000-0000-0000-000000000002', 'e2000000-0000-0000-0000-000000000002', '2026-03-01', 25000, 'e2000000-0000-0000-0000-000000000002') $$,
  '23514', 'early settlement month must be inside the debt schedule and before its end', 'settlement cannot replace the natural final installment'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-12-02', 210000, 'e1000000-0000-0000-0000-000000000001') $$,
  '23514', null, 'settlement month must be normalized'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-12-01', 0, 'e1000000-0000-0000-0000-000000000001') $$,
  '23514', null, 'settlement amount must be positive'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, reason, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-12-01', 1, repeat('x', 501), 'e1000000-0000-0000-0000-000000000001') $$,
  '23514', null, 'settlement reason is limited to 500 characters'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2026-12-01', 1, 'e2000000-0000-0000-0000-000000000002') $$,
  '23514', null, 'settlement author must be the debt owner'
);
select throws_ok(
  $$ update public.debt_early_settlements set reason = 'Alterado' where id = 'e1111100-0000-0000-0000-000000000001' $$,
  '55000', 'debt early settlements are append-only', 'settlements cannot be updated'
);
select throws_ok(
  $$ delete from public.debt_early_settlements where id = 'e1111100-0000-0000-0000-000000000001' $$,
  '55000', 'debt early settlements are append-only', 'settlements cannot be deleted'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'e1000000-0000-0000-0000-000000000001', true);

select is((select count(*) from public.debt_early_settlements), 1::bigint, 'authenticated users see only their settlements');
select is_empty(
  $$ select id from public.debt_early_settlements where user_id = 'e2000000-0000-0000-0000-000000000002' $$,
  'RLS hides another user settlements'
);
select throws_ok(
  $$ insert into public.debt_early_settlements (financial_item_id, user_id, reference_month, amount_cents, recorded_by) values ('e1110000-0000-0000-0000-000000000001', 'e1000000-0000-0000-0000-000000000001', '2027-01-01', 180000, 'e1000000-0000-0000-0000-000000000001') $$,
  '42501', null, 'authenticated clients cannot append settlements directly'
);

reset role;

select * from finish();
rollback;
