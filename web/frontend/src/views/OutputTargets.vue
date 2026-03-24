<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('output_targets.title') }}</h4>
        <div class="text-muted">{{ t('output_targets.subtitle') }}</div>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <i class="bi bi-plus-lg me-1"></i>{{ t('output_targets.create') }}
      </button>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('output_targets.total') }}</div>
            <div class="fs-3 fw-bold">{{ targets.length }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('output_targets.target_type_coverage') }}</div>
            <div class="fs-3 fw-bold">{{ targetTypeCount }}</div>
            <div class="small text-muted mt-2">{{ targetTypeList }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('output_targets.runtime_coverage') }}</div>
            <div class="fs-3 fw-bold">{{ runtimeCoverageCount }}</div>
            <div class="small text-muted mt-2">{{ runtimeCoverageList }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('common.type') }}</th>
                <th>{{ t('common.runtime') }}</th>
                <th>{{ t('output_targets.endpoint') }}</th>
                <th>{{ t('output_targets.summary') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="target in targets" :key="target.id">
                <td>
                  <div class="fw-semibold">{{ target.name }}</div>
                  <div class="small text-muted">{{ target.description || t('common.no_description') }}</div>
                </td>
                <td><span class="badge text-bg-secondary">{{ target.target_type }}</span></td>
                <td><span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(target.fluent_type) }}</span></td>
                <td>
                  <div><code>{{ target.endpoint || '-' }}</code></div>
                  <div v-if="targetSummary(target).primary" class="small text-muted mt-1">{{ targetSummary(target).primary }}</div>
                </td>
                <td>
                  <div class="d-flex flex-wrap gap-2">
                    <span
                      v-for="chip in targetSummary(target).chips"
                      :key="chip"
                      class="badge rounded-pill text-bg-light"
                    >
                      {{ chip }}
                    </span>
                    <span v-if="!targetSummary(target).chips.length" class="text-muted small">{{ t('output_targets.no_summary') }}</span>
                  </div>
                </td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(target)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="handleDelete(target)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!targets.length">
                <td colspan="6" class="text-center text-muted py-4">{{ t('output_targets.no_targets') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal fade" id="outputTargetModal" tabindex="-1">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editingId ? t('output_targets.edit_title') : t('output_targets.create_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3 mb-3">
              <div class="col-md-5">
                <label class="form-label">{{ t('common.name') }}</label>
                <input v-model="form.name" type="text" class="form-control" placeholder="opensearch-prod">
              </div>
              <div class="col-md-3">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="form.fluent_type" class="form-select">
                  <option value="shared">Shared</option>
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{ t('common.type') }}</label>
                <select v-model="form.target_type" class="form-select">
                  <option v-for="option in targetTypes" :key="option" :value="option">{{ option }}</option>
                </select>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <input v-model="form.description" type="text" class="form-control">
            </div>

            <div class="row g-4">
              <div class="col-lg-7">
                <div class="card border-0 bg-light-subtle h-100">
                  <div class="card-header bg-transparent d-flex justify-content-between align-items-center">
                    <div>
                      <div class="fw-semibold">{{ t('output_targets.guided_settings') }}</div>
                      <div class="small text-muted mt-1">{{ t('output_targets.guided_settings_hint') }}</div>
                    </div>
                    <div class="btn-group btn-group-sm">
                      <button
                        type="button"
                        class="btn"
                        :class="settingsMode === 'guided' ? 'btn-primary' : 'btn-outline-primary'"
                        @click="setSettingsMode('guided')"
                      >
                        {{ t('output_targets.guided_mode') }}
                      </button>
                      <button
                        type="button"
                        class="btn"
                        :class="settingsMode === 'json' ? 'btn-primary' : 'btn-outline-primary'"
                        @click="setSettingsMode('json')"
                      >
                        {{ t('output_targets.json_mode') }}
                      </button>
                    </div>
                  </div>
                  <div class="card-body">
                    <div class="mb-3">
                      <label class="form-label">{{ t('output_targets.endpoint') }}</label>
                      <input v-model="form.endpoint" type="text" class="form-control" :placeholder="endpointPlaceholder">
                    </div>

                    <div v-if="settingsMode === 'guided'">
                      <div class="row g-3">
                        <div v-for="field in currentFields" :key="field.key" class="col-md-6">
                          <label class="form-label">{{ field.label }}</label>
                          <select
                            v-if="field.type === 'boolean'"
                            v-model="settingsForm[field.key]"
                            class="form-select"
                          >
                            <option :value="true">true</option>
                            <option :value="false">false</option>
                          </select>
                          <select
                            v-else-if="field.type === 'select'"
                            v-model="settingsForm[field.key]"
                            class="form-select"
                          >
                            <option v-for="option in field.options || []" :key="String(option.value)" :value="option.value">
                              {{ option.label }}
                            </option>
                          </select>
                          <textarea
                            v-else-if="field.type === 'textarea'"
                            v-model="settingsForm[field.key]"
                            class="form-control font-monospace"
                            rows="4"
                            :placeholder="field.placeholder || ''"
                          ></textarea>
                          <input
                            v-else
                            v-model="settingsForm[field.key]"
                            :type="field.type === 'number' ? 'number' : (field.type === 'password' ? 'password' : 'text')"
                            class="form-control"
                            :placeholder="field.placeholder || ''"
                          >
                          <div v-if="field.hint" class="small text-muted mt-1">{{ field.hint }}</div>
                        </div>
                      </div>
                      <div class="small text-muted mt-3">{{ t('output_targets.guided_mode_footer') }}</div>
                    </div>

                    <div v-else>
                      <textarea
                        v-model="form.settings"
                        class="form-control font-monospace"
                        rows="14"
                        :placeholder="settingsPlaceholder"
                      ></textarea>
                      <div class="small text-muted mt-2">{{ t('output_targets.settings_hint') }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="col-lg-5">
                <div class="card border-0 shadow-sm h-100">
                  <div class="card-header bg-white">
                    <div class="fw-semibold">{{ t('output_targets.preview_title') }}</div>
                    <div class="small text-muted mt-1">{{ t('output_targets.preview_hint') }}</div>
                  </div>
                  <div class="card-body">
                    <div class="d-flex flex-wrap gap-2 mb-3">
                      <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(form.fluent_type) }}</span>
                      <span class="badge text-bg-secondary">{{ form.target_type }}</span>
                    </div>
                    <div class="mb-3">
                      <div class="small text-muted mb-1">{{ t('output_targets.endpoint') }}</div>
                      <code>{{ form.endpoint || '-' }}</code>
                    </div>
                    <div class="mb-0">
                      <div class="small text-muted mb-1">{{ t('output_targets.settings') }}</div>
                      <pre class="fm-settings-preview">{{ formattedSettingsPreview }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveTarget">
              {{ editingId ? t('save') : t('output_targets.create') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { createOutputTarget, deleteOutputTarget, getOutputTargets, updateOutputTarget } from '../api'
import { useI18n } from '../i18n'
import { summarizeOutputTarget } from '../utils/output_targets'

const { t } = useI18n()
const targets = ref([])
const editingId = ref(null)
const settingsMode = ref('guided')
const targetTypes = ['opensearch', 'loki', 'kafka', 'http', 's3', 'stdout', 'custom']

const fieldCatalog = {
  opensearch: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'host', label: 'Host', type: 'text', placeholder: 'opensearch.internal' },
    { key: 'port', label: 'Port', type: 'number', placeholder: '9200' },
    { key: 'index', label: 'Index', type: 'text', placeholder: 'logs-%Y.%m.%d' },
    { key: 'http_user', label: 'HTTP User', type: 'text', placeholder: 'admin' },
    { key: 'http_password', label: 'HTTP Password', type: 'password', placeholder: 'changeme' },
    { key: 'tls', label: 'TLS', type: 'boolean' },
    { key: 'replace_dots', label: 'Replace Dots', type: 'boolean' },
  ],
  loki: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'host', label: 'Host', type: 'text', placeholder: 'loki.internal' },
    { key: 'port', label: 'Port', type: 'number', placeholder: '3100' },
    { key: 'tenant_id', label: 'Tenant ID', type: 'text', placeholder: 'platform' },
    { key: 'labels', label: 'Labels', type: 'text', placeholder: 'job=fluent-manager,team=platform' },
    { key: 'line_format', label: 'Line Format', type: 'select', options: [{ label: 'json', value: 'json' }, { label: 'key_value', value: 'key_value' }] },
  ],
  kafka: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'brokers', label: 'Brokers', type: 'text', placeholder: 'kafka-1:9092,kafka-2:9092' },
    { key: 'topics', label: 'Topics', type: 'text', placeholder: 'logs.platform' },
    { key: 'format', label: 'Format', type: 'select', options: [{ label: 'json', value: 'json' }, { label: 'msgpack', value: 'msgpack' }] },
    { key: 'timestamp_key', label: 'Timestamp Key', type: 'text', placeholder: '@timestamp' },
  ],
  http: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'host', label: 'Host', type: 'text', placeholder: 'collector.internal' },
    { key: 'port', label: 'Port', type: 'number', placeholder: '8080' },
    { key: 'uri', label: 'URI', type: 'text', placeholder: '/ingest' },
    { key: 'format', label: 'Format', type: 'select', options: [{ label: 'json', value: 'json' }, { label: 'json_lines', value: 'json_lines' }, { label: 'msgpack', value: 'msgpack' }] },
    { key: 'header_authorization', label: 'Authorization Header', type: 'text', placeholder: 'Bearer <token>' },
  ],
  s3: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'bucket', label: 'Bucket', type: 'text', placeholder: 'company-log-archive' },
    { key: 'region', label: 'Region', type: 'text', placeholder: 'ap-northeast-1' },
    { key: 'path', label: 'Path Prefix', type: 'text', placeholder: 'fluent/%Y/%m/%d/' },
    { key: 'compression', label: 'Compression', type: 'select', options: [{ label: 'gzip', value: 'gzip' }, { label: 'arrow', value: 'arrow' }, { label: 'none', value: 'none' }] },
  ],
  stdout: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'format', label: 'Format', type: 'select', options: [{ label: 'json_lines', value: 'json_lines' }, { label: 'json', value: 'json' }] },
  ],
  custom: [
    { key: 'match', label: 'Match', type: 'text', placeholder: '*' },
    { key: 'plugin', label: 'Plugin', type: 'text', placeholder: 'custom_output' },
    { key: 'notes', label: 'Notes', type: 'textarea', placeholder: '{\n  "team": "platform"\n}' },
  ],
}

