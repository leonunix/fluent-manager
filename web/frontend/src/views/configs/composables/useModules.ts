// Module management composable – state, computed, and CRUD for config modules.
// Deps injected: { modules, t, onModuleDeleted?, onBatchModulesDeleted? }
// modules ref is owned by ConfigsPage and passed in.

import { computed, reactive, ref } from 'vue'
import {
  getModules, createModule, updateModule, deleteModule, deleteModules,
  createModuleVersion, getModuleVersions,
} from '../../../api'
import { getErrorMessage as getErrorMessageFn } from '../utils/config-format'
import {
  normalizeName, generateUniqueDraftName, countNonEmptyLines,
  extractTemplatePlaceholders, uniqueSorted, diffKeys,
} from '../utils/config-text'
import {
  parseVariablesMap, parseVariablesMapStrict, buildModuleVariableRows,
} from '../utils/config-variables'
import {
  createAIDraftState, resetAIDraftState, activateAIDraftState, areDraftConfirmationsComplete,
} from '../utils/config-ai-draft'
import { runtimeLabel } from '../utils/config-format'

export function useModules({ modules, t, onModuleDeleted, onBatchModulesDeleted }: any) {
  function getErrorMessage(e: any) { return getErrorMessageFn(e, t) }

  const editingModuleId = ref<number | null>(null)
  const currentModule = ref<any>(null)
  const moduleVersions = ref<any[]>([])
  const selectedModuleIds = ref<number[]>([])
  const moduleTableItems = ref<any[]>([])
  const moduleTableTotal = ref(0)
  const moduleTableLoading = ref(false)
  const moduleVariablesMode = ref('form')
  const moduleVariableRows = ref<any[]>([])
  const moduleVariablesFormError = ref('')

  const moduleQuery = reactive({
    search: '',
    fluent_type: '',
    module_type: '',
    page: 1,
    page_size: 20,
  })

  const moduleForm = reactive({
    name: '',
    description: '',
    module_type: 'input',
    fluent_type: 'fluentbit',
    content: '',
    content_fluentd: '',
    variables: '',
    is_builtin: false,
  })

  const moduleVersionForm = reactive({
    comment: '',
    variables: '{}',
    content: '',
  })

  const aiModuleDraftState = reactive(createAIDraftState())

  const moduleTypes = ['service', 'input', 'parser', 'filter', 'route', 'output']
  const managedModuleTypes = ['service', 'input', 'parser', 'filter', 'route', 'output']

  const moduleExamples: Record<string, any> = {
    fluentbit: {
      service: {
        variables: '{\n  "flush": 1,\n  "log_level": "info"\n}',
        content: '[SERVICE]\n    Flush        {{ .flush }}\n    Daemon       Off\n    Log_Level    {{ .log_level }}',
      },
      input: {
        variables: '{\n  "path": "/var/log/nginx/*.log",\n  "tag": "nginx.access",\n  "db_path": "/var/lib/fluent-bit/nginx.db"\n}',
        content: '[INPUT]\n    Name              tail\n    Path              {{ .path }}\n    Tag               {{ .tag }}\n    DB                {{ .db_path }}\n    Refresh_Interval  5',
      },
      parser: {
        variables: '{\n  "time_key": "time",\n  "time_format": "%d/%b/%Y:%H:%M:%S %z"\n}',
        content: '[PARSER]\n    Name         nginx_json\n    Format       json\n    Time_Key     {{ .time_key }}\n    Time_Format  {{ .time_format }}',
      },
      filter: {
        variables: '{\n  "match": "nginx.*",\n  "env": "prod"\n}',
        content: '[FILTER]\n    Name    modify\n    Match   {{ .match }}\n    Add     environment {{ .env }}',
      },
      route: {
        variables: '{\n  "match": "nginx.*"\n}',
        content: '# Fluent Bit usually routes in outputs by Match\n# Keep route intent documented here for reuse\n# Match: {{ .match }}',
      },
      output: {
        variables: '{\n  "match": "nginx.*",\n  "host": "10.0.0.15",\n  "port": 24224\n}',
        content: '[OUTPUT]\n    Name   forward\n    Match  {{ .match }}\n    Host   {{ .host }}\n    Port   {{ .port }}',
      },
    },
    fluentd: {
      service: {
        variables: '{\n  "workers": 2,\n  "log_level": "info"\n}',
        content: '<system>\n  workers {{ .workers }}\n  log_level {{ .log_level }}\n</system>',
      },
      input: {
        variables: '{\n  "path": "/var/log/nginx/access.log",\n  "tag": "nginx.access",\n  "pos_file": "/var/log/fluentd/nginx.pos"\n}',
        content: '<source>\n  @type tail\n  path {{ .path }}\n  pos_file {{ .pos_file }}\n  tag {{ .tag }}\n  <parse>\n    @type json\n  </parse>\n</source>',
      },
      parser: {
        variables: '{\n  "format": "/^(?<time>[^ ]*) (?<level>[^ ]*) (?<message>.*)$/",\n  "time_format": "%Y-%m-%dT%H:%M:%S"\n}',
        content: '<parse>\n  @type regexp\n  expression {{ .format }}\n  time_format {{ .time_format }}\n</parse>',
      },
      filter: {
        variables: '{\n  "match": "nginx.**",\n  "record_key": "environment",\n  "record_value": "prod"\n}',
        content: '<filter {{ .match }}>\n  @type record_transformer\n  <record>\n    {{ .record_key }} {{ .record_value }}\n  </record>\n</filter>',
      },
      route: {
        variables: '{\n  "match": "nginx.**",\n  "label": "@ARCHIVE"\n}',
        content: '<match {{ .match }}>\n  @type relabel\n  @label {{ .label }}\n</match>',
      },
      output: {
        variables: '{\n  "match": "nginx.**",\n  "host": "10.0.0.15",\n  "port": 24224\n}',
        content: '<match {{ .match }}>\n  @type forward\n  <server>\n    host {{ .host }}\n    port {{ .port }}\n  </server>\n</match>',
      },
    },
    shared: {
      service: { variables: '{\n  "flush_seconds": 5\n}', content: '# Shared service tuning\n# Flush interval: {{ .flush_seconds }}s' },
      input: { variables: '{\n  "tag": "app.logs"\n}', content: '# Shared input hints\n# Default tag {{ .tag }}' },
      parser: { variables: '{\n  "time_key": "time"\n}', content: '# Shared parser hint\n# Use time key {{ .time_key }}' },
      filter: { variables: '{\n  "match": "app.*",\n  "env": "prod"\n}', content: '# Shared filter intent\n# Match {{ .match }} and stamp env {{ .env }}' },
      route: { variables: '{\n  "match": "app.*",\n  "destination": "central-forward"\n}', content: '# Shared routing hint\n# Route {{ .match }} to {{ .destination }}' },
      output: { variables: '{\n  "host": "10.0.0.15",\n  "port": 24224\n}', content: '# Shared output hint\n# Forward to {{ .host }}:{{ .port }}' },
    },
  }

  // --- Computed ---

  const moduleCatalogCount = computed(() => modules.value.length)
  const visibleModules = computed(() => moduleTableItems.value)
  const visibleDeletableModules = computed(() => visibleModules.value.filter((item: any) => !item.is_builtin))
  const allVisibleDeletableModulesSelected = computed(() =>
    visibleDeletableModules.value.length > 0 &&
    visibleDeletableModules.value.every((item: any) => selectedModuleIds.value.includes(item.id))
  )
  const selectedDeletableModuleCount = computed(() => selectedModuleIds.value.length)
  const sharedModuleCount = computed(() => modules.value.filter((item: any) => item.fluent_type === 'shared').length)
  const usedModuleTypes = computed(() => managedModuleTypes.filter((type) => modules.value.some((item: any) => item.module_type === type)))
  const moduleTypeStats = computed(() =>
    managedModuleTypes
      .map((type) => ({ type, count: modules.value.filter((item: any) => item.module_type === type).length }))
      .filter((item) => item.count > 0)
  )
  const moduleTableTotalPages = computed(() => Math.max(1, Math.ceil(moduleTableTotal.value / Math.max(Number(moduleQuery.page_size) || 20, 1))))
  const moduleTableRangeStart = computed(() => (moduleTableTotal.value ? (moduleQuery.page - 1) * moduleQuery.page_size + 1 : 0))
  const moduleTableRangeEnd = computed(() => (moduleTableTotal.value ? Math.min(moduleQuery.page * moduleQuery.page_size, moduleTableTotal.value) : 0))
  const currentModuleExample = computed(() => {
    const runtimeExamples = moduleExamples[moduleForm.fluent_type] || moduleExamples.shared
    return runtimeExamples[moduleForm.module_type] || runtimeExamples.input || { variables: '{}', content: '# Example content' }
  })
  const aiModuleDraftSource = computed(() => [aiModuleDraftState.provider, aiModuleDraftState.accountName].filter(Boolean).join(' / '))
  const aiModuleDraftReady = computed(() => areDraftConfirmationsComplete(aiModuleDraftState))
  const aiModuleDraftComparison = computed(() => buildModuleDraftComparison())
  const aiModuleDraftCanSave = computed(() => {
    if (!aiModuleDraftReady.value) return false
    if (!aiModuleDraftState.active) return true
    return !!editingModuleId.value || !aiModuleDraftComparison.value?.hasConflict
  })

  // --- Functions ---

  function summarizeChangedKeys(added: string[], removed: string[]): string {
    const parts: string[] = []
    if (added.length) parts.push(t('configs_page.ai_draft_diff_added').replace('{items}', added.join(', ')))
    if (removed.length) parts.push(t('configs_page.ai_draft_diff_removed').replace('{items}', removed.join(', ')))
    return parts.join('；')
  }

  function buildModuleDraftComparison() {
    if (!aiModuleDraftState.active) return null
    const variableKeys = uniqueSorted(Object.keys(parseVariablesMap(moduleForm.variables)))
    const placeholderKeys = uniqueSorted(extractTemplatePlaceholders(moduleForm.content).map((item: string) => item.replace(/[{}\s.]/g, '')))
    const existing = modules.value.find((item: any) =>
      normalizeName(item.name) === normalizeName(moduleForm.name) &&
      item.module_type === moduleForm.module_type &&
      item.fluent_type === moduleForm.fluent_type
    )
    let identityMessage = t('configs_page.ai_draft_diff_new_asset')
    let existingName = '', changeMessage = '', changeDetail = '', suggestedName = ''
    const hasConflict = !!existing
    if (existing) {
      existingName = `${existing.name} / ${existing.module_type} / ${runtimeLabel(existing.fluent_type)}`
      identityMessage = t('configs_page.ai_draft_diff_existing_asset')
      suggestedName = generateUniqueDraftName(
        moduleForm.name,
        modules.value.filter((item: any) => item.module_type === moduleForm.module_type && item.fluent_type === moduleForm.fluent_type).map((item: any) => item.name),
        `ai-${moduleForm.module_type || 'module'}`
      )
      const prevKeys = uniqueSorted(Object.keys(parseVariablesMap(existing.variables)))
      const { added, removed } = diffKeys(variableKeys, prevKeys)
      changeMessage = existing.content === moduleForm.content
        ? t('configs_page.ai_draft_diff_content_same')
        : t('configs_page.ai_draft_diff_content_changed')
      changeDetail = summarizeChangedKeys(added, removed)
      if (!changeDetail) changeDetail = t('configs_page.ai_draft_diff_existing_review')
    }
    return {
      existingAsset: existing || null, hasConflict, identityMessage, existingName, suggestedName,
      variableCount: variableKeys.length, lineCount: countNonEmptyLines(moduleForm.content),
      placeholderCount: placeholderKeys.length, changeMessage, changeDetail,
    }
  }

  async function listAllModules(): Promise<any[]> {
    const pageSize = 100
    const collected: any[] = []
    let page = 1, total = 0
    do {
      const { data } = await getModules({ page, page_size: pageSize })
      const batch = data.data || []
      total = Number(data.total || 0)
      collected.push(...batch)
      if (!batch.length) break
      page += 1
    } while (collected.length < total)
    return collected
  }

  async function loadModules() {
    modules.value = await listAllModules()
    const deletableIDs = new Set(modules.value.filter((item: any) => !item.is_builtin).map((item: any) => item.id))
    selectedModuleIds.value = selectedModuleIds.value.filter((id) => deletableIDs.has(id))
  }

  async function loadModuleTable() {
    moduleTableLoading.value = true
    try {
      const params: Record<string, any> = { page: moduleQuery.page, page_size: moduleQuery.page_size }
      if (moduleQuery.search.trim()) params.search = moduleQuery.search.trim()
      if (moduleQuery.fluent_type) params.fluent_type = moduleQuery.fluent_type
      if (moduleQuery.module_type) params.module_type = moduleQuery.module_type
      const { data } = await getModules(params)
      moduleTableItems.value = data.data || []
      moduleTableTotal.value = Number(data.total || 0)
      const totalPages = Math.max(1, Math.ceil(moduleTableTotal.value / Math.max(Number(moduleQuery.page_size) || 20, 1)))
      if (moduleTableTotal.value > 0 && moduleQuery.page > totalPages) {
        moduleQuery.page = totalPages
        await loadModuleTable()
      }
    } catch (error) {
      moduleTableItems.value = []
      moduleTableTotal.value = 0
      alert(`${t('common.request_failed')}: ${getErrorMessage(error)}`)
    } finally {
      moduleTableLoading.value = false
    }
  }

  function applyModuleQuery() { moduleQuery.page = 1; loadModuleTable() }
  function resetModuleQuery() {
    moduleQuery.search = ''; moduleQuery.fluent_type = ''; moduleQuery.module_type = ''
    moduleQuery.page = 1; moduleQuery.page_size = 20; loadModuleTable()
  }
  function changeModulePage(nextPage: number) {
    if (nextPage < 1 || nextPage > moduleTableTotalPages.value || nextPage === moduleQuery.page) return
    moduleQuery.page = nextPage; loadModuleTable()
  }
  function setModuleTypeFilter(type = '') {
    moduleQuery.module_type = moduleQuery.module_type === type ? '' : type
    moduleQuery.page = 1; loadModuleTable()
  }
  function toggleModuleSelection(module: any) {
    if (!module || module.is_builtin) return
    if (selectedModuleIds.value.includes(module.id)) {
      selectedModuleIds.value = selectedModuleIds.value.filter((id) => id !== module.id)
    } else {
      selectedModuleIds.value = [...selectedModuleIds.value, module.id]
    }
  }
  function toggleSelectAllVisibleModules() {
    const visibleIDs = visibleDeletableModules.value.map((item: any) => item.id)
    if (!visibleIDs.length) return
    if (allVisibleDeletableModulesSelected.value) {
      selectedModuleIds.value = selectedModuleIds.value.filter((id) => !visibleIDs.includes(id))
    } else {
      selectedModuleIds.value = Array.from(new Set([...selectedModuleIds.value, ...visibleIDs]))
    }
  }

  function syncModuleVariablesFromRows(): boolean {
    try {
      const payload: Record<string, any> = {}
      for (const row of moduleVariableRows.value) {
        const key = String(row.key || '').trim()
        if (!key) continue
        payload[key] = row.type === 'boolean' ? row.value === 'true'
          : row.type === 'number' ? Number(row.value)
          : row.type === 'json' ? JSON.parse(row.value || 'null')
          : row.value
      }
      moduleForm.variables = JSON.stringify(payload, null, 2)
      moduleVariablesFormError.value = ''
      return true
    } catch {
      moduleVariablesFormError.value = t('configs_page.variable_form_invalid')
      return false
    }
  }
  function addModuleVariableRow() {
    moduleVariableRows.value = [...moduleVariableRows.value, { key: '', type: 'string', value: '' }]
  }
  function removeModuleVariableRow(index: number) {
    moduleVariableRows.value = moduleVariableRows.value.filter((_, i) => i !== index)
    if (!moduleVariableRows.value.length) moduleVariableRows.value = [{ key: '', type: 'string', value: '' }]
    syncModuleVariablesFromRows()
  }
  function setModuleVariablesMode(mode: string) {
    if (mode === moduleVariablesMode.value) return
    if (mode === 'json') {
      if (!syncModuleVariablesFromRows()) return
      moduleVariablesMode.value = 'json'
      return
    }
    try {
      moduleVariableRows.value = buildModuleVariableRows(moduleForm.variables)
      moduleVariablesFormError.value = ''
      moduleVariablesMode.value = 'form'
    } catch (error) {
      alert(`${t('configs_page.variable_mode_switch_failed')}: ${getErrorMessage(error)}`)
    }
  }

  function applyAIModuleVariables(raw: string) {
    moduleForm.variables = raw || '{}'
    try {
      moduleVariableRows.value = buildModuleVariableRows(moduleForm.variables)
      moduleVariablesMode.value = 'form'
      moduleVariablesFormError.value = ''
    } catch {
      moduleVariablesMode.value = 'json'
      moduleVariableRows.value = [{ key: '', type: 'string', value: '' }]
      moduleVariablesFormError.value = ''
    }
  }

  function resetModuleForm() {
    editingModuleId.value = null
    moduleForm.name = ''; moduleForm.description = ''; moduleForm.module_type = 'input'
    moduleForm.fluent_type = 'fluentbit'; moduleForm.content = ''; moduleForm.content_fluentd = ''
    moduleForm.variables = ''; moduleForm.is_builtin = false
    moduleVariablesMode.value = 'form'
    moduleVariableRows.value = [{ key: '', type: 'string', value: '' }]
    moduleVariablesFormError.value = ''
    resetAIDraftState(aiModuleDraftState)
  }

  function resetModuleVersionForm(module: any) {
    moduleVersionForm.comment = ''
    moduleVersionForm.variables = module?.variables || '{}'
    moduleVersionForm.content = module?.content || ''
  }

  function prepareModuleEdit(module: any) {
    resetAIDraftState(aiModuleDraftState)
    editingModuleId.value = module.id
    moduleForm.name = module.name; moduleForm.description = module.description || ''
    moduleForm.module_type = module.module_type; moduleForm.fluent_type = module.fluent_type
    moduleForm.content = module.content; moduleForm.content_fluentd = module.content_fluentd || ''
    moduleForm.variables = module.variables || '{}'; moduleForm.is_builtin = !!module.is_builtin
    moduleVariablesMode.value = 'form'
    moduleVariableRows.value = buildModuleVariableRows(moduleForm.variables)
    moduleVariablesFormError.value = ''
  }

  function applySuggestedModuleName() {
    const name = aiModuleDraftComparison.value?.suggestedName
    if (name) moduleForm.name = name
  }

  function openExistingModuleFromDraft(openEditModule: (m: any) => void) {
    const existing = aiModuleDraftComparison.value?.existingAsset
    if (!existing) return
    openEditModule(existing)
  }

  async function saveModule(): Promise<boolean> {
    try {
      if (moduleVariablesMode.value === 'form') {
        if (!syncModuleVariablesFromRows()) { alert(t('configs_page.variable_form_invalid')); return false }
      } else {
        parseVariablesMapStrict(moduleForm.variables)
      }
      if (editingModuleId.value) {
        await updateModule(editingModuleId.value, moduleForm)
      } else {
        await createModule(moduleForm)
      }
      resetAIDraftState(aiModuleDraftState)
      await Promise.all([loadModules(), loadModuleTable()])
      return true
    } catch (error) {
      alert(`${editingModuleId.value ? t('configs_page.save_module_failed') : t('configs_page.create_module_failed')}: ${getErrorMessage(error)}`)
      return false
    }
  }

  async function handleDeleteModule(module: any) {
    if (module.is_builtin) { alert(t('configs_page.builtin_module_protected')); return }
    if (!confirm(t('configs_page.confirm_delete_module').replace('{name}', module.name))) return
    try {
      await deleteModule(module.id)
      selectedModuleIds.value = selectedModuleIds.value.filter((id) => id !== module.id)
      onModuleDeleted?.(module.id)
      await Promise.all([loadModules(), loadModuleTable()])
    } catch (error) {
      alert(`${t('configs_page.delete_module_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function handleBatchDeleteModules() {
    if (!selectedModuleIds.value.length) return
    if (!confirm(t('configs_page.confirm_batch_delete_modules').replace('{count}', String(selectedModuleIds.value.length)))) return
    try {
      await deleteModules(selectedModuleIds.value)
      const deletedSet = new Set(selectedModuleIds.value)
      selectedModuleIds.value = []
      onBatchModulesDeleted?.(deletedSet)
      await Promise.all([loadModules(), loadModuleTable()])
    } catch (error) {
      alert(`${t('configs_page.batch_delete_modules_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function openModuleVersions(module: any) {
    currentModule.value = module
    resetModuleVersionForm(module)
    try {
      const { data } = await getModuleVersions(module.id)
      moduleVersions.value = data.data || []
      return true
    } catch (error) {
      alert(`${t('configs_page.load_module_versions_failed')}: ${getErrorMessage(error)}`)
      return false
    }
  }

  async function saveModuleVersion(): Promise<boolean> {
    if (!currentModule.value) return false
    try {
      await createModuleVersion(currentModule.value.id, moduleVersionForm)
      await openModuleVersions(currentModule.value)
      await Promise.all([loadModules(), loadModuleTable()])
      return true
    } catch (error) {
      alert(`${t('configs_page.create_module_version_failed')}: ${getErrorMessage(error)}`)
      return false
    }
  }

  function useAIModuleDraft(module: any, aiAssistantResult: any, aiAssistantForm: any) {
    if (!module) return
    resetModuleForm()
    editingModuleId.value = null
    moduleForm.name = module.name || `ai-${module.module_type}`
    moduleForm.description = aiAssistantResult?.summary || ''
    moduleForm.module_type = module.module_type || aiAssistantForm.module_type
    moduleForm.fluent_type = aiAssistantForm.fluent_type
    moduleForm.content = module.content || ''
    moduleForm.is_builtin = false
    applyAIModuleVariables(module.variables_json || '{}')
    const draftResult = {
      ...aiAssistantResult,
      notes: module.note ? [module.note, ...(aiAssistantResult?.notes || [])] : (aiAssistantResult?.notes || []),
      assembly_steps: [],
    }
    activateAIDraftState(aiModuleDraftState, draftResult, [
      t('configs_page.ai_draft_review_name'), t('configs_page.ai_draft_review_runtime'),
      t('configs_page.ai_draft_review_variables'), t('configs_page.ai_draft_review_module_content'),
    ], [
      t('configs_page.ai_draft_confirm_name'), t('configs_page.ai_draft_confirm_variables'),
      t('configs_page.ai_draft_confirm_target'), t('configs_page.ai_draft_confirm_module_content'),
    ])
  }

  return {
    editingModuleId, currentModule, moduleVersions, selectedModuleIds,
    moduleTableItems, moduleTableTotal, moduleTableLoading,
    moduleVariablesMode, moduleVariableRows, moduleVariablesFormError,
    moduleQuery, moduleForm, moduleVersionForm, aiModuleDraftState,
    moduleTypes, managedModuleTypes, moduleExamples,
    moduleCatalogCount, visibleModules, visibleDeletableModules,
    allVisibleDeletableModulesSelected, selectedDeletableModuleCount,
    sharedModuleCount, usedModuleTypes, moduleTypeStats,
    moduleTableTotalPages, moduleTableRangeStart, moduleTableRangeEnd,
    currentModuleExample, aiModuleDraftSource, aiModuleDraftReady,
    aiModuleDraftComparison, aiModuleDraftCanSave,
    listAllModules, loadModules, loadModuleTable,
    applyModuleQuery, resetModuleQuery, changeModulePage, setModuleTypeFilter,
    toggleModuleSelection, toggleSelectAllVisibleModules,
    syncModuleVariablesFromRows, addModuleVariableRow, removeModuleVariableRow, setModuleVariablesMode,
    applyAIModuleVariables, resetModuleForm, resetModuleVersionForm,
    prepareModuleEdit, applySuggestedModuleName, openExistingModuleFromDraft,
    saveModule, handleDeleteModule, handleBatchDeleteModules,
    openModuleVersions, saveModuleVersion, useAIModuleDraft, buildModuleDraftComparison,
  }
}
