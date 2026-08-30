import { Buffer } from 'buffer'
import type { BatchRunResult } from './batch'
import type { OutputFormat } from './types'

function toBase64(data: Uint8Array): string {
  return Buffer.from(data).toString('base64')
}

/**
 * 桌面端（Windows/macOS 的 Wails 应用）原生导出：
 * 通过 Go 端的系统对话框选择保存位置，再由 Go 端直接写入文件系统，
 * 不使用任何浏览器下载 API（URL.createObjectURL / <a download> 等）。
 */
export async function exportToDisk(run: BatchRunResult, format: OutputFormat): Promise<void> {
  const { AppService } = await import('../../bindings/textcleaner/index.js')

  if (format === 'zip') {
    const path = await AppService.ChooseSavePath('TextCleaner_Output.zip')
    if (!path) return
    const bytes = new Uint8Array(await run.zipBlob.arrayBuffer())
    await AppService.SaveFileBytes(path, toBase64(bytes))
    return
  }

  const files = run.outputFiles
  if (files.length === 1) {
    const name = files[0].path.split('/').pop() || 'output.txt'
    const path = await AppService.ChooseSavePath(name)
    if (!path) return
    await AppService.SaveFileBytes(path, toBase64(files[0].data))
    return
  }

  if (files.length > 1) {
    const dir = await AppService.ChooseDirectory()
    if (!dir) return
    await AppService.ExportFiles(dir, files.map((f) => ({ path: f.path, data: toBase64(f.data) })))
  }
}