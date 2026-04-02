<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('configs_page.title') }}</h4>
        <div class="text-muted">{{ t('configs_page.subtitle') }}</div>
      </div>
      <div class="d-flex gap-2">
        <button v-if="activeTab === 'import'" class="btn btn-success" :disabled="importAnalysisLoading || !importForm.content.trim()" @click="runImportAnalysis">
          <i class="bi bi-file-earmark-arrow-up me-1"></i>{{ importAnalysisLoading ? t('configs_page.import_analyzing') : t('configs_page.import_analyze') }}
        </button>
        <button v-if="activeTab === 'wizard'" class="btn btn-success" @click="runWizardPreview">
          <i class="bi bi-magic me-1"></i>{{ t('configs_page.generate_wizard_preview') }}
        </button>
        <button v-if="activeTab === 'assistant'" class="btn btn-success" :disabled="aiAssistantLoading || !aiAssistantForm.sample.trim()" @click="runAIAssistant">
          <i class="bi bi-stars me-1"></i>{{ aiAssistantLoading ? t('configs_page.ai_assistant_running') : t('configs_page.ai_assistant_run') }}
        </button>
        <button v-if="activeTab === 'templates'" class="btn btn-success" @click="openAssemblyTemplateBuilder">
          <i class="bi bi-diagram-3 me-1"></i>{{ t('configs_page.create_assembly_template') }}
        </button>
        <button v-if="activeTab === 'templates'" class="btn btn-outline-secondary" @click="openCreateTemplate">
          <i class="bi bi-code-square me-1"></i>{{ t('configs_page.create_manual_template') }}
        </button>
        <button v-if="activeTab === 'pipelines'" class="btn btn-primary" @click="openCreatePipeline">
          <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.create_pipeline') }}
        </button>
        <button v-if="activeTab === 'modules'" class="btn btn-outline-danger" :disabled="!selectedDeletableModuleCount" @click="handleBatchDeleteModules">
          <i class="bi bi-trash me-1"></i>{{ t('configs_page.batch_delete_modules').replace('{count}', String(selectedDeletableModuleCount)) }}
        </button>
        <button v-if="activeTab === 'modules'" class="btn btn-primary" @click="openCreateModule">
          <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.create_module') }}
        </button>
        <button v-if="activeTab === 'preview'" class="btn btn-success" @click="runPreview">
          <i class="bi bi-play-circle me-1"></i>{{ t('configs_page.generate_preview') }}
        </button>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body p-2">
        <div class="fm-config-explainer px-3 py-2 mb-2">
          <div class="fw-semibold">{{ t('configs_page.explainer_title') }}</div>
          <div class="small mb-0">{{ t('configs_page.explainer_body') }}</div>
        </div>
        <div class="nav nav-pills fm-config-tabs">
          <button class="nav-link" :class="{ active: activeTab === 'templates' }" @click="activeTab = 'templates'">
            {{ t('configs_page.templates') }}<span class="badge rounded-pill text-bg-light ms-2">{{ templates.length }}</span>
          </button>
          <button class="nav-link" :class="{ active: activeTab === 'import' }" @click="activeTab = 'import'">{{ t('configs_page.import_existing') }}</button>
          <button class="nav-link" :class="{ active: activeTab === 'wizard' }" @click="activeTab = 'wizard'">{{ t('configs_page.wizard') }}</button>
          <button class="nav-link" :class="{ active: activeTab === 'assistant' }" @click="activeTab = 'assistant'">{{ t('configs_page.ai_assistant') }}</button>
          <button class="nav-link" :class="{ active: activeTab === 'pipelines' }" @click="activeTab = 'pipelines'">
            {{ t('configs_page.pipelines') }}<span class="badge rounded-pill text-bg-light ms-2">{{ pipelines.length }}</span>
          </button>
          <button class="nav-link" :class="{ active: activeTab === 'modules' }" @click="activeTab = 'modules'">
            {{ t('configs_page.modules') }}<span class="badge rounded-pill text-bg-light ms-2">{{ moduleCatalogCount }}</span>
          </button>
          <button class="nav-link" :class="{ active: activeTab === 'preview' }" @click="activeTab = 'preview'">{{ t('configs_page.preview') }}</button>
        </div>
      </div>
    </div>

    <ConfigsTemplatesTab
      v-if="activeTab === 'templates'"
      :state="{ templates, assemblyTemplateCount, manualTemplateCount }"
      :actions="{ handleDeleteTemplate, openTemplateInWizard }"
      :helpers="{ runtimeLabel, templateSourceLabel, templateAssemblyModules, formatTime }"
    />
    <ConfigsImportTab
      v-else-if="activeTab === 'import'"
      :state="{
        importForm, importAnalysisLoading, importModulesLoading, importedConfigResult,
        importedWorkspaceModules, importedWorkspaceTemplate, importFlowPathLabel,
        importReuseDecisionCount, importCreateDecisionCount, importReusableMatchCount,
        importDestinationMatchCount, importSemanticChangeCount, importBlockingIssueCount,
      }"
      :actions="{ runImportAnalysis, importParsedModules, setAllImportedModuleActions, setImportedModuleAction }"
      :helpers="{ runtimeLabel, importValidationBadgeClass, importValidationLabel, importActionBadgeClass, importActionLabel, importDestinationMatchLabel, importedModuleNameIssue }"
    />
    <ConfigsWizardTab
      v-else-if="activeTab === 'wizard'"
      :state="{
        wizardForm, wizardServiceModuleId, wizardParserModuleIds,
        wizardServiceSearch, wizardParserSearch, wizardInputSearch, wizardFilterSearch, wizardRouteSearch, wizardOutputSearch,
        wizardStagePages, wizardPagedServiceModules, wizardPagedParserModules, wizardPagedInputModules,
        wizardPagedFilterModules, wizardPagedRouteModules, wizardPagedOutputTargets,
        wizardServiceModule, wizardSelectedParserModules, wizardGlobalVariableGroups,
        wizardPipelines, activeWizardPipelineId, activeWizardPipeline, wizardPipelineCards,
        wizardPipelineVariableGroups, wizardOutputResolutionWarnings, wizardIncompletePipelineLabels,
        wizardRenderSummary, renderedConfig, wizardAssemblyTemplates: wizardBuiltTemplates,
        wizardLoadedFromTemplate, wizardSaveButtonLabel, wizardCompatiblePipelines,
      }"
      :actions="{
        selectWizardServiceModule, toggleWizardParserModule,
        addWizardPipeline, duplicateWizardPipeline, removeWizardPipeline, selectWizardPipeline,
        setWizardPipelineInput, addWizardFilter, removeWizardFilter, moveWizardFilter,
        addWizardRoute, removeWizardRoute, moveWizardRoute,
        addWizardOutputTarget, removeWizardOutput, moveWizardOutput,
        changeWizardStagePage, runWizardPreview, saveWizardAsTemplate,
        openAdvancedPreviewFromWizard, loadWizardFromTemplate, clearWizardLoadedTemplate, addWizardPipelineFromSaved,
      }"
      :helpers="{ runtimeLabel, wizardGoalLabel, wizardPipelineDisplayName, matchingOutputModuleForTarget }"
    />
    <ConfigsAssistantTab
      v-else-if="activeTab === 'assistant'"
      :state="{ aiAssistantForm, aiAssistantLoading, aiAssistantResult, aiAssistantFeedback, moduleTypes, aiAssistantModules, aiAssistantModulesSaving }"
      :actions="{ runAIAssistant, useAIModuleDraft, useAITemplateDraft, saveAIModules, saveAIPipelineAsConfigPipeline, sendAIPipelineToWizard }"
      :helpers="{ runtimeLabel }"
    />
    <ConfigsPipelinesTab
      v-else-if="activeTab === 'pipelines'"
      :state="{ pipelines }"
      :actions="{ openEditPipeline, handleDeletePipeline, openCreatePipeline }"
      :helpers="{ runtimeLabel, runtimeBadgeClass, formatTime }"
    />
    <ConfigsModulesTab
      v-else-if="activeTab === 'modules'"
      :state="{
        moduleCatalogCount, sharedModuleCount, usedModuleTypes, moduleTypeStats, managedModuleTypes,
        moduleQuery, moduleTableRangeStart, moduleTableRangeEnd, moduleTableTotal,
        selectedDeletableModuleCount, moduleTableTotalPages, moduleTableLoading,
        visibleDeletableModules, allVisibleDeletableModulesSelected, visibleModules, selectedModuleIds,
      }"
      :actions="{ applyModuleQuery, resetModuleQuery, changeModulePage, setModuleTypeFilter, toggleSelectAllVisibleModules, toggleModuleSelection, openEditModule, openModuleVersions, handleDeleteModule }"
      :helpers="{ runtimeLabel, shortVariables, formatTime }"
    />
    <ConfigsPreviewTab
      v-else
      :state="{
        previewForm, previewAvailableOutputTargets, selectedPreviewOutputTargetIds,
        previewUnresolvedOutputTargets, previewModuleSearch, previewVisibleModules,
        selectedPreviewModuleIds, renderedConfig, previewFlowPathLabel, previewDestinationChips,
        previewSummaryModules, previewResolvedOutputTargets,
        analysisResult, compatibilityResult, replayResult, diffResult,
      }"
      :actions="{ togglePreviewOutputTarget, runPreview, runLint, runCompatibility, runReplay, runSemanticDiff, togglePreviewModule }"
      :helpers="{ runtimeLabel, findingBadgeClass, formatJson }"
    />

    <TemplateFormModal
      ref="templateModalRef"
      :form="templateForm"
      :ai-draft-state="aiTemplateDraftState"
      :ai-draft-source="aiTemplateDraftSource"
      :ai-draft-comparison="aiTemplateDraftComparison"
      :ai-draft-ready="aiTemplateDraftReady"
      :ai-draft-can-save="aiTemplateDraftCanSave"
      :current-example="currentTemplateExample"
      @save="doSaveTemplate"
      @apply-suggested-name="applySuggestedTemplateName"
      @open-existing="doOpenExistingTemplateFromDraft"
    />

    <ModuleFormModal
      ref="moduleModalRef"
      :form="moduleForm"
      :editing-id="editingModuleId"
      :ai-draft-state="aiModuleDraftState"
      :ai-draft-source="aiModuleDraftSource"
      :ai-draft-comparison="aiModuleDraftComparison"
      :ai-draft-ready="aiModuleDraftReady"
      :ai-draft-can-save="aiModuleDraftCanSave"
      :module-types="moduleTypes"
      :module-examples="moduleExamples"
      :current-example="currentModuleExample"
      :variables-mode="moduleVariablesMode"
      :variable-rows="moduleVariableRows"
      :variables-error="moduleVariablesFormError"
      @save="doSaveModule"
      @apply-suggested-name="applySuggestedModuleName"
      @open-existing="doOpenExistingModuleFromDraft"
      @set-variables-mode="setModuleVariablesMode"
      @add-variable-row="addModuleVariableRow"
      @remove-variable-row="removeModuleVariableRow"
      @sync-variables="syncModuleVariablesFromRows"
    />

    <PipelineFormModal
      ref="pipelineModalRef"
      :form="pipelineForm"
      :editing-id="editingPipelineId"
      :input-modules="pipelineInputModules"
      :filter-modules="pipelineFilterModules"
      :available-output-targets="pipelineAvailableOutputTargets"
      @save="doSavePipeline"
      @add-filter="addPipelineFilterModule"
      @remove-filter="removePipelineFilterModule"
      @move-filter="movePipelineFilterModule"
      @toggle-output="togglePipelineOutputTarget"
    />

    <ModuleVersionsModal
      ref="moduleVersionsModalRef"
      :current-module="currentModule"
      :versions="moduleVersions"
      :version-form="moduleVersionForm"
      :format-time="formatTime"
      @save-version="doSaveModuleVersion"
    />
  </div>
