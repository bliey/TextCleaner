<script setup lang="ts">
import { computed } from 'vue'
import SectionCard from './SectionCard.vue'
import { useOptions } from '../composables/useOptions'
import { useI18n } from '../i18n'
import { useInteraction } from '../composables/useInteraction'
import type { BasicCleanOptions } from '../core/types'

const { basicClean } = useOptions()
const { t } = useI18n()
const { locked } = useInteraction()

type BoolKey = {
  [K in keyof BasicCleanOptions]: BasicCleanOptions[K] extends boolean ? K : never
}[keyof BasicCleanOptions]

// 单一「基础功能」SectionCard：24 个 checkbox + 2 个子选项（maxBlankLines / lineEnding），
// 按用途分 8 组渲染。原「基础清理」与「基础功能」已合并——重复的类似项
// （CollapseSpaces、RemoveEmptyLines vs CollapseBlankLines 等）通过明确的 i18n
// 标签区分，避免歧义。
//
// 互斥项机制（三组 radio）：
//   - 标点方向（中↔英）         groupPunct      → 'none' | 'e2c' | 'c2e'
//   - 简繁转换方向                groupScript     → 'none' | 's2t' | 't2s'
//   - 空白行处理（严格/移除/限制） groupWhitespace 下的 blank-line 行
//                                                  → 'none' | 'collapse' | 'remove' | 'collapseMax'
// 下方三个 computed 自动同步 radio 选项与多个底层 bool 字段。后端 cleanrules
// 仍做防御性仲裁（万一 settings.json 直接被改），前端只是更直观的 UX。

interface FeatureGroup {
  title: string
  items: { key: BoolKey; label: string }[]
}
const featureGroups: FeatureGroup[] = [
  {
    title: 'feature.groupChar',
    items: [
      { key: 'trimLeadingWhitespace', label: 'feature.trimLeading' },
      { key: 'trimTrailingWhitespace', label: 'feature.trimTrailing' },
      { key: 'removeUTF8BOM', label: 'feature.removeBOM' },
      { key: 'removeZeroWidthChars', label: 'feature.removeZeroWidth' },
      { key: 'fullWidthToHalfWidth', label: 'feature.fullToHalf' },
      { key: 'replaceKangxiRadicals', label: 'feature.replaceKangxi' },
    ],
  },
  {
    title: 'feature.groupCite',
    items: [
      { key: 'removeCiteParen', label: 'feature.removeCiteParen' },
      { key: 'removeCiteBracket', label: 'feature.removeCiteBracket' },
    ],
  },
  {
    title: 'feature.groupCJK',
    items: [
      { key: 'removeSpaceBetweenCJK', label: 'feature.removeCJKSpaces' },
      { key: 'spaceAfterPunctuation', label: 'feature.spaceAfterPunct' },
      { key: 'removeSpaceAtDecimal', label: 'feature.removeSpaceAtDecimal' },
      { key: 'spaceBetweenLetterAndDigit', label: 'feature.spaceLetterDigit' },
      { key: 'removeSpaceAtColon', label: 'feature.removeSpaceAtColon' },
    ],
  },
  {
    title: 'feature.groupTypography',
    items: [{ key: 'normalizeChineseTypography', label: 'feature.normalizeTypography' }],
  },
]

// 独立（非互斥）的换行相关项，仅 newlineToSpace 留作独立 checkbox；
// 其它（collapseNewlines / removeEmptyLines / collapseBlankLines）已被
// blankLineMode radio 替代。
const newlineToSpaceItem = { key: 'newlineToSpace' as BoolKey, label: 'feature.newlineToSpace' }
const collapseSpacesItem = { key: 'collapseSpaces' as BoolKey, label: 'feature.collapseSpaces' }

// ============================================================
// 三组互斥 radio 的 get/set computed
// ============================================================

// 1) 标点方向
type PunctDir = 'none' | 'e2c' | 'c2e'
const punctDir = computed<PunctDir>({
  get: () => {
    if (basicClean.punctEnglishToChinese) return 'e2c'
    if (basicClean.punctChineseToEnglish) return 'c2e'
    return 'none'
  },
  set: (v) => {
    basicClean.punctEnglishToChinese = v === 'e2c'
    basicClean.punctChineseToEnglish = v === 'c2e'
  },
})

// 2) 简繁转换
type ScriptDir = 'none' | 's2t' | 't2s'
const scriptDir = computed<ScriptDir>({
  get: () => {
    if (basicClean.simplifiedToTraditional) return 's2t'
    if (basicClean.traditionalToSimplified) return 't2s'
    return 'none'
  },
  set: (v) => {
    basicClean.simplifiedToTraditional = v === 's2t'
    basicClean.traditionalToSimplified = v === 't2s'
  },
})

// 3) 空白行处理
//    'none'         → 三个 bool 均为 false（保留原文空行）
//    'collapse'     → 仅 collapseNewlines = true（连续换行严格合并为 1）
//    'remove'       → 仅 removeEmptyLines = true（删除全部空行）
//    'collapseMax'  → 仅 collapseBlankLines = true（按 maxBlankLines 合并）
type BlankLineMode = 'none' | 'collapse' | 'remove' | 'collapseMax'
const blankLineMode = computed<BlankLineMode>({
  get: () => {
    if (basicClean.removeEmptyLines) return 'remove'
    if (basicClean.collapseNewlines) return 'collapse'
    if (basicClean.collapseBlankLines) return 'collapseMax'
    return 'none'
  },
  set: (v) => {
    basicClean.collapseNewlines = v === 'collapse'
    basicClean.removeEmptyLines = v === 'remove'
    basicClean.collapseBlankLines = v === 'collapseMax'
  },
})
</script>

