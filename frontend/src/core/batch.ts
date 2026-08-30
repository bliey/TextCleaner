import JSZip from 'jszip'
import { decodeText, encodeText, type OutputEncoding } from './encoding'
import { processText } from './processor'
import type {
  BatchProgress,
  BatchSummary,
  BrowserFileEntry,
  FileResult,
  ProcessOptions,
} from './types'

export interface BatchOptions {
  outputEncoding: OutputEncoding
  maxConcurrency: number
}

export interface BatchRunResult {
  summary: BatchSummary
  zipBlob: Blob
  outputFiles: Array<{ path: string; data: Uint8Array }>
}

function normalizeZipPath(path: string): string {
  return path.replace(/\\/g, '/').replace(/^\/+/, '')
}

function uniquePath(path: string, used: Set<string>): string {
  const normalized = normalizeZipPath(path) || 'unnamed.txt'
  if (!used.has(normalized)) {
    used.add(normalized)
    return normalized
  }
  const slash = normalized.lastIndexOf('/')
  const dir = slash >= 0 ? normalized.slice(0, slash + 1) : ''
  const base = slash >= 0 ? normalized.slice(slash + 1) : normalized
  const dot = base.lastIndexOf('.')
  const stem = dot > 0 ? base.slice(0, dot) : base
  const ext = dot > 0 ? base.slice(dot) : ''
  for (let i = 1; i < 10000; i++) {
    const candidate = `${dir}${stem} (${i})${ext}`
    if (!used.has(candidate)) {
      used.add(candidate)
      return candidate
    }
  }
  return normalized
}

function basename(path: string): string {
  const normalized = normalizeZipPath(path)
  const i = normalized.lastIndexOf('/')
  return i >= 0 ? normalized.slice(i + 1) : normalized
}

export async function processBatch(
  entries: BrowserFileEntry[],
  processOptions: ProcessOptions,
  options: BatchOptions,
  onProgress: (progress: BatchProgress) => void,
  signal: AbortSignal,
): Promise<BatchRunResult> {
  const selected = entries.slice()
  const results: FileResult[] = new Array(selected.length)
  const outputFiles: Array<{ path: string; data: Uint8Array }> = []
  const usedPaths = new Set<string>()
  const concurrency = Math.max(1, Math.min(16, Math.round(options.maxConcurrency || 4)))
  let nextIndex = 0
  let done = 0

  const worker = async (): Promise<void> => {
    while (true) {
      if (signal.aborted) return
      const index = nextIndex++
      if (index >= selected.length) return
      const entry = selected[index]!
      try {
        const decoded = decodeText(new Uint8Array(await entry.file.arrayBuffer()))
        if (signal.aborted) return
        const processed = processText(decoded.text, processOptions)
        const data = encodeText(processed.text, options.outputEncoding, decoded.encoding)
        const outputPath = uniquePath(entry.relativePath || entry.name, usedPaths)
        outputFiles[index] = { path: outputPath, data }
        results[index] = {
          path: entry.path,
          success: true,
          deleted: processed.result.deletedMatches,
          replaced: processed.result.replacedMatches,
          changed: processed.result.changed,
        }
      } catch (error) {
        results[index] = {
          path: entry.path,
          success: false,
          error: error instanceof Error ? error.message : String(error),
          deleted: 0,
          replaced: 0,
          changed: false,
        }
      } finally {
        done++
        onProgress({
          done,
          total: selected.length,
          current: basename(entry.relativePath || entry.name),
          last: results[index],
        })
      }
    }
  }

  await Promise.all(Array.from({ length: Math.min(concurrency, selected.length) }, () => worker()))
  if (signal.aborted) throw new DOMException('Processing cancelled', 'AbortError')

  const zip = new JSZip()
  for (const output of outputFiles) {
    if (output) zip.file(output.path, output.data)
  }
  const zipBlob = await zip.generateAsync({ type: 'blob', compression: 'DEFLATE' })
  const succeeded = results.filter((result) => result?.success).length
  const failed = results.length - succeeded
  return {
    summary: {
      total: selected.length,
      processed: results.length,
      succeeded,
      failed,
      results,
      zipName: 'TextCleaner_Output.zip',
    },
    zipBlob,
    outputFiles: outputFiles.filter((output): output is { path: string; data: Uint8Array } => Boolean(output)),
  }
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  setTimeout(() => URL.revokeObjectURL(url), 0)
}
