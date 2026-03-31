<template>
  <div class="container-fluid px-0">
    <div
      class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4"
    >
      <div>
        <h4 class="mb-1">{{ t('deploys_page.title') }}</h4>
        <div class="text-muted">{{ t('deploys_page.subtitle') }}</div>
      </div>
      <button
        class="btn btn-outline-secondary"
        @click="refreshAll"
        :disabled="loadingAny"
      >
        <i class="bi bi-arrow-repeat me-1"></i>{{ t('common.refresh') }}
      </button>
    </div>

    <div class="row g-3 mb-4">
      <div class="col-6 col-xl-3">
        <div class="fm-deploy-stat-card">
          <div class="fm-deploy-stat-card__label">
            {{ t('deploys_page.total_tasks') }}
          </div>
          <div class="fm-deploy-stat-card__value">{{ total }}</div>
          <div class="fm-deploy-stat-card__meta">
            {{ t('deploys_page.page_tasks').replace('{count}', String(tasks.length)) }}
          </div>
        </div>
      </div>
      <div class="col-6 col-xl-3">
        <div class="fm-deploy-stat-card">
          <div class="fm-deploy-stat-card__label text-info-emphasis">
            {{ t('deploys_page.active_tasks') }}
          </div>
          <div class="fm-deploy-stat-card__value">{{ activeTaskCount }}</div>
          <div class="fm-deploy-stat-card__meta">
            {{ t('deploys_page.awaiting_nodes').replace('{count}', String(pagePendingNodes)) }}
          </div>
        </div>
      </div>
      <div class="col-6 col-xl-3">
        <div class="fm-deploy-stat-card">
          <div class="fm-deploy-stat-card__label text-success">
            {{ t('deploys_page.success_nodes') }}
          </div>
          <div class="fm-deploy-stat-card__value">{{ pageSuccessNodes }}</div>
          <div class="fm-deploy-stat-card__meta">
            {{ t('deploys_page.completed_batches').replace('{count}', String(completedTaskCount)) }}
          </div>
        </div>
      </div>
      <div class="col-6 col-xl-3">
        <div class="fm-deploy-stat-card">
          <div class="fm-deploy-stat-card__label text-danger">
            {{ t('deploys_page.failed_nodes') }}
          </div>
          <div class="fm-deploy-stat-card__value">{{ pageFailedNodes }}</div>
          <div class="fm-deploy-stat-card__meta">
            {{ t('deploys_page.failure_hint') }}
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div
          class="d-flex justify-content-between align-items-center p-3 border-bottom"
        >
          <div>
            <h5 class="mb-1">{{ t('deploys_page.task_list_title') }}</h5>
            <div class="text-muted small">
              {{ t('deploys_page.task_list_hint') }}
            </div>
          </div>
          <span class="badge bg-secondary-subtle text-secondary-emphasis">{{
            total
          }}</span>
        </div>

        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
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
              <tr
                v-for="task in tasks"
                :key="task.id"
                :class="{ 'fm-task-row-active': activeTaskID === task.id }"
              >
                <td>#{{ task.id }}</td>
                <td>
                  <div class="fw-semibold">
                    {{ deployConfigLabel(task) }}
                  </div>
                  <div class="small text-muted d-flex flex-wrap gap-2">
                    <span>{{ deployScopeText(task) }}</span>
                    <span>{{ task.total_nodes }} {{ t('common.nodes') }}</span>
                  </div>
                </td>
                <td>
                  <span class="badge" :class="statusClass(taskDisplayStatus(task))">{{
                    statusText(taskDisplayStatus(task))
                  }}</span>
                </td>
                <td>
                  <div class="progress" style="width: 148px">
                    <div
                      class="progress-bar bg-success"
                      :style="{ width: successPct(task) + '%' }"
                    ></div>
                    <div
                      class="progress-bar bg-danger"
                      :style="{ width: failPct(task) + '%' }"
                    ></div>
                    <div
                      class="progress-bar fm-progress-pending"
                      :style="{ width: pendingPct(task) + '%' }"
                    ></div>
                  </div>
                  <small>{{
                    `${task.success_count}/${task.fail_count}/${task.total_nodes}`
                  }}</small>
                </td>
                <td>{{ task.creator?.username || '-' }}</td>
                <td>{{ formatTime(task.created_at) }}</td>
                <td>
                  <button
                    class="btn btn-sm btn-outline-primary"
                    @click="viewDetail(task)"
                  >
                    <i class="bi bi-eye me-1"></i>{{ t('common.details') }}
                  </button>
                </td>
              </tr>
              <tr v-if="!tasks.length">
                <td colspan="7" class="text-center text-muted py-4">
                  {{ t('common.no_data') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center mb-0">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a
            class="page-link"
            href="#"
            @click.prevent="changePage(page - 1)"
            >{{ t('common.previous') }}</a
          >
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ totalPages() }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= totalPages() }">
          <a
            class="page-link"
            href="#"
            @click.prevent="changePage(page + 1)"
            >{{ t('common.next') }}</a
          >
        </li>
      </ul>
    </nav>

    <div class="modal fade" id="detailModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header border-0 pb-0">
            <div>
              <h5 class="modal-title">
                {{ t('deploys_page.detail_title').replace('{id}', activeTaskID || '') }}
              </h5>
              <div class="small text-muted">
                {{ t('deploys_page.detail_live_hint') }}
              </div>
            </div>
            <div class="d-flex align-items-center gap-2">
              <button
                type="button"
                class="btn btn-sm btn-outline-secondary"
                @click="refreshDetail()"
                :disabled="detailLoading || detailRefreshing || !activeTaskID"
              >
                <i
                  class="bi me-1"
                  :class="
                    detailRefreshing
                      ? 'bi-arrow-repeat fm-spin'
                      : 'bi-arrow-clockwise'
                  "
                ></i>
                {{ t('common.refresh') }}
              </button>
              <button
                type="button"
                class="btn-close"
                data-bs-dismiss="modal"
              ></button>
            </div>
          </div>

          <div class="modal-body">
            <div
              v-if="detailLoading && !detail"
              class="d-flex flex-column align-items-center justify-content-center text-center text-muted py-5"
            >
              <div class="spinner-border text-primary mb-3" role="status"></div>
              <div>{{ t('deploys_page.detail_loading') }}</div>
            </div>

            <template v-else-if="detail">
              <section class="fm-deploy-live-shell mb-4">
                <div
                  class="d-flex flex-wrap justify-content-between align-items-start gap-3"
                >
                  <div>
                    <div class="fm-live-kicker">
                      {{ t('deploys_page.detail_live_title') }}
                    </div>
                    <div class="h5 mb-1">{{ deployConfigLabel(detail.task) }}</div>
                    <div class="text-muted small">
                      {{ deployScopeText(detail.task) }}
                    </div>
                  </div>
                  <div class="text-start text-md-end">
                    <span
                      class="badge rounded-pill"
                      :class="statusClass(taskDisplayStatus(detail.task))"
                    >
                      {{ statusText(taskDisplayStatus(detail.task)) }}
                    </span>
                    <div class="small text-muted mt-2">
                      {{ detailStageText(detail.task) }}
                    </div>
                  </div>
                </div>

                <div
                  v-if="detailError"
                  class="alert alert-warning border-0 mt-3 mb-0 py-2"
                >
                  <i class="bi bi-exclamation-triangle me-2"></i>
                  {{ detailError }}
                </div>

                <div class="fm-deploy-flow">
                  <div class="fm-deploy-flow__rail">
                    <template
                      v-for="(stage, index) in detailFlowStages"
                      :key="stage.key"
                    >
                      <article
                        class="fm-deploy-flow__stage"
                        :class="`is-${stage.state}`"
                      >
                        <div class="fm-deploy-flow__icon">
                          <i
                            class="bi"
                            :class="[stage.icon, stage.spin ? 'fm-spin' : '']"
                          ></i>
                        </div>
                        <div class="fm-deploy-flow__content">
                          <div class="fm-deploy-flow__label">{{ stage.label }}</div>
                          <div class="fm-deploy-flow__meta">{{ stage.meta }}</div>
                        </div>
                      </article>
                      <div
                        v-if="index < detailFlowStages.length - 1"
                        class="fm-deploy-flow__connector"
                        :class="
                          flowConnectorClass(
                            stage,
                            detailFlowStages[index + 1],
                          )
                        "
                      >
                        <span class="fm-deploy-flow__connector-line"></span>
                      </div>
                    </template>
                  </div>
                </div>

                <div class="progress fm-detail-progress mt-3">
                  <div
                    class="progress-bar bg-success"
                    :style="{ width: detailSuccessPct + '%' }"
                  ></div>
                  <div
                    class="progress-bar bg-danger"
                    :style="{ width: detailFailPct + '%' }"
                  ></div>
                  <div
                    class="progress-bar fm-progress-pending"
                    :style="{ width: detailPendingPct + '%' }"
                  ></div>
                </div>

                <div class="row g-3 mt-1">
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label">
                        {{ t('deploys_page.total_nodes_label') }}
                      </div>
                      <div class="fm-detail-stat-value">{{ detailStats.total }}</div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-success">
                        {{ t('deploys_page.success_nodes') }}
                      </div>
                      <div class="fm-detail-stat-value">{{ detailStats.success }}</div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-danger">
                        {{ t('deploys_page.failed_nodes') }}
                      </div>
                      <div class="fm-detail-stat-value">{{ detailStats.failed }}</div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-info-emphasis">
                        {{ t('deploys_page.pending_nodes') }}
                      </div>
                      <div class="fm-detail-stat-value">{{ detailStats.pending }}</div>
                    </div>
                  </div>
                </div>

                <div
                  class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-3 pt-3 border-top"
                >
                  <div class="d-flex flex-wrap gap-2">
                    <span class="badge text-bg-light border">
                      {{
                        detailAutoRefreshEnabled
                          ? t('deploys_page.auto_refresh_on').replace(
                              '{seconds}',
                              String(pollIntervalSeconds),
                            )
                          : t('deploys_page.auto_refresh_off')
                      }}
                    </span>
                    <span class="badge text-bg-light border">
                      {{
                        t('deploys_page.last_refresh').replace(
                          '{time}',
                          formatTime(detailLastRefreshedAt),
                        )
                      }}
                    </span>
                  </div>
                  <div class="small text-muted d-flex flex-wrap gap-3">
                    <span>{{ t('deploys_page.created_at') }} {{ formatTime(detail.task.created_at) }}</span>
                    <span>{{ t('deploys_page.started_at') }} {{ formatTime(detail.task.started_at) }}</span>
                    <span>{{ t('deploys_page.finished_at') }} {{ formatTime(detail.task.finished_at) }}</span>
                  </div>
                </div>
              </section>

              <div class="table-responsive">
                <table class="table table-sm align-middle">
                  <thead>
                    <tr>
                      <th>{{ t('common.node') }}</th>
                      <th>IP</th>
                      <th>{{ t('common.cluster') }}</th>
                      <th>{{ t('status') }}</th>
                      <th>{{ t('deploys_page.message') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="record in detail.records" :key="record.id">
                      <td>
                        <div class="fw-semibold">{{ record.node?.hostname || record.node_id }}</div>
                        <div class="small text-muted">{{ record.node?.node_uid || '-' }}</div>
                      </td>
                      <td>{{ record.node?.ip_address || '-' }}</td>
                      <td>
                        {{
                          record.node?.cluster?.alias ||
                          record.node?.cluster?.name ||
                          t('nodes_page.unassigned')
                        }}
                      </td>
                      <td>
                        <span class="badge" :class="statusClass(recordDisplayStatus(record))">{{
                          statusText(recordDisplayStatus(record))
                        }}</span>
                      </td>
                      <td class="small">{{ record.message || '-' }}</td>
                    </tr>
                    <tr v-if="!detail.records?.length">
                      <td colspan="5" class="text-center text-muted py-4">
                        {{ t('deploys_page.no_records') }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <nav v-if="detailTotal > detailPageSize" class="mt-3">
                <ul class="pagination justify-content-center mb-0">
                  <li class="page-item" :class="{ disabled: detailPage <= 1 }">
                    <a
                      class="page-link"
                      href="#"
                      @click.prevent="changeDetailPage(detailPage - 1)"
                      >{{ t('common.previous') }}</a
                    >
                  </li>
                  <li class="page-item disabled">
                    <span class="page-link">{{ detailPage }} / {{ detailTotalPages() }}</span>
                  </li>
                  <li class="page-item" :class="{ disabled: detailPage >= detailTotalPages() }">
                    <a
                      class="page-link"
                      href="#"
                      @click.prevent="changeDetailPage(detailPage + 1)"
                      >{{ t('common.next') }}</a
                    >
                  </li>
                </ul>
              </nav>
            </template>
          </div>

          <div class="modal-footer border-0 pt-0">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
              {{ t('common.close') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getDeploys, getDeploy } from '../api'
import { useI18n } from '../i18n'

const { t, dateLocale } = useI18n()
const route = useRoute()
const router = useRouter()

const tasks = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loadingAny = ref(false)

const detail = ref(null)
const detailTotal = ref(0)
const detailPage = ref(1)
const detailPageSize = 20
const activeTaskID = ref(null)
const detailLoading = ref(false)
const detailRefreshing = ref(false)
const detailError = ref('')
const detailLastRefreshedAt = ref(null)

let modal = null
let detailModalEl = null
let detailModalHiddenHandler = null
let pollTimer = null
const pollIntervalMs = 5000
const pollIntervalSeconds = pollIntervalMs / 1000

const activeTaskCount = computed(() =>
  tasks.value.filter((task) => isActiveTaskStatus(task.status)).length,
)
const completedTaskCount = computed(() =>
  tasks.value.filter((task) => task.status === 'completed').length,
)
const pageSuccessNodes = computed(() =>
  tasks.value.reduce((sum, task) => sum + Number(task.success_count || 0), 0),
)
const pageFailedNodes = computed(() =>
  tasks.value.reduce((sum, task) => sum + Number(task.fail_count || 0), 0),
)
const pagePendingNodes = computed(() =>
  tasks.value.reduce((sum, task) => sum + pendingCount(task), 0),
)

const detailStats = computed(() => {
  const task = detail.value?.task
  const totalNodes = Number(task?.total_nodes || 0)
  const success = Number(task?.success_count || 0)
  const failed = Number(task?.fail_count || 0)
  return {
    total: totalNodes,
    success,
    failed,
    pending: Math.max(totalNodes - success - failed, 0),
  }
})

const detailSuccessPct = computed(() =>
  percentage(detailStats.value.success, detailStats.value.total),
)
const detailFailPct = computed(() =>
  percentage(detailStats.value.failed, detailStats.value.total),
)
const detailPendingPct = computed(() =>
  percentage(detailStats.value.pending, detailStats.value.total),
)

const detailAutoRefreshEnabled = computed(() =>
  isActiveTaskStatus(detail.value?.task?.status),
)

const detailFlowStages = computed(() => {
  const task = detail.value?.task
  if (!task) return []

  const displayStatus = taskDisplayStatus(task)
  const resolvedCount = detailStats.value.success + detailStats.value.failed
  const remainingCount = detailStats.value.pending
  const started =
    Boolean(task.started_at) || isActiveTaskStatus(task.status) || isFinishedStatus(displayStatus)
  const deliverActive = task.status === 'running' && resolvedCount === 0
  const applyActive = task.status === 'running' && resolvedCount > 0 && remainingCount > 0
  const convergeActive = task.status === 'running' && remainingCount === 0

  return [
    {
      key: 'queued',
      label: t('deploys_page.flow_queued'),
      meta: task.status === 'pending'
        ? t('deploys_page.flow_meta_waiting')
        : t('deploys_page.flow_meta_dispatched'),
      icon: 'bi-hourglass-split',
      state: task.status === 'pending' ? 'active' : 'done',
      spin: false,
    },
    {
      key: 'deliver',
      label: t('deploys_page.flow_deliver'),
      meta: !started
        ? t('deploys_page.flow_meta_waiting')
        : deliverActive
          ? t('deploys_page.flow_meta_delivering')
          : t('deploys_page.flow_meta_delivered'),
      icon: 'bi-broadcast-pin',
      state: !started ? 'pending' : deliverActive ? 'active' : 'done',
      spin: deliverActive,
    },
    {
      key: 'apply',
      label: t('deploys_page.flow_apply'),
      meta: !started
        ? t('deploys_page.flow_meta_apply_pending')
        : interpolate(t('deploys_page.flow_meta_applying'), {
            count: remainingCount,
            total: detailStats.value.total,
          }),
      icon: 'bi-sliders2-vertical',
      state: !started ? 'pending' : applyActive ? 'active' : resolvedCount > 0 || isFinishedStatus(displayStatus) ? 'done' : 'pending',
      spin: applyActive,
    },
    {
      key: 'converge',
      label: t('deploys_page.flow_converge'),
      meta: !started
        ? t('deploys_page.flow_meta_converging_pending')
        : interpolate(t('deploys_page.flow_meta_converging'), {
            done: resolvedCount,
            total: detailStats.value.total,
          }),
      icon: 'bi-clipboard-data',
      state: !started ? 'pending' : convergeActive ? 'active' : resolvedCount > 0 || isFinishedStatus(displayStatus) ? 'done' : 'pending',
      spin: convergeActive,
    },
    {
      key: 'finish',
      label: t('deploys_page.flow_finish'),
      meta: isFinishedStatus(displayStatus)
        ? interpolate(t('deploys_page.flow_meta_finished'), {
            success: detailStats.value.success,
            failed: detailStats.value.failed,
          })
        : t('deploys_page.flow_meta_finish_pending'),
      icon:
        displayStatus === 'failed'
          ? 'bi-x-octagon'
          : displayStatus === 'partial'
            ? 'bi-exclamation-circle'
            : 'bi-check2-circle',
      state:
        displayStatus === 'completed'
          ? 'success'
          : displayStatus === 'failed'
            ? 'failed'
            : displayStatus === 'partial'
              ? 'warning'
              : 'pending',
      spin: false,
    },
  ]
})

function percentage(value, totalValue) {
  return totalValue ? (value / totalValue) * 100 : 0
}

function interpolate(template, replacements = {}) {
  return Object.entries(replacements).reduce(
    (result, [key, value]) =>
      result.replace(new RegExp(`\\{${key}\\}`, 'g'), String(value)),
    template,
  )
}

function pendingCount(task) {
  return Math.max(
    Number(task?.total_nodes || 0) -
      Number(task?.success_count || 0) -
      Number(task?.fail_count || 0),
    0,
  )
}

function successPct(task) {
  return percentage(Number(task?.success_count || 0), Number(task?.total_nodes || 0))
}

function failPct(task) {
  return percentage(Number(task?.fail_count || 0), Number(task?.total_nodes || 0))
}

function pendingPct(task) {
  return percentage(pendingCount(task), Number(task?.total_nodes || 0))
}

function isActiveTaskStatus(status) {
  return status === 'pending' || status === 'running'
}

function isFinishedStatus(status) {
  return status === 'completed' || status === 'failed' || status === 'partial'
}

function taskDisplayStatus(task) {
  if (!task) return 'pending'
  if (task.status === 'completed' && Number(task.fail_count || 0) > 0) {
    if (Number(task.fail_count || 0) >= Number(task.total_nodes || 0)) {
      return 'failed'
    }
    return 'partial'
  }
  return task.status
}

function recordDisplayStatus(record) {
  if (!record) return 'pending'
  if (record.status === 'success') return 'completed'
  return record.status
}

function statusClass(status) {
  return {
    'bg-success': status === 'completed' || status === 'success',
    'bg-warning text-dark': status === 'pending' || status === 'running' || status === 'partial',
    'bg-danger': status === 'failed',
  }
}

function statusText(status) {
  return {
    pending: t('deploys_page.pending'),
    running: t('deploys_page.running'),
    completed: t('deploys_page.completed'),
    failed: t('deploys_page.failed'),
    partial: t('deploys_page.partial'),
    success: t('deploys_page.completed'),
  }[status] || status
}

function deployConfigLabel(task) {
  return `${task?.config?.template?.name || '-'} v${task?.config?.version || '-'}`
}

function deployScopeText(task) {
  const scopeType = task?.scope || 'node'
  const labels = {
    node: t('deploys_page.scope_selected_nodes'),
    cluster: t('common.cluster'),
    region: t('common.region'),
    datacenter: t('common.datacenter'),
    environment: t('common.environment'),
  }
  const label = labels[scopeType] || scopeType
  if (scopeType === 'node' || !task?.scope_id) {
    return label
  }
  return `${label} #${task.scope_id}`
}

function detailStageText(task) {
  const displayStatus = taskDisplayStatus(task)
  if (displayStatus === 'pending') {
    return t('deploys_page.stage_pending')
  }
  if (displayStatus === 'running') {
    return t('deploys_page.stage_running')
  }
  if (displayStatus === 'failed') {
    return t('deploys_page.stage_failed')
  }
  if (displayStatus === 'partial') {
    return t('deploys_page.stage_partial')
  }
  return t('deploys_page.stage_completed')
}

function flowConnectorClass(currentStage, nextStage) {
  return {
    'is-filled':
      currentStage.state !== 'pending' && nextStage.state !== 'pending',
    'is-active':
      currentStage.state === 'active' || nextStage.state === 'active',
    'is-failed':
      nextStage.state === 'failed',
    'is-success':
      nextStage.state === 'success',
    'is-warning':
      nextStage.state === 'warning',
  }
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function totalPages() {
  return Math.max(1, Math.ceil(total.value / pageSize))
}

function detailTotalPages() {
  return Math.max(1, Math.ceil(detailTotal.value / detailPageSize))
}

function parseRouteTaskID() {
  const raw = route.query.task
  const parsed = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}

function setTaskQuery(taskID) {
  const nextQuery = { ...route.query, task: String(taskID) }
  router.replace({ path: '/deploys', query: nextQuery })
}

function clearTaskQuery() {
  if (!route.query.task) return
  const nextQuery = { ...route.query }
  delete nextQuery.task
  router.replace({ path: '/deploys', query: nextQuery })
}

function resetDetailState() {
  detail.value = null
  detailTotal.value = 0
  detailPage.value = 1
  activeTaskID.value = null
  detailLoading.value = false
  detailRefreshing.value = false
  detailError.value = ''
  detailLastRefreshedAt.value = null
}

function ensureModal() {
  if (!detailModalEl) {
    detailModalEl = document.getElementById('detailModal')
  }
  if (!modal && detailModalEl) {
    modal = new window.bootstrap.Modal(detailModalEl)
  }
  if (detailModalEl && !detailModalHiddenHandler) {
    detailModalHiddenHandler = () => {
      resetDetailState()
      clearTaskQuery()
    }
    detailModalEl.addEventListener('hidden.bs.modal', detailModalHiddenHandler)
  }
}

async function loadDeploys() {
  const { data } = await getDeploys({ page: page.value, page_size: pageSize })
  tasks.value = data.data || []
  total.value = data.total || 0
  page.value = data.page || page.value
}

async function refreshAll() {
  loadingAny.value = true
  try {
    await loadDeploys()
  } finally {
    loadingAny.value = false
  }
}

async function refreshDetail(options = {}) {
  if (!activeTaskID.value) return null

  if (options.background) {
    detailRefreshing.value = true
  } else {
    detailLoading.value = true
  }

  try {
    const { data } = await getDeploy(activeTaskID.value, {
      page: detailPage.value,
      page_size: detailPageSize,
    })
    if (activeTaskID.value !== data.task?.id) {
      return null
    }
    detail.value = data
    detailTotal.value = data.total || 0
    detailPage.value = data.page || detailPage.value
    detailError.value = ''
    detailLastRefreshedAt.value = new Date().toISOString()
    return data
  } catch (error) {
    if (!options.background) {
      detailError.value =
        error?.response?.data?.error || error?.message || t('common.request_failed')
    }
    return null
  } finally {
    if (options.background) {
      detailRefreshing.value = false
    } else {
      detailLoading.value = false
    }
  }
}

async function openDetail(taskLike, options = {}) {
  const taskID = Number(taskLike?.id || 0)
  if (!taskID) return

  if (activeTaskID.value !== taskID) {
    detailPage.value = 1
  }

  activeTaskID.value = taskID
  detailError.value = ''

  if (taskLike?.config) {
    detail.value = {
      task: taskLike,
      records: detail.value?.task?.id === taskID ? detail.value.records || [] : [],
    }
  }

  ensureModal()
  modal?.show()
  if (options.syncRoute !== false) {
    setTaskQuery(taskID)
  }
  await refreshDetail()
}

async function viewDetail(task) {
  await openDetail(task)
}

async function changePage(nextPage) {
  if (nextPage < 1 || nextPage > totalPages() || nextPage === page.value) return
  page.value = nextPage
  await loadDeploys()
}

async function changeDetailPage(nextPage) {
  if (
    !activeTaskID.value ||
    nextPage < 1 ||
    nextPage > detailTotalPages() ||
    nextPage === detailPage.value
  ) return
  detailPage.value = nextPage
  await refreshDetail()
}

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(async () => {
    if (
      !tasks.value.some((task) => isActiveTaskStatus(task.status)) &&
      !detailAutoRefreshEnabled.value
    ) {
      return
    }
    try {
      await loadDeploys()
      if (activeTaskID.value && detailAutoRefreshEnabled.value) {
        await refreshDetail({ background: true })
      }
    } catch {}
  }, pollIntervalMs)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  await refreshAll()
  const routeTaskID = parseRouteTaskID()
  if (routeTaskID) {
    await openDetail({ id: routeTaskID }, { syncRoute: false })
  }
  startPolling()
})

onUnmounted(() => {
  stopPolling()
  if (detailModalEl && detailModalHiddenHandler) {
    detailModalEl.removeEventListener('hidden.bs.modal', detailModalHiddenHandler)
  }
})
</script>

<style scoped>
.fm-deploy-stat-card {
  height: 100%;
  padding: 1rem 1.1rem;
  border-radius: 1rem;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background:
    radial-gradient(circle at top right, rgba(251, 191, 36, 0.18), transparent 34%),
    linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.05);
}

.fm-deploy-stat-card__label {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--bs-secondary-color);
}

.fm-deploy-stat-card__value {
  margin-top: 0.4rem;
  font-size: 2rem;
  font-weight: 700;
  line-height: 1;
  color: var(--bs-body-color);
}

.fm-deploy-stat-card__meta {
  margin-top: 0.45rem;
  font-size: 0.82rem;
  color: #64748b;
}

.fm-task-row-active > td {
  background: linear-gradient(90deg, rgba(251, 191, 36, 0.12), rgba(14, 165, 233, 0.03));
}

.fm-deploy-live-shell {
  padding: 1.25rem;
  border: 1px solid rgba(251, 191, 36, 0.18);
  border-radius: 1rem;
  background:
    radial-gradient(circle at top left, rgba(250, 204, 21, 0.16), transparent 34%),
    radial-gradient(circle at top right, rgba(34, 211, 238, 0.16), transparent 30%),
    linear-gradient(180deg, rgba(255, 251, 235, 0.95), rgba(255, 255, 255, 1));
}

.fm-live-kicker {
  margin-bottom: 0.35rem;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--bs-secondary-color);
}

