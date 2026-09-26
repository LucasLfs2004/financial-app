create table public.debts (
  financial_item_id uuid primary key,
  user_id uuid not null,
  original_total_cents bigint,
  total_installments integer not null,
  first_projected_installment integer not null,
  scheduled_start_month date not null,
  scheduled_end_month date not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint debts_owner_matches_item
    foreign key (financial_item_id, user_id)
    references public.financial_items (id, user_id)
    on delete cascade,
  constraint debts_identity_scope_unique
    unique (financial_item_id, user_id),
  constraint debts_original_total_non_negative
    check (original_total_cents is null or original_total_cents >= 0),
  constraint debts_total_installments_positive
    check (total_installments > 0),
  constraint debts_first_projected_installment_valid
    check (
      first_projected_installment between 1 and total_installments
    ),
  constraint debts_start_month_normalized
    check (
      scheduled_start_month = date_trunc('month', scheduled_start_month)::date
    ),
  constraint debts_end_month_normalized
    check (
      scheduled_end_month = date_trunc('month', scheduled_end_month)::date
    ),
  constraint debts_schedule_within_operational_limit
    check (
      total_installments - first_projected_installment + 1 between 1 and 120
    ),
  constraint debts_scheduled_end_consistent
    check (
      scheduled_end_month = (
        scheduled_start_month
        + make_interval(
            months => total_installments - first_projected_installment
          )
      )::date
    )
);

comment on table public.debts is
  'Structural metadata for user-owned projected installment debts.';
comment on column public.debts.financial_item_id is
  'Debt public identity, shared with its specialized financial item.';
comment on column public.debts.original_total_cents is
  'Optional informational original total; it does not derive installments.';

create index debts_user_scheduled_end_idx
  on public.debts (user_id, scheduled_end_month);

create trigger debts_set_updated_at
before update on public.debts
for each row
execute function private.set_updated_at();

create or replace function private.assert_debt_schedule_integrity(
  target_financial_item_id uuid
)
returns void
language plpgsql
set search_path = ''
as $$
declare
  item_kind public.financial_item_kind;
  debt_row public.debts%rowtype;
  debt_exists boolean;
begin
  select item.kind
    into item_kind
  from public.financial_items as item
  where item.id = target_financial_item_id;

  if not found then
    return;
  end if;

  select debt.*
    into debt_row
  from public.debts as debt
  where debt.financial_item_id = target_financial_item_id;
  debt_exists := found;

  if item_kind = 'debt_installment' and not debt_exists then
    raise exception using
      errcode = '23514',
      message = 'debt installment items require debt metadata';
  end if;

  if item_kind <> 'debt_installment' and debt_exists then
    raise exception using
      errcode = '23514',
      message = 'debt metadata requires a debt installment item';
  end if;

  if not debt_exists then
    return;
  end if;

  if exists (
    select 1
    from public.financial_item_periods as period
    where period.financial_item_id = target_financial_item_id
      and (
        period.user_id <> debt_row.user_id
        or period.recurrence <> 'monthly'
        or period.end_month is null
        or period.amount_cents <= 0
        or period.start_month < debt_row.scheduled_start_month
        or period.end_month > debt_row.scheduled_end_month
      )
  ) then
    raise exception using
      errcode = '23514',
      message = 'debt periods must be positive monthly intervals inside the schedule';
  end if;

  if exists (
    select 1
    from generate_series(
      debt_row.scheduled_start_month::timestamp,
      debt_row.scheduled_end_month::timestamp,
      interval '1 month'
    ) as month_series(reference_month)
    where not exists (
      select 1
      from public.financial_item_periods as period
      where period.financial_item_id = target_financial_item_id
        and month_series.reference_month::date between
          period.start_month and period.end_month
    )
  ) then
    raise exception using
      errcode = '23514',
      message = 'debt periods must continuously cover the schedule';
  end if;
end;
$$;

create or replace function private.validate_debt_schedule_integrity()
returns trigger
language plpgsql
set search_path = ''
as $$
declare
  previous_financial_item_id uuid;
  current_financial_item_id uuid;
begin
  if tg_table_name = 'financial_items' then
    if tg_op <> 'INSERT' then
      previous_financial_item_id := old.id;
    end if;
    if tg_op <> 'DELETE' then
      current_financial_item_id := new.id;
    end if;
  else
    if tg_op <> 'INSERT' then
      previous_financial_item_id := old.financial_item_id;
    end if;
    if tg_op <> 'DELETE' then
      current_financial_item_id := new.financial_item_id;
    end if;
  end if;

  if previous_financial_item_id is not null
     and previous_financial_item_id is distinct from current_financial_item_id then
    perform private.assert_debt_schedule_integrity(previous_financial_item_id);
  end if;
  if current_financial_item_id is not null then
    perform private.assert_debt_schedule_integrity(current_financial_item_id);
  end if;

  return null;
end;
$$;

create constraint trigger financial_items_validate_debt_schedule
after insert or update or delete on public.financial_items
deferrable initially deferred
for each row
execute function private.validate_debt_schedule_integrity();

create constraint trigger financial_item_periods_validate_debt_schedule
after insert or update or delete on public.financial_item_periods
deferrable initially deferred
for each row
execute function private.validate_debt_schedule_integrity();

create constraint trigger debts_validate_schedule
after insert or update or delete on public.debts
deferrable initially deferred
for each row
execute function private.validate_debt_schedule_integrity();

alter table public.debts enable row level security;

create policy debts_select_own
on public.debts
for select
to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.debts from anon, authenticated;
grant select on table public.debts to authenticated;