</template>

<script setup>
import './configs.css'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ConfigsAssistantTab from './tabs/ConfigsAssistantTab.vue'
import ConfigsImportTab from './tabs/ConfigsImportTab.vue'
import ConfigsModulesTab from './tabs/ConfigsModulesTab.vue'
import ConfigsPipelinesTab from './tabs/ConfigsPipelinesTab.vue'
import ConfigsPreviewTab from './tabs/ConfigsPreviewTab.vue'
import ConfigsTemplatesTab from './tabs/ConfigsTemplatesTab.vue'
import ConfigsWizardTab from './tabs/ConfigsWizardTab.vue'
import TemplateFormModal from './modals/TemplateFormModal.vue'
import ModuleFormModal from './modals/ModuleFormModal.vue'
import PipelineFormModal from './modals/PipelineFormModal.vue'
import ModuleVersionsModal from './modals/ModuleVersionsModal.vue'
import { useI18n } from '../../i18n'
import {
  runtimeLabel, runtimeBadgeClass, shortVariables, formatJson, findingBadgeClass,
  getProviderErrorMessage, templateAssemblyModules, templateSourceLabel as templateSourceLabelFn,
  wizardGoalLabel as wizardGoalLabelFn,
} from './utils/config-format'
import {
  importActionBadgeClass, importValidationBadgeClass, setImportedModuleAction,
  importActionLabel as importActionLabelFn, importValidationLabel as importValidationLabelFn,
  importDestinationMatchLabel as importDestinationMatchLabelFn,
} from './utils/config-import-utils'
import { matchingOutputModuleForTarget, createWizardPipeline, createWizardInstanceID } from './utils/config-wizard-helpers'
import { usePreviewAnalysis } from './composables/usePreviewAnalysis'
import { useWizard } from './composables/useWizard'
import { useTemplates } from './composables/useTemplates'
import { usePipelines } from './composables/usePipelines'
import { useModules } from './composables/useModules'
import { useAIAssistant } from './composables/useAIAssistant'
import { useImport } from './composables/useImport'
import { getOutputTargets } from '../../api'

