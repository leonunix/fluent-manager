// Wizard composable – all wizard-related state, computed properties, and functions.
// External deps injected: { modules, outputTargets, pipelines, t, previewForm,
//   renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult,
//   activeTab, loadTemplates, router }

import { computed, reactive, ref, watch } from 'vue'
import { buildConfigFlowSummary } from '../../../utils/config_flow'
import { previewRenderedConfig, getRenderedConfig, createTemplate, updateTemplate, createVersion } from '../../../api'
import { matchesModuleSearch, matchesOutputTargetSearch } from '../utils/config-search'
import { paginateItems } from '../utils/config-text'
import {
  parseVariablesMap,
  normalizeWizardDraftValues,
  stringifyVariableValue,
  inferVariableKind,
  buildWizardVariableDraft,
} from '../utils/config-variables'
import {
  createWizardInstanceID,
  createWizardPipeline,
  moduleVariablesForWizard,
  parserNamesProvidedByModule,
  parserReferencesForInputModule,
  shouldAutoSyncWizardMatch,
  buildWizardModuleGroup,
  matchingOutputModuleForTarget,
  buildWizardModuleRef,
  ensureWizardModuleDraft,
  wizardPipelineDisplayName as wizardPipelineDisplayNameFn,
} from '../utils/config-wizard-helpers'
import { normalizeSearchText } from '../utils/config-search'
import { wizardGoalLabel as wizardGoalLabelFn, getErrorMessage as getErrorMessageFn } from '../utils/config-format'

