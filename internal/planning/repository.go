package planning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(ctx context.Context, ownerID string, input CreateInput) (Plan, error) {
	const query = `
		insert into public.plans (user_id, name, start_month, end_month, currency_code)
		values ($1, $2, $3, $4, $5)
		returning id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
	`

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query,
		ownerID,
		input.Name,
		input.StartMonth.Time(),
		input.EndMonth.Time(),
		input.CurrencyCode,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return Plan{}, ErrAlreadyExists
		}
		return Plan{}, fmt.Errorf("create plan: %w", err)
	}
	return plan, nil
}

func (repository *PostgresRepository) FindCurrent(ctx context.Context, ownerID string) (Plan, error) {
	const query = `
		select id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
		from public.plans
		where user_id = $1 and status <> 'archived'
		limit 1
	`

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, ErrNotFound
	}
	if err != nil {
		return Plan{}, fmt.Errorf("find current plan: %w", err)
	}
	return plan, nil
}

func (repository *PostgresRepository) UpdateDraft(ctx context.Context, ownerID string, input UpdateInput) (Plan, error) {
	const query = `
		update public.plans
		set
			name = coalesce($2, name),
			start_month = coalesce($3, start_month),
			end_month = coalesce($4, end_month),
			currency_code = coalesce($5, currency_code)
		where user_id = $1 and status = 'draft'
		returning id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
	`

	var startMonth any
	if input.StartMonth != nil {
		startMonth = input.StartMonth.Time()
	}
	var endMonth any
	if input.EndMonth != nil {
		endMonth = input.EndMonth.Time()
	}
	var name any
	if input.Name != nil {
		name = *input.Name
	}
	var currencyCode any
	if input.CurrencyCode != nil {
		currencyCode = *input.CurrencyCode
	}

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query, ownerID, name, startMonth, endMonth, currencyCode))
	if errors.Is(err, pgx.ErrNoRows) {
		current, findErr := repository.FindCurrent(ctx, ownerID)
		if errors.Is(findErr, ErrNotFound) {
			return Plan{}, ErrNotFound
		}
		if findErr != nil {
			return Plan{}, findErr
		}
		if current.Status != domain.PlanStatusDraft {
			return Plan{}, ErrNotDraft
		}
		return Plan{}, ErrNotFound
	}
	if err != nil {
		return Plan{}, fmt.Errorf("update draft plan: %w", err)
	}
	return plan, nil
}

func (repository *PostgresRepository) Activate(ctx context.Context, ownerID string) (Activation, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Activation{}, fmt.Errorf("begin activation: %w", err)
	}
	defer tx.Rollback(ctx)
	plan, err := scanPlan(tx.QueryRow(ctx, `select id,user_id,name,status,start_month,end_month,currency_code,activated_at,archived_at,created_at,updated_at from public.plans where user_id=$1 and status<>'archived' for update`, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Activation{}, ErrNotFound
	}
	if err != nil {
		return Activation{}, fmt.Errorf("lock plan: %w", err)
	}
	if plan.Status == domain.PlanStatusActive {
		snapshot, findErr := findOriginal(ctx, tx, ownerID, plan.ID)
		if findErr != nil {
			return Activation{}, findErr
		}
		return Activation{Plan: plan, Original: snapshot}, nil
	}
	var hasIncome, hasSavings bool
	err = tx.QueryRow(ctx, `select exists(
		select 1
		from public.financial_items i
		join public.financial_item_periods p
		  on p.financial_item_id=i.id and p.user_id=i.user_id
		where i.user_id=$2 and i.currency_code=$3 and i.status='active'
		  and i.kind in ('recurring_income','one_time_income')
		  and p.start_month <= $5
		  and (p.end_month is null or p.end_month >= $4)
	), exists(select 1 from public.saving_periods where plan_id=$1 and user_id=$2)`, plan.ID, ownerID, plan.CurrencyCode, plan.StartMonth.Time(), plan.EndMonth.Time()).Scan(&hasIncome, &hasSavings)
	if err != nil {
		return Activation{}, fmt.Errorf("validate activation: %w", err)
	}
	if !hasIncome || !hasSavings {
		return Activation{}, ErrNotActivatable
	}
	var documentText string
	err = tx.QueryRow(ctx, snapshotDocumentQuery, plan.ID, ownerID).Scan(&documentText)
	if err != nil {
		return Activation{}, fmt.Errorf("build snapshot: %w", err)
	}
	var snapshot Snapshot
	err = tx.QueryRow(ctx, `insert into public.plan_snapshots(plan_id,user_id,kind,schema_version,document) values($1,$2,'original',4,$3::jsonb) returning id,plan_id,kind,schema_version,created_at,document::text`, plan.ID, ownerID, documentText).Scan(&snapshot.ID, &snapshot.PlanID, &snapshot.Kind, &snapshot.SchemaVersion, &snapshot.CapturedAt, &documentText)
	if err != nil {
		return Activation{}, fmt.Errorf("create original snapshot: %w", err)
	}
	snapshot.Plan = json.RawMessage(documentText)
	plan, err = scanPlan(tx.QueryRow(ctx, `update public.plans set status='active',activated_at=now() where id=$1 and user_id=$2 returning id,user_id,name,status,start_month,end_month,currency_code,activated_at,archived_at,created_at,updated_at`, plan.ID, ownerID))
	if err != nil {
		return Activation{}, fmt.Errorf("activate plan: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Activation{}, fmt.Errorf("commit activation: %w", err)
	}
	return Activation{Plan: plan, Original: snapshot}, nil
}

func (repository *PostgresRepository) FindOriginal(ctx context.Context, ownerID string) (Snapshot, error) {
	plan, err := repository.FindCurrent(ctx, ownerID)
	if err != nil {
		return Snapshot{}, err
	}
	return findOriginal(ctx, repository.pool, ownerID, plan.ID)
}

type rowQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func findOriginal(ctx context.Context, database rowQueryer, ownerID, planID string) (Snapshot, error) {
	var result Snapshot
	var documentText string
	err := database.QueryRow(ctx, `select id,plan_id,kind,schema_version,created_at,document::text from public.plan_snapshots where plan_id=$1 and user_id=$2 and kind='original'`, planID, ownerID).Scan(&result.ID, &result.PlanID, &result.Kind, &result.SchemaVersion, &result.CapturedAt, &documentText)
	if errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrNotFound
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("find original snapshot: %w", err)
	}
	result.Plan = json.RawMessage(documentText)
	return result, nil
}

