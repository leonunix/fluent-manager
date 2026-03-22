<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">配置模板</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>新建模板
      </button>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>名称</th>
              <th>类型</th>
              <th>描述</th>
              <th>创建者</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="tpl in templates" :key="tpl.id">
              <td>
                <router-link :to="`/configs/${tpl.id}`" class="text-decoration-none">
                  <strong>{{ tpl.name }}</strong>
                </router-link>
              </td>
              <td><span class="badge bg-info">{{ tpl.fluent_type }}</span></td>
              <td>{{ tpl.description || '-' }}</td>
              <td>{{ tpl.creator?.username || '-' }}</td>
              <td>{{ formatTime(tpl.created_at) }}</td>
              <td>
                <router-link :to="`/configs/${tpl.id}`" class="btn btn-sm btn-outline-primary me-1">
                  <i class="bi bi-eye"></i>
                </router-link>
                <button class="btn btn-sm btn-outline-danger" @click="handleDelete(tpl)">
                  <i class="bi bi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Modal -->
    <div class="modal fade" id="createModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">新建配置模板</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row mb-3">
              <div class="col-md-6">
                <label class="form-label">名称</label>
                <input v-model="form.name" type="text" class="form-control" required>
              </div>
              <div class="col-md-6">
                <label class="form-label">类型</label>
                <select v-model="form.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">描述</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">配置内容</label>
              <textarea v-model="form.content" class="form-control font-monospace" rows="15"
                placeholder="[INPUT]&#10;    Name cpu&#10;    Tag  cpu.local&#10;&#10;[OUTPUT]&#10;    Name  stdout&#10;    Match *"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" @click="save">创建</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getTemplates, createTemplate, deleteTemplate } from '../api'

const templates = ref([])
const form = reactive({ name: '', description: '', fluent_type: 'fluentbit', content: '' })
let modal = null

function formatTime(t) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

async function loadTemplates() {
  const { data } = await getTemplates()
  templates.value = data.data || []
}

function openCreate() {
  form.name = ''
  form.description = ''
  form.fluent_type = 'fluentbit'
  form.content = ''
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('createModal'))
  modal.show()
}

async function save() {
  await createTemplate(form)
  modal.hide()
  loadTemplates()
}

async function handleDelete(tpl) {
  if (confirm(`确认删除模板 ${tpl.name}?`)) {
    await deleteTemplate(tpl.id)
    loadTemplates()
  }
}

onMounted(loadTemplates)
</script>
