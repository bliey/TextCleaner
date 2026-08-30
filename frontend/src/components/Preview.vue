<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import SectionCard from './SectionCard.vue'
import { useI18n } from '../i18n'
import { usePreview, DIFF_MAX_LINES } from '../composables/usePreview'
import { useInteraction } from '../composables/useInteraction'

const { t } = useI18n()
const { locked } = useInteraction()
const {
  files,
  previewPath,
  originalText,
  resultText,
  stats,
  loading,
  errorMsg,
  truncated,
  hasPreview,
  diffRows,
  diffStat,
  diffSkipped,
  preview,
  refresh,
  clear,
} = usePreview()

/**
 * 单次渲染的最大行数。
 * 与文件列表同理：一个上万行的小说文件会把 WebView 拖垮，
 * 因此只渲染首屏能看的部分，并明确告知用户被截断。
 */
const RENDER_LIMIT = 800

const view = ref<'diff' | 'split'>('diff')

const options = computed(() =>
  files.value.map((f) => ({ value: f.path, label: `${f.name} — ${f.path}` })),
)

// 选择的文件若被移出列表，自动清空预览，避免残留上一次的结果。
watch(
  () => files.value.some((f) => f.path === previewPath.value),
  (stillExists) => {
    if (previewPath.value && !stillExists) clear()
  },
)

const visibleDiff = computed(() => (diffRows.value ?? []).slice(0, RENDER_LIMIT))
const diffRemaining = computed(() =>
  Math.max(0, (diffRows.value ?? []).length - RENDER_LIMIT),
)

const originalLines = computed(() =>
  originalText.value.replace(/\r\n?/g, '\n').split('\n').slice(0, RENDER_LIMIT),
)
const resultLines = computed(() =>
  resultText.value.replace(/\r\n?/g, '\n').split('\n').slice(0, RENDER_LIMIT),
)

const statsText = computed(() => {
  if (!stats.value) return ''
  return t('preview.stats', {
    d: stats.value.deletedMatches,
    r: stats.value.replacedMatches,
  })
})

function onPick(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  if (value) void preview(value)
}
</script>

