// Template management composable – state, computed, and CRUD for config templates.
// Deps injected: { modules, templates, t, router }
// templates ref is owned by ConfigsPage and passed in (shared with useWizard etc.)

import { computed, reactive } from 'vue'
import { getTemplates, createTemplate, deleteTemplate } from '../../../api'
import { getErrorMessage as getErrorMessageFn } from '../utils/config-format'
import {
  normalizeName, generateUniqueDraftName, countNonEmptyLines,
  extractTemplatePlaceholders, uniqueSorted, diffKeys,
} from '../utils/config-text'
import {
  createAIDraftState, resetAIDraftState, activateAIDraftState, areDraftConfirmationsComplete,
} from '../utils/config-ai-draft'
import { runtimeLabel } from '../utils/config-format'

export function useTemplates({ modules, templates, t, router }: any) {
  function getErrorMessage(e: any) { return getErrorMessageFn(e, t) }

  const templateForm = reactive({
    name: '',
    description: '',
    fluent_type: 'fluentbit',
    content: '',
  })

  const aiTemplateDraftState = reactive(createAIDraftState())

  const templateExamples: Record<string, string> = {
    fluentbit: `[SERVICE]
    Flush        1
    Daemon       Off
    Log_Level    info

[INPUT]
    Name              tail
    Path              /var/log/nginx/*.log
    Tag               nginx.access
    DB                /var/lib/fluent-bit/nginx.db
    Refresh_Interval  5

[FILTER]
    Name    modify
    Match   nginx.*
    Add     environment prod

[OUTPUT]
    Name   forward
    Match  nginx.*
    Host   10.0.0.15
    Port   24224`,
    fluentd: `<system>
  workers 2
  log_level info
</system>

<source>
  @type tail
  path /var/log/nginx/access.log
  pos_file /var/log/fluentd/nginx.pos
  tag nginx.access
  <parse>
    @type json
  </parse>
</source>

<filter nginx.**>
  @type record_transformer
  <record>
    environment prod
  </record>
</filter>

<match nginx.**>
  @type forward
  <server>
    host 10.0.0.15
    port 24224
  </server>
</match>`,
  }

  // --- Computed ---

  const assemblyTemplates = computed(() => templates.value.filter((item: any) => item.source_type === 'module_assembly'))
  const assemblyTemplateCount = computed(() => assemblyTemplates.value.length)
  const manualTemplateCount = computed(() => templates.value.filter((item: any) => item.source_type !== 'module_assembly').length)
  const wizardBuiltTemplates = computed(() =>
    assemblyTemplates.value.filter((tpl: any) => {
      if (!tpl.flow_layout) return false
      try {
        const layout = typeof tpl.flow_layout === 'string' ? JSON.parse(tpl.flow_layout) : tpl.flow_layout
        return layout.builder === 'wizard'
      } catch { return false }
    })
  )
  const currentTemplateExample = computed(() => templateExamples[templateForm.fluent_type] || templateExamples.fluentbit)
  const aiTemplateDraftSource = computed(() => [aiTemplateDraftState.provider, aiTemplateDraftState.accountName].filter(Boolean).join(' / '))
  const aiTemplateDraftReady = computed(() => areDraftConfirmationsComplete(aiTemplateDraftState))
  const aiTemplateDraftComparison = computed(() => buildTemplateDraftComparison())
  const aiTemplateDraftCanSave = computed(() => {
    if (!aiTemplateDraftReady.value) return false
    if (!aiTemplateDraftState.active) return true
    return !aiTemplateDraftComparison.value?.hasConflict
  })

  // --- Functions ---

  function summarizeChangedKeys(added: string[], removed: string[]): string {
    const parts: string[] = []
    if (added.length) parts.push(t('configs_page.ai_draft_diff_added').replace('{items}', added.join(', ')))
    if (removed.length) parts.push(t('configs_page.ai_draft_diff_removed').replace('{items}', removed.join(', ')))
    return parts.join('；')
  }

  function buildTemplateDraftComparison() {
    if (!aiTemplateDraftState.active) return null
    const placeholderKeys = uniqueSorted(extractTemplatePlaceholders(templateForm.content).map((item: string) => item.replace(/[{}\s.]/g, '')))
    const existing = templates.value.find((item: any) => normalizeName(item.name) === normalizeName(templateForm.name))
    let identityMessage = t('configs_page.ai_draft_diff_new_asset')
    let existingName = '', changeMessage = '', changeDetail = '', suggestedName = ''
    const hasConflict = !!existing
    if (existing) {
      existingName = `${existing.name} / ${runtimeLabel(existing.fluent_type)}`
      identityMessage = t('configs_page.ai_draft_diff_existing_template')
      suggestedName = generateUniqueDraftName(
        templateForm.name, templates.value.map((item: any) => item.name),
        `ai-${templateForm.fluent_type || 'template'}-template`
      )
      changeMessage = existing.content === templateForm.content
        ? t('configs_page.ai_draft_diff_content_same')
        : t('configs_page.ai_draft_diff_content_changed')
      const prevKeys = uniqueSorted(extractTemplatePlaceholders(existing.content).map((item: string) => item.replace(/[{}\s.]/g, '')))
      const { added, removed } = diffKeys(uniqueSorted(placeholderKeys), prevKeys)
      changeDetail = summarizeChangedKeys(added, removed)
      if (!changeDetail) {
        const lineDelta = countNonEmptyLines(templateForm.content) - countNonEmptyLines(existing.content)
        changeDetail = lineDelta !== 0
          ? t('configs_page.ai_draft_diff_line_delta').replace('{count}', String(lineDelta))
          : t('configs_page.ai_draft_diff_existing_review')
      }
    }
    return {
      existingAsset: existing || null, hasConflict, identityMessage, existingName, suggestedName,
      lineCount: countNonEmptyLines(templateForm.content), placeholderCount: placeholderKeys.length,
      changeMessage, changeDetail,
    }
  }

  async function loadTemplates() {
    const { data } = await getTemplates()
    templates.value = data.data || []
  }

  function resetTemplateForm() {
    templateForm.name = ''
    templateForm.description = ''
    templateForm.fluent_type = 'fluentbit'
    templateForm.content = ''
    resetAIDraftState(aiTemplateDraftState)
  }

  async function saveTemplate(): Promise<boolean> {
    try {
      await createTemplate(templateForm)
      resetAIDraftState(aiTemplateDraftState)
      await loadTemplates()
      return true
    } catch (error) {
      alert(`${t('configs_page.create_template_failed')}: ${getErrorMessage(error)}`)
      return false
    }
  }

  function applySuggestedTemplateName() {
    const name = aiTemplateDraftComparison.value?.suggestedName
    if (name) templateForm.name = name
  }

  async function openExistingTemplateFromDraft(hideModal?: () => void) {
    const existing = aiTemplateDraftComparison.value?.existingAsset
    if (!existing) return
    hideModal?.()
    await router.push(`/configs/${existing.id}`)
  }

  async function handleDeleteTemplate(template: any) {
    if (!confirm(t('configs_page.confirm_delete_template').replace('{name}', template.name))) return
    try {
      await deleteTemplate(template.id)
      await loadTemplates()
    } catch (error) {
      alert(`${t('configs_page.delete_template_failed')}: ${getErrorMessage(error)}`)
    }
  }

  function useAITemplateDraft(pipeline: any, aiAssistantResult: any, aiAssistantForm: any) {
    if (!aiAssistantResult) return
    resetTemplateForm()
    if (pipeline) {
      templateForm.name = pipeline.name || `ai-${aiAssistantForm.fluent_type}-template`
      templateForm.description = pipeline.description || aiAssistantResult.summary || ''
      templateForm.content = pipeline.template_content || ''
    } else {
      templateForm.name = `ai-${aiAssistantForm.fluent_type}-template`
      templateForm.description = aiAssistantResult.summary || ''
      templateForm.content = (aiAssistantResult.pipelines && aiAssistantResult.pipelines[0]?.template_content) || ''
    }
    templateForm.fluent_type = aiAssistantForm.fluent_type
    activateAIDraftState(aiTemplateDraftState, aiAssistantResult, [
      t('configs_page.ai_draft_review_name'), t('configs_page.ai_draft_review_runtime'),
      t('configs_page.ai_draft_review_template_content'), t('configs_page.ai_draft_review_notes'),
    ], [
      t('configs_page.ai_draft_confirm_name'), t('configs_page.ai_draft_confirm_target'),
      t('configs_page.ai_draft_confirm_template_content'), t('configs_page.ai_draft_confirm_notes'),
    ])
  }

  return {
    templateForm, aiTemplateDraftState,
    assemblyTemplates, assemblyTemplateCount, manualTemplateCount, wizardBuiltTemplates,
    currentTemplateExample, aiTemplateDraftSource, aiTemplateDraftReady,
    aiTemplateDraftComparison, aiTemplateDraftCanSave,
    loadTemplates, resetTemplateForm, saveTemplate, applySuggestedTemplateName,
    openExistingTemplateFromDraft, handleDeleteTemplate, buildTemplateDraftComparison, useAITemplateDraft,
  }
}
