<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('aggregation_groups.title') }}</h4>
        <div class="text-muted">{{ t('aggregation_groups.subtitle') }}</div>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('aggregation_groups.create') }}
      </button>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('aggregation_groups.active_count') }}</div>
            <div class="fs-3 fw-bold">{{ groups.length }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('aggregation_groups.tls_enabled') }}</div>
            <div class="fs-3 fw-bold">{{ tlsEnabledCount }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('aggregation_groups.shared_key_configured') }}</div>
            <div class="fs-3 fw-bold">{{ sharedKeyCount }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('aggregation_groups.restorable') }}</div>
            <div class="fs-3 fw-bold">{{ deletedGroups.length }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body p-2">
        <div class="nav nav-pills fm-agg-tabs">
          <button
            class="nav-link"
            :class="{ active: activeTab === 'active' }"
            @click="activeTab = 'active'"
          >
            {{ t('aggregation_groups.active_tab') }}
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'deleted' }"
            @click="activeTab = 'deleted'"
          >
            {{ t('aggregation_groups.deleted_tab') }}
          </button>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('common.runtime') }}</th>
                <th>{{ t('common.mode') }}</th>
                <th>{{ t('aggregation_groups.endpoint') }}</th>
                <th>{{ t('common.cluster') }}</th>
                <th>TLS / Key</th>
                <th>{{ t('aggregation_groups.updated_at') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody v-if="activeTab === 'active'">
              <tr v-for="group in groups" :key="group.id">
                <td>
                  <div class="fw-semibold">{{ group.alias || group.name }}</div>
                  <div class="small text-muted">{{ group.name }}</div>
                </td>
                <td><span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(group.fluent_type) }}</span></td>
                <td><span class="badge text-bg-secondary">{{ group.mode }}</span></td>
                <td>
                  <code>{{ endpointText(group) }}</code>
                </td>
                <td>{{ clusterName(group.cluster_id) }}</td>
                <td>
                  <span class="badge me-1" :class="group.enable_tls ? 'text-bg-success' : 'text-bg-light'">
                    {{ group.enable_tls ? 'TLS' : t('aggregation_groups.no_tls') }}
                  </span>
                  <span class="badge" :class="group.has_shared_key ? 'text-bg-dark' : 'text-bg-light'">
                    {{ group.has_shared_key ? t('aggregation_groups.key_set') : t('aggregation_groups.no_key') }}
                  </span>
                </td>
                <td>{{ formatTime(group.updated_at) }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-secondary me-1" @click="loadMetrics(group)">
                    <i class="bi bi-bar-chart"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(group)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="handleDelete(group)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!groups.length">
                <td colspan="8" class="text-center text-muted py-4">{{ t('aggregation_groups.no_groups') }}</td>
              </tr>
            </tbody>
            <tbody v-else>
              <tr v-for="group in deletedGroups" :key="group.id">
                <td>
                  <div class="fw-semibold">{{ group.alias || group.name }}</div>
                  <div class="small text-muted">{{ group.name }}</div>
                </td>
                <td><span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(group.fluent_type) }}</span></td>
                <td><span class="badge text-bg-secondary">{{ group.mode }}</span></td>
                <td><code>{{ endpointText(group) }}</code></td>
                <td>{{ clusterName(group.cluster_id) }}</td>
                <td>
                  <span class="badge me-1" :class="group.enable_tls ? 'text-bg-success' : 'text-bg-light'">
                    {{ group.enable_tls ? 'TLS' : t('aggregation_groups.no_tls') }}
                  </span>
                  <span class="badge" :class="group.has_shared_key ? 'text-bg-dark' : 'text-bg-light'">
                    {{ group.has_shared_key ? t('aggregation_groups.key_set') : t('aggregation_groups.no_key') }}
                  </span>
                </td>
                <td>{{ formatTime(group.updated_at) }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-success" @click="handleRestore(group)">
                    <i class="bi bi-arrow-counterclockwise"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!deletedGroups.length">
                <td colspan="8" class="text-center text-muted py-4">{{ t('aggregation_groups.no_deleted_groups') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-if="metrics" class="card border-0 shadow-sm mt-4">
      <div class="card-header bg-white d-flex justify-content-between align-items-center">
        <div>
          <h6 class="mb-0">{{ t('aggregation_groups.metrics_title') }}</h6>
          <div class="small text-muted mt-1">{{ metrics.name }}</div>
        </div>
        <button class="btn btn-sm btn-outline-secondary" @click="metrics = null">
          <i class="bi bi-x-lg"></i>
        </button>
      </div>
      <div class="card-body">
        <div class="row g-3">
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.assigned_nodes') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.assigned_nodes }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.online_nodes') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.online_nodes }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.destination_pipelines') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.destination_pipelines }}</div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.source_pipelines') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.source_pipelines }}</div>
            </div>
          </div>
          <div class="col-md-4">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.avg_cpu') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.avg_cpu?.toFixed(1) || '0.0' }}%</div>
            </div>
          </div>
          <div class="col-md-4">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.avg_mem') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.avg_mem?.toFixed(1) || '0.0' }}%</div>
            </div>
          </div>
          <div class="col-md-4">
            <div class="border rounded-3 p-3 h-100">
              <div class="text-muted small mb-1">{{ t('aggregation_groups.tls_coverage') }}</div>
              <div class="fs-4 fw-bold">{{ metrics.tls_coverage_rate?.toFixed(0) || '0' }}%</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="aggregationGroupModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editingId ? t('aggregation_groups.edit_title') : t('aggregation_groups.create_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.name') }}</label>
                <input v-model="form.name" type="text" class="form-control">
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('common.alias') }}</label>
                <input v-model="form.alias" type="text" class="form-control">
              </div>
            </div>

            <div class="row g-3 mb-3">
              <div class="col-md-4">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="form.fluent_type" class="form-select">
                  <option value="fluentd">Fluentd</option>
                  <option value="fluentbit">Fluent Bit</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('common.mode') }}</label>
                <select v-model="form.mode" class="form-select">
                  <option value="forward">forward</option>
                  <option value="http">http</option>
                  <option value="custom">custom</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('common.cluster') }}</label>
                <select v-model="form.cluster_id" class="form-select">
                  <option :value="null">{{ t('common.none_assigned') }}</option>
                  <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
                    {{ cluster.alias || cluster.name }}
                  </option>
                </select>
              </div>
            </div>

            <div class="row g-3 mb-3">
              <div class="col-md-8">
                <label class="form-label">Endpoint Host</label>
                <input v-model="form.endpoint_host" type="text" class="form-control" placeholder="fluentd.internal.local">
              </div>
              <div class="col-md-4">
                <label class="form-label">Endpoint Port</label>
                <input v-model.number="form.endpoint_port" type="number" class="form-control" min="0" max="65535">
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <textarea v-model="form.description" class="form-control" rows="3"></textarea>
            </div>

            <div class="row g-3">
              <div class="col-md-8">
                <label class="form-label">
                  {{ editingId ? t('aggregation_groups.replace_shared_key') : t('aggregation_groups.shared_key') }}
                </label>
                <input
                  v-model="form.shared_key"
                  type="password"
                  class="form-control"
                  :placeholder="editingId ? t('aggregation_groups.shared_key_keep') : t('aggregation_groups.shared_key_optional')"
                >
              </div>
              <div class="col-md-4 d-flex flex-column justify-content-end">
                <div class="form-check mb-2">
                  <input id="aggEnableTLS" v-model="form.enable_tls" type="checkbox" class="form-check-input">
                  <label for="aggEnableTLS" class="form-check-label">{{ t('aggregation_groups.enable_tls') }}</label>
                </div>
                <div v-if="editingId" class="form-check">
                  <input id="aggClearKey" v-model="form.clear_shared_key" type="checkbox" class="form-check-input">
                  <label for="aggClearKey" class="form-check-label">{{ t('aggregation_groups.clear_shared_key') }}</label>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveGroup">
              {{ editingId ? t('save') : t('aggregation_groups.create') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from '../i18n'
import {
  createAggregationGroup,
  deleteAggregationGroup,
  getAggregationGroupMetrics,
  getAggregationGroups,
  getClusters,
  getDeletedAggregationGroups,
  restoreAggregationGroup,
  updateAggregationGroup,
} from '../api'

const activeTab = ref('active')
const groups = ref([])
const deletedGroups = ref([])
const clusters = ref([])
const editingId = ref(null)
const metrics = ref(null)
const { t, dateLocale } = useI18n()

const form = reactive({
  name: '',
  alias: '',
  description: '',
  fluent_type: 'fluentd',
  mode: 'forward',
  endpoint_host: '',
  endpoint_port: 24224,
  enable_tls: false,
  shared_key: '',
  clear_shared_key: false,
  cluster_id: null,
})

const tlsEnabledCount = computed(() => groups.value.filter((group) => group.enable_tls).length)
const sharedKeyCount = computed(() => groups.value.filter((group) => group.has_shared_key).length)

let modal = null

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function runtimeLabel(value) {
  return value === 'fluentbit' ? 'Fluent Bit' : 'Fluentd'
}

function endpointText(group) {
  if (!group.endpoint_host) return '-'
  return `${group.endpoint_host}:${group.endpoint_port || 0}`
}

function clusterName(clusterID) {
  if (!clusterID) return t('common.none_assigned')
  const cluster = clusters.value.find((item) => item.id === clusterID)
  return cluster ? (cluster.alias || cluster.name) : `#${clusterID}`
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function ensureModal() {
  if (!modal) {
    modal = new window.bootstrap.Modal(document.getElementById('aggregationGroupModal'))
  }
}

function resetForm() {
  editingId.value = null
  form.name = ''
  form.alias = ''
  form.description = ''
  form.fluent_type = 'fluentd'
  form.mode = 'forward'
  form.endpoint_host = ''
  form.endpoint_port = 24224
  form.enable_tls = false
  form.shared_key = ''
  form.clear_shared_key = false
  form.cluster_id = null
}

async function loadData() {
  const [groupsRes, deletedRes, clusterRes] = await Promise.all([
    getAggregationGroups(),
    getDeletedAggregationGroups(),
    getClusters(),
  ])

  groups.value = groupsRes || []
  deletedGroups.value = deletedRes || []
  clusters.value = clusterRes.data.data || []
}

function openCreate() {
  resetForm()
  ensureModal()
  modal.show()
}

function openEdit(group) {
  editingId.value = group.id
  form.name = group.name
  form.alias = group.alias || ''
  form.description = group.description || ''
  form.fluent_type = group.fluent_type || 'fluentd'
  form.mode = group.mode || 'forward'
  form.endpoint_host = group.endpoint_host || ''
  form.endpoint_port = group.endpoint_port || 24224
  form.enable_tls = !!group.enable_tls
  form.shared_key = ''
  form.clear_shared_key = false
  form.cluster_id = group.cluster_id || null
  ensureModal()
  modal.show()
}

function buildPayload() {
  const payload = {
    name: form.name,
    alias: form.alias,
    description: form.description,
    fluent_type: form.fluent_type,
    mode: form.mode,
    endpoint_host: form.endpoint_host,
    endpoint_port: Number(form.endpoint_port) || 0,
    enable_tls: form.enable_tls,
    cluster_id: form.cluster_id || null,
  }

  if (form.shared_key) {
    payload.shared_key = form.shared_key
  } else if (editingId.value && form.clear_shared_key) {
    payload.shared_key = ''
  }

  return payload
}

async function saveGroup() {
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateAggregationGroup(editingId.value, payload)
    } else {
      await createAggregationGroup(payload)
    }
    modal.hide()
    await loadData()
  } catch (error) {
    alert(`${editingId.value ? t('aggregation_groups.save_failed') : t('aggregation_groups.create_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDelete(group) {
  if (!confirm(t('aggregation_groups.confirm_delete').replace('{name}', group.alias || group.name))) return

  try {
    await deleteAggregationGroup(group.id)
    await loadData()
  } catch (error) {
    alert(`${t('aggregation_groups.delete_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleRestore(group) {
  if (!confirm(t('aggregation_groups.confirm_restore').replace('{name}', group.alias || group.name))) return

  try {
    await restoreAggregationGroup(group.id)
    activeTab.value = 'active'
    await loadData()
  } catch (error) {
    alert(`${t('aggregation_groups.restore_failed')}: ${getErrorMessage(error)}`)
  }
}

async function loadMetrics(group) {
  try {
    metrics.value = await getAggregationGroupMetrics(group.id)
  } catch (error) {
    alert(`${t('aggregation_groups.load_metrics_failed')}: ${getErrorMessage(error)}`)
  }
}

onMounted(loadData)
</script>

<style scoped>
.fm-agg-tabs .nav-link {
  border: 0;
  color: #475569;
  font-weight: 600;
  border-radius: 10px;
}

.fm-agg-tabs .nav-link.active {
  background: linear-gradient(135deg, #1d4ed8 0%, #2563eb 100%);
  color: #fff;
}
</style>
