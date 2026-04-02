// AI Assistant composable – state and functions for AI-assisted config generation.
// Deps injected: { modules, outputTargets, t, loadModules, loadModuleTable, loadPipelines }
// Cross-domain actions (useAIModuleDraft, useAITemplateDraft, sendAIPipelineToWizard) stay in ConfigsPage.

import { reactive, ref } from 'vue'
import { analyzeLogSampleAssistant, createModule, createModuleVersion } from '../../../api'
import { createConfigPipeline } from '../../../api/configs'
import { getErrorMessage as getErrorMessageFn, getProviderErrorMessage } from '../utils/config-format'
import { mergeAIModules } from '../utils/config-ai-draft'

export function useAIAssistant({ modules, outputTargets, t, loadModules, loadModuleTable, loadPipelines }: any) {
  function getErrorMessage(e: any) { return getErrorMessageFn(e, t) }

  const aiAssistantLoading = ref(false)
  const aiAssistantResult = ref<any>(null)
  const aiAssistantModules = ref<any[]>([])
  const aiAssistantModulesSaving = ref(false)

  const aiAssistantFeedback = reactive({
    type: '',
    message: '',
    detail: '',
    provider: '',
    providerDetail: '',
  })

  const aiAssistantForm = reactive({
    fluent_type: 'fluentbit',
    goal: 'both',
    module_type: 'input',
    sample: '',
    extra_requirements: '',
  })

  function clearAIAssistantFeedback() {
    aiAssistantFeedback.type = ''
    aiAssistantFeedback.message = ''
    aiAssistantFeedback.detail = ''
    aiAssistantFeedback.provider = ''
    aiAssistantFeedback.providerDetail = ''
  }

  function setAIAssistantFeedback(type: string, message: string, detail = '', provider = '', providerDetail = '') {
    aiAssistantFeedback.type = type
    aiAssistantFeedback.message = message
    aiAssistantFeedback.detail = detail
    aiAssistantFeedback.provider = provider
    aiAssistantFeedback.providerDetail = providerDetail
  }

  async function runAIAssistant() {
    if (!aiAssistantForm.sample.trim()) {
      aiAssistantResult.value = null
      setAIAssistantFeedback('danger', t('configs_page.require_sample_log'))
      return
    }
    aiAssistantLoading.value = true
    aiAssistantResult.value = null
    aiAssistantModules.value = []
    clearAIAssistantFeedback()
    try {
      const { data } = await analyzeLogSampleAssistant({
        fluent_type: aiAssistantForm.fluent_type,
        goal: aiAssistantForm.goal,
        module_type: aiAssistantForm.module_type,
        sample: aiAssistantForm.sample,
        extra_requirements: aiAssistantForm.extra_requirements,
      })
      aiAssistantResult.value = data
      const existingList = modules.value.filter(
        (m: any) => m.fluent_type === aiAssistantForm.fluent_type || m.fluent_type === 'shared'
      )
      aiAssistantModules.value = mergeAIModules(data.modules || [], existingList)
      setAIAssistantFeedback(
        'success', t('configs_page.ai_assistant_success'),
        t('configs_page.ai_assistant_ready'),
        [data.provider, data.account_name].filter(Boolean).join(' / ')
      )
    } catch (error: any) {
      aiAssistantResult.value = null
      aiAssistantModules.value = []
      setAIAssistantFeedback(
        'danger', t('configs_page.ai_assistant_failed'),
        getErrorMessage(error),
        error?.response?.data?.provider || '',
        getProviderErrorMessage(error)
      )
    } finally {
      aiAssistantLoading.value = false
    }
  }

  async function saveAIModules() {
    if (!aiAssistantModules.value.length) return
    aiAssistantModulesSaving.value = true
    const errors: string[] = []
    try {
      for (const item of aiAssistantModules.value) {
        if (item.decision === 'reuse_existing') continue
        try {
          if (item.decision === 'create_new') {
            const { data: created } = await createModule({
              name: item.name,
              module_type: item.module_type,
              fluent_type: aiAssistantForm.fluent_type,
              description: aiAssistantResult.value?.summary || '',
              variables: item.variables_json || '{}',
              content: item.content,
              is_builtin: false,
            })
            void created
          } else if (item.decision === 'update_existing') {
            if (!item.matchedModule) throw new Error(`No existing module matched. Change decision to "Create new" or re-run.`)
            await createModuleVersion(item.matchedModule.id, {
              content: item.content,
              variables: item.variables_json || '{}',
              comment: `AI-generated update from assistant`,
            })
          }
        } catch (e: any) {
          errors.push(`"${item.name}": ${e?.response?.data?.error || e?.message || ''}`)
        }
      }
      await loadModules()
      const rawGenerated = aiAssistantResult.value?.modules || []
      aiAssistantModules.value = mergeAIModules(rawGenerated, modules.value)
      if (errors.length) {
        setAIAssistantFeedback('warning', t('configs_page.ai_modules_save_partial'), errors.join('; '))
      } else {
        setAIAssistantFeedback('success', t('configs_page.ai_modules_saved'))
      }
    } finally {
      aiAssistantModulesSaving.value = false
    }
  }

  // Resolve AI pipeline module_names to catalog module objects.
  function resolveAIPipelineModules(pipeline: any) {
    const ft = aiAssistantForm.fluent_type
    const aiMetaByName = new Map(
      (aiAssistantResult.value?.modules || []).map((m: any) => [m.name, { type: m.module_type }])
    )
    const key3 = (name: string, type: string, ftype: string) => `${name}\0${type}\0${ftype}`
    const key2 = (name: string, type: string) => `${name}\0${type}`
    const byKey3 = new Map(modules.value.map((m: any) => [key3(m.name, m.module_type, m.fluent_type), m]))
    const byKey2 = new Map(modules.value.map((m: any) => [key2(m.name, m.module_type), m]))
    const byName = new Map(modules.value.map((m: any) => [m.name, m]))

    const resolved: any[] = []
    const unsaved: string[] = []
    for (const name of pipeline.module_names || []) {
      const aiMeta = aiMetaByName.get(name)
      const aiType = aiMeta?.type
      let mod = null
      if (aiType) {
        mod = byKey3.get(key3(name, aiType, ft))
             ?? byKey3.get(key3(name, aiType, 'shared'))
             ?? byKey2.get(key2(name, aiType))
      }
      if (!mod) mod = byName.get(name)
      if (!mod) unsaved.push(name)
      else resolved.push(mod)
    }
    return { resolved, unsaved }
  }

  // Split resolved modules into semantic slots.
  function splitAIPipelineModules(resolvedModules: any[]) {
    const inputMod = resolvedModules.find((m) => m.module_type === 'input') || null
    const outputMods = resolvedModules.filter((m) => m.module_type === 'output')
    const stageMods = resolvedModules.filter((m) => m.module_type !== 'input' && m.module_type !== 'output')
    const matchedTargets: any[] = []
    const unmatchedOutputMods: any[] = []
    for (const mod of outputMods) {
      const target = outputTargets.value.find((tgt: any) => tgt.name === mod.name)
      if (target) matchedTargets.push(target)
      else unmatchedOutputMods.push(mod)
    }
    return { inputMod, stageMods: [...stageMods, ...unmatchedOutputMods], matchedTargets, unmatchedOutputMods }
  }

  async function saveAIPipelineAsConfigPipeline(pipeline: any) {
    const { resolved, unsaved } = resolveAIPipelineModules(pipeline)
    if (unsaved.length) {
      alert(t('configs_page.ai_pipeline_modules_not_saved') + ' (' + unsaved.join(', ') + ')')
      return
    }
    const { inputMod, stageMods, matchedTargets, unmatchedOutputMods } = splitAIPipelineModules(resolved)
    try {
      const { data } = await createConfigPipeline({
        name: pipeline.name || `ai-pipeline-${Date.now()}`,
        description: pipeline.description || aiAssistantResult.value?.summary || '',
        fluent_type: aiAssistantForm.fluent_type,
        input_module_id: inputMod?.id ?? null,
        filter_module_ids: stageMods.map((m) => m.id),
        output_target_ids: matchedTargets.map((t: any) => t.id),
      })
      await loadPipelines()
      const msg = t('configs_page.ai_pipeline_saved_as_pipeline').replace('{name}', data.name)
      const detail = unmatchedOutputMods.length
        ? t('configs_page.ai_pipeline_output_stages_note').replace('{names}', unmatchedOutputMods.map((m) => m.name).join(', '))
        : ''
      setAIAssistantFeedback('success', msg, detail)
    } catch (e: any) {
      setAIAssistantFeedback('danger', t('common.request_failed'), e?.response?.data?.error || e?.message || '')
    }
  }

  return {
    aiAssistantLoading, aiAssistantResult, aiAssistantModules, aiAssistantModulesSaving,
    aiAssistantFeedback, aiAssistantForm,
    clearAIAssistantFeedback, setAIAssistantFeedback,
    runAIAssistant, saveAIModules,
    resolveAIPipelineModules, splitAIPipelineModules, saveAIPipelineAsConfigPipeline,
  }
}
