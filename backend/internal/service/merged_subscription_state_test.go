package service

import (
	"testing"
	"time"
)

func TestMergedSubscriptionState_ActivePlanIDs(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	futureStart := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	t.Run("nil state", func(t *testing.T) {
		var s *MergedSubscriptionState
		if got := s.ActivePlanIDs(); got != nil {
			t.Fatalf("nil state ActivePlanIDs = %v, want nil", got)
		}
	})

	t.Run("filters expired and dedups", func(t *testing.T) {
		s := &MergedSubscriptionState{
			FIFOQueue: []UserSubscription{
				{PlanID: 1, ExpiresAt: future},
				{PlanID: 2, ExpiresAt: past},   // expired -> excluded
				{PlanID: 1, ExpiresAt: future}, // duplicate plan -> deduped
				{PlanID: 3, ExpiresAt: future},
				{PlanID: 4, StartsAt: futureStart, ExpiresAt: future}, // not started -> excluded
			},
		}
		got := s.ActivePlanIDs()
		want := map[int64]bool{1: true, 3: true}
		if len(got) != len(want) {
			t.Fatalf("ActivePlanIDs = %v, want plans %v", got, want)
		}
		for _, id := range got {
			if !want[id] {
				t.Fatalf("ActivePlanIDs returned unexpected plan %d (got %v)", id, got)
			}
		}
	})

	t.Run("all expired yields empty", func(t *testing.T) {
		s := &MergedSubscriptionState{
			FIFOQueue: []UserSubscription{
				{PlanID: 1, ExpiresAt: past},
				{PlanID: 2, ExpiresAt: past},
			},
		}
		if got := s.ActivePlanIDs(); len(got) != 0 {
			t.Fatalf("ActivePlanIDs = %v, want empty", got)
		}
	})
}

func TestUserSubscription_IsActiveAtRequiresCurrentPeriod(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		sub  UserSubscription
		want bool
	}{
		{
			name: "inside period",
			sub: UserSubscription{
				Status:    SubscriptionStatusActive,
				StartsAt:  now.Add(-time.Hour),
				ExpiresAt: now.Add(time.Hour),
			},
			want: true,
		},
		{
			name: "not started",
			sub: UserSubscription{
				Status:    SubscriptionStatusActive,
				StartsAt:  now.Add(time.Hour),
				ExpiresAt: now.Add(2 * time.Hour),
			},
			want: false,
		},
		{
			name: "expired",
			sub: UserSubscription{
				Status:    SubscriptionStatusActive,
				StartsAt:  now.Add(-2 * time.Hour),
				ExpiresAt: now.Add(-time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sub.IsActiveAt(now); got != tt.want {
				t.Fatalf("IsActiveAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriptionService_CopyMergedStateRecalculatesAfterExpiry(t *testing.T) {
	now := time.Now()
	expiredLimit := 10.0
	activeLimit := 20.0
	totalLimit := expiredLimit + activeLimit
	src := &MergedSubscriptionState{
		EffectiveDailyLimit: &totalLimit,
		TotalDailyUsage:     11,
		NeedsMaintenance:    true,
		FIFOQueue: []UserSubscription{
			{
				ID:            1,
				Status:        SubscriptionStatusActive,
				StartsAt:      now.Add(-2 * time.Hour),
				ExpiresAt:     now.Add(-time.Minute),
				DailyUsageUSD: 9,
				Plan:          &SubscriptionPlan{DailyLimitUSD: &expiredLimit},
			},
			{
				ID:            2,
				Status:        SubscriptionStatusActive,
				StartsAt:      now.Add(-time.Hour),
				ExpiresAt:     now.Add(time.Hour),
				DailyUsageUSD: 2,
				Plan:          &SubscriptionPlan{DailyLimitUSD: &activeLimit},
			},
		},
	}

	got := (&SubscriptionService{}).copyMergedState(src)
	if len(got.FIFOQueue) != 1 || got.FIFOQueue[0].ID != 2 {
		t.Fatalf("FIFOQueue = %+v, want only active subscription 2", got.FIFOQueue)
	}
	if got.EffectiveDailyLimit == nil || *got.EffectiveDailyLimit != activeLimit {
		t.Fatalf("EffectiveDailyLimit = %v, want %v", got.EffectiveDailyLimit, activeLimit)
	}
	if got.TotalDailyUsage != 2 {
		t.Fatalf("TotalDailyUsage = %v, want 2", got.TotalDailyUsage)
	}
	if !got.NeedsMaintenance {
		t.Fatal("NeedsMaintenance was not preserved")
	}
}
