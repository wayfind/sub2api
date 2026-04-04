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
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">选择充值套餐</h2>
        </div>
        <div class="p-6">
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
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-gray-200 hover:border-primary-300 dark:border-dark-600'
              ]"
            >
              <div v-if="selectedPackage?.id === pkg.id" class="absolute right-2 top-2">
                <div class="flex h-5 w-5 items-center justify-center rounded-full bg-primary-500">
                  <svg class="h-3 w-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                  </svg>
                </div>
              </div>
              <p class="text-xl font-bold text-gray-900 dark:text-white">¥{{ pkg.cny_amount }}</p>
              <p class="mt-1 text-sm text-primary-600 dark:text-primary-400">到账 ${{ pkg.usd_amount.toFixed(2) }}</p>
              <p class="mt-0.5 text-xs text-gray-400">{{ pkg.name }}</p>
            </button>
          </div>

          <button
            v-if="selectedPackage"
            @click="createOrder"
            :disabled="creatingOrder"
            class="btn btn-primary mt-6 w-full py-3"
          >
            <svg v-if="creatingOrder" class="-ml-1 mr-2 h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
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
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">微信扫码支付</h3>
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
              <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" alt="微信支付二维码" class="h-48 w-48 rounded-lg" />
              <div v-else class="flex h-48 w-48 items-center justify-center rounded-lg bg-gray-100">
                <svg class="h-8 w-8 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              </div>
            </div>
            <p class="text-sm text-gray-500">请使用微信扫描上方二维码完成支付</p>
            <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">¥{{ selectedPackage?.cny_amount }}</p>
            <p class="text-sm text-primary-600">到账 ${{ selectedPackage?.usd_amount.toFixed(2) }}</p>
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { wechatPayAPI, type WechatPayPackage } from '@/api/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'

const authStore = useAuthStore()
const user = computed(() => authStore.user)

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
const MAX_POLL_COUNT = 200 // 3s × 200 = 600s = 10分钟硬上限

onMounted(async () => {
  loadingPackages.value = true
  try {
    packages.value = await wechatPayAPI.getPackages()
  } catch (e) {
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
    const resp = await wechatPayAPI.createOrder(selectedPackage.value.id)
    currentOrderNo.value = resp.order_no
    orderStatus.value = 'pending'

    // 计算剩余时间
    const expiresAt = new Date(resp.expires_at).getTime()
    remainingSeconds.value = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))

    // 生成二维码
    qrCodeDataUrl.value = await QRCode.toDataURL(resp.code_url, { width: 192, margin: 1 })

    showQRCode.value = true
    startPolling()
    startCountdown(expiresAt)
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
      if (orderStatus.value === 'pending') {
        orderStatus.value = 'expired'
      }
      return
    }
    try {
      const order = await wechatPayAPI.getOrderStatus(currentOrderNo.value)
      orderStatus.value = order.status
      if (order.status === 'paid') {
        stopPolling()
        // 刷新用户余额
        setTimeout(async () => {
          await authStore.refreshUser()
          closeQRCode()
        }, 2000)
      } else if (order.status === 'expired') {
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
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function closeQRCode() {
  stopPolling()
  showQRCode.value = false
  qrCodeDataUrl.value = ''
  currentOrderNo.value = ''
  orderStatus.value = 'pending'
}
</script>
