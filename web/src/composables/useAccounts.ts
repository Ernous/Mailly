import { ref } from 'vue'
import { api, type Account } from '../api/client'

const accounts = ref<Account[]>([])
const selectedAccount = ref<Account | null>(null)
const loading = ref(false)
const error = ref('')

export function useAccounts() {
  async function load() {
    loading.value = true
    error.value = ''
    try {
      accounts.value = await api.getAccounts()
    } catch (e: any) {
      accounts.value = []
    } finally {
      loading.value = false
    }
  }

  function select(acc: Account) {
    selectedAccount.value = acc
  }

  function deselect() {
    selectedAccount.value = null
  }

  return {
    accounts,
    selectedAccount,
    loading,
    error,
    load,
    select,
    deselect,
  }
}
