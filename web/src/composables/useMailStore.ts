import { ref, computed } from 'vue'
import { api, type Account, type Folder, type Message, type FullMessage, type QuotaInfo } from '../api/client'

const accounts = ref<Account[]>([])
const selectedAccountId = ref<string | null>(null)
const selectedFolderName = ref('INBOX')
const messages = ref<Message[]>([])
const selectedMessage = ref<FullMessage | null>(null)
const folders = ref<Folder[]>([])
const loading = ref(false)
const messagesLoading = ref(false)
const error = ref('')
const searchQuery = ref('')
const searchField = ref('subject')
const selectedMessages = ref<Set<number>>(new Set())
const sortBy = ref('date')
const sortAsc = ref(false)
const quota = ref<QuotaInfo>({ used: 0, total: 0 })

// ── Pagination ────────────────────────────────────────────────────────────
const currentPage = ref(1)
const totalPages = ref(1)
const totalMessages = ref(0)
const pageSize = ref(50)

// ── Persistence ───────────────────────────────────────────────────────────
const STORAGE_KEY_ACCOUNT = 'last-account-id'
const STORAGE_KEY_FOLDER  = 'last-folder'

function saveLastSession(accId: string, folder: string) {
  localStorage.setItem(STORAGE_KEY_ACCOUNT, accId)
  localStorage.setItem(STORAGE_KEY_FOLDER, folder)
}

function loadLastSession(): { accountId: string | null; folder: string } {
  return {
    accountId: localStorage.getItem(STORAGE_KEY_ACCOUNT),
    folder: localStorage.getItem(STORAGE_KEY_FOLDER) || 'INBOX',
  }
}

// ── Message cache ─────────────────────────────────────────────────────────
const messageCache = new Map<string, FullMessage>()
const MAX_CACHE = 100

function cacheKey(accId: string, folder: string, uid: number) {
  return `${accId}:${folder}:${uid}`
}

function cacheGet(accId: string, folder: string, uid: number) {
  return messageCache.get(cacheKey(accId, folder, uid)) ?? null
}

function cacheSet(accId: string, folder: string, uid: number, msg: FullMessage) {
  const key = cacheKey(accId, folder, uid)
  if (messageCache.size >= MAX_CACHE) {
    const first = messageCache.keys().next().value
    if (first !== undefined) messageCache.delete(first)
  }
  messageCache.set(key, msg)
}

// ── Auto-refresh ──────────────────────────────────────────────────────────
const AUTO_REFRESH_INTERVAL = 60_000
let refreshTimer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(silentRefresh, AUTO_REFRESH_INTERVAL)
}

