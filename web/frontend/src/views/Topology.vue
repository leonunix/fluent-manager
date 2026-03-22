<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div class="d-flex align-items-center gap-3">
        <h4 class="mb-0">基础设施拓扑</h4>
        <div class="btn-group btn-group-sm">
          <button class="btn" :class="viewMode === 'graph' ? 'btn-primary' : 'btn-outline-primary'" @click="viewMode = 'graph'">
            <i class="bi bi-diagram-3"></i> 拓扑图
          </button>
          <button class="btn" :class="viewMode === 'tree' ? 'btn-primary' : 'btn-outline-primary'" @click="viewMode = 'tree'">
            <i class="bi bi-list-nested"></i> 管理视图
          </button>
        </div>
      </div>
      <div class="btn-group" v-if="viewMode === 'tree'">
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

    <!-- ==================== Graph View ==================== -->
    <div v-if="viewMode === 'graph'">
      <TopologyGraph :tree="tree" @select="handleGraphSelect" />
      <!-- Detail card below graph when something is selected -->
      <div v-if="selected.type" class="card border-0 shadow-sm mt-3">
        <div class="card-body">
          <div v-if="selected.type === 'dc'" class="d-flex justify-content-between align-items-center">
            <div>
              <i class="bi bi-building text-primary me-2"></i>
              <strong>{{ selected.data.alias || selected.data.name }}</strong>
              <span class="badge bg-secondary ms-2">{{ selected.data.provider }}</span>
            </div>
            <button class="btn btn-sm btn-outline-primary" @click="viewMode = 'tree'; selectDC(selected.data)">详细管理</button>
          </div>
          <div v-if="selected.type === 'region'" class="d-flex justify-content-between align-items-center">
            <div>
              <i class="bi bi-globe2 text-info me-2"></i>
              <strong>{{ selected.data.alias || selected.data.name }}</strong>
              <span class="text-muted ms-2">{{ selected.dc?.alias || selected.dc?.name }}</span>
            </div>
            <button class="btn btn-sm btn-outline-primary" @click="viewMode = 'tree'">详细管理</button>
          </div>
          <div v-if="selected.type === 'cluster'" class="d-flex justify-content-between align-items-center">
            <div>
              <i class="bi bi-diagram-3 text-success me-2"></i>
              <strong>{{ selected.data.alias || selected.data.name }}</strong>
              <span v-if="selected.data.environment" class="badge ms-2" :style="{backgroundColor: selected.data.env_color}">{{ selected.data.environment }}</span>
              <span v-if="selected.data.is_default" class="badge bg-purple ms-1" style="background-color:#6f42c1">默认</span>
              <span class="text-muted ms-2">{{ selected.data.online_count }}/{{ selected.data.node_count }} 节点在线</span>
            </div>
            <div>
              <router-link :to="`/nodes?cluster_id=${selected.data.id}`" class="btn btn-sm btn-outline-success me-1">
                <i class="bi bi-hdd-network"></i> 查看节点
              </router-link>
              <button class="btn btn-sm btn-outline-primary" @click="viewMode = 'tree'">详细管理</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== Tree/Management View ==================== -->
    <div v-if="viewMode === 'tree'" class="row g-4">
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
                    <span v-if="cl.is_default" class="badge ms-1" style="background-color:#6f42c1">默认</span>
                    <span v-if="cl.environment" class="badge ms-1" :style="{ backgroundColor: cl.env_color }">
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
        <div v-if="selected.type === 'cluster'" class="card border-0 shadow-sm mb-3">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">
              <i class="bi bi-diagram-3 me-2"></i>{{ selected.data.alias || selected.data.name }}
              <span v-if="selected.data.is_default" class="badge ms-2" style="background-color:#6f42c1">默认集群</span>
            </h6>
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

        <!-- Match Rules for selected cluster -->
        <div v-if="selected.type === 'cluster'" class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0"><i class="bi bi-funnel me-2"></i>自动匹配规则</h6>
            <button class="btn btn-sm btn-outline-primary" @click="openRuleCreate"><i class="bi bi-plus"></i> 新增规则</button>
          </div>
          <div class="card-body p-0">
            <div v-if="!matchRules.length" class="p-3 text-center text-muted small">
              暂无匹配规则。新节点注册时不会自动分配到此集群。<br>
              <span v-if="selected.data.is_default">但此集群为<b>默认集群</b>，未匹配的节点将自动归入此处。</span>
            </div>
            <table v-else class="table table-sm table-hover mb-0">
              <thead><tr>
                <th>规则名</th><th>优先级</th><th>主机名</th><th>IP</th><th>类型</th><th>OS</th><th>标签</th><th>状态</th><th>操作</th>
              </tr></thead>
              <tbody>
                <tr v-for="rule in matchRules" :key="rule.id">
                  <td>{{ rule.name }}</td>
                  <td>{{ rule.priority }}</td>
                  <td><code>{{ rule.hostname_pattern || '*' }}</code></td>
                  <td><code>{{ rule.ip_pattern || '*' }}</code></td>
                  <td>{{ rule.fluent_type || '任意' }}</td>
                  <td>{{ rule.os_pattern || '*' }}</td>
                  <td><code class="small">{{ rule.label_selector || '-' }}</code></td>
                  <td><span :class="rule.is_active ? 'text-success' : 'text-muted'">{{ rule.is_active ? '启用' : '禁用' }}</span></td>
                  <td>
                    <button class="btn btn-sm btn-outline-primary me-1" @click="openRuleEdit(rule)"><i class="bi bi-pencil"></i></button>
                    <button class="btn btn-sm btn-outline-danger" @click="deleteRule(rule)"><i class="bi bi-trash"></i></button>
                  </td>
                </tr>
              </tbody>
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

    <!-- ==================== Create/Edit Topology Modal ==================== -->
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
            <div v-if="formType === 'cluster'" class="mb-3 form-check">
              <input v-model="form.is_default" type="checkbox" class="form-check-input" id="isDefaultCheck">
              <label class="form-check-label" for="isDefaultCheck">设为默认集群（未匹配的新节点将自动归入此集群）</label>
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

    <!-- ==================== Match Rule Modal ==================== -->
    <div class="modal fade" id="ruleModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ ruleForm.id ? '编辑' : '新增' }}匹配规则</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">规则名称</label>
              <input v-model="ruleForm.name" type="text" class="form-control" required placeholder="例: 生产Web服务器">
            </div>
            <div class="mb-3">
              <label class="form-label">优先级 <small class="text-muted">(数字越小优先级越高)</small></label>
              <input v-model.number="ruleForm.priority" type="number" class="form-control" min="0">
            </div>
            <div class="mb-3">
              <label class="form-label">主机名匹配 <small class="text-muted">(Glob: web-*, db-cn-*)</small></label>
              <input v-model="ruleForm.hostname_pattern" type="text" class="form-control" placeholder="留空匹配全部">
            </div>
            <div class="mb-3">
              <label class="form-label">IP匹配 <small class="text-muted">(CIDR: 10.0.0.0/16 或 Glob: 192.168.1.*)</small></label>
              <input v-model="ruleForm.ip_pattern" type="text" class="form-control" placeholder="留空匹配全部">
            </div>
            <div class="mb-3">
              <label class="form-label">Fluent类型</label>
              <select v-model="ruleForm.fluent_type" class="form-select">
                <option value="">任意</option>
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
            <div class="mb-3">
              <label class="form-label">OS匹配 <small class="text-muted">(Glob: linux*, windows*)</small></label>
              <input v-model="ruleForm.os_pattern" type="text" class="form-control" placeholder="留空匹配全部">
            </div>
            <div class="mb-3">
              <label class="form-label">标签选择器 <small class="text-muted">(JSON: {"env":"prod","role":"web"})</small></label>
              <input v-model="ruleForm.label_selector" type="text" class="form-control" placeholder="留空不检查标签">
            </div>
            <div class="form-check">
              <input v-model="ruleForm.is_active" type="checkbox" class="form-check-input" id="ruleActiveCheck">
              <label class="form-check-label" for="ruleActiveCheck">启用此规则</label>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" @click="saveRule">保存</button>
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
  getClusterRules, createClusterRule, updateClusterRule, deleteClusterRule,
} from '../api'

