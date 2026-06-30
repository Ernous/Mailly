import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'

export function useOAuth() {
  const router = useRouter()
  const connecting = ref(false)

  async function connect(provider: string) {
    connecting.value = true
    try {
      const res = await api.connect(provider)
      const popup = window.open(res.redirect_url, '_blank', 'width=600,height=700')

      const handler = (e: MessageEvent) => {
        if (e.data?.type === 'mailly:connected') {
          window.removeEventListener('message', handler)
          popup?.close()
          connecting.value = false
          router.push('/mail')
        }
      }
      window.addEventListener('message', handler)

      const poll = setInterval(() => {
        if (!popup || popup.closed) {
          clearInterval(poll)
          window.removeEventListener('message', handler)
          connecting.value = false
          router.push('/mail')
        }
      }, 1000)
    } catch {
      connecting.value = false
    }
  }

  return { connecting, connect }
}
