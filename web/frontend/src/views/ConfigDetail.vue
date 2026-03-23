<template>
  <div v-if="template">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <router-link to="/configs" class="text-decoration-none">&larr; {{ t('config_detail.back') }}</router-link>
        <h4 class="mt-2 mb-0">{{ template.name }}</h4>
        <span class="badge bg-info">{{ template.fluent_type }}</span>
        <span class="text-muted ms-2">{{ template.description }}</span>
      </div>
      <button class="btn btn-primary" @click="openNewVersion">
        <i class="bi bi-plus-lg me-1"></i>{{ t('config_detail.create_version') }}
      </button>
    </div>

    <div class="row g-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">{{ t('config_detail.version_list') }}</h6></div>
          <div class="list-group list-group-flush">
            <a
              v-for="v in versions"
              :key="v.id"
              href="#"
              class="list-group-item list-group-item-action"
              :class="{ active: selectedVersion?.id === v.id }"
              @click.prevent="selectedVersion = v"
            >
              <div class="d-flex justify-content-between">
                <strong>v{{ v.version }}</strong>
                <small>{{ formatTime(v.created_at) }}</small>
              </div>
              <small class="text-muted">{{ v.comment || t('configs_page.no_version_comment') }}</small>
            </a>
            <div v-if="!versions.length" class="list-group-item text-muted text-center">{{ t('config_detail.no_versions') }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-8">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">
              {{ selectedVersion ? t('config_detail.version_label').replace('{version}', selectedVersion.version) : t('config_detail.template_content') }}
            </h6>
            <button v-if="selectedVersion" class="btn btn-sm btn-success" @click="openDeploy">
              <i class="bi bi-rocket me-1"></i>{{ t('config_detail.deploy_version') }}
            </button>
          </div>
          <div class="card-body">
            <pre class="bg-dark text-light p-3 rounded fm-config-content">{{ currentContent }}</pre>
            <div v-if="selectedVersion" class="mt-2 text-muted small">
              SHA-256: {{ selectedVersion.hash }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mt-4">
      <div class="card-header bg-white d-flex justify-content-between align-items-center">
        <div>
          <h6 class="mb-0">{{ t('config_detail.analysis_workspace') }}</h6>
          <div class="small text-muted mt-1">{{ currentTargetLabel }}</div>
        </div>
        <span class="badge text-bg-light">{{ template.fluent_type }}</span>
      </div>
      <div class="card-body">
        <div class="row g-4">
          <div class="col-xl-4">
            <div class="border rounded-3 p-3 h-100">
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.target_version') }}</label>
                <input v-model="analysisForm.runtime_version" type="text" class="form-control" placeholder="3.1.0 / 1.16">
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.compatibility_node') }}</label>
                <input v-model="analysisForm.node_id" type="number" min="1" class="form-control" :placeholder="t('configs_page.compatibility_node')">
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.sample_tag') }}</label>
                <input v-model="analysisForm.sample_tag" type="text" class="form-control" placeholder="app.logs">
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.sample_log') }}</label>
                <textarea
                  v-model="analysisForm.sample_log"
                  class="form-control font-monospace"
                  rows="6"
                  placeholder='{"message":"hello fluent","level":"info"}'
                ></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('config_detail.compare_version') }}</label>
                <select v-model="analysisForm.compare_version_id" class="form-select" :disabled="!compareVersionOptions.length">
                  <option :value="null">{{ compareVersionOptions.length ? t('config_detail.select_compare_version') : t('config_detail.no_compare_versions') }}</option>
                  <option v-for="version in compareVersionOptions" :key="version.id" :value="version.id">
                    v{{ version.version }} - {{ formatTime(version.created_at) }}
                  </option>
                </select>
              </div>
              <div class="d-grid gap-2">
                <button class="btn btn-outline-primary" @click="runLint">
                  <i class="bi bi-shield-check me-1"></i>{{ t('configs_page.run_lint') }}
                </button>
                <button class="btn btn-outline-secondary" @click="runCompatibility">
                  <i class="bi bi-patch-check me-1"></i>{{ t('configs_page.run_compatibility') }}
                </button>
                <button class="btn btn-outline-dark" @click="runReplay">
                  <i class="bi bi-magic me-1"></i>{{ t('configs_page.run_replay') }}
                </button>
                <button class="btn btn-outline-info" :disabled="!compareVersionOptions.length" @click="runSemanticDiff">
                  <i class="bi bi-intersect me-1"></i>{{ t('configs_page.run_diff') }}
                </button>
              </div>
            </div>
          </div>

          <div class="col-xl-8">
            <div class="nav nav-pills fm-analysis-tabs mb-3">
              <button class="nav-link" :class="{ active: analysisTab === 'analysis' }" @click="analysisTab = 'analysis'">
                {{ t('configs_page.analysis') }}
              </button>
              <button class="nav-link" :class="{ active: analysisTab === 'compatibility' }" @click="analysisTab = 'compatibility'">
                {{ t('configs_page.compatibility') }}
              </button>
              <button class="nav-link" :class="{ active: analysisTab === 'replay' }" @click="analysisTab = 'replay'">
                {{ t('configs_page.replay') }}
              </button>
              <button class="nav-link" :class="{ active: analysisTab === 'diff' }" @click="analysisTab = 'diff'">
                {{ t('configs_page.semantic_diff') }}
              </button>
            </div>

            <div v-if="analysisTab === 'analysis'" class="card border-0 bg-light-subtle">
              <div class="card-body">
                <div v-if="analysisResult">
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <h6 class="mb-0">{{ t('configs_page.analysis') }}</h6>
                    <span class="badge text-bg-light">{{ analysisResult.summary }}</span>
                  </div>
                  <div class="list-group list-group-flush">
                    <div
                      v-for="finding in analysisResult.findings || []"
                      :key="`${finding.rule_code}-${finding.line}`"
                      class="list-group-item px-0 bg-transparent"
                    >
                      <div class="d-flex align-items-center gap-2 mb-1">
                        <span class="badge" :class="findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                        <code>{{ finding.rule_code }}</code>
                        <span class="small text-muted">Line {{ finding.line }}</span>
                      </div>
                      <div class="fw-semibold">{{ finding.message }}</div>
                      <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
                    </div>
                  </div>
                </div>
                <div v-else class="text-center text-muted py-5">
                  {{ t('configs_page.no_analysis') }}
                </div>
              </div>
            </div>

            <div v-else-if="analysisTab === 'compatibility'" class="card border-0 bg-light-subtle">
              <div class="card-body">
                <div v-if="compatibilityResult">
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <h6 class="mb-0">{{ t('configs_page.compatibility') }}</h6>
                    <span class="badge" :class="compatibilityResult.compatible ? 'text-bg-success' : 'text-bg-danger'">
                      {{ compatibilityResult.compatible ? 'compatible' : 'needs attention' }}
                    </span>
                  </div>
                  <div class="d-flex flex-wrap gap-2 mb-3">
                    <span class="badge text-bg-light">
                      {{ compatibilityResult.hot_reload_supported ? t('configs_page.hot_reload_available') : t('configs_page.hot_reload_unavailable') }}
                    </span>
                    <span class="badge text-bg-light">
                      {{ t('configs_page.missing_plugins') }} {{ compatibilityResult.missing_plugins?.length || 0 }}
                    </span>
                  </div>
                  <div v-if="compatibilityResult.missing_plugins?.length" class="alert alert-warning py-2">
                    {{ t('configs_page.missing_plugins') }}: {{ compatibilityResult.missing_plugins.join(', ') }}
                  </div>
                  <div class="list-group list-group-flush">
                    <div
                      v-for="finding in compatibilityResult.findings || []"
                      :key="`compat-${finding.rule_code}-${finding.line}`"
                      class="list-group-item px-0 bg-transparent"
                    >
                      <div class="d-flex align-items-center gap-2 mb-1">
                        <span class="badge" :class="findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                        <code>{{ finding.rule_code }}</code>
                      </div>
                      <div class="fw-semibold">{{ finding.message }}</div>
                      <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
                    </div>
                  </div>
                </div>
                <div v-else class="text-center text-muted py-5">
                  {{ t('configs_page.no_compatibility') }}
                </div>
              </div>
            </div>

            <div v-else-if="analysisTab === 'replay'" class="card border-0 bg-light-subtle">
              <div class="card-body">
                <div v-if="replayResult">
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <h6 class="mb-0">{{ t('configs_page.replay') }}</h6>
                    <span class="badge" :class="replayResult.route_matched ? 'text-bg-success' : 'text-bg-warning'">
                      {{ replayResult.route_matched ? t('configs_page.route_matched') : t('configs_page.no_route') }}
                    </span>
                  </div>
                  <div class="d-flex flex-wrap gap-2 mb-3">
                    <span class="badge text-bg-light">Parser {{ replayResult.detected_parser || '-' }}</span>
                    <span class="badge text-bg-light">{{ t('common.output') }} {{ replayResult.final_output || '-' }}</span>
                    <span class="badge text-bg-light">{{ t('common.type') }} {{ replayResult.final_output_type || '-' }}</span>
                  </div>
                  <div v-if="replayResult.warnings?.length" class="alert alert-warning py-2">
                    {{ replayResult.warnings.join('; ') }}
                  </div>
                  <div class="row g-4">
                    <div class="col-lg-6">
                      <h6 class="small text-muted text-uppercase">{{ t('configs_page.steps') }}</h6>
                      <div class="list-group list-group-flush">
                        <div v-for="step in replayResult.steps || []" :key="`${step.stage}-${step.name}`" class="list-group-item px-0 bg-transparent">
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
                      <pre class="fm-config-content mb-0">{{ formatJson(replayResult.parsed_record) }}</pre>
                    </div>
                  </div>
                </div>
                <div v-else class="text-center text-muted py-5">
                  {{ t('configs_page.no_replay') }}
                </div>
              </div>
            </div>

            <div v-else class="card border-0 bg-light-subtle">
              <div class="card-body">
                <div v-if="diffResult">
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <h6 class="mb-0">{{ t('configs_page.semantic_diff') }}</h6>
                    <span class="badge text-bg-light">{{ diffResult.summary }}</span>
                  </div>
                  <div class="list-group list-group-flush">
                    <div
                      v-for="change in diffResult.changes || []"
                      :key="`${change.category}-${change.change_type}-${change.item}`"
                      class="list-group-item px-0 bg-transparent"
                    >
                      <div class="d-flex align-items-center gap-2 mb-1">
                        <span class="badge text-bg-light">{{ change.category }}</span>
                        <span class="badge" :class="change.change_type === 'added' ? 'text-bg-success' : 'text-bg-danger'">{{ change.change_type }}</span>
                        <code>{{ change.item }}</code>
                      </div>
                      <div class="small text-muted">{{ change.detail }}</div>
                    </div>
                  </div>
                  <div v-if="!diffResult.changes?.length" class="text-center text-muted py-3">
                    {{ t('configs_page.no_changes') }}
                  </div>
                </div>
                <div v-else class="text-center text-muted py-5">
                  {{ t('configs_page.no_diff') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="versionModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('config_detail.new_version_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.version_comment') }}</label>
              <input v-model="versionForm.comment" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.template_content') }}</label>
              <textarea v-model="versionForm.content" class="form-control font-monospace" rows="15"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveVersion">{{ t('config_detail.create_version') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="deployModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('config_detail.deploy_title').replace('{version}', selectedVersion?.version || '') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('config_detail.deploy_scope') }}</label>
              <select v-model="deployForm.scope" class="form-select">
                <option value="cluster">{{ t('config_detail.by_cluster') }}</option>
                <option value="region">{{ t('config_detail.by_region') }}</option>
                <option value="datacenter">{{ t('config_detail.by_datacenter') }}</option>
                <option value="node">{{ t('config_detail.by_node') }}</option>
              </select>
            </div>
            <div v-if="deployForm.scope === 'cluster'" class="mb-3">
              <label class="form-label">{{ t('config_detail.select_cluster') }}</label>
              <select v-model="deployForm.cluster_id" class="form-select">
                <option v-for="cl in clusters" :key="cl.id" :value="cl.id">{{ cl.alias || cl.name }}</option>
              </select>
            </div>
            <div v-if="deployForm.scope === 'region'" class="mb-3">
              <label class="form-label">{{ t('config_detail.select_region') }}</label>
              <select v-model="deployForm.region_id" class="form-select">
                <option v-for="r in regions" :key="r.id" :value="r.id">{{ r.alias || r.name }}</option>
              </select>
            </div>
            <div v-if="deployForm.scope === 'datacenter'" class="mb-3">
              <label class="form-label">{{ t('config_detail.select_datacenter') }}</label>
              <select v-model="deployForm.datacenter_id" class="form-select">
                <option v-for="dc in datacenters" :key="dc.id" :value="dc.id">{{ dc.alias || dc.name }}</option>
              </select>
            </div>
            <div v-if="deployForm.scope === 'node'" class="mb-3">
              <label class="form-label">{{ t('config_detail.node_ids') }}</label>
              <input v-model="deployForm.node_ids_text" type="text" class="form-control" placeholder="1,2,3">
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-success" @click="submitDeploy">{{ t('config_detail.confirm_deploy') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  checkCompatibility,
  createDeploy,
  createVersion,
  diffConfig,
  getClusters,
  getDataCenters,
  getRegions,
  getTemplate,
  getVersions,
  lintConfig,
  replayConfig,
} from '../api'
import { useI18n } from '../i18n'

