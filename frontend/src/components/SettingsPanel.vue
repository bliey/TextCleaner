<script setup lang="ts">
import { ref } from 'vue'
import { useSettings } from '../composables/useSettings'
import { useFiles } from '../composables/useFiles'
import { useI18n } from '../i18n'
import alipayQr from '../assets/alipay.png'
import wechatQr from '../assets/wechat.png'
import paypalQr from '../assets/paypal.png'

const { settings, persist, reset, saveError } = useSettings()
// 会话内的“包含子文件夹”是临时开关，设置页改的是它的默认值：
// 保存时同步到当前会话，让“默认值”立刻反映到文件选择区。
const { includeSubfolders } = useFiles()
const { t } = useI18n()

const emit = defineEmits<{ (e: 'close'): void }>()

const savedHint = ref(false)
const sponsorQr = ref<{ name: string; src: string } | null>(null)
let hintTimer: ReturnType<typeof setTimeout> | null = null

const sponsors = [
  { key: 'alipay', nameKey: 'settings.sponsorAlipay', src: alipayQr },
  { key: 'wechat', nameKey: 'settings.sponsorWechat', src: wechatQr },
  { key: 'paypal', nameKey: 'settings.sponsorPaypal', src: paypalQr },
] as const

function showSponsor(name: string, src: string) {
  sponsorQr.value = sponsorQr.value?.src === src ? null : { name, src }
}

async function onSave() {
  await persist()
  includeSubfolders.value = settings.defaultIncludeSubfolders
  savedHint.value = true
  if (hintTimer) clearTimeout(hintTimer)
  hintTimer = setTimeout(() => {
    savedHint.value = false
    emit('close')
  }, 700)
}

function onReset() {
  reset()
}

function onClose() {
  emit('close')
}
</script>

<template>
  <div class="panel-backdrop" @click.self="onClose">
    <div class="panel" role="dialog" :aria-label="t('settings.title')">
      <header class="panel-head">
        <h2 class="panel-title">{{ t('settings.title') }}</h2>
        <button class="btn btn-sm panel-x" :title="t('settings.close')" @click="onClose">×</button>
      </header>

      <div class="panel-body">
        <!-- 主题 -->
        <div class="row-line">
          <div class="row-main">
            <span class="row-label">{{ t('settings.theme') }}</span>
          </div>
          <select v-model="settings.theme" class="sel-input">
            <option value="system">{{ t('settings.themeSystem') }}</option>
            <option value="light">{{ t('settings.themeLight') }}</option>
            <option value="dark">{{ t('settings.themeDark') }}</option>
          </select>
        </div>

        <!-- 语言 -->
        <div class="row-line">
          <div class="row-main">
            <span class="row-label">{{ t('settings.language') }}</span>
            <span class="row-desc">{{ t('settings.languageApplied') }}</span>
          </div>
          <select v-model="settings.language" class="sel-input">
            <option value="auto">{{ t('settings.langAuto') }}</option>
            <option value="zh">{{ t('settings.langZh') }}</option>
            <option value="en">{{ t('settings.langEn') }}</option>
          </select>
        </div>

        <!-- 默认包含子文件夹 -->
        <div class="row-line">
          <div class="row-main">
            <label class="check-inline">
              <input type="checkbox" v-model="settings.defaultIncludeSubfolders" />
              <span class="row-label">{{ t('settings.includeSubfolders') }}</span>
            </label>
            <span class="row-desc">{{ t('settings.includeSubfoldersDesc') }}</span>
          </div>
        </div>

        <!-- 默认并发数 -->
        <div class="row-line">
          <div class="row-main">
            <span class="row-label">{{ t('settings.concurrency') }}</span>
            <span class="row-desc">{{ t('settings.concurrencyDesc') }}</span>
          </div>
          <input
            type="number"
            min="1"
            max="32"
            v-model.number="settings.maxConcurrency"
            class="num-input"
          />
        </div>

        <!-- 赞助：参考 picocrypt-wails，点击支付方式显示二维码，同时只显示一张。 -->
        <div class="sponsor-block">
          <div class="row-main">
            <span class="row-label">{{ t('settings.sponsor') }}</span>
            <span class="row-desc">{{ t('settings.sponsorDesc') }}</span>
          </div>
          <div class="sponsor-links">
            <button
              v-for="item in sponsors"
              :key="item.key"
              class="sponsor-link"
              :class="{ 'is-active': sponsorQr?.src === item.src }"
              type="button"
              @click="showSponsor(t(item.nameKey), item.src)"
            >
              {{ t(item.nameKey) }}
            </button>
          </div>
          <div v-if="sponsorQr" class="sponsor-preview">
            <img :src="sponsorQr.src" :alt="sponsorQr.name" />
            <span>{{ sponsorQr.name }}</span>
          </div>
        </div>

        <p v-if="saveError" class="save-error">⚠ {{ saveError }}</p>
      </div>

      <footer class="panel-foot">
        <button class="btn btn-sm" @click="onReset">{{ t('settings.reset') }}</button>
        <span class="spacer" />
        <span v-if="savedHint" class="saved-hint">✓ {{ t('settings.saved') }}</span>
        <button class="btn btn-sm" @click="onClose">{{ t('settings.close') }}</button>
        <button class="btn btn-primary btn-sm" @click="onSave">{{ t('settings.save') }}</button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.panel-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(16, 24, 40, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  padding: 24px;
}
.panel {
  width: 520px;
  max-width: 100%;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  box-shadow: 0 12px 32px rgba(16, 24, 40, 0.22);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
}
.panel-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}
.panel-x {
  border: none;
  background: transparent;
  font-size: 16px;
  line-height: 1;
  padding: 2px 6px;
  color: var(--text-3);
}
.panel-body {
  padding: 8px 16px 16px;
  display: flex;
  flex-direction: column;
}
.row-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}
.row-line:last-child {
  border-bottom: none;
}
.row-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1 1 auto;
}
.row-label {
  font-size: 13px;
}
.row-desc {
  font-size: 12px;
  color: var(--text-3);
}
.check-inline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.check-inline input {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
}
.sel-input {
  width: 170px;
  flex: 0 0 auto;
}
.num-input {
  width: 90px;
  flex: 0 0 auto;
}
.sponsor-block {
  padding: 14px 0;
  border-bottom: 1px solid var(--border);
}
.sponsor-links {
  display: flex;
  gap: 14px;
  margin-top: 8px;
  flex-wrap: wrap;
}
.sponsor-link {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-3);
  font-size: 12px;
  cursor: pointer;
}
.sponsor-link:hover,
.sponsor-link.is-active {
  color: var(--accent);
  text-decoration: underline;
}
.sponsor-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  width: 160px;
  margin-top: 12px;
  color: var(--text-3);
  font-size: 12px;
}
.sponsor-preview img {
  width: 150px;
  height: 150px;
  padding: 4px;
  object-fit: contain;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 6px;
}
.save-error {
  color: var(--danger);
  font-size: 12.5px;
  margin: 4px 0 0;
}
.panel-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-top: 1px solid var(--border);
  background: var(--surface-2);
}
.saved-hint {
  font-size: 12px;
  color: var(--success);
}
</style>
