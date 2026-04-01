package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionPlanRepository struct {
	client *dbent.Client
}

func NewSubscriptionPlanRepository(client *dbent.Client) service.SubscriptionPlanRepository {
	return &subscriptionPlanRepository{client: client}
}

func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id int64) (*service.SubscriptionPlan, error) {
	m, err := r.client.SubscriptionPlan.Query().
		Where(
			subscriptionplan.ID(id),
			subscriptionplan.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPlanNotFound, nil)
	}
	return subscriptionPlanEntityToService(m), nil
}

func (r *subscriptionPlanRepository) List(ctx context.Context, params pagination.PaginationParams, visibility, status string) ([]service.SubscriptionPlan, *pagination.PaginationResult, error) {
	q := r.client.SubscriptionPlan.Query().
		Where(subscriptionplan.DeletedAtIsNil())

	if visibility != "" {
		q = q.Where(subscriptionplan.VisibilityEQ(visibility))
	}
	if status != "" {
		q = q.Where(subscriptionplan.StatusEQ(status))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	plans, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Asc(subscriptionplan.FieldSortOrder), dbent.Asc(subscriptionplan.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := make([]service.SubscriptionPlan, 0, len(plans))
	for i := range plans {
		out = append(out, *subscriptionPlanEntityToService(plans[i]))
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *subscriptionPlanRepository) ListAll(ctx context.Context) ([]service.SubscriptionPlan, error) {
	plans, err := r.client.SubscriptionPlan.Query().
		Where(
			subscriptionplan.DeletedAtIsNil(),
			subscriptionplan.StatusEQ(service.StatusActive),
		).
		Order(dbent.Asc(subscriptionplan.FieldSortOrder), dbent.Asc(subscriptionplan.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]service.SubscriptionPlan, 0, len(plans))
	for i := range plans {
		out = append(out, *subscriptionPlanEntityToService(plans[i]))
	}
	return out, nil
}

func (r *subscriptionPlanRepository) Create(ctx context.Context, plan *service.SubscriptionPlan) error {
	builder := r.client.SubscriptionPlan.Create().
		SetName(plan.Name).
		SetNillableDescription(nilIfEmpty(plan.Description)).
		SetVisibility(plan.Visibility).
		SetNillableDailyLimitUsd(plan.DailyLimitUSD).
		SetNillableWeeklyLimitUsd(plan.WeeklyLimitUSD).
		SetNillableMonthlyLimitUsd(plan.MonthlyLimitUSD).
		SetDefaultValidityDays(plan.DefaultValidityDays).
		SetNillablePrice(plan.Price).
		SetSortOrder(plan.SortOrder)

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	plan.ID = created.ID
	plan.Status = created.Status
	plan.CreatedAt = created.CreatedAt
	plan.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *subscriptionPlanRepository) Update(ctx context.Context, plan *service.SubscriptionPlan) error {
	builder := r.client.SubscriptionPlan.UpdateOneID(plan.ID).
		SetName(plan.Name).
		SetVisibility(plan.Visibility).
		SetStatus(plan.Status).
		SetDefaultValidityDays(plan.DefaultValidityDays).
		SetSortOrder(plan.SortOrder)

	// Description: 空字符串 → 清除，非空 → 设置
	if plan.Description != "" {
		builder = builder.SetDescription(plan.Description)
	} else {
		builder = builder.ClearDescription()
	}

	// Nullable float fields: nil → 清除，非 nil → 设置
	if plan.DailyLimitUSD != nil {
		builder = builder.SetDailyLimitUsd(*plan.DailyLimitUSD)
	} else {
		builder = builder.ClearDailyLimitUsd()
	}
	if plan.WeeklyLimitUSD != nil {
		builder = builder.SetWeeklyLimitUsd(*plan.WeeklyLimitUSD)
	} else {
		builder = builder.ClearWeeklyLimitUsd()
	}
	if plan.MonthlyLimitUSD != nil {
		builder = builder.SetMonthlyLimitUsd(*plan.MonthlyLimitUSD)
	} else {
		builder = builder.ClearMonthlyLimitUsd()
	}
	if plan.Price != nil {
		builder = builder.SetPrice(*plan.Price)
	} else {
		builder = builder.ClearPrice()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrPlanNotFound, nil)
	}
	plan.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *subscriptionPlanRepository) Delete(ctx context.Context, id int64) error {
	affected, err := r.client.SubscriptionPlan.Update().
		Where(
			subscriptionplan.ID(id),
			subscriptionplan.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrPlanNotFound
	}
	return nil
}

// nilIfEmpty returns nil if s is empty, otherwise returns a pointer to s.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
