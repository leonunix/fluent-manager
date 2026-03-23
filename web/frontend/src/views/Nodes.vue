<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ t('nodes_page.title') }}</h4>
      <div class="d-flex gap-2">
        <input v-model="search" type="text" class="form-control" :placeholder="t('nodes_page.search_placeholder')" style="width: 200px;" @input="loadNodes">
        <select v-model="statusFilter" class="form-select" style="width: 110px;" @change="loadNodes">
          <option value="">{{ t('common.all_status') }}</option>
          <option value="online">{{ t('nodes_page.online') }}</option>
          <option value="offline">{{ t('nodes_page.offline') }}</option>
          <option value="error">{{ t('nodes_page.error') }}</option>
        </select>
        <select v-model="clusterFilter" class="form-select" style="width: 160px;" @change="loadNodes">
          <option value="">{{ t('common.all_clusters') }}</option>
          <option v-for="cl in clusters" :key="cl.id" :value="cl.id">{{ cl.alias || cl.name }}</option>
        </select>
        <select v-model="envFilter" class="form-select" style="width: 130px;" @change="loadNodes">
          <option value="">{{ t('common.all_environments') }}</option>
          <option v-for="e in envs" :key="e.id" :value="e.id">{{ e.alias || e.name }}</option>
        </select>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>{{ t('nodes_page.hostname') }}</th>
              <th>IP</th>
              <th>{{ t('common.type') }}</th>
              <th>{{ t('status') }}</th>
              <th>{{ t('nodes_page.environment') }}</th>
              <th>{{ t('common.cluster') }}</th>
              <th>{{ t('nodes_page.current_config') }}</th>
              <th>{{ t('nodes_page.last_heartbeat') }}</th>
              <th>{{ t('actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="node in nodes" :key="node.id">
              <td>
                <router-link :to="`/nodes/${node.id}`" class="text-decoration-none">
                  <strong>{{ node.hostname }}</strong>
                </router-link>
                <div class="text-muted small">{{ node.node_uid?.substring(0, 12) }}...</div>
              </td>
              <td>{{ node.ip_address }}</td>
              <td><span class="badge bg-info">{{ node.fluent_type }}</span></td>
              <td>
                <span :class="statusClass(node.status)" class="badge">{{ statusText(node.status) }}</span>
              </td>
              <td>
                <span v-if="node.environment" class="badge" :style="{ backgroundColor: node.environment?.color }">
                  {{ node.environment?.alias || node.environment?.name }}
                </span>
                <span v-else-if="node.cluster?.environment" class="badge" :style="{ backgroundColor: node.cluster.environment?.color }">
                  {{ node.cluster.environment?.alias }}
                </span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span v-if="node.cluster">
                  {{ node.cluster.region?.datacenter?.alias || node.cluster.region?.datacenter?.name }} /
                  {{ node.cluster.region?.alias || node.cluster.region?.name }} /
                  {{ node.cluster.alias || node.cluster.name }}
                </span>
                <span v-else class="text-muted">{{ t('nodes_page.unassigned') }}</span>
              </td>
              <td>{{ node.config ? `v${node.config.version}` : '-' }}</td>
              <td>{{ formatTime(node.last_heartbeat) }}</td>
              <td>
                <router-link :to="`/nodes/${node.id}`" class="btn btn-sm btn-outline-primary me-1">
                  <i class="bi bi-eye"></i>
                </router-link>
                <button class="btn btn-sm btn-outline-danger" @click="handleDelete(node)">
                  <i class="bi bi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a class="page-link" href="#" @click.prevent="page--; loadNodes()">{{ t('common.previous') }}</a>
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= Math.ceil(total / pageSize) }">
          <a class="page-link" href="#" @click.prevent="page++; loadNodes()">{{ t('common.next') }}</a>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getNodes, deleteNode, getClusters, getEnvironments } from '../api'
import { useI18n } from '../i18n'

const route = useRoute()
const nodes = ref([])
const clusters = ref([])
const envs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')
const statusFilter = ref('')
const clusterFilter = ref(route.query.cluster_id || '')
const envFilter = ref('')
const { t, dateLocale } = useI18n()

function statusClass(s) {
  return { 'bg-success': s === 'online', 'bg-warning': s === 'offline', 'bg-danger': s === 'error' }
}
function statusText(s) {
  return {
    online: t('nodes_page.online'),
    offline: t('nodes_page.offline'),
    error: t('nodes_page.error'),
  }[s] || s
}
function formatTime(t) {
  return t ? new Date(t).toLocaleString(dateLocale.value) : '-'
}

async function loadNodes() {
  const params = { page: page.value, page_size: pageSize }
  if (search.value) params.search = search.value
  if (statusFilter.value) params.status = statusFilter.value
  if (clusterFilter.value) params.cluster_id = clusterFilter.value
  if (envFilter.value) params.environment_id = envFilter.value
  const { data } = await getNodes(params)
  nodes.value = data.data || []
  total.value = data.total
}

async function handleDelete(node) {
  if (confirm(t('nodes_page.confirm_delete').replace('{name}', node.hostname))) {
    await deleteNode(node.id)
    loadNodes()
  }
}

onMounted(async () => {
  const [, clRes, envRes] = await Promise.all([loadNodes(), getClusters(), getEnvironments()])
  clusters.value = clRes.data.data || []
  envs.value = envRes.data.data || []
})
</script>
