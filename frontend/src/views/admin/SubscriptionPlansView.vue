<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <!-- Left: search + filters -->
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-64">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.subscriptionPlans.searchPlans')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>
            <Select
              v-model="filters.status"
              :options="statusOptions"
              :placeholder="t('admin.subscriptionPlans.allStatus')"
              class="w-40"
              @change="loadPlans"
            />
            <Select
              v-model="filters.visibility"
              :options="visibilityOptions"
              :placeholder="t('admin.subscriptionPlans.allVisibility')"
              class="w-40"
              @change="loadPlans"
            />
          </div>

          <!-- Right: actions -->
          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadPlans"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button
              @click="showCreateModal = true"
              class="btn btn-primary"
            >
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.subscriptionPlans.createPlan') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="plans" :loading="loading">
          <template #cell-name="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-visibility="{ value }">
            <span
              :class="[
                'badge',
                value === 'public'
                  ? 'badge-blue'
                  : 'badge-orange'
              ]"
            >
              {{ t('admin.subscriptionPlans.visibility.' + value) }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'active'
                  ? 'badge-green'
                  : 'badge-gray'
              ]"
            >
              {{ t('admin.subscriptionPlans.status.' + value) }}
            </span>
          </template>

          <template #cell-daily_limit_usd="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value != null ? value + ' U' : t('admin.subscriptionPlans.unlimited') }}
            </span>
          </template>

          <template #cell-weekly_limit_usd="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value != null ? value + ' U' : t('admin.subscriptionPlans.unlimited') }}
            </span>
          </template>

          <template #cell-monthly_limit_usd="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value != null ? value + ' U' : t('admin.subscriptionPlans.unlimited') }}
            </span>
          </template>

          <template #cell-default_validity_days="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value }} {{ t('admin.subscriptionPlans.days') }}
            </span>
          </template>

          <template #cell-price="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value > 0 ? value + ' U' : t('admin.subscriptionPlans.notForSale') }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="handleEdit(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleDelete(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.subscriptionPlans.noPlansYet')"
              :description="t('admin.subscriptionPlans.createFirstPlan')"
              :action-text="t('admin.subscriptionPlans.createPlan')"
              @action="showCreateModal = true"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create Plan Modal -->
    <BaseDialog
      :show="showCreateModal"
      :title="t('admin.subscriptionPlans.createPlan')"
      width="normal"
      @close="closeCreateModal"
    >
      <form id="create-plan-form" @submit.prevent="handleCreatePlan" class="space-y-5">
        <div>
          <label class="input-label">{{ t('admin.subscriptionPlans.form.name') }}</label>
          <input
            v-model="createForm.name"
            type="text"
            class="input"
            :placeholder="t('admin.subscriptionPlans.form.namePlaceholder')"
            required
          />
        </div>

        <div>
          <label class="input-label">{{ t('admin.subscriptionPlans.form.description') }}</label>
          <textarea
            v-model="createForm.description"
            class="input"
            rows="2"
            :placeholder="t('admin.subscriptionPlans.form.descriptionPlaceholder')"
          ></textarea>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.visibility') }}</label>
            <Select
              v-model="createForm.visibility"
              :options="visibilityFormOptions"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.status') }}</label>
            <Select
              v-model="createForm.status"
              :options="statusFormOptions"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.dailyLimit') }}</label>
            <input
              v-model.number="createForm.daily_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
            <p class="input-hint">{{ t('admin.subscriptionPlans.form.limitHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.weeklyLimit') }}</label>
            <input
              v-model.number="createForm.weekly_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.monthlyLimit') }}</label>
            <input
              v-model.number="createForm.monthly_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.validityDays') }}</label>
            <input
              v-model.number="createForm.default_validity_days"
              type="number"
              min="1"
              class="input"
              placeholder="30"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.price') }}</label>
            <input
              v-model.number="createForm.price"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.pricePlaceholder')"
            />
            <p class="input-hint">{{ t('admin.subscriptionPlans.form.priceHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.sortOrder') }}</label>
            <input
              v-model.number="createForm.sort_order"
              type="number"
              min="0"
              class="input"
              placeholder="0"
            />
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="closeCreateModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="create-plan-form"
            :disabled="submitting"
            class="btn btn-primary"
          >
            <svg
              v-if="submitting"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ submitting ? t('admin.subscriptionPlans.creating') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Edit Plan Modal -->
    <BaseDialog
      :show="showEditModal"
      :title="t('admin.subscriptionPlans.editPlan')"
      width="normal"
      @close="closeEditModal"
    >
      <form id="edit-plan-form" @submit.prevent="handleUpdatePlan" class="space-y-5">
        <div>
          <label class="input-label">{{ t('admin.subscriptionPlans.form.name') }}</label>
          <input
            v-model="editForm.name"
            type="text"
            class="input"
            :placeholder="t('admin.subscriptionPlans.form.namePlaceholder')"
            required
          />
        </div>

        <div>
          <label class="input-label">{{ t('admin.subscriptionPlans.form.description') }}</label>
          <textarea
            v-model="editForm.description"
            class="input"
            rows="2"
            :placeholder="t('admin.subscriptionPlans.form.descriptionPlaceholder')"
          ></textarea>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.visibility') }}</label>
            <Select
              v-model="editForm.visibility"
              :options="visibilityFormOptions"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.status') }}</label>
            <Select
              v-model="editForm.status"
              :options="statusFormOptions"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.dailyLimit') }}</label>
            <input
              v-model.number="editForm.daily_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
            <p class="input-hint">{{ t('admin.subscriptionPlans.form.limitHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.weeklyLimit') }}</label>
            <input
              v-model.number="editForm.weekly_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.monthlyLimit') }}</label>
            <input
              v-model.number="editForm.monthly_limit_usd"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.limitPlaceholder')"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.validityDays') }}</label>
            <input
              v-model.number="editForm.default_validity_days"
              type="number"
              min="1"
              class="input"
              placeholder="30"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.price') }}</label>
            <input
              v-model.number="editForm.price"
              type="number"
              step="0.01"
              min="0"
              class="input"
              :placeholder="t('admin.subscriptionPlans.form.pricePlaceholder')"
            />
            <p class="input-hint">{{ t('admin.subscriptionPlans.form.priceHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.subscriptionPlans.form.sortOrder') }}</label>
            <input
              v-model.number="editForm.sort_order"
              type="number"
              min="0"
              class="input"
              placeholder="0"
            />
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="closeEditModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="edit-plan-form"
            :disabled="submitting"
            class="btn btn-primary"
          >
            <svg
              v-if="submitting"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ submitting ? t('admin.subscriptionPlans.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.subscriptionPlans.deletePlan')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { SubscriptionPlan, CreateSubscriptionPlanRequest, UpdateSubscriptionPlanRequest } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
