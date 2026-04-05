<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">支付宝配置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">配置支付宝当面付参数，套餐与微信支付共用</p>
      </div>

      <!-- 启用状态 -->
      <div class="card">
        <div class="flex items-center justify-between p-6">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">启用支付宝支付</h2>
            <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">开启后用户可在充值页面使用支付宝扫码支付</p>
          </div>
          <button
            @click="toggleEnabled"
            :class="[
              'relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none',
              enabled ? 'bg-blue-500' : 'bg-gray-200 dark:bg-dark-600'
            ]"
          >
            <span
              :class="[
                'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                enabled ? 'translate-x-6' : 'translate-x-1'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- 支付参数 -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">支付宝参数</h2>
        </div>
        <form @submit.prevent="saveConfig" class="space-y-4 p-6">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label class="input-label">AppID（开放平台应用）</label>
              <input v-model="config.app_id" type="text" class="input mt-1" placeholder="2021..." />
              <p class="input-hint">开放平台 → 应用详情 → AppID</p>
            </div>
            <div>
              <label class="input-label">环境</label>
              <select v-model="config.is_prod" class="input mt-1">
                <option :value="true">正式环境</option>
                <option :value="false">沙箱环境</option>
              </select>
            </div>
            <div class="sm:col-span-2">
              <label class="input-label">回调地址（自动生成）</label>
              <input :value="config.notify_url" type="text" class="input mt-1 bg-gray-50 dark:bg-dark-700 text-gray-500 cursor-default" readonly />
              <p class="input-hint">在支付宝开放平台应用详情 → 开发设置 → 异步通知地址 填入此地址</p>
            </div>
          </div>
          <div>
            <label class="input-label">应用私钥（PKCS1 或 PKCS8 PEM 格式）</label>
            <textarea
              v-model="config.private_key"
              class="input mt-1 font-mono text-xs"
              rows="6"
              :placeholder="config.private_key_set ? '已配置（留空保留原值）\n粘贴新私钥内容可替换' : '-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'"
            />
            <p class="input-hint">开放平台 → 开发设置 → 接口加签方式 → 生成/上传密钥对 → 下载应用私钥</p>
          </div>
          <div>
            <label class="input-label">支付宝公钥</label>
            <textarea
              v-model="config.public_key"
              class="input mt-1 font-mono text-xs"
              rows="4"
              :placeholder="config.public_key_set ? '已配置（留空保留原值）\n粘贴新公钥内容可替换' : '-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n或直接粘贴裸 base64 公钥'"
            />
            <p class="input-hint">开放平台 → 开发设置 → 接口加签方式 → 查看支付宝公钥</p>
          </div>
          <div class="flex justify-end">
            <button type="submit" :disabled="savingConfig" class="btn btn-primary">
              {{ savingConfig ? '保存中...' : '保存配置' }}
            </button>
          </div>
        </form>
      </div>

      <!-- 套餐说明 -->
      <div class="card">
        <div class="flex items-center gap-3 p-6">
          <svg class="h-5 w-5 flex-shrink-0 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-sm text-gray-600 dark:text-dark-300">
            支付宝与微信支付共用同一套充值套餐，请在
            <a href="/admin/wechat-pay" class="text-primary-500 hover:underline">微信支付配置</a>
            页面管理套餐。
          </p>
        </div>
      </div>

      <!-- 订单记录 -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">支付宝充值订单</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">订单号</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户ID</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">金额</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-dark-700">
              <tr v-if="loadingOrders">
                <td colspan="5" class="px-6 py-8 text-center text-gray-400">加载中...</td>
              </tr>
              <tr v-else-if="orders.length === 0">
                <td colspan="5" class="px-6 py-8 text-center text-gray-400">暂无订单</td>
              </tr>
              <tr v-else v-for="order in orders" :key="order.order_no" class="hover:bg-gray-50 dark:hover:bg-dark-750">
                <td class="px-6 py-3 font-mono text-xs text-gray-600 dark:text-dark-300">{{ order.order_no }}</td>
                <td class="px-6 py-3 text-gray-600 dark:text-dark-300">{{ order.user_id }}</td>
                <td class="px-6 py-3 text-gray-900 dark:text-white">
                  ¥{{ (order.cny_fee / 100).toFixed(2) }} → ${{ Number(order.usd_amount).toFixed(2) }}
                </td>
                <td class="px-6 py-3">
                  <span :class="statusClass(order.status)" class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium">
                    {{ statusLabel(order.status) }}
                  </span>
                </td>
                <td class="px-6 py-3 text-xs text-gray-400">{{ formatDate(order.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminAlipayAPI, type AlipayOrderRecord } from '@/api/admin/alipay'

const enabled = ref(false)
const savingConfig = ref(false)
const loadingOrders = ref(false)

const config = ref({
  app_id: '',
  private_key: '',
  public_key: '',
  is_prod: true,
  notify_url: '',
  private_key_set: false,
  public_key_set: false,
})

const orders = ref<AlipayOrderRecord[]>([])

onMounted(async () => {
  await Promise.allSettled([loadConfig(), loadOrders()])
})

async function loadConfig() {
  try {
    const cfg = await adminAlipayAPI.getConfig()
    config.value.notify_url = cfg.notify_url ?? ''
    if (cfg.configured) {
      config.value.app_id = cfg.app_id
      config.value.is_prod = cfg.is_prod
      config.value.private_key_set = cfg.private_key_set ?? false
      config.value.public_key_set = cfg.public_key_set ?? false
    }
  } catch {}
}

async function loadOrders() {
  loadingOrders.value = true
  try {
    const result = await adminAlipayAPI.listOrders(1, 50)
    orders.value = result.items
  } catch {} finally {
    loadingOrders.value = false
  }
}

async function toggleEnabled() {
  enabled.value = !enabled.value
  try {
    await adminAlipayAPI.setEnabled(enabled.value)
  } catch {
    enabled.value = !enabled.value
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    await adminAlipayAPI.updateConfig({
      app_id: config.value.app_id,
      private_key: config.value.private_key,
      public_key: config.value.public_key,
      is_prod: config.value.is_prod,
    })
    config.value.private_key = ''
    config.value.public_key = ''
    await loadConfig()
    alert('配置已保存')
  } catch (e: any) {
    alert(e?.response?.data?.message || '保存失败')
  } finally {
    savingConfig.value = false
  }
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending: '待支付', paid: '已支付', expired: '已过期', refunded: '已退款'
  }
  return map[status] ?? status
}

function statusClass(status: string) {
  const map: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    paid: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    expired: 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-400',
    refunded: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
  }
  return map[status] ?? ''
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString('zh-CN')
}
</script>
