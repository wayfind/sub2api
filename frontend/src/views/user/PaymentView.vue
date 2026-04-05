<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- 余额卡片 -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">当前余额</p>
          <p class="mt-2 text-4xl font-bold text-white">
            ${{ user?.balance?.toFixed(2) || '0.00' }}
          </p>
        </div>
      </div>

      <!-- 套餐选择 -->
      <div class="card">
        <!-- 选项卡 -->
        <div class="flex border-b border-gray-100 dark:border-dark-700">
          <button
            @click="activeTab = 'wechat'"
            :class="[
              'flex flex-1 items-center justify-center gap-2 px-6 py-4 text-sm font-medium transition-colors',
              activeTab === 'wechat'
                ? 'border-b-2 border-primary-500 text-primary-600 dark:text-primary-400'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400'
            ]"
          >
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M9.5 3C6.46 3 4 5.46 4 8.5c0 1.74.81 3.29 2.08 4.31L5.5 15l2.35-1.17c.53.15 1.08.24 1.65.24.18 0 .36-.01.53-.03A5.5 5.5 0 0 1 10 13c0-2.76 2.24-5 5-5 .06 0 .12 0 .19.01A5.498 5.498 0 0 0 9.5 3zM7 7.5c-.55 0-1-.45-1-1s.45-1 1-1 1 .45 1 1-.45 1-1 1zm5 0c-.55 0-1-.45-1-1s.45-1 1-1 1 .45 1 1-.45 1-1 1z"/>
              <path d="M15 9.5c-2.49 0-4.5 1.79-4.5 4S12.51 17.5 15 17.5c.57 0 1.12-.1 1.63-.27L18.5 18.5l-.54-1.83A4.04 4.04 0 0 0 19.5 13.5c0-2.21-2.01-4-4.5-4zm-1.5 4.5c-.41 0-.75-.34-.75-.75S13.09 12.5 13.5 12.5s.75.34.75.75-.34.75-.75.75zm3 0c-.41 0-.75-.34-.75-.75s.34-.75.75-.75.75.34.75.75-.34.75-.75.75z"/>
            </svg>
            微信支付
          </button>
          <button
            @click="activeTab = 'alipay'"
            :class="[
              'flex flex-1 items-center justify-center gap-2 px-6 py-4 text-sm font-medium transition-colors',
              activeTab === 'alipay'
                ? 'border-b-2 border-blue-500 text-blue-600 dark:text-blue-400'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400'
            ]"
          >
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M21.4 15.6c-.3-.1-2-.5-3.7-1 1.3-1.6 2.1-3.6 2.1-5.8C19.8 4.3 15.5 1 12 1S4.2 4.3 4.2 8.8c0 2.8 1.5 5.3 3.9 6.8-1.2.5-2.1.9-2.5 1.1-1.3.6-1.8 1.9-1.1 3 .5.8 1.5 1.3 2.7 1.3.7 0 1.4-.2 2.1-.5 1.1-.5 2.9-1.5 4.7-2.7 1.9.5 3.8.8 5 .8 2.1 0 3-.8 3-2 0-.5-.2-1-.6-1zm-9.4-.1c-1.3 0-2.5-.3-3.5-.7 1.3-1.2 2.7-2.6 3.9-4.1.5.6 1 1.3 1.4 2 .6 1.1 1 2.3 1.1 3.4-1 .3-2 .4-2.9.4z"/>
            </svg>
            支付宝
          </button>
        </div>

        <div class="p-6">
          <!-- 套餐列表（共用） -->
          <div v-if="loadingPackages" class="flex justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          </div>
          <div v-else-if="packages.length === 0" class="py-8 text-center text-gray-500">
            暂无可用套餐，请联系管理员
          </div>
          <div v-else class="grid grid-cols-2 gap-4 sm:grid-cols-3">
            <button
              v-for="pkg in packages"
              :key="pkg.id"
              @click="selectedPackage = pkg"
              :class="[
                'relative rounded-xl border-2 p-4 text-center transition-all',
                selectedPackage?.id === pkg.id
                  ? activeTab === 'wechat'
                    ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                    : 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-gray-200 hover:border-primary-300 dark:border-dark-600'
              ]"
            >
              <div v-if="selectedPackage?.id === pkg.id" class="absolute right-2 top-2">
                <div :class="['flex h-5 w-5 items-center justify-center rounded-full', activeTab === 'wechat' ? 'bg-primary-500' : 'bg-blue-500']">
                  <svg class="h-3 w-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                  </svg>
                </div>
              </div>
              <p class="text-xl font-bold text-gray-900 dark:text-white">¥{{ pkg.cny_amount }}</p>
              <p :class="['mt-1 text-sm', activeTab === 'wechat' ? 'text-primary-600 dark:text-primary-400' : 'text-blue-600 dark:text-blue-400']">到账 ${{ pkg.usd_amount.toFixed(2) }}</p>
              <p class="mt-0.5 text-xs text-gray-400">{{ pkg.name }}</p>
            </button>
          </div>

          <button
            v-if="selectedPackage"
            @click="createOrder"
            :disabled="creatingOrder"
            :class="['mt-6 w-full py-3 btn', activeTab === 'wechat' ? 'btn-primary' : 'bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors disabled:opacity-50']"
          >
            <svg v-if="creatingOrder" class="-ml-1 mr-2 inline-block h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            {{ creatingOrder ? '生成中...' : `立即充值 ¥${selectedPackage.cny_amount}` }}
          </button>
        </div>
      </div>

      <!-- 二维码弹窗 -->
      <div v-if="showQRCode" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
        <div class="w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ activeTab === 'wechat' ? '微信扫码支付' : '支付宝扫码支付' }}
            </h3>
            <button @click="closeQRCode" class="text-gray-400 hover:text-gray-600">
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div v-if="orderStatus === 'paid'" class="text-center py-4">
            <div class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <svg class="h-8 w-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <p class="text-lg font-semibold text-gray-900 dark:text-white">支付成功！</p>
            <p class="mt-1 text-sm text-gray-500">余额已到账，页面将自动刷新</p>
          </div>

          <div v-else-if="orderStatus === 'expired'" class="text-center py-4">
            <p class="text-gray-500">订单已过期，请重新下单</p>
            <button @click="closeQRCode" class="btn btn-primary mt-4 w-full">关闭</button>
          </div>

          <div v-else class="text-center">
            <div class="mb-3 flex justify-center">
              <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" alt="支付二维码" class="h-48 w-48 rounded-lg" />
              <div v-else class="flex h-48 w-48 items-center justify-center rounded-lg bg-gray-100">
                <svg class="h-8 w-8 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              </div>
            </div>
            <p class="text-sm text-gray-500">
              请使用{{ activeTab === 'wechat' ? '微信' : '支付宝' }}扫描上方二维码完成支付
            </p>
            <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">¥{{ selectedPackage?.cny_amount }}</p>
            <p :class="['text-sm', activeTab === 'wechat' ? 'text-primary-600' : 'text-blue-600']">到账 ${{ selectedPackage?.usd_amount.toFixed(2) }}</p>
            <div class="mt-3 flex items-center justify-center gap-1 text-xs text-gray-400">
              <svg class="h-3 w-3 animate-pulse text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <circle cx="10" cy="10" r="10" />
              </svg>
              等待支付中...（{{ remainingSeconds }}s）
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { wechatPayAPI, alipayAPI, type WechatPayPackage } from '@/api/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'