const appStore = useAppStore()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.subscriptionPlans.columns.name'), sortable: true },
  { key: 'visibility', label: t('admin.subscriptionPlans.columns.visibility'), sortable: true },
  { key: 'status', label: t('admin.subscriptionPlans.columns.status'), sortable: true },
  { key: 'daily_limit_usd', label: t('admin.subscriptionPlans.columns.dailyLimit'), sortable: true },
  { key: 'weekly_limit_usd', label: t('admin.subscriptionPlans.columns.weeklyLimit'), sortable: true },
  { key: 'monthly_limit_usd', label: t('admin.subscriptionPlans.columns.monthlyLimit'), sortable: true },
  { key: 'default_validity_days', label: t('admin.subscriptionPlans.columns.validityDays'), sortable: true },
  { key: 'price', label: t('admin.subscriptionPlans.columns.price'), sortable: true },
  { key: 'actions', label: t('admin.subscriptionPlans.columns.actions'), sortable: false }
])

// Data
const plans = ref<SubscriptionPlan[]>([])
const loading = ref(false)
const submitting = ref(false)
const searchQuery = ref('')
let abortController: AbortController | null = null

// Filters
const filters = reactive({
  status: '',
  visibility: ''
})

// Pagination
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

// Filter options
const statusOptions = computed(() => [
  { value: '', label: t('admin.subscriptionPlans.allStatus') },
  { value: 'active', label: t('admin.subscriptionPlans.status.active') },
  { value: 'inactive', label: t('admin.subscriptionPlans.status.inactive') }
])

