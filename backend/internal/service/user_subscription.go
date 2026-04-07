package service

import "time"

type UserSubscription struct {
	ID     int64
	UserID int64
	PlanID int64

	StartsAt  time.Time
	ExpiresAt time.Time
	Status    string

	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time

	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64

	AssignedBy *int64
	AssignedAt time.Time
	Notes      string

	CreatedAt time.Time
	UpdatedAt time.Time

	User           *User
	Plan           *SubscriptionPlan
	AssignedByUser *User
}

func (s *UserSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.ExpiresAt)
}

func (s *UserSubscription) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSubscription) DaysRemaining() int {
	if s.IsExpired() {
		return 0
	}
	return int(time.Until(s.ExpiresAt).Hours() / 24)
}

// NeedsWindowActivation 返回 true 表示至少有一个窗口尚未激活（window start 为 nil）。
// 与 NeedsAnyReset 独立：一个订阅可能同时需要激活某些 window 和重置另一些 window。
func (s *UserSubscription) NeedsWindowActivation() bool {
	return s.DailyWindowStart == nil || s.WeeklyWindowStart == nil || s.MonthlyWindowStart == nil
}

// NeedsAnyReset 返回 true 表示至少有一个窗口已过期需要重置。
func (s *UserSubscription) NeedsAnyReset() bool {
	return s.NeedsDailyReset() || s.NeedsWeeklyReset() || s.NeedsMonthlyReset()
}

func (s *UserSubscription) NeedsDailyReset() bool {
	if s.DailyWindowStart == nil {
		return false
	}
	return time.Since(*s.DailyWindowStart) >= 24*time.Hour
}

func (s *UserSubscription) NeedsWeeklyReset() bool {
	if s.WeeklyWindowStart == nil {
		return false
	}
	return time.Since(*s.WeeklyWindowStart) >= 7*24*time.Hour
}

func (s *UserSubscription) NeedsMonthlyReset() bool {
	if s.MonthlyWindowStart == nil {
		return false
	}
	return time.Since(*s.MonthlyWindowStart) >= 30*24*time.Hour
}

func (s *UserSubscription) DailyResetTime() *time.Time {
	if s.DailyWindowStart == nil {
		return nil
	}
	t := s.DailyWindowStart.Add(24 * time.Hour)
	return &t
}

func (s *UserSubscription) WeeklyResetTime() *time.Time {
	if s.WeeklyWindowStart == nil {
		return nil
	}
	t := s.WeeklyWindowStart.Add(7 * 24 * time.Hour)
	return &t
}

func (s *UserSubscription) MonthlyResetTime() *time.Time {
	if s.MonthlyWindowStart == nil {
		return nil
	}
	t := s.MonthlyWindowStart.Add(30 * 24 * time.Hour)
	return &t
}

// CheckDailyLimit 检查日额度（限额来自计划）
func (s *UserSubscription) CheckDailyLimit(plan *SubscriptionPlan, additionalCost float64) bool {
	if !plan.HasDailyLimit() {
		return true
	}
	return s.DailyUsageUSD+additionalCost <= *plan.DailyLimitUSD
}

// CheckWeeklyLimit 检查周额度（限额来自计划）
func (s *UserSubscription) CheckWeeklyLimit(plan *SubscriptionPlan, additionalCost float64) bool {
	if !plan.HasWeeklyLimit() {
		return true
	}
	return s.WeeklyUsageUSD+additionalCost <= *plan.WeeklyLimitUSD
}

// CheckMonthlyLimit 检查月额度（限额来自计划）
func (s *UserSubscription) CheckMonthlyLimit(plan *SubscriptionPlan, additionalCost float64) bool {
	if !plan.HasMonthlyLimit() {
		return true
	}
	return s.MonthlyUsageUSD+additionalCost <= *plan.MonthlyLimitUSD
}

func (s *UserSubscription) CheckAllLimits(plan *SubscriptionPlan, additionalCost float64) (daily, weekly, monthly bool) {
	daily = s.CheckDailyLimit(plan, additionalCost)
	weekly = s.CheckWeeklyLimit(plan, additionalCost)
	monthly = s.CheckMonthlyLimit(plan, additionalCost)
	return
}
