<script setup lang="ts">
import { ref } from 'vue'
import SectionCard from './SectionCard.vue'
import { useOptions } from '../composables/useOptions'
import { useInteraction } from '../composables/useInteraction'
import { readFileText } from '../core/encoding'

const { deleteEnabled, deleteMode, deleteContent, replaceRules, addSimpleRule, removeSimpleRule } =
  useOptions()
const { locked } = useInteraction()

async function pickTextFile(accept: string): Promise<File | null> {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = accept
  return new Promise((resolve) => {
    input.addEventListener('change', () => resolve(input.files?.[0] ?? null), { once: true })
    input.click()
  })
}

async function importTxt() {
  const file = await pickTextFile('.txt,text/plain')
  if (!file) return
  try {
    const content = await readFileText(file)
    deleteContent.value = content.text
    deleteEnabled.value = true
  } catch (e) {
    alert('读取文件失败：' + (e instanceof Error ? e.message : String(e)))
  }
}

function clearContent() {
  deleteContent.value = ''
}

// ---- 替换（表格 / 批量）----
const mode = ref<'table' | 'batch'>('table')
const batchText = ref('')
const batchError = ref('')

function applyBatch() {
  batchError.value = ''
  const lines = batchText.value.split(/\r?\n/)
  const pairs: { find: string; replace: string }[] = []
  let lineNo = 0
  for (const raw of lines) {
    lineNo++
    const line = raw.replace(/\r$/, '')
    if (line.trim() === '') continue
    let sep = line.indexOf('=>')
    let find = ''
    let replace = ''
    if (sep >= 0) {
      find = line.slice(0, sep)
      replace = line.slice(sep + 2)
    } else {
      const tab = line.indexOf('\t')
      if (tab >= 0) {
        find = line.slice(0, tab)
        replace = line.slice(tab + 1)
      } else {
        batchError.value = `第 ${lineNo} 行缺少分隔符（请用 “=>” 或制表符分隔查找与替换）。`
        return
      }
    }
    pairs.push({ find, replace })
  }
  if (pairs.length === 0) {
    batchError.value = '没有可应用的替换规则。'
    return
  }
  replaceRules.value = pairs.map((p) => ({
    enabled: true,
    find: p.find,
    replace: p.replace,
    regex: false,
  }))
}

async function importFile() {
  const file = await pickTextFile('.txt,.md,.log,.csv,text/plain,text/csv')
  if (!file) return
  try {
    const content = await readFileText(file)
    batchText.value = content.text
    mode.value = 'batch'
    batchError.value = ''
  } catch (e) {
    alert('读取文件失败：' + (e instanceof Error ? e.message : String(e)))
  }
}
</script>

