<template>
  <div>
    <h4 class="mb-4">{{ t('deploys_page.title') }}</h4>
    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>ID</th>
              <th>{{ t('deploys_page.config') }}</th>
              <th>{{ t('status') }}</th>
              <th>{{ t('deploys_page.progress') }}</th>
              <th>{{ t('deploys_page.creator') }}</th>
              <th>{{ t('deploys_page.created_at') }}</th>
              <th>{{ t('actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in tasks" :key="task.id">
              <td>#{{ task.id }}</td>
              <td>{{ task.config?.template?.name || '-' }} v{{ task.config?.version }}</td>
              <td><span :class="statusClass(task.status)" class="badge">{{ statusText(task.status) }}</span></td>
              <td>
                <div class="progress" style="width: 120px;">
                  <div class="progress-bar bg-success" :style="{ width: successPct(task) + '%' }"></div>
                  <div class="progress-bar bg-danger" :style="{ width: failPct(task) + '%' }"></div>
                </div>
                <small>{{ task.success_count }}/{{ task.fail_count }}/{{ task.total_nodes }}</small>
              </td>
              <td>{{ task.creator?.username || '-' }}</td>
              <td>{{ formatTime(task.created_at) }}</td>
              <td>
                <button class="btn btn-sm btn-outline-primary" @click="viewDetail(task)">
                  <i class="bi bi-eye"></i> {{ t('common.details') }}
                </button>
              </td>
            </tr>
            <tr v-if="!tasks.length">
              <td colspan="7" class="text-center text-muted py-4">{{ t('common.no_data') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a class="page-link" href="#" @click.prevent="changePage(page - 1)">{{ t('common.previous') }}</a>
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ totalPages() }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= totalPages() }">
          <a class="page-link" href="#" @click.prevent="changePage(page + 1)">{{ t('common.next') }}</a>
        </li>
      </ul>
    </nav>

    <!-- Detail Modal -->
    <div class="modal fade" id="detailModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('deploys_page.detail_title').replace('{id}', detail?.task?.id || '') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body" v-if="detail">
            <table class="table table-sm">
              <thead>
                <tr><th>{{ t('common.node') }}</th><th>IP</th><th>{{ t('status') }}</th><th>{{ t('deploys_page.message') }}</th></tr>
              </thead>
              <tbody>
                <tr v-for="r in detail.records" :key="r.id">
                  <td>{{ r.node?.hostname || r.node_id }}</td>
                  <td>{{ r.node?.ip_address || '-' }}</td>
                  <td><span :class="statusClass(r.status)" class="badge">{{ r.status }}</span></td>
                  <td>{{ r.message || '-' }}</td>
                </tr>
                <tr v-if="!detail.records?.length">
                  <td colspan="4" class="text-center text-muted py-4">{{ t('common.no_data') }}</td>
                </tr>
              </tbody>
            </table>

            <nav v-if="detailTotal > detailPageSize" class="mt-3">
              <ul class="pagination pagination-sm justify-content-center mb-0">
                <li class="page-item" :class="{ disabled: detailPage <= 1 }">
                  <a class="page-link" href="#" @click.prevent="changeDetailPage(detailPage - 1)">{{ t('common.previous') }}</a>
                </li>
                <li class="page-item disabled">
                  <span class="page-link">{{ detailPage }} / {{ detailTotalPages() }}</span>
                </li>
                <li class="page-item" :class="{ disabled: detailPage >= detailTotalPages() }">
                  <a class="page-link" href="#" @click.prevent="changeDetailPage(detailPage + 1)">{{ t('common.next') }}</a>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getDeploys, getDeploy } from '../api'
import { useI18n } from '../i18n'

const tasks = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const detail = ref(null)
const detailTotal = ref(0)
const detailPage = ref(1)
const detailPageSize = 20
const activeTaskID = ref(null)
let modal = null
const { t, dateLocale } = useI18n()

function statusClass(s) {
  return { 'bg-success': s === 'completed' || s === 'success', 'bg-warning': s === 'running' || s === 'pending', 'bg-danger': s === 'failed' }
}
function statusText(s) {
  return {
    pending: t('deploys_page.pending'),
    running: t('deploys_page.running'),
    completed: t('deploys_page.completed'),
    failed: t('deploys_page.failed'),
  }[s] || s
}
function successPct(t) { return t.total_nodes ? (t.success_count / t.total_nodes * 100) : 0 }
function failPct(t) { return t.total_nodes ? (t.fail_count / t.total_nodes * 100) : 0 }
function formatTime(t) { return t ? new Date(t).toLocaleString(dateLocale.value) : '-' }

function totalPages() {
  return Math.max(1, Math.ceil(total.value / pageSize))
}

function detailTotalPages() {
  return Math.max(1, Math.ceil(detailTotal.value / detailPageSize))
}

async function loadDeploys() {
  const { data } = await getDeploys({ page: page.value, page_size: pageSize })
  tasks.value = data.data || []
  total.value = data.total || 0
}

async function loadDetail(taskID, nextPage = detailPage.value) {
  const { data } = await getDeploy(taskID, { page: nextPage, page_size: detailPageSize })
  detail.value = data
  detailTotal.value = data.total || 0
  detailPage.value = data.page || nextPage
}

async function changePage(nextPage) {
  if (nextPage < 1 || nextPage > totalPages() || nextPage === page.value) return
  page.value = nextPage
  await loadDeploys()
}

async function viewDetail(task) {
  activeTaskID.value = task.id
  detailPage.value = 1
  await loadDetail(task.id, 1)
  if (!modal) modal = new window.bootstrap.Modal(document.getElementById('detailModal'))
  modal.show()
}

async function changeDetailPage(nextPage) {
  if (!activeTaskID.value || nextPage < 1 || nextPage > detailTotalPages() || nextPage === detailPage.value) return
  await loadDetail(activeTaskID.value, nextPage)
}

onMounted(loadDeploys)
</script>
