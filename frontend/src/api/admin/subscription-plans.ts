/**
 * Admin Subscription Plans API endpoints
 * Handles subscription plan management for administrators
 */

import { apiClient } from '../client'
import type {
  SubscriptionPlan,
  CreateSubscriptionPlanRequest,
  UpdateSubscriptionPlanRequest,
  PaginatedResponse
} from '@/types'

/**
 * List all subscription plans with pagination
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: 'active' | 'inactive'
    visibility?: 'public' | 'hidden'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<SubscriptionPlan>> {
  const { data } = await apiClient.get<PaginatedResponse<SubscriptionPlan>>(
    '/admin/subscription-plans',
    {
      params: {
        page,
        page_size: pageSize,
        ...filters
      },
      signal: options?.signal
    }
  )
  return data
}

/**
 * Get all active subscription plans (without pagination)
 */
export async function getAll(): Promise<SubscriptionPlan[]> {
  const { data } = await apiClient.get<SubscriptionPlan[]>('/admin/subscription-plans/all')
  return data
}

/**
 * Get subscription plan by ID
 */
export async function getById(id: number): Promise<SubscriptionPlan> {
  const { data } = await apiClient.get<SubscriptionPlan>(`/admin/subscription-plans/${id}`)
  return data
}

/**
 * Create a new subscription plan
 */
export async function create(plan: CreateSubscriptionPlanRequest): Promise<SubscriptionPlan> {
  const { data } = await apiClient.post<SubscriptionPlan>('/admin/subscription-plans', plan)
  return data
}

/**
 * Update a subscription plan
 */
export async function update(
  id: number,
  plan: UpdateSubscriptionPlanRequest
): Promise<SubscriptionPlan> {
  const { data } = await apiClient.put<SubscriptionPlan>(`/admin/subscription-plans/${id}`, plan)
  return data
}

/**
 * Delete a subscription plan
 */
export async function remove(id: number): Promise<void> {
  await apiClient.delete(`/admin/subscription-plans/${id}`)
}

export default { list, getAll, getById, create, update, delete: remove }