function stopAutoRefresh() {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function silentRefresh() {
  const acc = accounts.value.find(a => a.id === selectedAccountId.value) ?? null
  if (!acc || messagesLoading.value) return
  try {
    const res = await api.getMessages(acc, selectedFolderName.value, pageSize.value, 1)
    const incoming = res.messages || []
    if (incoming.length === 0) return
    const existingUids = new Set(messages.value.map(m => m.uid))
    const newMsgs = incoming.filter(m => !existingUids.has(m.uid))
    if (newMsgs.length > 0) {
      messages.value = [...newMsgs, ...messages.value]
      totalMessages.value = res.total
      totalPages.value = res.total_pages
    }
  } catch {
    // silent
  }
}

// ── Store ─────────────────────────────────────────────────────────────────
export function useMailStore() {
  const selectedAccount = computed(() =>
    accounts.value.find(a => a.id === selectedAccountId.value) || null
  )

  const selectedFolderDisplayName = computed(() => {
    const f = folders.value.find(f => f.full_name === selectedFolderName.value)
    return f ? f.name : selectedFolderName.value
  })

  async function loadAccounts() {
    try {
      accounts.value = await api.getAccounts()
      if (accounts.value.length > 0 && !selectedAccountId.value) {
        const { accountId } = loadLastSession()
        const restored = accountId && accounts.value.find(a => a.id === accountId)
        selectedAccountId.value = restored ? restored.id : accounts.value[0].id
      }
    } catch {
      accounts.value = []
    }
  }

  async function selectAccount(acc: Account) {
    selectedAccountId.value = acc.id
    selectedFolderName.value = 'INBOX'
    selectedMessage.value = null
    messages.value = []
    selectedMessages.value.clear()
    saveLastSession(acc.id, 'INBOX')
    await loadFolders(acc)
    await loadMessages(acc, 'INBOX')
    await loadQuota(acc)
  }

  async function selectFolder(folder: string) {
    selectedFolderName.value = folder
    selectedMessage.value = null
    selectedMessages.value.clear()
    const acc = selectedAccount.value
    if (acc) {
      await loadMessages(acc, folder)
    }
  }

  async function loadFolders(acc: Account) {
    loading.value = true
    try {
      folders.value = await api.getFolders(acc)
    } catch {
      folders.value = []
    } finally {
      loading.value = false
    }
  }

  async function loadMessages(acc: Account, folder: string, page = 1) {
    selectedFolderName.value = folder
    selectedMessage.value = null
    messages.value = []
    selectedMessages.value.clear()
    currentPage.value = page
    messagesLoading.value = true
    error.value = ''
    saveLastSession(acc.id, folder)
    try {
      const res = await api.getMessages(acc, folder, pageSize.value, page)
      messages.value = res.messages || []
      totalMessages.value = res.total
      totalPages.value = res.total_pages
      currentPage.value = res.page
    } catch (e: any) {
      error.value = e.message || 'Failed to load'
    } finally {
      messagesLoading.value = false
    }
  }

  // Prefetch a message into cache without displaying it
  function prefetchMessage(acc: Account, uid: number) {
    const folder = selectedFolderName.value
    if (cacheGet(acc.id, folder, uid)) return
    api.getMessage(acc, folder, uid)
      .then(msg => cacheSet(acc.id, folder, uid, msg))
      .catch(() => {})
  }

  async function openMessage(acc: Account, uid: number) {
    try {
      const folder = selectedFolderName.value
      const cached = cacheGet(acc.id, folder, uid)

      if (cached) {
        selectedMessage.value = cached
        if (!cached.is_read) {
          api.markRead(acc, folder, uid).then(() => {
            const updated = { ...cached, is_read: true }
            cacheSet(acc.id, folder, uid, updated)
            selectedMessage.value = updated
            const idx = messages.value.findIndex(m => m.uid === uid)
            if (idx !== -1) messages.value[idx] = { ...messages.value[idx], is_read: true }
          })
        }
        return
      }

      const msg = await api.getMessage(acc, folder, uid)
      cacheSet(acc.id, folder, uid, msg)
      selectedMessage.value = msg

      if (!msg.is_read) {
        api.markRead(acc, folder, uid).then(() => {
          const updated = { ...msg, is_read: true }
          cacheSet(acc.id, folder, uid, updated)
          if (selectedMessage.value?.uid === uid) selectedMessage.value = updated
          const idx = messages.value.findIndex(m => m.uid === uid)
          if (idx !== -1) messages.value[idx] = { ...messages.value[idx], is_read: true }
        })
      }
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function markMessageUnread(acc: Account, uid: number) {
    try {
      await api.markUnread(acc, selectedFolderName.value, uid)
      const idx = messages.value.findIndex(m => m.uid === uid)
      if (idx !== -1) {
        messages.value[idx] = { ...messages.value[idx], is_read: false }
      }
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function deleteMessage(acc: Account, uid: number) {
    try {
      await api.deleteMessage(acc, selectedFolderName.value, uid)
      messages.value = messages.value.filter(m => m.uid !== uid)
      if (selectedMessage.value?.uid === uid) {
        selectedMessage.value = null
      }
    } catch (e: any) {
      error.value = e.message
    }
  }

  function closeMessage() {
    selectedMessage.value = null
  }

  function toggleMessageSelect(uid: number) {
    if (selectedMessages.value.has(uid)) {
      selectedMessages.value.delete(uid)
    } else {
      selectedMessages.value.add(uid)
    }
  }

  function selectAllMessages() {
    if (selectedMessages.value.size === messages.value.length) {
      selectedMessages.value.clear()
    } else {
      messages.value.forEach(m => selectedMessages.value.add(m.uid))
    }
  }

  function clearSelection() {
    selectedMessages.value.clear()
  }

  async function loadQuota(acc: Account) {
    try {
      quota.value = await api.getQuota(acc)
    } catch {
      quota.value = { used: 0, total: 0 }
    }
  }

  return {
    accounts,
    selectedAccount,
    selectedAccountId,
    selectedFolderName,
    selectedFolderDisplayName,
    messages,
    selectedMessage,
    folders,
    loading,
    messagesLoading,
    error,
    searchQuery,
    searchField,
    selectedMessages,
    sortBy,
    sortAsc,
    quota,
    currentPage,
    totalPages,
    totalMessages,
    pageSize,
    loadAccounts,
    selectAccount,
    selectFolder,
    loadFolders,
    loadMessages,
    openMessage,
    closeMessage,
    prefetchMessage,
    startAutoRefresh,
    stopAutoRefresh,
    markMessageUnread,
    deleteMessage,
    toggleMessageSelect,
    selectAllMessages,
    clearSelection,
    loadQuota,
    loadLastSession,
  }
}