<template>
  <SectionCard title="删除与替换" hint="文本不保存；删除 = 替换为空" :locked="locked">
    <!-- 删除内容（即“替换为空白内容”） -->
    <div class="sub">
      <h3 class="sub-title">删除内容</h3>
      <p class="muted sub-note">删除即“替换为空白内容”，与普通替换共用同一套规则。</p>
      <label class="check-row">
        <input type="checkbox" v-model="deleteEnabled" />
        <span class="check-label">启用删除规则</span>
      </label>

      <div class="mode-row" :class="{ disabled: !deleteEnabled }">
        <span class="label">匹配模式</span>
        <label class="radio">
          <input type="radio" value="line" v-model="deleteMode" :disabled="!deleteEnabled" />
          <span>逐行（每行作为一条独立规则）</span>
        </label>
        <label class="radio">
          <input type="radio" value="block" v-model="deleteMode" :disabled="!deleteEnabled" />
          <span>整段（整个输入作为一条多行规则）</span>
        </label>
      </div>

      <textarea
        v-model="deleteContent"
        :disabled="!deleteEnabled"
        :placeholder="
          deleteMode === 'block'
            ? '粘贴要整体删除的文本内容（可多行）…'
            : '每行一条要删除的内容；空行会被忽略…'
        "
        class="content-area"
      />

      <div class="dz-actions">
        <button class="btn btn-sm" :disabled="!deleteEnabled" @click="importTxt">
          导入 TXT
        </button>
        <button class="btn btn-sm" :disabled="!deleteEnabled" @click="clearContent">
          清空
        </button>
        <span class="spacer" />
        <span class="muted count">{{ deleteContent.length }} 字符</span>
      </div>
    </div>

    <div class="divider" />

    <!-- 替换内容 -->
    <div class="sub">
      <h3 class="sub-title">替换内容</h3>
      <div class="mode-tabs">
        <button class="btn btn-sm" :class="{ 'is-active': mode === 'table' }" @click="mode = 'table'">
          表格模式
        </button>
        <button class="btn btn-sm" :class="{ 'is-active': mode === 'batch' }" @click="mode = 'batch'">
          批量文本模式
        </button>
        <span class="spacer" />
        <button class="btn btn-sm" @click="importFile">导入文件</button>
      </div>

      <div v-if="mode === 'table'" class="tbl">
        <div class="tbl-head">
          <span class="c-ena">启用</span>
          <span class="c-find">查找</span>
          <span class="c-rep">替换为</span>
          <span class="c-act" />
        </div>
        <div v-for="(r, i) in replaceRules" :key="i" class="tbl-row">
          <span class="c-ena"><input type="checkbox" v-model="r.enabled" /></span>
          <span class="c-find"><input type="text" v-model="r.find" placeholder="要查找的文本" /></span>
          <span class="c-rep"><input type="text" v-model="r.replace" placeholder="替换后的文本" /></span>
          <span class="c-act"><button class="btn btn-sm btn-danger" @click="removeSimpleRule(i)">×</button></span>
        </div>
        <button class="btn btn-sm add-row" @click="addSimpleRule">+ 添加一行</button>
      </div>

      <div v-else class="batch">
        <textarea
          v-model="batchText"
          class="batch-area"
          placeholder="每行一条规则，用 “=>” 或制表符分隔查找与替换：&#10;foo=>bar&#10;旧标题	新标题"
        />
        <p v-if="batchError" class="batch-error">⚠ {{ batchError }}</p>
        <div class="dz-actions">
          <button class="btn btn-primary btn-sm" @click="applyBatch">解析并应用</button>
          <span class="spacer" />
          <span class="muted">应用后将填充到上方表格</span>
        </div>
      </div>
    </div>
  </SectionCard>
</template>

<style scoped>
.sub-title {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-2);
}
.sub-note {
  margin: 0 0 10px;
}
.divider {
  height: 1px;
  background: var(--border);
  margin: 18px 0;
}
.mode-row {
  display: flex;
  align-items: center;
  gap: 18px;
  margin: 12px 0;
  flex-wrap: wrap;
}
.mode-row.disabled {
  opacity: 0.5;
}
.radio {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.content-area {
  min-height: 120px;
}
.dz-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
}
.count {
  font-variant-numeric: tabular-nums;
}
.mode-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.btn-sm.is-active {
  background: var(--accent-weak);
  border-color: var(--accent);
  color: var(--accent);
}
.tbl {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.tbl-head,
.tbl-row {
  display: grid;
  grid-template-columns: 54px 1fr 1fr 40px;
  gap: 10px;
  align-items: center;
}
.tbl-head {
  font-size: 12px;
  color: var(--text-3);
  padding: 0 2px;
}
.tbl-row input[type='text'] {
  width: 100%;
}
.c-ena {
  display: flex;
  justify-content: center;
}
.c-act {
  display: flex;
  justify-content: center;
}
.add-row {
  align-self: flex-start;
  margin-top: 4px;
}
.batch-area {
  min-height: 160px;
  font-family: 'SF Mono', 'Cascadia Code', Menlo, Consolas, monospace;
}
.batch-error {
  color: var(--danger);
  font-size: 13px;
  margin: 8px 0 0;
}
</style>
