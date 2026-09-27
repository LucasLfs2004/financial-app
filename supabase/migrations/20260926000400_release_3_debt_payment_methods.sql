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

  if item_kind not in (
    'fixed_expense',
    'projected_variable_expense',
    'debt_installment'
  ) then
    raise exception using
      errcode = '23514',
      message = 'payment periods can only be linked to expense items';
  end if;

  return new;
end;
$$;

comment on function private.validate_financial_item_payment_period() is
  'Allows payment choices only for projected expenses, including debts.';
