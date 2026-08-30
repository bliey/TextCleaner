import { computed } from 'vue'
import { useFiles } from './useFiles'
import { useBatch } from './useBatch'

// ============================================================
// useInteraction —— 全局“交互锁”状态（模块级单例）
//
// 对应需求：参考 picocrypt 的 refreshDisabled() 矩阵——“没有输入时整区禁用”。
// 当没有选中任何文件（或正在处理中）时，规则 / 输出 / 预览等所有功能不可用，
// 仅“文件选择”区保持可用。各区块通过 SectionCard 的 locked 属性接住本状态。
// ============================================================

const { selectedCount } = useFiles()
const { running } = useBatch()

// 未选择任何文件 → 所有处理相关功能不可用
const noInput = computed(() => selectedCount.value === 0)
// 处理中同样锁定编辑，避免运行期改动规则/输出
const locked = computed(() => noInput.value || running.value)

export function useInteraction() {
  return { noInput, locked, running }
}
