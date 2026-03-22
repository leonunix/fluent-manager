<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">用户管理</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>新建用户
      </button>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>用户名</th>
              <th>显示名</th>
              <th>邮箱</th>
              <th>认证来源</th>
              <th>角色</th>
              <th>状态</th>
              <th>最后登录</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td><strong>{{ u.username }}</strong></td>
              <td>{{ u.display_name || '-' }}</td>
              <td>{{ u.email || '-' }}</td>
              <td><span class="badge bg-secondary">{{ u.auth_source }}</span></td>
              <td>
                <span v-for="r in u.roles" :key="r.id" class="badge bg-primary me-1">{{ r.name }}</span>
              </td>
              <td>
                <span :class="u.is_active ? 'bg-success' : 'bg-danger'" class="badge">
                  {{ u.is_active ? '启用' : '禁用' }}
                </span>
              </td>
              <td>{{ formatTime(u.last_login_at) }}</td>
              <td>
                <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(u)">
                  <i class="bi bi-pencil"></i>
                </button>
                <button v-if="u.username !== 'admin'" class="btn btn-sm btn-outline-danger" @click="handleDelete(u)">
                  <i class="bi bi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <div class="modal fade" id="userModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? '编辑' : '新建' }}用户</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">用户名</label>
              <input v-model="form.username" type="text" class="form-control" :disabled="!!form.id" required>
            </div>
            <div class="mb-3">
              <label class="form-label">显示名</label>
              <input v-model="form.display_name" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">邮箱</label>
              <input v-model="form.email" type="email" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">密码{{ form.id ? ' (留空不修改)' : '' }}</label>
              <input v-model="form.password" type="password" class="form-control" :required="!form.id">
            </div>
            <div class="mb-3">
              <label class="form-label">角色</label>
              <div v-for="r in allRoles" :key="r.id" class="form-check">
                <input type="checkbox" :value="r.id" v-model="form.role_ids" class="form-check-input" :id="'role-'+r.id">
                <label class="form-check-label" :for="'role-'+r.id">{{ r.name }} - {{ r.description }}</label>
              </div>
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
import { getUsers, createUser, updateUser, deleteUser, getRoles } from '../api'

const users = ref([])
const allRoles = ref([])
const form = reactive({ id: null, username: '', display_name: '', email: '', password: '', role_ids: [] })
let modal = null

function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }
function getModal() {
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('userModal'))
  return modal
}

async function loadUsers() {
  const { data } = await getUsers({ page_size: 100 })
  users.value = data.data || []
}

function openCreate() {
  Object.assign(form, { id: null, username: '', display_name: '', email: '', password: '', role_ids: [] })
  getModal().show()
}

function openEdit(u) {
  Object.assign(form, {
    id: u.id, username: u.username, display_name: u.display_name,
    email: u.email, password: '', role_ids: (u.roles || []).map(r => r.id),
  })
  getModal().show()
}

async function save() {
  const data = { ...form }
  if (!data.password) delete data.password
  if (form.id) {
    await updateUser(form.id, data)
  } else {
    await createUser(data)
  }
  getModal().hide()
  loadUsers()
}

async function handleDelete(u) {
  if (confirm(`确认删除用户 ${u.username}?`)) {
    await deleteUser(u.id)
    loadUsers()
  }
}

onMounted(async () => {
  const [, r] = await Promise.all([loadUsers(), getRoles()])
  allRoles.value = r.data.data || []
})
</script>
