<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted } from 'vue'
import type { Account } from '../api/client'
import RichEditor from './RichEditor.vue'
import './ComposeDialog.css'

// Fullscreen on mobile
const windowWidth = ref(window.innerWidth)
const isMobileScreen = computed(() => windowWidth.value <= 767)
function onResize() { windowWidth.value = window.innerWidth }
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))

const props = defineProps<{
  show: boolean
  accounts: Account[]
  selectedAccount: Account | null
  replyTo?: { to: string; subject: string; body: string } | null
}>()

const emit = defineEmits<{
  close: []
  send: [data: { to: string; cc: string; bcc: string; subject: string; body: string; attachments: File[] }]
}>()

const selectedIdentity = ref<Account | null>(null)
const toRecipients = ref<string[]>([])
const ccRecipients = ref<string[]>([])
const bccRecipients = ref<string[]>([])
const toInput = ref('')
const ccInput = ref('')
const bccInput = ref('')
const showCcBcc = ref(false)
const subject = ref('')
const body = ref('')
const attachments = ref<File[]>([])
const sending = ref(false)

watch(() => props.show, (val) => {
  if (val) {
    selectedIdentity.value = props.selectedAccount
    if (props.replyTo) {
      // Parse "to" into recipients
      toRecipients.value = props.replyTo.to
        ? props.replyTo.to.split(',').map(s => s.trim()).filter(Boolean)
        : []
      subject.value = props.replyTo.subject.startsWith('Re:')
        ? props.replyTo.subject
        : `Re: ${props.replyTo.subject}`
      // Wrap reply body in blockquote HTML
      const quoted = props.replyTo.body
        ? `<p></p><blockquote>${props.replyTo.body.replace(/\n/g, '<br>')}</blockquote>`
        : ''
      body.value = quoted
      ccRecipients.value = []
      bccRecipients.value = []
    } else {
      toRecipients.value = []
      ccRecipients.value = []
      bccRecipients.value = []
      subject.value = ''
      body.value = ''
    }
    toInput.value = ''
    ccInput.value = ''
    bccInput.value = ''
    showCcBcc.value = false
    attachments.value = []
  }
})

function addRecipient(list: string[], inputVal: string, clearFn: () => void) {
  const email = inputVal.trim()
  if (email && !list.includes(email)) {
    list.push(email)
  }
  clearFn()
}

function onToKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ',' || e.key === ' ' || e.key === 'Tab') {
    e.preventDefault()
    addRecipient(toRecipients.value, toInput.value, () => { toInput.value = '' })
  } else if (e.key === 'Backspace' && toInput.value === '' && toRecipients.value.length > 0) {
    toRecipients.value.pop()
  }
}

function onCcKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ',' || e.key === ' ' || e.key === 'Tab') {
    e.preventDefault()
    addRecipient(ccRecipients.value, ccInput.value, () => { ccInput.value = '' })
  } else if (e.key === 'Backspace' && ccInput.value === '' && ccRecipients.value.length > 0) {
    ccRecipients.value.pop()
  }
}

function onBccKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ',' || e.key === ' ' || e.key === 'Tab') {
    e.preventDefault()
    addRecipient(bccRecipients.value, bccInput.value, () => { bccInput.value = '' })
  } else if (e.key === 'Backspace' && bccInput.value === '' && bccRecipients.value.length > 0) {
    bccRecipients.value.pop()
  }
}

function onToBlur() {
  if (toInput.value.trim()) {
    addRecipient(toRecipients.value, toInput.value, () => { toInput.value = '' })
  }
}

function onCcBlur() {
  if (ccInput.value.trim()) {
    addRecipient(ccRecipients.value, ccInput.value, () => { ccInput.value = '' })
  }
}

function onBccBlur() {
  if (bccInput.value.trim()) {
    addRecipient(bccRecipients.value, bccInput.value, () => { bccInput.value = '' })
  }
}

function removeToRecipient(email: string) {
  const idx = toRecipients.value.indexOf(email)
  if (idx !== -1) toRecipients.value.splice(idx, 1)
}

function removeCcRecipient(email: string) {
  const idx = ccRecipients.value.indexOf(email)
  if (idx !== -1) ccRecipients.value.splice(idx, 1)
}

function removeBccRecipient(email: string) {
  const idx = bccRecipients.value.indexOf(email)
  if (idx !== -1) bccRecipients.value.splice(idx, 1)
}

function handleSend() {
  // Flush any pending input
  if (toInput.value.trim()) {
    toRecipients.value.push(toInput.value.trim())
    toInput.value = ''
  }
  sending.value = true
  emit('send', {
    to: toRecipients.value.join(', '),
    cc: ccRecipients.value.join(', '),
    bcc: bccRecipients.value.join(', '),
    subject: subject.value,
    body: body.value,   // now HTML
    attachments: attachments.value,
  })
  setTimeout(() => {
    sending.value = false
    emit('close')
  }, 1000)
}

