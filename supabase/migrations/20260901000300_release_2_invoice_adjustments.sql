create type public.invoice_adjustment_status as enum (
  'active',
  'archived'
);

create type public.card_invoice_event_type as enum (
  'occurrence_moved',
  'adjustment_changed',
  'adjustment_archived'
);

create table public.card_invoice_adjustments (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null,
  user_id uuid not null,
  credit_card_id uuid not null,
  payment_month date not null,
  reference_month date,
  name text not null,
  amount_cents bigint not null,
  context text,
  status public.invoice_adjustment_status not null default 'active',
  archived_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint card_invoice_adjustments_owner_matches_plan
    foreign key (plan_id, user_id)
    references public.plans (id, user_id)
    on delete cascade,
  constraint card_invoice_adjustments_owner_matches_card
    foreign key (credit_card_id, user_id)
    references public.credit_cards (id, user_id),
  constraint card_invoice_adjustments_name_not_blank
    check (btrim(name) <> ''),
  constraint card_invoice_adjustments_name_length
    check (char_length(name) <= 120),
  constraint card_invoice_adjustments_amount_non_negative
    check (amount_cents >= 0),
  constraint card_invoice_adjustments_payment_month_normalized
    check (payment_month = date_trunc('month', payment_month)::date),
  constraint card_invoice_adjustments_reference_month_normalized
    check (
      reference_month is null
      or reference_month = date_trunc('month', reference_month)::date
    ),
  constraint card_invoice_adjustments_context_length
    check (context is null or char_length(context) <= 500),
  constraint card_invoice_adjustments_archive_state_consistent
    check (
      (status = 'active' and archived_at is null)
      or (status = 'archived' and archived_at is not null)
    ),
  constraint card_invoice_adjustments_identity_scope_unique
    unique (id, plan_id, user_id)
);

create or replace function private.validate_card_invoice_adjustment_horizon()
returns trigger
language plpgsql
set search_path = ''
as $$
declare
  plan_start date;
  plan_end date;
begin
  select start_month, end_month
    into plan_start, plan_end
  from public.plans
  where id = new.plan_id and user_id = new.user_id;

  if new.payment_month < plan_start
     or new.payment_month > (plan_end + interval '12 months')::date then
    raise exception using
      errcode = '23514',
      message = 'adjustment payment month is outside the operational horizon';
  end if;

  if new.reference_month is not null
     and (new.reference_month < plan_start or new.reference_month > plan_end) then
    raise exception using
      errcode = '23514',
      message = 'adjustment reference month is outside the plan horizon';
  end if;

  return new;
end;
$$;

create trigger card_invoice_adjustments_validate_horizon
before insert or update on public.card_invoice_adjustments
for each row
execute function private.validate_card_invoice_adjustment_horizon();

create trigger card_invoice_adjustments_set_updated_at
before update on public.card_invoice_adjustments
for each row
execute function private.set_updated_at();

create index card_invoice_adjustments_card_payment_idx
  on public.card_invoice_adjustments (credit_card_id, payment_month, status);

create index card_invoice_adjustments_plan_reference_idx
  on public.card_invoice_adjustments (plan_id, reference_month)
  where reference_month is not null;

create index card_invoice_adjustments_user_id_idx
  on public.card_invoice_adjustments (user_id);

create table public.card_invoice_audit_events (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null,
  user_id uuid not null,
  event_type public.card_invoice_event_type not null,
  financial_item_id uuid,
  reference_month date,
  adjustment_id uuid,
  from_credit_card_id uuid,
  from_payment_month date,
  to_credit_card_id uuid,
  to_payment_month date,
  before_document jsonb,
  after_document jsonb,
  reason text,
  recorded_at timestamptz not null default now(),
  constraint card_invoice_audit_events_owner_matches_plan
    foreign key (plan_id, user_id)
    references public.plans (id, user_id)
    on delete cascade,
  constraint card_invoice_audit_events_owner_matches_item
    foreign key (financial_item_id, plan_id, user_id)
    references public.financial_items (id, plan_id, user_id),
  constraint card_invoice_audit_events_owner_matches_adjustment
    foreign key (adjustment_id, plan_id, user_id)
    references public.card_invoice_adjustments (id, plan_id, user_id),
  constraint card_invoice_audit_events_owner_matches_from_card
    foreign key (from_credit_card_id, user_id)
    references public.credit_cards (id, user_id),
  constraint card_invoice_audit_events_owner_matches_to_card
    foreign key (to_credit_card_id, user_id)
    references public.credit_cards (id, user_id),
  constraint card_invoice_audit_events_reference_month_normalized
    check (
      reference_month is null
      or reference_month = date_trunc('month', reference_month)::date
    ),
  constraint card_invoice_audit_events_from_month_normalized
    check (
      from_payment_month is null
      or from_payment_month = date_trunc('month', from_payment_month)::date
    ),
  constraint card_invoice_audit_events_to_month_normalized
    check (
      to_payment_month is null
      or to_payment_month = date_trunc('month', to_payment_month)::date
    ),
  constraint card_invoice_audit_events_before_document_object
    check (before_document is null or jsonb_typeof(before_document) = 'object'),
  constraint card_invoice_audit_events_after_document_object
    check (after_document is null or jsonb_typeof(after_document) = 'object'),
  constraint card_invoice_audit_events_reason_length
    check (reason is null or char_length(reason) <= 500),
  constraint card_invoice_audit_events_shape_valid
    check (
      (
        event_type = 'occurrence_moved'
        and financial_item_id is not null
        and reference_month is not null
        and from_credit_card_id is not null
        and from_payment_month is not null
        and to_credit_card_id is not null
        and to_payment_month is not null
        and adjustment_id is null
      )
      or (
        event_type in ('adjustment_changed', 'adjustment_archived')
        and adjustment_id is not null
      )
    )
);

create or replace function private.prevent_card_invoice_audit_event_mutation()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  raise exception using
    errcode = '55000',
    message = 'card invoice audit events are append-only';
end;
$$;

create trigger card_invoice_audit_events_prevent_update_delete
before update or delete on public.card_invoice_audit_events
for each row
execute function private.prevent_card_invoice_audit_event_mutation();

create index card_invoice_audit_events_invoice_idx
  on public.card_invoice_audit_events (to_credit_card_id, to_payment_month, recorded_at, id)
  where to_credit_card_id is not null;

create index card_invoice_audit_events_reference_idx
  on public.card_invoice_audit_events (plan_id, financial_item_id, reference_month)
  where financial_item_id is not null;

create index card_invoice_audit_events_latest_move_idx
  on public.card_invoice_audit_events (
    plan_id, financial_item_id, reference_month, recorded_at desc, id desc
  )
  where event_type = 'occurrence_moved';

create index card_invoice_audit_events_adjustment_idx
  on public.card_invoice_audit_events (adjustment_id, recorded_at, id)
  where adjustment_id is not null;

create index card_invoice_audit_events_user_id_idx
  on public.card_invoice_audit_events (user_id);

alter table public.card_invoice_adjustments enable row level security;
alter table public.card_invoice_audit_events enable row level security;

create policy card_invoice_adjustments_select_own
on public.card_invoice_adjustments for select to authenticated
using ((select auth.uid()) = user_id);

create policy card_invoice_audit_events_select_own
on public.card_invoice_audit_events for select to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.card_invoice_adjustments from anon, authenticated;
revoke all on table public.card_invoice_audit_events from anon, authenticated;
grant select on table public.card_invoice_adjustments to authenticated;
grant select on table public.card_invoice_audit_events to authenticated;
