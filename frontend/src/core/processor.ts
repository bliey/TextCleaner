import { applyExtra, preClean } from './cleanrules'
import type { ProcessOptions, ProcessOutput, ReplaceRule } from './types'

function toLF(text: string): string {
  return text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
}

function detectLineEnding(text: string): 'lf' | 'crlf' {
  return text.includes('\r\n') || text.includes('\r') ? 'crlf' : 'lf'
}

function toEnding(text: string, ending: 'lf' | 'crlf'): string {
  return ending === 'crlf' ? text.replace(/\n/g, '\r\n') : text
}

function replaceAllLiteral(text: string, find: string, replacement: string): [string, number] {
  if (!find) return [text, 0]
  let count = 0
  let index = text.indexOf(find)
  while (index >= 0) {
    count++
    index = text.indexOf(find, index + find.length)
  }
  return [text.split(find).join(replacement), count]
}

function normalizeRegex(pattern: string): { pattern: string; flags: string } {
  let flags = 'gu'
  pattern = pattern.replace(/^\(\?i\)/u, () => { flags += 'i'; return '' })
  pattern = pattern.replace(/^\(\?m\)/u, () => { flags += 'm'; return '' })
  pattern = pattern.replace(/^\(\?s\)/u, () => { flags += 's'; return '' })
  return { pattern, flags: [...new Set(flags)].join('') }
}

function applyNormalRules(text: string, rules: ReplaceRule[]): [string, number, number] {
  let deleted = 0
  let replaced = 0
  for (const rule of rules) {
    if (!rule.enabled || rule.regex || !rule.find) continue
    const result = replaceAllLiteral(text, rule.find, rule.replace)
    text = result[0]
    if (rule.replace === '') deleted += result[1]
    else replaced += result[1]
  }
  return [text, deleted, replaced]
}

function applyRegexRules(text: string, rules: ReplaceRule[]): [string, number, number] {
  let deleted = 0
  let replaced = 0
  for (const rule of rules) {
    if (!rule.enabled || !rule.regex || !rule.find) continue
    try {
      const normalized = normalizeRegex(rule.find)
      const regex = new RegExp(normalized.pattern, normalized.flags)
      const matches = text.match(regex)
      if (!matches || matches.length === 0) continue
      const replacement = rule.replace.replace(/\$0/gu, '$&')
      text = text.replace(regex, replacement)
      if (rule.replace === '') deleted += matches.length
      else replaced += matches.length
    } catch {
      // Invalid JavaScript regex is skipped, matching the desktop's defensive behavior.
    }
  }
  return [text, deleted, replaced]
}

export function processText(input: string, options: ProcessOptions): ProcessOutput {
  const original = input
  let text = toLF(input)
  text = preClean(text, options.basicClean)
  const normal = applyNormalRules(text, options.replace)
  text = normal[0]
  const regex = applyRegexRules(text, options.replace)
  text = regex[0]
  text = applyExtra(text, options.basicClean)

  let ending = detectLineEnding(original)
  if (options.basicClean.normalizeLineEndings) {
    if (options.basicClean.lineEnding === 'lf') ending = 'lf'
    if (options.basicClean.lineEnding === 'crlf') ending = 'crlf'
  }
  let final = toEnding(text, ending)
  if (toLF(original).endsWith('\n') && final !== '') {
    const separator = ending === 'crlf' ? '\r\n' : '\n'
    if (!final.endsWith(separator)) final += separator
  }
  return {
    text: final,
    result: {
      deletedMatches: normal[1] + regex[1],
      replacedMatches: normal[2] + regex[2],
      changed: final !== original,
    },
  }
}