const activeTab = ref('templates')
const templates = ref([])
const modules = ref([])
const pipelines = ref([])
const outputTargets = ref([])

const route = useRoute()
const router = useRouter()
const { t, dateLocale } = useI18n()

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}
function templateSourceLabel(sourceType) { return templateSourceLabelFn(sourceType, t) }
function wizardGoalLabel(goal) { return wizardGoalLabelFn(goal, t) }
function importActionLabel(action) { return importActionLabelFn(action, t) }
function importValidationLabel(verdict) { return importValidationLabelFn(verdict, t) }
function importDestinationMatchLabel(matchType) { return importDestinationMatchLabelFn(matchType, t) }

// --- Composables ---

const {
  renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult,
  previewForm, selectedPreviewModuleIds, selectedPreviewOutputTargetIds,
  previewModuleSearch, previewModuleVariables,
  previewEligibleModules, previewVisibleModules, previewAvailableOutputTargets,
  previewSelectedModules, previewResolvedOutputTargets, previewUnresolvedOutputTargets,
  previewSelectedOutputModules, previewSummaryModules, previewFlowSummary,
  togglePreviewModule, togglePreviewOutputTarget, buildPreviewSelectedModuleRefs,
  runPreview, runLint, runCompatibility, runReplay, runSemanticDiff,
} = usePreviewAnalysis({ modules, outputTargets, activeTab, t })

