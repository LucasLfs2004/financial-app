begin;

create extension if not exists pgtap with schema extensions;
select plan(23);

select has_table('public', 'card_invoice_adjustments', 'invoice adjustments table exists');
select has_table('public', 'card_invoice_audit_events', 'invoice audit events table exists');
select has_index('public', 'card_invoice_adjustments', 'card_invoice_adjustments_card_payment_idx', 'adjustments have an invoice lookup index');
select has_index('public', 'card_invoice_adjustments', 'card_invoice_adjustments_user_reference_idx', 'adjustments have a user reference index');
select has_index('public', 'card_invoice_audit_events', 'card_invoice_audit_events_invoice_idx', 'events have an invoice lookup index');
select has_index('public', 'card_invoice_audit_events', 'card_invoice_audit_events_reference_idx', 'events have a reference lookup index');
select has_index('public', 'card_invoice_audit_events', 'card_invoice_audit_events_latest_move_idx', 'events have a latest movement index');
select has_column('public', 'card_invoice_adjustments', 'currency_code', 'adjustments store their currency');
select hasnt_column('public', 'card_invoice_adjustments', 'plan_id', 'adjustments are not owned by plans');
select hasnt_column('public', 'card_invoice_audit_events', 'plan_id', 'audit events are not owned by plans');

insert into auth.users (id, instance_id, aud, role, email, encrypted_password, email_confirmed_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at) values
  ('c1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'invoice-owner@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('c2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'invoice-other@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());
insert into public.plans (id, user_id, name, start_month, end_month) values
  ('c1100000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'Plano curto', '2026-01-01', '2026-12-01');
insert into public.financial_institutions (id, user_id, name) values
  ('c1110000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'Banco'),
  ('c2220000-0000-0000-0000-000000000002', 'c2000000-0000-0000-0000-000000000002', 'Outro banco');
insert into public.credit_cards (id, user_id, institution_id, name) values
  ('c1111000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'c1110000-0000-0000-0000-000000000001', 'Principal'),
  ('c2222000-0000-0000-0000-000000000002', 'c2000000-0000-0000-0000-000000000002', 'c2220000-0000-0000-0000-000000000002', 'Outro');
insert into public.financial_items (id, user_id, currency_code, name, kind) values
  ('c1111100-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'BRL', 'Seguro', 'fixed_expense');

select lives_ok(
  $$ insert into public.card_invoice_adjustments (id, user_id, currency_code, credit_card_id, payment_month, reference_month, name, amount_cents) values ('c1111110-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'BRL', 'c1111000-0000-0000-0000-000000000001', '2028-01-01', '2027-12-01', 'Ajuste futuro', 2500) $$,
  'adjustment can exist beyond the current plan horizon'
);
select throws_ok(
  $$ insert into public.card_invoice_adjustments (user_id, currency_code, credit_card_id, payment_month, name, amount_cents) values ('c1000000-0000-0000-0000-000000000001', 'BRL', 'c1111000-0000-0000-0000-000000000001', '2026-05-01', 'Negativo', -1) $$,
  '23514', null, 'adjustment amount cannot be negative'
);
select throws_ok(
  $$ insert into public.card_invoice_adjustments (user_id, currency_code, credit_card_id, payment_month, name, amount_cents) values ('c1000000-0000-0000-0000-000000000001', 'BRL', 'c1111000-0000-0000-0000-000000000001', '2026-05-02', 'Mês inválido', 1) $$,
  '23514', null, 'adjustment payment month must be normalized'
);
select throws_ok(
  $$ insert into public.card_invoice_adjustments (user_id, currency_code, credit_card_id, payment_month, name, amount_cents) values ('c1000000-0000-0000-0000-000000000001', 'BRL', 'c2222000-0000-0000-0000-000000000002', '2026-05-01', 'Outro cartão', 1) $$,
  '23503', null, 'adjustment card must belong to the user'
);
select lives_ok(
  $$ insert into public.card_invoice_audit_events (id, user_id, event_type, financial_item_id, reference_month, from_credit_card_id, from_payment_month, to_credit_card_id, to_payment_month, reason) values ('c1111111-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'occurrence_moved', 'c1111100-0000-0000-0000-000000000001', '2026-05-01', 'c1111000-0000-0000-0000-000000000001', '2026-06-01', 'c1111000-0000-0000-0000-000000000001', '2026-07-01', 'Movido pelo usuário') $$,
  'an occurrence move is accepted'
);
select lives_ok(
  $$ insert into public.card_invoice_audit_events (user_id, event_type, adjustment_id, before_document, after_document) values ('c1000000-0000-0000-0000-000000000001', 'adjustment_changed', 'c1111110-0000-0000-0000-000000000001', '{"amount_cents":2000}', '{"amount_cents":2500}') $$,
  'an adjustment change event is accepted'
);
select throws_ok(
  $$ insert into public.card_invoice_audit_events (user_id, event_type, financial_item_id) values ('c1000000-0000-0000-0000-000000000001', 'occurrence_moved', 'c1111100-0000-0000-0000-000000000001') $$,
  '23514', null, 'audit event shape is enforced'
);
select throws_ok(
  $$ update public.card_invoice_audit_events set reason = 'Alterado' where id = 'c1111111-0000-0000-0000-000000000001' $$,
  '55000', 'card invoice audit events are append-only', 'audit events cannot be updated'
);
select throws_ok(
  $$ delete from public.card_invoice_audit_events where id = 'c1111111-0000-0000-0000-000000000001' $$,
  '55000', 'card invoice audit events are append-only', 'audit events cannot be deleted'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'c1000000-0000-0000-0000-000000000001', true);
select is((select count(*) from public.card_invoice_adjustments), 1::bigint, 'users see their own adjustments');
select is((select count(*) from public.card_invoice_audit_events), 2::bigint, 'users see their own audit events');
select is_empty($$ select id from public.card_invoice_adjustments where user_id = 'c2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user adjustments');
select throws_ok(
  $$ insert into public.card_invoice_audit_events (user_id, event_type, adjustment_id) values ('c1000000-0000-0000-0000-000000000001', 'adjustment_archived', 'c1111110-0000-0000-0000-000000000001') $$,
  '42501', null, 'authenticated clients cannot append audit events directly'
);
reset role;

select * from finish();
rollback;
