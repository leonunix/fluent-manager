<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ t('users_page.title') }}</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('users_page.create') }}
      </button>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>{{ t('users_page.username') }}</th>
              <th>{{ t('users_page.display_name') }}</th>
              <th>{{ t('users_page.email') }}</th>
              <th>{{ t('users_page.auth_source') }}</th>
              <th>{{ t('users_page.roles') }}</th>
              <th>{{ t('users_page.scopes') }}</th>
              <th>{{ t('status') }}</th>
              <th>{{ t('users_page.last_login') }}</th>
              <th>{{ t('actions') }}</th>
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
                <span v-if="!userScopesMap[u.id] || userScopesMap[u.id].length === 0" class="badge bg-dark">{{ t('users_page.global') }}</span>
                <span v-for="s in (userScopesMap[u.id] || [])" :key="s.id" class="badge bg-info me-1">
                  {{ scopeIcon(s.scope_type) }} {{ s.scope_name || s.scope_type + ':' + s.scope_id }}
                </span>
              </td>
              <td>
                <span :class="u.is_active ? 'bg-success' : 'bg-danger'" class="badge">
                  {{ u.is_active ? t('users_page.active') : t('users_page.inactive') }}
                </span>
              </td>
              <td>{{ formatTime(u.last_login_at) }}</td>
              <td>
                <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(u)">
                  <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-info me-1" @click="openScopeEdit(u)" :title="t('users_page.scope_manage_title')">
                  <i class="bi bi-shield-lock"></i>
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

    <!-- User Modal -->
    <div class="modal fade" id="userModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? t('users_page.edit_user') : t('users_page.create_user') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.username') }}</label>
              <input v-model="form.username" type="text" class="form-control" :disabled="!!form.id" required>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.display_name') }}</label>
              <input v-model="form.display_name" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.email') }}</label>
              <input v-model="form.email" type="email" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">Password{{ form.id ? t('users_page.password_keep') : '' }}</label>
              <input v-model="form.password" type="password" class="form-control" :required="!form.id">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.roles') }}</label>
              <div v-for="r in allRoles" :key="r.id" class="form-check">
                <input type="checkbox" :value="r.id" v-model="form.role_ids" class="form-check-input" :id="'role-'+r.id">
                <label class="form-check-label" :for="'role-'+r.id">{{ r.name }} - {{ r.description }}</label>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="save">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Scope Modal -->
    <div class="modal fade" id="scopeModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title"><i class="bi bi-shield-lock me-2"></i>{{ t('users_page.scope_title').replace('{name}', scopeUser?.display_name || scopeUser?.username || '') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info small mb-3">
              <i class="bi bi-info-circle me-1"></i>
              {{ t('users_page.scope_help') }}
            </div>

            <div class="mb-3">
              <button class="btn btn-sm btn-outline-primary me-2" @click="addScopeRow('datacenter')">{{ t('users_page.add_datacenter') }}</button>
              <button class="btn btn-sm btn-outline-info me-2" @click="addScopeRow('region')">{{ t('users_page.add_region') }}</button>
              <button class="btn btn-sm btn-outline-success" @click="addScopeRow('cluster')">{{ t('users_page.add_cluster') }}</button>
            </div>

            <div v-if="!scopeRows.length" class="text-center text-muted py-3">
              {{ t('users_page.no_scope_limit') }}
            </div>
            <table v-else class="table table-sm">
              <thead><tr><th>{{ t('users_page.scope_type') }}</th><th>{{ t('users_page.target') }}</th><th></th></tr></thead>
              <tbody>
                <tr v-for="(row, idx) in scopeRows" :key="idx">
                  <td>
                    <span class="badge" :class="{'bg-primary': row.scope_type === 'datacenter', 'bg-info': row.scope_type === 'region', 'bg-success': row.scope_type === 'cluster'}">
                      {{ { datacenter: t('topology_page.dc_label'), region: t('topology_page.region_label'), cluster: t('topology_page.cluster_label') }[row.scope_type] }}
                    </span>
                  </td>
                  <td>
                    <select v-if="row.scope_type === 'datacenter'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="dc in allDCs" :key="dc.id" :value="dc.id">{{ dc.alias || dc.name }}</option>
                    </select>
                    <select v-if="row.scope_type === 'region'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="r in allRegions" :key="r.id" :value="r.id">{{ r.datacenter?.alias || '' }} / {{ r.alias || r.name }}</option>
                    </select>
                    <select v-if="row.scope_type === 'cluster'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="cl in allClusters" :key="cl.id" :value="cl.id">{{ cl.alias || cl.name }}</option>
                    </select>
                  </td>
                  <td><button class="btn btn-sm btn-outline-danger" @click="scopeRows.splice(idx, 1)"><i class="bi bi-trash"></i></button></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveScopes">{{ t('users_page.save_scopes') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getUsers, createUser, updateUser, deleteUser, getRoles, getUserScopes, setUserScopes, getDataCenters, getRegions, getClusters } from '../api'
import { useI18n } from '../i18n'

const users = ref([])
const allRoles = ref([])
const allDCs = ref([])
const allRegions = ref([])
const allClusters = ref([])
const userScopesMap = ref({})
const form = reactive({ id: null, username: '', display_name: '', email: '', password: '', role_ids: [] })
const scopeUser = ref(null)
const scopeRows = ref([])
let userModal = null
let scopeModal = null
const { t, dateLocale } = useI18n()

function formatTime(t) { return t ? new Date(t).toLocaleString(dateLocale.value) : '-' }
function scopeIcon(type) { return { datacenter: '🏢', region: '🌐', cluster: '⚙️' }[type] || '' }
function getUserModal() {
  if (!userModal) userModal = new window.bootstrap.Modal(document.getElementById('userModal'))
  return userModal
}
function getScopeModal() {
  if (!scopeModal) scopeModal = new window.bootstrap.Modal(document.getElementById('scopeModal'))
  return scopeModal
}

async function loadUsers() {
  const { data } = await getUsers({ page_size: 100 })
  users.value = data.data || []
  // Load scopes for all users
  for (const u of users.value) {
    try {
      const { data: scopeData } = await getUserScopes(u.id)
      userScopesMap.value[u.id] = scopeData.data || []
    } catch { userScopesMap.value[u.id] = [] }
  }
}

function openCreate() {
  Object.assign(form, { id: null, username: '', display_name: '', email: '', password: '', role_ids: [] })
  getUserModal().show()
}
function openEdit(u) {
  Object.assign(form, {
    id: u.id, username: u.username, display_name: u.display_name,
    email: u.email, password: '', role_ids: (u.roles || []).map(r => r.id),
  })
  getUserModal().show()
}

async function save() {
  const data = { ...form }
  if (!data.password) delete data.password
  if (form.id) await updateUser(form.id, data)
  else await createUser(data)
  getUserModal().hide()
  loadUsers()
}

async function handleDelete(u) {
  if (confirm(t('users_page.confirm_delete').replace('{name}', u.username))) {
    await deleteUser(u.id)
    loadUsers()
  }
}

function openScopeEdit(u) {
  scopeUser.value = u
  const existing = userScopesMap.value[u.id] || []
  scopeRows.value = existing.map(s => ({ scope_type: s.scope_type, scope_id: s.scope_id }))
  getScopeModal().show()
}
function addScopeRow(type) {
  let defaultID = null
  if (type === 'datacenter') defaultID = allDCs.value[0]?.id
  if (type === 'region') defaultID = allRegions.value[0]?.id
  if (type === 'cluster') defaultID = allClusters.value[0]?.id
  scopeRows.value.push({ scope_type: type, scope_id: defaultID })
}
async function saveScopes() {
  await setUserScopes(scopeUser.value.id, { scopes: scopeRows.value })
  getScopeModal().hide()
  loadUsers()
}

onMounted(async () => {
  const [, rolesRes, dcRes, regRes, clRes] = await Promise.all([
    loadUsers(), getRoles(), getDataCenters(), getRegions(), getClusters(),
  ])
  allRoles.value = rolesRes.data.data || []
  allDCs.value = dcRes.data.data || []
  allRegions.value = regRes.data.data || []
  allClusters.value = clRes.data.data || []
})
</script>