const tree = ref([])
const envs = ref([])
const allDCs = ref([])
const allRegions = ref([])
const matchRules = ref([])
const viewMode = ref('graph')
const selected = reactive({ type: '', id: null, data: null, dc: null, region: null })
const formType = ref('')
const form = reactive({ id: null, name: '', alias: '', provider: '', location: '', description: '', datacenter_id: null, region_id: null, environment_id: null, is_default: false })
const ruleForm = reactive({ id: null, name: '', priority: 0, hostname_pattern: '', ip_pattern: '', fluent_type: '', os_pattern: '', label_selector: '', is_active: true })
let topoModal = null
let ruleModal = null

const formTypeLabel = computed(() => ({ dc: '数据中心', region: '区域', cluster: '集群' }[formType.value] || ''))
function getTopoModal() {
  if (!topoModal) topoModal = new window.bootstrap.Modal(document.getElementById('topoModal'))
  return topoModal
}
function getRuleModal() {
  if (!ruleModal) ruleModal = new window.bootstrap.Modal(document.getElementById('ruleModal'))
  return ruleModal
}

function selectDC(dc) { Object.assign(selected, { type: 'dc', id: dc.id, data: dc, dc: null, region: null }); matchRules.value = [] }
function selectRegion(r, dc) { Object.assign(selected, { type: 'region', id: r.id, data: r, dc, region: null }); matchRules.value = [] }
async function selectCluster(cl, r, dc) {
  Object.assign(selected, { type: 'cluster', id: cl.id, data: cl, dc, region: r })
  await loadClusterRules(cl.id)
}

