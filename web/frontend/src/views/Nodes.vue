<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">节点管理</h4>
      <div class="d-flex gap-2">
        <input v-model="search" type="text" class="form-control" placeholder="搜索主机名/IP..." style="width: 250px;" @input="loadNodes">
        <select v-model="statusFilter" class="form-select" style="width: 130px;" @change="loadNodes">
          <option value="">全部状态</option>
          <option value="online">在线</option>
          <option value="offline">离线</option>
          <option value="error">异常</option>
        </select>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>主机名</th>
              <th>IP</th>
              <th>类型</th>
              <th>状态</th>
              <th>分组</th>
              <th>当前配置</th>
              <th>最后心跳</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="node in nodes" :key="node.id">
              <td>
                <strong>{{ node.hostname }}</strong>
                <div class="text-muted small">{{ node.node_uid }}</div>
              </td>
              <td>{{ node.ip_address }}</td>
              <td><span class="badge bg-info">{{ node.fluent_type }}</span></td>
              <td>
                <span :class="statusClass(node.status)" class="badge">{{ statusText(node.status) }}</span>
              </td>
              <td>{{ node.group?.name || '-' }}</td>
              <td>{{ node.config ? `v${node.config.version}` : '-' }}</td>
              <td>{{ formatTime(node.last_heartbeat) }}</td>
              <td>
                <button class="btn btn-sm btn-outline-primary me-1" @click="editNode(node)">
                  <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" @click="handleDelete(node)">
                  <i class="bi bi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a class="page-link" href="#" @click.prevent="page--; loadNodes()">上一页</a>
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= Math.ceil(total / pageSize) }">
          <a class="page-link" href="#" @click.prevent="page++; loadNodes()">下一页</a>
        </li>
      </ul>
    </nav>

    <!-- Edit Modal -->
    <div class="modal fade" id="editModal" tabindex="-1" ref="editModalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">编辑节点</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">分组</label>
              <select v-model="editForm.group_id" class="form-select">
                <option :value="null">无分组</option>
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
            </div>
            <div class="mb-3">
              <label class="form-label">标签 (JSON)</label>
              <textarea v-model="editForm.labels" class="form-control" rows="3"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" @click="saveNode">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getNodes, updateNode, deleteNode, getGroups } from '../api'

const nodes = ref([])
const groups = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')
const statusFilter = ref('')
const editForm = reactive({ id: null, group_id: null, labels: '' })
let editModal = null

function statusClass(s) {
  return { 'bg-success': s === 'online', 'bg-warning': s === 'offline', 'bg-danger': s === 'error' }
}
function statusText(s) {
  return { online: '在线', offline: '离线', error: '异常' }[s] || s
}
function formatTime(t) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

async function loadNodes() {
  const params = { page: page.value, page_size: pageSize }
  if (search.value) params.search = search.value
  if (statusFilter.value) params.status = statusFilter.value
  const { data } = await getNodes(params)
  nodes.value = data.data || []
  total.value = data.total
}

function editNode(node) {
  editForm.id = node.id
  editForm.group_id = node.group_id
  editForm.labels = node.labels || ''
  if (!editModal) {
    editModal = new window.bootstrap.Modal(document.getElementById('editModal'))
  }
  editModal.show()
}

async function saveNode() {
  await updateNode(editForm.id, { group_id: editForm.group_id, labels: editForm.labels })
  editModal.hide()
  loadNodes()
}

async function handleDelete(node) {
  if (confirm(`确认删除节点 ${node.hostname}?`)) {
    await deleteNode(node.id)
    loadNodes()
  }
}

onMounted(async () => {
  const [, g] = await Promise.all([loadNodes(), getGroups()])
  groups.value = g.data.data || []
})
</script>
