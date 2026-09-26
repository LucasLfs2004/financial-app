create table public.debt_early_settlements (
  id uuid primary key default gen_random_uuid(),
  financial_item_id uuid not null,
  user_id uuid not null,
  reference_month date not null,
  amount_cents bigint not null,
  reason text,
  recorded_by uuid not null,
  recorded_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  constraint debt_early_settlements_owner_matches_debt
    foreign key (financial_item_id, user_id)
    references public.debts (financial_item_id, user_id)
    on delete cascade,
  constraint debt_early_settlements_recorder_exists
    foreign key (recorded_by)
    references auth.users (id),
  constraint debt_early_settlements_one_per_debt
    unique (financial_item_id),
  constraint debt_early_settlements_reference_month_normalized
    check (reference_month = date_trunc('month', reference_month)::date),
  constraint debt_early_settlements_amount_positive
    check (amount_cents > 0),
  constraint debt_early_settlements_reason_length
    check (reason is null or char_length(reason) <= 500),
  constraint debt_early_settlements_recorder_is_owner
    check (recorded_by = user_id)
);

comment on table public.debt_early_settlements is
  'Immutable events that replace one regular installment and suppress later projections.';
comment on column public.debt_early_settlements.reference_month is
  'Competence month of the settlement, between debt start and penultimate installment.';

create or replace function private.validate_debt_early_settlement_month()
returns trigger
language plpgsql
set search_path = ''
as $$
declare
  debt_start date;
  debt_end date;
begin
  select debt.scheduled_start_month, debt.scheduled_end_month
    into debt_start, debt_end
  from public.debts as debt
  where debt.financial_item_id = new.financial_item_id
    and debt.user_id = new.user_id;

  if new.reference_month < debt_start or new.reference_month >= debt_end then
    raise exception using
      errcode = '23514',
      message = 'early settlement month must be inside the debt schedule and before its end';
  end if;

  return new;
end;
$$;

create trigger debt_early_settlements_validate_month
before insert on public.debt_early_settlements
for each row
execute function private.validate_debt_early_settlement_month();

create or replace function private.prevent_debt_early_settlement_mutation()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  raise exception using
    errcode = '55000',
    message = 'debt early settlements are append-only';
end;
$$;

create trigger debt_early_settlements_prevent_update_delete
before update or delete on public.debt_early_settlements
for each row
execute function private.prevent_debt_early_settlement_mutation();

create index debt_early_settlements_user_reference_idx
  on public.debt_early_settlements (user_id, reference_month);

alter table public.debt_early_settlements enable row level security;

create policy debt_early_settlements_select_own
on public.debt_early_settlements
for select
to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.debt_early_settlements from anon, authenticated;
grant select on table public.debt_early_settlements to authenticated;
