<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMailStore } from './composables/useMailStore'
import { api } from './api/client'

const router = useRouter()
const store = useMailStore()
const checking = ref(true)

onMounted(async () => {
  await store.loadAccounts()
  checking.value = false
  if (store.accounts.value.length > 0) {
    const { accountId, folder } = store.loadLastSession()
    // Use restored account if it still exists, otherwise first account
    const acc = (accountId && store.accounts.value.find(a => a.id === accountId))
      || store.accounts.value[0]
    router.replace({ name: 'mail-folder', params: { accountId: acc.id, folder } })
  } else {
    router.replace('/login')
  }
})
</script>

<template>
  <div v-if="checking" class="loading-screen">
    <v-progress-circular indeterminate color="primary" size="40" />
  </div>
  <router-view v-else />
</template>

<style>
html, body {
  margin: 0;
  padding: 0;
  overflow: hidden !important;
  background: #121212;
  height: 100%;
  max-height: 100vh;
  scrollbar-width: none;
  -ms-overflow-style: none;
}
html::-webkit-scrollbar, body::-webkit-scrollbar {
  display: none;
}

* {
  scrollbar-width: none;
  -ms-overflow-style: none;
}
*::-webkit-scrollbar {
  display: none;
}

.loading-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #121212;
}
</style>
