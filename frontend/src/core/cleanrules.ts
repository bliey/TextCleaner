import cn2t from 'opencc-js/cn2t'
import t2cn from 'opencc-js/t2cn'
import { kangxiRadicalMap } from './kangxi'
import type { BasicCleanOptions } from './types'

const convertS2T = cn2t
const convertT2S = t2cn

const citeParenRE = /[ \t　]*[\(（][\d,，\-—–\.]+[\)）][ \t　]*/gu
const citeBracketRE = /[ \t　]*[\[【][\d,，\-—–\.]+[\]】][ \t　]*/gu
const decimalSpaceRE = /(\d)\s*\.\s+(\d)/gu
const colonSpaceRE = /(\d)\s*:\s+(\d)/gu
const letterDigitRE1 = /([A-Za-z])(\d)/gu
const letterDigitRE2 = /(\d)([A-Za-z])/gu
const punctSpaceRE = /([,!?;:])([A-Za-z\p{Script=Han}])/gu
const cjkToAsciiLetterRE = /(\p{Script=Han})([A-Za-z])/gu
const asciiLetterToCJKRE = /([A-Za-z])(\p{Script=Han})/gu
const cjkToAsciiDigitRE = /(\p{Script=Han})(\d)/gu
const asciiDigitToCJKRE = /(\d)(\p{Script=Han})/gu
const degreeRE = /(\d)\s*°/gu
const percentRE = /(\d)\s*(?:%|％)/gu
const asciiDecimalRE = /(\d)\.(\d)/gu
const asciiTimeRE = /(\d):(\d)/gu

const punctToEnglish: Record<string, string> = {
  '，': ',', '。': '.', '；': ';', '：': ':', '！': '!', '？': '?',
  '（': '(', '）': ')', '【': '[', '】': ']', '“': '"', '”': '"',
  '‘': "'", '’': "'",
}
const punctToChinese: Record<string, string> = {
  ',': '，', '.': '。', ';': '；', ':': '：', '!': '！', '?': '？',
  '(': '（', ')': '）', '[': '【', ']': '】', '"': '”', "'": '’',
}

function replaceRuneMap(text: string, map: Record<string, string>): string {
  let changed = false
  const out = [...text].map((char) => {
    const replacement = map[char]
    if (replacement === undefined) return char
    changed = true
    return replacement
  }).join('')
  return changed ? out : text
}

function fullToHalf(text: string): string {
  return [...text].map((char) => {
    const code = char.codePointAt(0)!
    if (code === 0x3000) return ' '
    if (code >= 0xff01 && code <= 0xff5e) return String.fromCodePoint(code - 0xfee0)
    return char
  }).join('')
}

function replaceKangxiRadicals(text: string): string {
  return [...text].map((char) => kangxiRadicalMap[char] ?? char).join('')
}

function collapseNewlines(text: string): string {
  text = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  let out = ''
  let previousNewline = false
  for (const char of text) {
    if (char === '\n') {
      if (previousNewline) continue
      previousNewline = true
    } else {
      previousNewline = false
    }
    out += char
  }
  return out
}

function newlinesToSpace(text: string): string {
  return text.replace(/\r\n/g, '\n').replace(/\r/g, '\n').replace(/\n/g, ' ')
}

function collapseSpaces(text: string): string {
  let out = ''
  let previousSpace = false
  for (const char of text) {
    if (char === ' ') {
      if (previousSpace) continue
      previousSpace = true
    } else {
      previousSpace = false
    }
    out += char
  }
  return out
}

function perLineClean(text: string, opt: BasicCleanOptions): string {
  return text.split('\n').map((line) => {
    if (opt.trimLeadingWhitespace) line = line.replace(/^[ \t]+/u, '')
    if (opt.trimTrailingWhitespace) line = line.replace(/[ \t]+$/u, '')
    if (opt.removeZeroWidthChars) line = line.replace(/[\u200b\u200c\u200d\ufeff]/gu, '')
    return line
  }).join('\n')
}

function removeEmptyLines(text: string): string {
  const lines = text.split('\n')
  return lines.filter((line) => line.trim() !== '').join('\n')
}

function collapseBlankLines(text: string, max: number): string {
  max = Math.max(0, max)
  const lines = text.split('\n')
  const nonBlank = lines.filter((line) => line.trim() !== '').length
  if (nonBlank === 0) return max > 0 ? '\n' : ''
  const out: string[] = []
  let blankRun = 0
  for (const line of lines) {
    if (line.trim() === '') {
      if (blankRun < max) out.push(line)
      blankRun++
    } else {
      blankRun = 0
      out.push(line)
    }
  }
  return out.join('\n')
}

