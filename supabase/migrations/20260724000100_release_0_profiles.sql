create schema if not exists private;

revoke all on schema private from public;
revoke all on schema private from anon;
revoke all on schema private from authenticated;

create or replace function private.set_updated_at()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

create table public.profiles (
  id uuid primary key references auth.users(id) on delete cascade,
  display_name text,
  timezone text not null default 'America/Sao_Paulo',
  currency_code text not null default 'BRL',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint profiles_display_name_not_blank
    check (display_name is null or btrim(display_name) <> ''),
  constraint profiles_timezone_not_blank
    check (btrim(timezone) <> ''),
  constraint profiles_currency_code_format
    check (currency_code ~ '^[A-Z]{3}$')
);

comment on table public.profiles is
  'Application profile associated one-to-one with a Supabase Auth user.';
comment on column public.profiles.id is
  'Supabase Auth user identifier. It is also the ownership boundary.';

create trigger profiles_set_updated_at
before update on public.profiles
for each row
execute function private.set_updated_at();

alter table public.profiles enable row level security;

create policy profiles_select_own
on public.profiles
for select
to authenticated
using ((select auth.uid()) = id);

create policy profiles_update_own
on public.profiles
for update
to authenticated
using ((select auth.uid()) = id)
with check ((select auth.uid()) = id);

revoke all on table public.profiles from anon;
revoke all on table public.profiles from authenticated;
grant select (id, display_name, timezone, currency_code, created_at, updated_at)
  on table public.profiles to authenticated;
grant update (display_name, timezone, currency_code)
  on table public.profiles to authenticated;

create or replace function private.handle_new_auth_user()
returns trigger
language plpgsql
security definer
set search_path = ''
as $$
begin
  insert into public.profiles (id, display_name)
  values (
    new.id,
    nullif(
      btrim(coalesce(new.raw_user_meta_data ->> 'name', new.raw_user_meta_data ->> 'full_name')),
      ''
    )
  )
  on conflict (id) do nothing;

  return new;
end;
$$;

create trigger auth_user_created_create_profile
after insert on auth.users
for each row
execute function private.handle_new_auth_user();

insert into public.profiles (id, display_name)
select
  users.id,
  nullif(
    btrim(coalesce(users.raw_user_meta_data ->> 'name', users.raw_user_meta_data ->> 'full_name')),
    ''
  )
from auth.users as users
on conflict (id) do nothing;

