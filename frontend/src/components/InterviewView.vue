<template>
  <div class="interview-view">
    <div class="interview-toolbar">
      <div class="toolbar-left">
        <button
          class="btn btn-primary"
          :class="{ 'btn-stop': interview.isTranscribing }"
          @click="toggleTranscription"
        >
          <Icon :name="interview.isTranscribing ? 'mic-off' : 'mic'" :size="16" />
          <span>{{ interview.isTranscribing ? '停止转录' : '开始转录' }}</span>
        </button>
        <span v-if="interview.isTranscribing" class="status-badge live">转录中</span>
        <span v-else class="status-badge stopped">已停止</span>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-accent" :disabled="interview.isGeneratingHint" @click="generateHint">
          <Icon name="sparkles" :size="16" />
          <span>{{ interview.isGeneratingHint ? '生成中...' : '生成提示' }}</span>
        </button>
        <button class="btn btn-ghost" @click="clearContext">
          <Icon name="trash-2" :size="16" />
          <span>清空</span>
        </button>
      </div>
    </div>

    <div class="interview-status" v-if="interview.statusMessage">
      {{ interview.statusMessage }}
    </div>

    <div class="interview-content">
      <div class="panel transcripts-panel">
        <div class="panel-header">
          <Icon name="message-square" :size="16" />
          <span>实时转录</span>
        </div>
        <div ref="transcriptBox" class="transcript-box">
          <div v-if="interview.transcripts.length === 0" class="empty-state">
            点击「开始转录」后，面试官和面试者的语音会实时显示在这里。
          </div>
          <div
            v-for="(t, i) in interview.transcripts"
            :key="i"
            class="transcript-item"
            :class="{ latest: i === interview.transcripts.length - 1, interviewer: t.role === 'interviewer', interviewee: t.role === 'interviewee' }"
          >
            <span class="transcript-role">{{ t.role === 'interviewee' ? '我' : '面试官' }}</span>
            <span class="transcript-time">{{ t.timestamp }}</span>
            <span class="transcript-text">{{ t.text }}</span>
          </div>
        </div>
      </div>

      <div class="panel hint-panel">
        <div class="panel-header">
          <Icon name="lightbulb" :size="16" />
          <span>面试提示</span>
        </div>
        <div class="hint-box">
          <div v-if="!interview.hint && !interview.isGeneratingHint" class="empty-state">
            点击「生成提示」后，AI 会根据上方实时转录内容给出面试回答思路。
          </div>
          <div v-else-if="interview.isGeneratingHint" class="hint-loading">
            正在分析上下文并生成提示...
          </div>
          <div v-else class="hint-text">{{ interview.hint }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from 'vue'
import { useInterviewStore } from '../stores/interview'
import { useUIStore } from '../stores/ui'
import { api } from '../services/api'
import { on } from '../services/events'
import Icon from './Icon.vue'

const interview = useInterviewStore()
const ui = useUIStore()
const transcriptBox = ref(null)

watch(() => interview.transcripts.length, () => {
  nextTick(() => {
    if (transcriptBox.value) {
      transcriptBox.value.scrollTop = transcriptBox.value.scrollHeight
    }
  })
})

onMounted(() => {
  // 监听转录文本
  on('transcription', (timestamp, text, role) => {
    interview.addTranscript(timestamp, text, role)
  })

  // 监听自动生成的面试提示
  on('interview-hint', (hint) => {
    interview.setHint(hint || '')
  })

  // 监听转录状态
  on('transcription-status', (status, err) => {
    if (status === 'started') {
      interview.setTranscribing(true)
      interview.setStatusMessage('实时转录中...')
      ui.showToast('实时转录已启动', 'success', 1500)
    } else if (status === 'stopped') {
      interview.setTranscribing(false)
      interview.setStatusMessage('')
    } else if (status === 'error') {
      interview.setTranscribing(false)
      interview.setStatusMessage('转录出错: ' + (err || '未知错误'))
      ui.showToast('转录出错', 'error', 2000)
    }
  })
})

async function toggleTranscription() {
  if (interview.isTranscribing) {
    const err = await api.stopTranscription()
    if (err) {
      ui.showToast(err, 'error', 2000)
    }
  } else {
    // 面试官：系统音频/会议软件；面试者：MacBook Pro 麦克风
    const err = await api.startTranscription('OrayVirtualAudioDevice', 'MacBook Pro', './models/small', 'zh')
    if (err) {
      interview.setStatusMessage('启动失败: ' + err)
      ui.showToast('启动转录失败', 'error', 3000)
    }
  }
}

async function generateHint() {
  if (interview.isGeneratingHint) return
  interview.setGeneratingHint(true)
  try {
    const hint = await api.generateInterviewHint()
    interview.setHint(hint)
  } catch (e) {
    ui.showToast('生成提示失败', 'error', 2000)
  } finally {
    interview.setGeneratingHint(false)
  }
}

async function clearContext() {
  await api.clearInterviewContext()
  interview.clearTranscripts()
  interview.setHint('')
  interview.setStatusMessage('')
  ui.showToast('已清空面试上下文', 'info', 1500)
}
</script>

<style scoped>
.interview-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  pointer-events: auto;
}

.interview-view button,
.interview-view select {
  pointer-events: auto;
}

.interview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sp-3);
  padding: var(--sp-3) var(--sp-4);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
}

.status-badge {
  font-size: var(--text-xs);
  padding: var(--sp-1) var(--sp-2);
  border-radius: var(--radius-md);
  font-weight: 500;
}

.status-badge.live {
  background: rgba(40, 200, 64, 0.15);
  color: #28c840;
}

.status-badge.stopped {
  background: var(--surface-hover);
  color: var(--text-muted);
}

.interview-status {
  padding: var(--sp-2) var(--sp-4);
  font-size: var(--text-xs);
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.interview-content {
  display: flex;
  flex: 1;
  gap: var(--sp-3);
  padding: var(--sp-3);
  overflow: hidden;
}

.panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--surface-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3) var(--sp-4);
  border-bottom: 1px solid var(--border-subtle);
  font-weight: 600;
  font-size: var(--text-sm);
  color: var(--text-primary);
  flex-shrink: 0;
}

.transcript-box,
.hint-box {
  flex: 1;
  overflow-y: auto;
  padding: var(--sp-3) var(--sp-4);
}

.transcript-item {
  display: flex;
  gap: var(--sp-3);
  padding: var(--sp-2) 0;
  font-size: var(--text-sm);
  line-height: 1.6;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.transcript-item.latest {
  color: var(--text-primary);
}

.transcript-role {
  flex-shrink: 0;
  font-size: var(--text-xs);
  font-weight: 600;
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  background: var(--surface-hover);
  color: var(--text-muted);
}

.transcript-item.interviewer .transcript-role {
  background: rgba(254, 188, 46, 0.15);
  color: #febc2e;
}

.transcript-item.interviewee .transcript-role {
  background: rgba(0, 122, 255, 0.15);
  color: #007aff;
}

.transcript-time {
  flex-shrink: 0;
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: var(--text-xs);
}

.transcript-text {
  user-select: text;
  -webkit-user-select: text;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--text-muted);
  font-size: var(--text-sm);
  line-height: 1.6;
}

.hint-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
}

.hint-text {
  font-size: var(--text-sm);
  line-height: 1.8;
  color: var(--text-primary);
  white-space: pre-wrap;
  user-select: text;
  -webkit-user-select: text;
}

.btn-stop {
  background: var(--danger) !important;
  border-color: var(--danger) !important;
}
</style>
