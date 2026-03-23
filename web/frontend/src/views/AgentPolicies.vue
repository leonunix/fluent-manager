<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('agent_policies_page.title') }}</h4>
        <div class="text-muted">{{ t('agent_policies_page.subtitle') }}</div>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('agent_policies_page.create') }}
      </button>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-xl-5">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-header bg-white border-0 pb-0">
            <h6 class="mb-1">{{ t('agent_policies_page.global_defaults') }}</h6>
            <div class="text-muted small">{{ t('agent_policies_page.create_global_first') }}</div>
          </div>
          <div class="card-body">
            <div class="row g-3">
              <div
                v-for="item in settingsEntries(defaults)"
                :key="`default-${item.key}`"
                class="col-md-6"
              >
                <div class="border rounded-3 p-3 h-100">
                  <div class="text-muted small mb-1">{{ item.label }}</div>
                  <div class="fw-semibold text-break">{{ item.value }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-xl-7">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-header bg-white border-0 pb-0">
            <h6 class="mb-1">{{ t('agent_policies_page.resolved_preview') }}</h6>
            <div class="text-muted small">{{ t('agent_policies_page.select_node_preview') }}</div>
          </div>
          <div class="card-body">
            <div class="row g-3 align-items-end mb-4">
              <div class="col-md-8">
                <label class="form-label">{{ t('agent_policies_page.node_preview') }}</label>
                <select v-model="selectedNodeID" class="form-select" @change="loadResolvedPreview">
                  <option :value="null">{{ t('agent_policies_page.choose_node') }}</option>
                  <option v-for="node in nodes" :key="node.id" :value="node.id">
                    {{ node.hostname }} (#{{ node.id }})
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <button class="btn btn-outline-primary w-100" :disabled="!selectedNodeID" @click="loadResolvedPreview">
                  <i class="bi bi-arrow-repeat me-1"></i>{{ t('common.refresh') }}
                </button>
              </div>
            </div>

            <div v-if="resolved">
              <div class="row g-4">
                <div class="col-md-5">
                  <div class="border rounded-3 p-3 h-100">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <h6 class="mb-0">{{ t('agent_policies_page.matched_policies') }}</h6>
                      <span class="badge text-bg-secondary">{{ resolved.matched_policies.length }}</span>
                    </div>
                    <div v-if="resolved.matched_policies.length">
                      <div
                        v-for="policy in resolved.matched_policies"
                        :key="`matched-${policy.id}`"
                        class="fm-policy-match"
                      >
                        <div class="fw-semibold">{{ policy.name }}</div>
                        <div class="small text-muted">
                          {{ scopeLabel(policy.scope_type) }} · P{{ policy.priority }}
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-muted small">{{ t('agent_policies_page.no_matches') }}</div>
                  </div>
                </div>
                <div class="col-md-7">
                  <div class="border rounded-3 p-3 h-100">
                    <h6 class="mb-3">{{ t('agent_policies_page.effective_settings') }}</h6>
                    <div class="row g-3">
                      <div
                        v-for="item in settingsEntries(resolved.settings)"
                        :key="`resolved-${item.key}`"
                        class="col-md-6"
                      >
                        <div class="fm-setting-tile">
                          <div class="text-muted small mb-1">{{ item.label }}</div>
                          <div class="fw-semibold text-break">{{ item.value }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-muted small">{{ t('agent_policies_page.select_node_preview') }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-header bg-white border-0 pb-0">
        <h6 class="mb-1">{{ t('agent_policies_page.policies') }}</h6>
        <div class="text-muted small">{{ t('agent_policies_page.disabled_note') }}</div>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('agent_policies_page.scope') }}</th>
                <th>{{ t('agent_policies_page.target') }}</th>
                <th>{{ t('agent_policies_page.priority') }}</th>
                <th>{{ t('agent_policies_page.override_count') }}</th>
                <th>{{ t('agent_policies_page.enabled') }}</th>
                <th>{{ t('agent_policies_page.updated_at') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="policy in policies" :key="policy.id">
                <td>
                  <div class="fw-semibold">{{ policy.name }}</div>
                  <div class="small text-muted">{{ policy.description || '-' }}</div>
                </td>
                <td>
                  <span class="badge rounded-pill fm-scope-badge">{{ scopeLabel(policy.scope_type) }}</span>
                </td>
                <td class="text-break">{{ targetLabel(policy) }}</td>
                <td><span class="badge text-bg-light">P{{ policy.priority }}</span></td>
                <td>{{ overrideCount(policy.settings) }}</td>
                <td>
                  <span class="badge" :class="policy.is_enabled ? 'text-bg-success' : 'text-bg-secondary'">
                    {{ policy.is_enabled ? t('yes') : t('no') }}
                  </span>
                </td>
                <td>{{ formatTime(policy.updated_at) }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(policy)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="handleDelete(policy)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!policies.length">
                <td colspan="8" class="text-center text-muted py-4">{{ t('agent_policies_page.no_policies') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal fade" id="agentPolicyModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ editingId ? t('agent_policies_page.edit') : t('agent_policies_page.create') }}
            </h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-light border mb-4">{{ t('agent_policies_page.empty_inherits') }}</div>

            <h6 class="mb-3">{{ t('agent_policies_page.general_section') }}</h6>
            <div class="row g-3 mb-4">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.name') }}</label>
                <input v-model="form.name" type="text" class="form-control">
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('common.description') }}</label>
                <input v-model="form.description" type="text" class="form-control">
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('agent_policies_page.scope') }}</label>
                <select v-model="form.scope_type" class="form-select" @change="handleScopeChange">
                  <option value="global">{{ t('agent_policies_page.scope_global') }}</option>
                  <option value="environment">{{ t('agent_policies_page.scope_environment') }}</option>
                  <option value="cluster">{{ t('agent_policies_page.scope_cluster') }}</option>
                  <option value="label_selector">{{ t('agent_policies_page.scope_label_selector') }}</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('agent_policies_page.priority') }}</label>
                <input v-model.number="form.priority" type="number" class="form-control">
              </div>
              <div class="col-md-4 d-flex align-items-end">
                <div class="form-check">
                  <input id="agentPolicyEnabled" v-model="form.is_enabled" type="checkbox" class="form-check-input">
                  <label for="agentPolicyEnabled" class="form-check-label">{{ t('agent_policies_page.enabled') }}</label>
                </div>
              </div>
            </div>

            <h6 class="mb-3">{{ t('agent_policies_page.scope_target') }}</h6>
            <div class="row g-3 mb-4">
              <div v-if="form.scope_type === 'environment'" class="col-md-6">
                <label class="form-label">{{ t('agent_policies_page.choose_environment') }}</label>
                <select v-model="form.environment_id" class="form-select">
                  <option :value="null">{{ t('agent_policies_page.choose_environment') }}</option>
                  <option v-for="env in environments" :key="env.id" :value="env.id">
                    {{ env.alias || env.name }}
                  </option>
                </select>
              </div>
              <div v-if="form.scope_type === 'cluster'" class="col-md-6">
                <label class="form-label">{{ t('agent_policies_page.choose_cluster') }}</label>
                <select v-model="form.cluster_id" class="form-select">
                  <option :value="null">{{ t('agent_policies_page.choose_cluster') }}</option>
                  <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
                    {{ cluster.alias || cluster.name }}
                  </option>
                </select>
              </div>
              <div v-if="form.scope_type === 'label_selector'" class="col-12">
                <label class="form-label">{{ t('agent_policies_page.scope_label_selector') }}</label>
                <textarea
                  v-model="form.label_selector"
                  class="form-control font-monospace"
                  rows="4"
                  :placeholder="t('agent_policies_page.label_selector_help')"
                ></textarea>
              </div>
            </div>

            <h6 class="mb-3">{{ t('agent_policies_page.runtime_section') }}</h6>
            <div class="row g-3 mb-4">
              <div class="col-md-4" v-for="field in runtimeFields" :key="field.key">
                <label class="form-label">{{ field.label }}</label>
                <input
                  v-model="form[field.key]"
                  :type="field.type"
                  class="form-control"
                  :placeholder="field.placeholder"
                >
              </div>
            </div>

            <h6 class="mb-3">{{ t('agent_policies_page.fluent_section') }}</h6>
            <div class="row g-3">
              <div class="col-md-4">
                <label class="form-label">fluent_type</label>
                <select v-model="form.fluent_type" class="form-select">
                  <option value="">Inherit</option>
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div v-for="field in fluentFields" :key="field.key" :class="field.className || 'col-md-4'">
                <label class="form-label">{{ field.label }}</label>
                <textarea
                  v-if="field.as === 'textarea'"
                  v-model="form[field.key]"
                  class="form-control font-monospace"
                  :rows="field.rows || 3"
                  :placeholder="field.placeholder"
                ></textarea>
                <input
                  v-else
                  v-model="form[field.key]"
                  type="text"
                  class="form-control"
                  :placeholder="field.placeholder"
                >
                <div v-if="field.help" class="form-text">{{ field.help }}</div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="savePolicy">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import {
  createAgentPolicy,
  deleteAgentPolicy,
  getAgentPolicies,
  getClusters,
  getEnvironments,
  getNodes,
  resolveAgentPolicyForNode,
  updateAgentPolicy,
} from '../api'
import { useI18n } from '../i18n'