<template>
  <!-- 单一「基础功能」SectionCard -->
  <SectionCard :title="t('feature.title')" :hint="t('feature.hint')" :locked="locked">
    <!-- 非互斥的几个分组 -->
    <div
      v-for="group in featureGroups"
      :key="group.title"
      class="feature-group"
    >
      <div class="group-title">{{ t(group.title) }}</div>
      <div class="clean-grid">
        <label v-for="it in group.items" :key="it.key" class="check-row">
          <input type="checkbox" v-model="basicClean[it.key]" />
          <span class="check-label">{{ t(it.label) }}</span>
        </label>
      </div>
    </div>

    <!-- ========== 标点方向（互斥 radio） ========== -->
    <div class="feature-group">
      <div class="group-title">{{ t('feature.groupPunct') }}</div>
      <div class="mutex-grid">
        <label class="radio-row">
          <input type="radio" :value="'none'" v-model="punctDir" name="punctDir" />
          <span class="check-label">{{ t('feature.punctNone') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'e2c'" v-model="punctDir" name="punctDir" />
          <span class="check-label">{{ t('feature.punctE2C') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'c2e'" v-model="punctDir" name="punctDir" />
          <span class="check-label">{{ t('feature.punctC2E') }}</span>
        </label>
      </div>
    </div>

    <!-- ========== 简繁转换（互斥 radio） ========== -->
    <div class="feature-group">
      <div class="group-title">{{ t('feature.groupScript') }}</div>
      <div class="mutex-grid">
        <label class="radio-row">
          <input type="radio" :value="'none'" v-model="scriptDir" name="scriptDir" />
          <span class="check-label">{{ t('feature.scriptNone') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'s2t'" v-model="scriptDir" name="scriptDir" />
          <span class="check-label">{{ t('feature.s2t') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'t2s'" v-model="scriptDir" name="scriptDir" />
          <span class="check-label">{{ t('feature.t2s') }}</span>
        </label>
      </div>
    </div>

    <!-- ========== 空白与换行（混合：互斥 radio + 独立 checkbox） ========== -->
    <div class="feature-group">
      <div class="group-title">{{ t('feature.groupWhitespace') }}</div>

      <!-- 独立的换行→空格 / 合并空格选项 -->
      <div class="clean-grid">
        <label class="check-row">
          <input type="checkbox" v-model="basicClean[newlineToSpaceItem.key]" />
          <span class="check-label">{{ t(newlineToSpaceItem.label) }}</span>
        </label>
        <label class="check-row">
          <input type="checkbox" v-model="basicClean[collapseSpacesItem.key]" />
          <span class="check-label">{{ t(collapseSpacesItem.label) }}</span>
        </label>
      </div>

      <!-- 空白行处理：互斥 radio -->
      <div class="sub-label">{{ t('feature.blankLineMode') }}</div>
      <div class="mutex-grid">
        <label class="radio-row">
          <input type="radio" :value="'none'" v-model="blankLineMode" name="blankLineMode" />
          <span class="check-label">{{ t('feature.blankNone') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'collapse'" v-model="blankLineMode" name="blankLineMode" />
          <span class="check-label">{{ t('feature.blankCollapse') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'remove'" v-model="blankLineMode" name="blankLineMode" />
          <span class="check-label">{{ t('feature.blankRemove') }}</span>
        </label>
        <label class="radio-row">
          <input type="radio" :value="'collapseMax'" v-model="blankLineMode" name="blankLineMode" />
          <span class="check-label">{{ t('feature.blankCollapseMax') }}</span>
        </label>
      </div>

      <!-- collapseMax 选中时：连续空行上限 -->
      <div v-if="blankLineMode === 'collapseMax'" class="sub-field">
        <label class="label">{{ t('feature.maxBlank') }}</label>
        <input
          type="number"
          min="0"
          max="20"
          v-model.number="basicClean.maxBlankLines"
          class="num-input"
        />
      </div>
    </div>

    <!-- 输出换行符（在卡片底部另起一行；不归入上方任一分组） -->
    <div class="feature-group">
      <div class="group-title">{{ t('feature.groupOutput') }}</div>
      <div class="clean-grid">
        <label class="check-row">
          <input type="checkbox" v-model="basicClean.normalizeLineEndings" />
          <span class="check-label">{{ t('feature.normalizeLE') }}</span>
        </label>
      </div>
      <div v-if="basicClean.normalizeLineEndings" class="sub-field">
        <label class="label">{{ t('feature.lineEnding') }}</label>
        <select v-model="basicClean.lineEnding" class="sel-input">
          <option value="keep">{{ t('feature.encKeep') }}</option>
          <option value="lf">LF (Unix)</option>
          <option value="crlf">CRLF (Windows)</option>
        </select>
      </div>
    </div>
  </SectionCard>
</template>

<style scoped>
.clean-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 24px;
}
.mutex-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0 24px;
}
.feature-group {
  margin-top: 14px;
}
.feature-group:first-of-type {
  margin-top: 0;
}
.group-title {
  font-size: 12px;
  color: var(--text-3);
  font-weight: 600;
  letter-spacing: 0.3px;
  margin-bottom: 4px;
}
.sub-label {
  font-size: 11.5px;
  color: var(--text-3);
  margin-top: 10px;
  margin-bottom: 4px;
}
.check-row,
.radio-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  cursor: pointer;
}
.check-label {
  font-size: 13px;
}
.sub-field {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.num-input {
  width: 90px;
}
.sel-input {
  width: 200px;
}

@media (max-width: 720px) {
  .clean-grid {
    grid-template-columns: 1fr;
  }
  .mutex-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
