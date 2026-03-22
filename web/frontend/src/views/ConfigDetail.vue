<template>
  <div v-if="template">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <router-link to="/configs" class="text-decoration-none">&larr; 返回列表</router-link>
        <h4 class="mt-2 mb-0">{{ template.name }}</h4>
        <span class="badge bg-info">{{ template.fluent_type }}</span>
        <span class="text-muted ms-2">{{ template.description }}</span>
      </div>
      <button class="btn btn-primary" @click="openNewVersion">
        <i class="bi bi-plus-lg me-1"></i>新建版本
      </button>
    </div>

    <div class="row g-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">版本列表</h6></div>
          <div class="list-group list-group-flush">
            <a v-for="v in versions" :key="v.id"
               href="#" class="list-group-item list-group-item-action"
               :class="{ active: selectedVersion?.id === v.id }"
               @click.prevent="selectedVersion = v">
              <div class="d-flex justify-content-between">
                <strong>v{{ v.version }}</strong>
                <small>{{ formatTime(v.created_at) }}</small>
              </div>
              <small class="text-muted">{{ v.comment || '无备注' }}</small>
            </a>
            <div v-if="!versions.length" class="list-group-item text-muted text-center">暂无版本</div>
          </div>
        </div>
      </div>
      <div class="col-md-8">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">
              {{ selectedVersion ? `版本 v${selectedVersion.version}` : '模板内容' }}
            </h6>
            <button v-if="selectedVersion" class="btn btn-sm btn-success" @click="openDeploy">
              <i class="bi bi-rocket me-1"></i>部署此版本
            </button>
          </div>
          <div class="card-body">
            <pre class="bg-dark text-light p-3 rounded" style="max-height: 500px; overflow: auto;">{{ selectedVersion?.content || template.content }}</pre>
            <div v-if="selectedVersion" class="mt-2 text-muted small">
              SHA-256: {{ selectedVersion.hash }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- New Version Modal -->
    <div class="modal fade" id="versionModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">新建配置版本</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">版本备注</label>
              <input v-model="versionForm.comment" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">配置内容</label>
              <textarea v-model="versionForm.content" class="form-control font-monospace" rows="15"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" @click="saveVersion">创建版本</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Deploy Modal -->
    <div class="modal fade" id="deployModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">部署配置 v{{ selectedVersion?.version }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">选择目标分组</label>
              <select v-model="deployForm.group_id" class="form-select">
                <option :value="null">不按分组</option>
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
            </div>
            <div class="mb-3">
              <label class="form-label">或指定节点ID (逗号分隔)</label>
              <input v-model="deployForm.node_ids_text" type="text" class="form-control" placeholder="1,2,3">
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-success" @click="submitDeploy">确认部署</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTemplate, getVersions, createVersion, createDeploy, getGroups } from '../api'

const route = useRoute()
const router = useRouter()
const template = ref(null)
const versions = ref([])
const selectedVersion = ref(null)
const groups = ref([])

const versionForm = reactive({ content: '', comment: '' })
const deployForm = reactive({ group_id: null, node_ids_text: '' })
let versionModal = null
let deployModal = null

function formatTime(t) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

async function loadData() {
  const id = route.params.id
  const [tplRes, versRes, grpRes] = await Promise.all([
    getTemplate(id),
    getVersions(id),
    getGroups(),
  ])
  template.value = tplRes.data
  versions.value = versRes.data.data || []
  groups.value = grpRes.data.data || []
  if (versions.value.length) {
    selectedVersion.value = versions.value[0]
  }
}

function openNewVersion() {
  versionForm.content = template.value.content
  versionForm.comment = ''
  if (!versionModal) versionModal = new window.bootstrap.Modal(document.getElementById('versionModal'))
  versionModal.show()
}

async function saveVersion() {
  await createVersion(route.params.id, versionForm)
  versionModal.hide()
  loadData()
}

function openDeploy() {
  deployForm.group_id = null
  deployForm.node_ids_text = ''
  if (!deployModal) deployModal = new window.bootstrap.Modal(document.getElementById('deployModal'))
  deployModal.show()
}

async function submitDeploy() {
  const data = { config_version_id: selectedVersion.value.id }
  if (deployForm.group_id) data.group_id = deployForm.group_id
  if (deployForm.node_ids_text) {
    data.node_ids = deployForm.node_ids_text.split(',').map(s => parseInt(s.trim())).filter(n => n > 0)
  }
  await createDeploy(data)
  deployModal.hide()
  router.push('/deploys')
}

onMounted(loadData)
</script>
