<script setup lang="ts">
import { ref } from 'vue'
import SectionCard from './SectionCard.vue'
import { useBatch } from '../composables/useBatch'

const { running, progress, summary, errorMsg, downloadReady, downloadResults, cancel, clearResults } = useBatch()
const showErrors = ref(false)

function pct(p: { done: number; total: number } | null): number {
  if (!p || p.total === 0) return 0
  return Math.round((p.done / p.total) * 100)
}
function baseName(path: string): string {
  const i = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'))
  return i >= 0 ? path.slice(i + 1) : path
}
const failedList = () => (summary.value?.results ?? []).filter((result) => !result.success)
</script>

<template>
  <SectionCard title="处理进度" hint="浏览器本地执行 · 可取消">
    <div v-if="!running && !summary && !errorMsg" class="muted">
      点击底部“开始批量处理”后，这里显示实时进度与结果。
    </div>
    <div v-if="running" class="running">
      <div class="bar"><div class="bar-fill" :style="{ width: pct(progress) + '%' }" /></div>
      <div class="stat-row">
        <span class="pill">{{ progress ? progress.done : 0 }} / {{ progress ? progress.total : 0 }}</span>
        <span v-if="progress?.current" class="muted cur">当前：{{ progress.current }}</span>
        <span class="spacer" />
        <button class="btn btn-sm btn-danger" @click="cancel">取消</button>
      </div>
    </div>
    <div v-if="errorMsg" class="err-banner">⚠ {{ errorMsg }}</div>
    <div v-if="summary && !running" class="done">
      <div class="stat-row">
        <span class="pill">共 {{ summary.total }}</span>
        <span class="pill pill-ok">成功 {{ summary.succeeded }}</span>
        <span class="pill pill-bad">失败 {{ summary.failed }}</span>
        <span class="spacer" />
        <button v-if="downloadReady" class="btn btn-sm btn-primary" @click="downloadResults">下载结果</button>
        <button class="btn btn-sm" @click="clearResults">清除</button>
      </div>
      <p v-if="summary.failed > 0" class="muted">
        <button class="btn btn-sm" @click="showErrors = !showErrors">
          {{ showErrors ? '隐藏' : '查看' }}失败详情（{{ summary.failed }}）
        </button>
      </p>
      <ul v-if="showErrors && failedList().length" class="err-list">
        <li v-for="(result, index) in failedList()" :key="index">
          <span class="ef">{{ baseName(result.path) }}</span>
          <span class="em">{{ result.error }}</span>
        </li>
      </ul>
    </div>
  </SectionCard>
</template>

<style scoped>
.bar { height: 8px; background: var(--surface-2); border-radius: 999px; overflow: hidden; border: 1px solid var(--border); }
.bar-fill { height: 100%; background: var(--accent); transition: width .15s ease; }
.stat-row { display: flex; align-items: center; gap: 8px; margin-top: 10px; flex-wrap: wrap; }
.cur { word-break: break-all; }
.pill-ok { background: color-mix(in srgb, var(--success) 16%, var(--surface)); border-color: var(--success); color: var(--success); }
.pill-bad { background: var(--danger-weak); border-color: var(--danger); color: var(--danger); }
.err-banner { color: var(--danger); font-size: 13px; margin-top: 10px; }
.err-list { margin: 8px 0 0; padding-left: 18px; font-size: 12.5px; }
.err-list li { margin-bottom: 4px; }
.ef { color: var(--text); }
.em { color: var(--danger); margin-left: 8px; }
</style>
