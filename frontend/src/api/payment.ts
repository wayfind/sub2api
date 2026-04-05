/**
 * Wechat Pay & Alipay API endpoints
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

export interface AlipayOrder {
  order_no: string
  status: 'pending' | 'paid' | 'expired' | 'refunded'
  cny_fee: number
  usd_amount: number
  paid_at: string | null
  expires_at: string
}

export interface AlipayCreateOrderResponse {
  order_no: string
  qr_code: string
  expires_at: string
}

// ---- 微信支付 ----

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

// ---- 支付宝 ----

export async function alipayGetPackages(): Promise<WechatPayPackage[]> {
  const { data } = await apiClient.get<WechatPayPackage[]>('/payments/alipay/packages')
  return data
}

export async function alipayCreateOrder(packageId: number): Promise<AlipayCreateOrderResponse> {
  const { data } = await apiClient.post<AlipayCreateOrderResponse>('/payments/alipay/create-order', {
    package_id: packageId
  })
  return data
}

export async function alipayGetOrderStatus(orderNo: string): Promise<AlipayOrder> {
  const { data } = await apiClient.get<AlipayOrder>(`/payments/alipay/order/${orderNo}`)
  return data
}

export const alipayAPI = {
  getPackages: alipayGetPackages,
  createOrder: alipayCreateOrder,
  getOrderStatus: alipayGetOrderStatus
}

export default wechatPayAPI

