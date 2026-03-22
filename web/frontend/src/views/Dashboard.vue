<template>
  <div>
    <h4 class="mb-4">{{ t('dashboard.title') }}</h4>

    <!-- KPI Cards -->
    <div class="row g-4 mb-4">
      <div class="col-md-3">
        <div class="stat-card">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <div class="text-muted small">{{ t('dashboard.total_nodes') }}</div>
              <h3 class="mb-0 mt-1">{{ stats.total || 0 }}</h3>
            </div>
            <div class="stat-icon" style="background: var(--fm-primary-100); color: var(--fm-primary);">
              <i class="bi bi-hdd-network"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="stat-card">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <div class="text-muted small">{{ t('dashboard.online_nodes') }}</div>
              <h3 class="mb-0 mt-1 text-success">{{ getStatusCount('online') }}</h3>
            </div>
            <div class="stat-icon" style="background: #d1fae5; color: var(--fm-success);">
              <i class="bi bi-check-circle"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="stat-card">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <div class="text-muted small">{{ t('dashboard.offline_nodes') }}</div>
              <h3 class="mb-0 mt-1 text-warning">{{ getStatusCount('offline') }}</h3>
            </div>
            <div class="stat-icon" style="background: #fef3c7; color: var(--fm-warning);">
              <i class="bi bi-x-circle"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="stat-card">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <div class="text-muted small">{{ t('dashboard.error_nodes') }}</div>
              <h3 class="mb-0 mt-1 text-danger">{{ getStatusCount('error') }}</h3>
            </div>
            <div class="stat-icon" style="background: #fee2e2; color: var(--fm-danger);">
              <i class="bi bi-exclamation-triangle"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0">
          <div class="card-header"><h6 class="mb-0">{{ t('dashboard.status_dist') }}</h6></div>
          <div class="card-body">
            <v-chart :option="statusPieOption" style="height: 240px;" :autoresize="true" />
          </div>
        </div>
      </div>
      <div class="col-md-8">
        <div class="card border-0">
          <div class="card-header"><h6 class="mb-0">{{ t('dashboard.dc_overview') }}</h6></div>
          <div class="card-body">
            <v-chart v-if="dcBarOption.series" :option="dcBarOption" style="height: 240px;" :autoresize="true" />
            <div v-else class="text-center text-muted py-5">{{ t('dashboard.no_dc') }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom row -->
    <div class="row g-4">
      <div class="col-md-4">
        <div class="card border-0">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h6 class="mb-0">{{ t('dashboard.topo_overview') }}</h6>
            <router-link to="/topology" class="btn btn-sm btn-outline-primary">{{ t('dashboard.view_detail') }}</router-link>
          </div>
          <div class="card-body p-0">
            <div v-for="dc in tree" :key="dc.id" class="border-bottom px-3 py-2">
              <div class="d-flex align-items-center mb-1">
                <i class="bi bi-building text-primary me-2"></i>
                <strong class="small">{{ dc.alias || dc.name }}</strong>
                <span class="badge bg-secondary ms-auto">{{ dc.provider }}</span>
              </div>
              <div v-for="r in dc.regions" :key="r.id" class="ps-3">
                <div class="small text-muted"><i class="bi bi-globe2 me-1"></i>{{ r.alias || r.name }}</div>
                <div v-for="cl in r.clusters" :key="cl.id" class="ps-3 d-flex align-items-center small">
                  <i class="bi bi-diagram-3 text-success me-1"></i>
                  {{ cl.alias || cl.name }}
                  <span v-if="cl.environment" class="badge ms-1" style="font-size:0.65em" :style="{backgroundColor: cl.env_color}">{{ cl.environment }}</span>
                  <span class="ms-auto text-muted">{{ cl.online_count }}/{{ cl.node_count }}</span>
                </div>
              </div>
            </div>
            <div v-if="!tree.length" class="text-center text-muted py-3 small">{{ t('dashboard.no_dc') }}</div>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0">
          <div class="card-header"><h6 class="mb-0">{{ t('dashboard.recent_deploys') }}</h6></div>
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead><tr><th>{{ t('dashboard.deploy_id') }}</th><th>{{ t('dashboard.deploy_status') }}</th><th>{{ t('dashboard.deploy_result') }}</th><th>{{ t('dashboard.time') }}</th></tr></thead>
              <tbody>
                <tr v-for="d in deploys" :key="d.id">
                  <td>#{{ d.id }}</td>
                  <td><span :class="deployStatusClass(d.status)" class="badge">{{ d.status }}</span></td>
                  <td>{{ d.success_count }}/{{ d.fail_count }}/{{ d.total_nodes }}</td>
                  <td class="small">{{ formatTime(d.created_at) }}</td>
                </tr>
                <tr v-if="!deploys.length"><td colspan="4" class="text-center text-muted">{{ t('dashboard.no_deploy') }}</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0">
          <div class="card-header"><h6 class="mb-0">{{ t('dashboard.recent_audit') }}</h6></div>
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead><tr><th>{{ t('dashboard.user') }}</th><th>{{ t('dashboard.action') }}</th><th>{{ t('dashboard.resource') }}</th><th>{{ t('dashboard.time') }}</th></tr></thead>
              <tbody>
                <tr v-for="log in auditLogs" :key="log.id">
                  <td>{{ log.username }}</td>
                  <td>{{ log.action }}</td>
                  <td class="text-truncate small" style="max-width: 120px;">{{ log.resource }}</td>
                  <td class="small">{{ formatTime(log.created_at) }}</td>
                </tr>
                <tr v-if="!auditLogs.length"><td colspan="4" class="text-center text-muted">{{ t('dashboard.no_log') }}</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { PieChart, BarChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { getNodeStats, getDeploys, getAuditLogs, getTopologyTree } from '../api'
import { useI18n } from '../i18n'

use([CanvasRenderer, PieChart, BarChart, TooltipComponent, LegendComponent, GridComponent])

const { t } = useI18n()
const stats = ref({})
const deploys = ref([])
const auditLogs = ref([])
const tree = ref([])

function getStatusCount(status) {
  const s = stats.value.statuses?.find(s => s.status === status)
  return s?.count || 0
}

const statusPieOption = computed(() => {
  const online = getStatusCount('online')
  const offline = getStatusCount('offline')
  const error = getStatusCount('error')
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { fontSize: 11 } },
    series: [{
      type: 'pie', radius: ['40%', '70%'], center: ['50%', '45%'],
      avoidLabelOverlap: true, label: { show: false },
      data: [
        { value: online, name: t('dashboard.online'), itemStyle: { color: '#10b981' } },
        { value: offline, name: t('dashboard.offline'), itemStyle: { color: '#f59e0b' } },
        { value: error, name: t('dashboard.error'), itemStyle: { color: '#ef4444' } },
      ].filter(d => d.value > 0),
      emphasis: { itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0,0,0,0.2)' } },
    }],
  }
})