const form = reactive({
  name: '',
  description: '',
  fluent_type: 'shared',
  target_type: 'opensearch',
  endpoint: '',
  settings: '',
})

const settingsForm = reactive({})
const targetTypeCount = computed(() => new Set(targets.value.map((item) => item.target_type).filter(Boolean)).size)
const targetTypeList = computed(() => {
  const values = Array.from(new Set(targets.value.map((item) => item.target_type).filter(Boolean)))
  return values.length ? values.join(' / ') : '-'
})
const runtimeCoverageCount = computed(() => new Set(targets.value.map((item) => item.fluent_type).filter(Boolean)).size)
const runtimeCoverageList = computed(() => {
  const values = Array.from(new Set(targets.value.map((item) => item.fluent_type).filter(Boolean))).map((value) => runtimeLabel(value))
  return values.length ? values.join(' / ') : '-'
})
const currentFields = computed(() => fieldCatalog[form.target_type] || fieldCatalog.custom)
const endpointPlaceholder = computed(() => {
  if (form.target_type === 'opensearch') return 'https://opensearch.internal:9200'
  if (form.target_type === 'loki') return 'http://loki.internal:3100'
  if (form.target_type === 'http') return 'https://collector.internal/ingest'
  if (form.target_type === 'kafka') return 'kafka-1:9092,kafka-2:9092'
  if (form.target_type === 's3') return 's3://company-log-archive/fluent/'
  return 'destination endpoint'
})
const settingsPlaceholder = computed(() => JSON.stringify(defaultSettingsObject(form.target_type), null, 2))
const formattedSettingsPreview = computed(() => {
  try {
    return JSON.stringify(JSON.parse(form.settings || '{}'), null, 2)
  } catch {
    return form.settings || '{}'
  }
})

