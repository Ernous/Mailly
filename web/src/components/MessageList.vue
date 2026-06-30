<script setup lang="ts">
import { ref } from 'vue'
import type { Message } from '../api/client'
import { formatDateShort } from '../utils/format'
import './MessageList.css'

defineProps<{
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
}>()

const showSearch = ref(false)
const searchQuery = ref('')
</script>

<template>
  <div class="message-list">
    <div class="list-header">
      <v-btn icon size="x-small" variant="text" @click="showSearch = !showSearch">
        <v-icon size="small">mdi-magnify</v-icon>
      </v-btn>
      <span class="folder-title">{{ currentFolder || 'Inbox' }}</span>
      <v-spacer />
      <v-btn icon size="x-small" variant="text">
        <v-icon size="small">mdi-filter-variant</v-icon>
      </v-btn>
      <v-btn icon size="x-small" variant="text" @click="emit('sort', sortBy === 'date' ? 'from' : 'date')">
        <v-icon size="small">mdi-sort</v-icon>
      </v-btn>
      <v-btn icon size="x-small" variant="text" @click="emit('refresh')">
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

    <div v-else-if="messages.length === 0" class="empty-state">
      <span class="text-medium-emphasis">No message</span>
    </div>

    <div v-else class="message-items">
      <div
        v-for="msg in messages"
        :key="msg.uid"
        class="message-item"
        :class="{ 'message-unread': !msg.is_read }"
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
