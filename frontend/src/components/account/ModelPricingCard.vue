<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import accountsAPI, { type ModelPricingLookupResult } from '@/api/admin/accounts'

const { t } = useI18n()

interface ModelPricingEntry {
  source: string
  input_cost_per_token: number
  output_cost_per_token: number
  cache_creation_input_token_cost: number
  cache_read_input_token_cost: number
  confirmed_at?: string
}

const props = defineProps<{
  billingModelMapping: Record<string, string>
  modelPricing: Record<string, ModelPricingEntry>
}>()

const emit = defineEmits<{
  (e: 'update:billingModelMapping', value: Record<string, string>): void
  (e: 'update:modelPricing', value: Record<string, ModelPricingEntry>): void
}>()

// 合并展示：以 billing_model_mapping 为主，每行显示 假名字 → 真名字 + 价格
const mappingEntries = computed(() => {
  return Object.entries(props.billingModelMapping || {}).map(([claimed, actual]) => ({
    claimed,
    actual,
    pricing: (props.modelPricing || {})[actual] || null,
  }))
})

// 新增映射
const newClaimed = ref('')
const newActual = ref('')
const lookupLoading = ref(false)
const lookupResult = ref<ModelPricingLookupResult | null>(null)

// 模糊搜索
const searchResults = ref<string[]>([])
const searchLoading = ref(false)
const showSearchDropdown = ref(false)

let lookupTimer: ReturnType<typeof setTimeout> | null = null
let searchTimer: ReturnType<typeof setTimeout> | null = null

function onActualInput() {
  lookupResult.value = null
  if (lookupTimer) clearTimeout(lookupTimer)
  if (searchTimer) clearTimeout(searchTimer)

  const model = newActual.value.trim()
  if (!model || model.length < 2) {
    searchResults.value = []
    showSearchDropdown.value = false
    return
  }

  // 模糊搜索（300ms debounce）
  searchTimer = setTimeout(() => doSearch(model), 300)
}

async function doSearch(query: string) {
  searchLoading.value = true
  try {
    searchResults.value = await accountsAPI.searchPricingModels(query, 15)
    showSearchDropdown.value = searchResults.value.length > 0
  } catch {
    searchResults.value = []
    showSearchDropdown.value = false
  } finally {
    searchLoading.value = false
  }
}

function selectSearchResult(model: string) {
  newActual.value = model
  searchResults.value = []
  showSearchDropdown.value = false
  // 选中后立即查询完整定价
  doLookup(model)
}

function onActualBlur() {
  // 延迟关闭以允许点击搜索结果
  setTimeout(() => { showSearchDropdown.value = false }, 200)
}

async function doLookup(model: string) {
  lookupLoading.value = true
  try {
    lookupResult.value = await accountsAPI.lookupModelPricing(model)
  } catch {
    lookupResult.value = null
  } finally {
    lookupLoading.value = false
  }
}

function addMapping() {
  const claimed = newClaimed.value.trim().toLowerCase()
  const actual = newActual.value.trim().toLowerCase()
  if (!claimed || !actual) return

  // 更新映射
  const updatedMapping = { ...props.billingModelMapping }
  updatedMapping[claimed] = actual
  emit('update:billingModelMapping', updatedMapping)

  // 如果 LiteLLM 查到了价格且该模型未配价格，自动填入
  if (lookupResult.value?.found && lookupResult.value.pricing && !(actual in (props.modelPricing || {}))) {
    const p = lookupResult.value.pricing
    const updatedPricing = { ...props.modelPricing }
    updatedPricing[actual] = {
      source: 'litellm_confirmed',
      input_cost_per_token: p.input_cost_per_token,
      output_cost_per_token: p.output_cost_per_token,
      cache_creation_input_token_cost: p.cache_creation_input_token_cost,
      cache_read_input_token_cost: p.cache_read_input_token_cost,
      confirmed_at: new Date().toISOString(),
    }
    emit('update:modelPricing', updatedPricing)
  }

  newClaimed.value = ''
  newActual.value = ''
  lookupResult.value = null
}

