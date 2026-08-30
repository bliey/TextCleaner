<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import SectionCard from './SectionCard.vue'
import type { FileEntry } from '../types/models'
import { useI18n } from '../i18n'

const { t } = useI18n()

const props = defineProps<{
  files: FileEntry[]
  selected: Record<string, boolean>
  selectedCount: number
}>()

const emit = defineEmits<{
  (e: 'toggle', path: string): void
  (e: 'select-all'): void
  (e: 'deselect-all'): void
  (e: 'remove', path: string): void
  (e: 'clear-all'): void
  (e: 'preview', path: string): void
}>()

// ---- 搜索与排序（仅作用于界面展示，不修改底层数据）----
const search = ref('')
const sortKey = ref<'name' | 'size' | 'path'>('name')
const sortDir = ref<'asc' | 'desc'>('asc')

const displayFiles = computed<FileEntry[]>(() => {
  const q = search.value.trim().toLowerCase()
  let list = props.files
  if (q) {
    list = list.filter(
      (f) => f.name.toLowerCase().includes(q) || f.path.toLowerCase().includes(q),
    )
  }
  const dir = sortDir.value === 'asc' ? 1 : -1
  const key = sortKey.value
  return [...list].sort((a, b) => {
    if (key === 'size') return (a.size - b.size) * dir
    const av = key === 'name' ? a.name : a.path
    const bv = key === 'name' ? b.name : b.path
    return av.localeCompare(bv, 'zh-Hans-CN') * dir
  })
})

const hint = computed(() =>
  t('list.selected', { n: props.selectedCount, m: props.files.length }),
)

// ---- 虚拟列表（窗口化渲染，支持万级文件不卡顿）----
const rowHeight = 38
const viewportEl = ref<HTMLElement | null>(null)
const scrollTop = ref(0)
const viewportHeight = ref(420)

const totalHeight = computed(() => displayFiles.value.length * rowHeight)
const visibleCount = computed(() =>
  Math.ceil(viewportHeight.value / rowHeight) + 2,
)
const startIndex = computed(() =>
  Math.max(0, Math.floor(scrollTop.value / rowHeight)),
)
const endIndex = computed(() =>
  Math.min(displayFiles.value.length, startIndex.value + visibleCount.value),
)
const visibleRows = computed(() =>
  displayFiles.value.slice(startIndex.value, endIndex.value),
)

function onScroll(e: Event) {
  scrollTop.value = (e.target as HTMLElement).scrollTop
}

