<template>
  <div class="fm-agent-keys-page">
    <div class="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('agent_keys_page.title') }}</h4>
        <div class="text-muted">{{ t('agent_keys_page.subtitle') }}</div>
      </div>
      <button class="btn btn-outline-primary" :disabled="loading" @click="loadData">
        <i class="bi bi-arrow-clockwise me-1"></i>{{ t('common.refresh') }}
      </button>
    </div>

    <div v-if="createdSecret" class="alert alert-warning shadow-sm mb-4">
      <div class="fw-semibold mb-1">{{ t('agent_keys_page.plaintext_title') }}</div>
      <div class="small mb-3">{{ t('agent_keys_page.plaintext_hint') }}</div>
      <div class="input-group">
        <input :value="createdSecret" class="form-control font-monospace" readonly>
        <button class="btn btn-dark" type="button" @click="copyCreatedSecret">
          <i class="bi bi-clipboard me-1"></i>{{ t('agent_keys_page.copy_key') }}
        </button>
      </div>
    </div>

    <div class="row g-3 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small">{{ t('agent_keys_page.total_keys') }}</div>
            <div class="fs-3 fw-semibold">{{ keys.length }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small">{{ t('agent_keys_page.active_keys') }}</div>
            <div class="fs-3 fw-semibold text-success">{{ activeCount }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small">{{ t('agent_keys_page.bound_keys') }}</div>
            <div class="fs-3 fw-semibold text-primary">{{ boundCount }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body">
        <div class="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-3">
          <div>
            <div class="fw-semibold">{{ editingId ? t('agent_keys_page.edit_title') : t('agent_keys_page.create_title') }}</div>
            <div class="small text-muted">{{ t('agent_keys_page.create_hint') }}</div>
          </div>
          <button v-if="editingId" class="btn btn-sm btn-outline-secondary" type="button" @click="resetForm">
            {{ t('agent_keys_page.cancel_edit') }}
          </button>
        </div>

        <form class="row g-3" @submit.prevent="saveKey">
          <div class="col-md-4">
            <label class="form-label">{{ t('common.name') }}</label>
            <input v-model.trim="form.name" type="text" class="form-control" required :disabled="!canSubmit || saving">
          </div>
          <div class="col-md-4">
            <label class="form-label">{{ t('agent_keys_page.cluster_binding') }}</label>
            <select v-model="form.cluster_id" class="form-select" :disabled="!canSubmit || saving">
              <option :value="null">{{ t('agent_keys_page.unbound_cluster') }}</option>
              <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
                {{ clusterLabel(cluster) }}
              </option>
            </select>
          </div>
          <div class="col-md-2">
            <label class="form-label">{{ t('status') }}</label>
            <select v-model="form.is_active" class="form-select" :disabled="!canSubmit || saving">
              <option :value="true">{{ t('common.enabled') }}</option>
              <option :value="false">{{ t('common.disabled') }}</option>
            </select>
          </div>
          <div class="col-md-2 d-flex align-items-end">
            <button class="btn btn-primary w-100" type="submit" :disabled="!canSubmit || saving">
              <i class="bi" :class="editingId ? 'bi-check2-circle' : 'bi-plus-circle'"></i>
              <span class="ms-1">{{ editingId ? t('save') : t('agent_keys_page.create_action') }}</span>
            </button>
          </div>
          <div class="col-12">
            <label class="form-label">{{ t('common.description') }}</label>
            <textarea v-model.trim="form.description" rows="2" class="form-control" :placeholder="t('agent_keys_page.description_placeholder')" :disabled="!canSubmit || saving"></textarea>
          </div>
        </form>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('agent_keys_page.key_preview') }}</th>
                <th>{{ t('common.cluster') }}</th>
                <th>{{ t('status') }}</th>
                <th>{{ t('agent_keys_page.last_used_at') }}</th>
                <th>{{ t('agent_keys_page.created_at') }}</th>
                <th class="text-end">{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody v-if="keys.length">
              <tr v-for="item in keys" :key="item.id">
                <td>
                  <div class="fw-semibold">{{ item.name }}</div>
                  <div class="small text-muted">{{ item.description || t('common.no_description') }}</div>
                </td>
                <td><code>{{ item.key_preview }}</code></td>
                <td>{{ item.cluster ? clusterLabel(item.cluster) : t('agent_keys_page.unbound_cluster') }}</td>
                <td>
                  <span class="badge" :class="item.is_active ? 'text-bg-success' : 'text-bg-secondary'">
                    {{ item.is_active ? t('common.enabled') : t('common.disabled') }}
                  </span>
                </td>
                <td>{{ formatTime(item.last_used_at) }}</td>
                <td>{{ formatTime(item.created_at) }}</td>
                <td class="text-end">
                  <button class="btn btn-sm btn-outline-primary me-2" :disabled="!canEdit" @click="startEdit(item)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-warning me-2" :disabled="!canEdit" @click="toggleStatus(item)">
                    <i class="bi" :class="item.is_active ? 'bi-pause-circle' : 'bi-play-circle'"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" :disabled="!canDelete" @click="removeKey(item)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
            </tbody>
            <tbody v-else>
              <tr>
                <td colspan="7" class="text-center text-muted py-5">
                  {{ loading ? t('loading') : t('agent_keys_page.empty_state') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { createAgentAccessKey, deleteAgentAccessKey, getAgentAccessKeys, getClusters, updateAgentAccessKey } from '../api'
import { useI18n } from '../i18n'
import { useAuthStore } from '../store/auth'

const { t, locale } = useI18n()
const auth = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const keys = ref([])
const clusters = ref([])
const createdSecret = ref('')
const editingId = ref(null)
const form = reactive(createEmptyForm())

const canCreate = computed(() => auth.hasPermission('agent_keys', 'create'))
const canEdit = computed(() => auth.hasPermission('agent_keys', 'update'))
const canDelete = computed(() => auth.hasPermission('agent_keys', 'delete'))
const canSubmit = computed(() => editingId.value ? canEdit.value : canCreate.value)
const activeCount = computed(() => keys.value.filter(item => item.is_active).length)
const boundCount = computed(() => keys.value.filter(item => item.cluster_id).length)

function createEmptyForm() {
  return {
    name: '',
    cluster_id: null,
    description: '',
    is_active: true,
  }
}

function resetForm() {
  Object.assign(form, createEmptyForm())
  editingId.value = null
}

function clusterLabel(cluster) {
  const dc = cluster?.region?.datacenter?.alias || cluster?.region?.datacenter?.name
  const region = cluster?.region?.alias || cluster?.region?.name
  const name = cluster?.alias || cluster?.name
  return [dc, region, name].filter(Boolean).join(' / ')
}

function formatTime(value) {
  if (!value) return t('none')
  const currentLocale = locale.value === 'zh' ? 'zh-CN' : locale.value === 'ja' ? 'ja-JP' : 'en-US'
  return new Date(value).toLocaleString(currentLocale, { hour12: false })
}

async function loadData() {
  loading.value = true
  try {
    const requests = [getAgentAccessKeys()]
    if (auth.hasPermission('topology', 'read')) {
      requests.push(getClusters())
    }
    const [keysResponse, clusterResponse] = await Promise.all(requests)
    keys.value = keysResponse.data || []
    clusters.value = clusterResponse?.data?.data || []
  } catch (error) {
    alert(`${t('agent_keys_page.load_failed')}: ${getErrorMessage(error)}`)
  } finally {
    loading.value = false
  }
}

function startEdit(item) {
  editingId.value = item.id
  Object.assign(form, {
    name: item.name,
    cluster_id: item.cluster_id ?? null,
    description: item.description || '',
    is_active: !!item.is_active,
  })
}

async function saveKey() {
  if (!form.name) return
  saving.value = true
  try {
    const payload = {
      name: form.name,
      cluster_id: form.cluster_id ?? null,
      description: form.description,
      is_active: form.is_active,
    }
    if (editingId.value) {
      await updateAgentAccessKey(editingId.value, payload)
      createdSecret.value = ''
    } else {
      const result = await createAgentAccessKey(payload)
      createdSecret.value = result.plaintext_key || ''
    }
    resetForm()
    await loadData()
  } catch (error) {
    alert(`${t('agent_keys_page.save_failed')}: ${getErrorMessage(error)}`)
  } finally {
    saving.value = false
  }
}

async function toggleStatus(item) {
  try {
    await updateAgentAccessKey(item.id, {
      name: item.name,
      cluster_id: item.cluster_id ?? null,
      description: item.description,
      is_active: !item.is_active,
    })
    await loadData()
  } catch (error) {
    alert(`${t('agent_keys_page.save_failed')}: ${getErrorMessage(error)}`)
  }
}

async function removeKey(item) {
  if (!confirm(t('agent_keys_page.confirm_delete').replace('{name}', item.name))) return
  try {
    await deleteAgentAccessKey(item.id)
    if (editingId.value === item.id) {
      resetForm()
    }
    await loadData()
  } catch (error) {
    alert(`${t('agent_keys_page.delete_failed')}: ${getErrorMessage(error)}`)
  }
}

async function copyCreatedSecret() {
  try {
    await navigator.clipboard.writeText(createdSecret.value)
  } catch {
    alert(createdSecret.value)
  }
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

onMounted(loadData)
</script>
