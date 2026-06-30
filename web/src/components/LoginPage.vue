<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import { useOAuth } from '../composables/useOAuth'
import type { Account } from '../api/client'
import './LoginPage.css'

const router = useRouter()
const { connecting, connect } = useOAuth()
const existingAccounts = ref<Account[]>([])

const showCustomForm = ref(false)
const customConnecting = ref(false)
const customError = ref('')
const customForm = ref({
  email: '',
  password: '',
  imap_host: '',
  imap_port: 993,
  smtp_host: '',
  smtp_port: 465,
})

const providerMeta: Record<string, { label: string; icon: string; color: string }> = {
  google:    { label: 'Google (Gmail)',    icon: 'mdi-google',     color: 'red' },
  microsoft: { label: 'Microsoft (Outlook)', icon: 'mdi-microsoft', color: '#00a4ef' },
  custom:    { label: 'Custom IMAP',       icon: 'mdi-email',      color: 'grey' },
}

onMounted(async () => {
  try {
    existingAccounts.value = await api.getAccounts()
  } catch {
    existingAccounts.value = []
  }
})

function goToMail(acc: Account) {
  router.push(`/mail/${acc.id}/INBOX`)
}

function providerInfo(provider: string) {
  return providerMeta[provider] || { label: provider, icon: 'mdi-email', color: 'grey' }
}

async function connectCustom() {
  customError.value = ''
  customConnecting.value = true
  try {
    const res = await api.connectCustom(customForm.value)
    if (res.ok) {
      router.push(`/mail/${res.account_id}/INBOX`)
    }
  } catch (e: any) {
    customError.value = e.message || 'Failed to connect'
  } finally {
    customConnecting.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <v-icon size="48" color="primary">mdi-email</v-icon>
        <div class="text-h5 font-weight-medium mt-2">Mailly</div>
        <div class="text-body-2 text-medium-emphasis mt-1">Self-hosted email client</div>
      </div>

      <v-divider class="my-4" />

      <div v-if="!showCustomForm">
        <div v-if="existingAccounts.length > 0" class="existing-accounts">
          <div class="text-caption text-medium-emphasis mb-2">Connected accounts</div>
          <div
            v-for="acc in existingAccounts"
            :key="acc.id"
            class="existing-account-item"
            @click="goToMail(acc)"
          >
            <v-avatar size="28" color="#4d8080" class="mr-3">
              <v-img v-if="acc.photo_url" :src="acc.photo_url" alt="" cover />
              <v-icon v-else size="x-small" color="white">mdi-account</v-icon>
            </v-avatar>
            <div class="ea-info">
              <div class="ea-email">{{ acc.email }}</div>
              <div class="ea-provider text-caption text-medium-emphasis">
                <v-icon :color="providerInfo(acc.provider).color" size="x-small" class="mr-1">
                  {{ providerInfo(acc.provider).icon }}
                </v-icon>
                {{ providerInfo(acc.provider).label }}
              </div>
            </div>
            <v-icon size="small" color="grey">mdi-chevron-right</v-icon>
          </div>
          <v-divider class="my-3" />
          <div class="text-caption text-medium-emphasis mb-2">Add another account</div>
        </div>

        <div class="login-form">
          <v-btn
            block variant="outlined" size="large"
            class="mb-3 oauth-btn"
            :disabled="connecting"
            @click="connect('google')"
          >
            <v-icon start color="red">mdi-google</v-icon>
            Sign in with Google
          </v-btn>
          <v-btn
            block variant="outlined" size="large"
            class="mb-3 oauth-btn"
            :disabled="connecting"
            @click="connect('microsoft')"
          >
            <v-icon start color="#00a4ef">mdi-microsoft</v-icon>
            Sign in with Microsoft
          </v-btn>
          
          <v-btn
            block variant="text" size="large"
            class="oauth-btn"
            :disabled="connecting"
            @click="showCustomForm = true"
          >
            <v-icon start color="grey">mdi-email</v-icon>
            Custom IMAP / SMTP
          </v-btn>

          <v-progress-linear v-if="connecting" indeterminate color="primary" class="mt-4" />
        </div>
      </div>

      <!-- Custom IMAP Form -->
      <div v-else class="custom-form">
        <div class="d-flex align-center mb-4 position-relative">
          <v-btn icon="mdi-arrow-left" variant="text" size="small" density="comfortable" @click="showCustomForm = false" class="mr-3"></v-btn>
          <div class="text-subtitle-1 font-weight-bold">Connect Custom Email</div>
        </div>

        <v-text-field v-model="customForm.email" label="Email Address" density="compact" variant="outlined" class="mb-3" hide-details="auto" />
        <v-text-field v-model="customForm.password" label="Password" type="password" density="compact" variant="outlined" class="mb-3" hide-details="auto" />
        
        <div class="d-flex mb-3" style="gap: 12px">
          <v-text-field v-model="customForm.imap_host" label="IMAP Host" density="compact" variant="outlined" hide-details="auto" class="flex-grow-1" />
          <v-text-field v-model="customForm.imap_port" label="Port" type="number" density="compact" variant="outlined" hide-details="auto" style="width: 100px; flex-grow: 0" />
        </div>

        <div class="d-flex mb-4" style="gap: 12px">
          <v-text-field v-model="customForm.smtp_host" label="SMTP Host" density="compact" variant="outlined" hide-details="auto" class="flex-grow-1" />
          <v-text-field v-model="customForm.smtp_port" label="Port" type="number" density="compact" variant="outlined" hide-details="auto" style="width: 100px; flex-grow: 0" />
        </div>

        <v-alert v-if="customError" type="error" density="compact" class="mb-4 text-caption" variant="tonal">
          {{ customError }}
        </v-alert>

        <v-btn
          block
          color="primary"
          size="large"
          class="mt-2"
          :loading="customConnecting"
          :disabled="!customForm.email || !customForm.password || !customForm.imap_host"
          @click="connectCustom"
        >
          Connect
        </v-btn>
      </div>

    </div>
  </div>
</template>
