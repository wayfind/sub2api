/**
 * User Subscription API
 * API for regular users to view their own subscriptions and progress
 */

import { apiClient } from './client'
import type { UserSubscription, SubscriptionProgress } from '@/types'

/**
 * Subscription summary for user dashboard
 */
export interface SubscriptionSummary {
  active_count: number
  subscriptions: Array<{
    id: number
    group_name: string
    status: string
    daily_progress: number | null
    weekly_progress: number | null
    monthly_progress: number | null
    expires_at: string | null
    days_remaining: number | null
  }>
}

/**
 * Get list of current user's subscriptions
 */
export async function getMySubscriptions(): Promise<UserSubscription[]> {
  const response = await apiClient.get<UserSubscription[]>('/subscriptions')
  return response.data
}

/**
 * Get current user's active subscriptions
 */
export async function getActiveSubscriptions(): Promise<UserSubscription[]> {
  const response = await apiClient.get<UserSubscription[]>('/subscriptions/active')
  return response.data
}

/**
 * Get progress for all user's active subscriptions
 */
export async function getSubscriptionsProgress(): Promise<SubscriptionProgress[]> {
  const response = await apiClient.get<SubscriptionProgress[]>('/subscriptions/progress')
  return response.data
}

/**
 * Get subscription summary for dashboard display
 */
export async function getSubscriptionSummary(): Promise<SubscriptionSummary> {
  const response = await apiClient.get<SubscriptionSummary>('/subscriptions/summary')
  return response.data
}

/**
 * Get progress for a specific subscription
 */
export async function getSubscriptionProgress(
  subscriptionId: number
): Promise<SubscriptionProgress> {
  const response = await apiClient.get<SubscriptionProgress>(
    `/subscriptions/${subscriptionId}/progress`
  )
  return response.data
}

/**
 * Purchase a subscription plan using account balance
 */
export async function purchaseSubscription(planId: number): Promise<UserSubscription> {
  const response = await apiClient.post<UserSubscription>('/subscriptions/purchase', {
    plan_id: planId
  })
  return response.data
}

/**
 * Merged subscription state - combined limits across all active subscriptions
 */
export interface MergedSubscriptionState {
  has_active: boolean
  active_count: number
  daily_limit_usd: number | null
  daily_used_usd: number
  weekly_limit_usd: number | null
  weekly_used_usd: number
  monthly_limit_usd: number | null
  monthly_used_usd: number
}

/**
 * Get merged (combined) subscription state across all active subscriptions
 */
export async function getMergedSubscription(): Promise<MergedSubscriptionState> {
  const response = await apiClient.get<MergedSubscriptionState>('/subscriptions/merged')
  return response.data
}

export default {
  getMySubscriptions,
  getActiveSubscriptions,
  getSubscriptionsProgress,
  getSubscriptionSummary,
  getSubscriptionProgress,
  getMergedSubscription,
  purchaseSubscription
}
