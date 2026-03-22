<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">环境管理</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>新建环境
      </button>
    </div>

    <div class="row g-4">
      <div class="col-md-8">
        <div class="card border-0 shadow-sm">
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead>
                <tr>
                  <th>颜色</th>
                  <th>标识</th>
                  <th>别名</th>
                  <th>排序</th>
                  <th>描述</th>
                  <th>关联集群</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="env in envs" :key="env.id">
                  <td>
                    <span class="d-inline-block rounded-circle" :style="{ backgroundColor: env.color, width: '16px', height: '16px' }"></span>
                  </td>
                  <td><code>{{ env.name }}</code></td>
                  <td><span class="badge" :style="{ backgroundColor: env.color }">{{ env.alias || env.name }}</span></td>
                  <td>{{ env.sort_order }}</td>
                  <td class="text-muted small">{{ env.description || '-' }}</td>
                  <td>
                    <span class="badge bg-secondary">{{ clusterCountByEnv[env.id] || 0 }} 集群</span>
                  </td>
                  <td>
                    <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(env)"><i class="bi bi-pencil"></i></button>
                    <button class="btn btn-sm btn-outline-danger" @click="handleDelete(env)"><i class="bi bi-trash"></i></button>
                  </td>
                </tr>
                <tr v-if="!envs.length">
                  <td colspan="7" class="text-center text-muted py-3">暂无环境</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">环境说明</h6></div>
          <div class="card-body small text-muted">
            <p>环境用于区分基础设施的用途阶段。每个集群可以关联一个环境，节点继承所在集群的环境。</p>
            <p class="mb-1"><strong>典型环境：</strong></p>
            <ul class="mb-0">
              <li><span class="badge" style="background-color:#dc3545">生产环境</span> - 线上正式服务</li>
              <li><span class="badge" style="background-color:#ffc107;color:#333">预发布环境</span> - 发布前验证</li>
              <li><span class="badge" style="background-color:#17a2b8">开发环境</span> - 日常开发调试</li>
              <li><span class="badge" style="background-color:#6c757d">测试环境</span> - QA 测试</li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div class="modal fade" id="envModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? '编辑' : '新建' }}环境</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">标识名称 <small class="text-muted">(英文，唯一)</small></label>
              <input v-model="form.name" type="text" class="form-control" placeholder="例: production" required>
            </div>
            <div class="mb-3">
              <label class="form-label">显示别名</label>
              <input v-model="form.alias" type="text" class="form-control" placeholder="例: 生产环境">
            </div>
            <div class="mb-3">
              <label class="form-label">标识颜色</label>
              <div class="d-flex align-items-center gap-2">
                <input v-model="form.color" type="color" class="form-control form-control-color">
                <input v-model="form.color" type="text" class="form-control" style="width:120px" placeholder="#dc3545">
                <span class="badge" :style="{ backgroundColor: form.color }">{{ form.alias || form.name || '预览' }}</span>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">排序 <small class="text-muted">(数字越小越靠前)</small></label>
              <input v-model.number="form.sort_order" type="number" class="form-control" min="0">
            </div>
            <div class="mb-3">
              <label class="form-label">描述</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" @click="save">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getEnvironments, createEnvironment, updateEnvironment, deleteEnvironment, getClusters } from '../api'

const envs = ref([])
const clusterCountByEnv = ref({})
const form = reactive({ id: null, name: '', alias: '', color: '#0d6efd', sort_order: 0, description: '' })
let modal = null

function getModal() {
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('envModal'))
  return modal
}

async function loadData() {
  const [envRes, clRes] = await Promise.all([getEnvironments(), getClusters()])
  envs.value = envRes.data.data || []
  const clusters = clRes.data.data || []
  const counts = {}
  for (const cl of clusters) {
    if (cl.environment_id) {
      counts[cl.environment_id] = (counts[cl.environment_id] || 0) + 1
    }
  }
  clusterCountByEnv.value = counts
}

function openCreate() {
  Object.assign(form, { id: null, name: '', alias: '', color: '#0d6efd', sort_order: envs.value.length + 1, description: '' })
  getModal().show()
}

function openEdit(env) {
  Object.assign(form, { id: env.id, name: env.name, alias: env.alias, color: env.color || '#0d6efd', sort_order: env.sort_order, description: env.description })
  getModal().show()
}

async function save() {
  if (form.id) {
    await updateEnvironment(form.id, form)
  } else {
    await createEnvironment(form)
  }
  getModal().hide()
  loadData()
}

async function handleDelete(env) {
  if (clusterCountByEnv.value[env.id]) {
    alert(`${env.alias || env.name} 下还有 ${clusterCountByEnv.value[env.id]} 个集群，请先解绑`)
    return
  }
  if (confirm(`确认删除环境 ${env.alias || env.name}?`)) {
    await deleteEnvironment(env.id)
    loadData()
  }
}

onMounted(loadData)
</script>
