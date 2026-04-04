<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">微信支付配置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">配置微信支付参数和充值套餐</p>
      </div>

      <!-- 启用状态 -->
      <div class="card">
        <div class="flex items-center justify-between p-6">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">启用微信支付</h2>
            <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">开启后用户可在充值页面使用微信扫码支付</p>
          </div>
          <button
            @click="toggleEnabled"
            :class="[
              'relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none',
              enabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
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
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">微信支付参数</h2>
        </div>
        <form @submit.prevent="saveConfig" class="space-y-4 p-6">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label class="input-label">AppID（公众号/小程序 AppID）</label>
              <input v-model="config.appid" type="text" class="input mt-1" placeholder="wx..." />
            </div>
            <div>
              <label class="input-label">商户号（MchID）</label>
              <input v-model="config.mchid" type="text" class="input mt-1" placeholder="1234567890" />
            </div>
            <div>
              <label class="input-label">证书序列号（SerialNo）</label>
              <input v-model="config.serial_no" type="text" class="input mt-1" placeholder="证书序列号" />
            </div>
            <div>
              <label class="input-label">回调地址（NotifyURL）</label>
              <input v-model="config.notify_url" type="url" class="input mt-1" placeholder="https://example.com/api/v1/payments/wechat/notify" />
            </div>
          </div>
          <div>
            <label class="input-label">APIv3 密钥</label>
            <input v-model="config.api_key_v3" type="password" class="input mt-1" placeholder="32位 APIv3 密钥" autocomplete="new-password" />
          </div>
          <div>
            <label class="input-label">商户私钥（PEM 格式）</label>
            <textarea
              v-model="config.private_key"
              class="input mt-1 font-mono text-xs"
              rows="8"
              placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----"
            />
            <p class="input-hint">粘贴 apiclient_key.pem 文件内容</p>
          </div>
          <div class="flex justify-end">
            <button type="submit" :disabled="savingConfig" class="btn btn-primary">
              {{ savingConfig ? '保存中...' : '保存配置' }}
            </button>
          </div>
        </form>
      </div>

      <!-- 充值套餐 -->
      <div class="card">
        <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">充值套餐</h2>
          <button @click="addPackage" class="btn btn-secondary btn-sm">+ 添加套餐</button>
        </div>
        <div class="p-6">
          <div v-if="packages.length === 0" class="py-8 text-center text-gray-400">
            暂无套餐，点击右上角添加
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="(pkg, idx) in packages"
              :key="pkg.id"
              class="flex items-center gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-600"
            >
              <div class="flex-1 grid grid-cols-3 gap-3">
                <div>
                  <label class="text-xs text-gray-500">套餐名称</label>
                  <input v-model="pkg.name" type="text" class="input mt-0.5 py-1.5 text-sm" placeholder="套餐名称" />
                </div>
                <div>
                  <label class="text-xs text-gray-500">支付金额（CNY 元）</label>
                  <input v-model.number="pkg.cny_amount" type="number" step="0.01" min="0.01" class="input mt-0.5 py-1.5 text-sm" />
                </div>
                <div>
                  <label class="text-xs text-gray-500">到账余额（USD）</label>
                  <input v-model.number="pkg.usd_amount" type="number" step="0.01" min="0.01" class="input mt-0.5 py-1.5 text-sm" />
                </div>
              </div>
              <button @click="removePackage(idx)" class="text-gray-400 hover:text-red-500">
                <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
          <div class="mt-4 flex justify-end">
            <button @click="savePackages" :disabled="savingPackages" class="btn btn-primary">
              {{ savingPackages ? '保存中...' : '保存套餐' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 订单记录 -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">充值订单</h2>
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
import { adminWechatPayAPI, type WechatPayConfig, type WechatPayOrderRecord } from '@/api/admin/wechat_pay'
import type { WechatPayPackage } from '@/api/payment'

const enabled = ref(false)
const savingConfig = ref(false)
const savingPackages = ref(false)
const loadingOrders = ref(false)

const config = ref<WechatPayConfig>({
  appid: '',
  mchid: '',
  api_key_v3: '',
  serial_no: '',
  private_key: '',
  notify_url: ''
})

const packages = ref<WechatPayPackage[]>([])
const orders = ref<WechatPayOrderRecord[]>([])

let nextPackageId = 100

onMounted(async () => {
  await Promise.allSettled([loadConfig(), loadPackages(), loadOrders()])
})

async function loadConfig() {
  try {
    const cfg = await adminWechatPayAPI.getConfig()
    if (cfg.configured) {
      config.value.appid = cfg.appid
      config.value.mchid = cfg.mchid
      config.value.serial_no = cfg.serial_no
      config.value.notify_url = cfg.notify_url
    }
  } catch {}
}

async function loadPackages() {
  try {
    const pkgs = await adminWechatPayAPI.getPackages()
    packages.value = pkgs
    nextPackageId = Math.max(nextPackageId, ...pkgs.map((p) => p.id + 1))
  } catch {}
}

async function loadOrders() {
  loadingOrders.value = true
  try {
    const result = await adminWechatPayAPI.listOrders(1, 50)
    orders.value = result.items
  } catch {} finally {
    loadingOrders.value = false
  }
}

async function toggleEnabled() {
  enabled.value = !enabled.value
  try {
    await adminWechatPayAPI.setEnabled(enabled.value)
  } catch {
    enabled.value = !enabled.value
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    await adminWechatPayAPI.updateConfig(config.value)
    alert('配置已保存')
  } catch (e: any) {
    alert(e?.response?.data?.message || '保存失败')
  } finally {
    savingConfig.value = false
  }
}

function addPackage() {
  packages.value.push({ id: nextPackageId++, name: '', cny_amount: 10, usd_amount: 1 })
}

function removePackage(idx: number) {
  packages.value.splice(idx, 1)
}

async function savePackages() {
  savingPackages.value = true
  try {
    await adminWechatPayAPI.updatePackages(packages.value)
    alert('套餐已保存')
  } catch (e: any) {
    alert(e?.response?.data?.message || '保存失败')
  } finally {
    savingPackages.value = false
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
