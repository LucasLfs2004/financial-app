create extension if not exists btree_gist with schema extensions;

create table public.saving_periods (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null,
  user_id uuid not null,
  start_month date not null,
  end_month date,
  amount_cents bigint not null,
  context text,
  created_at timestamptz not null default now(),
  constraint saving_periods_owner_matches_plan
    foreign key (plan_id, user_id)
    references public.plans (id, user_id)
    on delete cascade,
  constraint saving_periods_start_month_normalized
    check (start_month = date_trunc('month', start_month)::date),
  constraint saving_periods_end_month_normalized
    check (
      end_month is null
      or end_month = date_trunc('month', end_month)::date
    ),
  constraint saving_periods_interval_valid
    check (end_month is null or start_month <= end_month),
  constraint saving_periods_amount_non_negative
    check (amount_cents >= 0),
  constraint saving_periods_context_length
    check (context is null or char_length(context) <= 500),
  constraint saving_periods_do_not_overlap
    exclude using gist (
      plan_id extensions.gist_uuid_ops with =,
      daterange(
        start_month,
        coalesce(end_month, 'infinity'::date),
        '[]'
      ) with &&
    )
);

comment on table public.saving_periods is
  'Explicit planned-saving amounts, valid over inclusive, non-overlapping monthly periods.';
comment on column public.saving_periods.amount_cents is
  'Unsigned cents. Zero is an explicit configuration and differs from no period.';

create index saving_periods_plan_start_idx
  on public.saving_periods (plan_id, start_month);

create index saving_periods_user_id_idx
  on public.saving_periods (user_id);

alter table public.saving_periods enable row level security;

create policy saving_periods_select_own
on public.saving_periods
for select
to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.saving_periods from anon;
revoke all on table public.saving_periods from authenticated;
grant select on table public.saving_periods to authenticated;
