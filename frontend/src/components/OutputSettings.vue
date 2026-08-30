<script setup lang="ts">
import SectionCard from './SectionCard.vue'
import { useBatch } from '../composables/useBatch'
import { useInteraction } from '../composables/useInteraction'
import { useI18n } from '../i18n'

const { outputFormat, outputEncoding, maxConcurrency } = useBatch()
const { locked } = useInteraction()
const { t } = useI18n()
</script>

<template>
  <SectionCard :title="t('out.title')" :hint="t('out.hint')" :locked="locked">
    <div class="sub-block">
      <div class="group-title">{{ t('out.exportMode') }}</div>
      <label class="mode-option">
        <input type="radio" value="zip" v-model="outputFormat" name="outputFormat" />
        <div class="mode-text">
          <div class="mode-name">{{ t('out.downloadZip') }}</div>
          <div class="mode-desc">{{ t('out.downloadZipDesc') }}</div>
        </div>
      </label>
      <label class="mode-option">
        <input type="radio" value="files" v-model="outputFormat" name="outputFormat" />
        <div class="mode-text">
          <div class="mode-name">{{ t('out.downloadFiles') }}</div>
          <div class="mode-desc">{{ t('out.downloadFilesDesc') }}</div>
        </div>
      </label>
    </div>

    <div class="sub-block">
      <div class="sub-field">
        <label class="label">{{ t('out.encoding') }}</label>
        <select v-model="outputEncoding" class="sel-input">
          <option value="keep">{{ t('out.encKeep') }}</option>
          <option value="utf-8">UTF-8（无 BOM）</option>
          <option value="utf-8-bom">UTF-8（带 BOM）</option>
        </select>
      </div>
      <div class="sub-field">
        <label class="label">{{ t('out.concurrency') }}</label>
        <input type="number" min="1" max="16" v-model.number="maxConcurrency" class="num-input" />
      </div>
    </div>

    <p class="privacy-note">{{ t('out.privacy') }}</p>
  </SectionCard>
</template>

<style scoped>
.sub-block { margin-top: 4px; }
.group-title {
  font-size: 12px;
  color: var(--text-3);
  font-weight: 600;
  margin-bottom: 6px;
}
.mode-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 10px;
  margin-top: 8px;
  border: 1px solid var(--border);
  border-radius: 8px;
  cursor: pointer;
}
.mode-option:has(input:checked) {
  border-color: var(--accent);
  background: var(--accent-weak);
}
.mode-option input { margin-top: 3px; accent-color: var(--accent); }
.mode-text { flex: 1; }
.mode-name { font-size: 13px; font-weight: 600; }
.mode-desc { margin-top: 2px; font-size: 12px; color: var(--text-3); }
.sub-field {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
}
.label { min-width: 90px; color: var(--text-2); font-size: 13px; }
.num-input { width: 90px; }
.sel-input { width: 200px; }
.privacy-note {
  margin: 14px 0 0;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--accent-weak);
  color: var(--text-2);
  font-size: 12px;
}
</style>
