<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Message } from '../api/client'
import { formatDateShort } from '../utils/format'
import './MessageList.css'

const props = defineProps<{
  messages: Message[]
  selectedMessages: Set<number>
  loading: boolean
  currentFolder: string
  sortBy: string
  sortAsc: boolean
}>()

const emit = defineEmits<{
  open: [uid: number]
  selectAll: []
  sort: [field: string]
  search: [query: string, field: string]
  refresh: []
  prefetch: [uid: number]
}>()

const showSearch = ref(false)
const searchQuery = ref('')

const filteredMessages = computed(() => {
  let m = props.messages
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    m = m.filter(msg => 
      (msg.subject && msg.subject.toLowerCase().includes(q)) || 
      (msg.from && msg.from.toLowerCase().includes(q))
    )
  }
  return m
})
</script>

<template>
  <div class="message-list">
    <div class="list-header">
      <v-btn icon size="small" variant="text" @click="showSearch = !showSearch">
        <v-icon size="small">mdi-magnify</v-icon>
      </v-btn>
      <span class="folder-title" style="font-size: 16px; font-weight: 500;">{{ currentFolder || 'Inbox' }}</span>
      <v-spacer />
      <v-btn icon size="small" variant="text" @click="emit('refresh')">
        <v-icon size="small">mdi-refresh</v-icon>
      </v-btn>
    </div>

    <v-expand-transition>
      <div v-if="showSearch" class="search-input">
        <v-text-field
          v-model="searchQuery"
          density="compact"
          variant="outlined"
          placeholder="Search..."
          hide-details
          clearable
          prepend-inner-icon="mdi-magnify"
          @keyup.enter="emit('search', searchQuery, '')"
        />
      </div>
    </v-expand-transition>

    <div v-if="loading" class="d-flex align-center justify-center pa-6">
      <v-progress-circular indeterminate color="primary" size="24" />
    </div>

    <div v-else-if="filteredMessages.length === 0" class="empty-state">
      <span class="text-medium-emphasis">No message found</span>
    </div>

    <div v-else class="message-items">
      <div
        v-for="msg in filteredMessages"
        :key="msg.uid"
        class="message-item"
        :class="{ 'message-unread': !msg.is_read }"
        @mouseenter="emit('prefetch', msg.uid)"
        @click="emit('open', msg.uid)"
      >
        <div class="msg-dot">
          <span v-if="!msg.is_read" class="unread-dot" />
        </div>
        <div class="msg-content">
          <div class="msg-top">
            <span class="msg-from">{{ msg.from.split('<')[0].trim() || msg.from }}</span>
            <span class="msg-date">{{ formatDateShort(msg.date) }}</span>
          </div>
          <div class="msg-bottom">
            <v-icon v-if="msg.has_attachments" size="x-small" color="#555">mdi-paperclip</v-icon>
            <span class="msg-subject">{{ msg.subject || '(No subject)' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