const dcBarOption = computed(() => {
  if (!tree.value.length) return {}
  const dcNames = [], onlineCounts = [], offlineCounts = []
  for (const dc of tree.value) {
    let online = 0, total = 0
    for (const r of (dc.regions || [])) for (const cl of (r.clusters || [])) { online += cl.online_count || 0; total += cl.node_count || 0 }
    dcNames.push(dc.alias || dc.name)
    onlineCounts.push(online)
    offlineCounts.push(total - online)
  }
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: [t('dashboard.online'), t('dashboard.offline_error')], bottom: 0, textStyle: { fontSize: 11 } },
    grid: { left: '3%', right: '4%', bottom: '15%', top: '5%', containLabel: true },
    xAxis: { type: 'category', data: dcNames, axisLabel: { fontSize: 11 } },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      { name: t('dashboard.online'), type: 'bar', stack: 'total', data: onlineCounts, itemStyle: { color: '#10b981', borderRadius: [4, 4, 0, 0] } },
      { name: t('dashboard.offline_error'), type: 'bar', stack: 'total', data: offlineCounts, itemStyle: { color: '#f59e0b', borderRadius: [4, 4, 0, 0] } },
    ],
  }
})

function deployStatusClass(status) {
  return { 'bg-success': status === 'completed', 'bg-warning': status === 'running', 'bg-danger': status === 'failed', 'bg-secondary': status === 'pending' }
}
function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }

onMounted(async () => {
  try {
    const [s, d, a, tr] = await Promise.all([getNodeStats(), getDeploys({ page_size: 5 }), getAuditLogs({ page_size: 8 }), getTopologyTree()])
    stats.value = s.data
    deploys.value = d.data.data || []
    auditLogs.value = a.data.data || []
    tree.value = tr.data.data || []
  } catch (e) { console.error('Dashboard load error:', e) }
})
</script>