function removeMapping(claimed: string) {
  const updatedMapping = { ...props.billingModelMapping }
  const actual = updatedMapping[claimed]
  delete updatedMapping[claimed]
  emit('update:billingModelMapping', updatedMapping)

  // 如果没有其他映射指向同一个 actual model，也清理其定价
  const stillUsed = Object.values(updatedMapping).includes(actual)
  if (!stillUsed && actual in (props.modelPricing || {})) {
    const updatedPricing = { ...props.modelPricing }
    delete updatedPricing[actual]
    emit('update:modelPricing', updatedPricing)
  }
}

type NumericField = 'input_cost_per_token' | 'output_cost_per_token' | 'cache_creation_input_token_cost' | 'cache_read_input_token_cost'

function updatePricingField(actual: string, field: NumericField, value: number) {
  const updatedPricing = { ...props.modelPricing }
  const entry = { ...(updatedPricing[actual] || { source: 'manual', input_cost_per_token: 0, output_cost_per_token: 0, cache_creation_input_token_cost: 0, cache_read_input_token_cost: 0 }) }
  entry[field] = value
  entry.confirmed_at = new Date().toISOString()
  updatedPricing[actual] = entry
  emit('update:modelPricing', updatedPricing)
}

function formatPrice(v: number): string {
  if (!v) return '0'
  return (v * 1e6).toFixed(4)
}

function parsePrice(s: string): number {
  const v = parseFloat(s)
  if (isNaN(v) || v < 0) return 0
  return v / 1e6 // 用户输入 U/MTok，存储为 U/token
}

function formatPricePerMTok(v: number): string {
  if (!v) return '0 U'
  return (v * 1e6).toFixed(2) + ' U/MTok'
}
</script>