const { t, dateLocale } = useI18n()

const policies = ref([])
const defaults = ref({})
const environments = ref([])
const clusters = ref([])
const nodes = ref([])
const resolved = ref(null)
const selectedNodeID = ref(null)
const editingId = ref(null)

const runtimeFields = [
  { key: 'heartbeat_interval', label: 'heartbeat_interval', type: 'number', placeholder: '30' },
  { key: 'metrics_interval', label: 'metrics_interval', type: 'number', placeholder: '60' },
  { key: 'log_upload_interval', label: 'log_upload_interval', type: 'number', placeholder: '120' },
  { key: 'log_buffer_lines', label: 'log_buffer_lines', type: 'number', placeholder: '500' },
  { key: 'health_port', label: 'health_port', type: 'number', placeholder: '9880' },
  { key: 'max_retries', label: 'max_retries', type: 'number', placeholder: '5' },
  { key: 'retry_base_delay', label: 'retry_base_delay', type: 'number', placeholder: '1000' },
  { key: 'max_backups', label: 'max_backups', type: 'number', placeholder: '10' },
]

const fluentFields = [
  { key: 'fluent_config_path', label: 'fluent_config_path', placeholder: '/etc/fluent-bit/fluent-bit.conf' },
  { key: 'fluent_config_dir', label: 'fluent_config_dir', placeholder: '/etc/fluent-bit/conf.d' },
  { key: 'fluent_binary', label: 'fluent_binary', placeholder: '/opt/fluent-bit/bin/fluent-bit' },
  { key: 'fluent_service_unit', label: 'fluent_service_unit', placeholder: 'fluent-bit.service' },
  { key: 'fluent_restart_cmd', label: 'fluent_restart_cmd', placeholder: 'systemctl restart fluent-bit' },
  { key: 'fluent_reload_cmd', label: 'fluent_reload_cmd', placeholder: 'systemctl reload fluent-bit' },
  { key: 'fluent_dry_run_cmd', label: 'fluent_dry_run_cmd', placeholder: 'fluent-bit --dry-run -c ...' },
  { key: 'fluent_log_path', label: 'fluent_log_path', placeholder: '/var/log/fluent-bit.log' },
  { key: 'fluent_metrics_url', label: 'fluent_metrics_url', placeholder: 'http://127.0.0.1:2020/api/v1/metrics/prometheus' },
  { key: 'fluent_metrics_format', label: 'fluent_metrics_format', placeholder: 'prometheus' },
  { key: 'backup_dir', label: 'backup_dir', placeholder: '/var/lib/fluent-manager-agent/backups' },
  {
    key: 'fluent_extra_files',
    label: 'fluent_extra_files',
    as: 'textarea',
    rows: 5,
    className: 'col-12',
    placeholder: '/etc/fluent-bit/parsers.conf',
    help: '',
  },
]

