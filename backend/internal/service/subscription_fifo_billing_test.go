//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── helpers ────────────────────────────────────────────────────────────────

// makePlan constructs a SubscriptionPlan with the given ID and limits.
// Passing distinct IDs prevents tests from accidentally sharing plan identity.
func makePlan(id int64, daily, weekly, monthly *float64) *SubscriptionPlan {
	return &SubscriptionPlan{
		ID:              id,
		Status:          StatusActive,
		DailyLimitUSD:   daily,
		WeeklyLimitUSD:  weekly,
		MonthlyLimitUSD: monthly,
	}
}

func activeSub(id int64, expiresIn time.Duration, plan *SubscriptionPlan) UserSubscription {
	now := time.Now()
	ws := now.Add(-time.Hour) // window started 1 hour ago
	return UserSubscription{
		ID:                 id,
		UserID:             1,
		PlanID:             plan.ID,
		Status:             SubscriptionStatusActive,
		ExpiresAt:          now.Add(expiresIn),
		DailyWindowStart:   ptrTime(ws),
		WeeklyWindowStart:  ptrTime(ws),
		MonthlyWindowStart: ptrTime(ws),
		Plan:               plan,
	}
}

// ─── fifoIncrementUsage stub ─────────────────────────────────────────────────

type fifoRepoStub struct {
	userSubRepoNoop
	calls   []fifoCall
	failIDs map[int64]bool // IDs to fail on IncrementUsage
	// usageByID provides per-ID current usage for GetCurrentUsage.
	// If nil or key absent, falls back to the usage already in the queue element
	// (which means the test exercises the snapshot path, triggering the "Warn" branch).
	usageByID map[int64][3]float64 // [daily, weekly, monthly]
}

type fifoCall struct {
	id     int64
	charge float64
}

func (r *fifoRepoStub) IncrementUsage(_ context.Context, id int64, charge float64) error {
	if r.failIDs[id] {
		return errors.New("db error")
	}
	r.calls = append(r.calls, fifoCall{id: id, charge: charge})
	return nil
}

// GetCurrentUsage returns per-ID usage from usageByID if present;
// otherwise returns (0,0,0,ErrSubscriptionNotFound) to trigger the fallback path.
func (r *fifoRepoStub) GetCurrentUsage(_ context.Context, id int64) (float64, float64, float64, error) {
	if r.usageByID != nil {
		if u, ok := r.usageByID[id]; ok {
			return u[0], u[1], u[2], nil
		}
	}
	return 0, 0, 0, ErrSubscriptionNotFound
}

// ─── fifoIncrementUsage tests ────────────────────────────────────────────────

func TestFifoIncrementUsage_EmptyQueue(t *testing.T) {
	repo := &fifoRepoStub{}
	fifoIncrementUsage(context.Background(), repo, nil, 10.0)
	assert.Empty(t, repo.calls, "empty queue: no DB writes expected")
}

func TestFifoIncrementUsage_SingleSub_FullyFits(t *testing.T) {
	plan := makePlan(1, ptrFloat(100), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyUsageUSD = 50

	repo := &fifoRepoStub{}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{sub}, 30)

	require.Len(t, repo.calls, 1)
	assert.Equal(t, int64(1), repo.calls[0].id)
	assert.InDelta(t, 30.0, repo.calls[0].charge, 1e-9)
}

// Single subscription is always "last" → absorbs all regardless of limits
func TestFifoIncrementUsage_SingleSub_ExceedsLimit(t *testing.T) {
	plan := makePlan(1, ptrFloat(10), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyUsageUSD = 8

	repo := &fifoRepoStub{}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{sub}, 50)

	require.Len(t, repo.calls, 1)
	assert.InDelta(t, 50.0, repo.calls[0].charge, 1e-9, "last sub absorbs full remaining")
}

