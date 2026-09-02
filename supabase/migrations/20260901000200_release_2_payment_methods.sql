create extension if not exists btree_gist with schema extensions;

create type public.payment_method_kind as enum (
  'direct',
  'credit_card'
);

create table public.financial_item_payment_periods (
  id uuid primary key default gen_random_uuid(),
  financial_item_id uuid not null,
  plan_id uuid not null,
  user_id uuid not null,
  start_month date not null,
  end_month date,
  method public.payment_method_kind not null,
  credit_card_id uuid,
  context text,
  recorded_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  constraint financial_item_payment_periods_owner_matches_item
    foreign key (financial_item_id, plan_id, user_id)
    references public.financial_items (id, plan_id, user_id)
    on delete cascade,
  constraint financial_item_payment_periods_owner_matches_card
    foreign key (credit_card_id, user_id)
    references public.credit_cards (id, user_id),
  constraint financial_item_payment_periods_start_month_normalized
    check (start_month = date_trunc('month', start_month)::date),
  constraint financial_item_payment_periods_end_month_normalized
    check (
      end_month is null
      or end_month = date_trunc('month', end_month)::date
    ),
  constraint financial_item_payment_periods_interval_valid
    check (end_month is null or start_month <= end_month),
  constraint financial_item_payment_periods_method_card_consistent
    check (
      (method = 'direct' and credit_card_id is null)
      or (method = 'credit_card' and credit_card_id is not null)
    ),
  constraint financial_item_payment_periods_context_length
    check (context is null or char_length(context) <= 500),
  constraint financial_item_payment_periods_do_not_overlap
    exclude using gist (
      financial_item_id extensions.gist_uuid_ops with =,
      daterange(start_month, coalesce(end_month, 'infinity'::date), '[]') with &&
    )
);

create or replace function private.validate_financial_item_payment_period()
returns trigger
language plpgsql
set search_path = ''
as $$
declare
  item_kind public.financial_item_kind;
  plan_start date;
  plan_end date;
begin
  select item.kind, plan.start_month, plan.end_month
    into item_kind, plan_start, plan_end
  from public.financial_items as item
  join public.plans as plan
    on plan.id = item.plan_id and plan.user_id = item.user_id
  where item.id = new.financial_item_id
    and item.plan_id = new.plan_id
    and item.user_id = new.user_id;

  if item_kind not in ('fixed_expense', 'projected_variable_expense') then
    raise exception using
      errcode = '23514',
      message = 'payment periods can only be linked to expense items';
  end if;

  if new.start_month < plan_start
     or coalesce(new.end_month, plan_end) > plan_end then
    raise exception using
      errcode = '23514',
      message = 'payment period must be contained in the plan horizon';
  end if;

  return new;
end;
$$;

create trigger financial_item_payment_periods_validate
before insert or update on public.financial_item_payment_periods
for each row
execute function private.validate_financial_item_payment_period();

create index financial_item_payment_periods_item_start_idx
  on public.financial_item_payment_periods (financial_item_id, start_month);

create index financial_item_payment_periods_plan_interval_idx
  on public.financial_item_payment_periods (plan_id, start_month, end_month);

create index financial_item_payment_periods_card_interval_idx
  on public.financial_item_payment_periods (credit_card_id, start_month, end_month)
  where credit_card_id is not null;

create index financial_item_payment_periods_user_id_idx
  on public.financial_item_payment_periods (user_id);

alter table public.financial_item_payment_periods enable row level security;

create policy financial_item_payment_periods_select_own
on public.financial_item_payment_periods for select to authenticated
using ((select auth.uid()) = user_id);

revoke all on table public.financial_item_payment_periods from anon, authenticated;
grant select on table public.financial_item_payment_periods to authenticated;