const visibilityOptions = computed(() => [
  { value: '', label: t('admin.subscriptionPlans.allVisibility') },
  { value: 'public', label: t('admin.subscriptionPlans.visibility.public') },
  { value: 'private', label: t('admin.subscriptionPlans.visibility.private') },
  { value: 'hidden', label: t('admin.subscriptionPlans.visibility.hidden') }
])

const visibilityFormOptions = computed(() => [
  { value: 'public', label: t('admin.subscriptionPlans.visibility.public') },
  { value: 'private', label: t('admin.subscriptionPlans.visibility.private') },
  { value: 'hidden', label: t('admin.subscriptionPlans.visibility.hidden') }
])

const statusFormOptions = computed(() => [
  { value: 'active', label: t('admin.subscriptionPlans.status.active') },
  { value: 'inactive', label: t('admin.subscriptionPlans.status.inactive') }
])

// Create modal
const showCreateModal = ref(false)
const createForm = reactive<CreateSubscriptionPlanRequest>({
  name: '',
  description: '',
  visibility: 'public',
  status: 'active',
  daily_limit_usd: undefined,
  weekly_limit_usd: undefined,
  monthly_limit_usd: undefined,
  default_validity_days: 30,
  price: undefined,
  sort_order: 0
})

// Edit modal
const showEditModal = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const editForm = reactive<UpdateSubscriptionPlanRequest>({
  name: '',
  description: '',
  visibility: 'public',
  status: 'active',
  daily_limit_usd: undefined,
  weekly_limit_usd: undefined,
  monthly_limit_usd: undefined,
  default_validity_days: 30,
  price: undefined,
  sort_order: 0
})

// Delete dialog
const showDeleteDialog = ref(false)
const deletingPlan = ref<SubscriptionPlan | null>(null)

const deleteConfirmMessage = computed(() => {
  if (!deletingPlan.value) return ''
  return t('admin.subscriptionPlans.deleteConfirm', { name: deletingPlan.value.name })
})

// Load plans
const loadPlans = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  const { signal } = currentController
  loading.value = true
  try {
    const response = await adminAPI.subscriptionPlans.list(
      pagination.page,
      pagination.page_size,
      {
        status: (filters.status as 'active' | 'inactive') || undefined,
        visibility: (filters.visibility as 'public' | 'hidden') || undefined
      },
      { signal }
    )
    if (signal.aborted) return
    plans.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (signal.aborted || error?.name === 'AbortError' || error?.code === 'ERR_CANCELED') {
      return
    }
    appStore.showError(t('admin.subscriptionPlans.failedToLoad'))
    console.error('Error loading subscription plans:', error)
  } finally {
    if (abortController === currentController && !signal.aborted) {
      loading.value = false
    }
  }
}

// Search
let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadPlans()
  }, 300)
}

// Pagination handlers
const handlePageChange = (page: number) => {
  pagination.page = page
  loadPlans()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadPlans()
}

// Create plan
const resetCreateForm = () => {
  createForm.name = ''
  createForm.description = ''
  createForm.visibility = 'public'
  createForm.status = 'active'
  createForm.daily_limit_usd = undefined
  createForm.weekly_limit_usd = undefined
  createForm.monthly_limit_usd = undefined
  createForm.default_validity_days = 30
  createForm.price = undefined
  createForm.sort_order = 0
}

