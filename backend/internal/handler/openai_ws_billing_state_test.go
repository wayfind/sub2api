package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type openAIWSTurnSubscriptionRepo struct {
	service.UserSubscriptionRepository
	listActive func(context.Context, int64) ([]service.UserSubscription, error)
}

func (r *openAIWSTurnSubscriptionRepo) ListActiveByUserID(ctx context.Context, userID int64) ([]service.UserSubscription, error) {
	return r.listActive(ctx, userID)
}

type openAIWSTurnBillingCache struct {
	service.BillingCache
	balance          float64
	subscriptionData *service.SubscriptionCacheData
}

func (c *openAIWSTurnBillingCache) GetUserBalance(context.Context, int64) (float64, error) {
	return c.balance, nil
}

func (c *openAIWSTurnBillingCache) GetSubscriptionCache(context.Context, int64, int64) (*service.SubscriptionCacheData, error) {
	return c.subscriptionData, nil
}

func TestResolveOpenAIWSTurnBillingState(t *testing.T) {
	now := time.Now()
	dailyLimit := 10.0
	windowStart := now.Add(-time.Hour)
	sub := service.UserSubscription{
		ID:                 7,
		UserID:             42,
		PlanID:             3,
		StartsAt:           now.Add(-time.Hour),
		ExpiresAt:          now.Add(time.Hour),
		Status:             service.SubscriptionStatusActive,
		DailyWindowStart:   &windowStart,
		WeeklyWindowStart:  &windowStart,
		MonthlyWindowStart: &windowStart,
		DailyUsageUSD:      5,
		Plan:               &service.SubscriptionPlan{ID: 3, DailyLimitUSD: &dailyLimit},
	}

	var listResult = []service.UserSubscription{sub}
	var listErr error
	repo := &openAIWSTurnSubscriptionRepo{
		listActive: func(context.Context, int64) ([]service.UserSubscription, error) {
			return listResult, listErr
		},
	}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	subscriptionService := service.NewSubscriptionService(nil, repo, nil, nil, nil, nil)
	t.Cleanup(subscriptionService.Stop)
	billingCache := &openAIWSTurnBillingCache{
		balance: 50,
		subscriptionData: &service.SubscriptionCacheData{
			Status:    service.SubscriptionStatusActive,
			ExpiresAt: now.Add(time.Hour),
		},
	}
	billingCacheService := service.NewBillingCacheService(billingCache, nil, nil, nil, cfg)
	t.Cleanup(billingCacheService.Stop)
	h := &OpenAIGatewayHandler{
		subscriptionService: subscriptionService,
		billingCacheService: billingCacheService,
	}
	apiKey := &service.APIKey{User: &service.User{ID: 42}}

	t.Run("active quota uses subscription", func(t *testing.T) {
		gotSub, gotState, inPeriod, err := h.resolveOpenAIWSTurnBillingState(context.Background(), apiKey)
		require.NoError(t, err)
		require.NotNil(t, gotSub)
		require.Equal(t, int64(7), gotSub.ID)
		require.NotNil(t, gotState)
		require.True(t, inPeriod)
	})

	t.Run("exhausted quota falls back to discounted balance", func(t *testing.T) {
		exhausted := sub
		exhausted.DailyUsageUSD = dailyLimit
		listResult = []service.UserSubscription{exhausted}

		gotSub, gotState, inPeriod, err := h.resolveOpenAIWSTurnBillingState(context.Background(), apiKey)
		require.NoError(t, err)
		require.Nil(t, gotSub)
		require.Nil(t, gotState)
		require.True(t, inPeriod)
	})

	t.Run("no active period falls back to full-rate balance", func(t *testing.T) {
		listResult = nil

		gotSub, gotState, inPeriod, err := h.resolveOpenAIWSTurnBillingState(context.Background(), apiKey)
		require.NoError(t, err)
		require.Nil(t, gotSub)
		require.Nil(t, gotState)
		require.False(t, inPeriod)
	})

	t.Run("lookup failure fails closed as service unavailable", func(t *testing.T) {
		listErr = errors.New("database unavailable")

		gotSub, gotState, inPeriod, err := h.resolveOpenAIWSTurnBillingState(context.Background(), apiKey)
		require.Nil(t, gotSub)
		require.Nil(t, gotState)
		require.False(t, inPeriod)
		require.ErrorIs(t, err, service.ErrBillingServiceUnavailable)
	})
}