const form = reactive(createEmptyForm())
let modal = null

function createEmptyForm() {
  return {
    name: '',
    description: '',
    scope_type: 'global',
    environment_id: null,
    cluster_id: null,
    label_selector: '',
    priority: 100,
    is_enabled: true,
    heartbeat_interval: '',
    metrics_interval: '',
    log_upload_interval: '',
    log_buffer_lines: '',
    health_port: '',
    max_retries: '',
    retry_base_delay: '',
    fluent_type: '',
    fluent_config_path: '',
    fluent_config_dir: '',
    fluent_binary: '',
    fluent_service_unit: '',
    fluent_restart_cmd: '',
    fluent_reload_cmd: '',
    fluent_dry_run_cmd: '',
    fluent_log_path: '',
    fluent_extra_files: '',
    fluent_metrics_url: '',
    fluent_metrics_format: '',
    backup_dir: '',
    max_backups: '',
  }
}

function ensureModal() {
  if (!modal) {
    modal = new window.bootstrap.Modal(document.getElementById('agentPolicyModal'))
  }
}

function resetForm() {
  Object.assign(form, createEmptyForm())
  editingId.value = null
}

function handleScopeChange() {
  if (form.scope_type !== 'environment') form.environment_id = null
  if (form.scope_type !== 'cluster') form.cluster_id = null
  if (form.scope_type !== 'label_selector') form.label_selector = ''
}