const authStore = useAuthStore()
const user = computed(() => authStore.user)

const activeTab = ref<'wechat' | 'alipay'>('wechat')
const packages = ref<WechatPayPackage[]>([])
const selectedPackage = ref<WechatPayPackage | null>(null)
const loadingPackages = ref(false)
const creatingOrder = ref(false)

const showQRCode = ref(false)
const qrCodeDataUrl = ref('')
const currentOrderNo = ref('')
const orderStatus = ref<'pending' | 'paid' | 'expired' | 'refunded'>('pending')
const remainingSeconds = ref(0)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null
let pollCount = 0
const MAX_POLL_COUNT = 200

onMounted(async () => {
  loadingPackages.value = true
  try {
    packages.value = await wechatPayAPI.getPackages()
  } catch {
    // 未启用时静默处理
  } finally {
    loadingPackages.value = false
  }
})

onUnmounted(() => {
  stopPolling()
})

async function createOrder() {
  if (!selectedPackage.value) return
  creatingOrder.value = true
  try {
    let qrCodeStr = ''
    let orderNo = ''
    let expiresAt = ''

    if (activeTab.value === 'wechat') {
      const resp = await wechatPayAPI.createOrder(selectedPackage.value.id)
      qrCodeStr = resp.code_url
      orderNo = resp.order_no
      expiresAt = resp.expires_at
    } else {
      const resp = await alipayAPI.createOrder(selectedPackage.value.id)
      qrCodeStr = resp.qr_code
      orderNo = resp.order_no
      expiresAt = resp.expires_at
    }

    currentOrderNo.value = orderNo
    orderStatus.value = 'pending'

    const expiresAtMs = new Date(expiresAt).getTime()
    remainingSeconds.value = Math.max(0, Math.floor((expiresAtMs - Date.now()) / 1000))

    qrCodeDataUrl.value = await QRCode.toDataURL(qrCodeStr, { width: 192, margin: 1 })
    showQRCode.value = true
    startPolling()
    startCountdown(expiresAtMs)
  } catch (e: any) {
    alert(e?.response?.data?.message || '创建订单失败，请稍后重试')
  } finally {
    creatingOrder.value = false
  }
}

function startPolling() {
  pollCount = 0
  pollTimer = setInterval(async () => {
    pollCount++
    if (pollCount >= MAX_POLL_COUNT) {
      stopPolling()
      if (orderStatus.value === 'pending') orderStatus.value = 'expired'
      return
    }
    try {
      let status: string
      if (activeTab.value === 'wechat') {
        const order = await wechatPayAPI.getOrderStatus(currentOrderNo.value)
        status = order.status
      } else {
        const order = await alipayAPI.getOrderStatus(currentOrderNo.value)
        status = order.status
      }
      orderStatus.value = status as any
      if (status === 'paid') {
        stopPolling()
        setTimeout(async () => {
          await authStore.refreshUser()
          closeQRCode()
        }, 2000)
      } else if (status === 'expired') {
        stopPolling()
      }
    } catch {
      // 忽略轮询错误
    }
  }, 3000)
}

function startCountdown(expiresAt: number) {
  countdownTimer = setInterval(() => {
    const remaining = Math.floor((expiresAt - Date.now()) / 1000)
    remainingSeconds.value = Math.max(0, remaining)
    if (remaining <= 0) {
      clearInterval(countdownTimer!)
      countdownTimer = null
      if (orderStatus.value === 'pending') {
        orderStatus.value = 'expired'
        stopPolling()
      }
    }
  }, 1000)
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

function closeQRCode() {
  stopPolling()
  showQRCode.value = false
  qrCodeDataUrl.value = ''
  currentOrderNo.value = ''
  orderStatus.value = 'pending'
}
</script>
