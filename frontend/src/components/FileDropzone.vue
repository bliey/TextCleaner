<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '../i18n'
import { useFiles } from '../composables/useFiles'

const { t } = useI18n()
const { files, selectedCount, scanning, includeSubfolders, pickFiles, pickFolder, clearAll } =
  useFiles()

// 标签动态文案：与 picocrypt-wails 一致——空态显示拖入提示，
// 有文件时显示已选数量（与底部状态栏形成层级关系）。
const labelText = computed(() => {
  if (files.value.length === 0) return t('file.dropMain')
  const n = selectedCount.value
  const total = files.value.length
  return t('app.selected', { n }) + ` · 共 ${total} 个`
})

// 清除按钮：无文件 / 扫描中均禁用。
// “处理中”由父组件 SectionCard 的 :locked 控制整个块；此处仅处理本块状态。
const canClear = computed(() => files.value.length > 0 && !scanning.value)
</script>

<template>
  <!--
    顶部文件选择入口（参考 picocrypt-wails 布局）：
      第一行：动态标签 + 清除按钮
      第二行：选择文件 / 选择文件夹 / 包含子文件夹
    整窗（#app-root）已设 data-file-drop-target，拖入即加入列表；
    高亮反馈见 theme.css 的 #app-root.file-drop-target-active。
  -->
  <div class="file-row">
    <span class="input-label" :title="labelText">{{ labelText }}</span>
    <button class="btn btn-sm" :disabled="!canClear" @click="clearAll">
      {{ t('file.clear') }}
    </button>
  </div>
  <div class="file-row">
    <button class="btn btn-sm btn-primary" :disabled="scanning" @click="pickFiles">
      {{ t('file.pickFiles') }}
    </button>
    <button class="btn btn-sm" :disabled="scanning" @click="pickFolder">
      {{ t('file.pickFolder') }}
    </button>
    <span class="spacer" />
    <label class="check-row dz-subfolder">
      <input
        type="checkbox"
        v-model="includeSubfolders"
      />
      <span class="check-label">{{ t('file.includeSubfolders') }}</span>
    </label>
  </div>
</template>

<style scoped>
/* 与 picocrypt-wails 的 .row / .input-label 对齐；最小自定义，复用全局 .btn .check-row .spacer */
.file-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 4px 0;
}
.input-label {
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}
.dz-subfolder {
  padding: 0;
}
</style>
