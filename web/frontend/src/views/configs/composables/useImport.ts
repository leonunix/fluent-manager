// Import composable – state and functions for importing existing configs.
// Deps injected: { modules, outputTargets, templates, t, previewForm, renderedConfig,
//   analysisResult, compatibilityResult, replayResult, diffResult,
//   selectedPreviewModuleIds, selectedPreviewOutputTargetIds, activeTab,
//   loadModules, loadModuleTable, loadOutputTargets, loadTemplates, router }

import { computed, reactive, ref } from 'vue'
import {
  importExistingConfig, createModule, createOutputTarget, createTemplate,
  previewRenderedConfig, getRenderedConfig,
} from '../../../api'
import { getErrorMessage as getErrorMessageFn } from '../utils/config-format'
import { normalizeName, generateUniqueDraftName } from '../utils/config-text'
import {
  importedModuleIdentity, uniqueImportedModuleName,
  findImportedModuleNameConflict, uniqueImportedOutputTargetName,
  createImportedOutputTargetSignature, findImportedOutputAdapterModule,
  buildImportedOutputRenderVariables, inferImportedOutputTargetType,
  buildImportedOutputTargetDraft, buildImportedModuleDescription as buildImportedModuleDescriptionFn,
  uniqueImportedDestinationList,
} from '../utils/config-import-utils'

