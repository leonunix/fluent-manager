<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div class="d-flex align-items-center gap-3">
        <h4 class="mb-0">基础设施拓扑</h4>
        <div class="btn-group">
          <button class="btn btn-sm" :class="viewMode === 'graph' ? 'btn-primary' : 'btn-outline-primary'" @click="viewMode = 'graph'">
            <i class="bi bi-diagram-3"></i> 拓扑图
          </button>
          <button class="btn btn-sm" :class="viewMode === 'tree' ? 'btn-primary' : 'btn-outline-primary'" @click="viewMode = 'tree'">
            <i class="bi bi-list-nested"></i> 管理视图
          </button>
        </div>
      </div>
      <div v-if="viewMode === 'tree'" class="btn-group">
        <button class="btn btn-outline-primary btn-sm" @click="openCreate('dc')">
          <i class="bi bi-building"></i> 新建数据中心
        </button>
        <button class="btn btn-outline-primary btn-sm" @click="openCreate('region')">
          <i class="bi bi-globe2"></i> 新建区域
        </button>
        <button class="btn btn-outline-primary btn-sm" @click="openCreate('cluster')">
          <i class="bi bi-diagram-3"></i> 新建集群
        </button>
      </div>
    </div>

    <!-- Graph View -->
    <TopologyGraph v-if="viewMode === 'graph'" :tree="tree" />

    <!-- Tree Management View -->
    <div v-else class="row g-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">拓扑树</h6></div>
          <div class="card-body p-0">
            <div v-if="!tree.length" class="p-3 text-center text-muted">暂无数据中心，请先创建</div>
            <div v-for="dc in tree" :key="'dc-'+dc.id" class="border-bottom">
              <div class="d-flex align-items-center px-3 py-2 bg-light cursor-pointer"
                   @click="selectDC(dc)" :class="{ 'border-start border-primary border-3': selected.type === 'dc' && selected.id === dc.id }">
                <i class="bi bi-building me-2 text-primary"></i>
                <strong>{{ dc.alias || dc.name }}</strong>
                <span class="badge bg-secondary ms-auto">{{ dc.provider }}</span>
              </div>
              <div v-for="r in dc.regions" :key="'r-'+r.id" class="border-bottom">
                <div class="d-flex align-items-center px-3 py-2 ps-4 cursor-pointer"
                     @click="selectRegion(r, dc)" :class="{ 'border-start border-info border-3': selected.type === 'region' && selected.id === r.id }">
                  <i class="bi bi-globe2 me-2 text-info"></i>
                  {{ r.alias || r.name }}
                </div>
                <div v-for="cl in r.clusters" :key="'cl-'+cl.id">
                  <div class="d-flex align-items-center px-3 py-2 ps-5 cursor-pointer"
                       @click="selectCluster(cl, r, dc)" :class="{ 'border-start border-success border-3': selected.type === 'cluster' && selected.id === cl.id }">
                    <i class="bi bi-diagram-3 me-2 text-success"></i>
                    {{ cl.alias || cl.name }}
                    <span v-if="cl.environment" class="badge ms-2" :style="{ backgroundColor: cl.env_color }">
                      {{ cl.environment }}
                    </span>
                    <span class="text-muted small ms-auto">{{ cl.online_count }}/{{ cl.node_count }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Detail panel -->
      <div class="col-md-8">
        <!-- DC detail -->
        <div v-if="selected.type === 'dc'" class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0"><i class="bi bi-building me-2"></i>{{ selected.data.alias || selected.data.name }}</h6>
            <div>
              <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit('dc', selected.data)"><i class="bi bi-pencil"></i></button>
              <button class="btn btn-sm btn-outline-danger" @click="handleDelete('dc', selected.data)"><i class="bi bi-trash"></i></button>
            </div>
          </div>
          <div class="card-body">
            <table class="table table-sm">
              <tr><td class="text-muted w-25">名称</td><td>{{ selected.data.name }}</td></tr>
              <tr><td class="text-muted">别名</td><td>{{ selected.data.alias || '-' }}</td></tr>
              <tr><td class="text-muted">供应商</td><td>{{ selected.data.provider || '-' }}</td></tr>
              <tr><td class="text-muted">位置</td><td>{{ selected.data.location || '-' }}</td></tr>
              <tr><td class="text-muted">区域数</td><td>{{ selected.data.regions?.length || 0 }}</td></tr>
            </table>
          </div>
        </div>

        <!-- Region detail -->
        <div v-if="selected.type === 'region'" class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0"><i class="bi bi-globe2 me-2"></i>{{ selected.data.alias || selected.data.name }}</h6>
            <div>
              <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit('region', selected.data)"><i class="bi bi-pencil"></i></button>
              <button class="btn btn-sm btn-outline-danger" @click="handleDelete('region', selected.data)"><i class="bi bi-trash"></i></button>
            </div>
          </div>
          <div class="card-body">
            <table class="table table-sm">
              <tr><td class="text-muted w-25">名称</td><td>{{ selected.data.name }}</td></tr>
              <tr><td class="text-muted">别名</td><td>{{ selected.data.alias || '-' }}</td></tr>
              <tr><td class="text-muted">数据中心</td><td>{{ selected.dc?.alias || selected.dc?.name }}</td></tr>
              <tr><td class="text-muted">集群数</td><td>{{ selected.data.clusters?.length || 0 }}</td></tr>
            </table>
          </div>
        </div>

        <!-- Cluster detail -->
        <div v-if="selected.type === 'cluster'" class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0"><i class="bi bi-diagram-3 me-2"></i>{{ selected.data.alias || selected.data.name }}</h6>
            <div>
              <router-link :to="`/nodes?cluster_id=${selected.data.id}`" class="btn btn-sm btn-outline-success me-1">
                <i class="bi bi-hdd-network"></i> 查看节点
              </router-link>
              <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit('cluster', selected.data)"><i class="bi bi-pencil"></i></button>
              <button class="btn btn-sm btn-outline-danger" @click="handleDelete('cluster', selected.data)"><i class="bi bi-trash"></i></button>
            </div>
          </div>
          <div class="card-body">
            <table class="table table-sm">
              <tr><td class="text-muted w-25">名称</td><td>{{ selected.data.name }}</td></tr>
              <tr><td class="text-muted">别名</td><td>{{ selected.data.alias || '-' }}</td></tr>
              <tr><td class="text-muted">区域</td><td>{{ selected.region?.alias || selected.region?.name }}</td></tr>
              <tr><td class="text-muted">数据中心</td><td>{{ selected.dc?.alias || selected.dc?.name }}</td></tr>
              <tr><td class="text-muted">环境</td><td>
                <span v-if="selected.data.environment" class="badge" :style="{ backgroundColor: selected.data.env_color }">{{ selected.data.environment }}</span>
                <span v-else>-</span>
              </td></tr>
              <tr><td class="text-muted">节点</td><td>{{ selected.data.online_count }} 在线 / {{ selected.data.node_count }} 总计</td></tr>
            </table>
          </div>
        </div>

        <div v-if="!selected.type" class="card border-0 shadow-sm">
          <div class="card-body text-center text-muted py-5">
            <i class="bi bi-diagram-3 display-4 d-block mb-3"></i>
            从左侧拓扑树选择一个数据中心、区域或集群查看详情
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div class="modal fade" id="topoModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? '编辑' : '新建' }}{{ formTypeLabel }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">名称</label>
              <input v-model="form.name" type="text" class="form-control" required>
            </div>
            <div class="mb-3">
              <label class="form-label">别名</label>
              <input v-model="form.alias" type="text" class="form-control">
            </div>
            <div v-if="formType === 'dc'" class="mb-3">
              <label class="form-label">供应商</label>
              <select v-model="form.provider" class="form-select">
                <option value="">自定义</option>
                <option value="aws">AWS</option>
                <option value="aliyun">阿里云</option>
                <option value="azure">Azure</option>
                <option value="gcp">GCP</option>
                <option value="tencent">腾讯云</option>
                <option value="huawei">华为云</option>
                <option value="idc">自建IDC</option>
              </select>
            </div>
            <div v-if="formType === 'dc'" class="mb-3">
              <label class="form-label">位置</label>
              <input v-model="form.location" type="text" class="form-control">
            </div>
            <div v-if="formType === 'region'" class="mb-3">
              <label class="form-label">所属数据中心</label>
              <select v-model="form.datacenter_id" class="form-select" required>
                <option v-for="dc in allDCs" :key="dc.id" :value="dc.id">{{ dc.alias || dc.name }}</option>
              </select>
            </div>
            <div v-if="formType === 'cluster'" class="mb-3">
              <label class="form-label">所属区域</label>
              <select v-model="form.region_id" class="form-select" required>
                <option v-for="r in allRegions" :key="r.id" :value="r.id">
                  {{ r.datacenter?.alias || r.datacenter?.name }} / {{ r.alias || r.name }}
                </option>
              </select>
            </div>
            <div v-if="formType === 'cluster'" class="mb-3">
              <label class="form-label">环境</label>
              <select v-model="form.environment_id" class="form-select">
                <option :value="null">不指定</option>
                <option v-for="e in envs" :key="e.id" :value="e.id">{{ e.alias || e.name }}</option>
              </select>
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
import { ref, reactive, computed, onMounted } from 'vue'
import TopologyGraph from '../components/TopologyGraph.vue'
import {
  getTopologyTree, getEnvironments,
  getDataCenters, createDataCenter, updateDataCenter, deleteDataCenter,
  getRegions, createRegion, updateRegion, deleteRegion,
  getClusters, createCluster, updateCluster, deleteCluster,
} from '../api'

const viewMode = ref('graph')
const tree = ref([])
const envs = ref([])
const allDCs = ref([])
const allRegions = ref([])
const selected = reactive({ type: '', id: null, data: null, dc: null, region: null })
const formType = ref('')
const form = reactive({ id: null, name: '', alias: '', provider: '', location: '', description: '', datacenter_id: null, region_id: null, environment_id: null })
let modal = null

const formTypeLabel = computed(() => ({ dc: '数据中心', region: '区域', cluster: '集群' }[formType.value] || ''))
function getModal() {
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('topoModal'))
  return modal
}