const closeCreateModal = () => {
  showCreateModal.value = false
  resetCreateForm()
}

const handleCreatePlan = async () => {
  submitting.value = true
  try {
    const payload: CreateSubscriptionPlanRequest = {
      name: createForm.name,
      description: createForm.description || undefined,
      visibility: createForm.visibility,
      status: createForm.status,
      daily_limit_usd: createForm.daily_limit_usd != null && createForm.daily_limit_usd !== ('' as any) ? createForm.daily_limit_usd : null,
      weekly_limit_usd: createForm.weekly_limit_usd != null && createForm.weekly_limit_usd !== ('' as any) ? createForm.weekly_limit_usd : null,
      monthly_limit_usd: createForm.monthly_limit_usd != null && createForm.monthly_limit_usd !== ('' as any) ? createForm.monthly_limit_usd : null,
      default_validity_days: createForm.default_validity_days || 30,
      price: createForm.price != null && createForm.price !== ('' as any) ? createForm.price : 0,
      sort_order: createForm.sort_order || 0
    }
    await adminAPI.subscriptionPlans.create(payload)
    appStore.showSuccess(t('admin.subscriptionPlans.planCreated'))
    closeCreateModal()
    loadPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.subscriptionPlans.failedToCreate'))
    console.error('Error creating subscription plan:', error)
  } finally {
    submitting.value = false
  }
}

// Edit plan
const handleEdit = (plan: SubscriptionPlan) => {
  editingPlan.value = plan
  editForm.name = plan.name
  editForm.description = plan.description
  editForm.visibility = plan.visibility
  editForm.status = plan.status
  editForm.daily_limit_usd = plan.daily_limit_usd ?? undefined
  editForm.weekly_limit_usd = plan.weekly_limit_usd ?? undefined
  editForm.monthly_limit_usd = plan.monthly_limit_usd ?? undefined
  editForm.default_validity_days = plan.default_validity_days
  editForm.price = plan.price
  editForm.sort_order = plan.sort_order
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingPlan.value = null
}

const handleUpdatePlan = async () => {
  if (!editingPlan.value) return
  submitting.value = true
  try {
    const payload: UpdateSubscriptionPlanRequest = {
      name: editForm.name,
      description: editForm.description,
      visibility: editForm.visibility,
      status: editForm.status,
      daily_limit_usd: editForm.daily_limit_usd != null && editForm.daily_limit_usd !== ('' as any) ? editForm.daily_limit_usd : null,
      weekly_limit_usd: editForm.weekly_limit_usd != null && editForm.weekly_limit_usd !== ('' as any) ? editForm.weekly_limit_usd : null,
      monthly_limit_usd: editForm.monthly_limit_usd != null && editForm.monthly_limit_usd !== ('' as any) ? editForm.monthly_limit_usd : null,
      default_validity_days: editForm.default_validity_days,
      price: editForm.price != null && editForm.price !== ('' as any) ? editForm.price : 0,
      sort_order: editForm.sort_order
    }
    await adminAPI.subscriptionPlans.update(editingPlan.value.id, payload)
    appStore.showSuccess(t('admin.subscriptionPlans.planUpdated'))
    closeEditModal()
    loadPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.subscriptionPlans.failedToUpdate'))
    console.error('Error updating subscription plan:', error)
  } finally {
    submitting.value = false
  }
}

// Delete plan
const handleDelete = (plan: SubscriptionPlan) => {
  deletingPlan.value = plan
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingPlan.value) return
  try {
    await adminAPI.subscriptionPlans.delete(deletingPlan.value.id)
    appStore.showSuccess(t('admin.subscriptionPlans.planDeleted'))
    showDeleteDialog.value = false
    deletingPlan.value = null
    loadPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.subscriptionPlans.failedToDelete'))
    console.error('Error deleting subscription plan:', error)
  }
}

onMounted(() => {
  loadPlans()
})
</script>
