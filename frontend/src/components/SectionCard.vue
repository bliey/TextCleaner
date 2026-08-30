<script setup lang="ts">
import { useI18n } from '../i18n'

defineProps<{
  title: string
  hint?: string
  locked?: boolean
}>()

const { t } = useI18n()
</script>

<template>
  <section class="section-card" :class="{ 'section-card--locked': locked }">
    <header class="section-head">
      <h2 class="section-title">{{ title }}</h2>
      <span class="section-head-right">
        <span v-if="locked" class="lock-badge" :title="t('ui.lockedHint')">🔒 {{ t('ui.locked') }}</span>
        <span v-if="hint" class="section-hint">{{ hint }}</span>
      </span>
    </header>
    <!-- inert：锁定后整块内容不可交互、不可聚焦（对应“所有功能不可用”） -->
    <div class="section-body" :inert="locked">
      <slot />
    </div>
  </section>
</template>