function scopeLabel(scopeType) {
  return t(`agent_policies_page.scope_${scopeType}`)
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function normalizeDisplayValue(value) {
  if (Array.isArray(value)) {
    return value.length ? value.join(', ') : '-'
  }
  if (typeof value === 'boolean') {
    return value ? t('yes') : t('no')
  }
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  return String(value)
}

function settingsEntries(settings) {
  return [
    'heartbeat_interval',
    'metrics_interval',
    'log_upload_interval',
    'log_buffer_lines',
    'health_port',
    'max_retries',
    'retry_base_delay',
    'fluent_type',
    'fluent_config_path',
    'fluent_config_dir',
    'fluent_binary',
    'fluent_service_unit',
    'fluent_restart_cmd',
    'fluent_reload_cmd',
    'fluent_dry_run_cmd',
    'fluent_log_path',
    'fluent_extra_files',
    'fluent_metrics_url',
    'fluent_metrics_format',
    'backup_dir',
    'max_backups',
  ].map((key) => ({
    key,
    label: key,
    value: normalizeDisplayValue(settings?.[key]),
  }))
}

function overrideCount(settings) {
  return Object.values(settings || {}).filter((value) => {
    if (Array.isArray(value)) return value.length > 0
    return value !== null && value !== undefined && value !== ''
  }).length
}

function targetLabel(policy) {
  if (policy.scope_type === 'global') return '-'
  if (policy.scope_type === 'environment') {
    return policy.environment?.alias || policy.environment?.name || `#${policy.environment_id}`
  }
  if (policy.scope_type === 'cluster') {
    return policy.cluster?.alias || policy.cluster?.name || `#${policy.cluster_id}`
  }
  return policy.label_selector || '-'
}

function parseOptionalInt(value) {
  if (value === '' || value === null || value === undefined) return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function parseOptionalString(value) {
  const trimmed = String(value || '').trim()
  return trimmed ? trimmed : undefined
}

function parseExtraFiles(value) {
  const items = String(value || '')
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  return items.length ? items : undefined
}

function buildPayload() {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    scope_type: form.scope_type,
    environment_id: form.scope_type === 'environment' ? form.environment_id || null : null,
    cluster_id: form.scope_type === 'cluster' ? form.cluster_id || null : null,
    label_selector: form.scope_type === 'label_selector' ? form.label_selector.trim() : '',
    priority: Number(form.priority) || 0,
    is_enabled: !!form.is_enabled,
    settings: {
      heartbeat_interval: parseOptionalInt(form.heartbeat_interval),
      metrics_interval: parseOptionalInt(form.metrics_interval),
      log_upload_interval: parseOptionalInt(form.log_upload_interval),
      log_buffer_lines: parseOptionalInt(form.log_buffer_lines),
      health_port: parseOptionalInt(form.health_port),
      max_retries: parseOptionalInt(form.max_retries),
      retry_base_delay: parseOptionalInt(form.retry_base_delay),
      fluent_type: parseOptionalString(form.fluent_type),
      fluent_config_path: parseOptionalString(form.fluent_config_path),
      fluent_config_dir: parseOptionalString(form.fluent_config_dir),
      fluent_binary: parseOptionalString(form.fluent_binary),
      fluent_service_unit: parseOptionalString(form.fluent_service_unit),
      fluent_restart_cmd: parseOptionalString(form.fluent_restart_cmd),
      fluent_reload_cmd: parseOptionalString(form.fluent_reload_cmd),
      fluent_dry_run_cmd: parseOptionalString(form.fluent_dry_run_cmd),
      fluent_log_path: parseOptionalString(form.fluent_log_path),
      fluent_extra_files: parseExtraFiles(form.fluent_extra_files),
      fluent_metrics_url: parseOptionalString(form.fluent_metrics_url),
      fluent_metrics_format: parseOptionalString(form.fluent_metrics_format),
      backup_dir: parseOptionalString(form.backup_dir),
      max_backups: parseOptionalInt(form.max_backups),
    },
  }
}

function applyPolicyToForm(policy) {
  resetForm()
  editingId.value = policy.id
  form.name = policy.name || ''
  form.description = policy.description || ''
  form.scope_type = policy.scope_type || 'global'
  form.environment_id = policy.environment_id || null
  form.cluster_id = policy.cluster_id || null
  form.label_selector = policy.label_selector || ''
  form.priority = policy.priority ?? 100
  form.is_enabled = !!policy.is_enabled

  const settings = policy.settings || {}
  for (const key of Object.keys(settings)) {
    if (key === 'fluent_extra_files') {
      form[key] = Array.isArray(settings[key]) ? settings[key].join('\n') : ''
      continue
    }
    form[key] = settings[key] === null || settings[key] === undefined ? '' : String(settings[key])
  }
}

async function loadPolicies() {
  const response = await getAgentPolicies()
  policies.value = response.data || []
  defaults.value = response.defaults || {}
}

async function loadReferenceData() {
  const [envRes, clusterRes, nodeRes] = await Promise.all([
    getEnvironments(),
    getClusters(),
    getNodes({ page: 1, page_size: 200 }),
  ])

  environments.value = envRes.data.data || []
  clusters.value = clusterRes.data.data || []
  nodes.value = nodeRes.data.data || []
}

async function loadResolvedPreview() {
  if (!selectedNodeID.value) {
    resolved.value = null
    return
  }

  try {
    resolved.value = await resolveAgentPolicyForNode(Number(selectedNodeID.value))
  } catch (error) {
    alert(`${t('agent_policies_page.resolve_failed')}: ${getErrorMessage(error)}`)
  }
}

function openCreate() {
  resetForm()
  ensureModal()
  modal.show()
}

function openEdit(policy) {
  applyPolicyToForm(policy)
  ensureModal()
  modal.show()
}

async function savePolicy() {
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateAgentPolicy(editingId.value, payload)
    } else {
      await createAgentPolicy(payload)
    }
    modal.hide()
    await loadPolicies()
    if (selectedNodeID.value) {
      await loadResolvedPreview()
    }
  } catch (error) {
    alert(`${t('agent_policies_page.save_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDelete(policy) {
  if (!confirm(t('agent_policies_page.confirm_delete').replace('{name}', policy.name))) return

  try {
    await deleteAgentPolicy(policy.id)
    await loadPolicies()
    if (selectedNodeID.value) {
      await loadResolvedPreview()
    }
  } catch (error) {
    alert(`${t('agent_policies_page.delete_failed')}: ${getErrorMessage(error)}`)
  }
}

onMounted(async () => {
  fluentFields[fluentFields.length - 1].help = t('agent_policies_page.fluent_extra_files_help')
  await Promise.all([loadPolicies(), loadReferenceData()])
})
</script>

<style scoped>
.fm-setting-tile {
  border: 1px solid rgba(148, 163, 184, 0.25);
  border-radius: 14px;
  padding: 0.9rem 1rem;
  height: 100%;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.95) 0%, #ffffff 100%);
}

.fm-policy-match {
  padding: 0.75rem 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 12px;
  margin-bottom: 0.75rem;
  background: #fff;
}

.fm-policy-match:last-child {
  margin-bottom: 0;
}

.fm-scope-badge {
  background: #e2e8f0;
  color: #0f172a;
  font-weight: 600;
}
</style>
