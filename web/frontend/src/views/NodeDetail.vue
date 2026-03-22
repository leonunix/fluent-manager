<template>
  <div v-if="node">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <router-link to="/nodes" class="text-decoration-none">&larr; 返回节点列表</router-link>
        <h4 class="mt-2 mb-0">{{ node.hostname }}</h4>
        <span class="text-muted small">{{ node.node_uid }}</span>
        <span :class="statusClass(node.status)" class="badge ms-2">{{ statusText(node.status) }}</span>
        <span class="badge bg-info ms-1">{{ node.fluent_type }}</span>
      </div>
      <div class="btn-group">
        <button class="btn btn-outline-success btn-sm" @click="sendCmd('restart')">
          <i class="bi bi-arrow-clockwise"></i> 重启服务
        </button>
        <button class="btn btn-outline-primary btn-sm" @click="sendCmd('reload')">
          <i class="bi bi-arrow-repeat"></i> 热重载
        </button>
        <button class="btn btn-outline-warning btn-sm" @click="sendCmd('status')">
          <i class="bi bi-info-circle"></i> 查看状态
        </button>
        <button class="btn btn-outline-secondary btn-sm" @click="sendCmd('validate')">
          <i class="bi bi-check2-circle"></i> 验证配置
        </button>
        <button class="btn btn-outline-danger btn-sm" @click="sendCmd('rollback')">
          <i class="bi bi-arrow-counterclockwise"></i> 回滚
        </button>
      </div>
    </div>

    <!-- Node Info + Metrics -->
    <div class="row g-4 mb-4">
      <div class="col-md-3">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <h6 class="text-muted">节点信息</h6>
            <table class="table table-sm table-borderless mb-0">
              <tr><td class="text-muted">IP</td><td>{{ node.ip_address || '-' }}</td></tr>
              <tr><td class="text-muted">系统</td><td>{{ node.os }}</td></tr>
              <tr><td class="text-muted">Agent</td><td>{{ node.agent_version }}</td></tr>
              <tr><td class="text-muted">Fluent</td><td>{{ node.fluent_version }}</td></tr>
              <tr><td class="text-muted">分组</td><td>{{ node.group?.name || '-' }}</td></tr>
              <tr><td class="text-muted">心跳</td><td>{{ formatTime(node.last_heartbeat) }}</td></tr>
            </table>
          </div>
        </div>
      </div>
      <div class="col-md-9" v-if="metrics">
        <div class="row g-3">
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">CPU 使用率</div>
                <h3 :class="colorClass(metrics.cpu_usage_percent, 80, 90)">{{ metrics.cpu_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.cpu_usage_percent, 80, 90)"
                    :style="{ width: metrics.cpu_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">Load: {{ metrics.load_avg_1?.toFixed(2) }} / {{ metrics.load_avg_5?.toFixed(2) }} / {{ metrics.load_avg_15?.toFixed(2) }}</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">内存使用</div>
                <h3 :class="colorClass(metrics.mem_usage_percent, 80, 90)">{{ metrics.mem_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.mem_usage_percent, 80, 90)"
                    :style="{ width: metrics.mem_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">{{ metrics.mem_used_mb }} / {{ metrics.mem_total_mb }} MB</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">磁盘使用</div>
                <h3 :class="colorClass(metrics.disk_usage_percent, 80, 90)">{{ metrics.disk_usage_percent?.toFixed(1) }}%</h3>
                <div class="progress" style="height: 6px;">
                  <div class="progress-bar" :class="barClass(metrics.disk_usage_percent, 80, 90)"
                    :style="{ width: metrics.disk_usage_percent + '%' }"></div>
                </div>
                <div class="text-muted small mt-1">{{ metrics.disk_used_gb }} / {{ metrics.disk_total_gb }} GB</div>
              </div>
            </div>
          </div>
          <div class="col-md-3">
            <div class="card border-0 shadow-sm">
              <div class="card-body text-center">
                <div class="text-muted small">Fluent 进程</div>
                <h3 :class="metrics.fluent_running ? 'text-success' : 'text-danger'">
                  {{ metrics.fluent_running ? '运行中' : '已停止' }}
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

    <!-- Commands and Logs tabs -->
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <a class="nav-link" :class="{ active: tab === 'commands' }" href="#" @click.prevent="tab = 'commands'">
          <i class="bi bi-terminal me-1"></i>命令历史
        </a>
      </li>
      <li class="nav-item">
        <a class="nav-link" :class="{ active: tab === 'logs' }" href="#" @click.prevent="tab = 'logs'; loadLogs()">
          <i class="bi bi-journal-text me-1"></i>Fluent 日志
        </a>
      </li>
    </ul>

    <!-- Commands Tab -->
    <div v-if="tab === 'commands'" class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>ID</th>
              <th>命令</th>
              <th>参数</th>
              <th>状态</th>
              <th>输出</th>
              <th>操作者</th>
              <th>时间</th>
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
                  <i class="bi bi-eye"></i> 查看
                </button>
                <span v-else class="text-muted">-</span>
              </td>
              <td>{{ cmd.creator?.username || '-' }}</td>
              <td>{{ formatTime(cmd.created_at) }}</td>
            </tr>
            <tr v-if="!commands.length">
              <td colspan="7" class="text-center text-muted">暂无命令记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Logs Tab -->
    <div v-if="tab === 'logs'" class="card border-0 shadow-sm">
      <div class="card-body">
        <div v-if="logs.length">
          <div v-for="logEntry in logs" :key="logEntry.id" class="mb-3">
            <div class="text-muted small mb-1">
              {{ formatTime(logEntry.created_at) }} - {{ logEntry.line_count }} 行
            </div>
            <pre class="bg-dark text-light p-2 rounded small" style="max-height: 300px; overflow: auto;">{{ logEntry.lines }}</pre>
          </div>
        </div>
        <div v-else class="text-center text-muted">暂无日志</div>
      </div>
    </div>

    <!-- Output Modal -->
    <div class="modal fade" id="outputModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">命令输出 #{{ selectedCmd?.id }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <pre class="bg-dark text-light p-3 rounded" style="max-height: 500px; overflow: auto;">{{ selectedCmd?.output }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getNode, getNodeMetrics, getNodeLogs, getNodeCommands, sendNodeCommand } from '../api'

const route = useRoute()
const node = ref(null)
const metrics = ref(null)
const commands = ref([])
const logs = ref([])
const tab = ref('commands')
const selectedCmd = ref(null)
let outputModal = null

function statusClass(s) {
  return { 'bg-success': s === 'online', 'bg-warning': s === 'offline', 'bg-danger': s === 'error' }
}
function statusText(s) {
  return { online: '在线', offline: '离线', error: '异常' }[s] || s
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
function formatTime(t) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

async function loadCommands() {
  try {
    const { data } = await getNodeCommands(route.params.id)
    commands.value = data.data || []
  } catch (e) { console.error(e) }
}

async function loadLogs() {
  try {
    const { data } = await getNodeLogs(route.params.id)
    logs.value = data.data || []
  } catch (e) { console.error(e) }
}

async function sendCmd(action) {
  if (action === 'rollback' && !confirm('确认回滚到上一个配置版本?')) return
  try {
    await sendNodeCommand(route.params.id, { action })
    setTimeout(loadCommands, 1000)
  } catch (e) {
    alert('命令发送失败: ' + (e.response?.data?.error || e.message))
  }
}

function showOutput(cmd) {
  selectedCmd.value = cmd
  if (!outputModal) outputModal = new window.bootstrap.Modal(document.getElementById('outputModal'))
  outputModal.show()
}

onMounted(async () => {
  const id = route.params.id
  const [nodeRes, metricsRes] = await Promise.all([
    getNode(id),
    getNodeMetrics(id).catch(() => ({ data: null })),
  ])
  node.value = nodeRes.data
  metrics.value = metricsRes.data
  loadCommands()
})
</script>
