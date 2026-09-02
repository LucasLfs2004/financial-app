create extension if not exists btree_gist with schema extensions;

create type public.financial_resource_status as enum (
  'active',
  'archived'
);

create table public.financial_institutions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  name text not null,
  status public.financial_resource_status not null default 'active',
  archived_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint financial_institutions_name_not_blank
    check (btrim(name) <> ''),
  constraint financial_institutions_name_length
    check (char_length(name) <= 120),
  constraint financial_institutions_archive_state_consistent
    check (
      (status = 'active' and archived_at is null)
      or (status = 'archived' and archived_at is not null)
    ),
  constraint financial_institutions_identity_scope_unique
    unique (id, user_id)
);

create index financial_institutions_user_status_idx
  on public.financial_institutions (user_id, status);

create unique index financial_institutions_active_name_idx
  on public.financial_institutions (user_id, lower(btrim(name)))
  where status = 'active';

create trigger financial_institutions_set_updated_at
before update on public.financial_institutions
for each row
execute function private.set_updated_at();

create table public.credit_cards (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  institution_id uuid not null,
  name text not null,
  status public.financial_resource_status not null default 'active',
  archived_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint credit_cards_owner_matches_institution
    foreign key (institution_id, user_id)
    references public.financial_institutions (id, user_id),
  constraint credit_cards_name_not_blank
    check (btrim(name) <> ''),
  constraint credit_cards_name_length
    check (char_length(name) <= 120),
  constraint credit_cards_archive_state_consistent
    check (
      (status = 'active' and archived_at is null)
      or (status = 'archived' and archived_at is not null)
    ),
  constraint credit_cards_identity_scope_unique
    unique (id, user_id)
);

create index credit_cards_user_status_idx
  on public.credit_cards (user_id, status);

create index credit_cards_institution_idx
  on public.credit_cards (institution_id);

create unique index credit_cards_active_name_idx
  on public.credit_cards (user_id, institution_id, lower(btrim(name)))
  where status = 'active';

create or replace function private.prevent_financial_resource_reactivation()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  if old.status = 'archived' and new.status = 'active' then
    raise exception using
      errcode = '55000',
      message = 'archived financial resources cannot be reactivated';
  end if;
  return new;
end;
$$;

create or replace function private.prevent_credit_card_institution_change()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  if old.institution_id is distinct from new.institution_id then
    raise exception using
      errcode = '55000',
      message = 'credit card institution is immutable';
  end if;
  return new;
end;
$$;

create trigger credit_cards_prevent_reactivation
before update on public.credit_cards
for each row
execute function private.prevent_financial_resource_reactivation();

create trigger credit_cards_prevent_institution_change
before update on public.credit_cards
for each row
execute function private.prevent_credit_card_institution_change();

create trigger credit_cards_set_updated_at
before update on public.credit_cards
for each row
execute function private.set_updated_at();

create table public.credit_card_periods (
  id uuid primary key default gen_random_uuid(),
  credit_card_id uuid not null,
  user_id uuid not null,
  start_month date not null,
  end_month date,
  nominal_due_day smallint not null,
  payment_month_offset smallint not null default 1,
  context text,
  recorded_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  constraint credit_card_periods_owner_matches_card
    foreign key (credit_card_id, user_id)
    references public.credit_cards (id, user_id)
    on delete cascade,
  constraint credit_card_periods_start_month_normalized
    check (start_month = date_trunc('month', start_month)::date),
  constraint credit_card_periods_end_month_normalized
    check (
      end_month is null
      or end_month = date_trunc('month', end_month)::date
    ),
  constraint credit_card_periods_interval_valid
    check (end_month is null or start_month <= end_month),
  constraint credit_card_periods_due_day_valid
    check (nominal_due_day between 1 and 31),
  constraint credit_card_periods_payment_offset_valid
    check (payment_month_offset between 0 and 12),
  constraint credit_card_periods_context_length
    check (context is null or char_length(context) <= 500),
  constraint credit_card_periods_do_not_overlap
    exclude using gist (
      credit_card_id extensions.gist_uuid_ops with =,
      daterange(start_month, coalesce(end_month, 'infinity'::date), '[]') with &&
    )
);

create index credit_card_periods_card_start_idx
  on public.credit_card_periods (credit_card_id, start_month);

create index credit_card_periods_user_interval_idx
  on public.credit_card_periods (user_id, start_month, end_month);

alter table public.financial_institutions enable row level security;
alter table public.credit_cards enable row level security;
alter table public.credit_card_periods enable row level security;

create policy financial_institutions_select_own
on public.financial_institutions for select to authenticated
using ((select auth.uid()) = user_id);

create policy credit_cards_select_own
on public.credit_cards for select to authenticated
using ((select auth.uid()) = user_id);

create policy credit_card_periods_select_own
on public.credit_card_periods for select to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.financial_institutions from anon, authenticated;
revoke all on table public.credit_cards from anon, authenticated;
revoke all on table public.credit_card_periods from anon, authenticated;
grant select on table public.financial_institutions to authenticated;
grant select on table public.credit_cards to authenticated;
grant select on table public.credit_card_periods to authenticated;
