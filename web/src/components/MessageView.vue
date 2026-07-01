<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
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
const iframeRef = ref<HTMLIFrameElement | null>(null)
const msgBodyRef = ref<HTMLDivElement | null>(null)

// Reset iframe height when message changes
watch(() => props.message?.uid, () => {
  iframeContentObserver?.disconnect()
  iframeContentObserver = null
  if (iframeRef.value) {
    iframeRef.value.style.height = '400px'
  }
})

const iframeSrcdoc = computed(() => {
  if (!props.message?.html_body) return ''
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  html {
    overflow-x: auto;
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
    overflow-x: auto;
    /* Encourage content to fit, but never lose it — anything that still
       doesn't fit gets its own horizontal scrollbar instead of being cut off */
    max-width: 100%;
    box-sizing: border-box;
  }
  * {
    background-color: transparent !important;
    color: #e0e0e0 !important;
    border-color: #444 !important;
    box-sizing: border-box;
  }
  /* Tables — don't force-squish layout (that mangles columns); if a table
     genuinely doesn't fit, it scrolls horizontally instead. */
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

function recalcScale() {
  const iframe = iframeRef.value
  if (!iframe) return
  try {
    const doc = iframe.contentDocument?.documentElement
    const body = iframe.contentDocument?.body
    if (!doc || !body) return

    const contentHeight = Math.max(doc.scrollHeight, body.scrollHeight)
    if (!contentHeight) return

    iframe.style.height = `${contentHeight}px`
  } catch {}
}

let recalcRaf = 0
let iframeContentObserver: ResizeObserver | null = null

function scheduleRecalc() {
  cancelAnimationFrame(recalcRaf)
  recalcRaf = requestAnimationFrame(recalcScale)
}

function onIframeLoad() {
  // Immediate recalc
  scheduleRecalc()
  // Re-measure after images/fonts load (mobile browsers are slower)
  setTimeout(scheduleRecalc, 300)
  setTimeout(scheduleRecalc, 1000)

  iframeContentObserver?.disconnect()
  iframeContentObserver = null

  try {
    const doc = iframeRef.value?.contentDocument
    if (!doc) return

    // Force all links to open in a new tab
    doc.querySelectorAll('a[href]').forEach(a => {
      a.setAttribute('target', '_blank')
      a.setAttribute('rel', 'noopener noreferrer')
    })

    const body = doc.body
    if (body && typeof ResizeObserver !== 'undefined') {
      iframeContentObserver = new ResizeObserver(scheduleRecalc)
      iframeContentObserver.observe(body)
    }
  } catch {}
}

// Recompute the iframe height whenever the available space changes — window
// resize, sidebar collapse/expand, device rotation, devtools resizing, etc.
// (container width changes can reflow text and change content height)
let resizeObserver: ResizeObserver | null = null
onMounted(() => {
  window.addEventListener('resize', scheduleRecalc)
  if (typeof ResizeObserver !== 'undefined' && msgBodyRef.value) {
    resizeObserver = new ResizeObserver(scheduleRecalc)
    resizeObserver.observe(msgBodyRef.value)
  }
})
onUnmounted(() => {
  window.removeEventListener('resize', scheduleRecalc)
  resizeObserver?.disconnect()
  iframeContentObserver?.disconnect()
  cancelAnimationFrame(recalcRaf)
})
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

      <div class="msg-body" ref="msgBodyRef">
        <iframe
          v-if="message.html_body"
          ref="iframeRef"
          :srcdoc="iframeSrcdoc"
          class="msg-iframe"
          sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
          frameborder="0"
          @load="onIframeLoad"
        />
        <div v-else class="msg-text">{{ message.text_body }}</div>
      </div>
    </template>
  </div>
</template>