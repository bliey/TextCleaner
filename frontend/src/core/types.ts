export type LineEnding = 'keep' | 'lf' | 'crlf'

export interface BasicCleanOptions {
  trimLeadingWhitespace: boolean
  trimTrailingWhitespace: boolean
  removeUTF8BOM: boolean
  removeZeroWidthChars: boolean
  collapseSpaces: boolean
  fullWidthToHalfWidth: boolean
  replaceKangxiRadicals: boolean
  punctEnglishToChinese: boolean
  punctChineseToEnglish: boolean
  removeCiteParen: boolean
  removeCiteBracket: boolean
  collapseNewlines: boolean
  newlineToSpace: boolean
  removeEmptyLines: boolean
  collapseBlankLines: boolean
  maxBlankLines: number
  removeSpaceBetweenCJK: boolean
  spaceAfterPunctuation: boolean
  removeSpaceAtDecimal: boolean
  spaceBetweenLetterAndDigit: boolean
  removeSpaceAtColon: boolean
  normalizeChineseTypography: boolean
  simplifiedToTraditional: boolean
  traditionalToSimplified: boolean
  normalizeLineEndings: boolean
  lineEnding: LineEnding
}

export interface ReplaceRule {
  enabled: boolean
  find: string
  replace: string
  regex: boolean
}

export interface ProcessOptions {
  basicClean: BasicCleanOptions
  replace: ReplaceRule[]
}

export interface ProcessResult {
  deletedMatches: number
  replacedMatches: number
  changed: boolean
}

export interface ProcessOutput {
  text: string
  result: ProcessResult
}

export interface BrowserFileEntry {
  id: string
  file: File
  /** User-visible virtual path, unique within the current selection. */
  path: string
  /** Path written into the downloaded ZIP. */
  relativePath: string
  name: string
  size: number
  ext: string
}

export interface BatchProgress {
  done: number
  total: number
  current: string
  last?: FileResult
}

export interface FileResult {
  path: string
  success: boolean
  error?: string
  deleted: number
  replaced: number
  changed: boolean
}

export interface BatchSummary {
  total: number
  processed: number
  succeeded: number
  failed: number
  results: FileResult[]
  zipName?: string
}

export type OutputFormat = 'zip' | 'files'
