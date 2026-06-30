<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import type { Account, Folder, QuotaInfo } from '../api/client'
import { api } from '../api/client'
import { folderIcon } from '../utils/format'
import './Sidebar.css'

const props = defineProps<{
  accounts: Account[]
  selectedAccount: Account | null
  folders: Folder[]
  selectedFolder: string
  quota: QuotaInfo
}>()

const emit = defineEmits<{
  selectFolder: [folder: string]
  switchAccount: [account: Account]
  addAccount: []
}>()

const menuOpen = ref(false)
const dragging = ref(false)
const sidebarWidth = ref(250)

onMounted(() => {
  setSidebarWidth(250)
})

function setSidebarWidth(w: number) {
  const shell = document.querySelector('.mail-shell') as HTMLElement | null
  if (shell) {
    shell.style.setProperty('--sidebar-w', w + 'px')
  }
}

function startResize(e: MouseEvent) {
  dragging.value = true
  const startX = e.clientX
  const startW = sidebarWidth.value

  function onMove(ev: MouseEvent) {
    const newW = Math.max(180, Math.min(500, startW + ev.clientX - startX))
    sidebarWidth.value = newW
    setSidebarWidth(newW)
  }

  function onUp() {
    dragging.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

const providerLabel = computed(() => {
  const p = props.selectedAccount?.provider
  if (p === 'google') return 'Google'
  if (p === 'microsoft') return 'Microsoft'
  return p || 'Account'
})

const providerIcon = computed(() => {
  const p = props.selectedAccount?.provider
  if (p === 'google') return 'mdi-google'
  if (p === 'microsoft') return 'mdi-microsoft'
  return 'mdi-email'
})

const providerColor = computed(() => {
  const p = props.selectedAccount?.provider
  if (p === 'google') return 'red'
  if (p === 'microsoft') return '#00a4ef'
  return 'grey'
})

const quotaPct = computed(() => {
  if (!props.quota.total) return 0
  return (props.quota.used / props.quota.total) * 100
})

const quotaLabel = computed(() => {
  if (!props.quota.total) return ''
  const pct = quotaPct.value.toFixed(1)
  const totalGb = (props.quota.total / 1024 / 1024 / 1024).toFixed(0)
  return `${pct}% used on ${totalGb} GB`
})

function otherAccounts() {
  return props.accounts.filter(a => a.id !== props.selectedAccount?.id)
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function closeMenu() {
  menuOpen.value = false
}

function handleSignOut() {
  const acc = props.selectedAccount
  if (!acc) return

  api.deleteAccount(acc.id).then(() => {
    window.location.reload()
  }).catch(() => {
    window.location.reload()
  })
  closeMenu()
}

function providerPage() {
  const p = props.selectedAccount?.provider
  if (p === 'microsoft') return 'https://account.microsoft.com'
  return 'https://myaccount.google.com'
}
</script>

<template>
  <div class="sidebar" :class="{ 'is-dragging': dragging }">
    <div class="user-header" @click="toggleMenu">
      <v-avatar size="32" color="#4d8080" class="mr-3">
        <v-img v-if="selectedAccount?.photo_url" :src="selectedAccount.photo_url" alt="" cover />
        <v-icon v-else size="small" color="white">mdi-account</v-icon>
      </v-avatar>
      <div class="user-info">
        <div class="user-name">{{ selectedAccount?.display_name || selectedAccount?.email || 'User' }}</div>
        <div class="user-email">{{ selectedAccount?.email || '' }}</div>
      </div>
      <v-icon size="small" color="grey">mdi-chevron-down</v-icon>
    </div>

    <div v-show="menuOpen" class="account-menu-backdrop" @click="closeMenu" />
    <div v-show="menuOpen" class="account-menu">
      <!-- Provider header -->
      <div class="am-provider">
        <div class="am-provider-left">
          <v-icon :color="providerColor" size="small">{{ providerIcon }}</v-icon>
          <span class="am-provider-name">{{ providerLabel }}</span>
        </div>
        <button class="am-signout" @click="handleSignOut">Sign out</button>
      </div>

      <!-- Current account: big avatar -->
      <div class="am-current">
        <div class="am-current-avatar">
          <img v-if="selectedAccount?.photo_url" :src="selectedAccount.photo_url" alt="" />
          <v-icon v-else size="large" color="white">mdi-account</v-icon>
        </div>
        <div class="am-current-name">{{ selectedAccount?.display_name || 'User' }}</div>
        <div class="am-current-email">{{ selectedAccount?.email || '' }}</div>
        <a
          v-if="selectedAccount?.provider"
          :href="providerPage()"
          target="_blank"
          class="am-current-link"
          @click.stop
        >My {{ providerLabel }} account</a>
      </div>

      <!-- Other accounts -->
      <div v-if="otherAccounts().length" class="am-accounts">
        <div
          v-for="acc in otherAccounts()"
          :key="acc.id"
          class="am-account"
          @click="emit('switchAccount', acc); closeMenu()"
        >
          <div class="am-account-avatar">
            <img v-if="acc.photo_url" :src="acc.photo_url" alt="" />
            <v-icon v-else size="x-small" color="white">mdi-account</v-icon>
          </div>
          <div class="am-account-info">
            <div class="am-account-name">{{ acc.display_name || acc.email }}</div>
            <div class="am-account-email">{{ acc.email }}</div>
          </div>
          <div class="am-account-more">
            <v-icon size="small" color="grey">mdi-dots-horizontal</v-icon>
          </div>
        </div>
      </div>

      <!-- Add another account -->
      <div class="am-add" @click="emit('addAccount'); closeMenu()">
        <div class="am-add-icon">
          <v-icon size="small" color="grey">mdi-plus</v-icon>
        </div>
        <span>Sign in with another account...</span>
      </div>
    </div>

    <div v-if="selectedAccount && quota.total > 0" class="quota-bar">
      <v-progress-linear
        :model-value="quotaPct"
        color="primary"
        bg-color="#333"
        height="2"
      />
      <div class="quota-text text-caption text-medium-emphasis mt-1">{{ quotaLabel }}</div>
    </div>

    <div v-if="folders.length > 0" class="folders-section">
      <div
        v-for="f in folders"
        :key="f.name"
        class="folder-item"
        :class="{ 'folder-active': selectedFolder === f.name }"
        @click="emit('selectFolder', f.name)"
      >
        <v-icon size="small" class="folder-icon" :class="{ 'folder-active-icon': selectedFolder === f.name }">
          {{ folderIcon(f.name) }}
        </v-icon>
        <span class="folder-name">{{ f.name }}</span>
      </div>
    </div>
    <div v-else class="folders-section d-flex align-center justify-center pa-4">
      <v-progress-circular indeterminate size="20" width="2" color="grey" />
    </div>
    <div
      class="resize-handle"
      :class="{ 'resize-active': dragging }"
      @mousedown.prevent="startResize"
    />
  </div>
</template>
