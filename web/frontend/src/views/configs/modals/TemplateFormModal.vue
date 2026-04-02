<template>
  <div ref="el" class="modal fade" tabindex="-1">
    <div class="modal-dialog modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ t('configs_page.create_manual_template_title') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
        </div>
        <div class="modal-body">
          <div class="alert alert-warning py-2">
            <div class="fw-semibold">{{ t('configs_page.manual_template_mode_title') }}</div>
            <div class="small mt-1">{{ t('configs_page.manual_template_mode_hint') }}</div>
          </div>
          <div v-if="aiDraftState.active" class="fm-ai-draft-panel mb-3">
            <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
              <div>
                <div class="fw-semibold">{{ t('configs_page.ai_draft_template_title') }}</div>
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
                  <div class="fw-semibold">{{ t('configs_page.ai_draft_diff_lines').replace('{count}', String(aiDraftComparison.lineCount)) }}</div>
                  <div class="small text-muted mt-1">{{ t('configs_page.ai_draft_diff_placeholders').replace('{count}', String(aiDraftComparison.placeholderCount)) }}</div>
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
                    <i class="bi bi-box-arrow-up-right me-1"></i>{{ t('configs_page.ai_draft_action_open_existing_template') }}
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
                  <li v-for="(item, index) in aiDraftState.reviewItems" :key="`tpl-review-${index}`">{{ item }}</li>
                </ul>
              </div>
              <div v-if="aiDraftState.notes.length || aiDraftState.steps.length" class="col-lg-6">
                <div v-if="aiDraftState.steps.length">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_steps_title') }}</div>
                  <ul class="mb-2 ps-3">
                    <li v-for="(step, index) in aiDraftState.steps" :key="`tpl-step-${index}`">{{ step }}</li>
                  </ul>
                </div>
                <div v-if="aiDraftState.notes.length">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_notes_title') }}</div>
                  <ul class="mb-0 ps-3">
                    <li v-for="(note, index) in aiDraftState.notes" :key="`tpl-note-${index}`">{{ note }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
          <div class="row mb-3">
            <div class="col-md-6">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.name') }}</span>
                <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <input v-model="form.name" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }" required>
            </div>
            <div class="col-md-6">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.runtime') }}</span>
                <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <select v-model="form.fluent_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiDraftState.active }">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
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
          <div class="mb-3">
            <label class="form-label d-flex align-items-center gap-2">
              <span>{{ t('configs_page.template_content') }}</span>
              <span v-if="aiDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
            </label>
            <div class="small text-muted mb-2">{{ t('configs_page.template_content_help') }}</div>
            <textarea
              v-model="form.content"
              class="form-control font-monospace fm-config-textarea"
              :class="{ 'fm-ai-draft-highlight': aiDraftState.active }"
              rows="15"
              :placeholder="currentExample"
            ></textarea>
          </div>
          <div class="small text-muted">{{ t('configs_page.template_content_placeholder_hint') }}</div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
          <div v-if="aiDraftState.active && !aiDraftReady" class="small text-muted me-auto">
            {{ t('configs_page.ai_draft_confirm_required') }}
          </div>
          <div v-else-if="aiDraftState.active && aiDraftComparison?.hasConflict" class="small text-warning me-auto">
            {{ t('configs_page.ai_draft_conflict_required') }}
          </div>
          <button type="button" class="btn btn-primary" :disabled="!aiDraftCanSave" @click="$emit('save')">
            {{ aiDraftState.active ? t('configs_page.ai_draft_confirm_template_cta') : t('create') }}
          </button>
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
  aiDraftState: { type: Object, required: true },
  aiDraftSource: { type: String, default: '' },
  aiDraftComparison: { type: Object, default: null },
  aiDraftReady: { type: Boolean, default: false },
  aiDraftCanSave: { type: Boolean, default: true },
  currentExample: { type: String, default: '' },
})
defineEmits(['save', 'apply-suggested-name', 'open-existing'])

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
