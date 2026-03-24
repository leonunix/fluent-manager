<template>
  <div class="row g-4">
    <div class="col-xl-5">
      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header bg-white">
          <h6 class="mb-0">{{ t('configs_page.ai_assistant') }}</h6>
        </div>
        <div class="card-body">
          <div class="alert alert-info py-2">
            {{ t('configs_page.ai_assistant_intro') }}
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label">{{ t('common.runtime') }}</label>
              <select v-model="state.aiAssistantForm.fluent_type" class="form-select">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('configs_page.module_type_coverage') }}</label>
              <select v-model="state.aiAssistantForm.module_type" class="form-select">
                <option v-for="type in state.moduleTypes" :key="type" :value="type">{{ type }}</option>
              </select>
            </div>
            <div class="col-md-12">
              <label class="form-label">{{ t('configs_page.ai_assistant_goal') }}</label>
              <select v-model="state.aiAssistantForm.goal" class="form-select">
                <option value="module">{{ t('configs_page.ai_assistant_goal_module') }}</option>
                <option value="template">{{ t('configs_page.ai_assistant_goal_template') }}</option>
                <option value="both">{{ t('configs_page.ai_assistant_goal_both') }}</option>
              </select>
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.sample_log') }}</label>
            <textarea
              v-model="state.aiAssistantForm.sample"
              class="form-control font-monospace"
              rows="12"
              :placeholder="t('configs_page.ai_assistant_sample_placeholder')"
            ></textarea>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('common.description') }}</label>
            <textarea
              v-model="state.aiAssistantForm.extra_requirements"
              class="form-control"
              rows="4"
              :placeholder="t('configs_page.ai_assistant_requirements_placeholder')"
            ></textarea>
          </div>
          <button class="btn btn-success w-100" :disabled="state.aiAssistantLoading || !state.aiAssistantForm.sample.trim()" @click="actions.runAIAssistant">
            <i class="bi bi-stars me-1"></i>{{ state.aiAssistantLoading ? t('configs_page.ai_assistant_running') : t('configs_page.ai_assistant_run') }}
          </button>
        </div>
      </div>
    </div>

    <div class="col-xl-7">
      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.ai_assistant_result') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.ai_assistant_result_hint') }}</div>
          </div>
          <div v-if="state.aiAssistantResult" class="d-flex gap-2">
            <button class="btn btn-sm btn-outline-primary" @click="actions.useAIModuleDraft">
              <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.ai_use_module') }}
            </button>
            <button class="btn btn-sm btn-outline-primary" @click="actions.useAITemplateDraft">
              <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.ai_use_template') }}
            </button>
          </div>
        </div>
        <div class="card-body">
          <div
            v-if="state.aiAssistantLoading || state.aiAssistantFeedback.message"
            class="fm-ai-assistant-feedback mb-3"
            :class="{
              'is-success': state.aiAssistantFeedback.type === 'success',
              'is-danger': state.aiAssistantFeedback.type === 'danger',
            }"
          >
            <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
              <div>
                <div class="fw-semibold">
                  {{ state.aiAssistantLoading ? t('configs_page.ai_assistant_running') : state.aiAssistantFeedback.message }}
                </div>
                <div
                  v-if="!state.aiAssistantLoading && state.aiAssistantFeedback.detail && state.aiAssistantFeedback.detail !== state.aiAssistantFeedback.message"
                  class="small text-muted mt-1"
                >
                  {{ state.aiAssistantFeedback.detail }}
                </div>
              </div>
              <div v-if="!state.aiAssistantLoading && state.aiAssistantFeedback.provider" class="small text-muted text-nowrap">
                {{ state.aiAssistantFeedback.provider }}
              </div>
            </div>
            <div
              v-if="!state.aiAssistantLoading && state.aiAssistantFeedback.providerDetail"
              class="small text-muted mt-2"
            >
              {{ t('configs_page.ai_provider_feedback') }}: {{ state.aiAssistantFeedback.providerDetail }}
            </div>
          </div>

          <div v-if="state.aiAssistantResult">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(state.aiAssistantForm.fluent_type) }}</span>
              <span class="badge text-bg-light">{{ state.aiAssistantResult.provider }}</span>
              <span class="badge text-bg-light">{{ state.aiAssistantResult.account_name }}</span>
            </div>

            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <div class="fm-ai-result-box">
                  <div class="fm-ai-result-box__label">{{ t('configs_page.ai_detected_format') }}</div>
                  <div>{{ state.aiAssistantResult.detected_format || '-' }}</div>
                </div>
              </div>
              <div class="col-md-6">
                <div class="fm-ai-result-box">
                  <div class="fm-ai-result-box__label">{{ t('configs_page.ai_summary') }}</div>
                  <div>{{ state.aiAssistantResult.summary || '-' }}</div>
                </div>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.variables_json') }}</label>
              <textarea class="form-control font-monospace fm-config-textarea" rows="7" readonly :value="state.aiAssistantResult.variables_json"></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.version_content') }}</label>
              <textarea class="form-control font-monospace fm-config-textarea" rows="10" readonly :value="state.aiAssistantResult.module_content"></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.template_content') }}</label>
              <textarea class="form-control font-monospace fm-config-textarea" rows="10" readonly :value="state.aiAssistantResult.template_content"></textarea>
            </div>

            <div class="row g-3">
              <div class="col-md-6">
                <div class="fm-ai-result-box">
                  <div class="fm-ai-result-box__label">{{ t('configs_page.ai_assembly_steps') }}</div>
                  <ul class="mb-0">
                    <li v-for="(step, index) in state.aiAssistantResult.assembly_steps || []" :key="index">{{ step }}</li>
                  </ul>
                </div>
              </div>
              <div class="col-md-6">
                <div class="fm-ai-result-box">
                  <div class="fm-ai-result-box__label">{{ t('configs_page.ai_notes') }}</div>
                  <ul class="mb-0">
                    <li v-for="(note, index) in state.aiAssistantResult.notes || []" :key="index">{{ note }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
          <div v-else-if="!state.aiAssistantLoading && !state.aiAssistantFeedback.message" class="text-center text-muted py-5">
            {{ t('configs_page.ai_assistant_empty') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from '../../../i18n'

defineProps({
  state: {
    type: Object,
    required: true,
  },
  actions: {
    type: Object,
    required: true,
  },
  helpers: {
    type: Object,
    required: true,
  },
})

const { t } = useI18n()
</script>