const {
  templateForm, aiTemplateDraftState,
  assemblyTemplates, assemblyTemplateCount, manualTemplateCount, wizardBuiltTemplates,
  currentTemplateExample, aiTemplateDraftSource, aiTemplateDraftReady,
  aiTemplateDraftComparison, aiTemplateDraftCanSave,
  loadTemplates, resetTemplateForm, saveTemplate, applySuggestedTemplateName,
  openExistingTemplateFromDraft, handleDeleteTemplate, useAITemplateDraft: applyAITemplateDraftFn,
} = useTemplates({ modules, templates, t, router })

const {
  editingPipelineId, pipelineForm, pipelineFilterPickerValue,
  pipelineEligibleModules, pipelineInputModules, pipelineFilterModules, pipelineAvailableOutputTargets,
  loadPipelines, preparePipelineCreate, preparePipelineEdit,
  savePipeline, handleDeletePipeline,
  addPipelineFilterModule, removePipelineFilterModule, movePipelineFilterModule, togglePipelineOutputTarget,
} = usePipelines({ modules, outputTargets, pipelines, t })

const {
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
  prepareModuleEdit, applySuggestedModuleName,
  saveModule, handleDeleteModule, handleBatchDeleteModules,
  openModuleVersions, saveModuleVersion, useAIModuleDraft: useAIModuleDraftFn,
} = useModules({
  modules, t,
  onModuleDeleted: (moduleId) => {
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => id !== moduleId)
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== moduleId)
    removeWizardModuleReferences(moduleId)
  },
  onBatchModulesDeleted: (deletedSet) => {
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => !deletedSet.has(id))
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => !deletedSet.has(id))
    removeWizardModuleReferencesBatch(Array.from(deletedSet))
  },
})

