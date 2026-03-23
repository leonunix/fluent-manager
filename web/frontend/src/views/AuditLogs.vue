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
                    <div class="small text-muted mb-3">
                      {{ t('audit_page.operation') }}: {{ parsedAuditDetail(log)?.operation || '-' }}
                    </div>
                    <div v-if="diffEntries(log).length" class="row g-3">
                      <div v-for="entry in diffEntries(log)" :key="entry.key" class="col-xl-6">
                        <div class="fm-audit-diff-card h-100">
                          <div class="d-flex justify-content-between align-items-center gap-2 mb-3">
                            <div class="fw-semibold">{{ entry.label }}</div>
                            <span v-if="entry.group" class="badge text-bg-light">{{ entry.group }}</span>
                          </div>
                          <div class="row g-3">
                            <div class="col-6">
                              <div class="small text-muted mb-2">{{ t('audit_page.before') }}</div>
                              <div class="fm-audit-diff-value">{{ formatAuditValue(entry.before, entry.key) }}</div>
                            </div>
                            <div class="col-6">
                              <div class="small text-muted mb-2">{{ t('audit_page.after') }}</div>
                              <div class="fm-audit-diff-value">{{ formatAuditValue(entry.after, entry.key) }}</div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-muted small">{{ t('audit_page.no_structured_changes') }}</div>
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
    return `${t('nav.agent_policies')} ${parsed.operation}`
  }
  return log.detail || '-'
}

function toggleExpanded(id) {
  expanded[id] = !expanded[id]
}

function prettifyFieldName(key) {
  return key
    .split('.')
    .pop()
    .split('_')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

function formatFieldLabel(key) {
  const labels = {
    name: t('common.name'),
    description: t('common.description'),
    scope_type: t('agent_policies_page.scope'),
    environment_id: t('common.environment'),
    cluster_id: t('common.cluster'),
    label_selector: t('agent_policies_page.scope_label_selector'),
    priority: t('agent_policies_page.priority'),
    is_enabled: t('agent_policies_page.enabled'),
  }

  if (key.startsWith('settings.')) {
    return `${t('audit_page.settings_group')} / ${prettifyFieldName(key)}`
  }
  return labels[key] || prettifyFieldName(key)
}

function scopeTypeLabel(value) {
  if (!value) return t('audit_page.no_value')
  return t(`agent_policies_page.scope_${value}`)
}

function formatAuditValue(value, key = '') {
  if (value === null || value === undefined || value === '') return t('audit_page.no_value')
  if (key === 'scope_type') return scopeTypeLabel(value)
  if (typeof value === 'boolean') return value ? t('yes') : t('no')
  if (Array.isArray(value)) return value.length ? value.join(', ') : t('audit_page.no_value')
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value, null, 2)
    } catch {
      return String(value)
    }
  }
  return String(value)
}

function areEqual(before, after) {
  return JSON.stringify(before) === JSON.stringify(after)
}

function flattenSettingsChange(change) {
  const before = change?.before || {}
  const after = change?.after || {}
  const keys = new Set([...Object.keys(before), ...Object.keys(after)])

  return [...keys]
    .filter((key) => !areEqual(before[key], after[key]))
    .map((key) => ({
      key: `settings.${key}`,
      label: formatFieldLabel(`settings.${key}`),
      group: t('audit_page.settings_group'),
      before: before[key],
      after: after[key],
    }))
}

function flattenSnapshot(snapshot, direction) {
  if (!snapshot) return []

  const entries = []
  for (const [key, value] of Object.entries(snapshot)) {
    if (key === 'id') continue
    if (key === 'settings' && value && typeof value === 'object') {
      for (const [settingKey, settingValue] of Object.entries(value)) {
        entries.push({
          key: `settings.${settingKey}`,
          label: formatFieldLabel(`settings.${settingKey}`),
          group: t('audit_page.settings_group'),
          before: direction === 'before' ? settingValue : null,
          after: direction === 'after' ? settingValue : null,
        })
      }
      continue
    }

    entries.push({
      key,
      label: formatFieldLabel(key),
      group: '',
      before: direction === 'before' ? value : null,
      after: direction === 'after' ? value : null,
    })
  }

  return entries
}

function diffEntries(log) {
  const parsed = parsedAuditDetail(log)
  if (!parsed) return []

  if (parsed.changes && typeof parsed.changes === 'object') {
    const entries = []
    for (const [key, change] of Object.entries(parsed.changes)) {
      if (key === 'settings') {
        entries.push(...flattenSettingsChange(change))
        continue
      }
      entries.push({
        key,
        label: formatFieldLabel(key),
        group: '',
        before: change?.before,
        after: change?.after,
      })
    }
    if (entries.length) return entries
  }

  if (parsed.operation === 'create') return flattenSnapshot(parsed.after, 'after')
  if (parsed.operation === 'delete') return flattenSnapshot(parsed.before, 'before')
  return []
}

async function loadLogs() {
  const { data } = await getAuditLogs({ page: page.value, page_size: pageSize })
  logs.value = data.data || []
  total.value = data.total
}

onMounted(loadLogs)
</script>

<style scoped>
.fm-audit-diff-card {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 14px;
  padding: 1rem;
  background: #fff;
}

.fm-audit-diff-value {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.8rem;
  min-height: 72px;
  max-height: 240px;
  overflow: auto;
  background: #0f172a;
  color: #e2e8f0;
  padding: 0.85rem;
  border-radius: 0.5rem;
}
</style>
