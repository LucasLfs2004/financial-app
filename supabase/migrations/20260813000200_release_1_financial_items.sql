create extension if not exists btree_gist with schema extensions;

create type public.financial_item_kind as enum (
  'recurring_income',
  'one_time_income',
  'fixed_expense',
  'projected_variable_expense'
);

create type public.financial_item_status as enum (
  'active',
  'archived'
);

create type public.recurrence_kind as enum (
  'monthly',
  'once'
);

create table public.financial_items (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null,
  user_id uuid not null,
  name text not null,
  kind public.financial_item_kind not null,
  description text,
  status public.financial_item_status not null default 'active',
  archived_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint financial_items_owner_matches_plan
    foreign key (plan_id, user_id)
    references public.plans (id, user_id)
    on delete cascade,
  constraint financial_items_name_not_blank
    check (btrim(name) <> ''),
  constraint financial_items_name_length
    check (char_length(name) <= 120),
  constraint financial_items_description_length
    check (description is null or char_length(description) <= 1000),
  constraint financial_items_archive_state_consistent
    check (
      (status = 'active' and archived_at is null)
      or (status = 'archived' and archived_at is not null)
    ),
  constraint financial_items_identity_scope_unique
    unique (id, plan_id, user_id)
);

comment on table public.financial_items is
  'Income and projected expense definitions owned by a financial plan.';
comment on column public.financial_items.kind is
  'Determines the financial effect; values remain stored as unsigned cents.';

create index financial_items_user_id_idx
  on public.financial_items (user_id);

create index financial_items_plan_status_idx
  on public.financial_items (plan_id, status);

create trigger financial_items_set_updated_at
before update on public.financial_items
for each row
execute function private.set_updated_at();

create table public.financial_item_periods (
  id uuid primary key default gen_random_uuid(),
  financial_item_id uuid not null,
  plan_id uuid not null,
  user_id uuid not null,
  start_month date not null,
  end_month date,
  amount_cents bigint not null,
  recurrence public.recurrence_kind not null,
  cash_month_offset smallint not null default 0,
  context text,
  recorded_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  constraint financial_item_periods_owner_matches_item
    foreign key (financial_item_id, plan_id, user_id)
    references public.financial_items (id, plan_id, user_id)
    on delete cascade,
  constraint financial_item_periods_start_month_normalized
    check (start_month = date_trunc('month', start_month)::date),
  constraint financial_item_periods_end_month_normalized
    check (
      end_month is null
      or end_month = date_trunc('month', end_month)::date
    ),
  constraint financial_item_periods_interval_valid
    check (end_month is null or start_month <= end_month),
  constraint financial_item_periods_amount_non_negative
    check (amount_cents >= 0),
  constraint financial_item_periods_cash_offset_valid
    check (cash_month_offset between 0 and 12),
  constraint financial_item_periods_context_length
    check (context is null or char_length(context) <= 500),
  constraint financial_item_periods_recurrence_valid
    check (
      recurrence = 'monthly'
      or (
        recurrence = 'once'
        and end_month is not null
        and start_month = end_month
      )
    ),
  constraint financial_item_periods_do_not_overlap
    exclude using gist (
      financial_item_id extensions.gist_uuid_ops with =,
      daterange(
        start_month,
        coalesce(end_month, 'infinity'::date),
        '[]'
      ) with &&
    )
);

comment on table public.financial_item_periods is
  'Inclusive, non-overlapping projected value periods for a financial item.';
comment on column public.financial_item_periods.cash_month_offset is
  'Number of months from economic reference to expected receipt or payment.';

create index financial_item_periods_item_start_idx
  on public.financial_item_periods (financial_item_id, start_month);

create index financial_item_periods_plan_interval_idx
  on public.financial_item_periods (plan_id, start_month, end_month);

create index financial_item_periods_user_id_idx
  on public.financial_item_periods (user_id);

alter table public.financial_items enable row level security;
alter table public.financial_item_periods enable row level security;

create policy financial_items_select_own
on public.financial_items
for select
to authenticated
using ((select auth.uid()) = user_id);

create policy financial_item_periods_select_own
on public.financial_item_periods
for select
to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.financial_items from anon;
revoke all on table public.financial_items from authenticated;
grant select on table public.financial_items to authenticated;

revoke all on table public.financial_item_periods from anon;
revoke all on table public.financial_item_periods from authenticated;
grant select on table public.financial_item_periods to authenticated;