function selectDC(dc) { Object.assign(selected, { type: 'dc', id: dc.id, data: dc, dc: null, region: null }) }
function selectRegion(r, dc) { Object.assign(selected, { type: 'region', id: r.id, data: r, dc, region: null }) }
function selectCluster(cl, r, dc) { Object.assign(selected, { type: 'cluster', id: cl.id, data: cl, dc, region: r }) }

function openCreate(type) {
  formType.value = type
  Object.assign(form, { id: null, name: '', alias: '', provider: '', location: '', description: '', datacenter_id: allDCs.value[0]?.id, region_id: allRegions.value[0]?.id, environment_id: null })
  getModal().show()
}
function openEdit(type, data) {
  formType.value = type
  Object.assign(form, { id: data.id, name: data.name, alias: data.alias, provider: data.provider || '', location: data.location || '', description: data.description || '', datacenter_id: data.datacenter_id, region_id: data.region_id, environment_id: data.environment_id })
  getModal().show()
}

async function save() {
  const t = formType.value
  if (form.id) {
    if (t === 'dc') await updateDataCenter(form.id, form)
    else if (t === 'region') await updateRegion(form.id, form)
    else if (t === 'cluster') await updateCluster(form.id, form)
  } else {
    if (t === 'dc') await createDataCenter(form)
    else if (t === 'region') await createRegion(form)
    else if (t === 'cluster') await createCluster(form)
  }
  getModal().hide()
  loadAll()
}

async function handleDelete(type, data) {
  const label = data.alias || data.name
  if (!confirm(`确认删除 ${label}?`)) return
  if (type === 'dc') await deleteDataCenter(data.id)
  else if (type === 'region') await deleteRegion(data.id)
  else if (type === 'cluster') await deleteCluster(data.id)
  selected.type = ''
  loadAll()
}

async function loadAll() {
  const [treeRes, envRes, dcRes, regRes] = await Promise.all([
    getTopologyTree(), getEnvironments(), getDataCenters(), getRegions(),
  ])
  tree.value = treeRes.data.data || []
  envs.value = envRes.data.data || []
  allDCs.value = dcRes.data.data || []
  allRegions.value = regRes.data.data || []
}

onMounted(loadAll)
</script>

<style scoped>
.cursor-pointer { cursor: pointer; }
.cursor-pointer:hover { background-color: #f8f9fa; }
</style>
