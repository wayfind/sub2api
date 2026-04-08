<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">充值订单</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">支付宝充值订单记录，支付宝参数请在服务器配置文件中配置</p>
      </div>

      <!-- 订单记录 -->
      <div class="card">
        <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">支付宝充值订单</h2>
          <div class="flex items-center gap-3">
            <select v-model="statusFilter" @change="onFilterChange" class="input py-1 text-sm">
              <option value="">全部状态</option>
              <option value="pending">待支付</option>
              <option value="paid">已支付</option>
              <option value="expired">已过期</option>
              <option value="refunded">已退款</option>
            </select>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">订单号</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户ID</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">金额</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">支付时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-dark-700">
              <tr v-if="loadingOrders">
                <td colspan="6" class="px-6 py-8 text-center text-gray-400">加载中...</td>
              </tr>
              <tr v-else-if="loadError">
                <td colspan="6" class="px-6 py-8 text-center text-red-400">{{ loadError }}</td>
              </tr>
              <tr v-else-if="orders.length === 0">
                <td colspan="6" class="px-6 py-8 text-center text-gray-400">暂无订单</td>
              </tr>
              <tr v-else v-for="order in orders" :key="order.order_no" class="hover:bg-gray-50 dark:hover:bg-dark-750">
                <td class="px-6 py-3 font-mono text-xs text-gray-600 dark:text-dark-300">{{ order.order_no }}</td>
                <td class="px-6 py-3 text-gray-600 dark:text-dark-300">{{ order.user_id }}</td>
                <td class="px-6 py-3 text-gray-900 dark:text-white">
                  ¥{{ (order.cny_fee / 100).toFixed(2) }} → {{ Number(order.usd_amount).toFixed(2) }} U
                </td>
                <td class="px-6 py-3">
                  <span :class="statusClass(order.status)" class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium">
                    {{ statusLabel(order.status) }}
                  </span>
                </td>
                <td class="px-6 py-3 text-xs text-gray-400">{{ order.paid_at ? formatDate(order.paid_at) : '—' }}</td>
                <td class="px-6 py-3 text-xs text-gray-400">{{ formatDate(order.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- 分页 -->
        <div v-if="total > pageSize" class="flex items-center justify-between border-t border-gray-100 px-6 py-3 dark:border-dark-700">
          <span class="text-sm text-gray-500">共 {{ total }} 条</span>
          <div class="flex items-center gap-2">
            <button @click="prevPage" :disabled="page <= 1" class="btn btn-secondary py-1 text-sm disabled:opacity-40">上一页</button>
            <span class="text-sm text-gray-600 dark:text-dark-300">第 {{ page }} 页</span>
            <button @click="nextPage" :disabled="page * pageSize >= total" class="btn btn-secondary py-1 text-sm disabled:opacity-40">下一页</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminAlipayAPI, type AlipayOrderRecord } from '@/api/admin/alipay'

const loadingOrders = ref(false)
const loadError = ref('')
const orders = ref<AlipayOrderRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const statusFilter = ref('')

onMounted(loadOrders)

async function loadOrders() {
  loadingOrders.value = true
  loadError.value = ''
  try {
    const result = await adminAlipayAPI.listOrders(page.value, pageSize, statusFilter.value)
    orders.value = result.items
    total.value = result.total
  } catch (e: any) {
    loadError.value = e?.response?.data?.message || '加载失败，请刷新重试'
  } finally {
    loadingOrders.value = false
  }
}

function onFilterChange() {
  page.value = 1
  loadOrders()
}

async function prevPage() {
  if (page.value > 1) {
    page.value--
    await loadOrders()
  }
}

async function nextPage() {
  if (page.value * pageSize < total.value) {
    page.value++
    await loadOrders()
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
