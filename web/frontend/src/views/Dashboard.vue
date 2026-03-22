<template>
  <div>
    <h4 class="mb-4">仪表盘</h4>

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

    <div class="row g-4">
      <div class="col-md-6">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white">
            <h6 class="mb-0">最近部署</h6>
          </div>
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>状态</th>
                  <th>成功/失败/总数</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="d in deploys" :key="d.id">
                  <td>#{{ d.id }}</td>
                  <td><span :class="deployStatusClass(d.status)" class="badge">{{ d.status }}</span></td>
                  <td>{{ d.success_count }}/{{ d.fail_count }}/{{ d.total_nodes }}</td>
                  <td>{{ formatTime(d.created_at) }}</td>
                </tr>
                <tr v-if="!deploys.length">
                  <td colspan="4" class="text-center text-muted">暂无部署记录</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div class="col-md-6">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white">
            <h6 class="mb-0">最近审计日志</h6>
          </div>
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead>
                <tr>
                  <th>用户</th>
                  <th>操作</th>
                  <th>资源</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="log in auditLogs" :key="log.id">
                  <td>{{ log.username }}</td>
                  <td>{{ log.action }}</td>
                  <td class="text-truncate" style="max-width: 200px;">{{ log.resource }}</td>
                  <td>{{ formatTime(log.created_at) }}</td>
                </tr>
                <tr v-if="!auditLogs.length">
                  <td colspan="4" class="text-center text-muted">暂无日志</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getNodeStats, getDeploys, getAuditLogs } from '../api'

const stats = ref({})
const deploys = ref([])
const auditLogs = ref([])

function getStatusCount(status) {
  const s = stats.value.statuses?.find(s => s.status === status)
  return s?.count || 0
}

function deployStatusClass(status) {
  return {
    'bg-success': status === 'completed',
    'bg-warning': status === 'running',
    'bg-danger': status === 'failed',
    'bg-secondary': status === 'pending',
  }
}

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(async () => {
  try {
    const [s, d, a] = await Promise.all([
      getNodeStats(),
      getDeploys({ page_size: 5 }),
      getAuditLogs({ page_size: 10 }),
    ])
    stats.value = s.data
    deploys.value = d.data.data || []
    auditLogs.value = a.data.data || []
  } catch (e) {
    console.error('Dashboard load error:', e)
  }
})
</script>
