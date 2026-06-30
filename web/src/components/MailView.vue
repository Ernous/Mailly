<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Sidebar from './Sidebar.vue'
import MessageList from './MessageList.vue'
import MessageView from './MessageView.vue'
import ComposeDialog from './ComposeDialog.vue'
import { useMailStore } from '../composables/useMailStore'
import './MailView.css'

const store = useMailStore()
const route = useRoute()
const router = useRouter()
const showCompose = ref(false)
const loading = ref(true)

onMounted(async () => {
  await store.loadAccounts()

  const accId = route.params.accountId as string
  const folder = route.params.folder as string

  if (accId) {
    store.selectedAccountId.value = accId
  }
  const acc = store.selectedAccount.value
  if (acc) {
    await Promise.all([
      store.loadFolders(acc),
      store.loadQuota(acc),
    ])
    await store.loadMessages(acc, folder || 'INBOX')
  }
  loading.value = false
})

watch(() => store.selectedAccountId.value, (id) => {
  if (id && store.selectedFolderName.value) {
    router.replace({ name: 'mail-folder', params: { accountId: id, folder: store.selectedFolderName.value } })
  }
})

watch(() => store.selectedFolderName.value, () => {
  if (store.selectedAccountId.value) {
    router.replace({ name: 'mail-folder', params: { accountId: store.selectedAccountId.value, folder: store.selectedFolderName.value } })
  }
})

async function onSwitchAccount(acc: import('../api/client').Account) {
  await store.selectAccount(acc)
}

function onFolderSelect(folder: string) {
  const acc = store.selectedAccount.value
  if (acc) {
    store.loadMessages(acc, folder)
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
  }
}
</script>

<template>
  <div class="loading-screen" v-if="loading">
    <v-progress-circular indeterminate color="primary" size="40" />
  </div>

  <div v-else class="mail-shell">
    <Sidebar
      :accounts="store.accounts.value"
      :selected-account="store.selectedAccount.value"
      :folders="store.folders.value"
      :selected-folder="store.selectedFolderName.value"
      :quota="store.quota.value"
      @select-folder="onFolderSelect"
      @switch-account="onSwitchAccount"
      @add-account="router.push('/login')"
    />

    <div class="mail-list-col">
      <MessageList
        :messages="store.messages.value"
        :selected-messages="store.selectedMessages.value"
        :loading="store.messagesLoading.value"
        :current-folder="store.selectedFolderName.value"
        :sort-by="store.sortBy.value"
        :sort-asc="store.sortAsc.value"
        @open="onMessageOpen"
        @select-all="store.selectAllMessages"
        @refresh="onRefresh"
        @sort="(f) => store.sortBy.value = f"
        @search="(q, f) => {}"
      />
    </div>

    <div class="mail-view-col">
      <MessageView
        :message="store.selectedMessage.value"
        @close="store.closeMessage"
        @reply="() => {}"
        @reply-all="() => {}"
        @forward="() => {}"
        @delete="() => {}"
      />
    </div>

    <v-btn
      icon
      color="primary"
      size="small"
      class="compose-fab"
      @click="showCompose = true"
    >
      <v-icon size="small">mdi-pencil</v-icon>
    </v-btn>

    <ComposeDialog
      :show="showCompose"
      :accounts="store.accounts.value"
      :selected-account="store.selectedAccount.value"
      @close="showCompose = false"
      @send="() => {}"
    />
  </div>
</template>
