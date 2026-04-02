// Preview and analysis composable – state, computed, and functions for the preview tab.
// External deps injected: { modules, outputTargets, activeTab, t }

import { computed, reactive, ref, watch } from 'vue'
import { buildConfigFlowSummary } from '../../../utils/config_flow'
import {
  previewRenderedConfig, getRenderedConfig,
  lintConfig, checkCompatibility, replayConfig, diffConfig,
} from '../../../api'
import { matchesModuleSearch } from '../utils/config-search'
import { matchingOutputModuleForTarget, buildOutputTargetModuleRefs } from '../utils/config-wizard-helpers'
import { getErrorMessage as getErrorMessageFn } from '../utils/config-format'

export function usePreviewAnalysis({ modules, outputTargets, activeTab, t }) {
  function getErrorMessage(error) { return getErrorMessageFn(error, t) }

  // --- State ---

  const renderedConfig = ref(null)
  const analysisResult = ref(null)
  const compatibilityResult = ref(null)
  const replayResult = ref(null)
  const diffResult = ref(null)
  const selectedPreviewModuleIds = ref([])
  const selectedPreviewOutputTargetIds = ref([])
  const previewModuleSearch = ref('')
  const previewModuleVariables = ref({})

  const previewForm = reactive({
    name: '',
    fluent_type: 'fluentbit',
    runtime_version: '',
    variables: '{}',
    node_id: '',
    sample_log: '',
    sample_tag: '',
    diff_content: '',
  })

  // --- Computed ---

  const previewEligibleModules = computed(() =>
    modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
  )
  const previewVisibleModules = computed(() =>
    previewEligibleModules.value.filter((item) => item.module_type !== 'output' && matchesModuleSearch(item, previewModuleSearch.value))
  )
  const previewAvailableOutputTargets = computed(() =>
    outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
  )
  const previewSelectedModules = computed(() =>
    previewEligibleModules.value.filter((item) => selectedPreviewModuleIds.value.includes(item.id))
  )
  const previewResolvedOutputTargets = computed(() =>
    previewAvailableOutputTargets.value.filter((item) => selectedPreviewOutputTargetIds.value.includes(item.id))
  )
  const previewUnresolvedOutputTargets = computed(() =>
    previewResolvedOutputTargets.value.filter((item) => !matchingOutputModuleForTarget(item, previewEligibleModules.value, previewForm.fluent_type))
  )
  const previewSelectedOutputModules = computed(() =>
    previewResolvedOutputTargets.value
      .map((target) => matchingOutputModuleForTarget(target, previewEligibleModules.value, previewForm.fluent_type))
      .filter(Boolean)
  )
  const previewSummaryModules = computed(() => [...previewSelectedModules.value, ...previewSelectedOutputModules.value])
  const previewFlowSummary = computed(() => buildConfigFlowSummary(previewSummaryModules.value, previewResolvedOutputTargets.value))

  // --- Functions ---

  function togglePreviewModule(moduleId) {
    if (selectedPreviewModuleIds.value.includes(moduleId)) {
      selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => id !== moduleId)
      return
    }
    selectedPreviewModuleIds.value = [...selectedPreviewModuleIds.value, moduleId]
  }

  function togglePreviewOutputTarget(targetId) {
    if (selectedPreviewOutputTargetIds.value.includes(targetId)) {
      selectedPreviewOutputTargetIds.value = selectedPreviewOutputTargetIds.value.filter((id) => id !== targetId)
      return
    }
    selectedPreviewOutputTargetIds.value = [...selectedPreviewOutputTargetIds.value, targetId]
  }

  function buildPreviewSelectedModuleRefs() {
    return selectedPreviewModuleIds.value.map((moduleId) => {
      const variables = previewModuleVariables.value[moduleId]
      if (!variables || variables === '{}') return { module_id: moduleId }
      return { module_id: moduleId, variables }
    })
  }

  async function runPreview(options: { switchTab?: boolean } = {}) {
    const outputModuleRefs = buildOutputTargetModuleRefs(previewResolvedOutputTargets.value, previewEligibleModules.value, previewForm.fluent_type)
    if (previewUnresolvedOutputTargets.value.length) {
      alert(t('configs_page.output_target_module_missing').replace('{targets}', previewUnresolvedOutputTargets.value.map((item) => item.name).join(', ')))
      return
    }
    if (!selectedPreviewModuleIds.value.length && !outputModuleRefs.length) {
      alert(t('configs_page.choose_modules')); return
    }
    try {
      const payload = {
        name: previewForm.name,
        fluent_type: previewForm.fluent_type,
        runtime_version: previewForm.runtime_version,
        variables: previewForm.variables,
        modules: [...buildPreviewSelectedModuleRefs(), ...outputModuleRefs],
      }
      const previewRes = await previewRenderedConfig(payload)
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
      if (options.switchTab !== false) activeTab.value = 'preview'
    } catch (error) {
      alert(`${t('configs_page.generate_failed')}: ${getErrorMessage(error)}`)
    }
  }

  function requireRenderedConfig(actionLabel) {
    if (!renderedConfig.value?.content) {
      alert(t('configs_page.require_preview').replace('{action}', actionLabel))
      return null
    }
    return renderedConfig.value.content
  }

  async function runLint() {
    const content = requireRenderedConfig(t('configs_page.run_lint'))
    if (!content) return
    try {
      analysisResult.value = await lintConfig({ fluent_type: renderedConfig.value.fluent_type, runtime_version: renderedConfig.value.runtime_version, content })
      activeTab.value = 'preview'
    } catch (error) {
      alert(`${t('configs_page.lint_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function runCompatibility() {
    const content = requireRenderedConfig(t('configs_page.run_compatibility'))
    if (!content) return
    try {
      const payload: { fluent_type: string; runtime_version?: string; content: string; node_id?: number } = { fluent_type: renderedConfig.value.fluent_type, runtime_version: renderedConfig.value.runtime_version, content }
      if (previewForm.node_id) payload.node_id = Number(previewForm.node_id)
      compatibilityResult.value = await checkCompatibility(payload)
      activeTab.value = 'preview'
    } catch (error) {
      alert(`${t('configs_page.compatibility_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function runReplay() {
    const content = requireRenderedConfig(t('configs_page.run_replay'))
    if (!content) return
    if (!previewForm.sample_log) { alert(t('configs_page.require_sample_log')); return }
    try {
      replayResult.value = await replayConfig({
        fluent_type: renderedConfig.value.fluent_type, runtime_version: renderedConfig.value.runtime_version,
        content, sample_log: previewForm.sample_log, sample_tag: previewForm.sample_tag,
      })
      activeTab.value = 'preview'
    } catch (error) {
      alert(`${t('configs_page.replay_failed')}: ${getErrorMessage(error)}`)
    }
  }

  async function runSemanticDiff() {
    const content = requireRenderedConfig(t('configs_page.run_diff'))
    if (!content) return
    if (!previewForm.diff_content) { alert(t('configs_page.require_diff_content')); return }
    try {
      diffResult.value = await diffConfig({ fluent_type: renderedConfig.value.fluent_type, before_content: previewForm.diff_content, after_content: content })
      activeTab.value = 'preview'
    } catch (error) {
      alert(`${t('configs_page.diff_failed')}: ${getErrorMessage(error)}`)
    }
  }

  // --- Watchers ---

  watch(
    () => previewForm.fluent_type,
    () => {
      const eligibleIds = new Set(previewEligibleModules.value.map((item) => item.id))
      selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => eligibleIds.has(id))
      previewModuleVariables.value = Object.fromEntries(
        Object.entries(previewModuleVariables.value).filter(([moduleId]) => eligibleIds.has(Number(moduleId)))
      )
      const availableIds = new Set(previewAvailableOutputTargets.value.map((item) => item.id))
      selectedPreviewOutputTargetIds.value = selectedPreviewOutputTargetIds.value.filter((id) => availableIds.has(id))
    }
  )

  return {
    // state
    renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult,
    previewForm, selectedPreviewModuleIds, selectedPreviewOutputTargetIds,
    previewModuleSearch, previewModuleVariables,
    // computed
    previewEligibleModules, previewVisibleModules, previewAvailableOutputTargets,
    previewSelectedModules, previewResolvedOutputTargets, previewUnresolvedOutputTargets,
    previewSelectedOutputModules, previewSummaryModules, previewFlowSummary,
    // functions
    togglePreviewModule, togglePreviewOutputTarget, buildPreviewSelectedModuleRefs,
    runPreview, runLint, runCompatibility, runReplay, runSemanticDiff,
  }
}
