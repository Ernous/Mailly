<script setup lang="ts">
import { onMounted, ref, watch, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Sidebar from './Sidebar.vue'
import MessageList from './MessageList.vue'
import MessageView from './MessageView.vue'
import ComposeDialog from './ComposeDialog.vue'
import { useMailStore } from '../composables/useMailStore'
import './MailView.css'

// Mobile navigation: 'sidebar' | 'list' | 'message'
const mobilePanel = ref<'sidebar' | 'list' | 'message'>('list')

// Tablet auto-collapse (768–1024px)
const COLLAPSE_BREAKPOINT = 1024
const MOBILE_BREAKPOINT = 768
const windowWidth = ref(window.innerWidth)
const tabletCollapsed = computed(
  () => windowWidth.value <= COLLAPSE_BREAKPOINT && windowWidth.value > MOBILE_BREAKPOINT
)
const isMobile = computed(() => windowWidth.value <= MOBILE_BREAKPOINT)

// Manual collapse (desktop toggle button) — persisted
const STORAGE_KEY_COLLAPSED = 'sidebar-collapsed'
const manualCollapsed = ref(localStorage.getItem(STORAGE_KEY_COLLAPSED) === 'true')

// Final collapsed state: auto on tablet, manual on desktop
const sidebarCollapsed = computed(() => {
  if (isMobile.value) return false          // mobile uses fullscreen panel, not rail
  if (tabletCollapsed.value) return true    // tablet: always collapsed
  return manualCollapsed.value              // desktop: user preference
})

function toggleSidebarCollapse() {
  manualCollapsed.value = !manualCollapsed.value
  localStorage.setItem(STORAGE_KEY_COLLAPSED, String(manualCollapsed.value))
}

function onResize() { windowWidth.value = window.innerWidth }
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  store.stopAutoRefresh()
})

// Keep --sidebar-w CSS var in sync with collapsed state
watch(sidebarCollapsed, (collapsed) => {
  const shell = document.querySelector('.mail-shell') as HTMLElement | null
  if (shell) shell.style.setProperty('--sidebar-w', collapsed ? '64px' : '250px')
}, { immediate: true })

const store = useMailStore()
const route = useRoute()
const router = useRouter()
const showCompose = ref(false)
const loading = ref(true)

// Compose dialog state for reply/forward
const composeReplyTo = ref<{ to: string; subject: string; body: string } | null>(null)

onMounted(async () => {
  // Set correct sidebar width CSS var immediately on mount
  const shell = document.querySelector('.mail-shell') as HTMLElement | null
  if (shell) shell.style.setProperty('--sidebar-w', sidebarCollapsed.value ? '64px' : (localStorage.getItem('sidebar-width') || '250') + 'px')

  await store.loadAccounts()

  const accId = route.params.accountId as string
  const folderParam = route.params.folder as string  // this is the display name from URL

  if (accId) {
    store.selectedAccountId.value = accId
  }
  const acc = store.selectedAccount.value
  if (acc) {
    await Promise.all([
      store.loadFolders(acc),
      store.loadQuota(acc),
    ])
    // Resolve display name from URL to full IMAP name
    let folderFullName = 'INBOX'
    if (folderParam) {
      const matched = store.folders.value.find(
        f => f.name === folderParam || f.full_name === folderParam
      )
      folderFullName = matched ? matched.full_name : folderParam
    }
    await store.loadMessages(acc, folderFullName)
  }
  loading.value = false
  store.startAutoRefresh()
})

watch(() => store.selectedAccountId.value, (id) => {
  if (id && store.selectedFolderName.value) {
    router.replace({ name: 'mail-folder', params: { accountId: id, folder: store.selectedFolderDisplayName.value } })
  }
})

watch(() => store.selectedFolderName.value, () => {
  if (store.selectedAccountId.value) {
    router.replace({ name: 'mail-folder', params: { accountId: store.selectedAccountId.value, folder: store.selectedFolderDisplayName.value } })
  }
})

async function onSwitchAccount(acc: import('../api/client').Account) {
  await store.selectAccount(acc)
}

function onFolderSelect(folder: string) {
  const acc = store.selectedAccount.value
  if (acc) {
    store.loadMessages(acc, folder)
    store.startAutoRefresh()
    mobilePanel.value = 'list'
  }
}

function onRefresh() {
  const acc = store.selectedAccount.value
  if (acc) {
    store.loadMessages(acc, store.selectedFolderName.value)
  }
}

function onMessageOpen(uid: number) {
  const acc = store.selectedAccount.value
  if (acc) {
    store.openMessage(acc, uid)
    // On mobile, navigate to message view after opening
    mobilePanel.value = 'message'
  }
}

function onReply() {
  const msg = store.selectedMessage.value
  if (!msg) return
  composeReplyTo.value = {
    to: msg.from,
    subject: msg.subject,
    body: msg.text_body || '',
  }
  showCompose.value = true
}

function onReplyAll() {
  const msg = store.selectedMessage.value
  if (!msg) return
  // Include original sender + all To recipients
  const recipients = [msg.from, ...(msg.to ? msg.to.split(',').map(s => s.trim()) : [])]
    .filter(Boolean)
    .join(', ')
  composeReplyTo.value = {
    to: recipients,
    subject: msg.subject,
    body: msg.text_body || '',
  }
  showCompose.value = true
}

