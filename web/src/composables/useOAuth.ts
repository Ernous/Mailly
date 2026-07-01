import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'

export function useOAuth() {
  const router = useRouter()
  const connecting = ref(false)
  const error = ref('')

  async function connect(provider: string) {
    connecting.value = true
    error.value = ''
    try {
      const res = await api.connect(provider)
      const popup = window.open(res.redirect_url, 'mailly_oauth', 'width=600,height=700,noopener=no')

      const handler = (e: MessageEvent) => {
        if (e.data?.type === 'mailly:connected') {
          window.removeEventListener('message', handler)
          popup?.close()
          connecting.value = false
          if (e.data.account_id) {
            router.push(`/mail/${e.data.account_id}/INBOX`)
          } else {
            router.push('/mail')
          }
        }
      }
      window.addEventListener('message', handler)

      const poll = setInterval(async () => {
        if (!popup || popup.closed) {
          clearInterval(poll)
          window.removeEventListener('message', handler)
          connecting.value = false

          // Popup closed — check if auth actually succeeded by fetching accounts
          try {
            const accounts = await api.getAccounts()
            if (accounts && accounts.length > 0) {
              // Auth succeeded (popup closed after postMessage or user closed after grant)
              router.push(`/mail/${accounts[0].id}/INBOX`)
            } else {
              // Popup closed without completing auth
              error.value = 'Authorization cancelled'
            }
          } catch {
            error.value = 'Authorization cancelled'
          }
        }
      }, 500)
    } catch {
      connecting.value = false
      error.value = 'Failed to start OAuth flow'
    }
  }

  return { connecting, error, connect }
}
