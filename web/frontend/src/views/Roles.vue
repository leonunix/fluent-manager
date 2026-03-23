<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ t('roles_page.title') }}</h4>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('roles_page.create') }}
      </button>
    </div>

    <div class="row g-3">
      <div class="col-md-6" v-for="role in roles" :key="role.id">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="d-flex justify-content-between">
              <h5>{{ role.name }}</h5>
              <div v-if="role.name !== 'admin'">
                <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(role)">
                  <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" @click="handleDelete(role)">
                  <i class="bi bi-trash"></i>
                </button>
              </div>
            </div>
            <p class="text-muted small">{{ role.description }}</p>
            <div>
              <span v-for="p in role.permissions" :key="p.id" class="badge bg-light text-dark border me-1 mb-1">
                {{ p.name }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div class="modal fade" id="roleModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? t('roles_page.edit_title') : t('roles_page.create_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('common.name') }}</label>
              <input v-model="form.name" type="text" class="form-control" required>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('roles_page.permissions') }}</label>
              <div class="row">
                <div v-for="resource in permResources" :key="resource" class="col-md-4 mb-2">
                  <h6 class="text-capitalize">{{ resource }}</h6>
                  <div v-for="p in permsByResource(resource)" :key="p.id" class="form-check">
                    <input type="checkbox" :value="p.id" v-model="form.permission_ids" class="form-check-input" :id="'perm-'+p.id">
                    <label class="form-check-label" :for="'perm-'+p.id">{{ p.action }}</label>
                  </div>
                </div>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { getRoles, createRole, updateRole, deleteRole, getPermissions } from '../api'
import { useI18n } from '../i18n'

const roles = ref([])
const allPermissions = ref([])
const form = reactive({ id: null, name: '', description: '', permission_ids: [] })
let modal = null
const { t } = useI18n()

const permResources = computed(() => [...new Set(allPermissions.value.map(p => p.resource))])
function permsByResource(r) { return allPermissions.value.filter(p => p.resource === r) }

function getModal() {
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('roleModal'))
  return modal
}

async function loadRoles() {
  const { data } = await getRoles()
  roles.value = data.data || []
}

function openCreate() {
  Object.assign(form, { id: null, name: '', description: '', permission_ids: [] })
  getModal().show()
}

function openEdit(role) {
  Object.assign(form, {
    id: role.id, name: role.name, description: role.description,
    permission_ids: (role.permissions || []).map(p => p.id),
  })
  getModal().show()
}

async function save() {
  if (form.id) {
    await updateRole(form.id, form)
  } else {
    await createRole(form)
  }
  getModal().hide()
  loadRoles()
}

async function handleDelete(role) {
  if (confirm(t('roles_page.confirm_delete').replace('{name}', role.name))) {
    await deleteRole(role.id)
    loadRoles()
  }
}

onMounted(async () => {
  const [, p] = await Promise.all([loadRoles(), getPermissions()])
  allPermissions.value = p.data.data || []
})
</script>