function onForward() {
  const msg = store.selectedMessage.value
  if (!msg) return
  composeReplyTo.value = {
    to: '',
    subject: msg.subject.startsWith('Fwd:') ? msg.subject : `Fwd: ${msg.subject}`,
    body: msg.text_body || '',
  }
  showCompose.value = true
}

async function onDelete() {
  const acc = store.selectedAccount.value
  const msg = store.selectedMessage.value
  if (!acc || !msg) return
  await store.deleteMessage(acc, msg.uid)
}

async function onMarkRead() {
  const acc = store.selectedAccount.value
  const msg = store.selectedMessage.value
  if (!acc || !msg) return
  // Already read - no-op via store re-use
  store.openMessage(acc, msg.uid)
}

async function onMarkUnread() {
  const acc = store.selectedAccount.value
  const msg = store.selectedMessage.value
  if (!acc || !msg) return
  await store.markMessageUnread(acc, msg.uid)
}

function onComposeClose() {
  showCompose.value = false
  composeReplyTo.value = null
}
</script>

<template>
  <div class="loading-screen" v-if="loading">
    <v-progress-circular indeterminate color="primary" size="40" />
  </div>

  <div v-else class="mail-shell" :data-mobile-panel="mobilePanel" :class="{ 'shell-dialog-open': showCompose }">
    <!-- Sidebar -->
    <div class="sidebar-col" :class="{ 'mobile-active': mobilePanel === 'sidebar' }">
      <Sidebar
        :accounts="store.accounts.value"
        :selected-account="store.selectedAccount.value"
        :folders="store.folders.value"
        :selected-folder="store.selectedFolderName.value"
        :quota="store.quota.value"
        :collapsed="sidebarCollapsed"
        :is-mobile="isMobile"
        @select-folder="onFolderSelect"
        @switch-account="onSwitchAccount"
        @add-account="router.push('/login')"
        @toggle-collapse="isMobile ? mobilePanel = 'list' : toggleSidebarCollapse()"
      />
    </div>

    <!-- Message list -->
    <div class="mail-list-col" :class="{ 'mobile-active': mobilePanel === 'list' }">
      <!-- Mobile top bar in list view -->
      <div class="mobile-topbar">
        <v-btn icon size="small" variant="text" @click="mobilePanel = 'sidebar'">
          <v-icon size="small">mdi-menu</v-icon>
        </v-btn>
        <span class="mobile-topbar-title">{{ store.selectedFolderDisplayName.value || 'Inbox' }}</span>
        <v-btn icon size="small" variant="text" @click="showCompose = true">
          <v-icon size="small">mdi-pencil</v-icon>
        </v-btn>
      </div>
      <MessageList
        :messages="store.messages.value"
        :selected-messages="store.selectedMessages.value"
        :loading="store.messagesLoading.value"
        :current-folder="store.selectedFolderDisplayName.value"
        :sort-by="store.sortBy.value"
        :sort-asc="store.sortAsc.value"
        :current-page="store.currentPage.value"
        :total-pages="store.totalPages.value"
        :total-messages="store.totalMessages.value"
        @open="onMessageOpen"
        @select-all="store.selectAllMessages"
        @refresh="onRefresh"
        @sort="(f) => store.sortBy.value = f"
        @search="(q, f) => {}"
        @prefetch="(uid) => { const acc = store.selectedAccount.value; if (acc) store.prefetchMessage(acc, uid) }"
        @go-to-page="(page) => { const acc = store.selectedAccount.value; if (acc) { store.loadMessages(acc, store.selectedFolderName.value, page); mobilePanel = 'list' } }"
      />
      <!-- Compose FAB: anchored inside the list column, not fixed to viewport -->
      <v-btn
        icon
        color="primary"
        size="small"
        class="compose-fab"
        @click="showCompose = true"
      >
        <v-icon size="small">mdi-pencil</v-icon>
      </v-btn>
    </div>

    <!-- Message view -->
    <div class="mail-view-col" :class="{ 'mobile-active': mobilePanel === 'message' }">
      <!-- Mobile back button inside message view -->
      <div class="mobile-topbar mobile-topbar-message">
        <v-btn icon size="small" variant="text" @click="mobilePanel = 'list'; store.closeMessage()">
          <v-icon size="small">mdi-arrow-left</v-icon>
        </v-btn>
        <span class="mobile-topbar-title">{{ store.selectedMessage.value?.subject || 'Message' }}</span>
      </div>
      <MessageView
        :message="store.selectedMessage.value"
        @close="() => { store.closeMessage(); mobilePanel = 'list' }"
        @reply="onReply"
        @reply-all="onReplyAll"
        @forward="onForward"
        @delete="onDelete"
        @mark-read="onMarkRead"
        @mark-unread="onMarkUnread"
      />
    </div>

    <ComposeDialog
      :show="showCompose"
      :accounts="store.accounts.value"
      :selected-account="store.selectedAccount.value"
      :reply-to="composeReplyTo"
      @close="onComposeClose"
      @send="() => {}"
    />
  </div>
</template>
