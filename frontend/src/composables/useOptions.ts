import { computed, ref } from 'vue'
import type { BasicCleanOptions, ProcessOptions, ReplaceRule } from '../core/types'
import { useSettings } from './useSettings'

export type DeleteMode = 'line' | 'block'

function blankSimpleRule(): ReplaceRule {
  return { enabled: true, find: '', replace: '', regex: false }
}
function blankRegexRule(): ReplaceRule {
  return { enabled: false, find: '', replace: '', regex: true }
}

const { settings } = useSettings()
const basicClean: BasicCleanOptions = settings.basicClean
const deleteEnabled = ref(false)
const deleteMode = ref<DeleteMode>('line')
const deleteContent = ref('')
const replaceRules = ref<ReplaceRule[]>([blankSimpleRule()])
const advancedRules = ref<ReplaceRule[]>([blankRegexRule()])

function deleteToRules(): ReplaceRule[] {
  if (!deleteEnabled.value) return []
  const norm = deleteContent.value.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  if (norm.trim() === '') return []
  if (deleteMode.value === 'block') {
    return [{ enabled: true, find: norm, replace: '', regex: false }]
  }
  return norm.split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((find) => ({ enabled: true, find, replace: '', regex: false }))
}

const processOptions = computed<ProcessOptions>(() => ({
  basicClean: { ...basicClean },
  replace: [
    ...deleteToRules(),
    ...replaceRules.value,
    ...advancedRules.value,
  ].filter((rule) => rule.enabled && rule.find !== '').map((rule) => ({ ...rule })),
}))

function addSimpleRule() { replaceRules.value.push(blankSimpleRule()) }
function removeSimpleRule(index: number) {
  replaceRules.value.splice(index, 1)
  if (replaceRules.value.length === 0) replaceRules.value.push(blankSimpleRule())
}
function addRegexRule() { advancedRules.value.push(blankRegexRule()) }
function removeRegexRule(index: number) {
  advancedRules.value.splice(index, 1)
  if (advancedRules.value.length === 0) advancedRules.value.push(blankRegexRule())
}
function clearTextRules() {
  deleteEnabled.value = false
  deleteMode.value = 'line'
  deleteContent.value = ''
  replaceRules.value = [blankSimpleRule()]
  advancedRules.value = [blankRegexRule()]
}

export function useOptions() {
  return {
    basicClean,
    deleteEnabled,
    deleteMode,
    deleteContent,
    replaceRules,
    advancedRules,
    processOptions,
    addSimpleRule,
    removeSimpleRule,
    addRegexRule,
    removeRegexRule,
    clearTextRules,
  }
}
