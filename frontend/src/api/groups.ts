/**
 * User Groups API endpoints (non-admin)
 * Handles group-related operations for regular users
 */

import { apiClient } from './client'
import type { Group } from '@/types'

/**
 * Get available groups that the current user can bind to API keys
 * This returns groups based on user's permissions:
 * - Standard groups: public (non-exclusive) or explicitly allowed
 * - Subscription groups: user has active subscription
 * @returns List of available groups
 */
export async function getAvailable(): Promise<Group[]> {
  const { data } = await apiClient.get<Group[]>('/groups/available')
  return data
}

/**
 * Get current user's custom group rate multipliers
 * @returns Map of group_id to custom rate_multiplier
 */
export async function getUserGroupRates(): Promise<Record<number, number>> {
  const { data } = await apiClient.get<Record<number, number> | null>('/groups/rates')
  return data || {}
}

export interface ModelPricing {
  model: string
  input_per_mtok: number
  output_per_mtok: number
}

/**
 * Get model pricing for a specific group (with discount applied)
 * @param groupId Group ID
 * @returns List of models with their pricing
 */
export async function getGroupModelPricing(groupId: number): Promise<ModelPricing[]> {
  const { data } = await apiClient.get<ModelPricing[]>(`/groups/${groupId}/model-pricing`)
  return data
}

export const userGroupsAPI = {
  getAvailable,
  getUserGroupRates,
  getGroupModelPricing
}

export default userGroupsAPI
