<template>
  <div>
    <h4 class="mb-4">部署任务</h4>
    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>ID</th>
              <th>配置</th>
              <th>状态</th>
              <th>进度</th>
              <th>创建者</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in tasks" :key="task.id">
              <td>#{{ task.id }}</td>
              <td>{{ task.config?.template?.name || '-' }} v{{ task.config?.version }}</td>
              <td><span :class="statusClass(task.status)" class="badge">{{ statusText(task.status) }}</span></td>
              <td>
                <div class="progress" style="width: 120px;">
                  <div class="progress-bar bg-success" :style="{ width: successPct(task) + '%' }"></div>
                  <div class="progress-bar bg-danger" :style="{ width: failPct(task) + '%' }"></div>
                </div>
                <small>{{ task.success_count }}/{{ task.fail_count }}/{{ task.total_nodes }}</small>
              </td>
              <td>{{ task.creator?.username || '-' }}</td>
              <td>{{ formatTime(task.created_at) }}</td>
              <td>
                <button class="btn btn-sm btn-outline-primary" @click="viewDetail(task)">
                  <i class="bi bi-eye"></i> 详情
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Detail Modal -->
    <div class="modal fade" id="detailModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">部署详情 #{{ detail?.task?.id }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body" v-if="detail">
            <table class="table table-sm">
              <thead>
                <tr><th>节点</th><th>IP</th><th>状态</th><th>消息</th></tr>
              </thead>
              <tbody>
                <tr v-for="r in detail.records" :key="r.id">
                  <td>{{ r.node?.hostname || r.node_id }}</td>
                  <td>{{ r.node?.ip_address || '-' }}</td>
                  <td><span :class="statusClass(r.status)" class="badge">{{ r.status }}</span></td>
                  <td>{{ r.message || '-' }}</td>
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
import { getDeploys, getDeploy } from '../api'

const tasks = ref([])
const detail = ref(null)
let modal = null

function statusClass(s) {
  return { 'bg-success': s === 'completed' || s === 'success', 'bg-warning': s === 'running' || s === 'pending', 'bg-danger': s === 'failed' }
}
function statusText(s) {
  return { pending: '待执行', running: '执行中', completed: '已完成', failed: '失败' }[s] || s
}
function successPct(t) { return t.total_nodes ? (t.success_count / t.total_nodes * 100) : 0 }
function failPct(t) { return t.total_nodes ? (t.fail_count / t.total_nodes * 100) : 0 }
function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }

async function viewDetail(task) {
  const { data } = await getDeploy(task.id)
  detail.value = data
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('detailModal'))
  modal.show()
}

onMounted(async () => {
  const { data } = await getDeploys({ page_size: 50 })
  tasks.value = data.data || []
})
</script>