.fm-deploy-flow {
  margin-top: 1rem;
}

.fm-deploy-flow__rail {
  display: flex;
  align-items: stretch;
  gap: 0.75rem;
  overflow-x: auto;
  padding-bottom: 0.35rem;
}

.fm-deploy-flow__stage {
  min-width: 176px;
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.06);
}

.fm-deploy-flow__icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  font-size: 1.15rem;
  color: #64748b;
  background: rgba(148, 163, 184, 0.16);
}

.fm-deploy-flow__content {
  min-width: 0;
}

.fm-deploy-flow__label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
}

.fm-deploy-flow__meta {
  margin-top: 0.4rem;
  font-size: 0.88rem;
  line-height: 1.4;
  color: #0f172a;
}

.fm-deploy-flow__stage.is-done {
  border-color: rgba(8, 145, 178, 0.18);
}

.fm-deploy-flow__stage.is-done .fm-deploy-flow__icon {
  color: #0f766e;
  background: rgba(45, 212, 191, 0.18);
}

.fm-deploy-flow__stage.is-active {
  transform: translateY(-2px);
  border-color: rgba(245, 158, 11, 0.32);
  background:
    radial-gradient(circle at top left, rgba(253, 224, 71, 0.28), transparent 42%),
    rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 44px rgba(245, 158, 11, 0.16);
}

