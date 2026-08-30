import { computed, ref } from 'vue'
import { processBatch, downloadBlob, type BatchRunResult } from '../core/batch'
import type { BatchProgress, BatchSummary, OutputFormat } from '../core/types'
import { isDesktop } from '../core/env'
import { useFiles } from './useFiles'
import { useOptions } from './useOptions'
import { useSettings } from './useSettings'

export type { BatchProgress, BatchSummary }

const { settings } = useSettings()
const outputEncoding = ref<'keep' | 'utf-8' | 'utf-8-bom'>('keep')
const outputFormat = ref<OutputFormat>('zip')
const running = ref(false)
const progress = ref<BatchProgress | null>(null)
const summary = ref<BatchSummary | null>(null)
const errorMsg = ref('')
const exportReady = ref(false)
const lastRun = ref<BatchRunResult | null>(null)
let controller: AbortController | null = null

const maxConcurrency = computed({
  get: () => settings.maxConcurrency,
  set: (value: number) => {
    const n = typeof value === 'number' && !Number.isNaN(value) ? Math.round(value) : 4
    settings.maxConcurrency = Math.min(16, Math.max(1, n))
  },
})

export interface PendingPlan {
  fileCount: number
  outputFormat: OutputFormat
}
const pendingPlan = ref<PendingPlan | null>(null)

const { files, selected } = useFiles()
const { processOptions } = useOptions()
const selectedEntries = computed(() => files.value.filter((file) => selected.value[file.path]))
const previewFileCount = computed(() => selectedEntries.value.length)

function start() {
  if (running.value || selectedEntries.value.length === 0) return
  pendingPlan.value = {
    fileCount: selectedEntries.value.length,
    outputFormat: outputFormat.value,
  }
  errorMsg.value = ''
}

async function confirmStart() {
  const plan = pendingPlan.value
  if (!plan || running.value) return
  pendingPlan.value = null
  running.value = true
  progress.value = { done: 0, total: plan.fileCount, current: '' }
  summary.value = null
  errorMsg.value = ''
  exportReady.value = false
  lastRun.value = null
  controller = new AbortController()

  try {
    const run = await processBatch(
      selectedEntries.value,
      processOptions.value,
      {
        outputEncoding: outputEncoding.value,
        maxConcurrency: maxConcurrency.value,
      },
      (next) => { progress.value = next },
      controller.signal,
    )
    lastRun.value = run
    summary.value = run.summary
    exportReady.value = run.summary.succeeded > 0
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      errorMsg.value = '处理已取消。'
    } else {
      errorMsg.value = error instanceof Error ? error.message : String(error)
    }
  } finally {
    running.value = false
    controller = null
  }
}

function cancel() {
  controller?.abort()
}

async function exportResults() {
  const run = lastRun.value
  if (!run || run.summary.succeeded === 0) return
  if (isDesktop()) {
    // 桌面端（Windows/macOS）：系统保存对话框 + Go 原生写盘，不使用浏览器下载 API
    const { exportToDisk } = await import('../core/desktop')
    try {
      await exportToDisk(run, outputFormat.value)
    } catch (error) {
      errorMsg.value = error instanceof Error ? error.message : String(error)
    }
    return
  }
  // Web 端：保留浏览器下载逻辑
  if (outputFormat.value === 'zip') {
    downloadBlob(run.zipBlob, 'TextCleaner_Output.zip')
    return
  }
  for (const file of run.outputFiles) {
    downloadBlob(new Blob([file.data], { type: 'text/plain;charset=utf-8' }), file.path.split('/').pop() || 'output.txt')
  }
}

function cancelPending() { pendingPlan.value = null }
function clearResults() {
  summary.value = null
  errorMsg.value = ''
  progress.value = null
  lastRun.value = null
  exportReady.value = false
}

export function useBatch() {
  return {
    outputEncoding,
    outputFormat,
    maxConcurrency,
    running,
    progress,
    summary,
    errorMsg,
    exportReady,
    selectedEntries,
    previewFileCount,
    start,
    confirmStart,
    cancelPending,
    cancel,
    exportResults,
    clearResults,
    pendingPlan,
  }
}