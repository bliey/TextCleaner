export type {
  BasicCleanOptions,
  BrowserFileEntry,
  FileResult,
  BatchProgress,
  BatchSummary,
  LineEnding,
  OutputFormat,
  ProcessOptions,
  ProcessOutput,
  ProcessResult,
  ReplaceRule,
} from '../core/types'

export type FileEntry = import('../core/types').BrowserFileEntry

export type ThemePref = 'light' | 'dark' | 'system'
export type LanguagePref = 'auto' | 'zh' | 'en'

export interface Settings {
  basicClean: import('../core/types').BasicCleanOptions
  defaultIncludeSubfolders: boolean
  maxConcurrency: number
  theme: ThemePref
  language: LanguagePref
}