function addAttachment() {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.onchange = (e: Event) => {
    const files = (e.target as HTMLInputElement).files
    if (files) {
      attachments.value.push(...Array.from(files))
    }
  }
  input.click()
}

function removeAttachment(index: number) {
  attachments.value.splice(index, 1)
}
</script>

<template>
  <v-dialog
    :model-value="show"
    max-width="700"
    :fullscreen="isMobileScreen"
    :transition="isMobileScreen ? 'dialog-bottom-transition' : 'dialog-transition'"
    @click:outside="emit('close')"
  >
    <v-card class="compose-card">
      <div class="compose-toolbar">
        <span class="text-body-2 font-weight-medium">
          {{ replyTo ? 'Reply' : 'Compose' }}
        </span>
        <v-spacer />
        <v-btn icon size="x-small" variant="text" @click="emit('close')">
          <v-icon size="small">mdi-close</v-icon>
        </v-btn>
      </div>

      <v-card-text class="pa-0">
        <div class="compose-field">
          <span class="field-label text-caption text-medium-emphasis">From:</span>
          <v-select
            v-model="selectedIdentity"
            :items="accounts"
            item-title="email"
            return-object
            density="compact"
            variant="plain"
            hide-details
            class="flex-grow-1"
          />
        </div>

        <!-- To field with tags -->
        <div class="compose-field compose-field-tags">
          <span class="field-label text-caption text-medium-emphasis">To:</span>
          <div class="tags-input-wrap">
            <span
              v-for="email in toRecipients"
              :key="email"
              class="recipient-tag"
            >
              {{ email }}
              <button class="tag-remove" @click.stop="removeToRecipient(email)">×</button>
            </span>
            <input
              v-model="toInput"
              class="tags-input"
              placeholder="Add recipients..."
              @keydown="onToKeydown"
              @blur="onToBlur"
            />
          </div>
          <button
            v-if="!showCcBcc"
            class="cc-bcc-toggle"
            @click="showCcBcc = true"
          >
            Cc/Bcc
          </button>
        </div>

        <!-- Cc field -->
        <div v-if="showCcBcc" class="compose-field compose-field-tags">
          <span class="field-label text-caption text-medium-emphasis">Cc:</span>
          <div class="tags-input-wrap">
            <span
              v-for="email in ccRecipients"
              :key="email"
              class="recipient-tag"
            >
              {{ email }}
              <button class="tag-remove" @click.stop="removeCcRecipient(email)">×</button>
            </span>
            <input
              v-model="ccInput"
              class="tags-input"
              placeholder="Add Cc..."
              @keydown="onCcKeydown"
              @blur="onCcBlur"
            />
          </div>
        </div>

        <!-- Bcc field -->
        <div v-if="showCcBcc" class="compose-field compose-field-tags">
          <span class="field-label text-caption text-medium-emphasis">Bcc:</span>
          <div class="tags-input-wrap">
            <span
              v-for="email in bccRecipients"
              :key="email"
              class="recipient-tag"
            >
              {{ email }}
              <button class="tag-remove" @click.stop="removeBccRecipient(email)">×</button>
            </span>
            <input
              v-model="bccInput"
              class="tags-input"
              placeholder="Add Bcc..."
              @keydown="onBccKeydown"
              @blur="onBccBlur"
            />
          </div>
        </div>

        <div class="compose-field">
          <span class="field-label text-caption text-medium-emphasis">Subject:</span>
          <v-text-field
            v-model="subject"
            density="compact"
            variant="plain"
            hide-details
            class="flex-grow-1"
          />
        </div>

        <v-divider />

        <div v-if="attachments.length" class="compose-attachments pa-3">
          <div
            v-for="(file, i) in attachments"
            :key="i"
            class="d-flex align-center mb-1"
          >
            <v-icon size="small" class="mr-2">mdi-file</v-icon>
            <span class="text-body-2">{{ file.name }}</span>
            <v-spacer />
            <span class="text-caption text-medium-emphasis mr-2">
              {{ Math.round(file.size / 1024) }}KB
            </span>
            <v-btn icon size="x-small" variant="text" @click="removeAttachment(i)">
              <v-icon size="small">mdi-close</v-icon>
            </v-btn>
          </div>
        </div>

        <RichEditor
          v-model="body"
          placeholder="Write your message..."
          class="compose-rich-editor"
        />
      </v-card-text>

      <v-card-actions class="compose-actions">
        <v-btn icon size="small" variant="text" @click="addAttachment">
          <v-icon size="small">mdi-attachment</v-icon>
        </v-btn>
        <v-spacer />
        <v-btn variant="text" size="small" @click="emit('close')">Cancel</v-btn>
        <v-btn
          color="primary"
          size="small"
          :loading="sending"
          :disabled="toRecipients.length === 0 && !toInput || !subject"
          @click="handleSend"
        >
          <v-icon start size="small">mdi-send</v-icon>
          Send
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
