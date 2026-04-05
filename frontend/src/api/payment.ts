/**
 * Alipay payment API endpoints
 */

import { apiClient } from './client'

export interface PaymentPackage {
  id: number
  name: string
  cny_amount: number
  usd_amount: number
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

// ---- 支付宝 ----

async function alipayGetPackages(): Promise<PaymentPackage[]> {
  const { data } = await apiClient.get<PaymentPackage[]>('/payments/alipay/packages')
  return data
}

async function alipayCreateOrder(packageId: number): Promise<AlipayCreateOrderResponse> {
  const { data } = await apiClient.post<AlipayCreateOrderResponse>('/payments/alipay/create-order', {
    package_id: packageId
  })
  return data
}

async function alipayGetOrderStatus(orderNo: string): Promise<AlipayOrder> {
  const { data } = await apiClient.get<AlipayOrder>(`/payments/alipay/order/${orderNo}`)
  return data
}

export const alipayAPI = {
  getPackages: alipayGetPackages,
  createOrder: alipayCreateOrder,
  getOrderStatus: alipayGetOrderStatus
}

// ---- 微信支付（暂时屏蔽）----
// export const wechatPayAPI = { ... }
