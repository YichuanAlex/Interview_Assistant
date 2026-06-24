<template>
  <div class="ocr-input-panel" style="--wails-draggable: no-drag">
    <div class="ocr-header">
      <Icon name="text" :size="13" />
      <span class="ocr-title">识别文本 / 输入消息</span>
      <span v-if="isDeepSeek" class="ocr-badge">DeepSeek 文本模式</span>
    </div>
    <textarea
      ref="textareaRef"
      v-model="localText"
      class="ocr-textarea"
      placeholder="截图后自动填入 OCR 文字；回答后可粘贴报错继续追问；Shift+Enter 换行"
      @keydown.enter.prevent="handleEnter"
    />
    <div class="ocr-actions">
      <span class="ocr-hint">
        <kbd>{{ sendShortcut }}</kbd> 发送
      </span>
      <button class="ocr-send-btn" @click="send" :disabled="!canSend || settingsStore.statusText === '正在思考...'">
        <Icon name="send" :size="14" />
        <span>发送</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useSolutionStore } from '../stores/solution'
import { useSettingsStore } from '../stores/settings'
import { api } from '../services/api'
import Icon from './Icon.vue'

const solution = useSolutionStore()
const settingsStore = useSettingsStore()

const localText = ref('')
const textareaRef = ref(null)

const sendShortcut = computed(() => settingsStore.sendShortcut || 'Cmd+J')

const isDeepSeek = computed(() => {
  const url = (settingsStore.settings.baseURL || '').toLowerCase()
  const model = (settingsStore.settings.model || '').toLowerCase()
  return url.includes('deepseek') || model.includes('deepseek')
})

const canSend = computed(() => localText.value.trim().length > 0)

// 当后端识别到新文字时，追加到输入框
watch(() => solution.ocrText, (val) => {
  if (val && val.trim()) {
    if (localText.value.trim()) {
      localText.value += '\n\n' + val.trim()
    } else {
      localText.value = val.trim()
    }
    nextTick(() => {
      if (textareaRef.value) {
        textareaRef.value.scrollTop = textareaRef.value.scrollHeight
      }
    })
  }
})

function handleEnter(e) {
  if (!e.shiftKey) {
    send()
  } else {
    localText.value += '\n'
  }
}

async function send() {
  const text = localText.value.trim()
  if (!text) return

  localText.value = ''
  // 通过 sendTextMessage 发送会作为当前对话的追问，保留上下文
  await api.sendTextMessage(text)
}
</script>

<style scoped>
.ocr-input-panel {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
  padding: var(--sp-3);
  background: var(--surface-card);
  border-top: 1px solid var(--border-subtle);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  pointer-events: auto;
}

.ocr-header {
  display: flex;
  align-items: center;
  gap: var(--sp-1-5);
  color: var(--text-secondary);
  font-size: var(--text-xs);
  font-weight: var(--weight-semibold);
}

.ocr-title {
  flex: 1;
}

.ocr-badge {
  background: rgba(59, 130, 246, 0.15);
  color: var(--color-info);
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  border: 1px solid rgba(59, 130, 246, 0.25);
}

.ocr-textarea {
  width: 100%;
  min-height: 72px;
  max-height: 160px;
  resize: vertical;
  background: var(--surface-input, rgba(0, 0, 0, 0.25));
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--sp-3);
  color: var(--text-primary);
  font-size: var(--text-sm);
  line-height: 1.5;
  font-family: var(--font-sans);
  outline: none;
  box-sizing: border-box;
  user-select: text;
  -webkit-user-select: text;
}

.ocr-textarea::placeholder {
  color: var(--text-muted);
}

.ocr-textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-light, rgba(16, 185, 129, 0.15));
}

.ocr-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ocr-hint {
  display: flex;
  align-items: center;
  gap: var(--sp-1);
  color: var(--text-tertiary);
  font-size: var(--text-xs);
}

.ocr-hint kbd {
  font-family: var(--font-mono);
  background: var(--surface-card-hover);
  border: 1px solid var(--border-default);
  padding: 1px 5px;
  border-radius: var(--radius-xs);
}

.ocr-send-btn {
  display: flex;
  align-items: center;
  gap: var(--sp-1-5);
  padding: var(--sp-2) var(--sp-4);
  border-radius: var(--radius-md);
  border: none;
  background: var(--accent);
  color: white;
  font-size: var(--text-sm);
  font-weight: var(--weight-semibold);
  cursor: pointer;
  transition: all var(--duration-fast) ease;
}

.ocr-send-btn:hover:not(:disabled) {
  background: var(--accent-hover);
}

.ocr-send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