function setSort(key: 'name' | 'size' | 'path') {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

function formatSize(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}

let ro: ResizeObserver | null = null
onMounted(() => {
  if (viewportEl.value) {
    viewportHeight.value = viewportEl.value.clientHeight
    ro = new ResizeObserver(() => {
      if (viewportEl.value) viewportHeight.value = viewportEl.value.clientHeight
    })
    ro.observe(viewportEl.value)
  }
})
onBeforeUnmount(() => {
  ro?.disconnect()
  ro = null
})
</script>

<template>
  <SectionCard :title="t('list.title')" :hint="hint">
    <!-- 工具栏：搜索 / 排序 / 批量操作 -->
    <div class="fl-toolbar">
      <input
        v-model="search"
        type="text"
        class="fl-search"
        :placeholder="t('list.search')"
      />
      <div class="row fl-sort">
        <span class="fl-sort-label">{{ t('list.sort') }}</span>
        <button
          class="btn btn-sm"
          :class="{ 'is-active': sortKey === 'name' }"
          @click="setSort('name')"
        >
          {{ t('list.name') }}{{ sortKey === 'name' ? (sortDir === 'asc' ? ' ↑' : ' ↓') : '' }}
        </button>
        <button
          class="btn btn-sm"
          :class="{ 'is-active': sortKey === 'path' }"
          @click="setSort('path')"
        >
          {{ t('list.path') }}{{ sortKey === 'path' ? (sortDir === 'asc' ? ' ↑' : ' ↓') : '' }}
        </button>
        <button
          class="btn btn-sm"
          :class="{ 'is-active': sortKey === 'size' }"
          @click="setSort('size')"
        >
          {{ t('list.size') }}{{ sortKey === 'size' ? (sortDir === 'asc' ? ' ↑' : ' ↓') : '' }}
        </button>
      </div>
      <span class="spacer" />
      <button class="btn btn-sm" :disabled="files.length === 0" @click="emit('select-all')">
        {{ t('list.selectAll') }}
      </button>
      <button class="btn btn-sm" :disabled="files.length === 0" @click="emit('deselect-all')">
        {{ t('list.deselectAll') }}
      </button>
      <button
        class="btn btn-sm"
        :disabled="files.length === 0"
        @click="emit('clear-all')"
      >
        {{ t('list.clear') }}
      </button>
    </div>

    <!-- 列表头（固定） -->
    <div class="fl-head">
      <span class="col-check" />
      <span class="col-name" @click="setSort('name')">{{ t('list.name') }}</span>
      <span class="col-path" @click="setSort('path')">{{ t('list.path') }}</span>
      <span class="col-size" @click="setSort('size')">{{ t('list.size') }}</span>
      <span class="col-act" />
    </div>

    <!-- 空状态 -->
    <div v-if="files.length === 0" class="fl-empty muted">
      {{ t('list.empty') }}
    </div>

    <!-- 虚拟列表视口 -->
    <div
      v-else
      ref="viewportEl"
      class="fl-viewport"
      @scroll="onScroll"
    >
      <div class="fl-spacer" :style="{ height: totalHeight + 'px' }">
        <div
          v-for="(f, i) in visibleRows"
          :key="f.path"
          class="fl-row"
          :class="{ 'is-selected': selected[f.path] }"
          :style="{ top: (startIndex + i) * rowHeight + 'px' }"
        >
          <span class="col-check">
            <input
              type="checkbox"
              :checked="!!selected[f.path]"
              @change="emit('toggle', f.path)"
            />
          </span>
          <span class="col-name" :title="f.name">
            <button class="fl-name" @click="emit('preview', f.path)">
              {{ f.name }}
            </button>
          </span>
          <span class="col-path" :title="f.path">{{ f.path }}</span>
          <span class="col-size">{{ formatSize(f.size) }}</span>
          <span class="col-act">
            <button
              class="btn btn-sm btn-danger"
              :title="t('list.removeTitle')"
              @click="emit('remove', f.path)"
            >
              {{ t('list.remove') }}
            </button>
          </span>
        </div>
      </div>
    </div>

    <!-- 过滤后为空 -->
    <div
      v-if="files.length > 0 && displayFiles.length === 0"
      class="fl-empty muted"
    >
      {{ t('list.emptyFilter', { q: search }) }}
    </div>
  </SectionCard>
</template>

<style scoped>
/* 文件名即预览入口：视觉上是可点击链接，但不引入额外列，避免挤压路径显示 */
.fl-name {
  padding: 0;
  border: none;
  background: none;
  font: inherit;
  color: var(--text);
  text-align: left;
  cursor: pointer;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.fl-name:hover {
  color: var(--accent);
  text-decoration: underline;
}
.fl-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.fl-search {
  flex: 1 1 240px;
  min-width: 180px;
}
.fl-sort {
  gap: 6px;
}
.fl-sort-label {
  font-size: 12px;
  color: var(--text-3);
}
.btn-sm.is-active {
  background: var(--accent-weak);
  border-color: var(--accent);
  color: var(--accent);
}

.fl-head,
.fl-row {
  display: grid;
  grid-template-columns: 34px 1.4fr 2.2fr 90px 64px;
  align-items: center;
  gap: 10px;
}
.fl-head {
  padding: 6px 10px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-3);
  user-select: none;
}
.fl-head .col-name,
.fl-head .col-path,
.fl-head .col-size {
  cursor: pointer;
}
.fl-head .col-name:hover,
.fl-head .col-path:hover,
.fl-head .col-size:hover {
  color: var(--accent);
}

.fl-viewport {
  position: relative;
  height: 420px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface-2);
}
.fl-spacer {
  position: relative;
}
.fl-row {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 38px;
  padding: 0 10px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  background: var(--surface);
}
.fl-row.is-selected {
  background: var(--accent-weak);
}
.fl-row:hover {
  background: var(--surface-2);
}
.col-name,
.col-path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.col-path {
  color: var(--text-3);
}
.col-size {
  text-align: right;
  color: var(--text-2);
  font-variant-numeric: tabular-nums;
}
.col-check {
  display: flex;
  justify-content: center;
}
.col-check input {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
}
.col-act {
  display: flex;
  justify-content: flex-end;
}

.fl-empty {
  padding: 24px 10px;
  text-align: center;
}
</style>