function handleGraphSelect(v) {
  if (v.type === 'dc') selectDC(v.data)
  else if (v.type === 'region') selectRegion(v.data, v.dc)
  else if (v.type === 'cluster') selectCluster(v.data, v.region, v.dc)
}

function openCreate(type) {
  formType.value = type
  Object.assign(form, { id: null, name: '', alias: '', provider: '', location: '', description: '', datacenter_id: allDCs.value[0]?.id, region_id: allRegions.value[0]?.id, environment_id: null, is_default: false })
  getTopoModal().show()
}
function openEdit(type, data) {
  formType.value = type
  Object.assign(form, { id: data.id, name: data.name, alias: data.alias, provider: data.provider || '', location: data.location || '', description: data.description || '', datacenter_id: data.datacenter_id, region_id: data.region_id, environment_id: data.environment_id, is_default: data.is_default || false })
  getTopoModal().show()
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
  getTopoModal().hide()
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

// Match rules
async function loadClusterRules(clusterID) {
  const { data } = await getClusterRules(clusterID)
  matchRules.value = data.data || []
}
function openRuleCreate() {
  Object.assign(ruleForm, { id: null, name: '', priority: 0, hostname_pattern: '', ip_pattern: '', fluent_type: '', os_pattern: '', label_selector: '', is_active: true })
  getRuleModal().show()
}
function openRuleEdit(rule) {
  Object.assign(ruleForm, rule)
  getRuleModal().show()
}
async function saveRule() {
  if (ruleForm.id) {
    await updateClusterRule(selected.id, ruleForm.id, ruleForm)
  } else {
    await createClusterRule(selected.id, ruleForm)
  }
  getRuleModal().hide()
  loadClusterRules(selected.id)
}
async function deleteRule(rule) {
  if (!confirm(`确认删除规则 ${rule.name}?`)) return
  await deleteClusterRule(selected.id, rule.id)
  loadClusterRules(selected.id)
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
