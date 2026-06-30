<script setup lang="ts">
import { computed, ref } from 'vue'
import type { FullMessage } from '../api/client'
import { formatDateTime } from '../utils/format'
import './MessageView.css'

const props = defineProps<{
  message: FullMessage | null
}>()

const emit = defineEmits<{
  close: []
  reply: []
  replyAll: []
  forward: []
  delete: []
  markUnread: []
  markRead: []
}>()

const showMoreMenu = ref(false)

const iframeSrcdoc = computed(() => {
  if (!props.message?.html_body) return ''
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  html {
    overflow-x: hidden;
  }
  body {
    margin: 0;
    padding: 16px;
    background: #1e1e1e !important;
    color: #e0e0e0 !important;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 14px;
    line-height: 1.6;
    word-wrap: break-word;
    overflow-x: hidden;
    /* Force content to never exceed viewport */
    max-width: 100%;
    box-sizing: border-box;
  }
  * {
    background-color: transparent !important;
    color: #e0e0e0 !important;
    border-color: #444 !important;
    /* Every element respects container width */
    max-width: 100% !important;
    box-sizing: border-box;
  }
  /* Tables are the main culprit in email layouts */
  table {
    width: 100% !important;
    table-layout: fixed !important;
  }
  td, th {
    word-break: break-word;
  }
  img {
    max-width: 100% !important;
    height: auto !important;
  }
  a { color: #80cbc4 !important; }
  blockquote { border-left: 3px solid #555; padding-left: 12px; color: #aaa; margin: 8px 0; }
  pre, code { background: #2a2a2a !important; color: #ccc !important; padding: 8px; border-radius: 4px; overflow-x: auto; white-space: pre-wrap; }
</style>
</head>
<body>${props.message.html_body}</body>
</html>`
})

function onIframeLoad(e: Event) {
  const iframe = e.target as HTMLIFrameElement
  try {
    const doc = iframe.contentDocument?.documentElement
    if (!doc) return
    // Set iframe height to match content (no JS scaling needed — CSS handles width)
    iframe.style.height = doc.scrollHeight + 'px'
  } catch {}
}
</script>

<template>
  <div class="message-view">
    <div v-if="!message" class="empty-state">
      <v-icon size="48" color="#555" class="mb-3">mdi-email-outline</v-icon>
      <span class="text-medium-emphasis">Select a message to read</span>
    </div>

    <template v-else>
      <!-- Action Toolbar -->
      <div class="msg-toolbar">
        <button class="toolbar-btn" title="Reply" @click="emit('reply')">
          <span class="toolbar-icon">↩</span>
          <span class="toolbar-label">Ответить</span>
        </button>

        <div class="toolbar-btn-group">
          <button class="toolbar-btn" title="Reply All" @click="emit('replyAll')">
            <span class="toolbar-icon">↩↩</span>
            <span class="toolbar-label">Ответить в...</span>
          </button>
        </div>

        <div class="toolbar-btn-group">
          <button class="toolbar-btn" title="Forward" @click="emit('forward')">
            <span class="toolbar-icon">↪</span>
            <span class="toolbar-label">Переслать</span>
          </button>
        </div>

        <div class="toolbar-divider"></div>

        <button class="toolbar-btn toolbar-btn-danger" title="Delete" @click="emit('delete')">
          <v-icon size="18">mdi-delete-outline</v-icon>
          <span class="toolbar-label">Удалить</span>
        </button>

        <div class="toolbar-divider"></div>

        <v-menu v-model="showMoreMenu" :close-on-content-click="true" location="bottom start">
          <template #activator="{ props: menuProps }">
            <button class="toolbar-btn" title="Mark" v-bind="menuProps">
              <v-icon size="18">mdi-tag-outline</v-icon>
              <span class="toolbar-label">Пометить</span>
            </button>
          </template>
          <v-list density="compact" class="toolbar-menu">
            <v-list-item @click="emit('markRead')" prepend-icon="mdi-email-open-outline">
              <v-list-item-title>Как прочитанное</v-list-item-title>
            </v-list-item>
            <v-list-item @click="emit('markUnread')" prepend-icon="mdi-email-outline">
              <v-list-item-title>Как непрочитанное</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

      <div class="msg-header">
        <div class="msg-subject">{{ message.subject || '(No subject)' }}</div>
        <div class="msg-headers">
          <div class="header-row">
            <span class="header-label">From</span>
            <span class="header-value">{{ message.from || '' }}</span>
          </div>
          <div v-if="message.to" class="header-row">
            <span class="header-label">To</span>
            <span class="header-value">{{ message.to }}</span>
          </div>
          <div v-if="message.cc" class="header-row">
            <span class="header-label">Cc</span>
            <span class="header-value">{{ message.cc }}</span>
          </div>
          <div class="header-row">
            <span class="header-label">Date</span>
            <span class="header-value">{{ formatDateTime(message.date) }}</span>
          </div>
        </div>
      </div>

      <div v-if="message.attachments?.length" class="msg-attachments">
        <div v-for="att in message.attachments" :key="att.filename" class="attachment-item">
          <v-icon size="small" class="mr-2">mdi-file</v-icon>
          <span>{{ att.filename }}</span>
          <span class="text-medium-emphasis ml-2">({{ Math.round(att.size / 1024) }}KB)</span>
        </div>
      </div>

      <div class="msg-body">
        <iframe
          v-if="message.html_body"
          :srcdoc="iframeSrcdoc"
          class="msg-iframe"
          sandbox="allow-same-origin"
          frameborder="0"
          @load="onIframeLoad"
        />
        <div v-else class="msg-text">{{ message.text_body }}</div>
      </div>
    </template>
  </div>
</template>