func TestFifoIncrementUsage_TwoSubs_FillsFirstThenOverflow(t *testing.T) {
	// A: daily 10, used 8 → 2 remaining
	// B: daily 100 → absorb rest
	planA := makePlan(1, ptrFloat(10), nil, nil)
	planB := makePlan(2, ptrFloat(100), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subA.DailyUsageUSD = 8
	subB := activeSub(2, 48*time.Hour, planB)

	repo := &fifoRepoStub{
		usageByID: map[int64][3]float64{1: {8, 8, 8}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 20)

	require.Len(t, repo.calls, 2)
	assert.Equal(t, int64(1), repo.calls[0].id)
	assert.InDelta(t, 2.0, repo.calls[0].charge, 1e-9, "A gets its 2 remaining")
	assert.Equal(t, int64(2), repo.calls[1].id)
	assert.InDelta(t, 18.0, repo.calls[1].charge, 1e-9, "B gets overflow 18")
}

func TestFifoIncrementUsage_FirstSubFull_SkipsToSecond(t *testing.T) {
	// A: daily 10, already used 10 → capacity=0 → charge=0 → continue
	// B: is last → charge = remaining
	planA := makePlan(1, ptrFloat(10), nil, nil)
	planB := makePlan(2, ptrFloat(50), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subA.DailyUsageUSD = 10
	subB := activeSub(2, 48*time.Hour, planB)

	repo := &fifoRepoStub{
		usageByID: map[int64][3]float64{1: {10, 10, 10}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 15)

	require.Len(t, repo.calls, 1, "A is full, only B gets charged")
	assert.Equal(t, int64(2), repo.calls[0].id)
	assert.InDelta(t, 15.0, repo.calls[0].charge, 1e-9)
}

func TestFifoIncrementUsage_CapacityMinOfDailyWeeklyMonthly(t *testing.T) {
	// daily remaining 5, weekly remaining 3, monthly remaining 10 → capacity = 3
	plan := makePlan(1, ptrFloat(10), ptrFloat(8), ptrFloat(20))
	subA := activeSub(1, 24*time.Hour, plan)
	subA.DailyUsageUSD = 5    // 10-5=5 daily remaining
	subA.WeeklyUsageUSD = 5   // 8-5=3 weekly remaining ← min
	subA.MonthlyUsageUSD = 10 // 20-10=10 monthly remaining
	subB := activeSub(2, 48*time.Hour, makePlan(2, nil, nil, nil))

	repo := &fifoRepoStub{
		usageByID: map[int64][3]float64{1: {5, 5, 10}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 10)

	require.Len(t, repo.calls, 2)
	assert.InDelta(t, 3.0, repo.calls[0].charge, 1e-9, "A gets min(5,3,10)=3")
	assert.InDelta(t, 7.0, repo.calls[1].charge, 1e-9, "B gets remainder 7")
}

func TestFifoIncrementUsage_DBFailure_ContinuesToNext(t *testing.T) {
	// cost=30, A capacity=50 → charge=30, DB fails → remaining stays 30
	// B is last → charge=remaining=30
	planA := makePlan(1, ptrFloat(50), nil, nil)
	planB := makePlan(2, ptrFloat(50), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subB := activeSub(2, 48*time.Hour, planB)

	repo := &fifoRepoStub{
		failIDs:   map[int64]bool{1: true},
		usageByID: map[int64][3]float64{1: {0, 0, 0}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 30)

	require.Len(t, repo.calls, 1, "only B succeeds")
	assert.Equal(t, int64(2), repo.calls[0].id)
	assert.InDelta(t, 30.0, repo.calls[0].charge, 1e-9)
}

// TestFifoIncrementUsage_DBFailure_RemainingNotReduced verifies the "retry on next"
// semantics when a DB write fails. Without failure: A absorbs 50, B absorbs 10.
// With failure on A: remaining is NOT decremented (50 is not subtracted), so B
// (as the last subscriber) absorbs the full remaining 60 instead of 10. This is
// intentional — we'd rather over-charge B than silently lose the cost record.
func TestFifoIncrementUsage_DBFailure_RemainingNotReduced(t *testing.T) {
	planA := makePlan(1, ptrFloat(50), nil, nil)
	planB := makePlan(2, ptrFloat(100), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subB := activeSub(2, 48*time.Hour, planB)

	repo := &fifoRepoStub{
		failIDs:   map[int64]bool{1: true},
		usageByID: map[int64][3]float64{1: {0, 0, 0}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 60)

	require.Len(t, repo.calls, 1)
	assert.Equal(t, int64(2), repo.calls[0].id)
	// B absorbs 60 (full remaining), not 10 (the overflow without failure)
	assert.InDelta(t, 60.0, repo.calls[0].charge, 1e-9)
}

func TestFifoIncrementUsage_AllDBFail_RemainingNonZero(t *testing.T) {
	plan := makePlan(1, ptrFloat(50), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)

	repo := &fifoRepoStub{failIDs: map[int64]bool{1: true}}
	// no panic, no data corruption — just log
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{sub}, 25)
	assert.Empty(t, repo.calls)
}

func TestFifoIncrementUsage_NilPlan_TreatedAsLast(t *testing.T) {
	// Sub without plan → treated as last (absorbs all)
	subA := UserSubscription{
		ID:        1,
		UserID:    1,
		Status:    SubscriptionStatusActive,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Plan:      nil,
	}

	repo := &fifoRepoStub{}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA}, 42)

	require.Len(t, repo.calls, 1)
	assert.InDelta(t, 42.0, repo.calls[0].charge, 1e-9)
}

// TestFifoIncrementUsage_RealtimeUsageOverridesSnapshot verifies that GetCurrentUsage
// is used instead of the stale snapshot value for capacity calculation.
// Snapshot says A used 2 (→ 8 remaining), but DB says A used 9 (→ 1 remaining).
// The charge to A must be based on the real-time value (1), not the snapshot (8).
func TestFifoIncrementUsage_RealtimeUsageOverridesSnapshot(t *testing.T) {
	planA := makePlan(1, ptrFloat(10), nil, nil)
	planB := makePlan(2, ptrFloat(100), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subA.DailyUsageUSD = 2 // stale snapshot: 8 remaining
	subB := activeSub(2, 48*time.Hour, planB)

	repo := &fifoRepoStub{
		// real-time DB value: A has used 9 → only 1 remaining
		usageByID: map[int64][3]float64{1: {9, 9, 9}},
	}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 20)

	require.Len(t, repo.calls, 2)
	assert.Equal(t, int64(1), repo.calls[0].id)
	assert.InDelta(t, 1.0, repo.calls[0].charge, 1e-9, "A gets real-time remaining=1, not snapshot remaining=8")
	assert.Equal(t, int64(2), repo.calls[1].id)
	assert.InDelta(t, 19.0, repo.calls[1].charge, 1e-9, "B absorbs the rest")
}

// TestFifoIncrementUsage_GetCurrentUsageFails_SkipsSubscription verifies that when
// GetCurrentUsage returns an error, the subscription is skipped (not charged with
// stale snapshot data), and the cost flows to the next subscription.
func TestFifoIncrementUsage_GetCurrentUsageFails_SkipsSubscription(t *testing.T) {
	planA := makePlan(1, ptrFloat(50), nil, nil)
	planB := makePlan(2, ptrFloat(100), nil, nil)
	subA := activeSub(1, 24*time.Hour, planA)
	subA.DailyUsageUSD = 0 // snapshot says plenty of room, but GetCurrentUsage will fail
	subB := activeSub(2, 48*time.Hour, planB)

	// usageByID is nil → GetCurrentUsage returns ErrSubscriptionNotFound for id=1
	repo := &fifoRepoStub{}
	fifoIncrementUsage(context.Background(), repo, []UserSubscription{subA, subB}, 30)

	require.Len(t, repo.calls, 1, "A skipped due to GetCurrentUsage failure, only B charged")
	assert.Equal(t, int64(2), repo.calls[0].id)
	assert.InDelta(t, 30.0, repo.calls[0].charge, 1e-9)
}

// ─── mergeSubscriptions tests ────────────────────────────────────────────────

func TestMergeSubscriptions_SortsByExpiry(t *testing.T) {
	plan := makePlan(1, ptrFloat(10), nil, nil)

	// deliberately out of order
	sub1 := activeSub(1, 72*time.Hour, plan) // expires latest
	sub2 := activeSub(2, 24*time.Hour, plan) // expires soonest
	sub3 := activeSub(3, 48*time.Hour, plan)

	state := mergeSubscriptions([]UserSubscription{sub1, sub2, sub3})

	require.Len(t, state.FIFOQueue, 3)
	assert.Equal(t, int64(2), state.FIFOQueue[0].ID, "soonest first")
	assert.Equal(t, int64(3), state.FIFOQueue[1].ID)
	assert.Equal(t, int64(1), state.FIFOQueue[2].ID)
}

func TestMergeSubscriptions_FiltersInactiveAndNilPlan(t *testing.T) {
	plan := makePlan(1, ptrFloat(10), nil, nil)

	active := activeSub(1, 24*time.Hour, plan)

	expired := activeSub(2, -1*time.Hour, plan)
	expired.Status = SubscriptionStatusExpired

	// construct directly — activeSub panics on nil plan
	noPlan := UserSubscription{
		ID:        3,
		UserID:    1,
		Status:    SubscriptionStatusActive,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Plan:      nil,
	}

	state := mergeSubscriptions([]UserSubscription{active, expired, noPlan})

	require.Len(t, state.FIFOQueue, 1)
	assert.Equal(t, int64(1), state.FIFOQueue[0].ID)
}

func TestMergeSubscriptions_AggregatesLimits(t *testing.T) {
	planA := makePlan(1, ptrFloat(10), ptrFloat(50), nil)
	planB := makePlan(2, ptrFloat(20), ptrFloat(100), nil)

	subA := activeSub(1, 24*time.Hour, planA)
	subB := activeSub(2, 48*time.Hour, planB)

	state := mergeSubscriptions([]UserSubscription{subA, subB})

	require.NotNil(t, state.EffectiveDailyLimit)
	assert.InDelta(t, 30.0, *state.EffectiveDailyLimit, 1e-9, "10+20")
	require.NotNil(t, state.EffectiveWeeklyLimit)
	assert.InDelta(t, 150.0, *state.EffectiveWeeklyLimit, 1e-9, "50+100")
	assert.Nil(t, state.EffectiveMonthlyLimit, "both unlimited → nil")
}

func TestMergeSubscriptions_UnlimitedPlanMakesLimitNil(t *testing.T) {
	planLimited := makePlan(1, ptrFloat(10), nil, nil)
	planUnlimited := makePlan(2, nil, nil, nil)

	subA := activeSub(1, 24*time.Hour, planLimited)
	subB := activeSub(2, 48*time.Hour, planUnlimited)

	state := mergeSubscriptions([]UserSubscription{subA, subB})

	assert.Nil(t, state.EffectiveDailyLimit, "one unlimited plan → merged limit is nil")
}

func TestMergeSubscriptions_AggregatesUsage(t *testing.T) {
	planA := makePlan(1, ptrFloat(100), ptrFloat(500), ptrFloat(2000))
	planB := makePlan(2, ptrFloat(100), ptrFloat(500), ptrFloat(2000))
	subA := activeSub(1, 24*time.Hour, planA)
	subA.DailyUsageUSD = 3
	subA.WeeklyUsageUSD = 10
	subA.MonthlyUsageUSD = 40

	subB := activeSub(2, 48*time.Hour, planB)
	subB.DailyUsageUSD = 7
	subB.WeeklyUsageUSD = 20
	subB.MonthlyUsageUSD = 60

	state := mergeSubscriptions([]UserSubscription{subA, subB})

	assert.InDelta(t, 10.0, state.TotalDailyUsage, 1e-9)
	assert.InDelta(t, 30.0, state.TotalWeeklyUsage, 1e-9)
	assert.InDelta(t, 100.0, state.TotalMonthlyUsage, 1e-9)
}

func TestMergeSubscriptions_Empty(t *testing.T) {
	state := mergeSubscriptions(nil)
	assert.Empty(t, state.FIFOQueue)
	assert.Nil(t, state.EffectiveDailyLimit)
	assert.Zero(t, state.TotalDailyUsage)
}

// ─── ValidateMergedState tests ───────────────────────────────────────────────

func newValidateService() *SubscriptionService {
	return NewSubscriptionService(nil, &userSubRepoNoop{}, nil, nil, nil, nil)
}

func makeActiveState(subs ...UserSubscription) *MergedSubscriptionState {
	return mergeSubscriptions(subs)
}

func TestValidateMergedState_NilState(t *testing.T) {
	svc := newValidateService()
	_, err := svc.ValidateMergedState(nil)
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestValidateMergedState_EmptyQueue(t *testing.T) {
	svc := newValidateService()
	state := &MergedSubscriptionState{}
	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestValidateMergedState_ExpiredStatus(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(100), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.Status = SubscriptionStatusExpired

	state := &MergedSubscriptionState{
		FIFOQueue:           []UserSubscription{sub},
		EffectiveDailyLimit: ptrFloat(100),
	}
	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrSubscriptionExpired)
}

func TestValidateMergedState_ExpiredByTime(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(100), nil, nil)
	sub := activeSub(1, -1*time.Second, plan)
	sub.Status = SubscriptionStatusActive // status field active, but ExpiresAt is past

	state := &MergedSubscriptionState{
		FIFOQueue:           []UserSubscription{sub},
		EffectiveDailyLimit: ptrFloat(100),
	}
	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrSubscriptionExpired)
}

func TestValidateMergedState_Suspended(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(100), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.Status = SubscriptionStatusSuspended

	state := &MergedSubscriptionState{
		FIFOQueue:           []UserSubscription{sub},
		EffectiveDailyLimit: ptrFloat(100),
	}
	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrSubscriptionSuspended)
}

func TestValidateMergedState_DailyLimitExceeded(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(10), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyUsageUSD = 10 // exactly at limit

	state := makeActiveState(sub)

	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrDailyLimitExceeded)
}

func TestValidateMergedState_WeeklyLimitExceeded(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, nil, ptrFloat(50), nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.WeeklyUsageUSD = 51

	state := makeActiveState(sub)

	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrWeeklyLimitExceeded)
}

func TestValidateMergedState_MonthlyLimitExceeded(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, nil, nil, ptrFloat(200))
	sub := activeSub(1, 24*time.Hour, plan)
	sub.MonthlyUsageUSD = 200

	state := makeActiveState(sub)

	_, err := svc.ValidateMergedState(state)
	require.ErrorIs(t, err, ErrMonthlyLimitExceeded)
}

func TestValidateMergedState_UnlimitedPlan_NoError(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, nil, nil, nil) // unlimited
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyUsageUSD = 9999
	sub.WeeklyUsageUSD = 9999
	sub.MonthlyUsageUSD = 9999

	state := makeActiveState(sub)

	_, err := svc.ValidateMergedState(state)
	require.NoError(t, err, "unlimited plan should never trigger limit errors")
}

