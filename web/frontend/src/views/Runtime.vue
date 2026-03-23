<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('runtime_page.title') }}</h4>
        <div class="text-muted">{{ t('runtime_page.subtitle') }}</div>
      </div>
      <button class="btn btn-outline-primary" @click="loadData">
        <i class="bi bi-arrow-clockwise me-1"></i>{{ t('common.refresh') }}
      </button>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-3" v-for="item in statusCards" :key="item.key">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ item.label }}</div>
            <div class="fs-3 fw-bold" :class="item.className">{{ item.count }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-header bg-white">
        <h6 class="mb-0">{{ t('runtime_page.health_graph') }}</h6>
      </div>
      <div class="card-body">
        <FluentFlowGraph :graph="graph" />
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-header bg-white d-flex justify-content-between align-items-center">
        <h6 class="mb-0">{{ t('runtime_page.recommendations') }}</h6>
        <span class="badge text-bg-light">{{ recommendations.length }}</span>
      </div>
      <div class="card-body">
        <div v-if="recommendations.length" class="list-group list-group-flush">
          <div v-for="item in recommendations" :key="`${item.scope_type}-${item.scope_id}-${item.title}`" class="list-group-item px-0">
            <div class="d-flex align-items-center gap-2 mb-1">
              <span class="badge" :class="recommendationBadgeClass(item.severity)">{{ item.severity }}</span>
              <span class="fw-semibold">{{ item.title }}</span>
              <span class="text-muted small">{{ item.scope_type }} #{{ item.scope_id }}</span>
            </div>
            <div class="small text-muted mb-1">{{ item.detail }}</div>
            <div class="small">{{ item.suggestion }}</div>
          </div>
        </div>
        <div v-else class="text-center text-muted py-4">
          {{ t('runtime_page.no_recommendations') }}
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-header bg-white">
        <h6 class="mb-0">{{ t('runtime_page.drift_list') }}</h6>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.node') }}</th>
                <th>{{ t('common.cluster') }}</th>
                <th>{{ t('flow_graph.aggregation_group') }}</th>
                <th>{{ t('status') }}</th>
                <th>{{ t('runtime_page.desired_hash') }}</th>
                <th>{{ t('runtime_page.effective_hash') }}</th>
                <th>{{ t('runtime_page.last_sync') }}</th>
                <th>{{ t('runtime_page.error') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in driftItems" :key="item.node_id">
                <td>
                  <router-link :to="`/nodes/${item.node_id}`" class="text-decoration-none fw-semibold">
                    {{ item.hostname }}
                  </router-link>
                </td>
                <td>{{ item.cluster_name || '-' }}</td>
                <td>{{ item.aggregation_group || '-' }}</td>
                <td>
                  <span class="badge" :class="driftBadgeClass(item.status)">{{ item.status }}</span>
                </td>
                <td><code>{{ shortHash(item.desired_config_hash) }}</code></td>
                <td><code>{{ shortHash(item.effective_config_hash) }}</code></td>
                <td>{{ formatTime(item.last_sync_at) }}</td>
                <td class="small text-danger">{{ item.last_error || '-' }}</td>
              </tr>
              <tr v-if="!driftItems.length">
                <td colspan="8" class="text-center text-muted py-4">{{ t('runtime_page.no_runtime_data') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { getRuntimeDrift, getRuntimeHealthGraph, getRuntimeRecommendations } from '../api'
import FluentFlowGraph from '../components/FluentFlowGraph.vue'
import { useI18n } from '../i18n'

const graph = ref({ nodes: [], edges: [] })
const driftItems = ref([])
const recommendations = ref([])
const { t, dateLocale } = useI18n()

const statusCards = computed(() => {
  const counts = driftItems.value.reduce((acc, item) => {
    acc[item.status] = (acc[item.status] || 0) + 1
    return acc
  }, {})
  return [
    { key: 'in_sync', label: t('runtime_page.in_sync'), count: counts.in_sync || 0, className: 'text-success' },
    { key: 'drifted', label: t('runtime_page.drifted'), count: counts.drifted || 0, className: 'text-danger' },
    { key: 'apply_failed', label: t('runtime_page.apply_failed'), count: counts.apply_failed || 0, className: 'text-warning' },
    { key: 'unknown', label: t('runtime_page.unknown'), count: counts.unknown || 0, className: 'text-secondary' },
  ]
})

function shortHash(value) {
  if (!value) return '-'
  return value.length > 12 ? `${value.slice(0, 12)}...` : value
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function driftBadgeClass(status) {
  if (status === 'in_sync') return 'text-bg-success'
  if (status === 'drifted') return 'text-bg-danger'
  if (status === 'apply_failed') return 'text-bg-warning'
  if (status === 'unknown') return 'text-bg-secondary'
  return 'text-bg-light'
}

function recommendationBadgeClass(severity) {
  if (severity === 'high') return 'text-bg-danger'
  if (severity === 'medium') return 'text-bg-warning'
  return 'text-bg-info'
}

async function loadData() {
  const [graphRes, driftRes, recommendationRes] = await Promise.all([
    getRuntimeHealthGraph(),
    getRuntimeDrift(),
    getRuntimeRecommendations(),
  ])
  graph.value = graphRes || { nodes: [], edges: [] }
  driftItems.value = driftRes || []
  recommendations.value = recommendationRes || []
}

onMounted(loadData)
</script>
