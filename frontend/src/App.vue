<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useI18n } from './i18n'
import { useFiles } from './composables/useFiles'
import { useSettings } from './composables/useSettings'
import { useBatch } from './composables/useBatch'
import { usePreview } from './composables/usePreview'
import FileDropzone from './components/FileDropzone.vue'
import FileList from './components/FileList.vue'
import BasicCleaner from './components/BasicCleaner.vue'
import RulesBlock from './components/RulesBlock.vue'
import AdvancedRules from './components/AdvancedRules.vue'
import Preview from './components/Preview.vue'
import OutputSettings from './components/OutputSettings.vue'
import ProcessingProgress from './components/ProcessingProgress.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import StartConfirmDialog from './components/StartConfirmDialog.vue'

const version = ref('Web')
const { t } = useI18n()

// 持久化设置（主题 / 语言 / 并发数 / 默认包含子文件夹 / 基础清理）
const { settings, ready: settingsReady, load: loadSettings, applyTheme } = useSettings()

// ---- 文件选择与扫描（useFiles 封装了 Go 端 scanner 调用、拖拽与对话框）----
const {
  files,
  selected,
  selectedCount,
  scanning,
  errorMessage,
  includeSubfolders,
  pickFiles,
  pickFolder,
  toggle,
  selectAll,
  deselectAll,
  remove,
  clearAll,
  bindDropEvents,
  unbindDropEvents,
} = useFiles()

// 批量处理：start() 只构造 pendingPlan（待确认），由用户在
// StartConfirmDialog 点确认后才开始浏览器本地处理。
const {
  running,
  pendingPlan,
  start,
  confirmStart,
  cancelPending,
  cancel,
} = useBatch()

// 预览（点击文件列表中的文件名触发）
const { preview: previewFile } = usePreview()

const showSettings = ref(false)

// ---------- 拖放高亮兜底清理 ----------
// Windows 上 dragleave(relatedTarget===null) 可能导致高亮残留，
// 这里在窗口级 leave / drop 时清理。
function clearDropHighlight() {
  document
    .querySelectorAll('.file-drop-target-active')
    .forEach((el) => el.classList.remove('file-drop-target-active'))
}
function onDragLeave(e: DragEvent) {
  if (e.relatedTarget === null) clearDropHighlight()
}

onMounted(async () => {
  // 先按默认值渲染一次主题，避免首屏闪白；load 完成后会再应用一次持久化值
  applyTheme()
  // 加载浏览器本地持久化设置；完成后初始化当前会话开关
  await loadSettings()
  if (settingsReady.value) {
    includeSubfolders.value = settings.defaultIncludeSubfolders
  }
  bindDropEvents()
  document.addEventListener('dragleave', onDragLeave)
  document.addEventListener('drop', clearDropHighlight)
})

onBeforeUnmount(() => {
  unbindDropEvents()
  document.removeEventListener('dragleave', onDragLeave)
  document.removeEventListener('drop', clearDropHighlight)
})

const canStart = computed(() => selectedCount.value > 0)
</script>

<template>
  <!-- 整个窗口都是浏览器原生文件拖放目标，拖入的文件会被加入列表。
       文件未选择时，规则/输出/预览等区块通过 SectionCard 的 locked 自动置灰。 -->
  <div id="app-root" data-file-drop-target>
    <header class="app-header">
      <div class="brand">
        <span class="brand-title">Text Cleaner</span>
        <span class="brand-version">v{{ version }}</span>
      </div>
      <div class="header-actions">
        <select v-model="settings.theme" class="btn btn-sm" :title="t('settings.theme')">
          <option value="system">🖥 {{ t('settings.themeSystem') }}</option>
          <option value="light">☀ {{ t('settings.themeLight') }}</option>
          <option value="dark">🌙 {{ t('settings.themeDark') }}</option>
        </select>
        <button
          class="btn btn-sm"
          :title="t('app.settings')"
          :aria-label="t('app.settings')"
          @click="showSettings = true"
        >
          ⚙
        </button>
      </div>
    </header>

    <main class="app-main">
      <div class="app-main-inner">
        <FileDropzone />
        <FileList
          :files="files"
          :selected="selected"
          :selected-count="selectedCount"
          @toggle="toggle"
          @select-all="selectAll"
          @deselect-all="deselectAll"
          @remove="remove"
          @clear-all="clearAll"
          @preview="previewFile"
        />
        <BasicCleaner />
        <RulesBlock />
        <AdvancedRules />
        <Preview />
        <OutputSettings />
        <ProcessingProgress />
      </div>
    </main>

    <div class="action-bar">
      <span class="action-info">
        <template v-if="errorMessage">⚠ {{ errorMessage }}</template>
        <template v-else-if="scanning">{{ t('app.scanning') }}</template>
        <template v-else-if="running">{{ t('app.running') }}</template>
        <template v-else>{{ t('app.selected', { n: selectedCount }) }}</template>
      </span>
      <button v-if="running" class="btn btn-danger" @click="cancel">
        {{ t('app.cancel') }}
      </button>
      <button class="btn btn-primary" :disabled="running || !canStart" @click="start">
        {{ t('app.start') }}
      </button>
    </div>

    <SettingsPanel v-if="showSettings" @close="showSettings = false" />

    <!-- 处理前显示输出计划（spec §8）：用户在 start() 构造 pendingPlan 后
         由 StartConfirmDialog 呈现计划与 [取消]/[开始处理] 按钮。 -->
    <StartConfirmDialog
      v-if="pendingPlan"
      :plan="pendingPlan"
      @confirm="confirmStart"
      @cancel="cancelPending"
    />
  </div>
</template>
