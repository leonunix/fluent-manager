<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('pipelines.title') }}</h4>
        <div class="text-muted">{{ t('pipelines.subtitle') }}</div>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('pipelines.create') }}
      </button>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('pipelines.total') }}</div>
            <div class="fs-3 fw-bold">{{ pipelines.length }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('pipelines.aggregation_targets') }}</div>
            <div class="fs-3 fw-bold">{{ aggregationTargetCount }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('pipelines.output_targets') }}</div>
            <div class="fs-3 fw-bold">{{ outputTargetCount }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body p-2">
        <div class="nav nav-pills">
          <button class="nav-link" :class="{ active: viewMode === 'graph' }" @click="viewMode = 'graph'">{{ t('pipelines.graph') }}</button>
          <button class="nav-link" :class="{ active: viewMode === 'table' }" @click="viewMode = 'table'">{{ t('pipelines.table') }}</button>
        </div>
      </div>
    </div>

    <div v-if="viewMode === 'graph'" class="card border-0 shadow-sm mb-4">
      <div class="card-header bg-white d-flex justify-content-between align-items-center">
        <h6 class="mb-0">{{ t('pipelines.flow_graph') }}</h6>
        <button class="btn btn-sm btn-outline-primary" @click="loadData">
          <i class="bi bi-arrow-clockwise"></i>
        </button>
      </div>
      <div class="card-body">
        <FluentFlowGraph :graph="graph" />
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('common.source') }}</th>
                <th>{{ t('common.target') }}</th>
                <th>{{ t('common.protocol') }}</th>
                <th>{{ t('pipelines.tag_strategy') }}</th>
                <th>{{ t('status') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="pipeline in pipelines" :key="pipeline.id">
                <td>
                  <div class="fw-semibold">{{ pipeline.name }}</div>
                  <div class="small text-muted">{{ pipeline.description || t('common.no_description') }}</div>
                </td>
                <td>{{ pipelineSourceLabel(pipeline) }}</td>
                <td>{{ pipelineTargetLabel(pipeline) }}</td>
                <td><span class="badge text-bg-secondary">{{ pipeline.protocol }}</span></td>
                <td><code>{{ pipeline.tag_strategy || '-' }}</code></td>
                <td>
                  <span :class="pipeline.enabled ? 'badge text-bg-success' : 'badge text-bg-light'">
                    {{ pipeline.enabled ? t('common.enabled') : t('common.disabled') }}
                  </span>
                </td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(pipeline)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="handleDelete(pipeline)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!pipelines.length">
                <td colspan="7" class="text-center text-muted py-4">{{ t('pipelines.no_pipelines') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal fade" id="pipelineModal" tabindex="-1">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editingId ? t('pipelines.edit_title') : t('pipelines.create_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3 mb-3">
              <div class="col-md-4">
                <label class="form-label">{{ t('common.name') }}</label>
                <input v-model="form.name" type="text" class="form-control">
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="form.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                  <option value="shared">Shared</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('common.protocol') }}</label>
                <select v-model="form.protocol" class="form-select">
                  <option value="forward">forward</option>
                  <option value="http">http</option>
                  <option value="kafka">kafka</option>
                  <option value="loki">loki</option>
                  <option value="custom">custom</option>
                </select>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>

            <div class="row g-4 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('pipelines.source_type') }}</label>
                <select v-model="form.source_type" class="form-select mb-2">
                  <option value="cluster">{{ t('common.cluster') }}</option>
                  <option value="aggregation_group">{{ t('flow_graph.aggregation_group') }}</option>
                  <option value="selector">{{ t('pipelines.selector') }}</option>
                </select>
                <select v-if="form.source_type === 'cluster'" v-model="form.source_cluster_id" class="form-select">
                  <option :value="null">{{ t('pipelines.select_cluster') }}</option>
                  <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">{{ cluster.alias || cluster.name }}</option>
                </select>
                <select v-else-if="form.source_type === 'aggregation_group'" v-model="form.source_aggregation_group_id" class="form-select">
                  <option :value="null">{{ t('pipelines.select_group') }}</option>
                  <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.alias || group.name }}</option>
                </select>
                <textarea
                  v-else
                  v-model="form.source_label_selector"
                  class="form-control font-monospace"
                  rows="4"
                  placeholder='{"app":"nginx","tier":"edge"}'
                ></textarea>
              </div>

              <div class="col-md-6">
                <label class="form-label">{{ t('pipelines.destination_type') }}</label>
                <select v-model="form.destination_type" class="form-select mb-2">
                  <option value="aggregation_group">{{ t('flow_graph.aggregation_group') }}</option>
                  <option value="output">{{ t('common.output') }}</option>
                </select>
                <select v-if="form.destination_type === 'aggregation_group'" v-model="form.destination_aggregation_group_id" class="form-select">
                  <option :value="null">{{ t('pipelines.select_group') }}</option>
                  <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.alias || group.name }}</option>
                </select>
                <div v-else>
                  <select v-model="form.destination_output_target_id" class="form-select">
                    <option :value="null">{{ t('pipelines.select_output_target') }}</option>
                    <option v-for="target in availableOutputTargets" :key="target.id" :value="target.id">
                      {{ target.name }} · {{ target.target_type }}
                    </option>
                  </select>
                  <div class="small text-muted mt-2">{{ t('pipelines.output_target_hint') }}</div>
                  <div v-if="selectedOutputTarget" class="border rounded-3 p-3 bg-light-subtle mt-3">
                    <div class="d-flex flex-wrap justify-content-between align-items-start gap-2 mb-2">
                      <div>
                        <div class="fw-semibold">{{ selectedOutputTarget.name }}</div>
                        <div class="small text-muted">{{ selectedOutputTarget.description || t('common.no_description') }}</div>
                      </div>
                      <span class="badge text-bg-secondary">{{ selectedOutputTarget.target_type }}</span>
                    </div>
                    <div class="small text-muted mb-2">{{ t('pipelines.selected_output_summary') }}</div>
                    <div class="mb-2"><code>{{ selectedOutputSummary.endpoint }}</code></div>
                    <div class="d-flex flex-wrap gap-2">
                      <span
                        v-for="chip in selectedOutputSummary.chips"
                        :key="chip"
                        class="badge rounded-pill text-bg-light"
                      >
                        {{ chip }}
                      </span>
                      <span v-if="!selectedOutputSummary.chips.length" class="text-muted small">{{ t('pipelines.no_output_summary') }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="row g-3">
              <div class="col-md-5">
                <label class="form-label">{{ t('pipelines.upstream_role') }}</label>
                <select v-model="form.upstream_role" class="form-select">
                  <option value="">{{ t('common.unspecified') }}</option>
                  <option value="edge_collector">edge_collector</option>
                  <option value="aggregator">aggregator</option>
                  <option value="gateway">gateway</option>
                  <option value="standalone">standalone</option>
                </select>
              </div>
              <div class="col-md-5">
                <label class="form-label">{{ t('pipelines.tag_strategy') }}</label>
                <input v-model="form.tag_strategy" type="text" class="form-control" placeholder="cluster.hostname / k8s.namespace.tag">
              </div>
              <div class="col-md-2 d-flex align-items-end">
                <div class="form-check mb-2">
                  <input id="pipelineEnabled" v-model="form.enabled" type="checkbox" class="form-check-input">
                  <label for="pipelineEnabled" class="form-check-label">{{ t('pipelines.enable') }}</label>
                </div>
              </div>
            </div>

            <div v-if="pipelineGuidance.items.length" class="alert mt-3 mb-0" :class="pipelineGuidance.alertClass">
              <div class="fw-semibold mb-2">{{ t('pipelines.guidance_title') }}</div>
              <div v-for="item in pipelineGuidance.items" :key="item.message" class="small mb-1">
                {{ item.message }}
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="savePipeline">
              {{ editingId ? t('save') : t('pipelines.create') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { createPipeline, deletePipeline, getAggregationGroups, getClusters, getOutputTargets, getPipelineGraph, getPipelines, updatePipeline } from '../api'
import FluentFlowGraph from '../components/FluentFlowGraph.vue'
import { useI18n } from '../i18n'
import { summarizeOutputTarget } from '../utils/output_targets'

const viewMode = ref('graph')
const pipelines = ref([])
const graph = ref({ nodes: [], edges: [] })
const clusters = ref([])
const groups = ref([])
const outputTargets = ref([])
const editingId = ref(null)
const { t } = useI18n()

const form = reactive({
  name: '',
  description: '',
  fluent_type: 'fluentbit',
  protocol: 'forward',
  source_type: 'cluster',
  source_cluster_id: null,
  source_aggregation_group_id: null,
  source_label_selector: '',
  upstream_role: '',
  destination_type: 'aggregation_group',
  destination_aggregation_group_id: null,
  destination_output_target_id: null,
  tag_strategy: '',
  enabled: true,
})

const aggregationTargetCount = computed(() => pipelines.value.filter((item) => item.destination_aggregation_group_id).length)
const outputTargetCount = computed(() => pipelines.value.filter((item) => item.destination_output_target_id || item.destination_output_name).length)
const availableOutputTargets = computed(() =>
  outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === form.fluent_type)
)
const selectedOutputTarget = computed(() =>
  availableOutputTargets.value.find((item) => item.id === form.destination_output_target_id) || null
)
const selectedOutputSummary = computed(() => summarizeOutputTarget(selectedOutputTarget.value))
const pipelineGuidance = computed(() => {
  const items = []
  let level = 'info'

  if (form.destination_type === 'aggregation_group') {
    if (form.protocol !== 'forward') {
      items.push({ level: 'warning', message: t('pipelines.guidance_group_forward') })
      level = 'warning'
    } else {
      items.push({ level: 'success', message: t('pipelines.guidance_group_forward_ok') })
      if (level !== 'warning') level = 'success'
    }
  }

  if (form.destination_type === 'output') {
    if (!selectedOutputTarget.value) {
      items.push({ level: 'info', message: t('pipelines.guidance_pick_target') })
    } else {
      const targetType = selectedOutputTarget.value.target_type
      if (targetType === 'kafka' && form.protocol !== 'kafka') {
        items.push({ level: 'warning', message: t('pipelines.guidance_kafka_protocol') })
        level = 'warning'
      }
      if (targetType === 'opensearch' && !['http', 'custom'].includes(form.protocol)) {
        items.push({ level: 'warning', message: t('pipelines.guidance_opensearch_protocol') })
        level = 'warning'
      }
      if (targetType === 'loki' && !['loki', 'http', 'custom'].includes(form.protocol)) {
        items.push({ level: 'warning', message: t('pipelines.guidance_loki_protocol') })
        level = 'warning'
      }
      if (targetType === 'http' && !['http', 'custom'].includes(form.protocol)) {
        items.push({ level: 'warning', message: t('pipelines.guidance_http_protocol') })
        level = 'warning'
      }
      if (targetType === 's3' && form.protocol === 'forward') {
        items.push({ level: 'warning', message: t('pipelines.guidance_s3_protocol') })
        level = 'warning'
      }
      if (targetType === 'stdout') {
        items.push({ level: 'info', message: t('pipelines.guidance_stdout_protocol') })
      }
      if (form.protocol === 'forward' && targetType !== 'stdout') {
        items.push({ level: 'info', message: t('pipelines.guidance_forward_terminal') })
      }
    }
  }

  return {
    items,
    alertClass: level === 'warning'
      ? 'alert-warning'
      : level === 'success'
        ? 'alert-success'
        : 'alert-info',
  }
})

let modal = null

function ensureModal() {
  if (!modal) {
    modal = new window.bootstrap.Modal(document.getElementById('pipelineModal'))
  }
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function resetForm() {
  editingId.value = null
  form.name = ''
  form.description = ''
  form.fluent_type = 'fluentbit'
  form.protocol = 'forward'
  form.source_type = 'cluster'
  form.source_cluster_id = null
  form.source_aggregation_group_id = null
  form.source_label_selector = ''
  form.upstream_role = ''
  form.destination_type = 'aggregation_group'
  form.destination_aggregation_group_id = null
  form.destination_output_target_id = null
  form.tag_strategy = ''
  form.enabled = true
}

function pipelineSourceLabel(pipeline) {
  if (pipeline.source_cluster) return pipeline.source_cluster.alias || pipeline.source_cluster.name
  if (pipeline.source_aggregation_group) return pipeline.source_aggregation_group.alias || pipeline.source_aggregation_group.name
  return pipeline.source_label_selector || '-'
}

function pipelineTargetLabel(pipeline) {
  if (pipeline.destination_aggregation_group) return pipeline.destination_aggregation_group.alias || pipeline.destination_aggregation_group.name
  if (pipeline.destination_output_target) return pipeline.destination_output_target.name
  return pipeline.destination_output_name || '-'
}

function openCreate() {
  resetForm()
  ensureModal()
  modal.show()
}

function openEdit(pipeline) {
  editingId.value = pipeline.id
  form.name = pipeline.name
  form.description = pipeline.description || ''
  form.fluent_type = pipeline.fluent_type
  form.protocol = pipeline.protocol
  form.source_type = pipeline.source_cluster_id ? 'cluster' : (pipeline.source_aggregation_group_id ? 'aggregation_group' : 'selector')
  form.source_cluster_id = pipeline.source_cluster_id || null
  form.source_aggregation_group_id = pipeline.source_aggregation_group_id || null
  form.source_label_selector = pipeline.source_label_selector || ''
  form.upstream_role = pipeline.upstream_role || ''
  form.destination_type = pipeline.destination_aggregation_group_id ? 'aggregation_group' : 'output'
  form.destination_aggregation_group_id = pipeline.destination_aggregation_group_id || null
  form.destination_output_target_id = pipeline.destination_output_target_id || null
  form.tag_strategy = pipeline.tag_strategy || ''
  form.enabled = !!pipeline.enabled
  ensureModal()
  modal.show()
}

function buildPayload() {
  const payload = {
    name: form.name,
    description: form.description,
    fluent_type: form.fluent_type,
    protocol: form.protocol,
    source_cluster_id: null,
    source_aggregation_group_id: null,
    source_label_selector: '',
    upstream_role: form.upstream_role,
    destination_aggregation_group_id: null,
    destination_output_target_id: null,
    destination_output_name: '',
    destination_output_type: '',
    tag_strategy: form.tag_strategy,
    enabled: form.enabled,
  }

  if (form.source_type === 'cluster') payload.source_cluster_id = form.source_cluster_id
  if (form.source_type === 'aggregation_group') payload.source_aggregation_group_id = form.source_aggregation_group_id
  if (form.source_type === 'selector') payload.source_label_selector = form.source_label_selector

  if (form.destination_type === 'aggregation_group') {
    payload.destination_aggregation_group_id = form.destination_aggregation_group_id
  } else {
    payload.destination_output_target_id = form.destination_output_target_id
  }

  return payload
}

async function savePipeline() {
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updatePipeline(editingId.value, payload)
    } else {
      await createPipeline(payload)
    }
    modal.hide()
    await loadData()
  } catch (error) {
    alert(`${editingId.value ? t('pipelines.save_failed') : t('pipelines.create_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDelete(pipeline) {
  if (!confirm(t('pipelines.confirm_delete').replace('{name}', pipeline.name))) return
  try {
    await deletePipeline(pipeline.id)
    await loadData()
  } catch (error) {
    alert(`${t('pipelines.delete_failed')}: ${getErrorMessage(error)}`)
  }
}

async function loadData() {
  const [pipelinesRes, graphRes, clustersRes, groupsRes, outputTargetsRes] = await Promise.all([
    getPipelines(),
    getPipelineGraph(),
    getClusters(),
    getAggregationGroups(),
    getOutputTargets(),
  ])
  pipelines.value = pipelinesRes || []
  graph.value = graphRes || { nodes: [], edges: [] }
  clusters.value = clustersRes.data.data || []
  groups.value = groupsRes || []
  outputTargets.value = outputTargetsRes || []
}

watch(
  () => form.fluent_type,
  () => {
    if (form.destination_type !== 'output') return
    const availableIds = new Set(availableOutputTargets.value.map((item) => item.id))
    if (!availableIds.has(form.destination_output_target_id)) {
      form.destination_output_target_id = null
    }
  }
)

onMounted(loadData)
</script>
