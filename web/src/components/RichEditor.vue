<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle, Color, FontSize } from '@tiptap/extension-text-style'
import Link from '@tiptap/extension-link'
import { watch } from 'vue'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit.configure({}),
    Underline,
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    TextStyle,
    FontSize,
    Color,
    Link.configure({ openOnClick: false }),
  ],
  editorProps: {
    attributes: {
      class: 'rich-editor-content',
      'data-placeholder': props.placeholder || 'Write your message...',
    },
  },
  onUpdate({ editor }) {
    emit('update:modelValue', editor.getHTML())
  },
})

// Sync external value changes (e.g. reply prefill)
watch(() => props.modelValue, (val) => {
  if (!editor.value) return
  const current = editor.value.getHTML()
  if (val !== current) {
    editor.value.commands.setContent(val || '', false)
  }
})

function setFontSize(size: string) {
  if (!size) return
  editor.value?.chain().focus().setFontSize(size).run()
}

function setLink() {
  const prev = editor.value?.getAttributes('link').href || ''
  const url = window.prompt('URL', prev)
  if (url === null) return
  if (url === '') {
    editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
  } else {
    editor.value?.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }
}

const FONT_SIZES = ['12px', '14px', '16px', '18px', '20px', '24px', '28px', '32px']
const COLORS = [
  '#e0e0e0', '#ffffff', '#ef9a9a', '#80cbc4', '#90caf9',
  '#ffcc80', '#ce93d8', '#a5d6a7', '#f48fb1', '#888888',
]

defineExpose({ editor })
</script>

<template>
  <div class="rich-editor-wrap">
    <!-- Toolbar -->
    <div class="re-toolbar">
      <!-- Font size -->
      <select class="re-select" title="Font size" @change="(e) => setFontSize((e.target as HTMLSelectElement).value)">
        <option value="" disabled selected>Size</option>
        <option v-for="s in FONT_SIZES" :key="s" :value="s">{{ s }}</option>
      </select>

      <div class="re-divider" />

      <!-- Bold / Italic / Underline / Strike -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('bold') }"
        title="Bold (Ctrl+B)"
        @click="editor?.chain().focus().toggleBold().run()"
      >
        <strong>B</strong>
      </button>
      <button
        class="re-btn re-italic"
        :class="{ active: editor?.isActive('italic') }"
        title="Italic (Ctrl+I)"
        @click="editor?.chain().focus().toggleItalic().run()"
      >
        <em>I</em>
      </button>
      <button
        class="re-btn re-underline"
        :class="{ active: editor?.isActive('underline') }"
        title="Underline (Ctrl+U)"
        @click="editor?.chain().focus().toggleUnderline().run()"
      >
        <span style="text-decoration: underline">U</span>
      </button>
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('strike') }"
        title="Strikethrough"
        @click="editor?.chain().focus().toggleStrike().run()"
      >
        <span style="text-decoration: line-through">S</span>
      </button>

      <div class="re-divider" />

      <!-- Headings -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('heading', { level: 1 }) }"
        title="Heading 1"
        @click="editor?.chain().focus().toggleHeading({ level: 1 }).run()"
      >H1</button>
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('heading', { level: 2 }) }"
        title="Heading 2"
        @click="editor?.chain().focus().toggleHeading({ level: 2 }).run()"
      >H2</button>

      <div class="re-divider" />

      <!-- Lists -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('bulletList') }"
        title="Bullet list"
        @click="editor?.chain().focus().toggleBulletList().run()"
      >
        <v-icon size="14">mdi-format-list-bulleted</v-icon>
      </button>
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('orderedList') }"
        title="Ordered list"
        @click="editor?.chain().focus().toggleOrderedList().run()"
      >
        <v-icon size="14">mdi-format-list-numbered</v-icon>
      </button>

      <div class="re-divider" />

      <!-- Alignment -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive({ textAlign: 'left' }) }"
        title="Align left"
        @click="editor?.chain().focus().setTextAlign('left').run()"
      >
        <v-icon size="14">mdi-format-align-left</v-icon>
      </button>
      <button
        class="re-btn"
        :class="{ active: editor?.isActive({ textAlign: 'center' }) }"
        title="Align center"
        @click="editor?.chain().focus().setTextAlign('center').run()"
      >
        <v-icon size="14">mdi-format-align-center</v-icon>
      </button>
      <button
        class="re-btn"
        :class="{ active: editor?.isActive({ textAlign: 'right' }) }"
        title="Align right"
        @click="editor?.chain().focus().setTextAlign('right').run()"
      >
        <v-icon size="14">mdi-format-align-right</v-icon>
      </button>

      <div class="re-divider" />

      <!-- Link -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('link') }"
        title="Insert link"
        @click="setLink"
      >
        <v-icon size="14">mdi-link</v-icon>
      </button>

      <!-- Blockquote -->
      <button
        class="re-btn"
        :class="{ active: editor?.isActive('blockquote') }"
        title="Blockquote"
        @click="editor?.chain().focus().toggleBlockquote().run()"
      >
        <v-icon size="14">mdi-format-quote-close</v-icon>
      </button>

      <div class="re-divider" />

      <!-- Text color -->
      <div class="re-color-wrap" title="Text color">
        <v-icon size="14" style="color: #bbb">mdi-format-color-text</v-icon>
        <div class="re-color-picker">
          <button
            v-for="c in COLORS"
            :key="c"
            class="re-color-swatch"
            :style="{ background: c }"
            @click="editor?.chain().focus().setColor(c).run()"
          />
          <button class="re-color-swatch re-color-reset" title="Reset color"
            @click="editor?.chain().focus().unsetColor().run()">×</button>
        </div>
      </div>

      <div class="re-divider" />

      <!-- Undo / Redo -->
      <button
        class="re-btn"
        title="Undo (Ctrl+Z)"
        :disabled="!editor?.can().undo()"
        @click="editor?.chain().focus().undo().run()"
      >
        <v-icon size="14">mdi-undo</v-icon>
      </button>
      <button
        class="re-btn"
        title="Redo (Ctrl+Y)"
        :disabled="!editor?.can().redo()"
        @click="editor?.chain().focus().redo().run()"
      >
        <v-icon size="14">mdi-redo</v-icon>
      </button>
    </div>

    <!-- Editor area -->
    <EditorContent :editor="editor" class="re-editor-area" />
  </div>
