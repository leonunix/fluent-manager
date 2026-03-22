<template>
  <div>
    <h4 class="mb-4">仪表盘</h4>

    <!-- KPI Cards -->
    <div class="row g-4 mb-4">
      <div class="col-md-3">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <div class="text-muted small">节点总数</div>
                <h3>{{ stats.total || 0 }}</h3>
              </div>
              <i class="bi bi-hdd-network text-primary" style="font-size: 2rem;"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <div class="text-muted small">在线节点</div>
                <h3 class="text-success">{{ getStatusCount('online') }}</h3>
              </div>
              <i class="bi bi-check-circle text-success" style="font-size: 2rem;"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <div class="text-muted small">离线节点</div>
                <h3 class="text-warning">{{ getStatusCount('offline') }}</h3>
              </div>
              <i class="bi bi-x-circle text-warning" style="font-size: 2rem;"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <div class="text-muted small">异常节点</div>
                <h3 class="text-danger">{{ getStatusCount('error') }}</h3>
              </div>
              <i class="bi bi-exclamation-triangle text-danger" style="font-size: 2rem;"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">节点状态分布</h6></div>
          <div class="card-body">
            <v-chart :option="statusPieOption" style="height: 240px;" :autoresize="true" />
          </div>
        </div>
      </div>
      <div class="col-md-8">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">数据中心节点概览</h6></div>
          <div class="card-body">
            <v-chart v-if="dcBarOption.series" :option="dcBarOption" style="height: 240px;" :autoresize="true" />
            <div v-else class="text-center text-muted py-5">暂无数据中心</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Topology mini + tables -->
    <div class="row g-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">拓扑概览</h6>
            <router-link to="/topology" class="btn btn-sm btn-outline-primary">查看详情</router-link>
          </div>
          <div class="card-body p-0">
            <div v-for="dc in tree" :key="dc.id" class="border-bottom px-3 py-2">
              <div class="d-flex align-items-center mb-1">
                <i class="bi bi-building text-primary me-2"></i>
                <strong class="small">{{ dc.alias || dc.name }}</strong>
                <span class="badge bg-secondary ms-auto small">{{ dc.provider }}</span>
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
            <div v-if="!tree.length" class="text-center text-muted py-3 small">暂无拓扑数据</div>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">最近部署</h6></div>
          <div class="card-body p-0">
            <table class="table table-hover table-sm mb-0">
              <thead><tr><th>ID</th><th>状态</th><th>成功/失败/总</th><th>时间</th></tr></thead>
              <tbody>
                <tr v-for="d in deploys" :key="d.id">
                  <td>#{{ d.id }}</td>
                  <td><span :class="deployStatusClass(d.status)" class="badge">{{ d.status }}</span></td>
                  <td>{{ d.success_count }}/{{ d.fail_count }}/{{ d.total_nodes }}</td>
                  <td class="small">{{ formatTime(d.created_at) }}</td>
                </tr>
                <tr v-if="!deploys.length"><td colspan="4" class="text-center text-muted">暂无</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">最近审计日志</h6></div>
          <div class="card-body p-0">
            <table class="table table-hover table-sm mb-0">
              <thead><tr><th>用户</th><th>操作</th><th>资源</th><th>时间</th></tr></thead>
              <tbody>
                <tr v-for="log in auditLogs" :key="log.id">
                  <td>{{ log.username }}</td>
                  <td>{{ log.action }}</td>
                  <td class="text-truncate small" style="max-width: 120px;">{{ log.resource }}</td>
                  <td class="small">{{ formatTime(log.created_at) }}</td>
                </tr>
                <tr v-if="!auditLogs.length"><td colspan="4" class="text-center text-muted">暂无</td></tr>
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

use([CanvasRenderer, PieChart, BarChart, TooltipComponent, LegendComponent, GridComponent])

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
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      label: { show: false },
      data: [
        { value: online, name: '在线', itemStyle: { color: '#198754' } },
        { value: offline, name: '离线', itemStyle: { color: '#ffc107' } },
        { value: error, name: '异常', itemStyle: { color: '#dc3545' } },
      ].filter(d => d.value > 0),
      emphasis: { itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0,0,0,0.2)' } },
    }],
  }
})

const dcBarOption = computed(() => {
  if (!tree.value.length) return {}
  const dcNames = []
  const onlineCounts = []
  const offlineCounts = []

  for (const dc of tree.value) {
    let online = 0, total = 0
    for (const r of (dc.regions || [])) {
      for (const cl of (r.clusters || [])) {
        online += cl.online_count || 0
        total += cl.node_count || 0
      }
    }
    dcNames.push(dc.alias || dc.name)
    onlineCounts.push(online)
    offlineCounts.push(total - online)
  }

  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['在线', '离线/异常'], bottom: 0, textStyle: { fontSize: 11 } },
    grid: { left: '3%', right: '4%', bottom: '15%', top: '5%', containLabel: true },
    xAxis: { type: 'category', data: dcNames, axisLabel: { fontSize: 11 } },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      { name: '在线', type: 'bar', stack: 'total', data: onlineCounts, itemStyle: { color: '#198754' } },
      { name: '离线/异常', type: 'bar', stack: 'total', data: offlineCounts, itemStyle: { color: '#ffc107' } },
    ],
  }
})

function deployStatusClass(status) {
  return { 'bg-success': status === 'completed', 'bg-warning': status === 'running', 'bg-danger': status === 'failed', 'bg-secondary': status === 'pending' }
}
function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }

onMounted(async () => {
  try {
    const [s, d, a, t] = await Promise.all([
      getNodeStats(),
      getDeploys({ page_size: 5 }),
      getAuditLogs({ page_size: 8 }),
      getTopologyTree(),
    ])
    stats.value = s.data
    deploys.value = d.data.data || []
    auditLogs.value = a.data.data || []
    tree.value = t.data.data || []
  } catch (e) {
    console.error('Dashboard load error:', e)
  }
})
</script>
