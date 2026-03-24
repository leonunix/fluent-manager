<template>
  <div class="fm-assembly-flow">
    <div class="fm-assembly-flow__rail">
      <template v-for="(stage, index) in visibleStages" :key="stage.key">
        <div class="fm-assembly-flow__stage">
          <div class="fm-assembly-flow__label">{{ stage.label }}</div>
          <div class="fm-assembly-flow__value">{{ stage.value }}</div>
          <div class="fm-assembly-flow__meta">{{ stage.items.length }} {{ t('configs_page.flow_stage_items') }}</div>
        </div>
        <div v-if="index < visibleStages.length - 1" class="fm-assembly-flow__arrow">
          <i class="bi bi-arrow-right"></i>
        </div>
      </template>
    </div>

    <div class="small text-muted mt-3">
      {{ resolvedPathLabel }}
    </div>

    <div v-if="sortedModules.length" class="d-flex flex-wrap gap-2 mt-3">
      <span
        v-for="module in sortedModules"
        :key="module._key"
        class="badge rounded-pill text-bg-light"
      >
        {{ module.module_type }} · {{ module.name }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from '../i18n'
import { buildConfigFlowSummary } from '../utils/config_flow'

const props = defineProps({
  modules: {
    type: Array,
    default: () => [],
  },
  destinations: {
    type: Array,
    default: () => [],
  },
  pathLabel: {
    type: String,
    default: '',
  },
})

const { t } = useI18n()

const moduleTypeOrder = {
  service: 0,
  input: 1,
  parser: 2,
  filter: 3,
  route: 4,
  output: 5,
}

const normalizedModules = computed(() => (props.modules || [])
  .filter(Boolean)
  .map((item, index) => ({
    _key: `${item.id || item.module_id || 'module'}-${index}`,
    id: item.id || item.module_id || index,
    name: item.name || item.module_name || `module-${index + 1}`,
    module_type: item.module_type || 'module',
  })))

const sortedModules = computed(() => [...normalizedModules.value].sort((left, right) => {
  const leftOrder = moduleTypeOrder[left.module_type] ?? 99
  const rightOrder = moduleTypeOrder[right.module_type] ?? 99
  if (leftOrder !== rightOrder) return leftOrder - rightOrder
  return left.name.localeCompare(right.name)
}))

const normalizedDestinations = computed(() => (props.destinations || [])
  .filter(Boolean)
  .map((item) => (typeof item === 'string' ? { name: item } : item)))

const summary = computed(() => buildConfigFlowSummary(normalizedModules.value, normalizedDestinations.value))

function summarize(items = []) {
  if (!items.length) return '-'
  if (items.length === 1) return items[0]
  if (items.length === 2) return items.join(' + ')
  return `${items[0]} + ${items[1]} + ...`
}

const visibleStages = computed(() => {
  const stages = [
    { key: 'service', label: t('configs_page.flow_stage_service'), items: summary.value.service || [] },
    { key: 'input', label: t('configs_page.flow_stage_input'), items: summary.value.inputs || [] },
    { key: 'processing', label: t('configs_page.flow_stage_processing'), items: summary.value.processors || [] },
    { key: 'output', label: t('configs_page.flow_stage_output'), items: summary.value.outputs || [] },
    { key: 'destination', label: t('configs_page.flow_stage_destination'), items: summary.value.destinationChips || [] },
  ]

  return stages
    .filter((stage) => stage.items.length)
    .map((stage) => ({
      ...stage,
      value: summarize(stage.items),
    }))
})

const resolvedPathLabel = computed(() => (
  props.pathLabel ||
  (summary.value.path.length ? summary.value.path.join(' -> ') : t('configs_page.no_solution_path'))
))
</script>

<style scoped>
.fm-assembly-flow {
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background:
    radial-gradient(circle at top left, rgba(45, 212, 191, 0.14), transparent 34%),
    linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  padding: 1rem;
}

.fm-assembly-flow__rail {
  display: flex;
  align-items: stretch;
  gap: 0.75rem;
  overflow-x: auto;
}

.fm-assembly-flow__stage {
  min-width: 170px;
  padding: 0.9rem 1rem;
  border-radius: 16px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.06);
}

.fm-assembly-flow__label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #0f766e;
  margin-bottom: 0.45rem;
}

.fm-assembly-flow__value {
  font-weight: 700;
  color: #0f172a;
  line-height: 1.35;
}

.fm-assembly-flow__meta {
  margin-top: 0.45rem;
  font-size: 0.78rem;
  color: #64748b;
}

.fm-assembly-flow__arrow {
  display: flex;
  align-items: center;
  color: #0d9488;
  font-size: 1.1rem;
  flex: 0 0 auto;
}

@media (max-width: 767.98px) {
  .fm-assembly-flow__stage {
    min-width: 150px;
  }
}
</style>
