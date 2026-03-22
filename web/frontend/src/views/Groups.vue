<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">节点分组</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>新建分组
      </button>
    </div>

    <div class="row g-3">
      <div class="col-md-4" v-for="group in groups" :key="group.id">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between">
              <h5>{{ group.name }}</h5>
              <div>
                <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(group)">
                  <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" @click="handleDelete(group)">
                  <i class="bi bi-trash"></i>
                </button>
              </div>
            </div>
            <p class="text-muted small mb-2">{{ group.description || '暂无描述' }}</p>
            <span class="badge bg-secondary">{{ group.node_count || 0 }} 个节点</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div class="modal fade" id="groupModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? '编辑' : '新建' }}分组</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">名称</label>
              <input v-model="form.name" type="text" class="form-control" required>
            </div>
            <div class="mb-3">
              <label class="form-label">描述</label>
              <textarea v-model="form.description" class="form-control" rows="2"></textarea>
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
import { getGroups, createGroup, updateGroup, deleteGroup } from '../api'

const groups = ref([])
const form = reactive({ id: null, name: '', description: '' })
let modal = null

function getModal() {
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('groupModal'))
  return modal
}

async function loadGroups() {
  const { data } = await getGroups()
  groups.value = data.data || []
}

function openCreate() {
  form.id = null
  form.name = ''
  form.description = ''
  getModal().show()
}

function openEdit(g) {
  form.id = g.id
  form.name = g.name
  form.description = g.description
  getModal().show()
}

async function save() {
  if (form.id) {
    await updateGroup(form.id, { name: form.name, description: form.description })
  } else {
    await createGroup({ name: form.name, description: form.description })
  }
  getModal().hide()
  loadGroups()
}

async function handleDelete(g) {
  if (confirm(`确认删除分组 ${g.name}?`)) {
    await deleteGroup(g.id)
    loadGroups()
  }
}

onMounted(loadGroups)
</script>
