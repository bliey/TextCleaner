import { reactive, ref, watch } from 'vue'
import type { BasicCleanOptions, Settings, ThemePref } from '../types/models'
import { setLang, langFromPref } from '../i18n'

const SETTINGS_KEY = 'textcleaner.settings.v1'

function defaultBasicClean(): BasicCleanOptions {
  return {
    trimLeadingWhitespace: true,
    trimTrailingWhitespace: true,
    removeUTF8BOM: true,
    removeZeroWidthChars: true,
    collapseSpaces: false,
    fullWidthToHalfWidth: false,
    replaceKangxiRadicals: false,
    punctEnglishToChinese: false,
    punctChineseToEnglish: false,
    removeCiteParen: false,
    removeCiteBracket: false,
    collapseNewlines: false,
    newlineToSpace: false,
    removeEmptyLines: false,
    collapseBlankLines: true,
    maxBlankLines: 1,
    removeSpaceBetweenCJK: false,
    spaceAfterPunctuation: false,
    removeSpaceAtDecimal: false,
    spaceBetweenLetterAndDigit: false,
    removeSpaceAtColon: false,
    normalizeChineseTypography: false,
    simplifiedToTraditional: false,
    traditionalToSimplified: false,
    normalizeLineEndings: false,
    lineEnding: 'keep',
  }
}

function defaultSettings(): Settings {
  return {
    basicClean: defaultBasicClean(),
    defaultIncludeSubfolders: true,
    maxConcurrency: 4,
    theme: 'system',
    language: 'auto',
  }
}

const settings = reactive<Settings>(defaultSettings())
const ready = ref(false)
const saving = ref(false)
const saveError = ref('')
const loadError = ref('')

type ResolvedTheme = 'light' | 'dark'
function resolveTheme(pref: string): ResolvedTheme {
  if (pref === 'light') return 'light'
  if (pref === 'dark') return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}
function applyTheme() {
  document.documentElement.setAttribute('data-theme', resolveTheme(settings.theme))
}
function applyLanguage() {
  setLang(langFromPref(settings.language))
}

async function persist() {
  saving.value = true
  saveError.value = ''
  try {
    localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings))
  } catch (error) {
    saveError.value = error instanceof Error ? error.message : String(error)
  } finally {
    saving.value = false
  }
}

async function load() {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as Partial<Settings>
      if (parsed.basicClean) Object.assign(settings.basicClean, parsed.basicClean)
      if (typeof parsed.defaultIncludeSubfolders === 'boolean') settings.defaultIncludeSubfolders = parsed.defaultIncludeSubfolders
      if (typeof parsed.maxConcurrency === 'number') settings.maxConcurrency = parsed.maxConcurrency
      if (parsed.theme === 'light' || parsed.theme === 'dark' || parsed.theme === 'system') settings.theme = parsed.theme
      if (parsed.language === 'auto' || parsed.language === 'zh' || parsed.language === 'en') settings.language = parsed.language
    }
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : String(error)
  } finally {
    ready.value = true
    applyTheme()
    applyLanguage()
  }
}

function reset() {
  const defaults = defaultSettings()
  Object.assign(settings.basicClean, defaults.basicClean)
  settings.defaultIncludeSubfolders = defaults.defaultIncludeSubfolders
  settings.maxConcurrency = defaults.maxConcurrency
  settings.theme = defaults.theme
  settings.language = defaults.language
}

watch(settings, () => {
  if (ready.value) void persist()
}, { deep: true })
watch(() => settings.theme, () => { if (ready.value) applyTheme() })
watch(() => settings.language, () => { if (ready.value) applyLanguage() })

export function useSettings() {
  return {
    settings,
    ready,
    saving,
    saveError,
    loadError,
    load,
    persist,
    reset,
    applyTheme,
    resolveTheme: (pref: ThemePref) => resolveTheme(pref),
  }
}
