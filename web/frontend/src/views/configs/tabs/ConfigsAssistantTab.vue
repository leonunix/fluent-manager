<template>
  <div class="row g-4">
    <div class="col-xl-5">
      <!-- Log Discovery Panel — only shown to users with nodes:read permission -->
      <div v-if="canDiscoverLogs" class="card border-0 shadow-sm mb-4">
        <div
          class="card-header bg-white d-flex justify-content-between align-items-center"
          style="cursor:pointer"
          @click="discovery.expanded = !discovery.expanded"
        >
          <h6 class="mb-0">
            <i class="bi bi-search me-2 text-primary"></i>{{ t('configs_page.log_discovery_title') }}
          </h6>
          <i :class="discovery.expanded ? 'bi bi-chevron-up' : 'bi bi-chevron-down'" class="text-muted"></i>
        </div>
        <div v-show="discovery.expanded" class="card-body">
          <div class="alert alert-secondary py-2 small mb-3">
            {{ t('configs_page.log_discovery_intro') }}
          </div>

          <!-- Node + Paths row -->
          <div class="row g-2 mb-2">
            <div class="col-md-5">
              <label class="form-label small mb-1">{{ t('configs_page.log_discovery_node') }}</label>
              <input
                v-model="discovery.nodeFilter"
                type="text"
                class="form-control form-control-sm mb-1"
                :placeholder="t('configs_page.log_discovery_node_placeholder')"
              />
              <select v-model="discovery.nodeId" class="form-select form-select-sm" size="4">
                <option v-for="n in filteredNodes" :key="n.id" :value="n.id">{{ n.hostname || n.ip_address }}{{ n.hostname && n.ip_address ? ` (${n.ip_address})` : '' }}</option>
              </select>
            </div>
            <div class="col-md-5">
              <label class="form-label small mb-1">{{ t('configs_page.log_discovery_paths') }}</label>
              <input
                v-model="discovery.paths"
                type="text"
                class="form-control form-control-sm"
                :placeholder="t('configs_page.log_discovery_paths_placeholder')"
              />
            </div>
            <div class="col-md-2 d-flex align-items-end">
              <button
                class="btn btn-sm btn-outline-primary w-100"
                :disabled="!discovery.nodeId || discovery.scanning"
                @click="scanLogs"
              >
                <span v-if="discovery.scanning"><span class="spinner-border spinner-border-sm me-1"></span></span>
                <span v-else><i class="bi bi-radar me-1"></i></span>
                {{ discovery.scanning ? t('configs_page.log_discovery_scanning') : t('configs_page.log_discovery_scan') }}
              </button>
            </div>
          </div>

          <!-- Error -->
          <div v-if="discovery.error" class="alert alert-danger py-2 small mb-2">
            {{ t('configs_page.log_discovery_error') }}: {{ discovery.error }}
          </div>

          <!-- File list -->
          <div v-if="discovery.files.length > 0" class="mt-2">
            <div class="d-flex justify-content-between align-items-center mb-1">
              <small class="text-muted">{{ t('configs_page.log_discovery_select_hint') }}</small>
              <div class="d-flex gap-2">
                <button class="btn btn-link btn-sm p-0 text-muted" @click="selectAllFiles">{{ t('configs_page.log_discovery_select_all') }}</button>
                <button class="btn btn-link btn-sm p-0 text-muted" @click="discovery.selectedFiles = []">{{ t('configs_page.log_discovery_select_none') }}</button>
              </div>
            </div>
            <div class="border rounded p-2" style="max-height:200px;overflow-y:auto;background:#f8f9fa">
              <div
                v-for="f in discovery.files"
                :key="f.path"
                class="form-check form-check-sm mb-1"
              >
                <input
                  class="form-check-input"
                  type="checkbox"
                  :id="'df-' + f.path"
                  :value="f.path"
                  v-model="discovery.selectedFiles"
                />
                <label class="form-check-label small font-monospace" :for="'df-' + f.path" style="user-select:none">
                  {{ f.path }}
                  <span class="text-muted ms-1">({{ formatBytes(f.size) }})</span>
                </label>
              </div>
            </div>

            <div class="mt-2 d-flex align-items-center gap-2">
              <button
                class="btn btn-sm btn-primary"
                :disabled="discovery.selectedFiles.length === 0 || discovery.fetching"
                @click="fetchSamples"
              >
                <span v-if="discovery.fetching"><span class="spinner-border spinner-border-sm me-1"></span></span>
                <span v-else><i class="bi bi-download me-1"></i></span>
                {{ discovery.fetching ? t('configs_page.log_discovery_fetching') : t('configs_page.log_discovery_fetch') }}
                <span v-if="discovery.selectedFiles.length > 0" class="badge bg-white text-primary ms-1">{{ discovery.selectedFiles.length }}</span>
              </button>
              <small class="text-muted">{{ t('configs_page.log_discovery_append_hint') }}</small>
            </div>
          </div>

          <div v-else-if="discovery.files.length === 0 && !discovery.scanning && !discovery.error" class="text-muted small mt-2">
            {{ t('configs_page.log_discovery_no_files') }}
          </div>
        </div>
      </div>

      <!-- AI Assistant Form -->
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
              <label class="form-label">{{ t('configs_page.ai_module_type_hint') }}</label>
              <select v-model="state.aiAssistantForm.module_type" class="form-select">
                <option value="">{{ t('configs_page.ai_module_type_auto') }}</option>
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
          <div v-if="state.aiAssistantResult && !(state.aiAssistantResult.pipelines && state.aiAssistantResult.pipelines.length)" class="d-flex gap-2">
            <button class="btn btn-sm btn-outline-primary" @click="actions.useAITemplateDraft()">
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
              'is-warning': state.aiAssistantFeedback.type === 'warning',
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

            <!-- Generated modules list with merge decisions -->
            <div v-if="state.aiAssistantModules && state.aiAssistantModules.length" class="mb-3">
              <div class="d-flex justify-content-between align-items-center mb-2">
                <label class="form-label mb-0">{{ t('configs_page.ai_generated_modules') }}</label>
                <button
                  class="btn btn-sm btn-success"
                  :disabled="state.aiAssistantModulesSaving"
                  @click="actions.saveAIModules"
                >
                  <span v-if="state.aiAssistantModulesSaving"><span class="spinner-border spinner-border-sm me-1"></span></span>
                  <i v-else class="bi bi-cloud-arrow-up me-1"></i>
                  {{ state.aiAssistantModulesSaving ? t('configs_page.ai_modules_saving') : t('configs_page.ai_modules_save_all') }}
                </button>
              </div>
              <div class="d-flex flex-column gap-2">
                <div
                  v-for="(mod, idx) in state.aiAssistantModules"
                  :key="idx"
                  class="border rounded p-3"
                  :class="mod.decision === 'reuse_existing' ? 'border-success-subtle bg-success-subtle' : mod.decision === 'update_existing' ? 'border-warning-subtle bg-warning-subtle' : 'border-primary-subtle bg-primary-subtle'"
                >
                  <div class="d-flex flex-wrap justify-content-between align-items-start gap-2 mb-2">
                    <div class="d-flex align-items-center gap-2">
                      <span class="badge text-bg-secondary font-monospace">{{ mod.module_type }}</span>
                      <span class="fw-semibold">{{ mod.name }}</span>
                    </div>
                    <div class="d-flex align-items-center gap-2">
                      <select
                        v-model="mod.decision"
                        class="form-select form-select-sm"
                        style="width:auto"
                      >
                        <option value="create_new">{{ t('configs_page.ai_module_create_new') }}</option>
                        <option value="reuse_existing">{{ t('configs_page.ai_module_reuse') }}</option>
                        <option value="update_existing">{{ t('configs_page.ai_module_update') }}</option>
                      </select>
                      <button
                        v-if="mod.decision !== 'reuse_existing'"
                        class="btn btn-sm btn-outline-primary"
                        @click="actions.useAIModuleDraft(mod)"
                        :title="t('configs_page.ai_use_module')"
                      >
                        <i class="bi bi-pencil-square"></i>
                      </button>
                    </div>
                  </div>
                  <div v-if="mod.matchedModule" class="small text-muted mb-2">
                    {{ t('configs_page.ai_module_matched') }}: <span class="fw-semibold">{{ mod.matchedModule.name }}</span>
                  </div>
                  <div v-if="mod.note" class="small text-muted mb-2">{{ mod.note }}</div>
                  <textarea
                    v-if="mod.decision !== 'reuse_existing'"
                    class="form-control form-control-sm font-monospace"
                    rows="6"
                    readonly
                    :value="mod.content"
                  ></textarea>
                </div>
              </div>
            </div>

            <!-- Generated pipelines list -->
            <div v-if="state.aiAssistantResult.pipelines && state.aiAssistantResult.pipelines.length" class="mb-3">
              <label class="form-label">{{ t('configs_page.ai_generated_pipelines') }}</label>
              <div class="d-flex flex-column gap-2">
                <div
                  v-for="(pipeline, idx) in state.aiAssistantResult.pipelines"
                  :key="idx"
                  class="border rounded p-3 border-info-subtle bg-info-subtle"
                >
                  <div class="d-flex flex-wrap justify-content-between align-items-start gap-2 mb-2">
                    <div>
                      <span class="fw-semibold">{{ pipeline.name }}</span>
                      <div v-if="pipeline.description" class="small text-muted mt-1">{{ pipeline.description }}</div>
                    </div>
                    <div class="d-flex gap-1 flex-wrap">
                      <button
                        class="btn btn-sm btn-outline-primary"
                        @click="actions.useAITemplateDraft(pipeline)"
                      >
                        <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.ai_pipeline_save_as_template') }}
                      </button>
                      <button
                        class="btn btn-sm btn-outline-success"
                        @click="actions.saveAIPipelineAsConfigPipeline(pipeline)"
                      >
                        <i class="bi bi-diagram-3 me-1"></i>{{ t('configs_page.ai_pipeline_save_as_config_pipeline') }}
                      </button>
                      <button
                        class="btn btn-sm btn-outline-secondary"
                        @click="actions.sendAIPipelineToWizard(pipeline)"
                      >
                        <i class="bi bi-magic me-1"></i>{{ t('configs_page.ai_pipeline_send_to_wizard') }}
                      </button>
                    </div>
                  </div>
                  <div v-if="pipeline.module_names && pipeline.module_names.length" class="mb-2">
                    <span class="small text-muted me-1">{{ t('configs_page.ai_pipeline_modules') }}:</span>
                    <span
                      v-for="modName in pipeline.module_names"
                      :key="modName"
                      class="badge bg-secondary-subtle text-secondary-emphasis font-monospace me-1"
                    >{{ modName }}</span>
                  </div>
                  <div v-if="pipeline.note" class="small text-muted mb-2">{{ pipeline.note }}</div>
                  <textarea
                    class="form-control form-control-sm font-monospace"
                    rows="8"
                    readonly
                    :value="pipeline.template_content"
                  ></textarea>
                </div>
              </div>
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
import { reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from '../../../i18n'
import { useAuthStore } from '../../../store/auth'
import { getNodes, sendNodeCommand, getNodeCommand } from '../../../api/nodes'

const props = defineProps({
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
const auth = useAuthStore()

const canDiscoverLogs = computed(() => auth.hasPermission('nodes', 'read') && auth.hasPermission('nodes', 'update'))

let pollAborted = false

onBeforeUnmount(() => { pollAborted = true })

const discovery = reactive({
  expanded: false,
  nodes: [],
  nodeFilter: '',
  nodeId: '',
  paths: '/var/log',
  scanning: false,
  files: [],
  selectedFiles: [],
  fetching: false,
  error: '',
})

const filteredNodes = computed(() => {
  const q = discovery.nodeFilter.trim().toLowerCase()
  if (!q) return discovery.nodes
  return discovery.nodes.filter(n =>
    (n.hostname || '').toLowerCase().includes(q) ||
    (n.ip_address || '').toLowerCase().includes(q)
  )
})

onMounted(async () => {
  if (!canDiscoverLogs.value || discovery.nodes.length > 0) return
  try {
    const { data } = await getNodes({ page_size: 1000 })
    discovery.nodes = data.data || data || []
  } catch {
    // non-blocking
  }
})

function formatBytes(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function selectAllFiles() {
  discovery.selectedFiles = discovery.files.map(f => f.path)
}

async function pollCommand(nodeId, cmdId, maxWaitMs = 90000) {
  const start = Date.now()
  while (Date.now() - start < maxWaitMs) {
    await new Promise(r => setTimeout(r, 2000))
    if (pollAborted) throw new Error('Cancelled')
    const { data } = await getNodeCommand(nodeId, cmdId)
    if (data.status === 'success' || data.status === 'failed' || data.status === 'completed') {
      return data
    }
  }
  throw new Error('Timed out waiting for command result')
}

async function scanLogs() {
  if (!discovery.nodeId) return
  discovery.scanning = true
  discovery.files = []
  discovery.selectedFiles = []
  discovery.error = ''

  try {
    const paths = discovery.paths
      .split(',')
      .map(p => p.trim())
      .filter(Boolean)

    const { data: cmd } = await sendNodeCommand(Number(discovery.nodeId), {
      action: 'scan_logs',
      args: JSON.stringify({ paths: paths.length ? paths : ['/var/log'] }),
    })

    const result = await pollCommand(Number(discovery.nodeId), cmd.id)
    if (result.status === 'failed') {
      discovery.error = result.output || 'scan failed'
      return
    }

    const parsed = JSON.parse(result.output || '{}')
    discovery.files = parsed.files || []
  } catch (e) {
    discovery.error = e.message || String(e)
  } finally {
    discovery.scanning = false
  }
}

async function fetchSamples() {
  if (!discovery.selectedFiles.length) return
  // Mirror the agent-side cap so the user gets immediate feedback in the UI
  const files = discovery.selectedFiles.slice(0, 10)
  discovery.fetching = true
  discovery.error = ''

  try {
    const { data: cmd } = await sendNodeCommand(Number(discovery.nodeId), {
      action: 'fetch_log_sample',
      args: JSON.stringify({ files, lines: 100 }),
    })

    const result = await pollCommand(Number(discovery.nodeId), cmd.id)
    if (result.status === 'failed') {
      discovery.error = result.output || 'fetch failed'
      return
    }

    const parsed = JSON.parse(result.output || '{}')
    const samples = parsed.samples || {}
    const parts = Object.keys(samples).sort().map(path => `# ${path}\n${samples[path]}`)
    if (parts.length) {
      const fetched = parts.join('\n\n')
      const existing = (props.state.aiAssistantForm.sample || '').trim()
      props.state.aiAssistantForm.sample = existing ? existing + '\n\n' + fetched : fetched
    }
  } catch (e) {
    discovery.error = e.message || String(e)
  } finally {
    discovery.fetching = false
  }
}
</script>