<template>
  <SectionCard :title="t('preview.title')" :hint="t('preview.hint')" :locked="locked">
    <!-- 选择要预览的文件 -->
    <div class="pv-toolbar">
      <select
        class="pv-select"
        :value="previewPath"
        :disabled="files.length === 0 || loading"
        @change="onPick"
      >
        <option value="">{{ t('preview.pickPlaceholder') }}</option>
        <option v-for="o in options" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <button
        class="btn btn-sm"
        :disabled="!previewPath || loading"
        @click="refresh"
      >
        {{ t('preview.refresh') }}
      </button>
      <span class="spacer" />
      <div v-if="hasPreview" class="pv-tabs">
        <button
          class="btn btn-sm"
          :class="{ 'is-active': view === 'diff' }"
          @click="view = 'diff'"
        >
          {{ t('preview.diff') }}
        </button>
        <button
          class="btn btn-sm"
          :class="{ 'is-active': view === 'split' }"
          @click="view = 'split'"
        >
          {{ t('preview.split') }}
        </button>
      </div>
    </div>

    <p v-if="files.length === 0" class="muted">{{ t('preview.noFiles') }}</p>
    <p v-else-if="!hasPreview && !loading" class="muted">
      {{ t('preview.none') }} <span class="pv-hint">{{ t('preview.pickHint') }}</span>
    </p>
    <p v-if="loading" class="muted">{{ t('preview.loading') }}</p>

    <div v-if="errorMsg" class="err-banner">⚠ {{ t('preview.failed', { msg: errorMsg }) }}</div>

    <template v-if="hasPreview && !loading && !errorMsg">
      <!-- 统计条 -->
      <div class="pv-stats">
        <span v-if="stats" class="pill">{{ statsText }}</span>
        <span class="pill" :class="stats?.changed ? 'pill-ok' : 'pill-mute'">
          {{ stats?.changed ? t('preview.diff') : t('preview.noChange') }}
        </span>
        <span v-if="view === 'diff' && !diffSkipped" class="pill pill-del">
          − {{ diffStat.del }}
        </span>
        <span v-if="view === 'diff' && !diffSkipped" class="pill pill-add">
          + {{ diffStat.add }}
        </span>
      </div>

      <p v-if="truncated" class="muted pv-note">
        ⓘ {{ t('preview.truncated', { n: '200,000' }) }}
      </p>
      <p v-if="diffSkipped" class="muted pv-note">
        ⓘ {{ t('preview.diffTooLarge') }}
      </p>

      <!-- 差异对比视图 -->
      <div v-if="view === 'diff' && !diffSkipped" class="pv-diff">
        <div
          v-for="(row, i) in visibleDiff"
          :key="i"
          class="pv-line"
          :class="{
            'pv-line-del': row.type === 'del',
            'pv-line-add': row.type === 'add',
          }"
        >
          <span class="pv-gutter">{{
            row.type === 'del' ? '−' : row.type === 'add' ? '+' : ' '
          }}</span>
          <span class="pv-text">{{ row.text === '' ? ' ' : row.text }}</span>
        </div>
        <p v-if="diffRemaining > 0" class="muted pv-note">
          ⓘ 仅显示前 {{ RENDER_LIMIT }} 行（还有 {{ diffRemaining }} 行未渲染，
          总行数上限 {{ DIFF_MAX_LINES }}）
        </p>
      </div>

      <!-- 双栏对照视图 -->
      <div v-else class="pv-split">
        <div class="pv-pane">
          <div class="pv-pane-head">{{ t('preview.original') }}</div>
          <div class="pv-pane-body">
            <div v-for="(line, i) in originalLines" :key="i" class="pv-line">
              <span class="pv-text">{{ line === '' ? ' ' : line }}</span>
            </div>
          </div>
        </div>
        <div class="pv-pane">
          <div class="pv-pane-head">{{ t('preview.result') }}</div>
          <div class="pv-pane-body">
            <div v-for="(line, i) in resultLines" :key="i" class="pv-line">
              <span class="pv-text">{{ line === '' ? ' ' : line }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </SectionCard>
</template>

<style scoped>
.pv-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.pv-select {
  flex: 1 1 260px;
  min-width: 0;
  padding: 6px 8px;
  font-size: 13px;
  color: var(--text);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
}
.pv-select:disabled {
  color: var(--text-3);
}
.spacer {
  flex: 1;
}
.pv-tabs {
  display: flex;
  gap: 4px;
}
.pv-stats {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}
.pv-note {
  font-size: 12.5px;
  margin: 6px 0;
}
.pv-hint {
  color: var(--text-3);
}
.err-banner {
  color: var(--danger);
  font-size: 13px;
  margin: 8px 0;
}

/* ---- Diff / 双栏共用行样式 ---- */
.pv-diff,
.pv-pane-body {
  max-height: 420px;
  overflow: auto;
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 8px;
}
.pv-line {
  display: flex;
  gap: 8px;
  padding: 1px 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-all;
}
.pv-gutter {
  flex: 0 0 12px;
  user-select: none;
  color: var(--text-3);
  text-align: center;
}
.pv-text {
  flex: 1;
  min-width: 0;
}
/* 删除行：红底（浅色主题下用浅红，深色主题同理，均由 CSS 变量派生） */
.pv-line-del {
  background: var(--danger-weak);
  color: var(--danger);
}
.pv-line-del .pv-gutter {
  color: var(--danger);
}
/* 新增行：绿底 */
.pv-line-add {
  background: color-mix(in srgb, var(--success) 14%, transparent);
  color: var(--success);
}
.pv-line-add .pv-gutter {
  color: var(--success);
}

/* ---- 双栏 ---- */
.pv-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
@media (max-width: 720px) {
  .pv-split {
    grid-template-columns: 1fr;
  }
}
.pv-pane {
  min-width: 0;
}
.pv-pane-head {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-2);
  padding: 4px 2px 6px;
}

/* ---- 统计标签 ---- */
.pill-del {
  background: var(--danger-weak);
  border-color: var(--danger);
  color: var(--danger);
}
.pill-add {
  background: color-mix(in srgb, var(--success) 16%, var(--surface));
  border-color: var(--success);
  color: var(--success);
}
.pill-mute {
  color: var(--text-3);
}
</style>
