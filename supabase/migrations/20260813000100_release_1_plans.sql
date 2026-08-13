create type public.plan_status as enum (
  'draft',
  'active',
  'archived'
);

create type public.snapshot_kind as enum (
  'original'
);

create table public.plans (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  name text not null,
  status public.plan_status not null default 'draft',
  start_month date not null,
  end_month date not null,
  currency_code text not null default 'BRL',
  activated_at timestamptz,
  archived_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint plans_name_not_blank
    check (btrim(name) <> ''),
  constraint plans_name_length
    check (char_length(name) <= 120),
  constraint plans_start_month_normalized
    check (start_month = date_trunc('month', start_month)::date),
  constraint plans_end_month_normalized
    check (end_month = date_trunc('month', end_month)::date),
  constraint plans_horizon_valid
    check (start_month <= end_month),
  constraint plans_currency_code_format
    check (currency_code ~ '^[A-Z]{3}$'),
  constraint plans_activation_state_consistent
    check (
      (status = 'draft' and activated_at is null and archived_at is null)
      or (status = 'active' and activated_at is not null and archived_at is null)
      or (status = 'archived' and archived_at is not null)
    ),
  constraint plans_id_user_id_unique
    unique (id, user_id)
);

comment on table public.plans is
  'Main financial plan. A user may have only one non-archived plan.';
comment on column public.plans.user_id is
  'Supabase Auth user identifier and ownership boundary.';
comment on column public.plans.start_month is
  'Inclusive horizon start, normalized to the first day of the month.';
comment on column public.plans.end_month is
  'Inclusive horizon end, normalized to the first day of the month.';

create index plans_user_id_idx
  on public.plans (user_id);

create unique index plans_one_current_per_user_idx
  on public.plans (user_id)
  where status <> 'archived';

create trigger plans_set_updated_at
before update on public.plans
for each row
execute function private.set_updated_at();

create table public.plan_snapshots (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null references public.plans(id) on delete cascade,
  user_id uuid not null references auth.users(id) on delete cascade,
  kind public.snapshot_kind not null,
  schema_version integer not null,
  document jsonb not null,
  created_at timestamptz not null default now(),
  constraint plan_snapshots_schema_version_positive
    check (schema_version > 0),
  constraint plan_snapshots_document_object
    check (jsonb_typeof(document) = 'object'),
  constraint plan_snapshots_owner_matches_plan
    foreign key (plan_id, user_id)
    references public.plans (id, user_id)
    on delete cascade,
  constraint plan_snapshots_one_kind_per_plan
    unique (plan_id, kind)
);

comment on table public.plan_snapshots is
  'Immutable, versioned snapshots captured during plan lifecycle transitions.';
comment on column public.plan_snapshots.document is
  'Deterministically ordered snapshot payload interpreted by schema_version.';

create index plan_snapshots_user_id_idx
  on public.plan_snapshots (user_id);

create or replace function private.prevent_plan_snapshot_update()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  raise exception using
    errcode = '55000',
    message = 'plan snapshots are immutable';
end;
$$;

create trigger plan_snapshots_prevent_update
before update on public.plan_snapshots
for each row
execute function private.prevent_plan_snapshot_update();

alter table public.plans enable row level security;
alter table public.plan_snapshots enable row level security;

create policy plans_select_own
on public.plans
for select
to authenticated
using ((select auth.uid()) = user_id);

create policy plan_snapshots_select_own
on public.plan_snapshots
for select
to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.plans from anon;
revoke all on table public.plans from authenticated;
grant select on table public.plans to authenticated;

revoke all on table public.plan_snapshots from anon;
revoke all on table public.plan_snapshots from authenticated;
grant select on table public.plan_snapshots to authenticated;
