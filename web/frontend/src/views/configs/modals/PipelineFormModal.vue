<template>
  <div ref="el" class="modal fade" tabindex="-1">
    <div class="modal-dialog modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ editingId ? t('configs_page.edit_pipeline_title') : t('configs_page.create_pipeline') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
        </div>
        <div class="modal-body">
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label">{{ t('common.name') }} <span class="text-danger">*</span></label>
              <input v-model="form.name" type="text" class="form-control" :placeholder="t('configs_page.pipeline_name_placeholder')" />
            </div>
            <div class="col-md-6">
              <label class="form-label">Runtime</label>
              <select v-model="form.fluent_type" class="form-select">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('common.description') }}</label>
            <input v-model="form.description" type="text" class="form-control" />
          </div>

          <!-- Input module -->
          <div class="mb-3">
            <label class="form-label fw-semibold">{{ t('configs_page.pipeline_input') }}</label>
            <select v-model="form.input_module_id" class="form-select">
              <option :value="null">— {{ t('none') }} —</option>
              <option v-for="m in inputModules" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
          </div>

          <!-- Filter modules -->
          <div class="mb-3">
            <label class="form-label fw-semibold">{{ t('configs_page.pipeline_filters') }}</label>
            <div v-if="form.filter_module_ids.length" class="list-group mb-2">
              <div
                v-for="(fid, idx) in form.filter_module_ids"
                :key="fid"
                class="list-group-item d-flex justify-content-between align-items-center py-1 px-2"
              >
                <span class="badge bg-secondary me-2">{{ idx + 1 }}</span>
                <span class="flex-grow-1 font-monospace small">{{ filterModules.find((m) => m.id === fid)?.name || fid }}</span>
                <div class="d-flex gap-1 ms-2">
                  <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" :disabled="idx === 0" @click="$emit('move-filter', idx, -1)"><i class="bi bi-arrow-up"></i></button>
                  <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" :disabled="idx === form.filter_module_ids.length - 1" @click="$emit('move-filter', idx, 1)"><i class="bi bi-arrow-down"></i></button>
                  <button type="button" class="btn btn-sm btn-outline-danger py-0 px-1" @click="$emit('remove-filter', idx)"><i class="bi bi-x"></i></button>
                </div>
              </div>
            </div>
            <div class="input-group">
              <select v-model="filterPickerValue" class="form-select form-select-sm">
                <option value="">{{ t('configs_page.pipeline_add_filter') }}…</option>
                <option v-for="m in filterModules" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
              <button type="button" class="btn btn-sm btn-outline-secondary" @click="addFilter">
                <i class="bi bi-plus"></i>
              </button>
            </div>
          </div>

          <!-- Output targets -->
          <div class="mb-3">
            <label class="form-label fw-semibold">{{ t('configs_page.pipeline_outputs') }}</label>
            <div class="d-flex flex-wrap gap-2">
              <label
                v-for="target in availableOutputTargets"
                :key="target.id"
                class="d-flex align-items-center gap-1 border rounded px-2 py-1 small"
                style="cursor:pointer"
              >
                <input type="checkbox" :checked="form.output_target_ids.includes(target.id)" @change="$emit('toggle-output', target.id)" />
                {{ target.name }}
              </label>
              <div v-if="!availableOutputTargets.length" class="text-muted small">{{ t('configs_page.no_output_targets') }}</div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
          <button type="button" class="btn btn-primary" @click="$emit('save')">{{ t('save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../../i18n'

const props = defineProps({
  form: { type: Object, required: true },
  editingId: { type: Number, default: null },
  inputModules: { type: Array, default: () => [] },
  filterModules: { type: Array, default: () => [] },
  availableOutputTargets: { type: Array, default: () => [] },
})
const emit = defineEmits(['save', 'add-filter', 'remove-filter', 'move-filter', 'toggle-output'])

const { t } = useI18n()
const el = ref(null)
const filterPickerValue = ref('')
let modal = null

onMounted(() => {
  modal = new window.bootstrap.Modal(el.value)
})

function addFilter() {
  if (filterPickerValue.value) {
    emit('add-filter', Number(filterPickerValue.value))
    filterPickerValue.value = ''
  }
}

defineExpose({
  show: () => { filterPickerValue.value = ''; modal?.show() },
  hide: () => modal?.hide(),
})
</script>
