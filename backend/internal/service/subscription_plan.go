package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// SubscriptionPlanRepository 订阅计划仓储接口
type SubscriptionPlanRepository interface {
	GetByID(ctx context.Context, id int64) (*SubscriptionPlan, error)
	List(ctx context.Context, params pagination.PaginationParams, visibility, status string) ([]SubscriptionPlan, *pagination.PaginationResult, error)
	ListAll(ctx context.Context) ([]SubscriptionPlan, error)
	Create(ctx context.Context, plan *SubscriptionPlan) error
	Update(ctx context.Context, plan *SubscriptionPlan) error
	Delete(ctx context.Context, id int64) error
}

// SubscriptionPlan 定义订阅计划模板
type SubscriptionPlan struct {
	ID          int64
	Name        string
	Description string
	Visibility  string // "public" | "private"
	Status      string

	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64

	DefaultValidityDays int
	Price               *float64

	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *SubscriptionPlan) IsActive() bool {
	return p.Status == StatusActive
}

func (p *SubscriptionPlan) IsPublic() bool {
	return p.Visibility == VisibilityPublic
}

func (p *SubscriptionPlan) HasDailyLimit() bool {
	return p.DailyLimitUSD != nil && *p.DailyLimitUSD > 0
}

func (p *SubscriptionPlan) HasWeeklyLimit() bool {
	return p.WeeklyLimitUSD != nil && *p.WeeklyLimitUSD > 0
}

func (p *SubscriptionPlan) HasMonthlyLimit() bool {
	return p.MonthlyLimitUSD != nil && *p.MonthlyLimitUSD > 0
}