export function useImport({
  modules, outputTargets, templates, t,
  previewForm, renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult,
  selectedPreviewModuleIds, selectedPreviewOutputTargetIds, activeTab,
  loadModules, loadModuleTable, loadOutputTargets, loadTemplates, router,
}: any) {
  function getErrorMessage(e: any) { return getErrorMessageFn(e, t) }
  function buildImportedModuleDescription(item: any) { return buildImportedModuleDescriptionFn(item, t) }

  const importAnalysisLoading = ref(false)
  const importModulesLoading = ref(false)
  const importedConfigResult = ref<any>(null)
  const importedWorkspaceModules = ref<any[]>([])
  const importedWorkspaceTemplate = ref<any>(null)

  const importForm = reactive({
    fluent_type: 'fluentbit',
    name_prefix: 'imported-config',
    content: '',
  })

  // --- Computed ---

  const importBlockingIssueCount = computed(() =>
    (importedConfigResult.value?.modules || []).filter((module: any) => !!importedModuleNameIssue(module)).length
  )
  const importFlowPathLabel = computed(() =>
    importedConfigResult.value?.flow_path?.length
      ? importedConfigResult.value.flow_path.join(' -> ')
      : t('configs_page.no_solution_path')
  )
  const importSemanticChangeCount = computed(() => importedConfigResult.value?.validation?.semantic_diff?.changes?.length || 0)
  const importReusableMatchCount = computed(() =>
    importedConfigResult.value?.modules?.filter((m: any) => m.module_type !== 'output' && m.existing_module_id).length || 0
  )
  const importDestinationMatchCount = computed(() => importedConfigResult.value?.destinations?.length || 0)
  const importReuseDecisionCount = computed(() =>
    importedConfigResult.value?.modules?.filter((m: any) => m.module_type !== 'output' && m.import_action === 'reuse_existing').length || 0
  )
  const importCreateDecisionCount = computed(() =>
    importedConfigResult.value?.modules?.filter((m: any) => m.module_type !== 'output' && m.import_action !== 'reuse_existing').length || 0
  )

  // --- Functions ---

  function importedModuleNameIssue(module: any) {
    if (!module || module.module_type === 'output' || module.import_action === 'reuse_existing') return null
    const name = String(module.name || '').trim()
    if (!name) return { type: 'required', message: t('configs_page.import_module_name_required').replace('{order}', String(module.order || '')) }
    const identity = importedModuleIdentity(name, module.module_type, module.fluent_type)
    const duplicateInBatch = (importedConfigResult.value?.modules || []).filter((item: any) => {
      if (item.module_type === 'output' || item.import_action === 'reuse_existing') return false
      return importedModuleIdentity(String(item.name || '').trim(), item.module_type, item.fluent_type) === identity
    }).length > 1
    if (duplicateInBatch) return { type: 'batch_duplicate', message: t('configs_page.import_module_name_duplicate_batch').replace('{name}', name) }
    if (module.module_type === 'parser') {
      const existsInWorkspace = modules.value.some((item: any) =>
        importedModuleIdentity(item.name, item.module_type, item.fluent_type) === identity
      )
      if (existsInWorkspace) return { type: 'existing_duplicate', message: t('configs_page.import_module_name_duplicate_existing').replace('{name}', name) }
    }
    return null
  }

  function importedOutputTargetNameSeed(module: any, fallbackPrefix: string) {
    const type = inferImportedOutputTargetType(module) || 'custom'
    const prefix = String(fallbackPrefix || importForm.name_prefix || 'imported-config').trim() || 'imported-config'
    return `${prefix}-${type}-destination`
  }

  function setAllImportedModuleActions(action: string) {
    if (!importedConfigResult.value?.modules?.length) return
    for (const module of importedConfigResult.value.modules) {
      if (module.module_type === 'output') continue
      if (action === 'reuse_existing' && !module.existing_module_id) { module.import_action = 'create_new'; continue }
      module.import_action = action
    }
  }

  async function runImportAnalysis() {
    if (!importForm.content.trim()) { importedConfigResult.value = null; return }
    importAnalysisLoading.value = true
    importedWorkspaceModules.value = []
    importedWorkspaceTemplate.value = null
    try {
      const { data } = await importExistingConfig({
        fluent_type: importForm.fluent_type,
        name_prefix: importForm.name_prefix,
        content: importForm.content,
      })
      importedConfigResult.value = data
    } catch (error) {
      importedConfigResult.value = null
      alert(`${t('configs_page.import_failed')}: ${getErrorMessage(error)}`)
    } finally {
      importAnalysisLoading.value = false
    }
  }

  async function importParsedModules() {
    if (!importedConfigResult.value?.modules?.length) return
    importModulesLoading.value = true
    try {
      const invalidNamedModule = importedConfigResult.value.modules.find((item: any) =>
        item.module_type !== 'output' && item.import_action !== 'reuse_existing' && !String(item.name || '').trim()
      )
      if (invalidNamedModule) throw new Error(t('configs_page.import_module_name_required').replace('{order}', String(invalidNamedModule.order)))

      // Fetch fresh full module list for name conflict detection
      const pageSize = 100
      const collected: any[] = []
      let page = 1, total = 0
      do {
        const { data: d } = await (await import('../../../api')).getModules({ page, page_size: pageSize })
        const batch = d.data || []; total = Number(d.total || 0); collected.push(...batch)
        if (!batch.length) break; page += 1
      } while (collected.length < total)
      const existingModules = collected

      const existingOutputTargets = [...outputTargets.value]
      const occupiedIdentities = new Set(existingModules.map((item: any) => importedModuleIdentity(item.name, item.module_type, item.fluent_type)))
      const importNameConflict = findImportedModuleNameConflict(importedConfigResult.value.modules, occupiedIdentities)
      if (importNameConflict?.type === 'batch_duplicate') throw new Error(t('configs_page.import_module_name_duplicate_batch').replace('{name}', String(importNameConflict.item?.name || '').trim()))
      if (importNameConflict?.type === 'existing_duplicate') throw new Error(t('configs_page.import_module_name_duplicate_existing').replace('{name}', String(importNameConflict.item?.name || '').trim()))

      const occupiedOutputTargetNames = new Set(existingOutputTargets.map((item: any) => normalizeName(item.name)).filter(Boolean))
      const reusableImportedOutputTargets = new Map()
      for (const target of existingOutputTargets) reusableImportedOutputTargets.set(createImportedOutputTargetSignature(target), target)

      const created: any[] = []
      const assembledModuleRefs: any[] = []
      const assembledModules: any[] = []
      const ensuredDestinations = Array.isArray(importedConfigResult.value.destinations)
        ? importedConfigResult.value.destinations.map((item: any) => ({ ...item }))
        : []

      for (const item of importedConfigResult.value.modules) {
        if (item.module_type === 'output') {
          let ensuredTarget = null
          if (item.output_target_id) {
            ensuredTarget = existingOutputTargets.find((target: any) => target.id === item.output_target_id) || null
          }
          if (!ensuredTarget) {
            const targetDraft = buildImportedOutputTargetDraft(item)
            const targetSignature = createImportedOutputTargetSignature(targetDraft)
            ensuredTarget = reusableImportedOutputTargets.get(targetSignature) || null
            if (!ensuredTarget) {
              const targetName = uniqueImportedOutputTargetName(
                importedOutputTargetNameSeed(item, importForm.name_prefix || importedConfigResult.value.name_prefix),
                occupiedOutputTargetNames
              )
              ensuredTarget = await createOutputTarget({
                name: targetName, description: buildImportedModuleDescription(item),
                fluent_type: 'shared', target_type: targetDraft.target_type,
                endpoint: targetDraft.endpoint, settings: targetDraft.settings,
              })
              existingOutputTargets.push(ensuredTarget)
              reusableImportedOutputTargets.set(targetSignature, ensuredTarget)
            }
          }
          item.output_target_id = ensuredTarget.id
          item.output_target_name = ensuredTarget.name
          item.output_target_type = ensuredTarget.target_type
          item.output_target_endpoint = ensuredTarget.endpoint
          item.output_target_match_type = item.output_target_match_type || 'created'
          const adapterModule = findImportedOutputAdapterModule(ensuredTarget.target_type, existingModules, importForm.fluent_type)
          if (!adapterModule) throw new Error(t('configs_page.import_output_adapter_missing').replace('{type}', ensuredTarget.target_type || 'output'))
          assembledModuleRefs.push({
            module_id: adapterModule.id,
            variables: JSON.stringify(buildImportedOutputRenderVariables(item, ensuredTarget, importForm.fluent_type), null, 2),
          })
          ensuredDestinations.push({
            output_module_name: adapterModule.name, output_module_order: item.order,
            output_target_id: ensuredTarget.id, name: ensuredTarget.name,
            target_type: ensuredTarget.target_type, endpoint: ensuredTarget.endpoint,
            match_type: item.output_target_match_type || 'created',
          })
          continue
        }

        if (item.import_action === 'reuse_existing' && item.existing_module_id) {
          assembledModuleRefs.push({ module_id: item.existing_module_id, variables: item.variables || '{}' })
          assembledModules.push({ id: Number(item.existing_module_id), module_type: item.module_type })
          continue
        }

        const name = uniqueImportedModuleName(item.name, item.module_type, item.fluent_type, occupiedIdentities)
        const { data } = await createModule({
          name, description: buildImportedModuleDescription(item),
          module_type: item.module_type, fluent_type: item.fluent_type,
          content: item.content, variables: item.variables || '{}', is_builtin: false,
        })
        created.push(data)
        assembledModuleRefs.push({ module_id: data.id, variables: item.variables || '{}' })
        assembledModules.push({ id: Number(data.id), module_type: item.module_type })
      }

      importedWorkspaceModules.value = created
      await Promise.all([loadModules(), loadModuleTable()])
      outputTargets.value = existingOutputTargets
      importedConfigResult.value.destinations = uniqueImportedDestinationList(ensuredDestinations)

      previewForm.name = importedConfigResult.value.suggested_template_name || `${importForm.name_prefix || 'imported-config'}-assembly`
      previewForm.fluent_type = importForm.fluent_type
      previewForm.runtime_version = ''
      previewForm.variables = '{}'
      selectedPreviewOutputTargetIds.value = []
      selectedPreviewModuleIds.value = assembledModules.filter((item) => item.module_type !== 'output').map((item) => item.id)

      if (!importedConfigResult.value.auto_assemble_supported) {
        renderedConfig.value = null
        analysisResult.value = null
        compatibilityResult.value = null
        replayResult.value = null
        diffResult.value = null
        importedWorkspaceTemplate.value = null
        importedConfigResult.value = null
        activeTab.value = 'modules'
        alert(t('configs_page.import_success_assets').replace('{count}', String(created.length)))
        return
      }

      const previewRes = await previewRenderedConfig({
        name: previewForm.name, fluent_type: previewForm.fluent_type,
        runtime_version: previewForm.runtime_version, variables: previewForm.variables,
        modules: assembledModuleRefs,
      })
      const previewId = previewRes.data?.id
      if (previewId) {
        const detailRes = await getRenderedConfig(previewId)
        renderedConfig.value = detailRes.data
      } else {
        renderedConfig.value = previewRes.data
      }
      analysisResult.value = null; compatibilityResult.value = null; replayResult.value = null; diffResult.value = null

      const templateName = generateUniqueDraftName(
        importedConfigResult.value.suggested_template_name || `${importForm.name_prefix || 'imported-config'}-assembly`,
        templates.value.map((item: any) => item.name),
        `${importForm.name_prefix || 'imported-config'}-assembly`
      )
      const matchedExistingCount = importedConfigResult.value.modules.filter((item: any) => item.existing_module_id).length
      const reusedExistingCount = importedConfigResult.value.modules.filter((item: any) => item.import_action === 'reuse_existing').length
      const { data: templateData } = await createTemplate({
        name: templateName,
        description: t('configs_page.import_template_description').replace('{prefix}', importForm.name_prefix || 'imported-config'),
        fluent_type: previewForm.fluent_type,
        content: renderedConfig.value?.content || importedConfigResult.value.template_draft_content || '',
        variables: previewForm.variables,
        source_type: 'module_assembly',
        source_modules: renderedConfig.value?.source_modules || '[]',
        flow_layout: JSON.stringify({
          ...(importedConfigResult.value.flow_layout || {}),
          matched_existing_count: matchedExistingCount,
          reused_existing_count: reusedExistingCount,
          destinations: uniqueImportedDestinationList(ensuredDestinations),
          validation: importedConfigResult.value.validation || {},
        }),
      })
      importedWorkspaceTemplate.value = templateData
      await loadTemplates()
      await router.push(`/configs/${templateData.id}`)
    } catch (error) {
      alert(`${t('configs_page.import_persist_failed')}: ${getErrorMessage(error)}`)
    } finally {
      importModulesLoading.value = false
    }
  }

  return {
    importForm, importAnalysisLoading, importModulesLoading,
    importedConfigResult, importedWorkspaceModules, importedWorkspaceTemplate,
    importBlockingIssueCount, importFlowPathLabel, importSemanticChangeCount,
    importReusableMatchCount, importDestinationMatchCount,
    importReuseDecisionCount, importCreateDecisionCount,
    importedModuleNameIssue, importedOutputTargetNameSeed, setAllImportedModuleActions,
    runImportAnalysis, importParsedModules,
  }
}
