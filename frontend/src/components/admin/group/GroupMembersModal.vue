<template>
  <BaseDialog :show="show" :title="modalTitle" width="wide" @close="$emit('close')">
    <div v-if="group" class="space-y-4">
      <!-- Group info -->
      <div class="flex flex-wrap items-center gap-3 rounded-lg bg-gray-50 px-4 py-2.5 text-sm dark:bg-dark-700">
        <span class="inline-flex items-center gap-1.5" :class="platformColorClass">
          <PlatformIcon :platform="group.platform" size="sm" />
          {{ t('admin.groups.platforms.' + group.platform) }}
        </span>
        <span class="text-gray-400">|</span>
        <span class="font-medium text-gray-900 dark:text-white">{{ group.name }}</span>
      </div>

      <!-- Add member -->
      <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-600">
        <h4 class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.groups.members.addMember') }}
        </h4>
        <div class="relative">
          <input
            v-model="searchQuery"
            type="text"
            autocomplete="off"
            class="input w-full"
            :placeholder="t('admin.groups.members.searchPlaceholder')"
            @input="handleSearchUsers"
            @focus="showDropdown = true"
          />
          <div
            v-if="showDropdown && searchResults.length > 0"
            class="absolute left-0 right-0 top-full z-10 mt-1 max-h-48 overflow-y-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-500 dark:bg-dark-700"
          >
            <button
              v-for="user in searchResults"
              :key="user.id"
              type="button"
              class="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-gray-50 dark:hover:bg-dark-600"
              @click="handleAddMember(user)"
            >
              <span class="text-gray-400">#{{ user.id }}</span>
              <span class="text-gray-900 dark:text-white">{{ user.username || user.email }}</span>
              <span v-if="user.username" class="text-xs text-gray-400">{{ user.email }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-8">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <!-- Members list -->
      <div v-else-if="members.length > 0" class="space-y-2">
        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.groups.members.memberList') }}
          <span class="text-xs text-gray-400">({{ members.length }})</span>
        </h4>
        <div class="max-h-64 overflow-y-auto rounded-lg border border-gray-200 dark:border-dark-600">
          <div
            v-for="member in members"
            :key="member.id"
            class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5 last:border-b-0 dark:border-dark-600"
          >
            <div class="flex items-center gap-3">
              <div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
                <span class="text-sm font-medium text-primary-600 dark:text-primary-400">{{ (member.email || '?').charAt(0).toUpperCase() }}</span>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-white">{{ member.username || member.email }}</p>
                <p v-if="member.username" class="text-xs text-gray-500 dark:text-gray-400">{{ member.email }}</p>
              </div>
            </div>
            <button
              @click="handleRemoveMember(member)"
              :disabled="removingId === member.id"
              class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              :title="t('admin.groups.members.remove')"
            >
              <svg v-if="removingId === member.id" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <Icon v-else name="trash" size="sm" />
            </button>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="flex flex-col items-center justify-center py-8 text-center">
        <div class="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
          <Icon name="users" size="md" class="text-gray-400" />
        </div>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.groups.members.noMembers') }}</p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary px-5">{{ t('common.close') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminGroup, AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
  group: AdminGroup | null
}>()
defineEmits<{ (e: 'close'): void }>()

const { t } = useI18n()
const appStore = useAppStore()

const members = ref<AdminUser[]>([])
const loading = ref(false)
const searchQuery = ref('')
const searchResults = ref<AdminUser[]>([])
const showDropdown = ref(false)
const removingId = ref<number | null>(null)

let searchTimer: ReturnType<typeof setTimeout> | null = null

const modalTitle = computed(() => {
  return group.value
    ? t('admin.groups.members.title', { name: group.value.name })
    : t('admin.groups.members.title', { name: '' })
})

// Use a local computed so template can reference group reactively
const group = computed(() => props.group)

const platformColorClass = computed(() => {
  const p = props.group?.platform
  if (p === 'anthropic') return 'text-orange-600 dark:text-orange-400'
  if (p === 'openai') return 'text-emerald-600 dark:text-emerald-400'
  if (p === 'antigravity') return 'text-purple-600 dark:text-purple-400'
  return 'text-blue-600 dark:text-blue-400'
})

watch(
  () => props.show,
  (v) => {
    if (v && props.group) {
      loadMembers()
    } else {
      // reset state on close
      members.value = []
      searchQuery.value = ''
      searchResults.value = []
      showDropdown.value = false
    }
  }
)

const loadMembers = async () => {
  if (!props.group) return
  loading.value = true
  try {
    members.value = await adminAPI.groups.getGroupMembers(props.group.id)
  } catch (error) {
    console.error('Failed to load group members:', error)
  } finally {
    loading.value = false
  }
}

const memberIds = computed(() => new Set(members.value.map((m) => m.id)))

const handleSearchUsers = () => {
  if (searchTimer) clearTimeout(searchTimer)
  const query = searchQuery.value.trim()
  if (!query) {
    searchResults.value = []
    showDropdown.value = false
    return
  }
  searchTimer = setTimeout(async () => {
    try {
      const res = await adminAPI.users.list(1, 10, { search: query })
      // Filter out existing members
      searchResults.value = res.items.filter((u) => !memberIds.value.has(u.id))
      showDropdown.value = true
    } catch (error) {
      console.error('Failed to search users:', error)
    }
  }, 300)
}

const handleAddMember = async (user: AdminUser) => {
  if (!props.group) return
  showDropdown.value = false
  searchQuery.value = ''
  searchResults.value = []
  try {
    await adminAPI.groups.addGroupMember(props.group.id, user.id)
    members.value.push(user)
    appStore.showSuccess(t('admin.groups.members.memberAdded'))
  } catch (error) {
    console.error('Failed to add member:', error)
  }
}

const handleRemoveMember = async (user: AdminUser) => {
  if (!props.group) return
  removingId.value = user.id
  try {
    await adminAPI.groups.removeGroupMember(props.group.id, user.id)
    members.value = members.value.filter((m) => m.id !== user.id)
    appStore.showSuccess(t('admin.groups.members.memberRemoved'))
  } catch (error) {
    console.error('Failed to remove member:', error)
  } finally {
    removingId.value = null
  }
}

// Close dropdown when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.relative')) {
    showDropdown.value = false
  }
}

watch(showDropdown, (v) => {
  if (v) {
    document.addEventListener('click', handleClickOutside)
  } else {
    document.removeEventListener('click', handleClickOutside)
  }
})
</script>
