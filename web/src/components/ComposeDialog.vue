<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Account } from '../api/client'
import './ComposeDialog.css'

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
const to = ref('')
const cc = ref('')
const bcc = ref('')
const subject = ref('')
const body = ref('')
const attachments = ref<File[]>([])
const sending = ref(false)

watch(() => props.show, (val) => {
  if (val) {
    selectedIdentity.value = props.selectedAccount
    if (props.replyTo) {
      to.value = props.replyTo.to
      subject.value = props.replyTo.subject.startsWith('Re:') ? props.replyTo.subject : `Re: ${props.replyTo.subject}`
      body.value = `\n\n--- Original Message ---\n${props.replyTo.body}`
      cc.value = ''
      bcc.value = ''
    } else {
      to.value = ''
      cc.value = ''
      bcc.value = ''
      subject.value = ''
      body.value = ''
    }
    attachments.value = []
  }
})

function handleSend() {
  sending.value = true
  emit('send', {
    to: to.value,
    cc: cc.value,
    bcc: bcc.value,
    subject: subject.value,
    body: body.value,
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
  <v-dialog :model-value="show" max-width="700" @click:outside="emit('close')">
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

        <div class="compose-field">
          <span class="field-label text-caption text-medium-emphasis">To:</span>
          <v-text-field
            v-model="to"
            density="compact"
            variant="plain"
            placeholder="recipients..."
            hide-details
            class="flex-grow-1"
          />
        </div>

        <div class="compose-field">
          <v-btn size="x-small" variant="text" class="text-caption">
            Cc/Bcc
          </v-btn>
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

        <v-textarea
          v-model="body"
          variant="plain"
          rows="12"
          placeholder="Write your message..."
          hide-details
          class="compose-body"
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
          :disabled="!to || !subject"
          @click="handleSend"
        >
          <v-icon start size="small">mdi-send</v-icon>
          Send
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
