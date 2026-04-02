<template>
  <div ref="el" class="modal fade" tabindex="-1">
    <div class="modal-dialog modal-xl">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ editingId ? t('configs_page.edit_module_title') : t('configs_page.create_module_title') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
        </div>
        <div class="modal-body">
          <div class="alert alert-info py-2">{{ t('configs_page.module_modal_hint') }}</div>
          <div v-if="aiDraftState.active" class="fm-ai-draft-panel mb-3">
            <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
              <div>
                <div class="fw-semibold">{{ t('configs_page.ai_draft_module_title') }}</div>
                <div v-if="aiDraftSource" class="small text-muted mt-1">{{ aiDraftSource }}</div>
              </div>
              <span class="badge bg-success-subtle text-success-emphasis">{{ t('configs_page.ai_draft_imported') }}</span>
            </div>
            <div v-if="aiDraftState.summary" class="small mt-2">{{ aiDraftState.summary }}</div>
            <div v-if="aiDraftComparison" class="mt-3">
              <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_diff_title') }}</div>
              <div class="fm-ai-draft-diff-grid">
                <div class="fm-ai-draft-diff-card">
                  <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_identity') }}</div>
                  <div class="fw-semibold">{{ aiDraftComparison.identityMessage }}</div>
                  <div v-if="aiDraftComparison.existingName" class="small text-muted mt-1">{{ aiDraftComparison.existingName }}</div>
                </div>
                <div class="fm-ai-draft-diff-card">
                  <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_scale') }}</div>
                  <div class="fw-semibold">{{ t('configs_page.ai_draft_diff_variables').replace('{count}', String(aiDraftComparison.variableCount)) }}</div>
                  <div class="small text-muted mt-1">
                    {{ t('configs_page.ai_draft_diff_lines').replace('{count}', String(aiDraftComparison.lineCount)) }}
                    · {{ t('configs_page.ai_draft_diff_placeholders').replace('{count}', String(aiDraftComparison.placeholderCount)) }}
                  </div>
                </div>
                <div v-if="aiDraftComparison.changeMessage" class="fm-ai-draft-diff-card">
                  <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_changes') }}</div>
                  <div class="fw-semibold">{{ aiDraftComparison.changeMessage }}</div>
                  <div v-if="aiDraftComparison.changeDetail" class="small text-muted mt-1">{{ aiDraftComparison.changeDetail }}</div>
                </div>
              </div>
              <div v-if="aiDraftComparison.hasConflict" class="fm-ai-draft-actions mt-3">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_action_title') }}</div>
                <div class="small text-muted mb-2">
                  {{ t('configs_page.ai_draft_action_existing_detail').replace('{name}', aiDraftComparison.existingName || form.name) }}
                </div>
                <div v-if="aiDraftComparison.suggestedName" class="small text-muted mb-3">
                  {{ t('configs_page.ai_draft_action_name_suggested').replace('{name}', aiDraftComparison.suggestedName) }}
                </div>
                <div class="d-flex flex-wrap gap-2">
                  <button type="button" class="btn btn-sm btn-outline-primary" @click="$emit('apply-suggested-name')">
                    <i class="bi bi-magic me-1"></i>{{ t('configs_page.ai_draft_action_auto_rename') }}
                  </button>
                  <button type="button" class="btn btn-sm btn-outline-secondary" @click="$emit('open-existing')">
                    <i class="bi bi-pencil-square me-1"></i>{{ t('configs_page.ai_draft_action_open_existing_module') }}
                  </button>
                </div>
              </div>
            </div>
            <div v-if="aiDraftState.confirmationItems.length" class="mt-3">
              <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_checklist_title') }}</div>
              <div class="small text-muted mb-2">{{ t('configs_page.ai_draft_checklist_hint') }}</div>
              <div class="fm-ai-draft-checklist">
                <label v-for="item in aiDraftState.confirmationItems" :key="item.key" class="form-check fm-ai-draft-checklist__item">
                  <input v-model="item.checked" class="form-check-input" type="checkbox">
                  <span class="form-check-label">{{ item.label }}</span>
                </label>
              </div>
            </div>
            <div class="row g-3 mt-1">
              <div v-if="aiDraftState.reviewItems.length" class="col-lg-6">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_review_title') }}</div>
                <ul class="mb-0 ps-3">
                  <li v-for="(item, index) in aiDraftState.reviewItems" :key="`mod-review-${index}`">{{ item }}</li>
                </ul>
              </div>
              <div v-if="aiDraftState.notes.length || aiDraftState.steps.length" class="col-lg-6">
                <div v-if="aiDraftState.steps.length">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_steps_title') }}</div>
                  <ul class="mb-2 ps-3">
                    <li v-for="(step, index) in aiDraftState.steps" :key="`mod-step-${index}`">{{ step }}</li>
                  </ul>
                </div>
                <div v-if="aiDraftState.notes.length">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_notes_title') }}</div>
                  <ul class="mb-0 ps-3">
                    <li v-for="(note, index) in aiDraftState.notes" :key="`mod-note-${index}`">{{ note }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <div class="row g-3 mb-3">
            <div class="col-md-4">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.name') }}</span>
                <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <input v-model="form.name" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }">
            </div>
            <div class="col-md-4">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('configs_page.module_type_coverage') }}</span>
                <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <select v-model="form.module_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }">
                <option v-for="type in moduleTypes" :key="type" :value="type">{{ type }}</option>
              </select>
            </div>
            <div class="col-md-4">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.runtime') }}</span>
                <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <select v-model="form.fluent_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
                <option value="shared">Shared</option>
              </select>
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label d-flex align-items-center gap-2">
              <span>{{ t('common.description') }}</span>
              <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
            </label>
            <input v-model="form.description" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }">
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-8">
              <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
                <label class="form-label mb-0 d-flex align-items-center gap-2">
                  <span>{{ t('configs_page.variables_json') }}</span>
                  <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <div class="btn-group btn-group-sm" role="group" :aria-label="t('configs_page.variable_input_mode')">
                  <button type="button" class="btn" :class="variablesMode === 'form' ? 'btn-primary' : 'btn-outline-primary'" @click="$emit('set-variables-mode', 'form')">
                    {{ t('configs_page.variable_mode_form') }}
                  </button>
                  <button type="button" class="btn" :class="variablesMode === 'json' ? 'btn-primary' : 'btn-outline-primary'" @click="$emit('set-variables-mode', 'json')">
                    {{ t('configs_page.variable_mode_json') }}
                  </button>
                </div>
              </div>
              <div class="small text-muted mb-2">{{ t('configs_page.variables_help') }}</div>
              <div v-if="variablesMode === 'form'" class="border rounded-3 p-3 bg-light-subtle">
                <div class="d-flex justify-content-between align-items-center mb-3">
                  <div class="small text-muted">{{ t('configs_page.variable_form_help') }}</div>
                  <button type="button" class="btn btn-sm btn-outline-secondary" @click="$emit('add-variable-row')">
                    <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.add_variable') }}
                  </button>
                </div>
                <div v-if="!variableRows.length" class="text-center text-muted py-3">{{ t('configs_page.no_variables_rows') }}</div>
                <div v-for="(row, index) in variableRows" :key="index" class="row g-2 align-items-start mb-2">
                  <div class="col-md-4">
                    <input v-model="row.key" type="text" class="form-control" :placeholder="t('configs_page.variable_name_placeholder')" @input="$emit('sync-variables')">
                  </div>
                  <div class="col-md-3">
                    <select v-model="row.type" class="form-select" @change="$emit('sync-variables')">
                      <option value="string">{{ t('configs_page.variable_type_string') }}</option>
                      <option value="number">{{ t('configs_page.variable_type_number') }}</option>
                      <option value="boolean">{{ t('configs_page.variable_type_boolean') }}</option>
                      <option value="json">{{ t('configs_page.variable_type_json') }}</option>
                    </select>
                  </div>
                  <div class="col-md-4">
                    <select v-if="row.type === 'boolean'" v-model="row.value" class="form-select" @change="$emit('sync-variables')">
                      <option value="true">true</option>
                      <option value="false">false</option>
                    </select>
                    <textarea v-else-if="row.type === 'json'" v-model="row.value" class="form-control font-monospace" rows="3" :placeholder="t('configs_page.variable_json_placeholder')" @input="$emit('sync-variables')"></textarea>
                    <input v-else v-model="row.value" type="text" class="form-control" :placeholder="row.type === 'number' ? '24224' : t('configs_page.variable_value_placeholder')" @input="$emit('sync-variables')">
                  </div>
                  <div class="col-md-1 d-grid">
                    <button type="button" class="btn btn-outline-danger" @click="$emit('remove-variable-row', index)"><i class="bi bi-trash"></i></button>
                  </div>
                </div>
                <div v-if="variablesError" class="small text-danger mt-2">{{ variablesError }}</div>
              </div>
              <textarea v-else v-model="form.variables" class="form-control font-monospace fm-config-textarea" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }" rows="5" :placeholder="currentExample.variables"></textarea>
            </div>
            <div class="col-md-4">
              <label class="form-label d-block">Props</label>
              <div class="form-check mt-2">
                <input id="moduleBuiltin" v-model="form.is_builtin" type="checkbox" class="form-check-input">
                <label for="moduleBuiltin" class="form-check-label">{{ t('configs_page.builtin_module') }}</label>
              </div>
              <div class="small text-muted mt-2">{{ t('configs_page.builtin_help') }}</div>
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label d-flex align-items-center gap-2">
              <span>{{ form.fluent_type === 'shared' ? 'Fluent Bit ' + t('configs_page.version_content') : t('configs_page.version_content') }}</span>
              <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
            </label>
            <div class="small text-muted mb-2">{{ t('configs_page.version_content_help') }}</div>
            <textarea v-model="form.content" class="form-control font-monospace fm-config-textarea" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }" rows="16" :placeholder="currentExample.content"></textarea>
          </div>
          <div v-if="form.fluent_type === 'shared'" class="mb-3">
            <label class="form-label">Fluentd {{ t('configs_page.version_content') }}</label>
            <div class="small text-muted mb-2">{{ t('configs_page.version_content_help') }}</div>
            <textarea v-model="form.content_fluentd" class="form-control font-monospace fm-config-textarea" rows="16" :placeholder="(moduleExamples.fluentd?.[form.module_type] || moduleExamples.fluentd?.input || {}).content || ''"></textarea>
            <div class="small text-muted mt-1">{{ t('configs_page.module_shared_fluentd_hint', 'Leave empty to reuse Fluent Bit content for Fluentd.') }}</div>
          </div>
          <div class="small text-muted">{{ t('configs_page.template_syntax_hint') }}</div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
          <div v-if="aiDraftState.active && !aiDraftReady" class="small text-muted me-auto">{{ t('configs_page.ai_draft_confirm_required') }}</div>
          <div v-else-if="aiDraftState.active && aiDraftComparison?.hasConflict && !editingId" class="small text-warning me-auto">{{ t('configs_page.ai_draft_conflict_required') }}</div>
          <button type="button" class="btn btn-primary" :disabled="!aiDraftCanSave" @click="$emit('save')">
            {{ editingId ? t('save') : (aiDraftState.active ? t('configs_page.ai_draft_confirm_module_cta') : t('configs_page.create_module')) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../../i18n'

defineProps({
  form: { type: Object, required: true },
  editingId: { type: Number, default: null },
  aiDraftState: { type: Object, required: true },
  aiDraftSource: { type: String, default: '' },
  aiDraftComparison: { type: Object, default: null },
  aiDraftReady: { type: Boolean, default: false },
  aiDraftCanSave: { type: Boolean, default: true },
  moduleTypes: { type: Array, default: () => [] },
  moduleExamples: { type: Object, default: () => ({}) },
  currentExample: { type: Object, default: () => ({ variables: '{}', content: '' }) },
  variablesMode: { type: String, default: 'form' },
  variableRows: { type: Array, default: () => [] },
  variablesError: { type: String, default: '' },
})
defineEmits(['save', 'apply-suggested-name', 'open-existing', 'set-variables-mode', 'add-variable-row', 'remove-variable-row', 'sync-variables'])

const { t } = useI18n()
const el = ref(null)
let modal = null

onMounted(() => {
  modal = new window.bootstrap.Modal(el.value)
})

defineExpose({
  show: () => modal?.show(),
  hide: () => modal?.hide(),
})
</script>
