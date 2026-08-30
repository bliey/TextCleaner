import { computed, ref } from 'vue'
import { readFileText } from '../core/encoding'
import { processText } from '../core/processor'
import { useFiles } from './useFiles'
import { useOptions } from './useOptions'

// ============================================================
// 预览状态与 Diff 计算
//
// 核心约束：预览必须与批量处理使用完全相同的 TypeScript 核心，
// 绝不能在前端另写一套替换实现。因此这里只做三件事：
//   1. 通过 Browser File API 读取原文；
//   2. 调用 core/processText 取结果；
//   3. 在前端对“两串文本”做行级 Diff —— 这属于展示层计算，不是文本处理逻辑。
// ============================================================

/**
 * 预览读取上限（字符数）。
 * 目的：避免把几十 MB 的小说合集整篇拉进 WebView 内存与 Diff 计算，
 * 导致界面卡死。注意此限制只影响「预览展示」，
 * 批量处理始终读取并处理完整文件。
 */
export const PREVIEW_MAX_CHARS = 200_000

/**
 * Diff 参与计算的最大行数。
 * LCS 是 O(n*m) 时间/空间，超出此规模直接不再计算 diff，
 * 由界面降级为双栏对照。
 */
export const DIFF_MAX_LINES = 2000

/** LCS 回溯的单元上限，超过则退化为「整段删除 + 整段新增」。 */
const LCS_CELL_BUDGET = 4_000_000

export type DiffType = 'same' | 'del' | 'add'

export interface DiffRow {
  type: DiffType
  text: string
}

export interface PreviewStats {
  deletedMatches: number
  replacedMatches: number
  changed: boolean
}

const { files } = useFiles()
const { processOptions } = useOptions()

const previewPath = ref('')
const originalText = ref('')
const resultText = ref('')
const stats = ref<PreviewStats | null>(null)
const loading = ref(false)
const errorMsg = ref('')
const truncated = ref(false)

const hasPreview = computed(() => previewPath.value !== '')

/**
 * 生成预览：读取 File → 交给统一 processText → 保存结果。
 * 预览用的 processOptions 与批量处理同源，保证「所见即所得」。
 */
async function preview(path: string): Promise<void> {
  if (!path) return
  const entry = files.value.find((file) => file.path === path)
  if (!entry) return
  loading.value = true
  errorMsg.value = ''
  try {
    const decoded = await readFileText(entry.file)
    const isTruncated = decoded.text.length > PREVIEW_MAX_CHARS
    // 截断只影响展示；仍按截断后的文本调用统一核心，确保所见即所得。
    const source = isTruncated ? decoded.text.slice(0, PREVIEW_MAX_CHARS) : decoded.text
    const out = processText(source, processOptions.value)

    previewPath.value = path
    originalText.value = source
    resultText.value = out.text
    stats.value = out.result
    truncated.value = isTruncated
  } catch (e) {
    errorMsg.value = e instanceof Error ? e.message : String(e)
    previewPath.value = path
    originalText.value = ''
    resultText.value = ''
    stats.value = null
  } finally {
    loading.value = false
  }
}

/** 按当前规则重新预览同一个文件（规则改动后调用）。 */
async function refresh(): Promise<void> {
  if (previewPath.value) await preview(previewPath.value)
}

function clear(): void {
  previewPath.value = ''
  originalText.value = ''
  resultText.value = ''
  stats.value = null
  errorMsg.value = ''
  truncated.value = false
}

// ------------------------------------------------------------
// 行级 Diff（LCS）
// ------------------------------------------------------------

function splitLines(text: string): string[] {
  // 统一按 \n 切分：原文的 \r\n 已在 Go 端归一化处理过，
  // 这里仅作展示层的兜底。
  return text.replace(/\r\n?/g, '\n').split('\n')
}

/**
 * 对两段“中间差异区”做标准 LCS DP。
 * 使用 Uint32Array 降低内存占用（JS number 数组开销约为其 8 倍）。
 */
function lcsDiff(a: string[], b: string[]): DiffRow[] {
  const n = a.length
  const m = b.length
  if (n === 0 && m === 0) return []
  if (n === 0) return b.map((text) => ({ type: 'add' as const, text }))
  if (m === 0) return a.map((text) => ({ type: 'del' as const, text }))

  // 规模过大时放弃精确 diff，直接整段替换，避免界面长时间无响应。
  if (n * m > LCS_CELL_BUDGET) {
    return [
      ...a.map((text) => ({ type: 'del' as const, text })),
      ...b.map((text) => ({ type: 'add' as const, text })),
    ]
  }

  const dp: Uint32Array[] = Array.from(
    { length: n + 1 },
    () => new Uint32Array(m + 1),
  )
  for (let i = n - 1; i >= 0; i--) {
    const row = dp[i]!
    const next = dp[i + 1]!
    for (let j = m - 1; j >= 0; j--) {
      row[j] = a[i] === b[j] ? next[j + 1]! + 1 : Math.max(next[j]!, row[j + 1]!)
    }
  }

  const out: DiffRow[] = []
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (a[i] === b[j]) {
      out.push({ type: 'same', text: a[i]! })
      i++
      j++
    } else if (dp[i + 1]![j]! >= dp[i]![j + 1]!) {
      out.push({ type: 'del', text: a[i]! })
      i++
    } else {
      out.push({ type: 'add', text: b[j]! })
      j++
    }
  }
  while (i < n) out.push({ type: 'del', text: a[i++]! })
  while (j < m) out.push({ type: 'add', text: b[j++]! })
  return out
}

/**
 * 行级 diff 入口。
 * 先剥离公共前后缀再进入 O(n*m) 的 DP —— 实际场景（删除广告段、替换词）
 * 差异通常集中在局部，这一步能把 DP 规模压掉几个数量级。
 */
function diffLines(a: string[], b: string[]): DiffRow[] {
  const out: DiffRow[] = []

  let start = 0
  while (start < a.length && start < b.length && a[start] === b[start]) {
    out.push({ type: 'same', text: a[start]! })
    start++
  }

  let endA = a.length
  let endB = b.length
  while (endA > start && endB > start && a[endA - 1] === b[endB - 1]) {
    endA--
    endB--
  }

  out.push(...lcsDiff(a.slice(start, endA), b.slice(start, endB)))

  for (let i = endA; i < a.length; i++) out.push({ type: 'same', text: a[i]! })
  return out
}

/**
 * 差异行。
 * 返回 null 表示“文件过大，不计算 diff”，界面会降级为双栏对照并给出提示。
 */
const diffRows = computed<DiffRow[] | null>(() => {
  if (!hasPreview.value) return null
  const a = splitLines(originalText.value)
  const b = splitLines(resultText.value)
  if (a.length > DIFF_MAX_LINES || b.length > DIFF_MAX_LINES) return null
  return diffLines(a, b)
})

/** 差异过大（DP 超预算）时，双栏视图更实用，提示用户切换。 */
const diffSkipped = computed(() => hasPreview.value && diffRows.value === null)

const diffStat = computed(() => {
  const rows = diffRows.value ?? []
  let del = 0
  let add = 0
  for (const r of rows) {
    if (r.type === 'del') del++
    else if (r.type === 'add') add++
  }
  return { del, add }
})

export function usePreview() {
  return {
    files,
    previewPath,
    originalText,
    resultText,
    stats,
    loading,
    errorMsg,
    truncated,
    hasPreview,
    diffRows,
    diffStat,
    diffSkipped,
    preview,
    refresh,
    clear,
  }
}
