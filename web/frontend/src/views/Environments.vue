<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ t('environments_page.title') }}</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('environments_page.create') }}
      </button>
    </div>

    <div class="row g-4">
      <div class="col-md-8">
        <div class="card border-0 shadow-sm">
          <div class="card-body p-0">
            <table class="table table-hover mb-0">
              <thead>
                <tr>
                  <th>{{ t('environments_page.color') }}</th>
                  <th>{{ t('environments_page.identifier') }}</th>
                  <th>{{ t('common.alias') }}</th>
                  <th>{{ t('environments_page.sort_order') }}</th>
                  <th>{{ t('common.description') }}</th>
                  <th>{{ t('environments_page.related_clusters') }}</th>
                  <th>{{ t('actions') }}</th>
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
                    <span class="badge bg-secondary">{{ t('environments_page.cluster_count').replace('{count}', clusterCountByEnv[env.id] || 0) }}</span>
                  </td>
                  <td>
                    <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(env)"><i class="bi bi-pencil"></i></button>
                    <button class="btn btn-sm btn-outline-danger" @click="handleDelete(env)"><i class="bi bi-trash"></i></button>
                  </td>
                </tr>
                <tr v-if="!envs.length">
                  <td colspan="7" class="text-center text-muted py-3">{{ t('environments_page.no_envs') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="col-md-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white"><h6 class="mb-0">{{ t('environments_page.guide_title') }}</h6></div>
          <div class="card-body small text-muted">
            <p>{{ t('environments_page.guide_intro') }}</p>
            <p class="mb-1"><strong>{{ t('environments_page.typical_env') }}</strong></p>
            <ul class="mb-0">
              <li><span class="badge" style="background-color:#dc3545">{{ t('environments_page.production') }}</span> - {{ t('environments_page.production_desc') }}</li>
              <li><span class="badge" style="background-color:#ffc107;color:#333">{{ t('environments_page.preprod') }}</span> - {{ t('environments_page.preprod_desc') }}</li>
              <li><span class="badge" style="background-color:#17a2b8">{{ t('environments_page.development') }}</span> - {{ t('environments_page.development_desc') }}</li>
              <li><span class="badge" style="background-color:#6c757d">{{ t('environments_page.testing') }}</span> - {{ t('environments_page.testing_desc') }}</li>
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
            <h5 class="modal-title">{{ form.id ? t('environments_page.edit_title') : t('environments_page.create_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('environments_page.identifier_help') }} <small class="text-muted">(English, unique)</small></label>
              <input v-model="form.name" type="text" class="form-control" :placeholder="t('environments_page.example_production')" required>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('environments_page.alias_help') }}</label>
              <input v-model="form.alias" type="text" class="form-control" :placeholder="t('environments_page.example_alias')">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('environments_page.color_label') }}</label>
              <div class="d-flex align-items-center gap-2">
                <input v-model="form.color" type="color" class="form-control form-control-color">
                <input v-model="form.color" type="text" class="form-control" style="width:120px" placeholder="#dc3545">
                <span class="badge" :style="{ backgroundColor: form.color }">{{ form.alias || form.name || t('environments_page.preview') }}</span>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('environments_page.sort_help') }}</label>
              <input v-model.number="form.sort_order" type="number" class="form-control" min="0">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="save">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getEnvironments, createEnvironment, updateEnvironment, deleteEnvironment, getClusters } from '../api'
import { useI18n } from '../i18n'

const envs = ref([])
const clusterCountByEnv = ref({})
const form = reactive({ id: null, name: '', alias: '', color: '#0d6efd', sort_order: 0, description: '' })
let modal = null
const { t } = useI18n()

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
    alert(t('environments_page.unbind_first')
      .replace('{name}', env.alias || env.name)
      .replace('{count}', clusterCountByEnv.value[env.id]))
    return
  }
  if (confirm(t('environments_page.confirm_delete').replace('{name}', env.alias || env.name))) {
    await deleteEnvironment(env.id)
    loadData()
  }
}

onMounted(loadData)
</script>