let modal = null

function ensureModal() {
  if (!modal) {
    modal = new window.bootstrap.Modal(document.getElementById('outputTargetModal'))
  }
}

function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || t('common.request_failed')
}

function runtimeLabel(value) {
  if (value === 'fluentbit') return 'Fluent Bit'
  if (value === 'fluentd') return 'Fluentd'
  if (value === 'shared') return 'Shared'
  return value || '-'
}

function shortSettings(value) {
  if (!value || value === '{}') return '{}'
  return value.length > 56 ? `${value.slice(0, 56)}...` : value
}

function targetSummary(target) {
  return summarizeOutputTarget(target)
}

function defaultSettingsObject(type) {
  if (type === 'opensearch') {
    return {
      match: '*',
      host: 'opensearch.internal',
      port: 9200,
      index: 'logs-%Y.%m.%d',
      http_user: 'admin',
      http_password: 'changeme',
      tls: true,
      replace_dots: true,
    }
  }
  if (type === 'loki') {
    return {
      match: '*',
      host: 'loki.internal',
      port: 3100,
      tenant_id: 'platform',
      labels: 'job=fluent-manager',
      line_format: 'json',
    }
  }
  if (type === 'kafka') {
    return {
      match: '*',
      brokers: 'kafka-1:9092,kafka-2:9092',
      topics: 'logs.platform',
      format: 'json',
      timestamp_key: '@timestamp',
    }
  }
  if (type === 'http') {
    return {
      match: '*',
      host: 'collector.internal',
      port: 8080,
      uri: '/ingest',
      format: 'json_lines',
      header_authorization: '',
    }
  }
  if (type === 's3') {
    return {
      match: '*',
      bucket: 'company-log-archive',
      region: 'ap-northeast-1',
      path: 'fluent/%Y/%m/%d/',
      compression: 'gzip',
    }
  }
  if (type === 'stdout') {
    return {
      match: '*',
      format: 'json_lines',
    }
  }
  return {
    match: '*',
    plugin: 'custom_output',
    notes: '',
  }
}