function removeSpacesBetweenNonASCII(text: string): string {
  const chars = [...text]
  let out = ''
  for (let i = 0; i < chars.length; i++) {
    if (chars[i] === ' ' && i > 0 && i + 1 < chars.length &&
      chars[i - 1]!.codePointAt(0)! >= 0x80 && chars[i + 1]!.codePointAt(0)! >= 0x80) continue
    out += chars[i]
  }
  return out
}

function normalizeChineseTypography(text: string): string {
  text = text.replace(cjkToAsciiLetterRE, '$1 $2')
  text = text.replace(asciiLetterToCJKRE, '$1 $2')
  text = text.replace(cjkToAsciiDigitRE, '$1 $2')
  text = text.replace(asciiDigitToCJKRE, '$1 $2')
  text = text.replace(degreeRE, '$1°')
  text = text.replace(percentRE, '$1%')
  return text
}

function punctuationChineseToEnglish(text: string): string {
  text = replaceRuneMap(text, punctToEnglish)
  return text.replace(/…/gu, '...')
}

function punctuationEnglishToChinese(text: string): string {
  text = text.replace(/\.\.\./gu, '……')
  text = text.replace(asciiDecimalRE, '$1\u0001$2')
  text = text.replace(asciiTimeRE, '$1\u0002$2')
  text = replaceRuneMap(text, punctToChinese)
  return text.replace(/\u0001/g, '.').replace(/\u0002/g, ':')
}

function arbitrateMutex(opt: BasicCleanOptions): BasicCleanOptions {
  const copy = { ...opt }
  if (copy.punctEnglishToChinese && copy.punctChineseToEnglish) copy.punctEnglishToChinese = false
  if (copy.simplifiedToTraditional && copy.traditionalToSimplified) copy.simplifiedToTraditional = false
  if (copy.removeEmptyLines) {
    copy.collapseNewlines = false
    copy.collapseBlankLines = false
  } else if (copy.collapseNewlines) {
    copy.collapseBlankLines = false
  }
  return copy
}

export function preClean(text: string, opt: BasicCleanOptions): string {
  if (opt.removeUTF8BOM) text = text.replace(/^\ufeff/u, '')
  if (opt.trimLeadingWhitespace || opt.trimTrailingWhitespace || opt.removeZeroWidthChars) {
    text = perLineClean(text, opt)
  }
  if (opt.collapseSpaces) text = text.split('\n').map(collapseSpaces).join('\n')
  return text
}

export function applyExtra(text: string, sourceOpt: BasicCleanOptions): string {
  const opt = arbitrateMutex(sourceOpt)
  text = preClean(text, opt)
  if (opt.fullWidthToHalfWidth) text = fullToHalf(text)
  if (opt.replaceKangxiRadicals) text = replaceKangxiRadicals(text)
  if (opt.removeCiteParen) text = text.replace(citeParenRE, '')
  if (opt.removeCiteBracket) text = text.replace(citeBracketRE, '')
  if (opt.collapseNewlines) text = collapseNewlines(text)
  if (opt.newlineToSpace) text = newlinesToSpace(text)
  if (opt.removeSpaceBetweenCJK) text = removeSpacesBetweenNonASCII(text)
  if (opt.removeSpaceAtDecimal) text = text.replace(decimalSpaceRE, '$1.$2')
  if (opt.removeSpaceAtColon) text = text.replace(colonSpaceRE, '$1:$2')
  if (opt.spaceBetweenLetterAndDigit) {
    text = text.replace(letterDigitRE1, '$1 $2').replace(letterDigitRE2, '$1 $2')
  }
  if (opt.spaceAfterPunctuation) text = text.replace(punctSpaceRE, '$1 $2')
  if (opt.punctChineseToEnglish) text = punctuationChineseToEnglish(text)
  if (opt.punctEnglishToChinese) text = punctuationEnglishToChinese(text)
  if (opt.normalizeChineseTypography) text = normalizeChineseTypography(text)
  if (opt.simplifiedToTraditional) text = convertS2T(text)
  if (opt.traditionalToSimplified) text = convertT2S(text)
  if (opt.removeEmptyLines) text = removeEmptyLines(text)
  else if (opt.collapseBlankLines) text = collapseBlankLines(text, opt.maxBlankLines)
  return text
}
