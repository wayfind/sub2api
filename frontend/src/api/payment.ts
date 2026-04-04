/**
 * Wechat Pay API endpoints
 */

import { apiClient } from './client'

export interface WechatPayPackage {
  id: number
  name: string
  cny_amount: number
  usd_amount: number
}

export interface WechatPayOrder {
  order_no: string
  status: 'pending' | 'paid' | 'expired' | 'refunded'
  cny_fee: number
  usd_amount: number
  paid_at: string | null
  expires_at: string
}

export interface CreateOrderResponse {
  order_no: string
  code_url: string
  expires_at: string
}

export async function getPackages(): Promise<WechatPayPackage[]> {
  const { data } = await apiClient.get<WechatPayPackage[]>('/payments/wechat/packages')
  return data
}

export async function createOrder(packageId: number): Promise<CreateOrderResponse> {
  const { data } = await apiClient.post<CreateOrderResponse>('/payments/wechat/create-order', {
    package_id: packageId
  })
  return data
}

export async function getOrderStatus(orderNo: string): Promise<WechatPayOrder> {
  const { data } = await apiClient.get<WechatPayOrder>(`/payments/wechat/order/${orderNo}`)
  return data
}

export const wechatPayAPI = {
  getPackages,
  createOrder,
  getOrderStatus
}

export default wechatPayAPI
