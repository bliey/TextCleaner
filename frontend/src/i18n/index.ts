import { ref } from 'vue'

// ============================================================
// 国际化（i18n）层
// 设计：单一响应式语言状态 currentLang；t(key, params) 即时翻译。
// 约定：字典以“平面键”组织（如 'app.start'），{n} 占位符通过 params 替换。
// 依赖方向：本模块不依赖任何 composable，避免循环引用。
// ============================================================

export type Lang = 'zh' | 'en'

type Dict = Record<string, string>

const zh: Dict = {
  // ---- 应用外壳 / 操作栏 ----
  'app.scanning': '正在扫描文件…',
  'app.running': '批量处理进行中…',
  'app.selected': '已选择 {n} 个文件 · 核心处理已在 Go 端验证',
  'app.start': '开始批量处理',
  'app.cancel': '取消',
  'app.settings': '设置',
  'app.offline': '（离线）',

  // ---- 交互锁（无输入时区块置灰提示） ----
  'ui.locked': '需先选择文件',
  'ui.lockedHint': '选择文件后即可编辑这些选项',

  // ---- 文件选择 ----
  'file.title': '文件选择',
  'file.hint': '支持 .txt / .md / .log / .csv',
  'file.dropMain': '将文件或文件夹拖入此窗口',
  'file.dropSub': '或点击下方按钮选择（支持多选）',
  'file.pickFiles': '选择文件…',
  'file.pickFolder': '选择文件夹…',
  'file.includeSubfolders': '包含子文件夹',

  // ---- 文件列表 ----
  'list.title': '文件列表',
  'list.selected': '已选择 {n} / {m} 个文件',
  'list.search': '搜索文件名或路径…',
  'list.sort': '排序',
  'list.name': '名称',
  'list.path': '路径',
  'list.size': '大小',
  'list.selectAll': '全选',
  'list.deselectAll': '取消全选',
  'list.clear': '清空',
  'list.empty': '尚未加载文件。请在上方选择文件或文件夹，或拖入文本文件。',
  'list.emptyFilter': '没有匹配 “{q}” 的文件。',
  'list.remove': '移除',
  'list.removeTitle': '移除该文件',

  // ---- 基础清理 ----
  'basic.title': '基础清理',
  'basic.hint': '通用选项，会被记住',
  'basic.trimLeading': '删除行首空白',
  'basic.trimTrailing': '删除行尾空白',
  'basic.removeEmptyLines': '删除完全空白的行',
  'basic.collapseBlank': '合并连续空行',
  'basic.collapseSpaces': '合并连续空格',
  'basic.removeZeroWidth': '删除零宽字符',
  'basic.normalizeLE': '统一换行符',
  'basic.removeBOM': '删除 UTF-8 BOM',
  'basic.maxBlank': '合并后保留的最大连续空行数',
  'basic.lineEnding': '目标换行符',

  // ---- 基础功能（24 个开关 + 2 个子选项，合并自原"基础清理"+"基础功能"） ----
  'feature.title': '基础功能',
  'feature.hint': '一键清理：字符、标点、引用、空白、CJK 间距、中文排版、简繁转换等 24 项常用规则',
  'feature.groupChar': '字符',
  'feature.groupPunct': '标点',
  'feature.groupCite': '引用角标',
  'feature.groupWhitespace': '空白与换行',
  'feature.groupCJK': '中英 / 数字间距',
  'feature.groupTypography': '中文排版',
  'feature.groupScript': '简繁转换',
  'feature.groupOutput': '输出换行符',
  // 字符（pre-clean：先于用户规则跑，让用户的 find 字符串能稳定命中）
  'feature.trimLeading': '删除行首空白',
  'feature.trimTrailing': '删除行尾空白',
  'feature.removeBOM': '删除 UTF-8 BOM',
  'feature.removeZeroWidth': '删除零宽字符',
  // 字符（post-clean：full-width → half-width、Kangxi 部首 → 现代字符）
  'feature.fullToHalf': '全角字符转半角字符',
  'feature.replaceKangxi': '康熙部首替换为正常字符',
  // 标点（互斥 radio：「不转换 / E→C / C→E」）
  'feature.punctE2C': '英文标点转换为中文标点',
  'feature.punctC2E': '中文标点转换为英文标点',
  'feature.punctNone': '不转换（保留原文标点）',
  // 引用角标
  'feature.removeCiteParen': '删除括号引用角标，如：(1)、(2,3)、(4-7)',
  'feature.removeCiteBracket': '删除方括号引用角标，如：[1]、【2,3】',
  // 空白与换行（互斥 radio：「不处理 / 严格合并 / 删除空行 / 合并为空行上限」）
  'feature.collapseNewlines': '删除重复的换行符（严格）',
  'feature.newlineToSpace': '将换行符替换为空格',
  'feature.collapseSpaces': '合并连续空格为单个',
  'feature.removeEmptyLines': '删除完全空白的行',
  'feature.collapseBlankLines': '合并连续空行（保留最多 N 个）',
  'feature.maxBlank': '合并后保留的最大连续空行数',
  // 空白行处理（radio 子块）
  'feature.blankLineMode': '空白行处理（互斥）：',
  'feature.blankNone': '不处理（保留原文空行）',
  'feature.blankCollapse': '严格合并（连续换行 → 1 个）',
  'feature.blankRemove': '删除全部空行',
  'feature.blankCollapseMax': '合并为最多 N 个连续空行',
  // CJK / 数字间距
  'feature.removeCJKSpaces': '删除 CJK 字符之间的空格',
  'feature.spaceAfterPunct': '在标点符号后添加空格',
  'feature.removeSpaceAtDecimal': '删除小数点和数字之间的空格',
  'feature.spaceLetterDigit': '在字母与数字之间添加空格',
  'feature.removeSpaceAtColon': '删除冒号和数字之间的空格',
  // 中文排版
  'feature.normalizeTypography': '规范中文排版（参考 sspai 排版指南）',
  // 简繁转换（互斥 radio：「不转换 / 简→繁 / 繁→简」）
  'feature.s2t': '简体中文转换为繁体中文',
  'feature.t2s': '繁体中文转换为简体中文',
  'feature.scriptNone': '不转换（保留原文）',
  // 输出换行符
  'feature.normalizeLE': '统一输出换行符（按下方选择）',
  'feature.lineEnding': '目标换行符',
  'feature.encKeep': '保持原有（keep）',

  // ---- 删除与替换 ----
  'rules.title': '删除与替换',
  'rules.hint': '文本不保存；删除 = 替换为空',
  'rules.deleteSection': '删除内容',
  'rules.deleteNote': '删除即“替换为空白内容”，与普通替换共用同一套规则。',
  'rules.enableDelete': '启用删除规则',
  'rules.mode': '匹配模式',
  'rules.modeLine': '逐行（每行作为一条独立规则）',
  'rules.modeBlock': '整段（整个输入作为一条多行规则）',
  'rules.blockPlaceholder': '粘贴要整体删除的文本内容（可多行）…',
  'rules.linePlaceholder': '每行一条要删除的内容；空行会被忽略…',
  'rules.importTxt': '导入 TXT',
  'rules.clear': '清空',
  'rules.chars': '{n} 字符',
  'rules.replaceSection': '替换内容',
  'rules.tableMode': '表格模式',
  'rules.batchMode': '批量文本模式',
  'rules.importFile': '导入文件',
  'rules.colEnable': '启用',
  'rules.colFind': '查找',
  'rules.colReplace': '替换为',
  'rules.findPlaceholder': '要查找的文本',
  'rules.replacePlaceholder': '替换后的文本',
  'rules.addRow': '+ 添加一行',
  'rules.batchPlaceholder':
    '每行一条规则，用 “=>” 或制表符分隔查找与替换：\nfoo=>bar\n旧标题\t新标题',
  'rules.batchErrorSep':
    '第 {n} 行缺少分隔符（请用 “=>” 或制表符分隔查找与替换）。',
  'rules.batchErrorEmpty': '没有可应用的替换规则。',
  'rules.applyBatch': '解析并应用',
  'rules.batchHint': '应用后将填充到上方表格',
  'rules.importTxtTitle': '导入要删除的文本（TXT）',
  'rules.importFileTitle': '导入查找 / 替换（每行 find=>replace）',
  'rules.readFail': '读取文件失败：',

  // ---- 高级规则 ----
  'adv.title': '高级规则',
  'adv.hint': '正则表达式（可折叠）',
  'adv.expand': '展开正则替换 ▼',
  'adv.collapse': '收起 ▲',
  'adv.note':
    '每条规则均按 Go 正则（RE2 语法）执行查找 / 替换；非法正则会在处理时被跳过并记录。',
  'adv.findPlaceholder': '如 \\d{4}-\\d{2}-\\d{2}',
  'adv.replacePlaceholder': '如 20XX-$0',
  'adv.addRow': '+ 添加正则规则',

  // ---- 预览 ----
  'preview.title': '预览',
  'preview.hint': '与批量处理共用同一套 Go 逻辑',
  'preview.pickPlaceholder': '选择一个文件…',
  'preview.none': '从上方文件列表选择一个文件，这里会显示处理前后的对比。',
  'preview.noFiles': '请先在上方加载文件，再选择要预览的文件。',
  'preview.original': '原始文本',
  'preview.result': '处理结果',
  'preview.diff': '差异对比',
  'preview.split': '双栏对照',
  'preview.loading': '正在生成预览…',
  'preview.refresh': '重新预览',
  'preview.noChange': '按当前规则处理后，该文件没有发生变化。',
  'preview.stats': '删除 {d} 处 · 替换 {r} 处',
  'preview.truncated': '文件较大，预览仅展示前 {n} 个字符（批量处理仍处理完整文件）。',
  'preview.diffTooLarge': '文件行数过多，已跳过差异对比，请查看双栏对照。',
  'preview.failed': '预览失败：{msg}',
  'preview.pickHint': '在文件列表中点击文件名即可预览',

  // ---- 输出与执行 ----
  'out.title': '输出与执行',
  'out.hint': '所有文件均在浏览器本地处理，不会上传到服务器',
  'out.exportMode': '导出方式',
  'out.downloadZip': '下载 ZIP（默认）',
  'out.downloadZipDesc': '保持文件夹和子目录结构，生成 TextCleaner_Output.zip',
  'out.downloadFiles': '下载单个文件',
  'out.downloadFilesDesc': '逐个下载处理后的文件（浏览器可能会请求允许多个下载）',
  'out.privacy': '隐私保护：所有文件均在您的浏览器本地处理，不会上传到任何服务器。',
  'out.encoding': '输出编码',
  'out.encKeep': '保持原编码（不支持编码时回退 UTF-8）',
  'out.encUtf8': 'UTF-8（无 BOM）',
  'out.encUtf8Bom': 'UTF-8（带 BOM）',
  'out.concurrency': '并发数（同时处理文件数）',

  // ---- 处理前显示输出计划 ----
  'plan.title': '确认开始处理',
  'plan.fileCount': '处理文件：',
  'plan.export': '导出方式：',
  'plan.cancel': '取消',
  'plan.confirm': '开始处理',

  // ---- 处理进度 ----
  'prog.title': '处理进度',
  'prog.hint': '后台执行 · 可取消',
  'prog.idle': '点击底部“开始批量处理”后，这里显示实时进度与结果。',
  'prog.cancel': '取消',
  'prog.total': '共 {n}',
  'prog.succeeded': '成功 {n}',
  'prog.failed': '失败 {n}',
  'prog.clear': '清除',
  'prog.viewErrors': '查看',
  'prog.hideErrors': '隐藏',
  'prog.failedDetail': '失败详情（{n}）',

  // ---- 设置页 ----
  'settings.title': '设置',
  'settings.theme': '主题',
  'settings.themeSystem': '跟随系统',
  'settings.themeLight': '浅色',
  'settings.themeDark': '深色',
  'settings.language': '语言',
  'settings.langAuto': '自动',
  'settings.langZh': '中文',
  'settings.langEn': 'English',
  'settings.includeSubfolders': '默认包含子文件夹',
  'settings.includeSubfoldersDesc':
    '新扫描默认递归子目录；可在文件选择区临时覆盖。',
  'settings.concurrency': '默认并发数',
  'settings.concurrencyDesc': '同时处理的文件数量（1–32）。',
  'settings.save': '保存',
  'settings.close': '关闭',
  'settings.reset': '恢复默认',
  'settings.saved': '设置已保存',
  'settings.languageApplied':
    '语言将在下次启动时完全生效；当前界面已即时切换。',
}