function parseSettingsObject(raw, fallbackType) {
  const fallback = defaultSettingsObject(fallbackType)
  if (!raw || !String(raw).trim()) return { ...fallback }
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return { ...fallback }
    }
    return { ...fallback, ...parsed }
  } catch {
    return { ...fallback }
  }
}

function applySettingsObjectToForm(settings) {
  for (const key of Object.keys(settingsForm)) {
    delete settingsForm[key]
  }
  for (const [key, value] of Object.entries(settings || {})) {
    settingsForm[key] = value
  }
}

function syncSettingsJSONFromForm() {
  const next = {}
  for (const field of currentFields.value) {
    let value = settingsForm[field.key]
    if (field.type === 'number') {
      value = value === '' || value === null || value === undefined ? null : Number(value)
      if (value === null || Number.isNaN(value)) continue
    }
    if (field.type === 'text' || field.type === 'password' || field.type === 'textarea') {
      value = String(value ?? '').trim()
      if (!value) continue
    }
    next[field.key] = value
  }
  form.settings = JSON.stringify(next, null, 2)
}

function syncSettingsFormFromJSON() {
  applySettingsObjectToForm(parseSettingsObject(form.settings, form.target_type))
}

function setSettingsMode(mode) {
  if (mode === settingsMode.value) return
  if (mode === 'json') {
    syncSettingsJSONFromForm()
    settingsMode.value = 'json'
    return
  }
  syncSettingsFormFromJSON()
  settingsMode.value = 'guided'
}

function resetForm() {
  editingId.value = null
  settingsMode.value = 'guided'
  form.name = ''
  form.description = ''
  form.fluent_type = 'shared'
  form.target_type = 'opensearch'
  form.endpoint = ''
  form.settings = JSON.stringify(defaultSettingsObject('opensearch'), null, 2)
  syncSettingsFormFromJSON()
}

async function loadTargets() {
  targets.value = await getOutputTargets()
}

function openCreate() {
  resetForm()
  ensureModal()
  modal.show()
}

function openEdit(target) {
  editingId.value = target.id
  settingsMode.value = 'guided'
  form.name = target.name
  form.description = target.description || ''
  form.fluent_type = target.fluent_type || 'shared'
  form.target_type = target.target_type || 'custom'
  form.endpoint = target.endpoint || ''
  form.settings = target.settings || '{}'
  syncSettingsFormFromJSON()
  ensureModal()
  modal.show()
}

async function saveTarget() {
  try {
    if (settingsMode.value === 'guided') {
      syncSettingsJSONFromForm()
    } else {
      const parsed = JSON.parse(form.settings || '{}')
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error(t('output_targets.invalid_settings'))
      }
      form.settings = JSON.stringify(parsed, null, 2)
    }

    const payload = {
      name: form.name,
      description: form.description,
      fluent_type: form.fluent_type,
      target_type: form.target_type,
      endpoint: form.endpoint,
      settings: form.settings,
    }
    if (editingId.value) {
      await updateOutputTarget(editingId.value, payload)
    } else {
      await createOutputTarget(payload)
    }
    modal.hide()
    await loadTargets()
  } catch (error) {
    alert(`${editingId.value ? t('output_targets.save_failed') : t('output_targets.create_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDelete(target) {
  if (!confirm(t('output_targets.confirm_delete').replace('{name}', target.name))) return
  try {
    await deleteOutputTarget(target.id)
    await loadTargets()
  } catch (error) {
    alert(`${t('output_targets.delete_failed')}: ${getErrorMessage(error)}`)
  }
}

watch(
  () => form.target_type,
  (next, previous) => {
    if (!next) return
    if (previous && previous !== next) {
      form.settings = JSON.stringify(defaultSettingsObject(next), null, 2)
      syncSettingsFormFromJSON()
    } else if (!Object.keys(settingsForm).length) {
      syncSettingsFormFromJSON()
    }
  },
  { immediate: true }
)

watch(
  settingsForm,
  () => {
    if (settingsMode.value === 'guided') {
      syncSettingsJSONFromForm()
    }
  },
  { deep: true }
)

onMounted(loadTargets)
</script>

<style scoped>
.fm-settings-preview {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #0f172a;
  color: #dbeafe;
  min-height: 280px;
  font-size: 0.82rem;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
