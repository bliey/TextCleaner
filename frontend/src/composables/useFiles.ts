import { computed, ref } from 'vue'
import type { BrowserFileEntry } from '../core/types'

export const DEFAULT_EXTENSIONS = ['.txt', '.md', '.log', '.csv']

const files = ref<BrowserFileEntry[]>([])
const selected = ref<Record<string, boolean>>({})
const scanning = ref(false)
const errorMessage = ref('')
const includeSubfolders = ref(true)
let nextId = 1
let dropBound = false
const onDragOver = (event: DragEvent) => {
  event.preventDefault()
  document.querySelector('#app-root')?.classList.add('file-drop-target-active')
}
const onDragLeave = (event: DragEvent) => {
  if (event.relatedTarget === null) document.querySelector('#app-root')?.classList.remove('file-drop-target-active')
}
const onDrop = (event: DragEvent) => {
  document.querySelector('#app-root')?.classList.remove('file-drop-target-active')
  void handleDrop(event)
}

const selectedCount = computed(() => files.value.filter((file) => selected.value[file.path]).length)

function isSupported(file: File): boolean {
  const lower = file.name.toLowerCase()
  return DEFAULT_EXTENSIONS.some((extension) => lower.endsWith(extension))
}

function basename(path: string): string {
  const normalized = path.replace(/\\/g, '/')
  const i = normalized.lastIndexOf('/')
  return i >= 0 ? normalized.slice(i + 1) : normalized
}

function uniqueVirtualPath(path: string): string {
  const used = new Set(files.value.map((file) => file.path))
  if (!used.has(path)) return path
  const slash = path.lastIndexOf('/')
  const dir = slash >= 0 ? path.slice(0, slash + 1) : ''
  const base = slash >= 0 ? path.slice(slash + 1) : path
  const dot = base.lastIndexOf('.')
  const stem = dot > 0 ? base.slice(0, dot) : base
  const ext = dot > 0 ? base.slice(dot) : ''
  for (let i = 1; i < 10000; i++) {
    const candidate = `${dir}${stem} (${i})${ext}`
    if (!used.has(candidate)) return candidate
  }
  return `${Date.now()}-${path}`
}

function addFiles(inputFiles: File[]): void {
  const accepted = inputFiles.filter((file) => isSupported(file))
  const known = new Set(files.value.map((entry) => `${entry.file.name}:${entry.file.size}:${entry.file.lastModified}:${entry.file.webkitRelativePath}`))
  for (const file of accepted) {
    const relativePath = (file.webkitRelativePath || file.name).replace(/\\/g, '/')
    const segments = relativePath.split('/').filter(Boolean)
    // A webkitdirectory selection includes the selected folder name as the
    // first segment. When recursion is disabled, keep only its direct files.
    if (!includeSubfolders.value && segments.length > 2) continue
    const fingerprint = `${file.name}:${file.size}:${file.lastModified}:${file.webkitRelativePath}`
    if (known.has(fingerprint)) continue
    known.add(fingerprint)
    const path = uniqueVirtualPath(relativePath)
    files.value.push({
      id: String(nextId++),
      file,
      path,
      relativePath,
      name: basename(path),
      size: file.size,
      ext: file.name.includes('.') ? file.name.slice(file.name.lastIndexOf('.')).toLowerCase() : '',
    })
    selected.value[path] = true
  }
}

async function selectWithInput(mode: 'files' | 'directory'): Promise<void> {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.accept = DEFAULT_EXTENSIONS.join(',')
  if (mode === 'directory') {
    input.accept = DEFAULT_EXTENSIONS.join(',')
    ;(input as HTMLInputElement & { webkitdirectory?: boolean }).webkitdirectory = true
  }
  await new Promise<void>((resolve) => {
    input.addEventListener('change', () => {
      addFiles(Array.from(input.files ?? []))
      resolve()
    }, { once: true })
    input.click()
  })
}

async function pickFiles() {
  await selectWithInput('files')
}

async function pickFolder() {
  await selectWithInput('directory')
}

interface FileSystemEntryLike {
  isFile: boolean
  isDirectory: boolean
  name: string
  file?: (callback: (file: File) => void, error?: (error: DOMException) => void) => void
  createReader?: () => { readEntries: (callback: (entries: FileSystemEntryLike[]) => void, error?: (error: DOMException) => void) => void }
}

async function readEntry(entry: FileSystemEntryLike, prefix = ''): Promise<File[]> {
  if (entry.isFile && entry.file) {
    const file = await new Promise<File | null>((resolve) => entry.file!(resolve, () => resolve(null)))
    return file ? [file] : []
  }
  if (!entry.isDirectory || !entry.createReader) return []
  const reader = entry.createReader()
  const children: FileSystemEntryLike[] = []
  const readAll = (): Promise<void> => new Promise((resolve) => {
    reader.readEntries((batch) => {
      if (batch.length === 0) return resolve()
      children.push(...batch)
      void readAll().then(resolve)
    }, () => resolve())
  })
  await readAll()
  const nested = await Promise.all(children.map((child) => readEntry(child, `${prefix}${entry.name}/`)))
  return nested.flat()
}

async function handleDrop(event: DragEvent): Promise<void> {
  event.preventDefault()
  const items = Array.from(event.dataTransfer?.items ?? [])
  scanning.value = true
  errorMessage.value = ''
  try {
    const entries = items.map((item) => {
      const getEntry = (item as DataTransferItem & { webkitGetAsEntry?: () => FileSystemEntryLike | null }).webkitGetAsEntry
      return getEntry ? getEntry.call(item) : null
    }).filter((entry): entry is FileSystemEntryLike => Boolean(entry))
    if (entries.length > 0) {
      const dropped = (await Promise.all(entries.map((entry) => readEntry(entry)))).flat()
      addFiles(dropped)
    } else {
      addFiles(Array.from(event.dataTransfer?.files ?? []))
    }
  } finally {
    scanning.value = false
  }
}

function bindDropEvents() {
  if (dropBound) return
  dropBound = true
  window.addEventListener('dragover', onDragOver)
  window.addEventListener('dragleave', onDragLeave)
  window.addEventListener('drop', onDrop)
}
function unbindDropEvents() {
  if (!dropBound) return
  window.removeEventListener('dragover', onDragOver)
  window.removeEventListener('dragleave', onDragLeave)
  window.removeEventListener('drop', onDrop)
  document.querySelector('#app-root')?.classList.remove('file-drop-target-active')
  dropBound = false
}

function toggle(path: string) { selected.value[path] = !selected.value[path] }
function selectAll() { for (const file of files.value) selected.value[file.path] = true }
function deselectAll() { for (const file of files.value) selected.value[file.path] = false }
function remove(path: string) {
  files.value = files.value.filter((file) => file.path !== path)
  delete selected.value[path]
}
function clearAll() {
  files.value = []
  selected.value = {}
}

export function useFiles() {
  return {
    files,
    selected,
    scanning,
    errorMessage,
    includeSubfolders,
    selectedCount,
    pickFiles,
    pickFolder,
    addFiles,
    toggle,
    selectAll,
    deselectAll,
    remove,
    clearAll,
    bindDropEvents,
    unbindDropEvents,
  }
}