const route = useRoute()
const router = useRouter()
const template = ref(null)
const versions = ref([])
const selectedVersion = ref(null)
const clusters = ref([])
const regions = ref([])
const datacenters = ref([])

const analysisTab = ref('analysis')
const analysisResult = ref(null)
const compatibilityResult = ref(null)
const replayResult = ref(null)
const diffResult = ref(null)

const versionForm = reactive({ content: '', comment: '' })
const deployForm = reactive({ scope: 'cluster', cluster_id: null, region_id: null, datacenter_id: null, node_ids_text: '' })
const analysisForm = reactive({
  runtime_version: '',
  node_id: '',
  sample_tag: 'app.logs',
  sample_log: '{"message":"hello fluent","level":"info"}',
  compare_version_id: null,
})

let versionModal = null
let deployModal = null
const { t, dateLocale } = useI18n()

const currentContent = computed(() => selectedVersion.value?.content || template.value?.content || '')
const currentTargetLabel = computed(() => (
  selectedVersion.value
    ? t('config_detail.version_label').replace('{version}', selectedVersion.value.version)
    : t('config_detail.template_content')
))
const compareVersionOptions = computed(() => versions.value.filter((version) => version.id !== selectedVersion.value?.id))
const compareVersion = computed(() => compareVersionOptions.value.find((version) => version.id === analysisForm.compare_version_id) || null)

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function formatJson(value) {
  try {
    return JSON.stringify(value || {}, null, 2)
  } catch {
    return String(value || '{}')
  }
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function findingBadgeClass(severity) {
  if (severity === 'error') return 'text-bg-danger'
  if (severity === 'warning') return 'text-bg-warning'
  return 'text-bg-info'
}

function resetAnalysisResults() {
  analysisResult.value = null
  compatibilityResult.value = null
  replayResult.value = null
  diffResult.value = null
}

function syncCompareVersion() {
  const available = compareVersionOptions.value
  if (!available.length) {
    analysisForm.compare_version_id = null
    return
  }
  if (!available.some((version) => version.id === analysisForm.compare_version_id)) {
    analysisForm.compare_version_id = available[0].id
  }
}

async function loadData() {
  const id = route.params.id
  const [tplRes, versRes, clRes, rRes, dcRes] = await Promise.all([
    getTemplate(id),
    getVersions(id),
    getClusters(),
    getRegions(),
    getDataCenters(),
  ])
  template.value = tplRes.data
  versions.value = versRes.data.data || []
  clusters.value = clRes.data.data || []
  regions.value = rRes.data.data || []
  datacenters.value = dcRes.data.data || []
  selectedVersion.value = versions.value[0] || null
  analysisForm.runtime_version = versions.value[0]?.runtime_version || ''
  syncCompareVersion()
}

function openNewVersion() {
  versionForm.content = currentContent.value
  versionForm.comment = ''
  if (!versionModal) versionModal = new window.bootstrap.Modal(document.getElementById('versionModal'))
  versionModal.show()
}

async function saveVersion() {
  await createVersion(route.params.id, versionForm)
  versionModal.hide()
  await loadData()
}

function openDeploy() {
  deployForm.scope = 'cluster'
  deployForm.cluster_id = clusters.value[0]?.id || null
  deployForm.region_id = regions.value[0]?.id || null
  deployForm.datacenter_id = datacenters.value[0]?.id || null
  deployForm.node_ids_text = ''
  if (!deployModal) deployModal = new window.bootstrap.Modal(document.getElementById('deployModal'))
  deployModal.show()
}

async function submitDeploy() {
  const data = { config_version_id: selectedVersion.value.id }
  if (deployForm.scope === 'cluster' && deployForm.cluster_id) data.cluster_id = deployForm.cluster_id
  if (deployForm.scope === 'region' && deployForm.region_id) data.region_id = deployForm.region_id
  if (deployForm.scope === 'datacenter' && deployForm.datacenter_id) data.datacenter_id = deployForm.datacenter_id
  if (deployForm.scope === 'node' && deployForm.node_ids_text) {
    data.node_ids = deployForm.node_ids_text.split(',').map((value) => parseInt(value.trim(), 10)).filter((value) => value > 0)
  }
  await createDeploy(data)
  deployModal.hide()
  router.push('/deploys')
}

function requireCurrentContent(actionLabel) {
  if (!currentContent.value) {
    alert(t('configs_page.require_preview').replace('{action}', actionLabel))
    return null
  }
  return currentContent.value
}

async function runLint() {
  const content = requireCurrentContent(t('configs_page.run_lint'))
  if (!content) return

  try {
    analysisResult.value = await lintConfig({
      fluent_type: template.value.fluent_type,
      runtime_version: analysisForm.runtime_version,
      content,
    })
    analysisTab.value = 'analysis'
  } catch (error) {
    alert(`${t('configs_page.lint_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runCompatibility() {
  const content = requireCurrentContent(t('configs_page.run_compatibility'))
  if (!content) return

  try {
    const payload = {
      fluent_type: template.value.fluent_type,
      runtime_version: analysisForm.runtime_version,
      content,
    }
    if (analysisForm.node_id) {
      payload.node_id = Number(analysisForm.node_id)
    }
    compatibilityResult.value = await checkCompatibility(payload)
    analysisTab.value = 'compatibility'
  } catch (error) {
    alert(`${t('configs_page.compatibility_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runReplay() {
  const content = requireCurrentContent(t('configs_page.run_replay'))
  if (!content) return
  if (!analysisForm.sample_log) {
    alert(t('configs_page.require_sample_log'))
    return
  }

  try {
    replayResult.value = await replayConfig({
      fluent_type: template.value.fluent_type,
      runtime_version: analysisForm.runtime_version,
      content,
      sample_log: analysisForm.sample_log,
      sample_tag: analysisForm.sample_tag,
    })
    analysisTab.value = 'replay'
  } catch (error) {
    alert(`${t('configs_page.replay_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runSemanticDiff() {
  const content = requireCurrentContent(t('configs_page.run_diff'))
  if (!content) return
  if (!compareVersion.value?.content) {
    alert(t('config_detail.require_compare_version'))
    return
  }

  try {
    diffResult.value = await diffConfig({
      fluent_type: template.value.fluent_type,
      before_content: compareVersion.value.content,
      after_content: content,
    })
    analysisTab.value = 'diff'
  } catch (error) {
    alert(`${t('configs_page.diff_failed')}: ${getErrorMessage(error)}`)
  }
}

watch(selectedVersion, (version) => {
  analysisForm.runtime_version = version?.runtime_version || analysisForm.runtime_version || ''
  resetAnalysisResults()
  syncCompareVersion()
})

watch(versions, syncCompareVersion)

onMounted(loadData)
</script>

<style scoped>
.fm-config-content {
  max-height: 500px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.fm-analysis-tabs {
  gap: 0.5rem;
}

.fm-analysis-tabs .nav-link {
  border: 0;
  color: #475569;
  font-weight: 600;
  border-radius: 10px;
}

.fm-analysis-tabs .nav-link.active {
  background: linear-gradient(135deg, #0f766e 0%, #0d9488 100%);
  color: #fff;
}
</style>
