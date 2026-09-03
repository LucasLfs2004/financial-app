begin;

create extension if not exists pgtap with schema extensions;

select plan(23);

select has_type('public', 'financial_resource_status', 'financial resource status enum exists');
select has_table('public', 'financial_institutions', 'financial institutions table exists');
select has_table('public', 'credit_cards', 'credit cards table exists');
select has_table('public', 'credit_card_periods', 'credit card periods table exists');
select has_index('public', 'financial_institutions', 'financial_institutions_active_name_idx', 'active institution names are indexed');
select has_index('public', 'credit_cards', 'credit_cards_active_name_idx', 'active card names are indexed');
select has_index('public', 'credit_card_periods', 'credit_card_periods_card_start_idx', 'card periods have a lookup index');

insert into auth.users (id, instance_id, aud, role, email, encrypted_password, email_confirmed_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at)
values
  ('a1000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'cards-owner-1@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now()),
  ('a2000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', 'cards-owner-2@example.com', '', now(), '{"provider":"email","providers":["email"]}', '{}', now(), now());

insert into public.financial_institutions (id, user_id, name)
values
  ('a1100000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Banco Alfa'),
  ('a2200000-0000-0000-0000-000000000002', 'a2000000-0000-0000-0000-000000000002', 'Banco Beta');

select throws_ok(
  $$ insert into public.financial_institutions (user_id, name) values ('a1000000-0000-0000-0000-000000000001', '  banco ALFA ') $$,
  '23505', null, 'active institution names are unique ignoring case and surrounding spaces'
);

insert into public.credit_cards (id, user_id, institution_id, name)
values
  ('a1110000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'a1100000-0000-0000-0000-000000000001', 'Cartão principal'),
  ('a1120000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'a1100000-0000-0000-0000-000000000001', 'Cartão reserva'),
  ('a2220000-0000-0000-0000-000000000002', 'a2000000-0000-0000-0000-000000000002', 'a2200000-0000-0000-0000-000000000002', 'Cartão secundário');

select throws_ok(
  $$ insert into public.credit_cards (user_id, institution_id, name) values ('a1000000-0000-0000-0000-000000000001', 'a2200000-0000-0000-0000-000000000002', 'Cartão inválido') $$,
  '23503', null, 'a card cannot use another user institution'
);

select throws_ok(
  $$ insert into public.credit_cards (user_id, institution_id, name) values ('a1000000-0000-0000-0000-000000000001', 'a1100000-0000-0000-0000-000000000001', ' cartão PRINCIPAL ') $$,
  '23505', null, 'active card names are unique per institution ignoring case and spaces'
);

select throws_ok(
  $$ update public.credit_cards set institution_id = 'a2200000-0000-0000-0000-000000000002' where id = 'a1110000-0000-0000-0000-000000000001' $$,
  '55000', 'credit card institution is immutable', 'a card cannot change institution after creation'
);

insert into public.credit_card_periods (id, credit_card_id, user_id, start_month, end_month, nominal_due_day, payment_month_offset)
values
  ('a1111000-0000-0000-0000-000000000001', 'a1110000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', '2026-01-01', '2026-06-01', 10, 1),
  ('a2222000-0000-0000-0000-000000000002', 'a2220000-0000-0000-0000-000000000002', 'a2000000-0000-0000-0000-000000000002', '2026-01-01', null, 5, 1);

select lives_ok(
  $$ insert into public.credit_card_periods (credit_card_id, user_id, start_month, end_month, nominal_due_day, payment_month_offset) values ('a1110000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', '2026-07-01', null, 15, 1) $$,
  'adjacent card configuration periods are accepted'
);

select throws_ok(
  $$ insert into public.credit_card_periods (credit_card_id, user_id, start_month, end_month, nominal_due_day) values ('a1110000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', '2026-06-01', '2026-08-01', 20) $$,
  '23P01', null, 'card configuration periods cannot overlap'
);

select throws_ok(
  $$ insert into public.credit_card_periods (credit_card_id, user_id, start_month, nominal_due_day) values ('a1110000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', '2028-01-01', 32) $$,
  '23514', null, 'nominal due day must be between one and thirty-one'
);

select throws_ok(
  $$ insert into public.credit_card_periods (credit_card_id, user_id, start_month, nominal_due_day) values ('a1120000-0000-0000-0000-000000000001', 'a2000000-0000-0000-0000-000000000002', '2028-01-01', 10) $$,
  '23503', null, 'card period ownership must match card ownership'
);

update public.credit_cards set status = 'archived', archived_at = now() where id = 'a1110000-0000-0000-0000-000000000001';
select throws_ok(
  $$ update public.credit_cards set status = 'active', archived_at = null where id = 'a1110000-0000-0000-0000-000000000001' $$,
  '55000', 'archived financial resources cannot be reactivated', 'archived cards cannot be reactivated'
);

set local role authenticated;
select set_config('request.jwt.claim.sub', 'a1000000-0000-0000-0000-000000000001', true);

select is((select count(*) from public.financial_institutions), 1::bigint, 'users see their own institutions');
select is_empty($$ select id from public.financial_institutions where user_id = 'a2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user institutions');
select is((select count(*) from public.credit_cards), 2::bigint, 'users see their own cards');
select is_empty($$ select id from public.credit_cards where user_id = 'a2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user cards');
select is((select count(*) from public.credit_card_periods), 2::bigint, 'users see their own card periods');
select is_empty($$ select id from public.credit_card_periods where user_id = 'a2000000-0000-0000-0000-000000000002' $$, 'users cannot see another user card periods');
select throws_ok(
  $$ insert into public.financial_institutions (user_id, name) values ('a1000000-0000-0000-0000-000000000001', 'Sem grant') $$,
  '42501', null, 'authenticated clients cannot write institutions directly'
);

select * from finish();
rollback;
