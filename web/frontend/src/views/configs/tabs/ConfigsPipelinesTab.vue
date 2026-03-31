<template>
  <div>
    <!-- Search + filter bar -->
    <div class="d-flex flex-wrap gap-2 align-items-center mb-3">
      <div class="input-group" style="max-width:280px">
        <span class="input-group-text"><i class="bi bi-search"></i></span>
        <input
          v-model="localSearch"
          type="text"
          class="form-control"
          :placeholder="t('common.search')"
        />
      </div>
      <select
        v-model="localFluentType"
        class="form-select"
        style="max-width:160px"
      >
        <option value="">{{ t('common.all') }} Runtime</option>
        <option value="fluentbit">Fluent Bit</option>
        <option value="fluentd">Fluentd</option>
      </select>
    </div>

    <!-- Pipeline list -->
    <div v-if="filteredPipelines.length" class="row g-3">
      <div v-for="pipeline in filteredPipelines" :key="pipeline.id" class="col-xl-6">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start gap-2 mb-2">
              <div>
                <div class="fw-semibold">{{ pipeline.name }}</div>
                <div v-if="pipeline.description" class="small text-muted mt-1">{{ pipeline.description }}</div>
              </div>
              <div class="d-flex gap-1 flex-shrink-0">
                <span class="badge" :class="helpers.runtimeBadgeClass(pipeline.fluent_type)">
                  {{ helpers.runtimeLabel(pipeline.fluent_type) }}
                </span>
              </div>
            </div>

            <!-- Input module -->
            <div class="d-flex flex-wrap gap-2 mb-2 small">
              <div class="d-flex align-items-center gap-1">
                <span class="badge bg-primary-subtle text-primary-emphasis">{{ t('configs_page.pipeline_stage_input') }}</span>
                <span class="font-monospace">{{ pipeline.input_module?.name || '—' }}</span>
              </div>
            </div>

            <!-- Filter modules -->
            <div v-if="pipeline.filter_modules && pipeline.filter_modules.length" class="d-flex flex-wrap gap-1 mb-2">
              <span class="badge bg-secondary-subtle text-secondary-emphasis">{{ t('configs_page.pipeline_stage_filter') }}</span>
              <span
                v-for="(m, i) in pipeline.filter_modules"
                :key="m.id"
                class="badge bg-light text-dark border font-monospace"
              >{{ i + 1 }}. {{ m.name }}</span>
            </div>

            <!-- Output targets -->
            <div v-if="pipeline.output_targets && pipeline.output_targets.length" class="d-flex flex-wrap gap-1 mb-3">
              <span class="badge bg-success-subtle text-success-emphasis">{{ t('configs_page.pipeline_stage_output') }}</span>
              <span
                v-for="t_ in pipeline.output_targets"
                :key="t_.id"
                class="badge bg-light text-dark border font-monospace"
              >{{ t_.name }}</span>
            </div>

            <div class="d-flex justify-content-between align-items-center">
              <div class="small text-muted">{{ helpers.formatTime(pipeline.created_at) }}</div>
              <div class="d-flex gap-1">
                <button class="btn btn-sm btn-outline-secondary" @click="actions.openEditPipeline(pipeline)">
                  <i class="bi bi-pencil me-1"></i>{{ t('common.edit') }}
                </button>
                <button class="btn btn-sm btn-outline-danger" @click="actions.handleDeletePipeline(pipeline)">
                  <i class="bi bi-trash"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center text-muted py-5">
      <i class="bi bi-diagram-3 fs-2 d-block mb-2 opacity-25"></i>
      {{ t('configs_page.pipeline_no_pipelines') }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from '../../../i18n'

const props = defineProps({
  state: { type: Object, required: true },
  actions: { type: Object, required: true },
  helpers: { type: Object, required: true },
})

const { t } = useI18n()

const localSearch = ref('')
const localFluentType = ref('')

const filteredPipelines = computed(() => {
  let list = props.state.pipelines || []
  const q = localSearch.value.trim().toLowerCase()
  if (q) list = list.filter((p) => p.name.toLowerCase().includes(q) || (p.description || '').toLowerCase().includes(q))
  if (localFluentType.value) list = list.filter((p) => p.fluent_type === localFluentType.value)
  return list
})
</script>