</template>

<style scoped>
.rich-editor-wrap {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* ── Toolbar ─────────────────────────────────── */
.re-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 1px;
  padding: 4px 10px;
  background: #252525;
  border-bottom: 1px solid #333;
  flex-shrink: 0;
}

.re-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0 5px;
  background: transparent;
  border: none;
  border-radius: 3px;
  color: #aaa;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
  user-select: none;
}
.re-btn:hover {
  background: rgba(255,255,255,0.08);
  color: #e0e0e0;
}
.re-btn:disabled {
  opacity: 0.3;
  cursor: default;
}
.re-btn.active {
  background: rgba(77, 128, 128, 0.3);
  color: #80cbc4;
}

.re-select {
  height: 26px;
  padding: 0 4px;
  background: #2a2a2a;
  border: 1px solid #3a3a3a;
  border-radius: 3px;
  color: #aaa;
  font-size: 12px;
  cursor: pointer;
  outline: none;
}
.re-select:hover {
  border-color: #555;
}

.re-divider {
  width: 1px;
  height: 18px;
  background: #3a3a3a;
  margin: 0 3px;
  flex-shrink: 0;
}

/* ── Color picker ──────────────────────────── */
.re-color-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0 5px;
  border-radius: 3px;
  cursor: pointer;
}
.re-color-wrap:hover {
  background: rgba(255,255,255,0.08);
}
.re-color-picker {
  display: none;
  position: absolute;
  top: 30px;
  left: 0;
  background: #2a2a2a;
  border: 1px solid #444;
  border-radius: 6px;
  padding: 6px;
  gap: 4px;
  flex-wrap: wrap;
  width: 130px;
  z-index: 200;
  box-shadow: 0 4px 16px rgba(0,0,0,0.5);
}
.re-color-wrap:hover .re-color-picker,
.re-color-wrap:focus-within .re-color-picker {
  display: flex;
}
.re-color-swatch {
  width: 18px;
  height: 18px;
  border-radius: 3px;
  border: 1px solid rgba(255,255,255,0.1);
  cursor: pointer;
  transition: transform 0.1s, border-color 0.1s;
}
.re-color-swatch:hover {
  transform: scale(1.2);
  border-color: rgba(255,255,255,0.4);
}
.re-color-reset {
  background: #333;
  color: #aaa;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ── Editor content area ─────────────────── */
.re-editor-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: none;
}
.re-editor-area::-webkit-scrollbar { display: none; }

:deep(.rich-editor-content) {
  outline: none;
  min-height: 200px;
  padding: 12px 16px;
  font-size: 14px;
  color: #e0e0e0;
  line-height: 1.7;
  caret-color: #80cbc4;
}

/* Placeholder */
:deep(.rich-editor-content p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: #555;
  float: left;
  height: 0;
  pointer-events: none;
}
:deep(.rich-editor-content p.is-empty::before) {
  content: attr(data-placeholder);
  color: #555;
  float: left;
  height: 0;
  pointer-events: none;
}

/* Typography inside editor */
:deep(.rich-editor-content h1) { font-size: 22px; font-weight: 700; color: #e0e0e0; margin: 12px 0 6px; }
:deep(.rich-editor-content h2) { font-size: 18px; font-weight: 600; color: #e0e0e0; margin: 10px 0 4px; }
:deep(.rich-editor-content p) { margin: 0 0 4px; }
:deep(.rich-editor-content strong) { color: #f0f0f0; }
:deep(.rich-editor-content em) { color: #d0d0d0; }
:deep(.rich-editor-content u) { text-decoration: underline; }
:deep(.rich-editor-content s) { text-decoration: line-through; color: #999; }
:deep(.rich-editor-content a) { color: #80cbc4; text-decoration: underline; cursor: pointer; }
:deep(.rich-editor-content blockquote) {
  border-left: 3px solid #4d8080;
  padding-left: 12px;
  color: #aaa;
  margin: 8px 0;
}
:deep(.rich-editor-content ul) { padding-left: 20px; margin: 4px 0; }
:deep(.rich-editor-content ol) { padding-left: 20px; margin: 4px 0; }
:deep(.rich-editor-content li) { margin: 2px 0; }
:deep(.rich-editor-content code) {
  background: #2a2a2a;
  border-radius: 3px;
  padding: 1px 4px;
  font-family: monospace;
  font-size: 13px;
  color: #ce93d8;
}
:deep(.rich-editor-content pre) {
  background: #2a2a2a;
  border-radius: 6px;
  padding: 12px;
  overflow-x: auto;
}
:deep(.rich-editor-content pre code) {
  background: none;
  padding: 0;
  color: #ccc;
}
</style>
