<template>
  <div class="row g-4">
    <div class="col-xl-5">
      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white">
          <h6 class="mb-0">{{ t('configs_page.import_existing') }}</h6>
        </div>
        <div class="card-body">
          <div class="alert alert-info py-2">
            {{ t('configs_page.import_intro') }}
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label">{{ t('common.runtime') }}</label>
              <select v-model="state.importForm.fluent_type" class="form-select">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('configs_page.import_name_prefix') }}</label>
              <input v-model="state.importForm.name_prefix" type="text" class="form-control" :placeholder="t('configs_page.import_name_prefix_placeholder')">
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.import_content') }}</label>
            <textarea
              v-model="state.importForm.content"
              class="form-control font-monospace fm-config-textarea"
              rows="18"
              :placeholder="t('configs_page.import_content_placeholder')"
            ></textarea>
          </div>
          <button class="btn btn-success w-100" :disabled="state.importAnalysisLoading || !state.importForm.content.trim()" @click="actions.runImportAnalysis">
            <i class="bi bi-file-earmark-arrow-up me-1"></i>{{ state.importAnalysisLoading ? t('configs_page.import_analyzing') : t('configs_page.import_analyze') }}
          </button>
        </div>
      </div>
    </div>

    <div class="col-xl-7">
      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.import_result') }}</h6>
            <div class="small text-muted mt-1">
              {{
                state.importedConfigResult && !state.importedConfigResult.auto_assemble_supported
                  ? t('configs_page.import_result_hint_assets')
                  : t('configs_page.import_result_hint')
              }}
            </div>
          </div>
          <button
            v-if="state.importedConfigResult?.modules?.length"
            class="btn btn-sm btn-outline-primary"
            :disabled="state.importModulesLoading || state.importBlockingIssueCount > 0"
            @click="actions.importParsedModules"
          >
            <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ state.importModulesLoading ? t('configs_page.import_persisting') : (state.importedConfigResult?.auto_assemble_supported ? t('configs_page.import_persist_modules') : t('configs_page.import_persist_assets')) }}
          </button>
        </div>
        <div class="card-body">
          <div v-if="state.importedConfigResult">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(state.importedConfigResult.fluent_type) }}</span>
              <span class="badge text-bg-light">{{ state.importedConfigResult.modules.length }} {{ t('configs_page.modules') }}</span>
              <span class="badge text-bg-light">{{ state.importedConfigResult.suggested_template_name }}</span>
            </div>
            <div class="small text-muted mb-3">{{ state.importedConfigResult.summary }}</div>
            <div
              v-if="!state.importedConfigResult.auto_assemble_supported"
              class="alert alert-info py-2"
            >
              <div class="fw-semibold mb-1">{{ t('configs_page.import_workspace_assets_title') }}</div>
              <div>{{ t('configs_page.import_workspace_assets_hint') }}</div>
            </div>

            <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-3">
              <div class="d-flex flex-wrap gap-2">
                <span class="badge text-bg-light">
                  {{ t('configs_page.import_decision_reuse').replace('{count}', String(state.importReuseDecisionCount)) }}
                </span>
                <span class="badge text-bg-light">
                  {{ t('configs_page.import_decision_create').replace('{count}', String(state.importCreateDecisionCount)) }}
                </span>
                <span class="badge text-bg-light">
                  {{ t('configs_page.import_existing_matches').replace('{count}', String(state.importReusableMatchCount)) }}
                </span>
                <span class="badge text-bg-light">
                  {{ t('configs_page.import_destination_matches').replace('{count}', String(state.importDestinationMatchCount)) }}
                </span>
              </div>
              <div class="btn-group btn-group-sm" role="group">
                <button
                  type="button"
                  class="btn btn-outline-success"
                  :disabled="!state.importReusableMatchCount"
                  @click="actions.setAllImportedModuleActions('reuse_existing')"
                >
                  {{ t('configs_page.import_apply_reuse_all') }}
                </button>
                <button
                  type="button"
                  class="btn btn-outline-primary"
                  @click="actions.setAllImportedModuleActions('create_new')"
                >
                  {{ t('configs_page.import_apply_create_all') }}
                </button>
              </div>
            </div>
            <div class="small text-muted mb-3">{{ t('configs_page.import_batch_decision_hint') }}</div>

            <div v-if="state.importBlockingIssueCount > 0" class="alert alert-warning py-2">
              <div class="fw-semibold mb-1">{{ t('configs_page.import_name_conflict_title') }}</div>
              <div>{{ t('configs_page.import_name_conflict_hint').replace('{count}', String(state.importBlockingIssueCount)) }}</div>
            </div>

            <div v-if="state.importedConfigResult.warnings?.length" class="alert alert-warning py-2">
              <div class="fw-semibold mb-1">{{ t('configs_page.import_warnings') }}</div>
              <div v-for="(warning, index) in state.importedConfigResult.warnings" :key="`import-warning-${index}`">
                {{ warning }}
              </div>
            </div>

            <div class="card border-0 bg-light-subtle mb-3">
              <div class="card-body">
                <div class="d-flex flex-wrap justify-content-between align-items-start gap-2 mb-3">
                  <div>
                    <div class="fw-semibold">{{ t('configs_page.import_validation_title') }}</div>
                    <div class="small text-muted mt-1">{{ state.importedConfigResult.validation.summary }}</div>
                  </div>
                  <span class="badge" :class="helpers.importValidationBadgeClass(state.importedConfigResult.validation.verdict)">
                    {{ helpers.importValidationLabel(state.importedConfigResult.validation.verdict) }}
                  </span>
                </div>
                <div class="row g-3">
                  <div class="col-md-6">
                    <div class="small text-muted mb-1">{{ t('configs_page.semantic_diff') }}</div>
                    <div class="fw-semibold">
                      {{ state.importedConfigResult.validation.semantic_diff?.summary || t('configs_page.no_diff') }}
                    </div>
                  </div>
                  <div class="col-md-6">
                    <div class="small text-muted mb-1">{{ t('configs_page.analysis') }}</div>
                    <div class="fw-semibold">
                      {{ state.importedConfigResult.validation.lint_summary || '-' }}
                    </div>
                  </div>
                </div>
                <div class="small text-muted mt-3">
                  {{ t('configs_page.import_semantic_change_count').replace('{count}', String(state.importSemanticChangeCount)) }}
                </div>
              </div>
            </div>

            <ConfigAssemblyFlow
              v-if="state.importedConfigResult.auto_assemble_supported"
              class="mb-3"
              :modules="state.importedConfigResult.modules"
              :destinations="state.importedConfigResult.destinations"
              :path-label="state.importFlowPathLabel"
            />

            <div class="row g-3">
              <div v-for="module in state.importedConfigResult.modules" :key="`${module.order}-${module.module_type}`" class="col-lg-6">
                <div class="fm-import-module-card h-100">
                  <div class="mb-2">
                    <label class="form-label small text-muted mb-1">{{ t('common.name') }}</label>
                    <input
                      v-model="module.name"
                      type="text"
                      class="form-control form-control-sm"
                      :class="{ 'is-invalid': !!helpers.importedModuleNameIssue(module) }"
                      :disabled="module.module_type === 'output' || module.import_action === 'reuse_existing'"
                      :placeholder="t('configs_page.import_module_name_placeholder')"
                    >
                    <div v-if="module.module_type === 'output'" class="small text-muted mt-1">
                      {{ t('configs_page.import_module_name_output_hint') }}
                    </div>
                    <div v-else-if="module.import_action === 'reuse_existing'" class="small text-muted mt-1">
                      {{ t('configs_page.import_module_name_reuse_hint') }}
                    </div>
                    <div v-else-if="helpers.importedModuleNameIssue(module)" class="invalid-feedback d-block">
                      {{ helpers.importedModuleNameIssue(module).message }}
                    </div>
                  </div>
                  <div class="d-flex align-items-center gap-2 flex-wrap mb-2">
                    <span class="badge text-bg-secondary">{{ module.module_type }}</span>
                    <span class="badge" :class="helpers.importActionBadgeClass(module.import_action)">
                      {{ helpers.importActionLabel(module.import_action) }}
                    </span>
                    <span v-if="module.detected_plugin" class="badge text-bg-light">{{ module.detected_plugin }}</span>
                  </div>
                  <div class="small text-muted mb-2">{{ module.summary || t('common.no_description') }}</div>
                  <div v-if="module.existing_module_name" class="small text-muted mb-2">
                    {{ t('configs_page.import_existing_match').replace('{name}', module.existing_module_name) }}
                  </div>
                  <div v-if="module.output_target_name" class="small text-muted mb-2">
                    {{ t('configs_page.import_destination_match').replace('{name}', module.output_target_name) }}
                    <span class="badge text-bg-light ms-2">{{ helpers.importDestinationMatchLabel(module.output_target_match_type) }}</span>
                  </div>
                  <div v-if="module.module_type === 'output'" class="small text-muted mb-2">
                    {{ t('configs_page.import_output_managed_by_destination') }}
                  </div>
                  <div v-else class="btn-group btn-group-sm mb-2" role="group">
                    <button
                      type="button"
                      class="btn"
                      :class="module.import_action === 'create_new' ? 'btn-primary' : 'btn-outline-primary'"
                      @click="actions.setImportedModuleAction(module, 'create_new')"
                    >
                      {{ t('configs_page.import_action_create') }}
                    </button>
                    <button
                      type="button"
                      class="btn"
                      :disabled="!module.existing_module_id"
                      :class="module.import_action === 'reuse_existing' ? 'btn-success' : 'btn-outline-success'"
                      @click="actions.setImportedModuleAction(module, 'reuse_existing')"
                    >
                      {{ t('configs_page.import_action_reuse') }}
                    </button>
                  </div>
                  <div v-if="module.variable_keys?.length" class="d-flex flex-wrap gap-2 mb-2">
                    <span
                      v-for="key in module.variable_keys"
                      :key="`${module.name}-${key}`"
                      class="badge rounded-pill text-bg-light"
                    >
                      {{ key }}
                    </span>
                  </div>
                  <pre class="fm-module-snippet">{{ module.content }}</pre>
                </div>
              </div>
            </div>

            <div v-if="state.importedConfigResult.auto_assemble_supported" class="card border-0 bg-light-subtle mt-3">
              <div class="card-body">
                <div class="fw-semibold mb-2">{{ t('configs_page.import_template_draft') }}</div>
                <div class="small text-muted mb-3">{{ t('configs_page.import_template_draft_hint') }}</div>
                <pre class="fm-render-preview">{{ state.importedConfigResult.template_draft_content }}</pre>
              </div>
            </div>

            <div v-if="state.importedWorkspaceModules.length" class="alert alert-success py-2 mt-3 mb-0">
              {{
                (state.importedConfigResult.auto_assemble_supported
                  ? t('configs_page.import_success')
                  : t('configs_page.import_success_assets')
                ).replace('{count}', String(state.importedWorkspaceModules.length))
              }}
            </div>
            <div v-if="state.importedWorkspaceTemplate" class="alert alert-success py-2 mt-3 mb-0">
              {{ t('configs_page.import_template_created').replace('{name}', state.importedWorkspaceTemplate.name) }}
            </div>
          </div>
          <div v-else class="text-center text-muted py-5">
            {{ t('configs_page.import_empty') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import ConfigAssemblyFlow from '../../../components/ConfigAssemblyFlow.vue'
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
