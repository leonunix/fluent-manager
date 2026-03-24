<template>
  <div class="row g-4">
    <div class="col-xl-4">
      <div class="card border-0 shadow-sm h-100">
        <div class="card-header bg-white">
          <h6 class="mb-0">{{ t('configs_page.render_params') }}</h6>
        </div>
        <div class="card-body">
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.preview_name') }}</label>
            <input v-model="state.previewForm.name" type="text" class="form-control" placeholder="preview-fluentbit-edge">
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label">{{ t('common.runtime') }}</label>
              <select v-model="state.previewForm.fluent_type" class="form-select">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('configs_page.target_version') }}</label>
              <input v-model="state.previewForm.runtime_version" type="text" class="form-control" placeholder="3.1.0 / 1.16">
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.variables_json') }}</label>
            <textarea
              v-model="state.previewForm.variables"
              class="form-control font-monospace"
              rows="8"
              placeholder='{"path":"/var/log/*.log","match":"*"}'
            ></textarea>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.output_destination') }}</label>
            <div class="d-grid gap-2">
              <button
                v-for="target in state.previewAvailableOutputTargets"
                :key="target.id"
                type="button"
                class="btn text-start"
                :class="state.selectedPreviewOutputTargetIds.includes(target.id) ? 'btn-primary' : 'btn-outline-secondary'"
                @click="actions.togglePreviewOutputTarget(target.id)"
              >
                <div class="fw-semibold">{{ target.name }}</div>
                <div class="small opacity-75">{{ target.target_type }} · {{ target.endpoint || t('common.unspecified') }}</div>
              </button>
            </div>
            <div v-if="state.previewUnresolvedOutputTargets.length" class="alert alert-warning py-2 mt-2 mb-0">
              {{ t('configs_page.output_target_module_missing').replace('{targets}', state.previewUnresolvedOutputTargets.map((item) => item.name).join(', ')) }}
            </div>
            <div class="small text-muted mt-2">{{ t('configs_page.preview_output_destination_hint') }}</div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.compatibility_node') }}</label>
            <input v-model="state.previewForm.node_id" type="number" min="1" class="form-control" :placeholder="t('configs_page.compatibility_node')">
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.sample_tag') }}</label>
            <input v-model="state.previewForm.sample_tag" type="text" class="form-control" placeholder="app.logs">
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.sample_log') }}</label>
            <textarea
              v-model="state.previewForm.sample_log"
              class="form-control font-monospace"
              rows="6"
              placeholder='{"message":"hello fluent","level":"info"}'
            ></textarea>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('configs_page.diff_content') }}</label>
            <textarea
              v-model="state.previewForm.diff_content"
              class="form-control font-monospace"
              rows="6"
              :placeholder="t('configs_page.diff_content')"
            ></textarea>
          </div>
          <button class="btn btn-success w-100" @click="actions.runPreview">
            <i class="bi bi-play-circle me-1"></i>{{ t('configs_page.generate_preview') }}
          </button>
          <button class="btn btn-outline-primary w-100 mt-2" @click="actions.runLint">
            <i class="bi bi-shield-check me-1"></i>{{ t('configs_page.run_lint') }}
          </button>
          <button class="btn btn-outline-secondary w-100 mt-2" @click="actions.runCompatibility">
            <i class="bi bi-patch-check me-1"></i>{{ t('configs_page.run_compatibility') }}
          </button>
          <button class="btn btn-outline-dark w-100 mt-2" @click="actions.runReplay">
            <i class="bi bi-magic me-1"></i>{{ t('configs_page.run_replay') }}
          </button>
          <button class="btn btn-outline-info w-100 mt-2" @click="actions.runSemanticDiff">
            <i class="bi bi-intersect me-1"></i>{{ t('configs_page.run_diff') }}
          </button>
        </div>
      </div>
    </div>

    <div class="col-xl-8">
      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <h6 class="mb-0">{{ t('configs_page.available_modules') }}</h6>
          <span class="text-muted small">{{ t('configs_page.runtime_help').replace('{runtime}', helpers.runtimeLabel(state.previewForm.fluent_type)) }}</span>
        </div>
        <div class="card-body">
          <div class="row g-3 align-items-end mb-4">
            <div class="col-lg-8">
              <label class="form-label">{{ t('common.search') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-search"></i></span>
                <input
                  v-model="state.previewModuleSearch"
                  type="text"
                  class="form-control"
                  :placeholder="t('configs_page.module_picker_search_placeholder')"
                >
                <button
                  v-if="state.previewModuleSearch"
                  type="button"
                  class="btn btn-outline-secondary"
                  @click="state.previewModuleSearch = ''"
                >
                  {{ t('cancel') }}
                </button>
              </div>
            </div>
            <div class="col-lg-4">
              <div class="small text-muted">
                {{ t('configs_page.module_picker_search_hint').replace('{count}', String(state.previewVisibleModules.length)) }}
              </div>
            </div>
          </div>
          <div class="row g-3">
            <div
              v-for="module in state.previewVisibleModules"
              :key="module.id"
              class="col-lg-6"
            >
              <label class="fm-module-choice h-100" :class="{ selected: state.selectedPreviewModuleIds.includes(module.id) }">
                <input
                  :checked="state.selectedPreviewModuleIds.includes(module.id)"
                  type="checkbox"
                  class="form-check-input"
                  @change="actions.togglePreviewModule(module.id)"
                >
                <div class="flex-grow-1">
                  <div class="d-flex align-items-center gap-2 mb-2">
                    <span class="fw-semibold">{{ module.name }}</span>
                    <span class="badge text-bg-secondary">{{ module.module_type }}</span>
                    <span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(module.fluent_type) }}</span>
                  </div>
                  <div class="small text-muted mb-2">{{ module.description || t('common.no_description') }}</div>
                  <pre class="fm-module-snippet">{{ module.content }}</pre>
                </div>
              </label>
            </div>
            <div v-if="!state.previewVisibleModules.length" class="col-12 text-center text-muted py-4">
              {{ state.previewModuleSearch ? t('configs_page.module_picker_no_results') : t('configs_page.current_runtime_no_modules') }}
            </div>
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.render_result') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.render_result') }}</div>
          </div>
          <div v-if="state.renderedConfig" class="text-end">
            <div class="small text-muted">{{ t('configs_page.render_hash') }}</div>
            <code>{{ state.renderedConfig.hash }}</code>
          </div>
        </div>
        <div class="card-body">
          <div v-if="state.renderedConfig">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge bg-success-subtle text-success-emphasis">ID {{ state.renderedConfig.id }}</span>
              <span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(state.renderedConfig.fluent_type) }}</span>
              <span class="badge text-bg-light">{{ state.renderedConfig.runtime_version || t('common.unspecified') }}</span>
            </div>
            <div class="card border-0 bg-light-subtle mb-3">
              <div class="card-body">
                <div class="fw-semibold mb-2">{{ t('configs_page.render_chain_summary') }}</div>
                <div class="small text-muted mb-3">{{ t('configs_page.render_chain_hint') }}</div>
                <div class="mb-2">
                  <div class="small text-muted mb-1">{{ t('configs_page.solution_path') }}</div>
                  <div class="fw-semibold">{{ state.previewFlowPathLabel }}</div>
                </div>
                <div class="d-flex flex-wrap gap-2">
                  <span
                    v-for="chip in state.previewDestinationChips"
                    :key="`preview-${chip}`"
                    class="badge rounded-pill text-bg-light"
                  >
                    {{ chip }}
                  </span>
                  <span v-if="!state.previewDestinationChips.length" class="text-muted small">{{ t('configs_page.no_destination_summary') }}</span>
                </div>
              </div>
            </div>
            <ConfigAssemblyFlow
              class="mb-3"
              :modules="state.previewSummaryModules"
              :destinations="state.previewResolvedOutputTargets"
              :path-label="state.previewFlowPathLabel"
            />
            <pre class="fm-render-preview">{{ state.renderedConfig.content }}</pre>
          </div>
          <div v-else class="text-center text-muted py-5">
            {{ t('configs_page.no_preview') }}
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm mt-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.analysis') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.analysis') }}</div>
          </div>
          <span v-if="state.analysisResult" class="badge text-bg-light">{{ state.analysisResult.summary }}</span>
        </div>
        <div class="card-body">
          <div v-if="state.analysisResult">
            <div class="list-group list-group-flush">
              <div
                v-for="finding in state.analysisResult.findings || []"
                :key="`${finding.rule_code}-${finding.line}`"
                class="list-group-item px-0"
              >
                <div class="d-flex align-items-center gap-2 mb-1">
                  <span class="badge" :class="helpers.findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                  <code>{{ finding.rule_code }}</code>
                  <span class="small text-muted">Line {{ finding.line }}</span>
                </div>
                <div class="fw-semibold">{{ finding.message }}</div>
                <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-muted py-4">
            {{ t('configs_page.no_analysis') }}
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm mt-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.compatibility') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.compatibility') }}</div>
          </div>
          <span
            v-if="state.compatibilityResult"
            class="badge"
            :class="state.compatibilityResult.compatible ? 'text-bg-success' : 'text-bg-danger'"
          >
            {{ state.compatibilityResult.compatible ? 'compatible' : 'needs attention' }}
          </span>
        </div>
        <div class="card-body">
          <div v-if="state.compatibilityResult">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge text-bg-light">{{ state.compatibilityResult.hot_reload_supported ? t('configs_page.hot_reload_available') : t('configs_page.hot_reload_unavailable') }}</span>
              <span class="badge text-bg-light">{{ t('configs_page.missing_plugins') }} {{ state.compatibilityResult.missing_plugins?.length || 0 }}</span>
            </div>
            <div v-if="state.compatibilityResult.missing_plugins?.length" class="alert alert-warning py-2">
              {{ t('configs_page.missing_plugins') }}: {{ state.compatibilityResult.missing_plugins.join(', ') }}
            </div>
            <div class="list-group list-group-flush">
              <div
                v-for="finding in state.compatibilityResult.findings || []"
                :key="`compat-${finding.rule_code}-${finding.line}`"
                class="list-group-item px-0"
              >
                <div class="d-flex align-items-center gap-2 mb-1">
                  <span class="badge" :class="helpers.findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                  <code>{{ finding.rule_code }}</code>
                </div>
                <div class="fw-semibold">{{ finding.message }}</div>
                <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-muted py-4">
            {{ t('configs_page.no_compatibility') }}
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm mt-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.replay') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.replay') }}</div>
          </div>
          <span v-if="state.replayResult" class="badge" :class="state.replayResult.route_matched ? 'text-bg-success' : 'text-bg-warning'">
            {{ state.replayResult.route_matched ? t('configs_page.route_matched') : t('configs_page.no_route') }}
          </span>
        </div>
        <div class="card-body">
          <div v-if="state.replayResult">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge text-bg-light">Parser {{ state.replayResult.detected_parser || '-' }}</span>
              <span class="badge text-bg-light">{{ t('common.output') }} {{ state.replayResult.final_output || '-' }}</span>
            </div>
            <div v-if="state.replayResult.warnings?.length" class="alert alert-warning py-2">
              {{ state.replayResult.warnings.join('；') }}
            </div>
            <div class="row g-4">
              <div class="col-lg-6">
                <h6 class="small text-muted text-uppercase">{{ t('configs_page.steps') }}</h6>
                <div class="list-group list-group-flush">
                  <div v-for="step in state.replayResult.steps || []" :key="`${step.stage}-${step.name}`" class="list-group-item px-0">
                    <div class="d-flex align-items-center gap-2 mb-1">
                      <span class="badge text-bg-light">{{ step.stage }}</span>
                      <span class="fw-semibold">{{ step.name }}</span>
                    </div>
                    <div class="small text-muted">{{ step.detail }}</div>
                  </div>
                </div>
              </div>
              <div class="col-lg-6">
                <h6 class="small text-muted text-uppercase">{{ t('configs_page.parsed_record') }}</h6>
                <pre class="fm-render-preview mb-0">{{ helpers.formatJson(state.replayResult.parsed_record) }}</pre>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-muted py-4">
            {{ t('configs_page.no_replay') }}
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm mt-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.semantic_diff') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.semantic_diff') }}</div>
          </div>
          <span v-if="state.diffResult" class="badge text-bg-light">{{ state.diffResult.summary }}</span>
        </div>
        <div class="card-body">
          <div v-if="state.diffResult">
            <div class="list-group list-group-flush">
              <div v-for="change in state.diffResult.changes || []" :key="`${change.category}-${change.change_type}-${change.item}`" class="list-group-item px-0">
                <div class="d-flex align-items-center gap-2 mb-1">
                  <span class="badge text-bg-light">{{ change.category }}</span>
                  <span class="badge" :class="change.change_type === 'added' ? 'text-bg-success' : 'text-bg-danger'">{{ change.change_type }}</span>
                  <code>{{ change.item }}</code>
                </div>
                <div class="small text-muted">{{ change.detail }}</div>
              </div>
            </div>
            <div v-if="!state.diffResult.changes?.length" class="text-center text-muted py-3">
              {{ t('configs_page.no_changes') }}
            </div>
          </div>
          <div v-else class="text-center text-muted py-4">
            {{ t('configs_page.no_diff') }}
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