func TestValidateMergedState_NeedsWindowActivation_SetsMaintenance(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(100), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyWindowStart = nil // window not yet activated

	state := makeActiveState(sub)

	needsMaint, err := svc.ValidateMergedState(state)
	require.NoError(t, err)
	assert.True(t, needsMaint, "unactivated window should flag maintenance")
}

func TestValidateMergedState_NeedsDailyReset_ZerosUsageAndSetsMaintenance(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(10), nil, nil)
	sub := activeSub(1, 24*time.Hour, plan)
	// window started 25 hours ago → NeedsDailyReset = true
	ws := time.Now().Add(-25 * time.Hour)
	sub.DailyWindowStart = &ws
	sub.DailyUsageUSD = 8 // would exceed limit of 10, but reset zeroes it first

	state := makeActiveState(sub)

	needsMaint, err := svc.ValidateMergedState(state)
	require.NoError(t, err, "after reset usage = 0, should not exceed limit")
	assert.True(t, needsMaint)
	assert.Zero(t, state.TotalDailyUsage, "recalcUsage should reflect zeroed usage")
}

func TestValidateMergedState_ValidState_NoError(t *testing.T) {
	svc := newValidateService()
	plan := makePlan(1, ptrFloat(100), ptrFloat(500), ptrFloat(2000))
	sub := activeSub(1, 24*time.Hour, plan)
	sub.DailyUsageUSD = 50
	sub.WeeklyUsageUSD = 200
	sub.MonthlyUsageUSD = 800

	state := makeActiveState(sub)

	needsMaint, err := svc.ValidateMergedState(state)
	require.NoError(t, err)
	assert.False(t, needsMaint)
}
