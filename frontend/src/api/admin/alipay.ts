/**
 * Admin Alipay API endpoints
 */

import { apiClient } from '../client'

export interface AlipayConfig {
  app_id: string
  private_key: string
  public_key: string
  is_prod: boolean
}

export interface AlipayConfigResponse {
  app_id: string
  notify_url: string
  is_prod: boolean
  private_key_set: boolean
  public_key_set: boolean
  configured: boolean
}

export interface AlipayOrderRecord {
  id: number
  order_no: string
  user_id: number
  package_id: number
  cny_fee: number
  usd_amount: number
  status: string
  alipay_trade_no: string | null
  expires_at: string
  paid_at: string | null
  created_at: string
}

export async function getConfig(): Promise<AlipayConfigResponse> {
  const { data } = await apiClient.get<AlipayConfigResponse>('/admin/alipay/config')
  return data
}

export async function updateConfig(cfg: AlipayConfig): Promise<void> {
  await apiClient.put('/admin/alipay/config', cfg)
}

export async function setEnabled(enabled: boolean): Promise<void> {
  await apiClient.put('/admin/alipay/enabled', { enabled })
}

export async function listOrders(page = 1, pageSize = 20, status = ''): Promise<{
  items: AlipayOrderRecord[]
  total: number
}> {
  const params: Record<string, unknown> = { page, page_size: pageSize }
  if (status) params.status = status
  const { data } = await apiClient.get<{ items: AlipayOrderRecord[]; total: number }>(
    '/admin/alipay/orders',
    { params }
  )
  return data
}

export const adminAlipayAPI = {
  getConfig,
  updateConfig,
  setEnabled,
  listOrders
}

export default adminAlipayAPI
