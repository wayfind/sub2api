package service

import (
	"context"
	"fmt"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// CreateSubscriptionPlanInput 创建订阅计划输入
type CreateSubscriptionPlanInput struct {
	Name                string
	Description         string
	Visibility          string
	DailyLimitUSD       *float64
	WeeklyLimitUSD      *float64
	MonthlyLimitUSD     *float64
	DefaultValidityDays int
	Price               *float64
	SortOrder           int
}

// UpdateSubscriptionPlanInput 更新订阅计划输入（nil 表示不修改）
type UpdateSubscriptionPlanInput struct {
	Name                *string
	Description         *string
	Visibility          *string
	Status              *string
	DailyLimitUSD       *float64
	WeeklyLimitUSD      *float64
	MonthlyLimitUSD     *float64
	DefaultValidityDays *int
	Price               *float64
	SortOrder           *int
}

// SubscriptionPlanService 订阅计划管理服务
type SubscriptionPlanService struct {
	planRepo    SubscriptionPlanRepository
	userSubRepo UserSubscriptionRepository
}

// NewSubscriptionPlanService creates a new SubscriptionPlanService.
func NewSubscriptionPlanService(planRepo SubscriptionPlanRepository, userSubRepo UserSubscriptionRepository) *SubscriptionPlanService {
	return &SubscriptionPlanService{planRepo: planRepo, userSubRepo: userSubRepo}
}

// GetByID 获取订阅计划详情
func (s *SubscriptionPlanService) GetByID(ctx context.Context, id int64) (*SubscriptionPlan, error) {
	return s.planRepo.GetByID(ctx, id)
}

// List 分页查询订阅计划
func (s *SubscriptionPlanService) List(ctx context.Context, page, pageSize int, visibility, status string) ([]SubscriptionPlan, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	return s.planRepo.List(ctx, params, visibility, status)
}

// ListAll 返回所有活跃计划（不分页，用于下拉选择）
func (s *SubscriptionPlanService) ListAll(ctx context.Context) ([]SubscriptionPlan, error) {
	return s.planRepo.ListAll(ctx)
}

// Create 创建订阅计划
func (s *SubscriptionPlanService) Create(ctx context.Context, input *CreateSubscriptionPlanInput) (*SubscriptionPlan, error) {
	// 输入验证
	if input.Name == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "plan name is required")
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = VisibilityPublic
	}
	if visibility != VisibilityPublic && visibility != VisibilityPrivate {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "visibility must be 'public' or 'private'")
	}

	if input.DailyLimitUSD != nil && *input.DailyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "daily_limit_usd must be non-negative")
	}
	if input.WeeklyLimitUSD != nil && *input.WeeklyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "weekly_limit_usd must be non-negative")
	}
	if input.MonthlyLimitUSD != nil && *input.MonthlyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "monthly_limit_usd must be non-negative")
	}

	defaultValidityDays := input.DefaultValidityDays
	if defaultValidityDays <= 0 {
		defaultValidityDays = 30
	}

	plan := &SubscriptionPlan{
		Name:                input.Name,
		Description:         input.Description,
		Visibility:          visibility,
		DailyLimitUSD:       input.DailyLimitUSD,
		WeeklyLimitUSD:      input.WeeklyLimitUSD,
		MonthlyLimitUSD:     input.MonthlyLimitUSD,
		DefaultValidityDays: defaultValidityDays,
		Price:               input.Price,
		SortOrder:           input.SortOrder,
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// Update 更新订阅计划（部分更新语义）
func (s *SubscriptionPlanService) Update(ctx context.Context, id int64, input *UpdateSubscriptionPlanInput) (*SubscriptionPlan, error) {
	// 输入验证
	if input.Visibility != nil && *input.Visibility != VisibilityPublic && *input.Visibility != VisibilityPrivate {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "visibility must be 'public' or 'private'")
	}
	if input.DailyLimitUSD != nil && *input.DailyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "daily_limit_usd must be non-negative")
	}
	if input.WeeklyLimitUSD != nil && *input.WeeklyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "weekly_limit_usd must be non-negative")
	}
	if input.MonthlyLimitUSD != nil && *input.MonthlyLimitUSD < 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "monthly_limit_usd must be non-negative")
	}

	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 仅覆盖非 nil 字段
	if input.Name != nil {
		plan.Name = *input.Name
	}
	if input.Description != nil {
		plan.Description = *input.Description
	}
	if input.Visibility != nil {
		plan.Visibility = *input.Visibility
	}
	if input.Status != nil {
		plan.Status = *input.Status
	}
	if input.DefaultValidityDays != nil {
		plan.DefaultValidityDays = *input.DefaultValidityDays
	}
	if input.SortOrder != nil {
		plan.SortOrder = *input.SortOrder
	}

	// 对于 nullable float 字段，Update request 中的 nil 表示"不修改"。
	// 要清除值需要显式传 0 值指针（前端约定）。
	// 但这里采用简化逻辑：input 中的 field 只要出现在 JSON 中就覆盖。
	// Handler 层通过 request struct 的指针语义控制是否传递。
	if input.DailyLimitUSD != nil {
		plan.DailyLimitUSD = input.DailyLimitUSD
	}
	if input.WeeklyLimitUSD != nil {
		plan.WeeklyLimitUSD = input.WeeklyLimitUSD
	}
	if input.MonthlyLimitUSD != nil {
		plan.MonthlyLimitUSD = input.MonthlyLimitUSD
	}
	if input.Price != nil {
		plan.Price = input.Price
	}

	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// Delete 软删除订阅计划（有活跃订阅时拒绝）
func (s *SubscriptionPlanService) Delete(ctx context.Context, id int64) error {
	activeCount, err := s.userSubRepo.CountActiveByPlanID(ctx, id)
	if err != nil {
		return fmt.Errorf("check active subscriptions: %w", err)
	}
	if activeCount > 0 {
		return ErrPlanHasActiveSubscriptions
	}
	return s.planRepo.Delete(ctx, id)
}
