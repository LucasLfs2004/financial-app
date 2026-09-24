-- Release 2.1: financial records belong to the user. Plans define projection
-- windows and keep only plan-specific configuration such as saving periods.

alter table public.financial_items
  add column currency_code text;

alter table public.card_invoice_adjustments
  add column currency_code text;

update public.financial_items as item
set currency_code = plan.currency_code
from public.plans as plan
where plan.id = item.plan_id
  and plan.user_id = item.user_id;

update public.card_invoice_adjustments as adjustment
set currency_code = plan.currency_code
from public.plans as plan
where plan.id = adjustment.plan_id
  and plan.user_id = adjustment.user_id;

alter table public.financial_items
  alter column currency_code set not null,
  add constraint financial_items_currency_code_format
    check (currency_code ~ '^[A-Z]{3}$'),
  add constraint financial_items_user_owner
    foreign key (user_id)
    references auth.users (id)
    on delete cascade,
  add constraint financial_items_id_user_id_unique
    unique (id, user_id);

alter table public.card_invoice_adjustments
  alter column currency_code set not null,
  add constraint card_invoice_adjustments_currency_code_format
    check (currency_code ~ '^[A-Z]{3}$'),
  add constraint card_invoice_adjustments_user_owner
    foreign key (user_id)
    references auth.users (id)
    on delete cascade,
  add constraint card_invoice_adjustments_id_user_id_unique
    unique (id, user_id);

alter table public.card_invoice_audit_events
  add constraint card_invoice_audit_events_user_owner
    foreign key (user_id)
    references auth.users (id)
    on delete cascade;

alter table public.financial_item_periods
  add constraint financial_item_periods_owner_matches_item_v2
    foreign key (financial_item_id, user_id)
    references public.financial_items (id, user_id)
    on delete cascade;

alter table public.financial_item_payment_periods
  add constraint financial_item_payment_periods_owner_matches_item_v2
    foreign key (financial_item_id, user_id)
    references public.financial_items (id, user_id)
    on delete cascade;

alter table public.card_invoice_audit_events
  add constraint card_invoice_audit_events_owner_matches_item_v2
    foreign key (financial_item_id, user_id)
    references public.financial_items (id, user_id),
  add constraint card_invoice_audit_events_owner_matches_adjustment_v2
    foreign key (adjustment_id, user_id)
    references public.card_invoice_adjustments (id, user_id);

drop trigger financial_item_payment_periods_validate
  on public.financial_item_payment_periods;
drop function private.validate_financial_item_payment_period();

drop trigger card_invoice_adjustments_validate_horizon
  on public.card_invoice_adjustments;
drop function private.validate_card_invoice_adjustment_horizon();

drop index public.financial_items_plan_status_idx;
drop index public.financial_item_periods_plan_interval_idx;
drop index public.financial_item_payment_periods_plan_interval_idx;
drop index public.card_invoice_adjustments_plan_reference_idx;
drop index public.card_invoice_audit_events_reference_idx;
drop index public.card_invoice_audit_events_latest_move_idx;

alter table public.card_invoice_audit_events
  drop constraint card_invoice_audit_events_owner_matches_plan,
  drop constraint card_invoice_audit_events_owner_matches_item,
  drop constraint card_invoice_audit_events_owner_matches_adjustment;

alter table public.financial_item_payment_periods
  drop constraint financial_item_payment_periods_owner_matches_item;

alter table public.financial_item_periods
  drop constraint financial_item_periods_owner_matches_item;

alter table public.card_invoice_adjustments
  drop constraint card_invoice_adjustments_owner_matches_plan,
  drop constraint card_invoice_adjustments_identity_scope_unique;

alter table public.financial_items
  drop constraint financial_items_owner_matches_plan,
  drop constraint financial_items_identity_scope_unique;

alter table public.card_invoice_audit_events
  drop column plan_id;

alter table public.card_invoice_adjustments
  drop column plan_id;

alter table public.financial_item_payment_periods
  drop column plan_id;

alter table public.financial_item_periods
  drop column plan_id;

alter table public.financial_items
  drop column plan_id;

create or replace function private.validate_financial_item_payment_period()
returns trigger
language plpgsql
set search_path = ''
as $$
declare
  item_kind public.financial_item_kind;
begin
  select item.kind
    into item_kind
  from public.financial_items as item
  where item.id = new.financial_item_id
    and item.user_id = new.user_id;

  if item_kind not in ('fixed_expense', 'projected_variable_expense') then
    raise exception using
      errcode = '23514',
      message = 'payment periods can only be linked to expense items';
  end if;

  return new;
end;
$$;

create trigger financial_item_payment_periods_validate
before insert or update on public.financial_item_payment_periods
for each row
execute function private.validate_financial_item_payment_period();

create index financial_items_user_currency_status_idx
  on public.financial_items (user_id, currency_code, status);

create index financial_item_periods_user_interval_idx
  on public.financial_item_periods (user_id, start_month, end_month);

create index financial_item_payment_periods_user_interval_idx
  on public.financial_item_payment_periods (user_id, start_month, end_month);

create index card_invoice_adjustments_user_reference_idx
  on public.card_invoice_adjustments (user_id, currency_code, reference_month)
  where reference_month is not null;

create index card_invoice_audit_events_reference_idx
  on public.card_invoice_audit_events (
    user_id, financial_item_id, reference_month
  )
  where financial_item_id is not null;

create index card_invoice_audit_events_latest_move_idx
  on public.card_invoice_audit_events (
    user_id, financial_item_id, reference_month, recorded_at desc, id desc
  )
  where event_type = 'occurrence_moved';

comment on table public.financial_items is
  'User-owned income and projected expense definitions, independent of plans.';
comment on column public.financial_items.currency_code is
  'Currency interpreted by projections; no implicit conversion is performed.';
comment on table public.financial_item_periods is
  'Inclusive projected value periods independent of any plan horizon.';
comment on table public.financial_item_payment_periods is
  'User-owned payment choices for financial item periods.';
comment on table public.card_invoice_adjustments is
  'User-owned projected invoice adjustments independent of plan lifecycle.';
comment on column public.card_invoice_adjustments.currency_code is
  'Currency interpreted by projections; no implicit conversion is performed.';
comment on table public.card_invoice_audit_events is
  'Append-only user-owned invoice events independent of plan lifecycle.';