const snapshotDocumentQuery = `
select jsonb_build_object(
  'name', p.name, 'start_month', to_char(p.start_month,'YYYY-MM'),
  'end_month', to_char(p.end_month,'YYYY-MM'), 'currency_code',p.currency_code,
  'items', coalesce((select jsonb_agg(jsonb_build_object(
    'id',i.id,'currency_code',i.currency_code,'name',i.name,'kind',i.kind,'description',i.description,
    'status',i.status,'archived_at',i.archived_at,'created_at',i.created_at,'updated_at',i.updated_at,
    'periods',coalesce((select jsonb_agg(jsonb_build_object(
      'id',ip.id,'start_month',to_char(ip.start_month,'YYYY-MM'),'end_month',case when ip.end_month is null then null else to_char(ip.end_month,'YYYY-MM') end,
      'amount_cents',ip.amount_cents,'recurrence',ip.recurrence,'cash_month_offset',ip.cash_month_offset,'context',ip.context,'recorded_at',ip.recorded_at,'created_at',ip.created_at
    ) order by ip.start_month,ip.id) from public.financial_item_periods ip where ip.financial_item_id=i.id and ip.user_id=i.user_id),'[]'::jsonb)
  ) order by i.created_at,i.id) from public.financial_items i
    where i.user_id=$2 and i.currency_code=p.currency_code
      and exists(select 1 from public.financial_item_periods visible
        where visible.financial_item_id=i.id and visible.user_id=i.user_id
          and visible.start_month<=p.end_month
          and (visible.end_month is null or visible.end_month>=p.start_month))),'[]'::jsonb),
  'savings',jsonb_build_object('configured',exists(select 1 from public.saving_periods s where s.plan_id=p.id and s.user_id=$2),
    'periods',coalesce((select jsonb_agg(jsonb_build_object('id',s.id,'start_month',to_char(s.start_month,'YYYY-MM'),'end_month',case when s.end_month is null then null else to_char(s.end_month,'YYYY-MM') end,'amount_cents',s.amount_cents,'context',s.context,'created_at',s.created_at) order by s.start_month,s.id) from public.saving_periods s where s.plan_id=p.id and s.user_id=$2),'[]'::jsonb)),
  'release_2',jsonb_build_object(
    'financial_institutions',coalesce((select jsonb_agg(jsonb_build_object('id',fi.id,'name',fi.name,'status',fi.status) order by fi.id) from public.financial_institutions fi where fi.user_id=$2 and exists(select 1 from public.credit_cards cc where cc.institution_id=fi.id and cc.user_id=$2 and (exists(select 1 from public.financial_item_payment_periods pp join public.financial_items pi on pi.id=pp.financial_item_id and pi.user_id=pp.user_id where pp.credit_card_id=cc.id and pp.user_id=$2 and pi.currency_code=p.currency_code) or exists(select 1 from public.card_invoice_adjustments ia where ia.credit_card_id=cc.id and ia.user_id=$2 and ia.currency_code=p.currency_code) or exists(select 1 from public.card_invoice_audit_events ae join public.financial_items mi on mi.id=ae.financial_item_id and mi.user_id=ae.user_id where ae.user_id=$2 and mi.currency_code=p.currency_code and (ae.from_credit_card_id=cc.id or ae.to_credit_card_id=cc.id))))),'[]'::jsonb),
    'credit_cards',coalesce((select jsonb_agg(jsonb_build_object('id',cc.id,'institution_id',cc.institution_id,'name',cc.name,'status',cc.status,'periods',coalesce((select jsonb_agg(jsonb_build_object('id',cp.id,'start_month',to_char(cp.start_month,'YYYY-MM'),'end_month',case when cp.end_month is null then null else to_char(cp.end_month,'YYYY-MM') end,'nominal_due_day',cp.nominal_due_day,'payment_month_offset',cp.payment_month_offset) order by cp.start_month,cp.id) from public.credit_card_periods cp where cp.credit_card_id=cc.id and cp.user_id=$2),'[]'::jsonb)) order by cc.id) from public.credit_cards cc where cc.user_id=$2 and (exists(select 1 from public.financial_item_payment_periods pp join public.financial_items pi on pi.id=pp.financial_item_id and pi.user_id=pp.user_id where pp.credit_card_id=cc.id and pp.user_id=$2 and pi.currency_code=p.currency_code) or exists(select 1 from public.card_invoice_adjustments ia where ia.credit_card_id=cc.id and ia.user_id=$2 and ia.currency_code=p.currency_code) or exists(select 1 from public.card_invoice_audit_events ae join public.financial_items mi on mi.id=ae.financial_item_id and mi.user_id=ae.user_id where ae.user_id=$2 and mi.currency_code=p.currency_code and (ae.from_credit_card_id=cc.id or ae.to_credit_card_id=cc.id)))),'[]'::jsonb),
    'payment_methods',coalesce((select jsonb_agg(to_jsonb(pp) order by pp.start_month,pp.id) from public.financial_item_payment_periods pp join public.financial_items pi on pi.id=pp.financial_item_id and pi.user_id=pp.user_id where pp.user_id=$2 and pi.currency_code=p.currency_code and exists(select 1 from public.financial_item_periods visible where visible.financial_item_id=pi.id and visible.user_id=pi.user_id and visible.start_month<=p.end_month and (visible.end_month is null or visible.end_month>=p.start_month))),'[]'::jsonb),
    'invoice_adjustments',coalesce((select jsonb_agg(to_jsonb(ia) order by ia.payment_month,ia.id) from public.card_invoice_adjustments ia where ia.user_id=$2 and ia.currency_code=p.currency_code and ((ia.reference_month between p.start_month and p.end_month) or (ia.payment_month between p.start_month and (p.end_month + interval '12 months')::date))),'[]'::jsonb),
    'invoice_moves',coalesce((select jsonb_agg(to_jsonb(ae) order by ae.recorded_at,ae.id) from public.card_invoice_audit_events ae join public.financial_items mi on mi.id=ae.financial_item_id and mi.user_id=ae.user_id where ae.user_id=$2 and mi.currency_code=p.currency_code and ae.reference_month between p.start_month and p.end_month),'[]'::jsonb)),
  'release_3',jsonb_build_object(
    'debts',coalesce((select jsonb_agg(jsonb_build_object(
      'financial_item_id',d.financial_item_id,
      'original_total_cents',d.original_total_cents,
      'total_installments',d.total_installments,
      'first_projected_installment',d.first_projected_installment,
      'scheduled_start_month',to_char(d.scheduled_start_month,'YYYY-MM'),
      'scheduled_end_month',to_char(d.scheduled_end_month,'YYYY-MM'),
      'created_at',d.created_at,'updated_at',d.updated_at,
      'periods',coalesce((select jsonb_agg(to_jsonb(dp) order by dp.start_month,dp.id)
        from public.financial_item_periods dp
        where dp.financial_item_id=d.financial_item_id and dp.user_id=$2),'[]'::jsonb),
      'payment_methods',coalesce((select jsonb_agg(to_jsonb(pm) order by pm.start_month,pm.id)
        from public.financial_item_payment_periods pm
        where pm.financial_item_id=d.financial_item_id and pm.user_id=$2),'[]'::jsonb),
      'early_settlement',(select to_jsonb(es) from public.debt_early_settlements es
        where es.financial_item_id=d.financial_item_id and es.user_id=$2)
    ) order by d.scheduled_start_month,d.financial_item_id)
      from public.debts d
      join public.financial_items di on di.id=d.financial_item_id and di.user_id=d.user_id
      where d.user_id=$2 and di.currency_code=p.currency_code
        and d.scheduled_start_month<=p.end_month and d.scheduled_end_month>=p.start_month),'[]'::jsonb))
)::text from public.plans p where p.id=$1 and p.user_id=$2`

func scanPlan(row pgx.Row) (Plan, error) {
	var result Plan
	var status string
	var startMonth time.Time
	var endMonth time.Time
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.Name,
		&status,
		&startMonth,
		&endMonth,
		&result.CurrencyCode,
		&result.ActivatedAt,
		&result.ArchivedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return Plan{}, err
	}

	parsedStatus, err := domain.ParsePlanStatus(status)
	if err != nil {
		return Plan{}, fmt.Errorf("parse plan status: %w", err)
	}
	parsedStartMonth, err := domain.NewYearMonth(startMonth.Year(), startMonth.Month())
	if err != nil {
		return Plan{}, fmt.Errorf("parse start month: %w", err)
	}
	parsedEndMonth, err := domain.NewYearMonth(endMonth.Year(), endMonth.Month())
	if err != nil {
		return Plan{}, fmt.Errorf("parse end month: %w", err)
	}
	result.Status = parsedStatus
	result.StartMonth = parsedStartMonth
	result.EndMonth = parsedEndMonth
	return result, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
