/**
 * Admin Wechat Pay API endpoints
 */

import { apiClient } from '../client'
import type { PaymentPackage } from '../payment'

// 微信支付暂时屏蔽，类型别名保持兼容
export type WechatPayPackage = PaymentPackage

export interface WechatPayConfig {
  appid: string
  mchid: string
  api_key_v3: string
  serial_no: string
  private_key: string
  public_key_id: string
  public_key: string
}

export interface WechatPayConfigResponse {
  appid: string
  mchid: string
  serial_no: string
  notify_url: string
  public_key_id: string
  private_key_set: boolean
  api_key_v3_set: boolean
  public_key_set: boolean
  configured: boolean
}

export interface WechatPayOrderRecord {
  id: number
  order_no: string
  user_id: number
  package_id: number
  cny_fee: number
  usd_amount: number
  status: string
  wechat_trade_no: string | null
  expires_at: string
  paid_at: string | null
  created_at: string
}

export async function getConfig(): Promise<WechatPayConfigResponse> {
  const { data } = await apiClient.get<WechatPayConfigResponse>('/admin/wechat-pay/config')
  return data
}

export async function updateConfig(cfg: WechatPayConfig): Promise<void> {
  await apiClient.put('/admin/wechat-pay/config', cfg)
}

export async function setEnabled(enabled: boolean): Promise<void> {
  await apiClient.put('/admin/wechat-pay/enabled', { enabled })
}

export async function getPackages(): Promise<WechatPayPackage[]> {
  const { data } = await apiClient.get<WechatPayPackage[]>('/admin/wechat-pay/packages')
  return data
}

export async function updatePackages(pkgs: WechatPayPackage[]): Promise<void> {
  await apiClient.put('/admin/wechat-pay/packages', pkgs)
}

export async function listOrders(page = 1, pageSize = 20, status = ''): Promise<{
  items: WechatPayOrderRecord[]
  total: number
}> {
  const params: Record<string, unknown> = { page, page_size: pageSize }
  if (status) params.status = status
  const { data } = await apiClient.get<{ items: WechatPayOrderRecord[]; total: number }>(
    '/admin/wechat-pay/orders',
    { params }
  )
  return data
}

export const adminWechatPayAPI = {
  getConfig,
  updateConfig,
  setEnabled,
  getPackages,
  updatePackages,
  listOrders
}

export default adminWechatPayAPI
