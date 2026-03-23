<template>
  <div v-if="node">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <router-link to="/nodes" class="text-decoration-none">&larr; {{ t('node_detail.back') }}</router-link>
        <h4 class="mt-2 mb-0">{{ node.hostname }}</h4>
        <span class="text-muted small">{{ node.node_uid }}</span>
        <span :class="statusClass(node.status)" class="badge ms-2">{{ statusText(node.status) }}</span>
        <span class="badge bg-info ms-1">{{ node.fluent_type }}</span>
      </div>
      <div class="btn-group">
        <button class="btn btn-outline-success btn-sm" @click="sendCmd('restart')">
          <i class="bi bi-arrow-clockwise"></i> {{ t('node_detail.restart_service') }}
        </button>
        <button class="btn btn-outline-primary btn-sm" @click="sendCmd('reload')">
          <i class="bi bi-arrow-repeat"></i> {{ t('node_detail.hot_reload') }}
        </button>
        <button class="btn btn-outline-warning btn-sm" @click="sendCmd('status')">
          <i class="bi bi-info-circle"></i> {{ t('node_detail.view_status') }}
        </button>
        <button class="btn btn-outline-secondary btn-sm" @click="sendCmd('validate')">
          <i class="bi bi-check2-circle"></i> {{ t('node_detail.validate_config') }}
        </button>
        <button class="btn btn-outline-danger btn-sm" @click="sendCmd('rollback')">
          <i class="bi bi-arrow-counterclockwise"></i> {{ t('node_detail.rollback') }}
        </button>
      </div>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <h6 class="text-muted">{{ t('node_detail.node_info') }}</h6>
            <table class="table table-sm table-borderless mb-0">
              <tr><td class="text-muted">IP</td><td>{{ node.ip_address || '-' }}</td></tr>
              <tr><td class="text-muted">{{ t('node_detail.system') }}</td><td>{{ node.os }}</td></tr>
              <tr><td class="text-muted">{{ t('node_detail.agent') }}</td><td>{{ node.agent_version }}</td></tr>
              <tr><td class="text-muted">{{ t('node_detail.fluent') }}</td><td>{{ node.fluent_version }}</td></tr>
              <tr><td class="text-muted">{{ t('common.role') }}</td><td>{{ node.node_role || '-' }}</td></tr>
              <tr><td class="text-muted">{{ t('flow_graph.aggregation_group') }}</td><td>{{ node.aggregation_group?.alias || node.aggregation_group?.name || '-' }}</td></tr>
              <tr><td class="text-muted">{{ t('node_detail.heartbeat') }}</td><td>{{ formatTime(node.last_heartbeat) }}</td></tr>
            </table>
          </div>
        </div>
      </div>
      <div class="col-md-9" v-if="metrics">
        <div class="row g-3">
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">{{ t('node_detail.cpu_usage') }}</div>
                <h3 :class="colorClass(metrics.cpu_usage_percent, 80, 90)">{{ metrics.cpu_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.cpu_usage_percent, 80, 90)" :style="{ width: metrics.cpu_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">Load: {{ metrics.load_avg_1?.toFixed(2) }} / {{ metrics.load_avg_5?.toFixed(2) }} / {{ metrics.load_avg_15?.toFixed(2) }}</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">{{ t('node_detail.memory_usage') }}</div>
                <h3 :class="colorClass(metrics.mem_usage_percent, 80, 90)">{{ metrics.mem_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.mem_usage_percent, 80, 90)" :style="{ width: metrics.mem_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">{{ metrics.mem_used_mb }} / {{ metrics.mem_total_mb }} MB</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">{{ t('node_detail.disk_usage') }}</div>
                <h3 :class="colorClass(metrics.disk_usage_percent, 80, 90)">{{ metrics.disk_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.disk_usage_percent, 80, 90)" :style="{ width: metrics.disk_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">{{ metrics.disk_used_gb }} / {{ metrics.disk_total_gb }} GB</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">{{ t('node_detail.fluent_process') }}</div>
                <h3 :class="metrics.fluent_running ? 'text-success' : 'text-danger'">
                  {{ metrics.fluent_running ? t('node_detail.running') : t('node_detail.stopped') }}
                </h3>
                <div class="text-muted small" v-if="metrics.fluent_running">
                  PID: {{ metrics.fluent_pid }} |
                  CPU: {{ metrics.fluent_cpu_percent?.toFixed(1) }}% |
                  Mem: {{ metrics.fluent_mem_mb?.toFixed(1) }}MB
                </div>
                <div class="text-muted small" v-if="metrics.fluent_running">
                  FDs: {{ metrics.fluent_open_fds }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-header bg-white d-flex justify-content-between align-items-center">
        <h6 class="mb-0">{{ t('node_detail.fluent_profile') }}</h6>
        <button class="btn btn-sm btn-outline-primary" @click="openProfileEditor">
          <i class="bi bi-pencil-square me-1"></i>{{ t('node_detail.edit_profile') }}
        </button>
      </div>
      <div class="card-body">
        <div v-if="!node.fluent_profile" class="alert alert-warning py-2">
          {{ t('node_detail.profile_missing') }}
        </div>
        <div class="row g-3">
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('node_detail.hot_reload') }}</div>
              <div class="fw-semibold">{{ profileSupport(node.fluent_profile?.supports_hot_reload) }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('node_detail.multiline') }}</div>
              <div class="fw-semibold">{{ profileSupport(node.fluent_profile?.supports_multiline) }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('node_detail.storage_layer') }}</div>
              <div class="fw-semibold">{{ profileSupport(node.fluent_profile?.supports_storage_layer) }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('node_detail.metrics_api') }}</div>
              <div class="fw-semibold">{{ profileSupport(node.fluent_profile?.supports_metrics_api) }}</div>
            </div>
          </div>
        </div>
        <div class="row g-3 mt-1">
          <div class="col-md-6">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-2">{{ t('node_detail.loaded_plugins') }}</div>
              <code>{{ node.fluent_profile?.loaded_plugins || '-' }}</code>
            </div>
          </div>
          <div class="col-md-6">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-2">{{ t('node_detail.last_reported') }}</div>
              <div class="fw-semibold">{{ formatTime(node.fluent_profile?.last_reported_at) }}</div>
              <div class="small text-muted mt-2">
                {{ t('node_detail.forward_tls') }}: {{ profileSupport(node.fluent_profile?.supports_forward_tls) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <a class="nav-link" :class="{ active: tab === 'commands' }" href="#" @click.prevent="tab = 'commands'">
          <i class="bi bi-terminal me-1"></i>{{ t('node_detail.command_history') }}
        </a>
      </li>
      <li class="nav-item">
        <a class="nav-link" :class="{ active: tab === 'logs' }" href="#" @click.prevent="tab = 'logs'; loadLogs()">
          <i class="bi bi-journal-text me-1"></i>{{ t('node_detail.fluent_logs') }}
        </a>
      </li>
    </ul>

    <div v-if="tab === 'commands'" class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>ID</th>
              <th>{{ t('node_detail.command') }}</th>
              <th>{{ t('node_detail.args') }}</th>
              <th>{{ t('status') }}</th>
              <th>{{ t('node_detail.output') }}</th>
              <th>{{ t('node_detail.operator') }}</th>
              <th>{{ t('dashboard.time') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cmd in commands" :key="cmd.id">
              <td>#{{ cmd.id }}</td>
              <td><span class="badge bg-secondary">{{ cmd.action }}</span></td>
              <td class="text-truncate" style="max-width: 150px;">{{ cmd.args || '-' }}</td>
              <td>
                <span :class="cmdStatusClass(cmd.status)" class="badge">{{ cmd.status }}</span>
              </td>
              <td>
                <button v-if="cmd.output" class="btn btn-sm btn-outline-secondary" @click="showOutput(cmd)">
                  <i class="bi bi-eye"></i> {{ t('common.view') }}
                </button>
                <span v-else class="text-muted">-</span>
              </td>
              <td>{{ cmd.creator?.username || '-' }}</td>
              <td>{{ formatTime(cmd.created_at) }}</td>
            </tr>
            <tr v-if="!commands.length">
              <td colspan="7" class="text-center text-muted">{{ t('node_detail.no_commands') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="tab === 'logs'" class="card border-0 shadow-sm">
      <div class="card-body">
        <div v-if="logs.length">
          <div v-for="logEntry in logs" :key="logEntry.id" class="mb-3">
            <div class="text-muted small mb-1">
              {{ formatTime(logEntry.created_at) }} - {{ logEntry.line_count }} {{ t('node_detail.lines') }}
            </div>
            <pre class="bg-dark text-light p-2 rounded small fm-log-block">{{ logEntry.lines }}</pre>
          </div>
        </div>
        <div v-else class="text-center text-muted">{{ t('node_detail.no_logs') }}</div>
      </div>
    </div>

    <div class="modal fade" id="profileModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('node_detail.edit_profile') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.role') }}</label>
                <select v-model="profileForm.node_role" class="form-select">
                  <option value="standalone">standalone</option>
                  <option value="edge_collector">edge_collector</option>
                  <option value="aggregator">aggregator</option>
                  <option value="gateway">gateway</option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('flow_graph.aggregation_group') }}</label>
                <select v-model="profileForm.aggregation_group_id" class="form-select">
                  <option :value="null">{{ t('node_detail.no_aggregation_group') }}</option>
                  <option v-for="group in aggregationGroups" :key="group.id" :value="group.id">
                    {{ group.alias || group.name }}
                  </option>
                </select>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('node_detail.loaded_plugins') }}</label>
              <textarea v-model="profileForm.loaded_plugins" class="form-control font-monospace" rows="4"></textarea>
            </div>

            <div class="row g-3 mb-3">
              <div class="col-md-4">
                <div class="form-check">
                  <input id="profileHotReload" v-model="profileForm.supports_hot_reload" type="checkbox" class="form-check-input">
                  <label for="profileHotReload" class="form-check-label">{{ t('node_detail.hot_reload') }}</label>
                </div>
              </div>
              <div class="col-md-4">
                <div class="form-check">
                  <input id="profileMultiline" v-model="profileForm.supports_multiline" type="checkbox" class="form-check-input">
                  <label for="profileMultiline" class="form-check-label">{{ t('node_detail.multiline') }}</label>
                </div>
              </div>
              <div class="col-md-4">
                <div class="form-check">
                  <input id="profileStorage" v-model="profileForm.supports_storage_layer" type="checkbox" class="form-check-input">
                  <label for="profileStorage" class="form-check-label">{{ t('node_detail.storage_layer') }}</label>
                </div>
              </div>
              <div class="col-md-4">
                <div class="form-check">
                  <input id="profileForwardTLS" v-model="profileForm.supports_forward_tls" type="checkbox" class="form-check-input">
                  <label for="profileForwardTLS" class="form-check-label">{{ t('node_detail.forward_tls') }}</label>
                </div>
              </div>
              <div class="col-md-4">
                <div class="form-check">
                  <input id="profileMetrics" v-model="profileForm.supports_metrics_api" type="checkbox" class="form-check-input">
                  <label for="profileMetrics" class="form-check-label">{{ t('node_detail.metrics_api') }}</label>
                </div>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('node_detail.metadata') }}</label>
              <textarea v-model="profileForm.metadata" class="form-control font-monospace" rows="5" :placeholder="t('node_detail.metadata_placeholder')"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveProfile">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="outputModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('node_detail.command_output').replace('{id}', selectedCmd?.id || '') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <pre class="bg-dark text-light p-3 rounded fm-log-block">{{ selectedCmd?.output }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  getAggregationGroups,
  getNode,
  getNodeCommands,
  getNodeLogs,
  getNodeMetrics,
  sendNodeCommand,
  updateNodeFluentProfile,
} from '../api'
import { useI18n } from '../i18n'

const route = useRoute()
const node = ref(null)
const metrics = ref(null)
const commands = ref([])
const logs = ref([])
const aggregationGroups = ref([])
const tab = ref('commands')
const selectedCmd = ref(null)

const profileForm = reactive({
  node_role: 'standalone',
  aggregation_group_id: null,
  loaded_plugins: '',
  supports_hot_reload: false,
  supports_multiline: false,
  supports_storage_layer: false,
  supports_forward_tls: false,
  supports_metrics_api: false,
  metadata: '',
})

let outputModal = null
let profileModal = null
const { t, dateLocale } = useI18n()

function statusClass(s) {
  return { 'bg-success': s === 'online', 'bg-warning': s === 'offline', 'bg-danger': s === 'error' }
}

function statusText(s) {
  return { online: t('nodes_page.online'), offline: t('nodes_page.offline'), error: t('nodes_page.error') }[s] || s
}

function cmdStatusClass(s) {
  return { 'bg-success': s === 'success', 'bg-warning': s === 'pending' || s === 'delivered', 'bg-danger': s === 'failed' }
}

function colorClass(val, warn, danger) {
  if (val >= danger) return 'text-danger'
  if (val >= warn) return 'text-warning'
  return 'text-success'
}

function barClass(val, warn, danger) {
  if (val >= danger) return 'bg-danger'
  if (val >= warn) return 'bg-warning'
  return 'bg-success'
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function profileSupport(value) {
  return value ? t('node_detail.supported') : t('node_detail.unsupported')
}

function syncProfileForm() {
  const profile = node.value?.fluent_profile || {}
  profileForm.node_role = node.value?.node_role || 'standalone'
  profileForm.aggregation_group_id = node.value?.aggregation_group_id || null
  profileForm.loaded_plugins = profile.loaded_plugins || ''
  profileForm.supports_hot_reload = !!profile.supports_hot_reload
  profileForm.supports_multiline = !!profile.supports_multiline
  profileForm.supports_storage_layer = !!profile.supports_storage_layer
  profileForm.supports_forward_tls = !!profile.supports_forward_tls
  profileForm.supports_metrics_api = !!profile.supports_metrics_api
  profileForm.metadata = profile.metadata || ''
}

async function loadCommands() {
  try {
    const { data } = await getNodeCommands(route.params.id)
    commands.value = data.data || []
  } catch (error) {
    console.error(error)
  }
}

async function loadLogs() {
  try {
    const { data } = await getNodeLogs(route.params.id)
    logs.value = data.data || []
  } catch (error) {
    console.error(error)
  }
}

async function loadNodeData() {
  const id = route.params.id
  const [nodeRes, metricsRes, groupsRes] = await Promise.all([
    getNode(id),
    getNodeMetrics(id).catch(() => ({ data: null })),
    getAggregationGroups().catch(() => []),
  ])
  node.value = nodeRes.data
  metrics.value = metricsRes.data
  aggregationGroups.value = groupsRes || []
  syncProfileForm()
}

async function sendCmd(action) {
  if (action === 'rollback' && !confirm(t('node_detail.confirm_rollback'))) return
  try {
    await sendNodeCommand(route.params.id, { action })
    setTimeout(loadCommands, 1000)
  } catch (error) {
    alert(`${t('node_detail.command_failed')}: ${getErrorMessage(error)}`)
  }
}

function showOutput(cmd) {
  selectedCmd.value = cmd
  if (!outputModal) outputModal = new window.bootstrap.Modal(document.getElementById('outputModal'))
  outputModal.show()
}

function openProfileEditor() {
  syncProfileForm()
  if (!profileModal) profileModal = new window.bootstrap.Modal(document.getElementById('profileModal'))
  profileModal.show()
}

async function saveProfile() {
  try {
    await updateNodeFluentProfile(Number(route.params.id), {
      node_role: profileForm.node_role,
      aggregation_group_id: profileForm.aggregation_group_id || null,
      loaded_plugins: profileForm.loaded_plugins,
      supports_hot_reload: profileForm.supports_hot_reload,
      supports_multiline: profileForm.supports_multiline,
      supports_storage_layer: profileForm.supports_storage_layer,
      supports_forward_tls: profileForm.supports_forward_tls,
      supports_metrics_api: profileForm.supports_metrics_api,
      metadata: profileForm.metadata,
    })
    profileModal.hide()
    await loadNodeData()
  } catch (error) {
    alert(`${t('node_detail.save_profile_failed')}: ${getErrorMessage(error)}`)
  }
}

onMounted(async () => {
  await loadNodeData()
  loadCommands()
})
</script>

<style scoped>
.fm-log-block {
  max-height: 500px;
  overflow: auto;
}
</style>
