import { ref } from 'vue'
import { api, type Account, type Folder, type Message, type FullMessage } from '../api/client'

const folders = ref<Folder[]>([])
const messages = ref<Message[]>([])
const selectedFolder = ref('INBOX')
const selectedMessage = ref<FullMessage | null>(null)
const loading = ref(false)
const error = ref('')

export function useMessages() {
  async function loadFolders(acc: Account) {
    loading.value = true
    error.value = ''
    try {
      folders.value = await api.getFolders(acc)
    } catch (e: any) {
      error.value = e.message || 'Failed to load folders'
      folders.value = []
    } finally {
      loading.value = false
    }
  }

  async function loadMessages(acc: Account, folder: string) {
    selectedFolder.value = folder
    selectedMessage.value = null
    loading.value = true
    error.value = ''
    try {
      const res = await api.getMessages(acc, folder)
      messages.value = res.messages || []
    } catch (e: any) {
      error.value = e.message || 'Failed to load messages'
      messages.value = []
    } finally {
      loading.value = false
    }
  }

  async function openMessage(acc: Account, folder: string, uid: number) {
    loading.value = true
    error.value = ''
    try {
      selectedMessage.value = await api.getMessage(acc, folder, uid)
    } catch (e: any) {
      error.value = e.message || 'Failed to load message'
    } finally {
      loading.value = false
    }
  }

  function closeMessage() {
    selectedMessage.value = null
  }

  function reset() {
    folders.value = []
    messages.value = []
    selectedMessage.value = null
    selectedFolder.value = 'INBOX'
  }

  return {
    folders,
    messages,
    selectedFolder,
    selectedMessage,
    loading,
    error,
    loadFolders,
    loadMessages,
    openMessage,
    closeMessage,
    reset,
  }
}