const en: Dict = {
  // ---- App shell / action bar ----
  'app.scanning': 'Scanning files…',
  'app.running': 'Processing in progress…',
  'app.selected': 'Selected {n} files · core processing verified in Go',
  'app.start': 'Start batch processing',
  'app.cancel': 'Cancel',
  'app.settings': 'Settings',
  'app.offline': '(offline)',

  // ---- Interaction lock (blocks dimmed until input is selected) ----
  'ui.locked': 'Select files first',
  'ui.lockedHint': 'These options become editable after you select files',

  // ---- File picker ----
  'file.title': 'Files',
  'file.hint': 'Supports .txt / .md / .log / .csv',
  'file.dropMain': 'Drop files or folders into this window',
  'file.dropSub': 'Or choose below (multiple selection supported)',
  'file.pickFiles': 'Choose files…',
  'file.pickFolder': 'Choose folder…',
  'file.includeSubfolders': 'Include subfolders',

  // ---- File list ----
  'list.title': 'File list',
  'list.selected': 'Selected {n} / {m} files',
  'list.search': 'Search file name or path…',
  'list.sort': 'Sort',
  'list.name': 'Name',
  'list.path': 'Path',
  'list.size': 'Size',
  'list.selectAll': 'Select all',
  'list.deselectAll': 'Deselect all',
  'list.clear': 'Clear',
  'list.empty':
    'No files loaded yet. Choose files/folders above, or drop text files.',
  'list.emptyFilter': 'No files match “{q}”.',
  'list.remove': 'Remove',
  'list.removeTitle': 'Remove this file',

  // ---- Basic cleaning ----
  'basic.title': 'Basic cleaning',
  'basic.hint': 'General options, remembered',
  'basic.trimLeading': 'Trim leading whitespace',
  'basic.trimTrailing': 'Trim trailing whitespace',
  'basic.removeEmptyLines': 'Remove blank lines',
  'basic.collapseBlank': 'Collapse consecutive blank lines',
  'basic.collapseSpaces': 'Collapse consecutive spaces',
  'basic.removeZeroWidth': 'Remove zero-width chars',
  'basic.normalizeLE': 'Normalize line endings',
  'basic.removeBOM': 'Remove UTF-8 BOM',
  'basic.maxBlank': 'Max consecutive blank lines to keep',
  'basic.lineEnding': 'Target line ending',

  // ---- Basic features (24 toggles + 2 sub-options, merged from old "Basic cleaning" + "Basic features") ----
  'feature.title': 'Basic features',
  'feature.hint': 'One-click cleanup: 24 common rules for characters, punctuation, citations, whitespace, CJK spacing, Chinese typography, S↔T, etc.',
  'feature.groupChar': 'Characters',
  'feature.groupPunct': 'Punctuation',
  'feature.groupCite': 'Citation marks',
  'feature.groupWhitespace': 'Whitespace & newlines',
  'feature.groupCJK': 'CJK / digit spacing',
  'feature.groupTypography': 'Chinese typography',
  'feature.groupScript': 'Simplified ↔ Traditional',
  'feature.groupOutput': 'Output line ending',
  // Characters (pre-clean: run before user rules so find strings match reliably)
  'feature.trimLeading': 'Trim leading whitespace',
  'feature.trimTrailing': 'Trim trailing whitespace',
  'feature.removeBOM': 'Remove UTF-8 BOM',
  'feature.removeZeroWidth': 'Remove zero-width characters',
  // Characters (post-clean: full-width → half-width, Kangxi → modern)
  'feature.fullToHalf': 'Full-width → half-width characters',
  'feature.replaceKangxi': 'Replace Kangxi radicals with normal characters',
  // Punctuation (mutex radio: none / E→C / C→E)
  'feature.punctE2C': 'English → Chinese punctuation',
  'feature.punctC2E': 'Chinese → English punctuation',
  'feature.punctNone': 'No conversion (keep original)',
  // Citation marks
  'feature.removeCiteParen': 'Delete parenthesized citations (1), (2,3), (4-7)',
  'feature.removeCiteBracket': 'Delete bracketed citations [1], [2,3], [4-7]',
  // Whitespace & newlines (mutex radio: leave / strict / remove / collapseMax)
  'feature.collapseNewlines': 'Collapse duplicate newlines (strict)',
  'feature.newlineToSpace': 'Replace newlines with spaces',
  'feature.collapseSpaces': 'Collapse consecutive spaces',
  'feature.removeEmptyLines': 'Remove blank lines',
  'feature.collapseBlankLines': 'Collapse consecutive blank lines (keep up to N)',
  'feature.maxBlank': 'Max consecutive blank lines to keep',
  'feature.blankLineMode': 'Blank-line handling (mutually exclusive):',
  'feature.blankNone': 'Leave as is (keep original blanks)',
  'feature.blankCollapse': 'Strict: runs of newlines → 1',
  'feature.blankRemove': 'Remove all blank lines',
  'feature.blankCollapseMax': 'Collapse to at most N blank lines',
  // CJK / digit spacing
  'feature.removeCJKSpaces': 'Delete spaces between CJK characters',
  'feature.spaceAfterPunct': 'Add space after punctuation',
  'feature.removeSpaceAtDecimal': 'Delete spaces around decimal point',
  'feature.spaceLetterDigit': 'Add space between letters and digits',
  'feature.removeSpaceAtColon': 'Delete spaces around colon',
  // Chinese typography
  'feature.normalizeTypography': 'Normalize Chinese typography (sspai guidelines)',
  // S↔T (mutex radio: none / S→T / T→S)
  'feature.s2t': 'Simplified → Traditional Chinese',
  'feature.t2s': 'Traditional → Simplified Chinese',
  'feature.scriptNone': 'No conversion (keep original)',
  // Output line ending
  'feature.normalizeLE': 'Normalize output line ending (per dropdown below)',
  'feature.lineEnding': 'Target line ending',
  'feature.encKeep': 'Keep original (keep)',

  // ---- Delete & replace ----
  'rules.title': 'Delete & replace',
  'rules.hint': 'Text not saved; delete = replace with empty',
  'rules.deleteSection': 'Delete content',
  'rules.deleteNote':
    'Delete means “replace with empty content”, sharing the same rule engine as replace.',
  'rules.enableDelete': 'Enable delete rules',
  'rules.mode': 'Match mode',
  'rules.modeLine': 'Line by line (each line as a rule)',
  'rules.modeBlock': 'Block (whole input as one multi-line rule)',
  'rules.blockPlaceholder': 'Paste the block of text to delete (multi-line)…',
  'rules.linePlaceholder': 'One item per line to delete; blank lines ignored…',
  'rules.importTxt': 'Import TXT',
  'rules.clear': 'Clear',
  'rules.chars': '{n} chars',
  'rules.replaceSection': 'Replace content',
  'rules.tableMode': 'Table mode',
  'rules.batchMode': 'Batch text mode',
  'rules.importFile': 'Import file',
  'rules.colEnable': 'Enable',
  'rules.colFind': 'Find',
  'rules.colReplace': 'Replace',
  'rules.findPlaceholder': 'Text to find',
  'rules.replacePlaceholder': 'Replacement text',
  'rules.addRow': '+ Add row',
  'rules.batchPlaceholder':
    'One rule per line, separate find/replace with “=>” or tab:\nfoo=>bar\nold\tnew',
  'rules.batchErrorSep':
    'Line {n} is missing a separator (use “=>” or tab between find and replace).',
  'rules.batchErrorEmpty': 'No replace rules to apply.',
  'rules.applyBatch': 'Parse & apply',
  'rules.batchHint': 'Applied rules will fill the table above',
  'rules.importTxtTitle': 'Import text to delete (TXT)',
  'rules.importFileTitle': 'Import find / replace (one find=>replace per line)',
  'rules.readFail': 'Failed to read file: ',

  // ---- Advanced rules ----
  'adv.title': 'Advanced rules',
  'adv.hint': 'Regex (collapsible)',
  'adv.expand': 'Show regex replace ▼',
  'adv.collapse': 'Hide ▲',
  'adv.note':
    'Each rule runs as a Go regex (RE2 syntax); invalid patterns are skipped and logged during processing.',
  'adv.findPlaceholder': 'e.g. \\d{4}-\\d{2}-\\d{2}',
  'adv.replacePlaceholder': 'e.g. 20XX-$0',
  'adv.addRow': '+ Add regex rule',

  // ---- Preview ----
  'preview.title': 'Preview',
  'preview.hint': 'Same Go logic as batch processing',
  'preview.pickPlaceholder': 'Choose a file…',
  'preview.none': 'Pick a file from the list above to see before/after comparison.',
  'preview.noFiles': 'Load files above first, then choose one to preview.',
  'preview.original': 'Original',
  'preview.result': 'Result',
  'preview.diff': 'Diff',
  'preview.split': 'Side by side',
  'preview.loading': 'Generating preview…',
  'preview.refresh': 'Refresh',
  'preview.noChange': 'No changes for this file with the current rules.',
  'preview.stats': 'Deleted {d} · Replaced {r}',
  'preview.truncated': 'Large file: preview shows the first {n} characters only (batch processing still handles the full file).',
  'preview.diffTooLarge': 'Too many lines for a diff; use the side-by-side view.',
  'preview.failed': 'Preview failed: {msg}',
  'preview.pickHint': 'Click a file name in the list to preview it',

  // ---- Output & run ----
  'out.title': 'Output & run',
  'out.hint': 'All files are processed locally in your browser; nothing is uploaded.',
  'out.exportMode': 'Export mode',
  'out.downloadZip': 'Download ZIP (default)',
  'out.downloadZipDesc': 'Preserves folders and subfolders in TextCleaner_Output.zip',
  'out.downloadFiles': 'Download individual files',
  'out.downloadFilesDesc': 'Downloads each processed file separately; the browser may ask permission for multiple downloads.',
  'out.privacy': 'Privacy: all files are processed locally in your browser and are never uploaded.',
  'out.encoding': 'Output encoding',
  'out.encKeep': 'Keep original (falls back to UTF-8 when unavailable)',
  'out.encUtf8': 'UTF-8 (no BOM)',
  'out.encUtf8Bom': 'UTF-8 (with BOM)',
  'out.concurrency': 'Concurrency (files processed at once)',

  // ---- Start confirmation ----
  'plan.title': 'Confirm processing',
  'plan.fileCount': 'Files:',
  'plan.export': 'Export:',
  'plan.cancel': 'Cancel',
  'plan.confirm': 'Start processing',

  // ---- Progress ----
  'prog.title': 'Progress',
  'prog.hint': 'Runs in background · cancellable',
  'prog.idle':
    'Real-time progress and results appear here after you click “Start” below.',
  'prog.cancel': 'Cancel',
  'prog.total': 'Total {n}',
  'prog.succeeded': 'Succeeded {n}',
  'prog.failed': 'Failed {n}',
  'prog.clear': 'Clear',
  'prog.viewErrors': 'View',
  'prog.hideErrors': 'Hide',
  'prog.failedDetail': 'Failure details ({n})',

  // ---- Settings ----
  'settings.title': 'Settings',
  'settings.theme': 'Theme',
  'settings.themeSystem': 'Follow system',
  'settings.themeLight': 'Light',
  'settings.themeDark': 'Dark',
  'settings.language': 'Language',
  'settings.langAuto': 'Auto',
  'settings.langZh': '中文',
  'settings.langEn': 'English',
  'settings.includeSubfolders': 'Include subfolders by default',
  'settings.includeSubfoldersDesc':
    'New scans recurse into subfolders by default; can be overridden in the file picker.',
  'settings.concurrency': 'Default concurrency',
  'settings.concurrencyDesc': 'Number of files processed simultaneously (1–32).',
  'settings.save': 'Save',
  'settings.close': 'Close',
  'settings.reset': 'Restore defaults',
  'settings.saved': 'Settings saved',
  'settings.languageApplied':
    'Language fully applies on next launch; the current UI switches immediately.',
}

const dict: Record<Lang, Dict> = { zh, en }

const currentLang = ref<Lang>('zh')

export function setLang(lang: Lang): void {
  currentLang.value = lang
}

export function getLang(): Lang {
  return currentLang.value
}

// 根据偏好（auto/zh/en）解析出实际语言；auto 跟随浏览器 locale。
export function langFromPref(pref: string): Lang {
  if (pref === 'en') return 'en'
  if (pref === 'zh') return 'zh'
  if (typeof navigator !== 'undefined' && navigator.language) {
    return navigator.language.toLowerCase().startsWith('zh') ? 'zh' : 'en'
  }
  return 'zh'
}

export function t(key: string, params?: Record<string, string | number>): string {
  const table = dict[currentLang.value] ?? zh
  let s = table[key] ?? zh[key] ?? key
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      s = s.replace(new RegExp('\\{' + k + '\\}', 'g'), String(v))
    }
  }
  return s
}

export function useI18n() {
  return { t, currentLang, setLang, getLang, langFromPref }
}