const {
  aiAssistantLoading, aiAssistantResult, aiAssistantModules, aiAssistantModulesSaving,
  aiAssistantFeedback, aiAssistantForm, setAIAssistantFeedback,
  runAIAssistant, saveAIModules,
  resolveAIPipelineModules, splitAIPipelineModules, saveAIPipelineAsConfigPipeline,
} = useAIAssistant({ modules, outputTargets, t, loadModules, loadModuleTable, loadPipelines })

const {
  importForm, importAnalysisLoading, importModulesLoading,
  importedConfigResult, importedWorkspaceModules, importedWorkspaceTemplate,
  importBlockingIssueCount, importFlowPathLabel, importSemanticChangeCount,
  importReusableMatchCount, importDestinationMatchCount,
  importReuseDecisionCount, importCreateDecisionCount,
  importedModuleNameIssue, setAllImportedModuleActions,
  runImportAnalysis, importParsedModules,
} = useImport({
  modules, outputTargets, templates, t,
  previewForm, renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult,
  selectedPreviewModuleIds, selectedPreviewOutputTargetIds, activeTab,
  loadModules, loadModuleTable,
  loadOutputTargets: async () => {
    try { outputTargets.value = await getOutputTargets() } catch { outputTargets.value = [] }
  },
  loadTemplates, router,
})

const {
  wizardForm, wizardServiceModuleId, wizardParserModuleIds, wizardGlobalModuleVariables,
  wizardPipelines, activeWizardPipelineId, wizardLoadedFromTemplate,
  wizardServiceSearch, wizardParserSearch, wizardInputSearch, wizardFilterSearch, wizardRouteSearch, wizardOutputSearch, wizardModuleSearch,
  wizardStagePages, selectedWizardModuleIds, selectedWizardOutputTargetIds, wizardModuleVariableValues,
  wizardPipelineModuleTypes, wizardPipelineStageTotal,
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
  matchingWizardParserModules, autoAttachWizardParsersForInputModule,
  wizardPipelineInputTag, wizardPipelineModuleDefaults, buildWizardPipelineDraft,
  mergeWizardModuleVariableValues, ensureWizardBaselineModules,
  changeWizardStagePage, resetWizardStagePage,
  selectWizardServiceModule, toggleWizardParserModule,
  addWizardPipeline, duplicateWizardPipeline, removeWizardPipeline, selectWizardPipeline, updateWizardPipeline,
  setWizardPipelineInput, addWizardFilter, removeWizardFilter, moveWizardFilter,
  addWizardRoute, removeWizardRoute, moveWizardRoute,
  buildWizardOutputDraft, addWizardOutputTarget, removeWizardOutput, moveWizardOutput,
  pruneWizardStateForRuntime, removeWizardModuleReferences, removeWizardModuleReferencesBatch,
  applyWizardInputPreset, toggleWizardOutputTarget,
  buildWizardRenderModuleRefs, preparePreviewMetaFromWizard,
  runWizardPreview, saveWizardAsTemplate, openAdvancedPreviewFromWizard,
  wizardHasContent, loadWizardFromTemplate, clearWizardLoadedTemplate,
  addWizardPipelineFromSaved, resetWizardForm,
  wizardPipelineDisplayName,
} = useWizard({ modules, outputTargets, pipelines, t, previewForm, renderedConfig, analysisResult, compatibilityResult, replayResult, diffResult, activeTab, loadTemplates, router })

// --- Computed ---

const wizardFlowPathLabel = computed(() =>
  wizardFlowSummary.value.path.length ? wizardFlowSummary.value.path.join(' -> ') : t('configs_page.no_solution_path')
)
const previewFlowPathLabel = computed(() =>
  previewFlowSummary.value.path.length ? previewFlowSummary.value.path.join(' -> ') : t('configs_page.no_solution_path')
)
const previewDestinationChips = computed(() => previewFlowSummary.value.destinationChips || [])

// --- Modal refs ---

const templateModalRef = ref(null)
const moduleModalRef = ref(null)
const pipelineModalRef = ref(null)
const moduleVersionsModalRef = ref(null)

// --- Template modal actions ---

