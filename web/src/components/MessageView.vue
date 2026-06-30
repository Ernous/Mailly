<script setup lang="ts">
import { computed } from 'vue'
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
}>()

const iframeSrcdoc = computed(() => {
  if (!props.message?.html_body) return ''
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
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
  }
  * {
    background-color: transparent !important;
    color: #e0e0e0 !important;
    border-color: #444 !important;
  }
  a { color: #80cbc4 !important; }
  img { max-width: 100%; height: auto; }
  table { max-width: 100%; }
  blockquote { border-left: 3px solid #555; padding-left: 12px; color: #aaa; }
  pre, code { background: #2a2a2a !important; color: #ccc !important; padding: 8px; border-radius: 4px; overflow-x: auto; }
</style>
</head>
<body>${props.message.html_body}</body>
</html>`
})
</script>

<template>
  <div class="message-view">
    <div v-if="!message" class="empty-state">
      <v-icon size="48" color="#555" class="mb-3">mdi-email-outline</v-icon>
      <span class="text-medium-emphasis">Select a message to read</span>
    </div>

    <template v-else>
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
        />
        <div v-else class="msg-text">{{ message.text_body }}</div>
      </div>
    </template>
  </div>
</template>
