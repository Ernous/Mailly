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

export function useMailStore() {
  const selectedAccount = computed(() =>
    accounts.value.find(a => a.id === selectedAccountId.value) || null
  )

  async function loadAccounts() {
    try {
      accounts.value = await api.getAccounts()
      if (accounts.value.length > 0 && !selectedAccountId.value) {
        selectedAccountId.value = accounts.value[0].id
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

  async function loadMessages(acc: Account, folder: string) {
    selectedFolderName.value = folder
    selectedMessage.value = null
    messages.value = []
    selectedMessages.value.clear()
    messagesLoading.value = true
    error.value = ''
    try {
      const res = await api.getMessages(acc, folder, 100)
      messages.value = res.messages || []
    } catch (e: any) {
      error.value = e.message || 'Failed to load'
    } finally {
      messagesLoading.value = false
    }
  }

  async function openMessage(acc: Account, uid: number) {
    try {
      selectedMessage.value = await api.getMessage(acc, selectedFolderName.value, uid)
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
    loadAccounts,
    selectAccount,
    selectFolder,
    loadFolders,
    loadMessages,
    openMessage,
    closeMessage,
    toggleMessageSelect,
    selectAllMessages,
    clearSelection,
    loadQuota,
  }
}
