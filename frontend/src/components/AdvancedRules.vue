<script setup lang="ts">
import { ref } from 'vue'
import SectionCard from './SectionCard.vue'
import { useOptions } from '../composables/useOptions'
import { useInteraction } from '../composables/useInteraction'

const { advancedRules, addRegexRule, removeRegexRule } = useOptions()
const { locked } = useInteraction()

const open = ref(false)
</script>

<template>
  <SectionCard title="高级规则" hint="正则表达式（可折叠）" :locked="locked">
    <button class="btn btn-sm toggle" @click="open = !open">
      {{ open ? '收起 ▲' : '展开正则替换 ▼' }}
    </button>

    <div v-if="open" class="adv">
      <p class="muted adv-note">
        每条规则均按 Go 正则（RE2 语法）执行查找 / 替换；非法正则会在处理时被跳过并记录。
      </p>
      <div class="tbl-head">
        <span class="c-ena">启用</span>
        <span class="c-find">正则查找</span>
        <span class="c-rep">替换为</span>
        <span class="c-act" />
      </div>
      <div v-for="(r, i) in advancedRules" :key="i" class="tbl-row">
        <span class="c-ena">
          <input type="checkbox" v-model="r.enabled" />
        </span>
        <span class="c-find">
          <input type="text" v-model="r.find" placeholder="如 \d{4}-\d{2}-\d{2}" />
        </span>
        <span class="c-rep">
          <input type="text" v-model="r.replace" placeholder="如 20XX-$0" />
        </span>
        <span class="c-act">
          <button class="btn btn-sm btn-danger" @click="removeRegexRule(i)">×</button>
        </span>
      </div>
      <button class="btn btn-sm add-row" @click="addRegexRule">+ 添加正则规则</button>
    </div>
  </SectionCard>
</template>

<style scoped>
.toggle {
  margin-bottom: 4px;
}
.adv {
  margin-top: 10px;
}
.adv-note {
  margin: 0 0 12px;
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
  margin-top: 8px;
}
</style>
