<template>
  <div>
    <h4 class="mb-4">{{ t('audit_page.title') }}</h4>
    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>{{ t('dashboard.time') }}</th>
              <th>{{ t('dashboard.user') }}</th>
              <th>{{ t('dashboard.action') }}</th>
              <th>{{ t('dashboard.resource') }}</th>
              <th>{{ t('audit_page.detail') }}</th>
              <th>IP</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="log in logs" :key="log.id">
              <tr>
                <td>{{ formatTime(log.created_at) }}</td>
                <td>{{ log.username || '-' }}</td>
                <td><span class="badge bg-info">{{ log.action }}</span></td>
                <td>{{ log.resource }}</td>
                <td style="max-width: 320px;">
                  <div v-if="hasAgentPolicyDiff(log)">
                    <div class="fw-semibold">{{ summarizeAuditDetail(log) }}</div>
                    <button class="btn btn-sm btn-outline-secondary mt-2" @click="toggleExpanded(log.id)">
                      {{ expanded[log.id] ? t('audit_page.hide_diff') : t('audit_page.show_diff') }}
                    </button>
                  </div>
                  <div v-else class="text-truncate">{{ summarizeAuditDetail(log) }}</div>
                </td>
                <td>{{ log.ip }}</td>
              </tr>
              <tr v-if="expanded[log.id] && hasAgentPolicyDiff(log)" class="table-light">
                <td colspan="6">
                  <div class="p-2">
                    <div class="small text-muted mb-2">
                      {{ t('audit_page.operation') }}: {{ parsedAuditDetail(log)?.operation || '-' }}
                    </div>
                    <div class="row g-3">
                      <div class="col-md-6">
                        <div class="border rounded p-3 h-100">
                          <div class="fw-semibold mb-2">{{ t('audit_page.before') }}</div>
                          <pre class="fm-audit-json mb-0">{{ formatJSON(parsedAuditDetail(log)?.before) }}</pre>
                        </div>
                      </div>
                      <div class="col-md-6">
                        <div class="border rounded p-3 h-100">
                          <div class="fw-semibold mb-2">{{ t('audit_page.after') }}</div>
                          <pre class="fm-audit-json mb-0">{{ formatJSON(parsedAuditDetail(log)?.after) }}</pre>
                        </div>
                      </div>
                      <div class="col-12">
                        <div class="border rounded p-3">
                          <div class="fw-semibold mb-2">{{ t('audit_page.changes') }}</div>
                          <pre class="fm-audit-json mb-0">{{ formatJSON(parsedAuditDetail(log)?.changes) }}</pre>
                        </div>
                      </div>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a class="page-link" href="#" @click.prevent="page--; loadLogs()">{{ t('common.previous') }}</a>
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= Math.ceil(total / pageSize) }">
          <a class="page-link" href="#" @click.prevent="page++; loadLogs()">{{ t('common.next') }}</a>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { getAuditLogs } from '../api'
import { useI18n } from '../i18n'

const logs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 50
const expanded = reactive({})
const { t, dateLocale } = useI18n()

function formatTime(t) { return t ? new Date(t).toLocaleString(dateLocale.value) : '-' }

function parsedAuditDetail(log) {
  if (log.resource_type !== 'agent_policy') return null
  try {
    return JSON.parse(log.detail)
  } catch {
    return null
  }
}

function hasAgentPolicyDiff(log) {
  return !!parsedAuditDetail(log)
}

function summarizeAuditDetail(log) {
  const parsed = parsedAuditDetail(log)
  if (parsed?.operation) {
    return `${log.resource_type || 'resource'} ${parsed.operation}`
  }
  return log.detail || '-'
}

function toggleExpanded(id) {
  expanded[id] = !expanded[id]
}

function formatJSON(value) {
  if (value === null || value === undefined) return '-'
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

async function loadLogs() {
  const { data } = await getAuditLogs({ page: page.value, page_size: pageSize })
  logs.value = data.data || []
  total.value = data.total
}

onMounted(loadLogs)
</script>

<style scoped>
.fm-audit-json {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.8rem;
  max-height: 320px;
  overflow: auto;
  background: #0f172a;
  color: #e2e8f0;
  padding: 0.85rem;
  border-radius: 0.5rem;
}
</style>