<template>
  <div class="space-y-3">
    <!-- 已配置的映射 -->
    <div v-for="entry in mappingEntries" :key="entry.claimed" class="rounded-lg border border-gray-200 p-3 dark:border-dark-500">
      <div class="mb-2 flex items-center justify-between">
        <div class="text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ entry.claimed }}</span>
          <span class="mx-1.5 text-gray-400">→</span>
          <span class="font-medium text-amber-600 dark:text-amber-400">{{ entry.actual }}</span>
        </div>
        <button
          type="button"
          class="text-xs text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
          @click="removeMapping(entry.claimed)"
        >
          {{ t('common.delete') }}
        </button>
      </div>

      <!-- 价格编辑 -->
      <div v-if="entry.pricing" class="grid grid-cols-2 gap-2">
        <div>
          <label class="text-xs text-gray-500 dark:text-gray-400">Input U/MTok</label>
          <input type="text" class="input-field mt-0.5 text-xs"
            :value="formatPrice(entry.pricing.input_cost_per_token)"
            @change="updatePricingField(entry.actual, 'input_cost_per_token', parsePrice(($event.target as HTMLInputElement).value))" />
        </div>
        <div>
          <label class="text-xs text-gray-500 dark:text-gray-400">Output U/MTok</label>
          <input type="text" class="input-field mt-0.5 text-xs"
            :value="formatPrice(entry.pricing.output_cost_per_token)"
            @change="updatePricingField(entry.actual, 'output_cost_per_token', parsePrice(($event.target as HTMLInputElement).value))" />
        </div>
        <div>
          <label class="text-xs text-gray-500 dark:text-gray-400">Cache Create U/MTok</label>
          <input type="text" class="input-field mt-0.5 text-xs"
            :value="formatPrice(entry.pricing.cache_creation_input_token_cost)"
            @change="updatePricingField(entry.actual, 'cache_creation_input_token_cost', parsePrice(($event.target as HTMLInputElement).value))" />
        </div>
        <div>
          <label class="text-xs text-gray-500 dark:text-gray-400">Cache Read U/MTok</label>
          <input type="text" class="input-field mt-0.5 text-xs"
            :value="formatPrice(entry.pricing.cache_read_input_token_cost)"
            @change="updatePricingField(entry.actual, 'cache_read_input_token_cost', parsePrice(($event.target as HTMLInputElement).value))" />
        </div>
      </div>
      <div v-else class="text-xs text-gray-400 italic">
        No pricing configured for {{ entry.actual }}
      </div>

      <div v-if="entry.pricing" class="mt-1 text-[10px] text-gray-400">
        {{ entry.pricing.source === 'manual' ? 'Manual' : entry.pricing.source === 'litellm_confirmed' ? 'LiteLLM confirmed' : entry.pricing.source }}
        <span v-if="entry.pricing.confirmed_at"> · {{ new Date(entry.pricing.confirmed_at).toLocaleDateString() }}</span>
      </div>
    </div>

    <!-- 添加新映射 -->
    <div class="rounded-lg border border-dashed border-gray-300 p-3 dark:border-dark-400">
      <div class="mb-2 text-xs font-medium text-gray-600 dark:text-gray-300">Add model mapping</div>
      <div class="flex items-center gap-2">
        <input v-model="newClaimed" type="text" class="input-field text-xs"
          placeholder="Claimed name (e.g. claude-sonnet-4)"
          @keydown.enter.prevent="addMapping" />
        <span class="text-gray-400">→</span>
        <div class="relative flex-1">
          <input v-model="newActual" type="text" class="input-field text-xs w-full"
            placeholder="Search billing model..."
            @input="onActualInput"
            @focus="newActual.trim().length >= 2 && searchResults.length > 0 && (showSearchDropdown = true)"
            @blur="onActualBlur"
            @keydown.enter.prevent="addMapping"
            @keydown.escape="showSearchDropdown = false" />
          <div v-if="searchLoading" class="absolute right-2 top-1/2 -translate-y-1/2">
            <span class="block h-3 w-3 animate-spin rounded-full border border-gray-400 border-t-transparent"></span>
          </div>
          <!-- 搜索结果下拉 -->
          <div v-if="showSearchDropdown && searchResults.length > 0"
            class="absolute left-0 right-0 top-full z-50 mt-1 max-h-48 overflow-y-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-500 dark:bg-dark-800">
            <button
              v-for="result in searchResults"
              :key="result"
              type="button"
              class="block w-full px-3 py-1.5 text-left text-xs hover:bg-primary-50 dark:hover:bg-dark-600"
              @mousedown.prevent="selectSearchResult(result)"
            >
              <span class="text-gray-900 dark:text-white">{{ result }}</span>
            </button>
          </div>
        </div>
        <button type="button"
          class="whitespace-nowrap rounded-md bg-primary-600 px-3 py-1.5 text-xs text-white hover:bg-primary-700 disabled:opacity-50"
          :disabled="!newClaimed.trim() || !newActual.trim() || lookupLoading"
          @click="addMapping">
          {{ t('common.add') }}
        </button>
      </div>

      <!-- LiteLLM 查价结果 -->
      <div v-if="lookupLoading" class="mt-2 text-xs text-gray-500">
        Querying LiteLLM pricing...
      </div>
      <div v-else-if="lookupResult?.found && lookupResult.pricing" class="mt-2 rounded bg-green-50 p-2 text-xs dark:bg-green-900/20">
        <div class="mb-1 font-medium text-green-700 dark:text-green-400">LiteLLM price found for {{ lookupResult.model }}:</div>
        <div class="grid grid-cols-2 gap-x-4 gap-y-0.5 text-green-600 dark:text-green-300">
          <span>Input: {{ formatPricePerMTok(lookupResult.pricing.input_cost_per_token) }}</span>
          <span>Output: {{ formatPricePerMTok(lookupResult.pricing.output_cost_per_token) }}</span>
          <span>Cache Create: {{ formatPricePerMTok(lookupResult.pricing.cache_creation_input_token_cost) }}</span>
          <span>Cache Read: {{ formatPricePerMTok(lookupResult.pricing.cache_read_input_token_cost) }}</span>
        </div>
        <div class="mt-1 text-[10px] text-green-500">Prices will be auto-filled when you click "Add"</div>
      </div>
      <div v-else-if="lookupResult && !lookupResult.found && newActual.trim()" class="mt-2 rounded bg-amber-50 p-2 text-xs dark:bg-amber-900/20">
        <span class="text-amber-700 dark:text-amber-400">
          "{{ lookupResult.model }}" not found in LiteLLM — you can still add it and set prices manually.
        </span>
      </div>

      <!-- 通配符提示 -->
      <div class="mt-2 text-[10px] text-gray-400">
        Tip: Use <code class="rounded bg-gray-100 px-1 dark:bg-dark-500">*</code> as claimed name to match all unmapped models.
      </div>
    </div>
  </div>
</template>