export function useWizard({
  modules,
  outputTargets,
  pipelines,
  t,
  previewForm,
  renderedConfig,
  analysisResult,
  compatibilityResult,
  replayResult,
  diffResult,
  activeTab,
  loadTemplates,
  router,
}) {
  function wizardGoalLabel(goal) { return wizardGoalLabelFn(goal, t) }
  function wizardPipelineDisplayName(pipelineOrCard, index = 0) { return wizardPipelineDisplayNameFn(pipelineOrCard, index, t) }
  function getErrorMessage(error) { return getErrorMessageFn(error, t) }

  // --- State ---

  const wizardForm = reactive({
    goal: 'edge_collection',
    name: '',
    description: '',
    fluent_type: 'fluentbit',
    runtime_version: '',
  })
  const wizardServiceModuleId = ref(null)
  const wizardParserModuleIds = ref([])
  const wizardGlobalModuleVariables = ref({})
  const wizardPipelines = ref([])
  const activeWizardPipelineId = ref('')
  const wizardLoadedFromTemplate = ref(null)
  const wizardServiceSearch = ref('')
  const wizardParserSearch = ref('')
  const wizardInputSearch = ref('')
  const wizardFilterSearch = ref('')
  const wizardRouteSearch = ref('')
  const wizardOutputSearch = ref('')
  const wizardModuleSearch = ref('')
  const wizardStagePages = reactive({ service: 1, parser: 1, input: 1, filter: 1, route: 1, output: 1 })
  const selectedWizardModuleIds = ref([])
  const selectedWizardOutputTargetIds = ref([])
  const wizardModuleVariableValues = ref({})

  const wizardPipelineModuleTypes = ['input', 'filter', 'route']
  const wizardPipelineStageTotal = 3

  // --- Computed ---

  const wizardSaveButtonLabel = computed(() => (
    wizardLoadedFromTemplate.value?.id
      ? t('configs_page.wizard_update_template')
      : t('configs_page.wizard_save_as_template')
  ))

  const wizardEligibleModules = computed(() =>
    modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === wizardForm.fluent_type)
  )
  const wizardServiceModules = computed(() =>
    wizardEligibleModules.value.filter((item) => item.module_type === 'service' && matchesModuleSearch(item, wizardServiceSearch.value))
  )
  const wizardParserModules = computed(() =>
    wizardEligibleModules.value.filter((item) => item.module_type === 'parser' && matchesModuleSearch(item, wizardParserSearch.value))
  )
  const wizardInputModules = computed(() =>
    wizardEligibleModules.value.filter((item) => item.module_type === 'input' && matchesModuleSearch(item, wizardInputSearch.value))
  )
  const wizardFilterModules = computed(() =>
    wizardEligibleModules.value.filter((item) => item.module_type === 'filter' && matchesModuleSearch(item, wizardFilterSearch.value))
  )
  const wizardRouteModules = computed(() =>
    wizardEligibleModules.value.filter((item) => item.module_type === 'route' && matchesModuleSearch(item, wizardRouteSearch.value))
  )
  const wizardVisibleModules = computed(() =>
    wizardEligibleModules.value.filter((item) =>
      wizardPipelineModuleTypes.includes(item.module_type) && matchesModuleSearch(item, wizardModuleSearch.value)
    )
  )
  const wizardInputPresets = computed(() =>
    wizardVisibleModules.value.filter((item) => item.module_type === 'input' && item.preset_kind === 'input')
  )
  const wizardAvailableOutputTargets = computed(() =>
    outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === wizardForm.fluent_type)
  )
  const wizardOutputTargets = computed(() =>
    wizardAvailableOutputTargets.value.filter((item) => matchesOutputTargetSearch(item, wizardOutputSearch.value))
  )
  const wizardCompatiblePipelines = computed(() =>
    pipelines.value.filter((p) => p.fluent_type === wizardForm.fluent_type)
  )
  const wizardPagedServiceModules = computed(() => paginateItems(wizardServiceModules.value, wizardStagePages.service))
  const wizardPagedParserModules = computed(() => paginateItems(wizardParserModules.value, wizardStagePages.parser))
  const wizardPagedInputModules = computed(() => paginateItems(wizardInputModules.value, wizardStagePages.input))
  const wizardPagedFilterModules = computed(() => paginateItems(wizardFilterModules.value, wizardStagePages.filter))
  const wizardPagedRouteModules = computed(() => paginateItems(wizardRouteModules.value, wizardStagePages.route))
  const wizardPagedOutputTargets = computed(() => paginateItems(wizardOutputTargets.value, wizardStagePages.output))
  const wizardServiceModule = computed(() =>
    wizardEligibleModules.value.find((item) => item.id === wizardServiceModuleId.value) || null
  )
  const wizardSelectedParserModules = computed(() =>
    wizardEligibleModules.value.filter((item) => wizardParserModuleIds.value.includes(item.id))
  )
  const activeWizardPipeline = computed(() =>
    wizardPipelines.value.find((item) => item.id === activeWizardPipelineId.value) || wizardPipelines.value[0] || null
  )
  const wizardPipelineCards = computed(() =>
    wizardPipelines.value.map((pipeline, index) => {
      const inputModule = wizardEligibleModules.value.find((item) => item.id === pipeline.input?.module_id) || null
      const filterModules = pipeline.filters
        .map((instance) => wizardEligibleModules.value.find((item) => item.id === instance.module_id) || null)
        .filter(Boolean)
      const routeModules = (pipeline.routes || [])
        .map((instance) => wizardEligibleModules.value.find((item) => item.id === instance.module_id) || null)
        .filter(Boolean)
      const outputTargetsForPipeline = pipeline.outputs
        .map((instance) => wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id) || null)
        .filter(Boolean)
      const outputModulesForPipeline = outputTargetsForPipeline
        .map((target) => matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type))
        .filter(Boolean)
      const summary = buildConfigFlowSummary(
        [inputModule, ...filterModules, ...routeModules, ...outputModulesForPipeline].filter(Boolean),
        outputTargetsForPipeline
      )
      const missing = []
      if (!inputModule) missing.push(t('configs_page.pipeline_stage_input'))
      if (!outputTargetsForPipeline.length) missing.push(t('configs_page.pipeline_stage_output'))
      return {
        id: pipeline.id, index, name: pipeline.name,
        inputModule, filterModules, routeModules,
        outputTargets: outputTargetsForPipeline, outputModules: outputModulesForPipeline,
        summary, missing, complete: missing.length === 0,
      }
    })
  )
  const wizardRenderablePipelineCards = computed(() => wizardPipelineCards.value.filter((item) => item.complete))
  const wizardIncompletePipelineLabels = computed(() =>
    wizardPipelineCards.value
      .filter((item) => !item.complete)
      .map((item, index) => wizardPipelineDisplayName(item, index))
  )
  const wizardRenderSummary = computed(() => ({
    pipelineCount: wizardRenderablePipelineCards.value.length,
    outputCount: wizardRenderablePipelineCards.value.reduce((total, item) => total + item.outputTargets.length, 0),
  }))
  const wizardGlobalVariableGroups = computed(() => {
    const groups = []
    const globalSection = { key: 'global', title: t('configs_page.wizard_global_resources'), kind: 'global' }
    if (wizardServiceModule.value) {
      groups.push(buildWizardModuleGroup(
        `wizard-service-${wizardServiceModule.value.id}`,
        wizardServiceModule.value.name,
        t('configs_page.wizard_service_baseline'),
        wizardServiceModule.value,
        wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`],
        {},
        globalSection
      ))
    }
    for (const module of wizardSelectedParserModules.value) {
      groups.push(buildWizardModuleGroup(
        `wizard-parser-${module.id}`,
        module.name,
        t('configs_page.wizard_parser_assets'),
        module,
        wizardGlobalModuleVariables.value[`parser:${module.id}`],
        {},
        globalSection
      ))
    }
    return groups.filter(Boolean)
  })
  const wizardPipelineVariableGroups = computed(() => {
    const groups = []
    for (const [pipelineIndex, pipeline] of wizardPipelines.value.entries()) {
      const pipelineLabel = wizardPipelineDisplayName(pipeline, pipelineIndex)
      const pipelineSection = { key: `pipeline:${pipeline.id}`, title: pipelineLabel, kind: 'pipeline', ref: pipeline.id }
      if (pipeline.input) {
        const module = wizardEligibleModules.value.find((item) => item.id === pipeline.input.module_id)
        groups.push(buildWizardModuleGroup(
          pipeline.input.id,
          module?.name || t('configs_page.pipeline_stage_input'),
          `${pipelineLabel} · ${t('configs_page.pipeline_stage_input')}`,
          module, pipeline.input.variables, wizardPipelineModuleDefaults(module, pipeline), pipelineSection
        ))
      }
      pipeline.filters.forEach((instance, filterIndex) => {
        const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
        groups.push(buildWizardModuleGroup(
          instance.id,
          module?.name || t('configs_page.pipeline_stage_filter'),
          `${pipelineLabel} · ${t('configs_page.pipeline_stage_filter')} ${filterIndex + 1}`,
          module, instance.variables, wizardPipelineModuleDefaults(module, pipeline), pipelineSection
        ))
      });
      (pipeline.routes || []).forEach((instance, routeIndex) => {
        const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
        groups.push(buildWizardModuleGroup(
          instance.id,
          module?.name || t('configs_page.pipeline_stage_route', 'Route'),
          `${pipelineLabel} · ${t('configs_page.pipeline_stage_route', 'Route')} ${routeIndex + 1}`,
          module, instance.variables, wizardPipelineModuleDefaults(module, pipeline), pipelineSection
        ))
      })
      pipeline.outputs.forEach((instance, outputIndex) => {
        const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
        const module = matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)
        groups.push(buildWizardModuleGroup(
          instance.id,
          target?.name || t('configs_page.pipeline_stage_output'),
          `${pipelineLabel} · ${t('configs_page.pipeline_stage_output')} ${outputIndex + 1}`,
          module, instance.variables,
          wizardPipelineModuleDefaults(module, pipeline, {
            ...parseVariablesMap(target?.settings),
            output_target_name: target?.name || '',
            output_target_type: target?.target_type || '',
          }),
          pipelineSection
        ))
      })
    }
    return groups.filter(Boolean)
  })
  const wizardOutputResolutionWarnings = computed(() =>
    wizardPipelineCards.value
      .flatMap((card, index) => card.outputTargets
        .filter((target) => !matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type))
        .map((target) => ({ pipeline: wizardPipelineDisplayName(card, index), target: target.name })))
  )
  const wizardSelectedModules = computed(() =>
    wizardEligibleModules.value.filter((item) => selectedWizardModuleIds.value.includes(item.id))
  )
  const selectedWizardInputPresetKeys = computed(() =>
    wizardSelectedModules.value
      .filter((item) => item.module_type === 'input' && item.preset_kind === 'input' && item.preset_key)
      .map((item) => item.preset_key)
  )
  const wizardSelectedOutputTargets = computed(() =>
    wizardAvailableOutputTargets.value.filter((item) => selectedWizardOutputTargetIds.value.includes(item.id))
  )
  const wizardUnresolvedOutputTargets = computed(() =>
    wizardSelectedOutputTargets.value.filter((item) => !matchingOutputModuleForTarget(item, wizardEligibleModules.value, wizardForm.fluent_type))
  )
  const wizardInputPresetsSelected = computed(() =>
    wizardSelectedModules.value.filter((item) => item.module_type === 'input' && item.preset_kind === 'input')
  )
  const wizardSelectedInputModule = computed(() =>
    wizardSelectedModules.value.find((item) => item.module_type === 'input') || null
  )
  const wizardSelectedFilterModules = computed(() =>
    wizardSelectedModules.value.filter((item) => item.module_type === 'filter')
  )
  const wizardSelectedOutputModules = computed(() =>
    wizardSelectedOutputTargets.value
      .map((target) => matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type))
      .filter(Boolean)
  )
  const wizardSummaryModules = computed(() => [...wizardSelectedModules.value, ...wizardSelectedOutputModules.value])
  const wizardFlowSummary = computed(() => buildConfigFlowSummary(wizardSummaryModules.value, wizardSelectedOutputTargets.value))
  const wizardModulesByType = computed(() =>
    wizardPipelineModuleTypes
      .map((type) => ({
        type,
        modules: wizardVisibleModules.value.filter((item) => item.module_type === type),
        selectionType: type === 'input' ? 'radio' : 'checkbox',
        hint: type === 'input' ? t('configs_page.wizard_input_group_hint') : t('configs_page.wizard_filter_group_hint'),
      }))
      .filter((group) => group.modules.length)
  )
  const wizardVariableGroups = computed(() =>
    wizardSelectedModules.value
      .map((module) => {
        const variables = parseVariablesMap(module.variables)
        const fields = Object.entries(variables).map(([key, value]) => ({
          key, defaultValue: stringifyVariableValue(value), kind: inferVariableKind(value), description: module.description || '',
        }))
        return {
          moduleId: module.id, moduleName: module.name, moduleType: module.module_type,
          moduleTypeLabel: module.module_type === 'input' ? t('configs_page.pipeline_stage_input') : t('configs_page.pipeline_stage_filter'),
          fields,
        }
      })
      .filter((group) => group.fields.length > 0)
  )
  const wizardVariableFieldCount = computed(() =>
    wizardVariableGroups.value.reduce((total, group) => total + group.fields.length, 0)
  )
  const wizardPipelineCompletedStages = computed(() => {
    let total = 0
    if (wizardSelectedInputModule.value) total += 1
    if (wizardSelectedFilterModules.value.length) total += 1
    if (wizardSelectedOutputTargets.value.length) total += 1
    return total
  })
  const wizardMissingRequirements = computed(() => {
    const missing = []
    if (!wizardSelectedInputModule.value) missing.push(t('configs_page.pipeline_stage_input'))
    if (!wizardSelectedOutputTargets.value.length) missing.push(t('configs_page.pipeline_stage_output'))
    return missing
  })

  // --- Functions ---

  function matchingWizardParserModules(parserNames) {
    const wanted = new Set((parserNames || []).map((item) => normalizeSearchText(item)).filter(Boolean))
    if (!wanted.size) return []
    return wizardEligibleModules.value.filter((item) => {
      if (item.module_type !== 'parser') return false
      return parserNamesProvidedByModule(item).some((name) => wanted.has(normalizeSearchText(name)))
    })
  }

  function autoAttachWizardParsersForInputModule(module, draft = null) {
    for (const parserModule of matchingWizardParserModules(parserReferencesForInputModule(module, draft))) {
      if (!wizardParserModuleIds.value.includes(parserModule.id)) {
        toggleWizardParserModule(parserModule.id)
      }
    }
  }

  function wizardPipelineInputTag(pipeline) {
    const inputModuleID = pipeline?.input?.module_id
    if (!inputModuleID) return ''
    const module = wizardEligibleModules.value.find((item) => item.id === inputModuleID && item.module_type === 'input')
    const merged = { ...moduleVariablesForWizard(module), ...(pipeline?.input?.variables || {}) }
    return String(merged.tag || '').trim()
  }

  function wizardPipelineModuleDefaults(module, pipeline, extraDefaults = {}) {
    const defaults: Record<string, any> = { ...moduleVariablesForWizard(module), ...(extraDefaults || {}) }
    const inputTag = wizardPipelineInputTag(pipeline)
    if (inputTag && Object.prototype.hasOwnProperty.call(defaults, 'match')) {
      defaults.match = inputTag
    }
    return defaults
  }

  function buildWizardPipelineDraft(key, module, pipeline, existingDraft = null, extraDefaults = {}, previousTag = '') {
    const defaults = wizardPipelineModuleDefaults(module, pipeline, extraDefaults)
    if (!existingDraft || !Object.keys(existingDraft).length) {
      return buildWizardVariableDraft(defaults)
    }
    const nextDraft: Record<string, any> = { ...existingDraft }
    if (Object.prototype.hasOwnProperty.call(defaults, 'match') && shouldAutoSyncWizardMatch(nextDraft.match, previousTag)) {
      nextDraft.match = stringifyVariableValue(defaults.match)
    }
    return nextDraft
  }

  function mergeWizardModuleVariableValues(moduleId, nextValues) {
    if (!moduleId) return
    const existing = { ...(wizardModuleVariableValues.value[moduleId] || {}) }
    for (const [key, value] of Object.entries(nextValues || {})) {
      existing[key] = stringifyVariableValue(value)
    }
    wizardModuleVariableValues.value = { ...wizardModuleVariableValues.value, [moduleId]: existing }
  }

  function ensureWizardBaselineModules() {
    if (!wizardPipelines.value.length) {
      const pipeline = createWizardPipeline()
      wizardPipelines.value = [pipeline]
      activeWizardPipelineId.value = pipeline.id
      return
    }
    if (!wizardPipelines.value.some((item) => item.id === activeWizardPipelineId.value)) {
      activeWizardPipelineId.value = wizardPipelines.value[0].id
    }
  }

  function changeWizardStagePage(stage, nextPage) {
    const metaMap = {
      service: wizardPagedServiceModules.value, parser: wizardPagedParserModules.value,
      input: wizardPagedInputModules.value, filter: wizardPagedFilterModules.value, output: wizardPagedOutputTargets.value,
    }
    const meta = metaMap[stage]
    if (!meta) return
    wizardStagePages[stage] = Math.min(Math.max(nextPage, 1), meta.totalPages)
  }

  function resetWizardStagePage(stage) { wizardStagePages[stage] = 1 }

  function selectWizardServiceModule(moduleId) {
    if (wizardServiceModuleId.value === moduleId) { wizardServiceModuleId.value = null; return }
    const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'service')
    if (!module) return
    wizardServiceModuleId.value = moduleId
    wizardGlobalModuleVariables.value = {
      ...wizardGlobalModuleVariables.value,
      [`service:${moduleId}`]: ensureWizardModuleDraft(`service:${moduleId}`, module, wizardGlobalModuleVariables.value[`service:${moduleId}`]),
    }
  }

  function toggleWizardParserModule(moduleId) {
    const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'parser')
    if (!module) return
    if (wizardParserModuleIds.value.includes(moduleId)) {
      wizardParserModuleIds.value = wizardParserModuleIds.value.filter((id) => id !== moduleId)
      return
    }
    wizardParserModuleIds.value = [...wizardParserModuleIds.value, moduleId]
    wizardGlobalModuleVariables.value = {
      ...wizardGlobalModuleVariables.value,
      [`parser:${moduleId}`]: ensureWizardModuleDraft(`parser:${moduleId}`, module, wizardGlobalModuleVariables.value[`parser:${moduleId}`]),
    }
  }

  function addWizardPipeline() {
    const pipeline = createWizardPipeline()
    wizardPipelines.value = [...wizardPipelines.value, pipeline]
    activeWizardPipelineId.value = pipeline.id
  }

  function duplicateWizardPipeline(pipelineId) {
    const pipeline = wizardPipelines.value.find((item) => item.id === pipelineId)
    if (!pipeline) return
    const cloned = {
      id: createWizardInstanceID('wizard-pipeline'),
      name: pipeline.name ? `${pipeline.name}-copy` : '',
      input: pipeline.input ? { id: createWizardInstanceID('wizard-input'), module_id: pipeline.input.module_id, variables: { ...pipeline.input.variables } } : null,
      filters: pipeline.filters.map((instance) => ({ id: createWizardInstanceID('wizard-filter'), module_id: instance.module_id, variables: { ...instance.variables } })),
      routes: (pipeline.routes || []).map((instance) => ({ id: createWizardInstanceID('wizard-route'), module_id: instance.module_id, variables: { ...instance.variables } })),
      outputs: pipeline.outputs.map((instance) => ({ id: createWizardInstanceID('wizard-output'), target_id: instance.target_id, variables: { ...instance.variables } })),
    }
    wizardPipelines.value = [...wizardPipelines.value, cloned]
    activeWizardPipelineId.value = cloned.id
  }

  function removeWizardPipeline(pipelineId) {
    if (wizardPipelines.value.length <= 1) {
      wizardPipelines.value = [createWizardPipeline()]
      activeWizardPipelineId.value = wizardPipelines.value[0].id
      return
    }
    wizardPipelines.value = wizardPipelines.value.filter((item) => item.id !== pipelineId)
    ensureWizardBaselineModules()
  }

  function selectWizardPipeline(pipelineId) { activeWizardPipelineId.value = pipelineId }

  function updateWizardPipeline(pipelineId, updater) {
    wizardPipelines.value = wizardPipelines.value.map((pipeline) => pipeline.id !== pipelineId ? pipeline : updater(pipeline))
  }

  function setWizardPipelineInput(pipelineId, moduleId) {
    const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'input')
    if (!module) return
    const currentPipeline = wizardPipelines.value.find((item) => item.id === pipelineId) || null
    const previousTag = wizardPipelineInputTag(currentPipeline)
    const draft = ensureWizardModuleDraft(
      `input:${module.id}`, module,
      currentPipeline?.input?.module_id === module.id ? currentPipeline?.input?.variables : null
    )
    const nextInput = { id: createWizardInstanceID('wizard-input'), module_id: module.id, variables: draft }
    updateWizardPipeline(pipelineId, (pipeline) => {
      const nextPipeline = { ...pipeline, input: nextInput }
      return {
        ...nextPipeline,
        filters: pipeline.filters.map((instance) => {
          const filterModule = wizardEligibleModules.value.find((item) => item.id === instance.module_id && item.module_type === 'filter')
          return { ...instance, variables: buildWizardPipelineDraft(`filter:${instance.module_id}`, filterModule, nextPipeline, instance.variables, {}, previousTag) }
        }),
        outputs: pipeline.outputs.map((instance) => {
          const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
          return { ...instance, variables: buildWizardOutputDraft(target, wizardForm.fluent_type, nextPipeline, instance.variables, previousTag) }
        }),
      }
    })
    autoAttachWizardParsersForInputModule(module, draft)
  }

  function addWizardFilter(pipelineId, moduleId) {
    const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'filter')
    if (!module) return
    updateWizardPipeline(pipelineId, (pipeline) => ({
      ...pipeline,
      filters: [...pipeline.filters, { id: createWizardInstanceID('wizard-filter'), module_id: module.id, variables: buildWizardPipelineDraft(`filter:${module.id}`, module, pipeline) }],
    }))
  }

  function removeWizardFilter(pipelineId, instanceId) {
    updateWizardPipeline(pipelineId, (pipeline) => ({ ...pipeline, filters: pipeline.filters.filter((item) => item.id !== instanceId) }))
  }

  function moveWizardFilter(pipelineId, instanceId, direction) {
    updateWizardPipeline(pipelineId, (pipeline) => {
      const next = [...pipeline.filters]
      const index = next.findIndex((item) => item.id === instanceId)
      if (index === -1) return pipeline
      const targetIndex = direction === 'up' ? index - 1 : index + 1
      if (targetIndex < 0 || targetIndex >= next.length) return pipeline
      const [moved] = next.splice(index, 1)
      next.splice(targetIndex, 0, moved)
      return { ...pipeline, filters: next }
    })
  }

  function addWizardRoute(pipelineId, moduleId) {
    const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'route')
    if (!module) return
    updateWizardPipeline(pipelineId, (pipeline) => ({
      ...pipeline,
      routes: [...(pipeline.routes || []), { id: createWizardInstanceID('wizard-route'), module_id: module.id, variables: buildWizardPipelineDraft(`route:${module.id}`, module, pipeline) }],
    }))
  }

  function removeWizardRoute(pipelineId, instanceId) {
    updateWizardPipeline(pipelineId, (pipeline) => ({ ...pipeline, routes: (pipeline.routes || []).filter((item) => item.id !== instanceId) }))
  }

  function moveWizardRoute(pipelineId, instanceId, direction) {
    updateWizardPipeline(pipelineId, (pipeline) => {
      const next = [...(pipeline.routes || [])]
      const index = next.findIndex((item) => item.id === instanceId)
      if (index === -1) return pipeline
      const targetIndex = direction === 'up' ? index - 1 : index + 1
      if (targetIndex < 0 || targetIndex >= next.length) return pipeline
      const [moved] = next.splice(index, 1)
      next.splice(targetIndex, 0, moved)
      return { ...pipeline, routes: next }
    })
  }

  function buildWizardOutputDraft(target, fluentType, pipeline = null, existingDraft = null, previousTag = '') {
    const outputModule = matchingOutputModuleForTarget(target, wizardEligibleModules.value, fluentType)
    const defaults = {
      ...parseVariablesMap(target?.settings),
      output_target_name: target?.name || '',
      output_target_type: target?.target_type || '',
    }
    return buildWizardPipelineDraft(`output:${target?.id}`, outputModule, pipeline, existingDraft, defaults, previousTag)
  }

  function addWizardOutputTarget(pipelineId, targetId) {
    const target = wizardAvailableOutputTargets.value.find((item) => item.id === targetId)
    if (!target) return
    updateWizardPipeline(pipelineId, (pipeline) => ({
      ...pipeline,
      outputs: [...pipeline.outputs, { id: createWizardInstanceID('wizard-output'), target_id: target.id, variables: buildWizardOutputDraft(target, wizardForm.fluent_type, pipeline) }],
    }))
  }

  function updateWizardOutputVariable(pipelineId, instanceId, key, value) {
    updateWizardPipeline(pipelineId, (pipeline) => ({
      ...pipeline,
      outputs: pipeline.outputs.map((instance) =>
        instance.id !== instanceId ? instance : {
          ...instance,
          variables: { ...instance.variables, [key]: value },
        }
      ),
    }))
  }

  function removeWizardOutput(pipelineId, instanceId) {
    updateWizardPipeline(pipelineId, (pipeline) => ({ ...pipeline, outputs: pipeline.outputs.filter((item) => item.id !== instanceId) }))
  }

  function moveWizardOutput(pipelineId, instanceId, direction) {
    updateWizardPipeline(pipelineId, (pipeline) => {
      const next = [...pipeline.outputs]
      const index = next.findIndex((item) => item.id === instanceId)
      if (index === -1) return pipeline
      const targetIndex = direction === 'up' ? index - 1 : index + 1
      if (targetIndex < 0 || targetIndex >= next.length) return pipeline
      const [moved] = next.splice(index, 1)
      next.splice(targetIndex, 0, moved)
      return { ...pipeline, outputs: next }
    })
  }

  function pruneWizardStateForRuntime() {
    const eligibleModuleIDs = new Set(wizardEligibleModules.value.map((item) => item.id))
    const availableTargetIDs = new Set(wizardAvailableOutputTargets.value.map((item) => item.id))
    if (wizardServiceModuleId.value && !eligibleModuleIDs.has(wizardServiceModuleId.value)) {
      wizardServiceModuleId.value = null
    }
    wizardParserModuleIds.value = wizardParserModuleIds.value.filter((id) => eligibleModuleIDs.has(id))
    wizardGlobalModuleVariables.value = Object.fromEntries(
      Object.entries(wizardGlobalModuleVariables.value).filter(([key]) => {
        const parts = String(key).split(':')
        return eligibleModuleIDs.has(Number(parts[1]))
      })
    )
    wizardPipelines.value = wizardPipelines.value.map((pipeline) => ({
      ...pipeline,
      input: pipeline.input && eligibleModuleIDs.has(pipeline.input.module_id) ? pipeline.input : null,
      filters: pipeline.filters.filter((instance) => eligibleModuleIDs.has(instance.module_id)),
      routes: (pipeline.routes || []).filter((instance) => eligibleModuleIDs.has(instance.module_id)),
      outputs: pipeline.outputs
        .filter((instance) => availableTargetIDs.has(instance.target_id))
        .map((instance) => {
          const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
          return { ...instance, variables: buildWizardOutputDraft(target, wizardForm.fluent_type, pipeline, instance.variables) }
        }),
    }))
    ensureWizardBaselineModules()
  }

  function removeWizardModuleReferences(moduleId) {
    if (wizardServiceModuleId.value === moduleId) wizardServiceModuleId.value = null
    wizardParserModuleIds.value = wizardParserModuleIds.value.filter((id) => id !== moduleId)
    wizardPipelines.value = wizardPipelines.value.map((pipeline) => ({
      ...pipeline,
      input: pipeline.input?.module_id === moduleId ? null : pipeline.input,
      filters: pipeline.filters.filter((instance) => instance.module_id !== moduleId),
      routes: (pipeline.routes || []).filter((instance) => instance.module_id !== moduleId),
    }))
    ensureWizardBaselineModules()
  }

  function removeWizardModuleReferencesBatch(moduleIds) {
    const deleted = new Set(moduleIds)
    if (wizardServiceModuleId.value && deleted.has(wizardServiceModuleId.value)) wizardServiceModuleId.value = null
    wizardParserModuleIds.value = wizardParserModuleIds.value.filter((id) => !deleted.has(id))
    wizardPipelines.value = wizardPipelines.value.map((pipeline) => ({
      ...pipeline,
      input: pipeline.input && deleted.has(pipeline.input.module_id) ? null : pipeline.input,
      filters: pipeline.filters.filter((instance) => !deleted.has(instance.module_id)),
      routes: (pipeline.routes || []).filter((instance) => !deleted.has(instance.module_id)),
    }))
    ensureWizardBaselineModules()
  }

  function applyWizardInputPreset(module) {
    if (!module) return
    const currentInputIDs = wizardSelectedModules.value.filter((item) => item.module_type === 'input').map((item) => item.id)
    if (selectedWizardModuleIds.value.includes(module.id)) {
      selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== module.id)
      return
    }
    selectedWizardModuleIds.value = [...selectedWizardModuleIds.value.filter((id) => !currentInputIDs.includes(id)), module.id]
    mergeWizardModuleVariableValues(module.id, parseVariablesMap(module.variables))
  }

  function toggleWizardOutputTarget(targetId) {
    if (selectedWizardOutputTargetIds.value.includes(targetId)) {
      selectedWizardOutputTargetIds.value = selectedWizardOutputTargetIds.value.filter((id) => id !== targetId)
      return
    }
    selectedWizardOutputTargetIds.value = [...selectedWizardOutputTargetIds.value, targetId]
  }

  function buildWizardRenderModuleRefs() {
    const refs = []
    if (wizardServiceModule.value) {
      refs.push(buildWizardModuleRef(wizardServiceModule.value, wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`]))
    }
    for (const module of wizardSelectedParserModules.value) {
      refs.push(buildWizardModuleRef(module, wizardGlobalModuleVariables.value[`parser:${module.id}`]))
    }
    for (const card of wizardRenderablePipelineCards.value) {
      const pipeline = wizardPipelines.value.find((item) => item.id === card.id)
      if (!pipeline) continue
      if (pipeline.input) {
        const inputModule = wizardEligibleModules.value.find((item) => item.id === pipeline.input.module_id)
        refs.push(buildWizardModuleRef(inputModule, pipeline.input.variables))
      }
      for (const instance of pipeline.filters) {
        const filterModule = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
        refs.push(buildWizardModuleRef(filterModule, instance.variables))
      }
      for (const instance of (pipeline.routes || [])) {
        const routeModule = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
        refs.push(buildWizardModuleRef(routeModule, instance.variables))
      }
      for (const instance of pipeline.outputs) {
        const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
        const outputModule = matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)
        refs.push(buildWizardModuleRef(outputModule, instance.variables, {
          ...parseVariablesMap(target?.settings),
          output_target_name: target?.name || '',
          output_target_type: target?.target_type || '',
        }))
      }
    }
    return refs.filter(Boolean)
  }

  function preparePreviewMetaFromWizard() {
    previewForm.name = wizardForm.name || `preview-${wizardForm.goal}`
    previewForm.fluent_type = wizardForm.fluent_type
    previewForm.runtime_version = wizardForm.runtime_version
    previewForm.variables = '{}'
  }

  async function runWizardPreview() {
    if (!wizardPipelines.value.length) { alert(t('configs_page.wizard_require_pipeline')); return }
    if (!wizardRenderablePipelineCards.value.length) {
      if (wizardIncompletePipelineLabels.value.length) {
        alert(t('configs_page.wizard_incomplete_pipelines').replace('{items}', wizardIncompletePipelineLabels.value.join(', ')))
        return
      }
      alert(t('configs_page.wizard_require_pipeline')); return
    }
    if (wizardOutputResolutionWarnings.value.length) {
      alert(wizardOutputResolutionWarnings.value.map((item) => `${item.pipeline}: ${item.target}`).join('\n'))
      return
    }
    preparePreviewMetaFromWizard()
    try {
      const previewRes = await previewRenderedConfig({
        name: previewForm.name,
        fluent_type: wizardForm.fluent_type,
        runtime_version: wizardForm.runtime_version,
        variables: '{}',
        modules: buildWizardRenderModuleRefs(),
      })
      const previewId = previewRes.data?.id
      if (previewId) {
        const detailRes = await getRenderedConfig(previewId)
        renderedConfig.value = detailRes.data
      } else {
        renderedConfig.value = previewRes.data
      }
      analysisResult.value = null
      compatibilityResult.value = null
      replayResult.value = null
      diffResult.value = null
    } catch (error) {
      alert(`${t('configs_page.generate_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function saveWizardAsTemplate() {
    if (!renderedConfig.value?.content) {
      alert(t('configs_page.require_preview').replace('{action}', wizardSaveButtonLabel.value)); return
    }
    const pipelineStateByID = new Map(wizardPipelines.value.map((pipeline) => [pipeline.id, pipeline]))
    const payload = {
      name: wizardForm.name || renderedConfig.value.name || `wizard-${wizardForm.goal}`,
      description: wizardForm.description || t('configs_page.wizard_default_description').replace('{goal}', wizardGoalLabel(wizardForm.goal)),
      fluent_type: wizardForm.fluent_type,
      content: renderedConfig.value.content,
      variables: previewForm.variables,
      source_type: 'module_assembly',
      source_modules: renderedConfig.value.source_modules || '[]',
      flow_layout: JSON.stringify({
        builder: 'wizard',
        goal: wizardForm.goal,
        runtime: wizardForm.fluent_type,
        global: {
          service_module_id: wizardServiceModule.value?.id || null,
          service_module: wizardServiceModule.value
            ? { id: wizardServiceModule.value.id, variables: normalizeWizardDraftValues(wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`], moduleVariablesForWizard(wizardServiceModule.value)) }
            : null,
          parser_module_ids: wizardSelectedParserModules.value.map((item) => item.id),
          parser_modules: wizardSelectedParserModules.value.map((item) => ({
            id: item.id,
            variables: normalizeWizardDraftValues(wizardGlobalModuleVariables.value[`parser:${item.id}`], moduleVariablesForWizard(item)),
          })),
        },
        pipelines: wizardPipelineCards.value.map((card) => {
          const pipelineState = pipelineStateByID.get(card.id)
          return {
            name: wizardPipelineDisplayName(card, card.index),
            complete: card.complete,
            input_module_id: card.inputModule?.id || null,
            input: card.inputModule
              ? { module_id: card.inputModule.id, variables: normalizeWizardDraftValues(pipelineState?.input?.variables, moduleVariablesForWizard(card.inputModule)) }
              : null,
            filter_module_ids: card.filterModules.map((item) => item.id),
            filters: (pipelineState?.filters || []).map((instance) => {
              const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
              return { module_id: instance.module_id, variables: normalizeWizardDraftValues(instance.variables, moduleVariablesForWizard(module)) }
            }),
            routes: (pipelineState?.routes || []).map((instance) => {
              const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
              return { module_id: instance.module_id, variables: normalizeWizardDraftValues(instance.variables, moduleVariablesForWizard(module)) }
            }),
            output_targets: card.outputTargets.map((target) => {
              const instance = (pipelineState?.outputs || []).find((output) => output.target_id === target.id)
              return {
                id: target.id, name: target.name, target_type: target.target_type,
                endpoint: target.endpoint, fluent_type: target.fluent_type,
                variables: normalizeWizardDraftValues(instance?.variables, {
                  ...moduleVariablesForWizard(matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)),
                  ...parseVariablesMap(target?.settings),
                  output_target_name: target?.name || '',
                  output_target_type: target?.target_type || '',
                }),
              }
            }),
          }
        }),
      }),
    }
    try {
      if (wizardLoadedFromTemplate.value?.id) {
        const templateID = Number(wizardLoadedFromTemplate.value.id)
        const { data: updatedTemplate } = await updateTemplate(templateID, payload)
        await createVersion(templateID, {
          content: renderedConfig.value.content,
          comment: t('configs_page.wizard_version_comment').replace('{goal}', wizardGoalLabel(wizardForm.goal)),
        })
        wizardLoadedFromTemplate.value = updatedTemplate || wizardLoadedFromTemplate.value
        await loadTemplates()
        await router.push(`/configs/${templateID}`)
        return
      }
      await createTemplate(payload)
      await loadTemplates()
      activeTab.value = 'templates'
    } catch (error) {
      alert(`${wizardLoadedFromTemplate.value ? t('configs_page.create_version_failed') : t('configs_page.create_template_failed')}: ${getErrorMessage(error)}`)
    }
  }

  function openAdvancedPreviewFromWizard() {
    preparePreviewMetaFromWizard()
    activeTab.value = 'preview'
  }

  function wizardHasContent() {
    if (wizardServiceModuleId.value) return true
    if (wizardParserModuleIds.value.length) return true
    return wizardPipelines.value.some((p) => p.input || p.filters.length || (p.routes || []).length || p.outputs.length)
  }

  function loadWizardFromTemplate(template) {
    if (!template?.flow_layout) { alert(t('configs_page.wizard_load_incompatible')); return }
    let layout
    try {
      layout = typeof template.flow_layout === 'string' ? JSON.parse(template.flow_layout) : template.flow_layout
    } catch { alert(t('configs_page.wizard_load_incompatible')); return }
    if (layout.builder !== 'wizard') { alert(t('configs_page.wizard_load_incompatible')); return }
    if (wizardHasContent() && !confirm(t('configs_page.wizard_load_overwrite_confirm'))) return

    wizardPipelines.value = []
    wizardServiceModuleId.value = null
    wizardParserModuleIds.value = []
    wizardGlobalModuleVariables.value = {}
    renderedConfig.value = null

    if (layout.runtime) wizardForm.fluent_type = layout.runtime
    if (layout.goal) wizardForm.goal = layout.goal
    wizardForm.name = template.name || ''
    wizardForm.description = template.description || ''

    const restoredServiceModuleID = layout.global?.service_module?.id || layout.global?.service_module_id
    if (restoredServiceModuleID) {
      selectWizardServiceModule(restoredServiceModuleID)
      const module = wizardEligibleModules.value.find((item) => item.id === restoredServiceModuleID && item.module_type === 'service')
      const restoredVariables = layout.global?.service_module?.variables
      if (module && restoredVariables && typeof restoredVariables === 'object') {
        wizardGlobalModuleVariables.value = {
          ...wizardGlobalModuleVariables.value,
          [`service:${restoredServiceModuleID}`]: ensureWizardModuleDraft(`service:${restoredServiceModuleID}`, module, restoredVariables),
        }
      }
    }
    const restoredParserModules = Array.isArray(layout.global?.parser_modules)
      ? layout.global.parser_modules
      : (layout.global?.parser_module_ids || []).map((id) => ({ id }))
    for (const parserEntry of restoredParserModules) {
      const moduleID = Number(parserEntry?.id || parserEntry)
      if (!moduleID) continue
      toggleWizardParserModule(moduleID)
      const module = wizardEligibleModules.value.find((item) => item.id === moduleID && item.module_type === 'parser')
      const restoredVariables = parserEntry?.variables
      if (module && restoredVariables && typeof restoredVariables === 'object') {
        wizardGlobalModuleVariables.value = {
          ...wizardGlobalModuleVariables.value,
          [`parser:${moduleID}`]: ensureWizardModuleDraft(`parser:${moduleID}`, module, restoredVariables),
        }
      }
    }

    const inputModuleIDForLayout = (p) => p.input?.module_id || p.input_module_id
    const restored = (layout.pipelines || []).map((p) => {
      const pid = createWizardInstanceID('wizard-pipeline')
      const inputModuleID = inputModuleIDForLayout(p)
      const inputModule = wizardEligibleModules.value.find((item) => item.id === inputModuleID && item.module_type === 'input')
      return {
        id: pid, name: p.name || '',
        input: inputModuleID ? { id: createWizardInstanceID('wizard-input'), module_id: inputModuleID, variables: ensureWizardModuleDraft(`input:${inputModuleID}`, inputModule, p.input?.variables) } : null,
        filters: (Array.isArray(p.filters) ? p.filters : (p.filter_module_ids || []).map((mid) => ({ module_id: mid })))
          .map((entry) => {
            const moduleID = Number(entry?.module_id || entry)
            const module = wizardEligibleModules.value.find((item) => item.id === moduleID && item.module_type === 'filter')
            if (!moduleID) return null
            return { id: createWizardInstanceID('wizard-filter'), module_id: moduleID, variables: ensureWizardModuleDraft(`filter:${moduleID}`, module, entry?.variables) }
          }).filter(Boolean),
        routes: (Array.isArray(p.routes) ? p.routes : [])
          .map((entry) => {
            const moduleID = Number(entry?.module_id || entry)
            const module = wizardEligibleModules.value.find((item) => item.id === moduleID && item.module_type === 'route')
            if (!moduleID) return null
            return { id: createWizardInstanceID('wizard-route'), module_id: moduleID, variables: ensureWizardModuleDraft(`route:${moduleID}`, module, entry?.variables) }
          }).filter(Boolean),
        outputs: (p.output_targets || [])
          .map((ot) => {
            const targetID = Number(ot?.id)
            const target = wizardAvailableOutputTargets.value.find((item) => item.id === targetID)
            if (!targetID) return null
            return {
              id: createWizardInstanceID('wizard-output'), target_id: targetID,
              variables: buildWizardOutputDraft(target, wizardForm.fluent_type, { input: inputModuleID ? { module_id: inputModuleID, variables: p.input?.variables } : null }, ot?.variables),
            }
          }).filter(Boolean),
      }
    })
    wizardPipelines.value = restored.length ? restored : [createWizardPipeline()]
    activeWizardPipelineId.value = wizardPipelines.value[0].id
    wizardLoadedFromTemplate.value = template
  }

  function clearWizardLoadedTemplate() { wizardLoadedFromTemplate.value = null }

  function addWizardPipelineFromSaved(pipelineId) {
    const saved = pipelines.value.find((p) => p.id === Number(pipelineId))
    if (!saved) return
    const newPipeline = createWizardPipeline()
    newPipeline.input = saved.input_module_id ? { id: createWizardInstanceID('wizard-input'), module_id: saved.input_module_id } : null as any
    newPipeline.filters = (saved.filter_modules || []).map((m) => ({ id: createWizardInstanceID('wizard-filter'), module_id: m.id }))
    newPipeline.outputs = (saved.output_targets || []).map((t) => ({ id: createWizardInstanceID('wizard-output'), target_id: t.id }))
    wizardPipelines.value.push(newPipeline)
    activeWizardPipelineId.value = newPipeline.id
  }

  function resetWizardForm() {
    wizardForm.goal = 'edge_collection'
    wizardForm.name = ''
    wizardForm.description = ''
    wizardForm.fluent_type = 'fluentbit'
    wizardForm.runtime_version = ''
    wizardServiceModuleId.value = null
    wizardParserModuleIds.value = []
    wizardGlobalModuleVariables.value = {}
    wizardServiceSearch.value = ''
    wizardParserSearch.value = ''
    wizardInputSearch.value = ''
    wizardFilterSearch.value = ''
    wizardRouteSearch.value = ''
    wizardOutputSearch.value = ''
    wizardStagePages.service = 1
    wizardStagePages.parser = 1
    wizardStagePages.input = 1
    wizardStagePages.filter = 1
    wizardStagePages.route = 1
    wizardStagePages.output = 1
    const initialPipeline = createWizardPipeline()
    wizardPipelines.value = [initialPipeline]
    activeWizardPipelineId.value = initialPipeline.id
    selectedWizardModuleIds.value = []
    selectedWizardOutputTargetIds.value = []
    wizardModuleVariableValues.value = {}
  }

  // --- Watchers ---

  watch(
    () => wizardForm.fluent_type,
    (runtime) => {
      if (!wizardForm.name) wizardForm.name = `guided-${runtime}`
      const eligibleIds = new Set(wizardEligibleModules.value.map((item) => item.id))
      selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => eligibleIds.has(id))
      wizardModuleVariableValues.value = Object.fromEntries(
        Object.entries(wizardModuleVariableValues.value).filter(([moduleId]) => eligibleIds.has(Number(moduleId)))
      )
      const availableIds = new Set(wizardAvailableOutputTargets.value.map((item) => item.id))
      selectedWizardOutputTargetIds.value = selectedWizardOutputTargetIds.value.filter((id) => availableIds.has(id))
      ensureWizardBaselineModules()
      pruneWizardStateForRuntime()
    },
    { immediate: true }
  )

  watch(selectedWizardModuleIds, () => {
    if (!wizardForm.description) {
      wizardForm.description = t('configs_page.wizard_default_description').replace('{goal}', wizardGoalLabel(wizardForm.goal))
    }
  })

  watch(
    wizardVariableGroups,
    (groups) => {
      const next = {}
      for (const group of groups) {
        const existingValues = wizardModuleVariableValues.value[group.moduleId] || {}
        next[group.moduleId] = {}
        for (const field of group.fields) {
          next[group.moduleId][field.key] = Object.prototype.hasOwnProperty.call(existingValues, field.key)
            ? existingValues[field.key]
            : field.defaultValue
        }
      }
      wizardModuleVariableValues.value = next
    },
    { immediate: true }
  )

  return {
    // state
    wizardForm, wizardServiceModuleId, wizardParserModuleIds, wizardGlobalModuleVariables,
    wizardPipelines, activeWizardPipelineId, wizardLoadedFromTemplate,
    wizardServiceSearch, wizardParserSearch, wizardInputSearch, wizardFilterSearch, wizardRouteSearch, wizardOutputSearch, wizardModuleSearch,
    wizardStagePages, selectedWizardModuleIds, selectedWizardOutputTargetIds, wizardModuleVariableValues,
    wizardPipelineModuleTypes, wizardPipelineStageTotal,
    // computed
    wizardSaveButtonLabel, wizardEligibleModules, wizardServiceModules, wizardParserModules,
    wizardInputModules, wizardFilterModules, wizardRouteModules, wizardVisibleModules,
    wizardInputPresets, wizardAvailableOutputTargets, wizardOutputTargets, wizardCompatiblePipelines,
    wizardPagedServiceModules, wizardPagedParserModules, wizardPagedInputModules, wizardPagedFilterModules, wizardPagedRouteModules, wizardPagedOutputTargets,
    wizardServiceModule, wizardSelectedParserModules, activeWizardPipeline,
    wizardPipelineCards, wizardRenderablePipelineCards, wizardIncompletePipelineLabels, wizardRenderSummary,
    wizardGlobalVariableGroups, wizardPipelineVariableGroups, wizardOutputResolutionWarnings,
    wizardSelectedModules, selectedWizardInputPresetKeys, wizardSelectedOutputTargets, wizardUnresolvedOutputTargets,
    wizardInputPresetsSelected, wizardSelectedInputModule, wizardSelectedFilterModules, wizardSelectedOutputModules,
    wizardSummaryModules, wizardFlowSummary, wizardModulesByType,
    wizardVariableGroups, wizardVariableFieldCount, wizardPipelineCompletedStages, wizardMissingRequirements,
    // functions
    matchingWizardParserModules, autoAttachWizardParsersForInputModule,
    wizardPipelineInputTag, wizardPipelineModuleDefaults, buildWizardPipelineDraft,
    mergeWizardModuleVariableValues, ensureWizardBaselineModules,
    changeWizardStagePage, resetWizardStagePage,
    selectWizardServiceModule, toggleWizardParserModule,
    addWizardPipeline, duplicateWizardPipeline, removeWizardPipeline, selectWizardPipeline, updateWizardPipeline,
    setWizardPipelineInput, addWizardFilter, removeWizardFilter, moveWizardFilter,
    addWizardRoute, removeWizardRoute, moveWizardRoute,
    buildWizardOutputDraft, addWizardOutputTarget, updateWizardOutputVariable, removeWizardOutput, moveWizardOutput,
    pruneWizardStateForRuntime, removeWizardModuleReferences, removeWizardModuleReferencesBatch,
    applyWizardInputPreset, toggleWizardOutputTarget,
    buildWizardRenderModuleRefs, preparePreviewMetaFromWizard,
    runWizardPreview, saveWizardAsTemplate, openAdvancedPreviewFromWizard,
    wizardHasContent, loadWizardFromTemplate, clearWizardLoadedTemplate,
    addWizardPipelineFromSaved, resetWizardForm,
    wizardPipelineDisplayName,
  }
}
