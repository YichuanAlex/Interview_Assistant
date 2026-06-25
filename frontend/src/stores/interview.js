import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useInterviewStore = defineStore('interview', () => {
  // 实时转录文本列表
  const transcripts = ref([])
  // 面试提示
  const hint = ref('')
  // 是否正在生成提示
  const isGeneratingHint = ref(false)
  // 是否正在转录
  const isTranscribing = ref(false)
  // 转录状态/错误信息
  const statusMessage = ref('')
  // 当前选择的音频输入设备
  const inputDevice = ref(-1)
  // 可用音频设备列表
  const audioDevices = ref([])

  const transcriptText = computed(() => {
    return transcripts.value.map(t => `[${t.timestamp}] ${t.text}`).join('\n')
  })

  function addTranscript(timestamp, text, role = '') {
    transcripts.value.push({ timestamp, text, role })
    // 保留最近 100 条
    if (transcripts.value.length > 100) {
      transcripts.value = transcripts.value.slice(-100)
    }
  }

  function clearTranscripts() {
    transcripts.value = []
  }

  function setHint(value) {
    hint.value = value
  }

  function setGeneratingHint(value) {
    isGeneratingHint.value = value
  }

  function setTranscribing(value) {
    isTranscribing.value = value
  }

  function setStatusMessage(value) {
    statusMessage.value = value
  }

  function setAudioDevices(devices) {
    audioDevices.value = devices
  }

  function setInputDevice(device) {
    inputDevice.value = device
  }

  return {
    transcripts,
    hint,
    isGeneratingHint,
    isTranscribing,
    statusMessage,
    inputDevice,
    audioDevices,
    transcriptText,
    addTranscript,
    clearTranscripts,
    setHint,
    setGeneratingHint,
    setTranscribing,
    setStatusMessage,
    setAudioDevices,
    setInputDevice,
  }
})