function openCreateTemplate() {
  resetTemplateForm()
  templateModalRef.value?.show()
}

function openAssemblyTemplateBuilder() {
  activeTab.value = 'wizard'
}

function openTemplateInWizard(template, options = {}) {
  if (!template) return
  activeTab.value = 'wizard'
  loadWizardFromTemplate(template)
  if (!options.suppressRouteSync) {
    router.replace({ path: '/configs', query: { tab: 'wizard' } })
  }
}

async function doSaveTemplate() {
  const ok = await saveTemplate()
  if (ok) templateModalRef.value?.hide()
}

async function doOpenExistingTemplateFromDraft() {
  await openExistingTemplateFromDraft(() => templateModalRef.value?.hide())
}

// --- Module modal actions ---

function openCreateModule() {
  resetModuleForm()
  moduleModalRef.value?.show()
}

function openEditModule(module) {
  prepareModuleEdit(module)
  moduleModalRef.value?.show()
}

function doOpenExistingModuleFromDraft() {
  const existing = aiModuleDraftComparison.value?.existingAsset
  if (!existing) return
  openEditModule(existing)
}

async function doSaveModule() {
  const ok = await saveModule()
  if (ok) moduleModalRef.value?.hide()
}

async function openModuleVersionsModal(module) {
  const ok = await openModuleVersions(module)
  if (ok) moduleVersionsModalRef.value?.show()
}

async function doSaveModuleVersion() {
  await saveModuleVersion()
}

// --- Pipeline modal actions ---

function openCreatePipeline() {
  preparePipelineCreate()
  pipelineModalRef.value?.show()
}

function openEditPipeline(pipeline) {
  preparePipelineEdit(pipeline)
  pipelineModalRef.value?.show()
}

async function doSavePipeline() {
  const ok = await savePipeline()
  if (ok) pipelineModalRef.value?.hide()
}

// --- AI cross-domain coordination ---

function useAIModuleDraft(module) {
  useAIModuleDraftFn(module, aiAssistantResult.value, aiAssistantForm)
  moduleModalRef.value?.show()
}

function useAITemplateDraft(pipeline) {
  applyAITemplateDraftFn(pipeline, aiAssistantResult.value, aiAssistantForm)
  templateModalRef.value?.show()
}

function sendAIPipelineToWizard(pipeline) {
  const { resolved, unsaved } = resolveAIPipelineModules(pipeline)
  if (unsaved.length) {
    alert(t('configs_page.ai_pipeline_modules_not_saved') + ' (' + unsaved.join(', ') + ')')
    return
  }
  const { inputMod, stageMods, matchedTargets, unmatchedOutputMods } = splitAIPipelineModules(resolved)
  const newPipeline = createWizardPipeline()
  newPipeline.name = pipeline.name || ''
  newPipeline.input = inputMod ? { id: createWizardInstanceID('wizard-input'), module_id: inputMod.id } : null
  newPipeline.filters = stageMods.map((m) => ({ id: createWizardInstanceID('wizard-filter'), module_id: m.id }))
  newPipeline.outputs = matchedTargets.map((tgt) => ({ id: createWizardInstanceID('wizard-output'), target_id: tgt.id }))
  wizardPipelines.value.push(newPipeline)
  activeWizardPipelineId.value = newPipeline.id
  activeTab.value = 'wizard'
  if (unmatchedOutputMods.length) {
    alert(t('configs_page.ai_pipeline_output_stages_note').replace('{names}', unmatchedOutputMods.map((m) => m.name).join(', ')))
  }
}

// --- Lifecycle ---

async function loadOutputTargets() {
  try { outputTargets.value = await getOutputTargets() } catch { outputTargets.value = [] }
}

onMounted(async () => {
  resetWizardForm()
  await Promise.all([loadTemplates(), loadModules(), loadModuleTable(), loadOutputTargets(), loadPipelines()])
  ensureWizardBaselineModules()

  if (route.query.tab === 'wizard') activeTab.value = 'wizard'

  const loadTemplateID = Number(route.query.load_template || 0)
  if (loadTemplateID) {
    const template = templates.value.find((item) => item.id === loadTemplateID)
    if (template) {
      openTemplateInWizard(template, { suppressRouteSync: true })
      await router.replace({ path: '/configs', query: { tab: 'wizard' } })
    }
  }
})
</script>
