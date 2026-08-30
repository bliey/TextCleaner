<script setup lang="ts">
import { useI18n } from '../i18n'
import type { PendingPlan } from '../composables/useBatch'
import { isDesktop } from '../core/env'

const { t } = useI18n()
const props = defineProps<{ plan: PendingPlan }>()
const emit = defineEmits<{ (e: 'confirm'): void; (e: 'cancel'): void }>()
</script>

<template>
  <div class="overlay" @click.self="emit('cancel')">
    <div class="dialog" role="dialog" aria-modal="true">
      <header class="dlg-header">
        <div class="dlg-title">{{ t('plan.title') }}</div>
        <button class="dlg-close" :aria-label="t('plan.cancel')" @click="emit('cancel')">✕</button>
      </header>
      <div class="dlg-body">
        <div class="row big">
          <span class="label">{{ t('plan.fileCount') }}</span>
          <span class="value">{{ props.plan.fileCount }}</span>
        </div>
        <div class="divider" />
        <div class="row">
          <span class="label">{{ t('plan.export') }}</span>
          <span class="value mode">
            {{ props.plan.outputFormat === 'zip' ? t('out.downloadZip') : (isDesktop() ? t('out.exportFolder') : t('out.downloadFiles')) }}
          </span>
        </div>
        <div class="row privacy">
          <span class="value">{{ t('out.privacy') }}</span>
        </div>
      </div>
      <footer class="dlg-footer">
        <button class="btn" @click="emit('cancel')">{{ t('plan.cancel') }}</button>
        <button class="btn btn-primary" @click="emit('confirm')">{{ t('plan.confirm') }}</button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, .45); display: flex; align-items: center; justify-content: center; z-index: 100; }
.dialog { width: min(520px, 92vw); background: var(--surface); border: 1px solid var(--border); border-radius: 10px; box-shadow: 0 12px 40px rgba(0, 0, 0, .18); overflow: hidden; }
.dlg-header { display: flex; align-items: center; justify-content: space-between; padding: 14px 18px; border-bottom: 1px solid var(--border); }
.dlg-title { font-size: 15px; font-weight: 600; }
.dlg-close { background: none; border: none; font-size: 16px; color: var(--text-3); cursor: pointer; padding: 4px 8px; border-radius: 4px; }
.dlg-close:hover { background: var(--surface-2); color: var(--text); }
.dlg-body { padding: 16px 18px; }
.row { display: flex; align-items: flex-start; gap: 12px; padding: 6px 0; font-size: 13px; }
.row.big { font-size: 14px; padding: 4px 0 10px; }
.label { min-width: 90px; color: var(--text-3); }
.value { flex: 1; color: var(--text); }
.value.mode { font-weight: 600; }
.privacy { margin-top: 10px; padding: 8px 10px; border-radius: 6px; background: var(--accent-weak); }
.divider { height: 1px; background: var(--border); margin: 4px 0; }
.dlg-footer { display: flex; justify-content: flex-end; gap: 8px; padding: 12px 18px; background: var(--surface-2); border-top: 1px solid var(--border); }
</style>