.fm-deploy-flow__stage.is-active .fm-deploy-flow__icon {
  color: #b45309;
  background: linear-gradient(135deg, rgba(253, 224, 71, 0.42), rgba(34, 211, 238, 0.16));
  animation: fm-flow-pulse 1.8s ease-in-out infinite;
}

.fm-deploy-flow__stage.is-pending {
  border-style: dashed;
  opacity: 0.82;
}

.fm-deploy-flow__stage.is-success {
  border-color: rgba(34, 197, 94, 0.25);
  background:
    radial-gradient(circle at top left, rgba(134, 239, 172, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-deploy-flow__stage.is-success .fm-deploy-flow__icon {
  color: #15803d;
  background: rgba(74, 222, 128, 0.18);
}

.fm-deploy-flow__stage.is-failed {
  border-color: rgba(239, 68, 68, 0.28);
  background:
    radial-gradient(circle at top left, rgba(252, 165, 165, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-deploy-flow__stage.is-failed .fm-deploy-flow__icon {
  color: #b91c1c;
  background: rgba(248, 113, 113, 0.18);
}

.fm-deploy-flow__stage.is-warning {
  border-color: rgba(245, 158, 11, 0.28);
  background:
    radial-gradient(circle at top left, rgba(253, 224, 71, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-deploy-flow__stage.is-warning .fm-deploy-flow__icon {
  color: #b45309;
  background: rgba(251, 191, 36, 0.2);
}

.fm-deploy-flow__connector {
  position: relative;
  flex: 0 0 54px;
  display: flex;
  align-items: center;
}

.fm-deploy-flow__connector-line {
  position: relative;
  display: block;
  width: 100%;
  height: 4px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.28);
}

.fm-deploy-flow__connector-line::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  transform: scaleX(0);
  transform-origin: left center;
  background: linear-gradient(90deg, #f59e0b, #0ea5e9);
  transition: transform 0.25s ease;
}

.fm-deploy-flow__connector::after {
  content: '';
  position: absolute;
  right: -3px;
  top: 50%;
  transform: translateY(-50%);
  border-left: 8px solid rgba(148, 163, 184, 0.45);
  border-top: 6px solid transparent;
  border-bottom: 6px solid transparent;
}

.fm-deploy-flow__connector.is-filled .fm-deploy-flow__connector-line::after,
.fm-deploy-flow__connector.is-active .fm-deploy-flow__connector-line::after,
.fm-deploy-flow__connector.is-failed .fm-deploy-flow__connector-line::after,
.fm-deploy-flow__connector.is-success .fm-deploy-flow__connector-line::after,
.fm-deploy-flow__connector.is-warning .fm-deploy-flow__connector-line::after {
  transform: scaleX(1);
}

.fm-deploy-flow__connector.is-active .fm-deploy-flow__connector-line::after {
  background: linear-gradient(90deg, #facc15, #f59e0b, #0ea5e9);
  background-size: 200% 100%;
  animation: fm-flow-glide 1.6s linear infinite;
}

.fm-deploy-flow__connector.is-active::after {
  border-left-color: #f59e0b;
}

.fm-deploy-flow__connector.is-failed .fm-deploy-flow__connector-line::after {
  background: linear-gradient(90deg, #f87171, #ef4444);
}

.fm-deploy-flow__connector.is-failed::after {
  border-left-color: #ef4444;
}

.fm-deploy-flow__connector.is-success .fm-deploy-flow__connector-line::after {
  background: linear-gradient(90deg, #4ade80, #22c55e);
}

.fm-deploy-flow__connector.is-success::after {
  border-left-color: #22c55e;
}

.fm-deploy-flow__connector.is-warning .fm-deploy-flow__connector-line::after {
  background: linear-gradient(90deg, #facc15, #f59e0b);
}

.fm-deploy-flow__connector.is-warning::after {
  border-left-color: #f59e0b;
}

.fm-detail-progress {
  height: 0.8rem;
  border-radius: 999px;
  background: rgba(108, 117, 125, 0.12);
}

.fm-progress-pending {
  background: rgba(148, 163, 184, 0.44);
}

.fm-detail-stat {
  height: 100%;
  padding: 0.9rem 1rem;
  border-radius: 0.9rem;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(33, 37, 41, 0.08);
}

.fm-detail-stat-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--bs-secondary-color);
}

.fm-detail-stat-value {
  margin-top: 0.3rem;
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1;
  color: var(--bs-body-color);
}

.fm-spin {
  animation: fm-spin 1s linear infinite;
}

@keyframes fm-flow-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.2);
  }

  50% {
    box-shadow: 0 0 0 10px rgba(245, 158, 11, 0);
  }
}

@keyframes fm-flow-glide {
  from {
    background-position: 0% 0%;
  }

  to {
    background-position: 200% 0%;
  }
}

@keyframes fm-spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 767.98px) {
  .fm-deploy-flow__stage {
    min-width: 164px;
  }

  .fm-deploy-flow__connector {
    flex-basis: 40px;
  }
}
</style>
