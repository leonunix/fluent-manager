<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('configs_page.title') }}</h4>
        <div class="text-muted">{{ t('configs_page.subtitle') }}</div>
      </div>
      <div class="d-flex gap-2">
        <button
          v-if="activeTab === 'import'"
          class="btn btn-success"
          :disabled="importAnalysisLoading || !importForm.content.trim()"
          @click="runImportAnalysis"
        >
          <i class="bi bi-file-earmark-arrow-up me-1"></i>{{ importAnalysisLoading ? t('configs_page.import_analyzing') : t('configs_page.import_analyze') }}
        </button>
        <button
          v-if="activeTab === 'wizard'"
          class="btn btn-success"
          @click="runWizardPreview"
        >
          <i class="bi bi-magic me-1"></i>{{ t('configs_page.generate_wizard_preview') }}
        </button>
        <button
          v-if="activeTab === 'assistant'"
          class="btn btn-success"
          :disabled="aiAssistantLoading || !aiAssistantForm.sample.trim()"
          @click="runAIAssistant"
        >
          <i class="bi bi-stars me-1"></i>{{ aiAssistantLoading ? t('configs_page.ai_assistant_running') : t('configs_page.ai_assistant_run') }}
        </button>
        <button
          v-if="activeTab === 'templates'"
          class="btn btn-success"
          @click="openAssemblyTemplateBuilder"
        >
          <i class="bi bi-diagram-3 me-1"></i>{{ t('configs_page.create_assembly_template') }}
        </button>
        <button
          v-if="activeTab === 'templates'"
          class="btn btn-outline-secondary"
          @click="openCreateTemplate"
        >
          <i class="bi bi-code-square me-1"></i>{{ t('configs_page.create_manual_template') }}
        </button>
        <button
          v-if="activeTab === 'pipelines'"
          class="btn btn-primary"
          @click="openCreatePipeline"
        >
          <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.create_pipeline') }}
        </button>
        <button
          v-if="activeTab === 'modules'"
          class="btn btn-outline-danger"
          :disabled="!selectedDeletableModuleCount"
          @click="handleBatchDeleteModules"
        >
          <i class="bi bi-trash me-1"></i>{{ t('configs_page.batch_delete_modules').replace('{count}', String(selectedDeletableModuleCount)) }}
        </button>
        <button
          v-if="activeTab === 'modules'"
          class="btn btn-primary"
          @click="openCreateModule"
        >
          <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.create_module') }}
        </button>
        <button
          v-if="activeTab === 'preview'"
          class="btn btn-success"
          @click="runPreview"
        >
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
          <button
            class="nav-link"
            :class="{ active: activeTab === 'templates' }"
            @click="activeTab = 'templates'"
          >
            {{ t('configs_page.templates') }}
            <span class="badge rounded-pill text-bg-light ms-2">{{ templates.length }}</span>
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'import' }"
            @click="activeTab = 'import'"
          >
            {{ t('configs_page.import_existing') }}
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'wizard' }"
            @click="activeTab = 'wizard'"
          >
            {{ t('configs_page.wizard') }}
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'assistant' }"
            @click="activeTab = 'assistant'"
          >
            {{ t('configs_page.ai_assistant') }}
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'pipelines' }"
            @click="activeTab = 'pipelines'"
          >
            {{ t('configs_page.pipelines') }}
            <span class="badge rounded-pill text-bg-light ms-2">{{ pipelines.length }}</span>
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'modules' }"
            @click="activeTab = 'modules'"
          >
            {{ t('configs_page.modules') }}
            <span class="badge rounded-pill text-bg-light ms-2">{{ moduleCatalogCount }}</span>
          </button>
          <button
            class="nav-link"
            :class="{ active: activeTab === 'preview' }"
            @click="activeTab = 'preview'"
          >
            {{ t('configs_page.preview') }}
          </button>
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
        importForm,
        importAnalysisLoading,
        importModulesLoading,
        importedConfigResult,
        importedWorkspaceModules,
        importedWorkspaceTemplate,
        importFlowPathLabel,
        importReuseDecisionCount,
        importCreateDecisionCount,
        importReusableMatchCount,
        importDestinationMatchCount,
        importSemanticChangeCount,
        importBlockingIssueCount,
      }"
      :actions="{ runImportAnalysis, importParsedModules, setAllImportedModuleActions, setImportedModuleAction }"
      :helpers="{
        runtimeLabel,
        importValidationBadgeClass,
        importValidationLabel,
        importActionBadgeClass,
        importActionLabel,
        importDestinationMatchLabel,
        importedModuleNameIssue,
      }"
    />

    <ConfigsWizardTab
      v-else-if="activeTab === 'wizard'"
      :state="{
        wizardForm,
        wizardServiceModuleId,
        wizardParserModuleIds,
        wizardServiceSearch,
        wizardParserSearch,
        wizardInputSearch,
        wizardFilterSearch,
        wizardRouteSearch,
        wizardOutputSearch,
        wizardStagePages,
        wizardPagedServiceModules,
        wizardPagedParserModules,
        wizardPagedInputModules,
        wizardPagedFilterModules,
        wizardPagedRouteModules,
        wizardPagedOutputTargets,
        wizardServiceModule,
        wizardSelectedParserModules,
        wizardGlobalVariableGroups,
        wizardPipelines,
        activeWizardPipelineId,
        activeWizardPipeline,
        wizardPipelineCards,
        wizardPipelineVariableGroups,
        wizardOutputResolutionWarnings,
        wizardIncompletePipelineLabels,
        wizardRenderSummary,
        renderedConfig,
        wizardAssemblyTemplates: wizardBuiltTemplates,
        wizardLoadedFromTemplate,
        wizardSaveButtonLabel,
        wizardCompatiblePipelines,
      }"
      :actions="{
        selectWizardServiceModule,
        toggleWizardParserModule,
        addWizardPipeline,
        duplicateWizardPipeline,
        removeWizardPipeline,
        selectWizardPipeline,
        setWizardPipelineInput,
        addWizardFilter,
        removeWizardFilter,
        moveWizardFilter,
        addWizardRoute,
        removeWizardRoute,
        moveWizardRoute,
        addWizardOutputTarget,
        removeWizardOutput,
        moveWizardOutput,
        changeWizardStagePage,
        runWizardPreview,
        saveWizardAsTemplate,
        openAdvancedPreviewFromWizard,
        loadWizardFromTemplate,
        clearWizardLoadedTemplate,
        addWizardPipelineFromSaved,
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
        moduleCatalogCount,
        sharedModuleCount,
        usedModuleTypes,
        moduleTypeStats,
        managedModuleTypes,
        moduleQuery,
        moduleTableRangeStart,
        moduleTableRangeEnd,
        moduleTableTotal,
        selectedDeletableModuleCount,
        moduleTableTotalPages,
        moduleTableLoading,
        visibleDeletableModules,
        allVisibleDeletableModulesSelected,
        visibleModules,
        selectedModuleIds,
      }"
      :actions="{
        applyModuleQuery,
        resetModuleQuery,
        changeModulePage,
        setModuleTypeFilter,
        toggleSelectAllVisibleModules,
        toggleModuleSelection,
        openEditModule,
        openModuleVersions,
        handleDeleteModule,
      }"
      :helpers="{ runtimeLabel, shortVariables, formatTime }"
    />

    <ConfigsPreviewTab
      v-else
      :state="{
        previewForm,
        previewAvailableOutputTargets,
        selectedPreviewOutputTargetIds,
        previewUnresolvedOutputTargets,
        previewModuleSearch,
        previewVisibleModules,
        selectedPreviewModuleIds,
        renderedConfig,
        previewFlowPathLabel,
        previewDestinationChips,
        previewSummaryModules,
        previewResolvedOutputTargets,
        analysisResult,
        compatibilityResult,
        replayResult,
        diffResult,
      }"
      :actions="{ togglePreviewOutputTarget, runPreview, runLint, runCompatibility, runReplay, runSemanticDiff, togglePreviewModule }"
      :helpers="{ runtimeLabel, findingBadgeClass, formatJson }"
    />

    <div class="modal fade" id="templateModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('configs_page.create_manual_template_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-warning py-2">
              <div class="fw-semibold">{{ t('configs_page.manual_template_mode_title') }}</div>
              <div class="small mt-1">{{ t('configs_page.manual_template_mode_hint') }}</div>
            </div>
            <div v-if="aiTemplateDraftState.active" class="fm-ai-draft-panel mb-3">
              <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
                <div>
                  <div class="fw-semibold">{{ t('configs_page.ai_draft_template_title') }}</div>
                  <div v-if="aiTemplateDraftSource" class="small text-muted mt-1">{{ aiTemplateDraftSource }}</div>
                </div>
                <span class="badge bg-success-subtle text-success-emphasis">{{ t('configs_page.ai_draft_imported') }}</span>
              </div>
              <div v-if="aiTemplateDraftState.summary" class="small mt-2">{{ aiTemplateDraftState.summary }}</div>
              <div v-if="aiTemplateDraftComparison" class="mt-3">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_diff_title') }}</div>
                <div class="fm-ai-draft-diff-grid">
                  <div class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_identity') }}</div>
                    <div class="fw-semibold">{{ aiTemplateDraftComparison.identityMessage }}</div>
                    <div v-if="aiTemplateDraftComparison.existingName" class="small text-muted mt-1">{{ aiTemplateDraftComparison.existingName }}</div>
                  </div>
                  <div class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_scale') }}</div>
                    <div class="fw-semibold">{{ t('configs_page.ai_draft_diff_lines').replace('{count}', String(aiTemplateDraftComparison.lineCount)) }}</div>
                    <div class="small text-muted mt-1">{{ t('configs_page.ai_draft_diff_placeholders').replace('{count}', String(aiTemplateDraftComparison.placeholderCount)) }}</div>
                  </div>
                  <div v-if="aiTemplateDraftComparison.changeMessage" class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_changes') }}</div>
                    <div class="fw-semibold">{{ aiTemplateDraftComparison.changeMessage }}</div>
                    <div v-if="aiTemplateDraftComparison.changeDetail" class="small text-muted mt-1">{{ aiTemplateDraftComparison.changeDetail }}</div>
                  </div>
                </div>
                <div v-if="aiTemplateDraftComparison.hasConflict" class="fm-ai-draft-actions mt-3">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_action_title') }}</div>
                  <div class="small text-muted mb-2">
                    {{ t('configs_page.ai_draft_action_existing_detail').replace('{name}', aiTemplateDraftComparison.existingName || templateForm.name) }}
                  </div>
                  <div v-if="aiTemplateDraftComparison.suggestedName" class="small text-muted mb-3">
                    {{ t('configs_page.ai_draft_action_name_suggested').replace('{name}', aiTemplateDraftComparison.suggestedName) }}
                  </div>
                  <div class="d-flex flex-wrap gap-2">
                    <button type="button" class="btn btn-sm btn-outline-primary" @click="applySuggestedTemplateName">
                      <i class="bi bi-magic me-1"></i>{{ t('configs_page.ai_draft_action_auto_rename') }}
                    </button>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="openExistingTemplateFromDraft">
                      <i class="bi bi-box-arrow-up-right me-1"></i>{{ t('configs_page.ai_draft_action_open_existing_template') }}
                    </button>
                  </div>
                </div>
              </div>
              <div v-if="aiTemplateDraftState.confirmationItems.length" class="mt-3">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_checklist_title') }}</div>
                <div class="small text-muted mb-2">{{ t('configs_page.ai_draft_checklist_hint') }}</div>
                <div class="fm-ai-draft-checklist">
                  <label v-for="item in aiTemplateDraftState.confirmationItems" :key="item.key" class="form-check fm-ai-draft-checklist__item">
                    <input v-model="item.checked" class="form-check-input" type="checkbox">
                    <span class="form-check-label">{{ item.label }}</span>
                  </label>
                </div>
              </div>
              <div class="row g-3 mt-1">
                <div v-if="aiTemplateDraftState.reviewItems.length" class="col-lg-6">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_review_title') }}</div>
                  <ul class="mb-0 ps-3">
                    <li v-for="(item, index) in aiTemplateDraftState.reviewItems" :key="`tpl-review-${index}`">{{ item }}</li>
                  </ul>
                </div>
                <div v-if="aiTemplateDraftState.notes.length || aiTemplateDraftState.steps.length" class="col-lg-6">
                  <div v-if="aiTemplateDraftState.steps.length">
                    <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_steps_title') }}</div>
                    <ul class="mb-2 ps-3">
                      <li v-for="(step, index) in aiTemplateDraftState.steps" :key="`tpl-step-${index}`">{{ step }}</li>
                    </ul>
                  </div>
                  <div v-if="aiTemplateDraftState.notes.length">
                    <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_notes_title') }}</div>
                    <ul class="mb-0 ps-3">
                      <li v-for="(note, index) in aiTemplateDraftState.notes" :key="`tpl-note-${index}`">{{ note }}</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
            <div class="row mb-3">
              <div class="col-md-6">
                <label class="form-label d-flex align-items-center gap-2">
                  <span>{{ t('common.name') }}</span>
                  <span v-if="aiTemplateDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <input v-model="templateForm.name" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiTemplateDraftState.active }" required>
              </div>
              <div class="col-md-6">
                <label class="form-label d-flex align-items-center gap-2">
                  <span>{{ t('common.runtime') }}</span>
                  <span v-if="aiTemplateDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <select v-model="templateForm.fluent_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiTemplateDraftState.active }">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.description') }}</span>
                <span v-if="aiTemplateDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <input v-model="templateForm.description" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiTemplateDraftState.active }">
            </div>
            <div class="mb-3">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('configs_page.template_content') }}</span>
                <span v-if="aiTemplateDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <div class="small text-muted mb-2">
                {{ t('configs_page.template_content_help') }}
              </div>
              <textarea
                v-model="templateForm.content"
                class="form-control font-monospace fm-config-textarea"
                :class="{ 'fm-ai-draft-highlight': aiTemplateDraftState.active }"
                rows="15"
                :placeholder="currentTemplateExample"
              ></textarea>
            </div>
            <div class="small text-muted">
              {{ t('configs_page.template_content_placeholder_hint') }}
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <div v-if="aiTemplateDraftState.active && !aiTemplateDraftReady" class="small text-muted me-auto">
              {{ t('configs_page.ai_draft_confirm_required') }}
            </div>
            <div v-else-if="aiTemplateDraftState.active && aiTemplateDraftComparison?.hasConflict" class="small text-warning me-auto">
              {{ t('configs_page.ai_draft_conflict_required') }}
            </div>
            <button type="button" class="btn btn-primary" :disabled="!aiTemplateDraftCanSave" @click="saveTemplate">
              {{ aiTemplateDraftState.active ? t('configs_page.ai_draft_confirm_template_cta') : t('create') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="moduleModal" tabindex="-1">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editingModuleId ? t('configs_page.edit_module_title') : t('configs_page.create_module_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info py-2">
              {{ t('configs_page.module_modal_hint') }}
            </div>
            <div v-if="aiModuleDraftState.active" class="fm-ai-draft-panel mb-3">
              <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
                <div>
                  <div class="fw-semibold">{{ t('configs_page.ai_draft_module_title') }}</div>
                  <div v-if="aiModuleDraftSource" class="small text-muted mt-1">{{ aiModuleDraftSource }}</div>
                </div>
                <span class="badge bg-success-subtle text-success-emphasis">{{ t('configs_page.ai_draft_imported') }}</span>
              </div>
              <div v-if="aiModuleDraftState.summary" class="small mt-2">{{ aiModuleDraftState.summary }}</div>
              <div v-if="aiModuleDraftComparison" class="mt-3">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_diff_title') }}</div>
                <div class="fm-ai-draft-diff-grid">
                  <div class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_identity') }}</div>
                    <div class="fw-semibold">{{ aiModuleDraftComparison.identityMessage }}</div>
                    <div v-if="aiModuleDraftComparison.existingName" class="small text-muted mt-1">{{ aiModuleDraftComparison.existingName }}</div>
                  </div>
                  <div class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_scale') }}</div>
                    <div class="fw-semibold">{{ t('configs_page.ai_draft_diff_variables').replace('{count}', String(aiModuleDraftComparison.variableCount)) }}</div>
                    <div class="small text-muted mt-1">
                      {{ t('configs_page.ai_draft_diff_lines').replace('{count}', String(aiModuleDraftComparison.lineCount)) }}
                      ·
                      {{ t('configs_page.ai_draft_diff_placeholders').replace('{count}', String(aiModuleDraftComparison.placeholderCount)) }}
                    </div>
                  </div>
                  <div v-if="aiModuleDraftComparison.changeMessage" class="fm-ai-draft-diff-card">
                    <div class="fm-ai-draft-diff-card__label">{{ t('configs_page.ai_draft_diff_changes') }}</div>
                    <div class="fw-semibold">{{ aiModuleDraftComparison.changeMessage }}</div>
                    <div v-if="aiModuleDraftComparison.changeDetail" class="small text-muted mt-1">{{ aiModuleDraftComparison.changeDetail }}</div>
                  </div>
                </div>
                <div v-if="aiModuleDraftComparison.hasConflict" class="fm-ai-draft-actions mt-3">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_action_title') }}</div>
                  <div class="small text-muted mb-2">
                    {{ t('configs_page.ai_draft_action_existing_detail').replace('{name}', aiModuleDraftComparison.existingName || moduleForm.name) }}
                  </div>
                  <div v-if="aiModuleDraftComparison.suggestedName" class="small text-muted mb-3">
                    {{ t('configs_page.ai_draft_action_name_suggested').replace('{name}', aiModuleDraftComparison.suggestedName) }}
                  </div>
                  <div class="d-flex flex-wrap gap-2">
                    <button type="button" class="btn btn-sm btn-outline-primary" @click="applySuggestedModuleName">
                      <i class="bi bi-magic me-1"></i>{{ t('configs_page.ai_draft_action_auto_rename') }}
                    </button>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="openExistingModuleFromDraft">
                      <i class="bi bi-pencil-square me-1"></i>{{ t('configs_page.ai_draft_action_open_existing_module') }}
                    </button>
                  </div>
                </div>
              </div>
              <div v-if="aiModuleDraftState.confirmationItems.length" class="mt-3">
                <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_checklist_title') }}</div>
                <div class="small text-muted mb-2">{{ t('configs_page.ai_draft_checklist_hint') }}</div>
                <div class="fm-ai-draft-checklist">
                  <label v-for="item in aiModuleDraftState.confirmationItems" :key="item.key" class="form-check fm-ai-draft-checklist__item">
                    <input v-model="item.checked" class="form-check-input" type="checkbox">
                    <span class="form-check-label">{{ item.label }}</span>
                  </label>
                </div>
              </div>
              <div class="row g-3 mt-1">
                <div v-if="aiModuleDraftState.reviewItems.length" class="col-lg-6">
                  <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_review_title') }}</div>
                  <ul class="mb-0 ps-3">
                    <li v-for="(item, index) in aiModuleDraftState.reviewItems" :key="`mod-review-${index}`">{{ item }}</li>
                  </ul>
                </div>
                <div v-if="aiModuleDraftState.notes.length || aiModuleDraftState.steps.length" class="col-lg-6">
                  <div v-if="aiModuleDraftState.steps.length">
                    <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_steps_title') }}</div>
                    <ul class="mb-2 ps-3">
                      <li v-for="(step, index) in aiModuleDraftState.steps" :key="`mod-step-${index}`">{{ step }}</li>
                    </ul>
                  </div>
                  <div v-if="aiModuleDraftState.notes.length">
                    <div class="fm-ai-draft-panel__title">{{ t('configs_page.ai_draft_notes_title') }}</div>
                    <ul class="mb-0 ps-3">
                      <li v-for="(note, index) in aiModuleDraftState.notes" :key="`mod-note-${index}`">{{ note }}</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-4">
                <label class="form-label d-flex align-items-center gap-2">
                  <span>{{ t('common.name') }}</span>
                  <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <input v-model="moduleForm.name" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }">
              </div>
              <div class="col-md-4">
                <label class="form-label d-flex align-items-center gap-2">
                  <span>{{ t('configs_page.module_type_coverage') }}</span>
                  <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <select v-model="moduleForm.module_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }">
                  <option v-for="type in moduleTypes" :key="type" :value="type">{{ type }}</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label d-flex align-items-center gap-2">
                  <span>{{ t('common.runtime') }}</span>
                  <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                </label>
                <select v-model="moduleForm.fluent_type" class="form-select" :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                  <option value="shared">Shared</option>
                </select>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ t('common.description') }}</span>
                <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <input v-model="moduleForm.description" type="text" class="form-control" :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }">
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-8">
                <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
                  <label class="form-label mb-0 d-flex align-items-center gap-2">
                    <span>{{ t('configs_page.variables_json') }}</span>
                    <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
                  </label>
                  <div class="btn-group btn-group-sm" role="group" :aria-label="t('configs_page.variable_input_mode')">
                    <button
                      type="button"
                      class="btn"
                      :class="moduleVariablesMode === 'form' ? 'btn-primary' : 'btn-outline-primary'"
                      @click="setModuleVariablesMode('form')"
                    >
                      {{ t('configs_page.variable_mode_form') }}
                    </button>
                    <button
                      type="button"
                      class="btn"
                      :class="moduleVariablesMode === 'json' ? 'btn-primary' : 'btn-outline-primary'"
                      @click="setModuleVariablesMode('json')"
                    >
                      {{ t('configs_page.variable_mode_json') }}
                    </button>
                  </div>
                </div>
                <div class="small text-muted mb-2">
                  {{ t('configs_page.variables_help') }}
                </div>
                <div v-if="moduleVariablesMode === 'form'" class="border rounded-3 p-3 bg-light-subtle">
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <div class="small text-muted">{{ t('configs_page.variable_form_help') }}</div>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="addModuleVariableRow()">
                      <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.add_variable') }}
                    </button>
                  </div>
                  <div v-if="!moduleVariableRows.length" class="text-center text-muted py-3">
                    {{ t('configs_page.no_variables_rows') }}
                  </div>
                  <div v-for="(row, index) in moduleVariableRows" :key="index" class="row g-2 align-items-start mb-2">
                    <div class="col-md-4">
                      <input
                        v-model="row.key"
                        type="text"
                        class="form-control"
                        :placeholder="t('configs_page.variable_name_placeholder')"
                        @input="syncModuleVariablesFromRows"
                      >
                    </div>
                    <div class="col-md-3">
                      <select v-model="row.type" class="form-select" @change="syncModuleVariablesFromRows">
                        <option value="string">{{ t('configs_page.variable_type_string') }}</option>
                        <option value="number">{{ t('configs_page.variable_type_number') }}</option>
                        <option value="boolean">{{ t('configs_page.variable_type_boolean') }}</option>
                        <option value="json">{{ t('configs_page.variable_type_json') }}</option>
                      </select>
                    </div>
                    <div class="col-md-4">
                      <select
                        v-if="row.type === 'boolean'"
                        v-model="row.value"
                        class="form-select"
                        @change="syncModuleVariablesFromRows"
                      >
                        <option value="true">true</option>
                        <option value="false">false</option>
                      </select>
                      <textarea
                        v-else-if="row.type === 'json'"
                        v-model="row.value"
                        class="form-control font-monospace"
                        rows="3"
                        :placeholder="t('configs_page.variable_json_placeholder')"
                        @input="syncModuleVariablesFromRows"
                      ></textarea>
                      <input
                        v-else
                        v-model="row.value"
                        type="text"
                        class="form-control"
                        :placeholder="row.type === 'number' ? '24224' : t('configs_page.variable_value_placeholder')"
                        @input="syncModuleVariablesFromRows"
                      >
                    </div>
                    <div class="col-md-1 d-grid">
                      <button type="button" class="btn btn-outline-danger" @click="removeModuleVariableRow(index)">
                        <i class="bi bi-trash"></i>
                      </button>
                    </div>
                  </div>
                  <div v-if="moduleVariablesFormError" class="small text-danger mt-2">
                    {{ moduleVariablesFormError }}
                  </div>
                </div>
                <textarea
                  v-else
                  v-model="moduleForm.variables"
                  class="form-control font-monospace fm-config-textarea"
                  :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }"
                  rows="5"
                  :placeholder="currentModuleExample.variables"
                ></textarea>
              </div>
              <div class="col-md-4">
                <label class="form-label d-block">Props</label>
                <div class="form-check mt-2">
                  <input id="moduleBuiltin" v-model="moduleForm.is_builtin" type="checkbox" class="form-check-input">
                  <label for="moduleBuiltin" class="form-check-label">{{ t('configs_page.builtin_module') }}</label>
                </div>
                <div class="small text-muted mt-2">
                  {{ t('configs_page.builtin_help') }}
                </div>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label d-flex align-items-center gap-2">
                <span>{{ moduleForm.fluent_type === 'shared' ? 'Fluent Bit ' + t('configs_page.version_content') : t('configs_page.version_content') }}</span>
                <span v-if="aiModuleDraftState.active" class="badge text-bg-light">{{ t('configs_page.ai_draft_filled') }}</span>
              </label>
              <div class="small text-muted mb-2">
                {{ t('configs_page.version_content_help') }}
              </div>
              <textarea
                v-model="moduleForm.content"
                class="form-control font-monospace fm-config-textarea"
                :class="{ 'fm-ai-draft-highlight': aiModuleDraftState.active }"
                rows="16"
                :placeholder="currentModuleExample.content"
              ></textarea>
            </div>
            <div v-if="moduleForm.fluent_type === 'shared'" class="mb-3">
              <label class="form-label">Fluentd {{ t('configs_page.version_content') }}</label>
              <div class="small text-muted mb-2">{{ t('configs_page.version_content_help') }}</div>
              <textarea
                v-model="moduleForm.content_fluentd"
                class="form-control font-monospace fm-config-textarea"
                rows="16"
                :placeholder="(moduleExamples.fluentd?.[moduleForm.module_type] || moduleExamples.fluentd?.input || {}).content || ''"
              ></textarea>
              <div class="small text-muted mt-1">{{ t('configs_page.module_shared_fluentd_hint', 'Leave empty to reuse Fluent Bit content for Fluentd.') }}</div>
            </div>
            <div class="small text-muted">
              {{ t('configs_page.template_syntax_hint') }}
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <div v-if="aiModuleDraftState.active && !aiModuleDraftReady" class="small text-muted me-auto">
              {{ t('configs_page.ai_draft_confirm_required') }}
            </div>
            <div v-else-if="aiModuleDraftState.active && aiModuleDraftComparison?.hasConflict && !editingModuleId" class="small text-warning me-auto">
              {{ t('configs_page.ai_draft_conflict_required') }}
            </div>
            <button type="button" class="btn btn-primary" :disabled="!aiModuleDraftCanSave" @click="saveModule">
              {{ editingModuleId ? t('save') : (aiModuleDraftState.active ? t('configs_page.ai_draft_confirm_module_cta') : t('configs_page.create_module')) }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="pipelineModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editingPipelineId ? t('configs_page.edit_pipeline_title') : t('configs_page.create_pipeline') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.name') }} <span class="text-danger">*</span></label>
                <input v-model="pipelineForm.name" type="text" class="form-control" :placeholder="t('configs_page.pipeline_name_placeholder')" />
              </div>
              <div class="col-md-6">
                <label class="form-label">Runtime</label>
                <select v-model="pipelineForm.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <input v-model="pipelineForm.description" type="text" class="form-control" />
            </div>

            <!-- Input module -->
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ t('configs_page.pipeline_input') }}</label>
              <select v-model="pipelineForm.input_module_id" class="form-select">
                <option :value="null">— {{ t('none') }} —</option>
                <option v-for="m in pipelineInputModules" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
            </div>

            <!-- Filter modules -->
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ t('configs_page.pipeline_filters') }}</label>
              <div v-if="pipelineForm.filter_module_ids.length" class="list-group mb-2">
                <div
                  v-for="(fid, idx) in pipelineForm.filter_module_ids"
                  :key="fid"
                  class="list-group-item d-flex justify-content-between align-items-center py-1 px-2"
                >
                  <span class="badge bg-secondary me-2">{{ idx + 1 }}</span>
                  <span class="flex-grow-1 font-monospace small">{{ pipelineEligibleModules.find((m) => m.id === fid)?.name || fid }}</span>
                  <div class="d-flex gap-1 ms-2">
                    <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" :disabled="idx === 0" @click="movePipelineFilterModule(idx, -1)"><i class="bi bi-arrow-up"></i></button>
                    <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" :disabled="idx === pipelineForm.filter_module_ids.length - 1" @click="movePipelineFilterModule(idx, 1)"><i class="bi bi-arrow-down"></i></button>
                    <button type="button" class="btn btn-sm btn-outline-danger py-0 px-1" @click="removePipelineFilterModule(idx)"><i class="bi bi-x"></i></button>
                  </div>
                </div>
              </div>
              <div class="input-group">
                <select v-model="pipelineFilterPickerValue" class="form-select form-select-sm">
                  <option value="">{{ t('configs_page.pipeline_add_filter') }}…</option>
                  <option v-for="m in pipelineFilterModules" :key="m.id" :value="m.id">{{ m.name }}</option>
                </select>
                <button type="button" class="btn btn-sm btn-outline-secondary" @click="addPipelineFilterModule(Number(pipelineFilterPickerValue)); pipelineFilterPickerValue = ''">
                  <i class="bi bi-plus"></i>
                </button>
              </div>
            </div>

            <!-- Output targets -->
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ t('configs_page.pipeline_outputs') }}</label>
              <div class="d-flex flex-wrap gap-2">
                <label
                  v-for="target in pipelineAvailableOutputTargets"
                  :key="target.id"
                  class="d-flex align-items-center gap-1 border rounded px-2 py-1 small"
                  style="cursor:pointer"
                >
                  <input type="checkbox" :checked="pipelineForm.output_target_ids.includes(target.id)" @change="togglePipelineOutputTarget(target.id)" />
                  {{ target.name }}
                </label>
                <div v-if="!pipelineAvailableOutputTargets.length" class="text-muted small">{{ t('configs_page.no_output_targets') }}</div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="savePipeline">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="moduleVersionsModal" tabindex="-1">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <div>
              <h5 class="modal-title mb-1">{{ t('configs_page.module_versions_title') }}</h5>
              <div class="small text-muted">{{ currentModule?.name || '-' }}</div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-4">
              <div class="col-lg-4">
                <div class="card border-0 bg-light h-100">
                  <div class="card-body">
                    <h6 class="mb-3">{{ t('configs_page.history_versions') }}</h6>
                    <div class="list-group list-group-flush rounded overflow-hidden">
                      <div v-for="version in moduleVersions" :key="version.id" class="list-group-item">
                        <div class="d-flex justify-content-between align-items-start mb-1">
                          <strong>v{{ version.version }}</strong>
                          <small class="text-muted">{{ formatTime(version.created_at) }}</small>
                        </div>
                        <div class="small text-muted mb-1">{{ version.comment || t('configs_page.no_version_comment') }}</div>
                        <code class="small">{{ version.hash }}</code>
                      </div>
                      <div v-if="!moduleVersions.length" class="list-group-item text-muted">{{ t('configs_page.no_history_versions') }}</div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-lg-8">
                <h6 class="mb-3">{{ t('configs_page.create_new_version') }}</h6>
                <div class="mb-3">
                  <label class="form-label">{{ t('configs_page.version_comment') }}</label>
                  <input v-model="moduleVersionForm.comment" type="text" class="form-control">
                </div>
                <div class="mb-3">
                  <label class="form-label">{{ t('configs_page.variables_json') }}</label>
                  <textarea v-model="moduleVersionForm.variables" class="form-control font-monospace" rows="4"></textarea>
                </div>
                <div class="mb-3">
                  <label class="form-label">{{ t('configs_page.version_content') }}</label>
                  <textarea v-model="moduleVersionForm.content" class="form-control font-monospace" rows="14"></textarea>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('common.close') }}</button>
            <button type="button" class="btn btn-primary" @click="saveModuleVersion">{{ t('configs_page.create_new_version') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import './configs.css'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ConfigsAssistantTab from './tabs/ConfigsAssistantTab.vue'
import ConfigsImportTab from './tabs/ConfigsImportTab.vue'
import ConfigsModulesTab from './tabs/ConfigsModulesTab.vue'
import ConfigsPipelinesTab from './tabs/ConfigsPipelinesTab.vue'
import ConfigsPreviewTab from './tabs/ConfigsPreviewTab.vue'
import ConfigsTemplatesTab from './tabs/ConfigsTemplatesTab.vue'
import ConfigsWizardTab from './tabs/ConfigsWizardTab.vue'
import { useI18n } from '../../i18n'
import { buildConfigFlowSummary } from '../../utils/config_flow'
import {
  analyzeLogSampleAssistant,
  checkCompatibility,
  createModule,
  createModuleVersion,
  createOutputTarget,
  createVersion,
  createTemplate,
  deleteModules,
  deleteModule,
  deleteTemplate,
  diffConfig,
  getModuleVersions,
  getModules,
  getOutputTargets,
  getRenderedConfig,
  getTemplates,
  importExistingConfig,
  lintConfig,
  previewRenderedConfig,
  replayConfig,
  updateTemplate,
  updateModule,
} from '../../api'
import {
  getConfigPipelines,
  createConfigPipeline,
  updateConfigPipeline,
  deleteConfigPipeline,
} from '../../api/configs'

const activeTab = ref('templates')
const templates = ref([])
const modules = ref([])
const pipelines = ref([])
const editingPipelineId = ref(null)
const pipelineForm = reactive({
  name: '',
  description: '',
  fluent_type: 'fluentbit',
  input_module_id: null,
  filter_module_ids: [],
  output_target_ids: [],
})
const pipelineFilterPickerValue = ref('')
const pipelineFormInitializing = ref(false)
const moduleTableItems = ref([])
const moduleTableTotal = ref(0)
const moduleTableLoading = ref(false)
const outputTargets = ref([])
const renderedConfig = ref(null)
const analysisResult = ref(null)
const compatibilityResult = ref(null)
const replayResult = ref(null)
const diffResult = ref(null)
const currentModule = ref(null)
const moduleVersions = ref([])
const selectedModuleIds = ref([])
const selectedPreviewModuleIds = ref([])
const selectedPreviewOutputTargetIds = ref([])
const selectedWizardModuleIds = ref([])
const selectedWizardOutputTargetIds = ref([])
const aiAssistantLoading = ref(false)
const aiAssistantResult = ref(null)
// aiAssistantModules: generated modules enriched with a local merge decision.
// decision: 'create_new' | 'reuse_existing' | 'update_existing'
// matchedModule: the existing module that was matched (if any)
const aiAssistantModules = ref([])
const aiAssistantModulesSaving = ref(false)
const importAnalysisLoading = ref(false)
const importModulesLoading = ref(false)
const importedConfigResult = ref(null)
const importedWorkspaceModules = ref([])
const importedWorkspaceTemplate = ref(null)
const aiAssistantFeedback = reactive({
  type: '',
  message: '',
  detail: '',
  provider: '',
  providerDetail: '',
})
const aiModuleDraftState = reactive(createAIDraftState())
const aiTemplateDraftState = reactive(createAIDraftState())
const editingModuleId = ref(null)
const moduleVariablesMode = ref('form')
const moduleVariableRows = ref([])
const moduleVariablesFormError = ref('')
const previewModuleSearch = ref('')
const wizardModuleSearch = ref('')
const previewModuleVariables = ref({})
const wizardModuleVariableValues = ref({})
const moduleQuery = reactive({
  search: '',
  fluent_type: '',
  module_type: '',
  page: 1,
  page_size: 20,
})

const moduleTypes = ['service', 'input', 'parser', 'filter', 'route', 'output']
const managedModuleTypes = ['service', 'input', 'parser', 'filter', 'route', 'output']
const wizardPipelineModuleTypes = ['input', 'filter', 'route']
const wizardPipelineStageTotal = 3
const wizardStagePageSize = 6
let wizardSequence = 0
const moduleExamples = {
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
    service: {
      variables: '{\n  "flush_seconds": 5\n}',
      content: '# Shared service tuning\n# Flush interval: {{ .flush_seconds }}s',
    },
    input: {
      variables: '{\n  "tag": "app.logs"\n}',
      content: '# Shared input hints\n# Default tag {{ .tag }}',
    },
    parser: {
      variables: '{\n  "time_key": "time"\n}',
      content: '# Shared parser hint\n# Use time key {{ .time_key }}',
    },
    filter: {
      variables: '{\n  "match": "app.*",\n  "env": "prod"\n}',
      content: '# Shared filter intent\n# Match {{ .match }} and stamp env {{ .env }}',
    },
    route: {
      variables: '{\n  "match": "app.*",\n  "destination": "central-forward"\n}',
      content: '# Shared routing hint\n# Route {{ .match }} to {{ .destination }}',
    },
    output: {
      variables: '{\n  "host": "10.0.0.15",\n  "port": 24224\n}',
      content: '# Shared output hint\n# Forward to {{ .host }}:{{ .port }}',
    },
  },
}
const templateExamples = {
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

const templateForm = reactive({
  name: '',
  description: '',
  fluent_type: 'fluentbit',
  content: '',
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

const previewForm = reactive({
  name: '',
  fluent_type: 'fluentbit',
  runtime_version: '',
  variables: '{\n  "path": "/var/log/*.log",\n  "match": "*"\n}',
  node_id: '',
  sample_tag: 'app.logs',
  sample_log: '{"message":"hello fluent","level":"info"}',
  diff_content: '',
})

const importForm = reactive({
  fluent_type: 'fluentbit',
  name_prefix: 'imported-config',
  content: '',
})

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
const wizardSaveButtonLabel = computed(() => (
  wizardLoadedFromTemplate.value
    ? t('configs_page.save_wizard_version')
    : t('configs_page.save_wizard_template')
))
const wizardServiceSearch = ref('')
const wizardParserSearch = ref('')
const wizardInputSearch = ref('')
const wizardFilterSearch = ref('')
const wizardRouteSearch = ref('')
const wizardOutputSearch = ref('')
const wizardStagePages = reactive({
  service: 1,
  parser: 1,
  input: 1,
  filter: 1,
  route: 1,
  output: 1,
})
const aiAssistantForm = reactive({
  fluent_type: 'fluentbit',
  goal: 'both',
  module_type: 'input',
  sample: '',
  extra_requirements: '',
})

const moduleCatalogCount = computed(() => modules.value.length)
const visibleModules = computed(() => moduleTableItems.value)
const visibleDeletableModules = computed(() => visibleModules.value.filter((item) => !item.is_builtin))
const allVisibleDeletableModulesSelected = computed(() =>
  visibleDeletableModules.value.length > 0 && visibleDeletableModules.value.every((item) => selectedModuleIds.value.includes(item.id))
)
const selectedDeletableModuleCount = computed(() => selectedModuleIds.value.length)
const assemblyTemplates = computed(() => templates.value.filter((item) => item.source_type === 'module_assembly'))
const assemblyTemplateCount = computed(() => assemblyTemplates.value.length)
const wizardBuiltTemplates = computed(() =>
  assemblyTemplates.value.filter((tpl) => {
    if (!tpl.flow_layout) return false
    try {
      const layout = typeof tpl.flow_layout === 'string' ? JSON.parse(tpl.flow_layout) : tpl.flow_layout
      return layout.builder === 'wizard'
    } catch {
      return false
    }
  })
)
const manualTemplateCount = computed(() => templates.value.filter((item) => item.source_type !== 'module_assembly').length)
const sharedModuleCount = computed(() => modules.value.filter((item) => item.fluent_type === 'shared').length)
const usedModuleTypes = computed(() => managedModuleTypes.filter((type) => modules.value.some((item) => item.module_type === type)))
const moduleTypeStats = computed(() =>
  managedModuleTypes
    .map((type) => ({
      type,
      count: modules.value.filter((item) => item.module_type === type).length,
    }))
    .filter((item) => item.count > 0)
)
const moduleTableTotalPages = computed(() => Math.max(1, Math.ceil(moduleTableTotal.value / Math.max(Number(moduleQuery.page_size) || 20, 1))))
const moduleTableRangeStart = computed(() => (moduleTableTotal.value ? (moduleQuery.page - 1) * moduleQuery.page_size + 1 : 0))
const moduleTableRangeEnd = computed(() => (moduleTableTotal.value ? Math.min(moduleQuery.page * moduleQuery.page_size, moduleTableTotal.value) : 0))
const previewEligibleModules = computed(() =>
  modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
)
const previewVisibleModules = computed(() =>
  previewEligibleModules.value.filter((item) => item.module_type !== 'output' && matchesModuleSearch(item, previewModuleSearch.value))
)
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
const pipelineEligibleModules = computed(() =>
  modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === pipelineForm.fluent_type)
)
const pipelineInputModules = computed(() =>
  pipelineEligibleModules.value.filter((item) => item.module_type === 'input')
)
const pipelineFilterModules = computed(() =>
  pipelineEligibleModules.value.filter((item) => item.module_type === 'filter')
)
const pipelineAvailableOutputTargets = computed(() =>
  outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === pipelineForm.fluent_type)
)
const previewAvailableOutputTargets = computed(() =>
  outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
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
      id: pipeline.id,
      index,
      name: pipeline.name,
      inputModule,
      filterModules,
      routeModules,
      outputTargets: outputTargetsForPipeline,
      outputModules: outputModulesForPipeline,
      summary,
      missing,
      complete: missing.length === 0,
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
  const globalSection = {
    key: 'global',
    title: t('configs_page.wizard_global_resources'),
    kind: 'global',
  }
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
    const pipelineSection = {
      key: `pipeline:${pipeline.id}`,
      title: pipelineLabel,
      kind: 'pipeline',
      ref: pipeline.id,
    }

    if (pipeline.input) {
      const module = wizardEligibleModules.value.find((item) => item.id === pipeline.input.module_id)
      groups.push(buildWizardModuleGroup(
        pipeline.input.id,
        module?.name || t('configs_page.pipeline_stage_input'),
        `${pipelineLabel} · ${t('configs_page.pipeline_stage_input')}`,
        module,
        pipeline.input.variables,
        wizardPipelineModuleDefaults(module, pipeline),
        pipelineSection
      ))
    }

    pipeline.filters.forEach((instance, filterIndex) => {
      const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
      groups.push(buildWizardModuleGroup(
        instance.id,
        module?.name || t('configs_page.pipeline_stage_filter'),
        `${pipelineLabel} · ${t('configs_page.pipeline_stage_filter')} ${filterIndex + 1}`,
        module,
        instance.variables,
        wizardPipelineModuleDefaults(module, pipeline),
        pipelineSection
      ))
    });

    (pipeline.routes || []).forEach((instance, routeIndex) => {
      const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
      groups.push(buildWizardModuleGroup(
        instance.id,
        module?.name || t('configs_page.pipeline_stage_route', 'Route'),
        `${pipelineLabel} · ${t('configs_page.pipeline_stage_route', 'Route')} ${routeIndex + 1}`,
        module,
        instance.variables,
        wizardPipelineModuleDefaults(module, pipeline),
        pipelineSection
      ))
    })

    pipeline.outputs.forEach((instance, outputIndex) => {
      const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
      const module = matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)
      groups.push(buildWizardModuleGroup(
        instance.id,
        target?.name || t('configs_page.pipeline_stage_output'),
        `${pipelineLabel} · ${t('configs_page.pipeline_stage_output')} ${outputIndex + 1}`,
        module,
        instance.variables,
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
      .map((target) => ({
        pipeline: wizardPipelineDisplayName(card, index),
        target: target.name,
      })))
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
      hint: type === 'input'
        ? t('configs_page.wizard_input_group_hint')
        : t('configs_page.wizard_filter_group_hint'),
    }))
    .filter((group) => group.modules.length)
)
const wizardVariableGroups = computed(() =>
  wizardSelectedModules.value
    .map((module) => {
      const variables = parseVariablesMap(module.variables)
      const fields = Object.entries(variables).map(([key, value]) => ({
        key,
        defaultValue: stringifyVariableValue(value),
        kind: inferVariableKind(value),
        description: module.description || '',
      }))
      return {
        moduleId: module.id,
        moduleName: module.name,
        moduleType: module.module_type,
        moduleTypeLabel: module.module_type === 'input'
          ? t('configs_page.pipeline_stage_input')
          : t('configs_page.pipeline_stage_filter'),
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
const currentModuleExample = computed(() => {
  const runtimeExamples = moduleExamples[moduleForm.fluent_type] || moduleExamples.shared
  return runtimeExamples[moduleForm.module_type] || runtimeExamples.input || {
    variables: '{}',
    content: '# Example content',
  }
})
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
const currentTemplateExample = computed(() => templateExamples[templateForm.fluent_type] || templateExamples.fluentbit)
const aiModuleDraftSource = computed(() => [aiModuleDraftState.provider, aiModuleDraftState.accountName].filter(Boolean).join(' / '))
const aiTemplateDraftSource = computed(() => [aiTemplateDraftState.provider, aiTemplateDraftState.accountName].filter(Boolean).join(' / '))
const aiModuleDraftReady = computed(() => areDraftConfirmationsComplete(aiModuleDraftState))
const aiTemplateDraftReady = computed(() => areDraftConfirmationsComplete(aiTemplateDraftState))
const aiModuleDraftComparison = computed(() => buildModuleDraftComparison())
const aiTemplateDraftComparison = computed(() => buildTemplateDraftComparison())
const aiModuleDraftCanSave = computed(() => {
  if (!aiModuleDraftReady.value) return false
  if (!aiModuleDraftState.active) return true
  return !!editingModuleId.value || !aiModuleDraftComparison.value?.hasConflict
})
const aiTemplateDraftCanSave = computed(() => {
  if (!aiTemplateDraftReady.value) return false
  if (!aiTemplateDraftState.active) return true
  return !aiTemplateDraftComparison.value?.hasConflict
})

let templateModal = null
let moduleModal = null
let moduleVersionsModal = null
let pipelineModal = null
const route = useRoute()
const router = useRouter()
const { t, dateLocale } = useI18n()

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : '-'
}

function runtimeLabel(value) {
  if (value === 'fluentbit') return 'Fluent Bit'
  if (value === 'fluentd') return 'Fluentd'
  if (value === 'shared') return 'Shared'
  return value || '-'
}

function runtimeBadgeClass(value) {
  if (value === 'fluentbit') return 'bg-info-subtle text-info-emphasis'
  if (value === 'fluentd') return 'bg-warning-subtle text-warning-emphasis'
  return 'bg-secondary-subtle text-secondary-emphasis'
}

function shortVariables(value) {
  if (!value || value === '{}') return '{}'
  return value.length > 42 ? `${value.slice(0, 42)}...` : value
}

function parseJSONList(value) {
  if (!value) return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function formatJson(value) {
  try {
    return JSON.stringify(value || {}, null, 2)
  } catch {
    return String(value || '{}')
  }
}

function normalizeSearchText(value) {
  return String(value || '').trim().toLowerCase()
}

function matchesModuleSearch(module, keyword) {
  const normalizedKeyword = normalizeSearchText(keyword)
  if (!normalizedKeyword) return true

  const haystack = [
    module?.name,
    module?.description,
    module?.module_type,
    module?.fluent_type,
    module?.preset_kind,
    module?.preset_key,
    module?.content,
  ]
    .filter(Boolean)
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalizedKeyword)
}

function matchesOutputTargetSearch(target, keyword) {
  const normalizedKeyword = normalizeSearchText(keyword)
  if (!normalizedKeyword) return true

  const haystack = [
    target?.name,
    target?.description,
    target?.target_type,
    target?.endpoint,
    target?.settings,
  ]
    .filter(Boolean)
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalizedKeyword)
}

function templateSourceLabel(sourceType) {
  return sourceType === 'module_assembly'
    ? t('configs_page.template_source_module_assembly')
    : t('configs_page.template_source_manual')
}

function templateAssemblyModules(template) {
  return parseJSONList(template?.source_modules)
}

function buildImportedModuleDescription(item) {
  const details = [t('configs_page.imported_from_existing_config')]
  if (item?.summary) {
    details.push(item.summary)
  }
  return details.join(' · ')
}

function importedModuleIdentity(name, moduleType, fluentType) {
  return [normalizeName(name), String(moduleType || '').trim(), String(fluentType || '').trim()].join('::')
}

function uniqueImportedModuleName(baseName, moduleType, fluentType, occupiedIdentities) {
  const seed = normalizeName(baseName) ? String(baseName).trim() : 'imported-module'
  if (moduleType === 'parser') {
    return seed
  }
  const seedIdentity = importedModuleIdentity(seed, moduleType, fluentType)
  if (!occupiedIdentities.has(seedIdentity)) {
    occupiedIdentities.add(seedIdentity)
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (occupiedIdentities.has(importedModuleIdentity(candidate, moduleType, fluentType))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  occupiedIdentities.add(importedModuleIdentity(candidate, moduleType, fluentType))
  return candidate
}

function findImportedModuleNameConflict(modules, occupiedIdentities) {
  const batchIdentities = new Set()
  for (const item of modules || []) {
    if (item.module_type === 'output' || item.import_action === 'reuse_existing') {
      continue
    }
    const name = String(item.name || '').trim()
    if (!name) {
      continue
    }
    const identity = importedModuleIdentity(name, item.module_type, item.fluent_type)
    if (batchIdentities.has(identity)) {
      return {
        type: 'batch_duplicate',
        item,
      }
    }
    batchIdentities.add(identity)
    if (item.module_type === 'parser' && occupiedIdentities.has(identity)) {
      return {
        type: 'existing_duplicate',
        item,
      }
    }
  }
  return null
}

function importedModuleNameIssue(module) {
  if (!module || module.module_type === 'output' || module.import_action === 'reuse_existing') {
    return null
  }

  const name = String(module.name || '').trim()
  if (!name) {
    return {
      type: 'required',
      message: t('configs_page.import_module_name_required').replace('{order}', String(module.order || '')),
    }
  }

  const identity = importedModuleIdentity(name, module.module_type, module.fluent_type)
  const duplicateInBatch = (importedConfigResult.value?.modules || []).filter((item) => {
    if (item.module_type === 'output' || item.import_action === 'reuse_existing') {
      return false
    }
    return importedModuleIdentity(String(item.name || '').trim(), item.module_type, item.fluent_type) === identity
  }).length > 1
  if (duplicateInBatch) {
    return {
      type: 'batch_duplicate',
      message: t('configs_page.import_module_name_duplicate_batch').replace('{name}', name),
    }
  }

  if (module.module_type === 'parser') {
    const existsInWorkspace = modules.value.some((item) =>
      importedModuleIdentity(item.name, item.module_type, item.fluent_type) === identity
    )
    if (existsInWorkspace) {
      return {
        type: 'existing_duplicate',
        message: t('configs_page.import_module_name_duplicate_existing').replace('{name}', name),
      }
    }
  }

  return null
}

const importBlockingIssueCount = computed(() =>
  (importedConfigResult.value?.modules || []).filter((module) => !!importedModuleNameIssue(module)).length
)

function importedOutputTargetNameSeed(module, fallbackPrefix) {
  const type = inferImportedOutputTargetType(module) || 'custom'
  const prefix = String(fallbackPrefix || importForm.name_prefix || 'imported-config').trim() || 'imported-config'
  return `${prefix}-${type}-destination`
}

function uniqueImportedOutputTargetName(baseName, occupiedNames) {
  const seed = normalizeName(baseName) ? String(baseName).trim() : 'imported-destination'
  if (!occupiedNames.has(normalizeName(seed))) {
    occupiedNames.add(normalizeName(seed))
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (occupiedNames.has(normalizeName(candidate))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  occupiedNames.add(normalizeName(candidate))
  return candidate
}

function normalizeImportedOutputModuleContent(content) {
  return String(content || '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .join('\n')
}

function createImportedOutputModuleSignature(module) {
  return [
    String(module?.module_type || '').trim().toLowerCase(),
    String(module?.fluent_type || '').trim().toLowerCase(),
    normalizeImportedOutputModuleContent(module?.content),
  ].join('::')
}

function sortObjectKeys(value) {
  if (Array.isArray(value)) {
    return value.map((item) => sortObjectKeys(item))
  }
  if (!value || typeof value !== 'object') {
    return value
  }
  return Object.keys(value)
    .sort()
    .reduce((acc, key) => {
      acc[key] = sortObjectKeys(value[key])
      return acc
    }, {})
}

function stringifySortedObject(value) {
  return JSON.stringify(sortObjectKeys(value || {}))
}

function pickImportedOutputTargetSettings(rawSettings, targetType) {
  const settings = { ...(rawSettings || {}) }
  const keepByTargetType = {
    opensearch: [
      'host', 'port', 'http_user', 'http_passwd', 'cloud_id', 'path', 'uri',
      'tls', 'tls.verify', 'aws_auth', 'aws_region', 'aws_service_name',
    ],
    loki: ['host', 'port', 'uri', 'tenant_id', 'http_user', 'http_passwd', 'tls', 'tls.verify'],
    kafka: ['brokers', 'client_id', 'security.protocol', 'sasl.username', 'sasl.password', 'tls', 'tls.verify'],
    http: ['host', 'port', 'uri', 'http_user', 'http_passwd', 'proxy', 'tls', 'tls.verify'],
    s3: ['bucket', 'region', 'endpoint', 'host', 'port', 'uri', 'access_key_id', 'secret_access_key', 'role_arn', 'tls', 'tls.verify'],
    stdout: [],
  }
  const keepKeys = keepByTargetType[targetType] || []
  if (keepKeys.length) {
    return keepKeys.reduce((acc, key) => {
      if (settings[key] !== undefined && settings[key] !== null && String(settings[key]).trim() !== '') {
        acc[key] = settings[key]
      }
      return acc
    }, {})
  }

  const dropKeys = new Set([
    'match', 'logstash_format', 'logstash_prefix', 'logstash_dateformat', 'index', 'topic', 'topics',
    'labels', 'label_keys', 'remove_keys', 'generate_id', 'retry_limit', 'workers', 'compress',
    'replace_dots', 'suppress_type_name', 'trace_error', 'time_key', 'time_key_format',
  ])
  return Object.entries(settings).reduce((acc, [key, value]) => {
    if (dropKeys.has(key)) return acc
    if (value === undefined || value === null || String(value).trim() === '') return acc
    acc[key] = value
    return acc
  }, {})
}

function createImportedOutputTargetSignature(target) {
  const targetType = String(target?.target_type || '').trim().toLowerCase()
  const endpoint = String(target?.endpoint || '').trim().toLowerCase()
  const settings = typeof target?.settings === 'string' ? parseVariablesMap(target.settings) : (target?.settings || {})
  const filteredSettings = pickImportedOutputTargetSettings(settings, targetType)
  return [targetType, endpoint, stringifySortedObject(filteredSettings)].join('::')
}

function uniqueImportedDestinationList(destinations) {
  const seen = new Set()
  const filtered = []
  for (const item of destinations || []) {
    const id = Number(item?.output_target_id || 0)
    if (!id || seen.has(id)) continue
    seen.add(id)
    filtered.push(item)
  }
  return filtered
}

function findImportedOutputAdapterModule(targetType, modules, fluentType) {
  const normalizedType = String(targetType || '').trim().toLowerCase()
  if (!normalizedType) return null
  const matches = (modules || []).filter((item) =>
    item.module_type === 'output' &&
    item.preset_kind === 'output' &&
    String(item.preset_key || '').trim().toLowerCase() === normalizedType &&
    !!item.is_builtin
  )
  return matches.find((item) => item.fluent_type === fluentType) ||
    matches.find((item) => item.fluent_type === 'shared') ||
    matches[0] ||
    null
}

function normalizeImportedOutputRenderVariables(targetType, fluentType, rawVariables) {
  const normalizedType = String(targetType || '').trim().toLowerCase()
  const normalizedRuntime = String(fluentType || '').trim().toLowerCase()
  const variables = { ...(rawVariables || {}) }

  if (variables.http_passwd !== undefined && variables.http_password === undefined) {
    variables.http_password = variables.http_passwd
  }
  if (variables.http_password !== undefined && variables.http_passwd === undefined) {
    variables.http_passwd = variables.http_password
  }
  if (variables['tls.verify'] !== undefined && variables.tls_verify === undefined) {
    variables.tls_verify = variables['tls.verify']
  }
  if (variables.tls_verify !== undefined && variables['tls.verify'] === undefined) {
    variables['tls.verify'] = variables.tls_verify
  }

  if (normalizedType === 'opensearch' && normalizedRuntime === 'fluentd') {
    if (variables.http_user !== undefined && variables.user === undefined) {
      variables.user = variables.http_user
    }
    if (variables.http_passwd !== undefined && variables.password === undefined) {
      variables.password = variables.http_passwd
    }
    if (variables.http_password !== undefined && variables.password === undefined) {
      variables.password = variables.http_password
    }
    if (variables.tls_verify !== undefined && variables.ssl_verify === undefined) {
      variables.ssl_verify = variables.tls_verify
    }
    if (variables.logstash_prefix !== undefined && variables.index_name === undefined) {
      variables.index_name = variables.logstash_prefix
    }
  }

  if (normalizedType === 'http' && normalizedRuntime === 'fluentd' && variables.endpoint === undefined) {
    if (variables.uri && String(variables.uri).startsWith('http')) {
      variables.endpoint = variables.uri
    } else if (variables.host) {
      const portPart = variables.port ? `:${variables.port}` : ''
      const uriPart = variables.uri ? String(variables.uri) : ''
      variables.endpoint = `http://${variables.host}${portPart}${uriPart}`
    }
  }

  return variables
}

function buildImportedOutputRenderVariables(module, target, fluentType) {
  const targetSettings = parseVariablesMap(target?.settings)
  const instanceVariables = parseVariablesMap(module?.variables)
  return normalizeImportedOutputRenderVariables(target?.target_type, fluentType, {
    ...targetSettings,
    ...instanceVariables,
    output_target_name: target?.name || '',
    output_target_type: target?.target_type || '',
  })
}

async function listAllModules() {
  const pageSize = 100
  const collected = []
  let page = 1
  let total = 0

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

function inferImportedOutputTargetType(module) {
  const plugin = String(module?.output_target_type || module?.detected_plugin || '').trim().toLowerCase()
  if (plugin === 'es' || plugin === 'opensearch' || plugin === 'elasticsearch') return 'opensearch'
  if (plugin === 'loki') return 'loki'
  if (plugin === 'kafka' || plugin === 'rdkafka') return 'kafka'
  if (plugin === 'http') return 'http'
  if (plugin === 's3') return 's3'
  if (plugin === 'stdout') return 'stdout'

  const content = String(module?.content || '').toLowerCase()
  if (content.includes('opensearch') || content.includes('elasticsearch')) return 'opensearch'
  if (content.includes('loki')) return 'loki'
  if (content.includes('kafka')) return 'kafka'
  if (content.includes('stdout')) return 'stdout'
  if (content.includes('s3')) return 's3'
  if (content.includes('http')) return 'http'
  return 'custom'
}

function buildImportedOutputTargetDraft(module) {
  const targetType = inferImportedOutputTargetType(module)
  const rawSettings = {
    ...parseVariablesMap(module?.variables),
  }
  if (targetType === 'custom') {
    rawSettings.plugin = module?.detected_plugin || rawSettings.plugin || 'custom_output'
  }
  const settings = pickImportedOutputTargetSettings(rawSettings, targetType)

  let endpoint = ''
  if (rawSettings.uri) {
    endpoint = String(rawSettings.uri)
  } else if (rawSettings.endpoint) {
    endpoint = String(rawSettings.endpoint)
  } else if (rawSettings.host && rawSettings.port) {
    endpoint = `${rawSettings.host}:${rawSettings.port}`
  } else if (rawSettings.host) {
    endpoint = String(rawSettings.host)
  } else if (targetType === 'kafka' && rawSettings.brokers) {
    endpoint = String(rawSettings.brokers)
  } else if (targetType === 's3' && rawSettings.bucket) {
    const path = rawSettings.path ? `/${String(rawSettings.path).replace(/^\/+/, '')}` : ''
    endpoint = `s3://${rawSettings.bucket}${path}`
  } else if (targetType === 'stdout') {
    endpoint = 'stdout'
  }

  return {
    target_type: targetType,
    endpoint,
    settings: JSON.stringify(settings, null, 2),
  }
}

function importActionLabel(action) {
  if (action === 'reuse_existing') return t('configs_page.import_action_reuse')
  return t('configs_page.import_action_create')
}

function setImportedModuleAction(module, action) {
  if (!module) return
  if (module.module_type === 'output') return
  if (action === 'reuse_existing' && !module.existing_module_id) return
  module.import_action = action
}

function setAllImportedModuleActions(action) {
  if (!importedConfigResult.value?.modules?.length) return
  for (const module of importedConfigResult.value.modules) {
    if (module.module_type === 'output') continue
    if (action === 'reuse_existing' && !module.existing_module_id) {
      module.import_action = 'create_new'
      continue
    }
    module.import_action = action
  }
}

function importActionBadgeClass(action) {
  return action === 'reuse_existing'
    ? 'bg-success-subtle text-success-emphasis'
    : 'text-bg-light'
}

function importValidationBadgeClass(verdict) {
  if (verdict === 'equivalent') return 'bg-success-subtle text-success-emphasis'
  if (verdict === 'mostly_equivalent') return 'bg-warning-subtle text-warning-emphasis'
  return 'bg-danger-subtle text-danger-emphasis'
}

function importValidationLabel(verdict) {
  if (verdict === 'equivalent') return t('configs_page.import_validation_equivalent')
  if (verdict === 'mostly_equivalent') return t('configs_page.import_validation_mostly_equivalent')
  return t('configs_page.import_validation_needs_review')
}

function importDestinationMatchLabel(matchType) {
  if (matchType === 'exact') return t('configs_page.import_destination_match_exact')
  if (matchType === 'created') return t('configs_page.import_destination_match_created')
  return t('configs_page.import_destination_match_type')
}

const wizardFlowPathLabel = computed(() =>
  wizardFlowSummary.value.path.length ? wizardFlowSummary.value.path.join(' -> ') : t('configs_page.no_solution_path')
)
const wizardProcessorChainLabel = computed(() =>
  wizardFlowSummary.value.processors.length ? wizardFlowSummary.value.processors.join(' -> ') : t('configs_page.no_processors')
)
const wizardSourcePresetChips = computed(() => wizardInputPresetsSelected.value.map((item) => item.name))
const wizardDestinationChips = computed(() => wizardFlowSummary.value.destinationChips || [])
const previewFlowPathLabel = computed(() =>
  previewFlowSummary.value.path.length ? previewFlowSummary.value.path.join(' -> ') : t('configs_page.no_solution_path')
)
const previewDestinationChips = computed(() => previewFlowSummary.value.destinationChips || [])
const importFlowPathLabel = computed(() =>
  importedConfigResult.value?.flow_path?.length ? importedConfigResult.value.flow_path.join(' -> ') : t('configs_page.no_solution_path')
)
const importSemanticChangeCount = computed(() => importedConfigResult.value?.validation?.semantic_diff?.changes?.length || 0)
const importReusableMatchCount = computed(() =>
  importedConfigResult.value?.modules?.filter((module) => module.module_type !== 'output' && module.existing_module_id).length || 0
)
const importDestinationMatchCount = computed(() => importedConfigResult.value?.destinations?.length || 0)
const importReuseDecisionCount = computed(() =>
  importedConfigResult.value?.modules?.filter((module) => module.module_type !== 'output' && module.import_action === 'reuse_existing').length || 0
)
const importCreateDecisionCount = computed(() =>
  importedConfigResult.value?.modules?.filter((module) => module.module_type !== 'output' && module.import_action !== 'reuse_existing').length || 0
)

function getErrorMessage(error) {
  return error?.response?.data?.user_message || error?.response?.data?.error || error?.message || t('common.request_failed')
}

function getProviderErrorMessage(error) {
  return error?.response?.data?.provider_message || ''
}

function clearAIAssistantFeedback() {
  aiAssistantFeedback.type = ''
  aiAssistantFeedback.message = ''
  aiAssistantFeedback.detail = ''
  aiAssistantFeedback.provider = ''
  aiAssistantFeedback.providerDetail = ''
}

function setAIAssistantFeedback(type, message, detail = '', provider = '', providerDetail = '') {
  aiAssistantFeedback.type = type
  aiAssistantFeedback.message = message
  aiAssistantFeedback.detail = detail
  aiAssistantFeedback.provider = provider
  aiAssistantFeedback.providerDetail = providerDetail
}

function createAIDraftState() {
  return {
    active: false,
    provider: '',
    accountName: '',
    summary: '',
    steps: [],
    notes: [],
    reviewItems: [],
    confirmationItems: [],
  }
}

function resetAIDraftState(state) {
  state.active = false
  state.provider = ''
  state.accountName = ''
  state.summary = ''
  state.steps = []
  state.notes = []
  state.reviewItems = []
  state.confirmationItems = []
}

function buildDraftConfirmationItems(labels) {
  return labels.map((label, index) => ({
    key: `confirm-${index}`,
    label,
    checked: false,
  }))
}

function activateAIDraftState(state, result, reviewItems, confirmationLabels) {
  state.active = true
  state.provider = result?.provider || ''
  state.accountName = result?.account_name || ''
  state.summary = result?.summary || ''
  state.steps = Array.isArray(result?.assembly_steps) ? result.assembly_steps : []
  state.notes = Array.isArray(result?.notes) ? result.notes : []
  state.reviewItems = reviewItems
  state.confirmationItems = buildDraftConfirmationItems(confirmationLabels)
}

function areDraftConfirmationsComplete(state) {
  return !state.active || state.confirmationItems.every((item) => item.checked)
}

function normalizeName(value) {
  return String(value || '').trim().toLowerCase()
}

function generateUniqueDraftName(baseName, existingNames, fallbackPrefix = 'ai-draft') {
  const normalizedExisting = new Set(existingNames.map((item) => normalizeName(item)).filter(Boolean))
  const seed = String(baseName || '').trim() || fallbackPrefix
  if (!normalizedExisting.has(normalizeName(seed))) {
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (normalizedExisting.has(normalizeName(candidate))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  return candidate
}

function countNonEmptyLines(content) {
  return String(content || '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean).length
}

function extractTemplatePlaceholders(content) {
  const matches = String(content || '').match(/{{\s*\.[^}]+}}/g)
  return matches || []
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort()
}

function diffKeys(current, previous) {
  const currentSet = new Set(current)
  const previousSet = new Set(previous)
  return {
    added: current.filter((item) => !previousSet.has(item)),
    removed: previous.filter((item) => !currentSet.has(item)),
  }
}

function summarizeChangedKeys(added, removed) {
  const parts = []
  if (added.length) {
    parts.push(t('configs_page.ai_draft_diff_added').replace('{items}', added.join(', ')))
  }
  if (removed.length) {
    parts.push(t('configs_page.ai_draft_diff_removed').replace('{items}', removed.join(', ')))
  }
  return parts.join('；')
}

function buildModuleDraftComparison() {
  if (!aiModuleDraftState.active) return null

  const variableKeys = uniqueSorted(Object.keys(parseVariablesMap(moduleForm.variables)))
  const placeholderKeys = uniqueSorted(extractTemplatePlaceholders(moduleForm.content).map((item) => item.replace(/[{}\s.]/g, '')))
  const existing = modules.value.find((item) =>
    normalizeName(item.name) === normalizeName(moduleForm.name) &&
    item.module_type === moduleForm.module_type &&
    item.fluent_type === moduleForm.fluent_type
  )

  let identityMessage = t('configs_page.ai_draft_diff_new_asset')
  let existingName = ''
  let changeMessage = ''
  let changeDetail = ''
  let suggestedName = ''
  const hasConflict = !!existing

  if (existing) {
    existingName = `${existing.name} / ${existing.module_type} / ${runtimeLabel(existing.fluent_type)}`
    identityMessage = t('configs_page.ai_draft_diff_existing_asset')
    suggestedName = generateUniqueDraftName(
      moduleForm.name,
      modules.value
        .filter((item) => item.module_type === moduleForm.module_type && item.fluent_type === moduleForm.fluent_type)
        .map((item) => item.name),
      `ai-${moduleForm.module_type || 'module'}`
    )
    const previousKeys = uniqueSorted(Object.keys(parseVariablesMap(existing.variables)))
    const { added, removed } = diffKeys(variableKeys, previousKeys)
    changeMessage = existing.content === moduleForm.content
      ? t('configs_page.ai_draft_diff_content_same')
      : t('configs_page.ai_draft_diff_content_changed')
    changeDetail = summarizeChangedKeys(added, removed)
    if (!changeDetail) {
      changeDetail = t('configs_page.ai_draft_diff_existing_review')
    }
  }

  return {
    existingAsset: existing || null,
    hasConflict,
    identityMessage,
    existingName,
    suggestedName,
    variableCount: variableKeys.length,
    lineCount: countNonEmptyLines(moduleForm.content),
    placeholderCount: placeholderKeys.length,
    changeMessage,
    changeDetail,
  }
}

function buildTemplateDraftComparison() {
  if (!aiTemplateDraftState.active) return null

  const placeholderKeys = uniqueSorted(extractTemplatePlaceholders(templateForm.content).map((item) => item.replace(/[{}\s.]/g, '')))
  const existing = templates.value.find((item) => normalizeName(item.name) === normalizeName(templateForm.name))

  let identityMessage = t('configs_page.ai_draft_diff_new_asset')
  let existingName = ''
  let changeMessage = ''
  let changeDetail = ''
  let suggestedName = ''
  const hasConflict = !!existing

  if (existing) {
    existingName = `${existing.name} / ${runtimeLabel(existing.fluent_type)}`
    identityMessage = t('configs_page.ai_draft_diff_existing_template')
    suggestedName = generateUniqueDraftName(
      templateForm.name,
      templates.value.map((item) => item.name),
      `ai-${templateForm.fluent_type || 'template'}-template`
    )
    changeMessage = existing.content === templateForm.content
      ? t('configs_page.ai_draft_diff_content_same')
      : t('configs_page.ai_draft_diff_content_changed')

    const currentKeys = uniqueSorted(placeholderKeys)
    const previousKeys = uniqueSorted(extractTemplatePlaceholders(existing.content).map((item) => item.replace(/[{}\s.]/g, '')))
    const { added, removed } = diffKeys(currentKeys, previousKeys)
    changeDetail = summarizeChangedKeys(added, removed)
    if (!changeDetail) {
      const lineDelta = countNonEmptyLines(templateForm.content) - countNonEmptyLines(existing.content)
      if (lineDelta !== 0) {
        changeDetail = t('configs_page.ai_draft_diff_line_delta').replace('{count}', String(lineDelta))
      } else {
        changeDetail = t('configs_page.ai_draft_diff_existing_review')
      }
    }
  }

  return {
    existingAsset: existing || null,
    hasConflict,
    identityMessage,
    existingName,
    suggestedName,
    lineCount: countNonEmptyLines(templateForm.content),
    placeholderCount: placeholderKeys.length,
    changeMessage,
    changeDetail,
  }
}

function parseVariablesMap(value) {
  if (!value) return {}
  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

function parseVariablesMapStrict(value) {
  const raw = String(value || '').trim()
  if (!raw) return {}
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(t('configs_page.variable_json_object_required'))
  }
  return parsed
}

function inferVariableKind(value) {
  if (typeof value === 'boolean') return 'boolean'
  if (value && typeof value === 'object') return 'json'
  return 'text'
}

function stringifyVariableValue(value) {
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  if (value && typeof value === 'object') return JSON.stringify(value, null, 2)
  if (value === undefined || value === null) return ''
  return String(value)
}

function buildWizardVariableDraft(defaults) {
  const draft = {}
  for (const [key, value] of Object.entries(defaults || {})) {
    draft[key] = stringifyVariableValue(value)
  }
  return draft
}

function createWizardInstanceID(prefix) {
  wizardSequence += 1
  return `${prefix}-${wizardSequence}`
}

function createWizardPipeline() {
  return {
    id: createWizardInstanceID('wizard-pipeline'),
    name: '',
    input: null,
    filters: [],
    routes: [],
    outputs: [],
  }
}

function paginateItems(items, page, pageSize = wizardStagePageSize) {
  const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
  const currentPage = Math.min(Math.max(Number(page) || 1, 1), totalPages)
  const start = (currentPage - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    totalPages,
    currentPage,
    total: items.length,
  }
}

function moduleVariablesForWizard(module) {
  return parseVariablesMap(module?.variables)
}

function parserNamesProvidedByModule(module) {
  const names = []
  let parserSection = false
  for (const rawLine of String(module?.content || '').split('\n')) {
    const trimmed = rawLine.trim()
    if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith(';')) continue
    if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
      const section = trimmed.slice(1, -1).trim().toUpperCase()
      parserSection = section === 'PARSER' || section === 'MULTILINE_PARSER'
      continue
    }
    if (!parserSection) continue
    const parts = trimmed.split(/\s+/)
    if (parts.length < 2 || parts[0].toLowerCase() !== 'name') continue
    const name = trimmed.slice(parts[0].length).trim()
    if (name && !name.includes('{{')) {
      names.push(name)
    }
  }
  return uniqueSorted(names)
}

function parserReferencesForInputModule(module, draft = null) {
  const names = []
  const merged = {
    ...moduleVariablesForWizard(module),
    ...(draft || {}),
  }
  for (const [key, value] of Object.entries(merged)) {
    const normalizedKey = String(key || '').trim().toLowerCase()
    if (!(normalizedKey === 'parser' || normalizedKey === 'parser_firstline' || normalizedKey === 'multiline.parser' || normalizedKey.startsWith('parser_'))) {
      continue
    }
    const name = String(value || '').trim()
    if (name && !name.includes('{{')) {
      names.push(name)
    }
  }
  return uniqueSorted(names)
}

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
  const merged = {
    ...moduleVariablesForWizard(module),
    ...(pipeline?.input?.variables || {}),
  }
  return String(merged.tag || '').trim()
}

function wizardPipelineModuleDefaults(module, pipeline, extraDefaults = {}) {
  const defaults = {
    ...moduleVariablesForWizard(module),
    ...(extraDefaults || {}),
  }
  const inputTag = wizardPipelineInputTag(pipeline)
  if (inputTag && Object.prototype.hasOwnProperty.call(defaults, 'match')) {
    defaults.match = inputTag
  }
  return defaults
}

function shouldAutoSyncWizardMatch(currentValue, previousTag) {
  const current = String(currentValue ?? '').trim()
  return !current || current === '*' || current === '**' || (!!previousTag && current === previousTag)
}

function buildWizardPipelineDraft(key, module, pipeline, existingDraft = null, extraDefaults = {}, previousTag = '') {
  const defaults = wizardPipelineModuleDefaults(module, pipeline, extraDefaults)
  if (!existingDraft || !Object.keys(existingDraft).length) {
    return buildWizardVariableDraft(defaults)
  }
  const nextDraft = { ...existingDraft }
  if (Object.prototype.hasOwnProperty.call(defaults, 'match') && shouldAutoSyncWizardMatch(nextDraft.match, previousTag)) {
    nextDraft.match = stringifyVariableValue(defaults.match)
  }
  return nextDraft
}

function normalizeWizardDraftValues(draft, defaults) {
  const normalized = {}
  const mergedDefaults = { ...(defaults || {}) }
  const keys = new Set([...Object.keys(mergedDefaults), ...Object.keys(draft || {})])
  for (const key of keys) {
    normalized[key] = normalizeWizardVariableValue(draft?.[key], inferVariableKind(mergedDefaults[key]))
  }
  return normalized
}

function buildWizardModuleGroup(key, title, subtitle, module, model, extraDefaults = {}, section = null) {
  if (!module || !model) return null
  const defaults = {
    ...moduleVariablesForWizard(module),
    ...(extraDefaults || {}),
  }
  const fields = Object.entries(defaults).map(([fieldKey, value]) => ({
    key: fieldKey,
    kind: inferVariableKind(value),
    description: module.description || '',
  }))
  if (!fields.length) return null
  return {
    key,
    title,
    subtitle,
    sectionKey: section?.key || 'default',
    sectionTitle: section?.title || '',
    sectionKind: section?.kind || 'default',
    sectionRef: section?.ref || '',
    fields,
    model,
  }
}

function inferModuleVariableRowType(value) {
  if (typeof value === 'boolean') return 'boolean'
  if (typeof value === 'number') return 'number'
  if (value && typeof value === 'object') return 'json'
  return 'string'
}

function buildModuleVariableRows(raw) {
  const parsed = parseVariablesMapStrict(raw)
  const entries = Object.entries(parsed)
  if (!entries.length) {
    return [{ key: '', type: 'string', value: '' }]
  }
  return entries.map(([key, value]) => ({
    key,
    type: inferModuleVariableRowType(value),
    value: stringifyVariableValue(value),
  }))
}

function parseModuleVariableRowValue(row) {
  if (row.type === 'boolean') {
    return row.value === true || row.value === 'true'
  }
  if (row.type === 'number') {
    const trimmed = String(row.value ?? '').trim()
    if (trimmed === '') return 0
    const parsed = Number(trimmed)
    if (Number.isNaN(parsed)) {
      throw new Error('invalid number')
    }
    return parsed
  }
  if (row.type === 'json') {
    const trimmed = String(row.value ?? '').trim()
    if (!trimmed) return {}
    return JSON.parse(trimmed)
  }
  return String(row.value ?? '')
}

function syncModuleVariablesFromRows() {
  try {
    const payload = {}
    for (const row of moduleVariableRows.value) {
      const key = String(row.key || '').trim()
      if (!key) continue
      payload[key] = parseModuleVariableRowValue(row)
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

function removeModuleVariableRow(index) {
  moduleVariableRows.value = moduleVariableRows.value.filter((_, rowIndex) => rowIndex !== index)
  if (!moduleVariableRows.value.length) {
    moduleVariableRows.value = [{ key: '', type: 'string', value: '' }]
  }
  syncModuleVariablesFromRows()
}

function setModuleVariablesMode(mode) {
  if (mode === moduleVariablesMode.value) return

  if (mode === 'json') {
    if (!syncModuleVariablesFromRows()) {
      return
    }
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

function normalizeWizardVariableValue(value, kind) {
  if (kind === 'boolean') {
    return value === true || value === 'true'
  }
  if (kind === 'json') {
    if (!String(value || '').trim()) return {}
    try {
      return JSON.parse(value)
    } catch {
      return value
    }
  }
  return value
}

function mergeWizardModuleVariableValues(moduleId, nextValues) {
  if (!moduleId) return
  const existing = { ...(wizardModuleVariableValues.value[moduleId] || {}) }
  for (const [key, value] of Object.entries(nextValues || {})) {
    existing[key] = stringifyVariableValue(value)
  }
  wizardModuleVariableValues.value = {
    ...wizardModuleVariableValues.value,
    [moduleId]: existing,
  }
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

function matchingOutputModuleForTarget(target, eligibleModules, fluentType) {
  if (!target) return null
  const matches = eligibleModules.filter((item) =>
    item.module_type === 'output' &&
    item.preset_kind === 'output' &&
    item.preset_key === target.target_type
  )
  return matches.find((item) => item.fluent_type === fluentType) || matches.find((item) => item.fluent_type === 'shared') || null
}

function wizardPipelineDisplayName(pipelineOrCard, index = 0) {
  const pipelineName = String(pipelineOrCard?.name || '').trim()
  return pipelineName || t('configs_page.wizard_pipeline_fallback').replace('{index}', String(index + 1))
}

function changeWizardStagePage(stage, nextPage) {
  const metaMap = {
    service: wizardPagedServiceModules.value,
    parser: wizardPagedParserModules.value,
    input: wizardPagedInputModules.value,
    filter: wizardPagedFilterModules.value,
    output: wizardPagedOutputTargets.value,
  }
  const meta = metaMap[stage]
  if (!meta) return
  const normalized = Math.min(Math.max(nextPage, 1), meta.totalPages)
  wizardStagePages[stage] = normalized
}

function resetWizardStagePage(stage) {
  wizardStagePages[stage] = 1
}

function ensureWizardModuleDraft(key, module, existingDraft = null, extraDefaults = {}) {
  const defaults = {
    ...moduleVariablesForWizard(module),
    ...(extraDefaults || {}),
  }
  return existingDraft && Object.keys(existingDraft).length ? { ...existingDraft } : buildWizardVariableDraft(defaults)
}

function selectWizardServiceModule(moduleId) {
  if (wizardServiceModuleId.value === moduleId) {
    wizardServiceModuleId.value = null
    return
  }
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
    input: pipeline.input
      ? {
          id: createWizardInstanceID('wizard-input'),
          module_id: pipeline.input.module_id,
          variables: { ...pipeline.input.variables },
        }
      : null,
    filters: pipeline.filters.map((instance) => ({
      id: createWizardInstanceID('wizard-filter'),
      module_id: instance.module_id,
      variables: { ...instance.variables },
    })),
    routes: (pipeline.routes || []).map((instance) => ({
      id: createWizardInstanceID('wizard-route'),
      module_id: instance.module_id,
      variables: { ...instance.variables },
    })),
    outputs: pipeline.outputs.map((instance) => ({
      id: createWizardInstanceID('wizard-output'),
      target_id: instance.target_id,
      variables: { ...instance.variables },
    })),
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

function selectWizardPipeline(pipelineId) {
  activeWizardPipelineId.value = pipelineId
}

function updateWizardPipeline(pipelineId, updater) {
  wizardPipelines.value = wizardPipelines.value.map((pipeline) => {
    if (pipeline.id !== pipelineId) return pipeline
    return updater(pipeline)
  })
}

function setWizardPipelineInput(pipelineId, moduleId) {
  const module = wizardEligibleModules.value.find((item) => item.id === moduleId && item.module_type === 'input')
  if (!module) return
  const currentPipeline = wizardPipelines.value.find((item) => item.id === pipelineId) || null
  const previousTag = wizardPipelineInputTag(currentPipeline)
  const draft = ensureWizardModuleDraft(
    `input:${module.id}`,
    module,
    currentPipeline?.input?.module_id === module.id ? currentPipeline?.input?.variables : null
  )
  const nextInput = {
    id: createWizardInstanceID('wizard-input'),
    module_id: module.id,
    variables: draft,
  }
  updateWizardPipeline(pipelineId, (pipeline) => {
    const nextPipeline = {
      ...pipeline,
      input: nextInput,
    }
    return {
      ...nextPipeline,
      filters: pipeline.filters.map((instance) => {
        const filterModule = wizardEligibleModules.value.find((item) => item.id === instance.module_id && item.module_type === 'filter')
        return {
          ...instance,
          variables: buildWizardPipelineDraft(`filter:${instance.module_id}`, filterModule, nextPipeline, instance.variables, {}, previousTag),
        }
      }),
      outputs: pipeline.outputs.map((instance) => {
        const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
        return {
          ...instance,
          variables: buildWizardOutputDraft(target, wizardForm.fluent_type, nextPipeline, instance.variables, previousTag),
        }
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
    filters: [
      ...pipeline.filters,
      {
        id: createWizardInstanceID('wizard-filter'),
        module_id: module.id,
        variables: buildWizardPipelineDraft(`filter:${module.id}`, module, pipeline),
      },
    ],
  }))
}

function removeWizardFilter(pipelineId, instanceId) {
  updateWizardPipeline(pipelineId, (pipeline) => ({
    ...pipeline,
    filters: pipeline.filters.filter((item) => item.id !== instanceId),
  }))
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
    routes: [
      ...(pipeline.routes || []),
      {
        id: createWizardInstanceID('wizard-route'),
        module_id: module.id,
        variables: buildWizardPipelineDraft(`route:${module.id}`, module, pipeline),
      },
    ],
  }))
}

function removeWizardRoute(pipelineId, instanceId) {
  updateWizardPipeline(pipelineId, (pipeline) => ({
    ...pipeline,
    routes: (pipeline.routes || []).filter((item) => item.id !== instanceId),
  }))
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
    outputs: [
      ...pipeline.outputs,
      {
        id: createWizardInstanceID('wizard-output'),
        target_id: target.id,
        variables: buildWizardOutputDraft(target, wizardForm.fluent_type, pipeline),
      },
    ],
  }))
}

function removeWizardOutput(pipelineId, instanceId) {
  updateWizardPipeline(pipelineId, (pipeline) => ({
    ...pipeline,
    outputs: pipeline.outputs.filter((item) => item.id !== instanceId),
  }))
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
        return {
          ...instance,
          variables: buildWizardOutputDraft(target, wizardForm.fluent_type, pipeline, instance.variables),
        }
      }),
  }))

  ensureWizardBaselineModules()
}

function removeWizardModuleReferences(moduleId) {
  if (wizardServiceModuleId.value === moduleId) {
    wizardServiceModuleId.value = null
  }
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
  if (wizardServiceModuleId.value && deleted.has(wizardServiceModuleId.value)) {
    wizardServiceModuleId.value = null
  }
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
  const currentInputIDs = wizardSelectedModules.value
    .filter((item) => item.module_type === 'input')
    .map((item) => item.id)

  if (selectedWizardModuleIds.value.includes(module.id)) {
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== module.id)
    return
  }
  selectedWizardModuleIds.value = [
    ...selectedWizardModuleIds.value.filter((id) => !currentInputIDs.includes(id)),
    module.id,
  ]
  mergeWizardModuleVariableValues(module.id, parseVariablesMap(module.variables))
}

function toggleWizardOutputTarget(targetId) {
  if (selectedWizardOutputTargetIds.value.includes(targetId)) {
    selectedWizardOutputTargetIds.value = selectedWizardOutputTargetIds.value.filter((id) => id !== targetId)
    return
  }
  selectedWizardOutputTargetIds.value = [...selectedWizardOutputTargetIds.value, targetId]
}

function togglePreviewOutputTarget(targetId) {
  if (selectedPreviewOutputTargetIds.value.includes(targetId)) {
    selectedPreviewOutputTargetIds.value = selectedPreviewOutputTargetIds.value.filter((id) => id !== targetId)
    return
  }
  selectedPreviewOutputTargetIds.value = [...selectedPreviewOutputTargetIds.value, targetId]
}

function buildOutputTargetModuleRefs(targets, eligibleModules, fluentType) {
  return targets
    .map((target) => {
      const outputModule = matchingOutputModuleForTarget(target, eligibleModules, fluentType)
      if (!outputModule) return null
      const settings = parseVariablesMap(target.settings)
      const variables = {
        ...settings,
        output_target_name: target.name,
        output_target_type: target.target_type,
      }
      return {
        module_id: outputModule.id,
        variables: JSON.stringify(variables, null, 2),
      }
    })
    .filter(Boolean)
}

function buildPreviewSelectedModuleRefs() {
  return selectedPreviewModuleIds.value.map((moduleId) => {
    const variables = previewModuleVariables.value[moduleId]
    if (!variables || variables === '{}') {
      return { module_id: moduleId }
    }
    return {
      module_id: moduleId,
      variables,
    }
  })
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

async function loadTemplates() {
  const { data } = await getTemplates()
  templates.value = data.data || []
}

async function loadPipelines() {
  const { data } = await getConfigPipelines()
  pipelines.value = data.data || []
}

function openCreatePipeline() {
  pipelineFormInitializing.value = true
  editingPipelineId.value = null
  pipelineForm.name = ''
  pipelineForm.description = ''
  pipelineForm.fluent_type = 'fluentbit'
  pipelineForm.input_module_id = null
  pipelineForm.filter_module_ids = []
  pipelineForm.output_target_ids = []
  pipelineFilterPickerValue.value = ''
  pipelineFormInitializing.value = false
  ensurePipelineModal()
  pipelineModal.show()
}

function openEditPipeline(pipeline) {
  pipelineFormInitializing.value = true
  editingPipelineId.value = pipeline.id
  pipelineForm.name = pipeline.name
  pipelineForm.description = pipeline.description || ''
  pipelineForm.fluent_type = pipeline.fluent_type || 'fluentbit'
  pipelineForm.input_module_id = pipeline.input_module_id || null
  pipelineForm.filter_module_ids = (pipeline.filter_modules || []).map((m) => m.id)
  pipelineForm.output_target_ids = (pipeline.output_targets || []).map((t) => t.id)
  pipelineFilterPickerValue.value = ''
  pipelineFormInitializing.value = false
  ensurePipelineModal()
  pipelineModal.show()
}

async function savePipeline() {
  if (!pipelineForm.name.trim()) {
    alert('Name is required')
    return
  }
  try {
    const payload = {
      name: pipelineForm.name.trim(),
      description: pipelineForm.description.trim(),
      fluent_type: pipelineForm.fluent_type,
      input_module_id: pipelineForm.input_module_id,
      filter_module_ids: pipelineForm.filter_module_ids,
      output_target_ids: pipelineForm.output_target_ids,
    }
    if (editingPipelineId.value) {
      await updateConfigPipeline(editingPipelineId.value, payload)
    } else {
      await createConfigPipeline(payload)
    }
    await loadPipelines()
    pipelineModal.hide()
  } catch (e) {
    alert(`${t('common.request_failed')}: ${e?.response?.data?.error || e?.message || ''}`)
  }
}

async function handleDeletePipeline(pipeline) {
  if (!confirm(t('configs_page.pipeline_delete_confirm').replace('{name}', pipeline.name))) return
  try {
    await deleteConfigPipeline(pipeline.id)
    await loadPipelines()
  } catch (e) {
    alert(`${t('common.request_failed')}: ${e?.response?.data?.error || e?.message || ''}`)
  }
}

function addPipelineFilterModule(moduleId) {
  if (moduleId && !pipelineForm.filter_module_ids.includes(moduleId)) {
    pipelineForm.filter_module_ids.push(moduleId)
  }
}

function removePipelineFilterModule(index) {
  pipelineForm.filter_module_ids.splice(index, 1)
}

function movePipelineFilterModule(index, direction) {
  const arr = pipelineForm.filter_module_ids
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= arr.length) return
  const tmp = arr[index]
  arr[index] = arr[newIndex]
  arr[newIndex] = tmp
}

function togglePipelineOutputTarget(targetId) {
  const idx = pipelineForm.output_target_ids.indexOf(targetId)
  if (idx === -1) {
    pipelineForm.output_target_ids.push(targetId)
  } else {
    pipelineForm.output_target_ids.splice(idx, 1)
  }
}

function addWizardPipelineFromSaved(pipelineId) {
  const saved = pipelines.value.find((p) => p.id === Number(pipelineId))
  if (!saved) return
  const newPipeline = createWizardPipeline()
  newPipeline.input = saved.input_module_id ? { id: createWizardInstanceID('wizard-input'), module_id: saved.input_module_id } : null
  newPipeline.filters = (saved.filter_modules || []).map((m) => ({ id: createWizardInstanceID('wizard-filter'), module_id: m.id }))
  newPipeline.outputs = (saved.output_targets || []).map((t) => ({ id: createWizardInstanceID('wizard-output'), target_id: t.id }))
  wizardPipelines.value.push(newPipeline)
  activeWizardPipelineId.value = newPipeline.id
}

async function loadModules() {
  modules.value = await listAllModules()
  const deletableIDs = new Set(modules.value.filter((item) => !item.is_builtin).map((item) => item.id))
  selectedModuleIds.value = selectedModuleIds.value.filter((id) => deletableIDs.has(id))
}

async function loadModuleTable() {
  moduleTableLoading.value = true
  try {
    const params = {
      page: moduleQuery.page,
      page_size: moduleQuery.page_size,
    }
    if (moduleQuery.search.trim()) {
      params.search = moduleQuery.search.trim()
    }
    if (moduleQuery.fluent_type) {
      params.fluent_type = moduleQuery.fluent_type
    }
    if (moduleQuery.module_type) {
      params.module_type = moduleQuery.module_type
    }

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

function applyModuleQuery() {
  moduleQuery.page = 1
  loadModuleTable()
}

function resetModuleQuery() {
  moduleQuery.search = ''
  moduleQuery.fluent_type = ''
  moduleQuery.module_type = ''
  moduleQuery.page = 1
  moduleQuery.page_size = 20
  loadModuleTable()
}

function changeModulePage(nextPage) {
  if (nextPage < 1 || nextPage > moduleTableTotalPages.value || nextPage === moduleQuery.page) {
    return
  }
  moduleQuery.page = nextPage
  loadModuleTable()
}

function setModuleTypeFilter(type = '') {
  moduleQuery.module_type = moduleQuery.module_type === type ? '' : type
  moduleQuery.page = 1
  loadModuleTable()
}

function toggleModuleSelection(module) {
  if (!module || module.is_builtin) return
  if (selectedModuleIds.value.includes(module.id)) {
    selectedModuleIds.value = selectedModuleIds.value.filter((id) => id !== module.id)
    return
  }
  selectedModuleIds.value = [...selectedModuleIds.value, module.id]
}

function toggleSelectAllVisibleModules() {
  const visibleIDs = visibleDeletableModules.value.map((item) => item.id)
  if (!visibleIDs.length) return

  if (allVisibleDeletableModulesSelected.value) {
    selectedModuleIds.value = selectedModuleIds.value.filter((id) => !visibleIDs.includes(id))
    return
  }

  selectedModuleIds.value = Array.from(new Set([...selectedModuleIds.value, ...visibleIDs]))
}

async function loadOutputTargets() {
  try {
    outputTargets.value = await getOutputTargets()
  } catch {
    outputTargets.value = []
  }
}

function ensureTemplateModal() {
  if (!templateModal) {
    templateModal = new window.bootstrap.Modal(document.getElementById('templateModal'))
  }
}

function ensureModuleModal() {
  if (!moduleModal) {
    moduleModal = new window.bootstrap.Modal(document.getElementById('moduleModal'))
  }
}

function ensureModuleVersionsModal() {
  if (!moduleVersionsModal) {
    moduleVersionsModal = new window.bootstrap.Modal(document.getElementById('moduleVersionsModal'))
  }
}

function ensurePipelineModal() {
  if (!pipelineModal) {
    pipelineModal = new window.bootstrap.Modal(document.getElementById('pipelineModal'))
  }
}

function resetTemplateForm() {
  templateForm.name = ''
  templateForm.description = ''
  templateForm.fluent_type = 'fluentbit'
  templateForm.content = ''
  resetAIDraftState(aiTemplateDraftState)
}

function resetModuleForm() {
  editingModuleId.value = null
  moduleForm.name = ''
  moduleForm.description = ''
  moduleForm.module_type = 'input'
  moduleForm.fluent_type = 'fluentbit'
  moduleForm.content = ''
  moduleForm.content_fluentd = ''
  moduleForm.variables = ''
  moduleForm.is_builtin = false
  moduleVariablesMode.value = 'form'
  moduleVariableRows.value = [{ key: '', type: 'string', value: '' }]
  moduleVariablesFormError.value = ''
  resetAIDraftState(aiModuleDraftState)
}

function resetModuleVersionForm(module) {
  moduleVersionForm.comment = ''
  moduleVersionForm.variables = module?.variables || '{}'
  moduleVersionForm.content = module?.content || ''
}

function openCreateTemplate() {
  resetTemplateForm()
  ensureTemplateModal()
  templateModal.show()
}

function openAssemblyTemplateBuilder() {
  activeTab.value = 'wizard'
}

function openTemplateInWizard(template, options = {}) {
  if (!template) return
  activeTab.value = 'wizard'
  loadWizardFromTemplate(template)
  if (!options.suppressRouteSync) {
    router.replace({
      path: '/configs',
      query: { tab: 'wizard' },
    })
  }
}

function applyAIModuleVariables(raw) {
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

async function saveTemplate() {
  try {
    await createTemplate(templateForm)
    templateModal.hide()
    resetAIDraftState(aiTemplateDraftState)
    await loadTemplates()
  } catch (error) {
    alert(`${t('configs_page.create_template_failed')}: ${getErrorMessage(error)}`)
  }
}

function applySuggestedTemplateName() {
  const suggestedName = aiTemplateDraftComparison.value?.suggestedName
  if (suggestedName) {
    templateForm.name = suggestedName
  }
}

async function openExistingTemplateFromDraft() {
  const existing = aiTemplateDraftComparison.value?.existingAsset
  if (!existing) return
  templateModal?.hide()
  await router.push(`/configs/${existing.id}`)
}

async function handleDeleteTemplate(template) {
  if (!confirm(t('configs_page.confirm_delete_template').replace('{name}', template.name))) return

  try {
    await deleteTemplate(template.id)
    await loadTemplates()
  } catch (error) {
    alert(`${t('configs_page.delete_template_failed')}: ${getErrorMessage(error)}`)
  }
}

function openCreateModule() {
  resetModuleForm()
  ensureModuleModal()
  moduleModal.show()
}

function openEditModule(module) {
  resetAIDraftState(aiModuleDraftState)
  editingModuleId.value = module.id
  moduleForm.name = module.name
  moduleForm.description = module.description || ''
  moduleForm.module_type = module.module_type
  moduleForm.fluent_type = module.fluent_type
  moduleForm.content = module.content
  moduleForm.content_fluentd = module.content_fluentd || ''
  moduleForm.variables = module.variables || '{}'
  moduleForm.is_builtin = !!module.is_builtin
  moduleVariablesMode.value = 'form'
  moduleVariableRows.value = buildModuleVariableRows(moduleForm.variables)
  moduleVariablesFormError.value = ''
  ensureModuleModal()
  moduleModal.show()
}

function applySuggestedModuleName() {
  const suggestedName = aiModuleDraftComparison.value?.suggestedName
  if (suggestedName) {
    moduleForm.name = suggestedName
  }
}

function openExistingModuleFromDraft() {
  const existing = aiModuleDraftComparison.value?.existingAsset
  if (!existing) return
  openEditModule(existing)
}

async function saveModule() {
  try {
    if (moduleVariablesMode.value === 'form') {
      if (!syncModuleVariablesFromRows()) {
        alert(t('configs_page.variable_form_invalid'))
        return
      }
    } else {
      parseVariablesMapStrict(moduleForm.variables)
    }
    if (editingModuleId.value) {
      await updateModule(editingModuleId.value, moduleForm)
    } else {
      await createModule(moduleForm)
    }
    moduleModal.hide()
    resetAIDraftState(aiModuleDraftState)
    await Promise.all([loadModules(), loadModuleTable()])
  } catch (error) {
    alert(`${editingModuleId.value ? t('configs_page.save_module_failed') : t('configs_page.create_module_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDeleteModule(module) {
  if (module.is_builtin) {
    alert(t('configs_page.builtin_module_protected'))
    return
  }
  if (!confirm(t('configs_page.confirm_delete_module').replace('{name}', module.name))) return

  try {
    await deleteModule(module.id)
    selectedModuleIds.value = selectedModuleIds.value.filter((id) => id !== module.id)
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => id !== module.id)
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== module.id)
    removeWizardModuleReferences(module.id)
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
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => !deletedSet.has(id))
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => !deletedSet.has(id))
    removeWizardModuleReferencesBatch(Array.from(deletedSet))
    await Promise.all([loadModules(), loadModuleTable()])
  } catch (error) {
    alert(`${t('configs_page.batch_delete_modules_failed')}: ${getErrorMessage(error)}`)
  }
}

async function openModuleVersions(module) {
  currentModule.value = module
  resetModuleVersionForm(module)
  ensureModuleVersionsModal()
  try {
    const { data } = await getModuleVersions(module.id)
    moduleVersions.value = data.data || []
    moduleVersionsModal.show()
  } catch (error) {
    alert(`${t('configs_page.load_module_versions_failed')}: ${getErrorMessage(error)}`)
  }
}

async function saveModuleVersion() {
  if (!currentModule.value) return

  try {
    await createModuleVersion(currentModule.value.id, moduleVersionForm)
    await openModuleVersions(currentModule.value)
    await Promise.all([loadModules(), loadModuleTable()])
  } catch (error) {
    alert(`${t('configs_page.create_module_version_failed')}: ${getErrorMessage(error)}`)
  }
}

function togglePreviewModule(moduleId) {
  if (selectedPreviewModuleIds.value.includes(moduleId)) {
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => id !== moduleId)
    return
  }
  selectedPreviewModuleIds.value = [...selectedPreviewModuleIds.value, moduleId]
}

function toggleWizardModule(moduleId) {
  const module = wizardEligibleModules.value.find((item) => item.id === moduleId)
  if (!module) return
  if (selectedWizardModuleIds.value.includes(moduleId)) {
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== moduleId)
    return
  }
  if (module.module_type === 'input') {
    const inputIds = wizardSelectedModules.value
      .filter((item) => item.module_type === 'input')
      .map((item) => item.id)
    selectedWizardModuleIds.value = [
      ...selectedWizardModuleIds.value.filter((id) => !inputIds.includes(id)),
      moduleId,
    ]
    mergeWizardModuleVariableValues(moduleId, parseVariablesMap(module.variables))
    return
  }
  selectedWizardModuleIds.value = [...selectedWizardModuleIds.value, moduleId]
  mergeWizardModuleVariableValues(moduleId, parseVariablesMap(module.variables))
}

function buildWizardModuleRef(module, draft, defaults = {}) {
  if (!module) return null
  const normalized = normalizeWizardDraftValues(draft, {
    ...moduleVariablesForWizard(module),
    ...(defaults || {}),
  })
  if (!Object.keys(normalized).length) {
    return { module_id: module.id }
  }
  return {
    module_id: module.id,
    variables: JSON.stringify(normalized, null, 2),
  }
}

function buildWizardRenderModuleRefs() {
  const refs = []

  if (wizardServiceModule.value) {
    refs.push(buildWizardModuleRef(
      wizardServiceModule.value,
      wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`]
    ))
  }

  for (const module of wizardSelectedParserModules.value) {
    refs.push(buildWizardModuleRef(
      module,
      wizardGlobalModuleVariables.value[`parser:${module.id}`]
    ))
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
  if (!wizardPipelines.value.length) {
    alert(t('configs_page.wizard_require_pipeline'))
    return
  }
  if (!wizardRenderablePipelineCards.value.length) {
    if (wizardIncompletePipelineLabels.value.length) {
      alert(t('configs_page.wizard_incomplete_pipelines').replace('{items}', wizardIncompletePipelineLabels.value.join(', ')))
      return
    }
    alert(t('configs_page.wizard_require_pipeline'))
    return
  }
  if (wizardOutputResolutionWarnings.value.length) {
    alert(wizardOutputResolutionWarnings.value
      .map((item) => `${item.pipeline}: ${item.target}`)
      .join('\n'))
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
    alert(t('configs_page.require_preview').replace('{action}', wizardSaveButtonLabel.value))
    return
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
          ? {
              id: wizardServiceModule.value.id,
              variables: normalizeWizardDraftValues(
                wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`],
                moduleVariablesForWizard(wizardServiceModule.value)
              ),
            }
          : null,
        parser_module_ids: wizardSelectedParserModules.value.map((item) => item.id),
        parser_modules: wizardSelectedParserModules.value.map((item) => ({
          id: item.id,
          variables: normalizeWizardDraftValues(
            wizardGlobalModuleVariables.value[`parser:${item.id}`],
            moduleVariablesForWizard(item)
          ),
        })),
      },
      pipelines: wizardPipelineCards.value.map((card) => {
        const pipelineState = pipelineStateByID.get(card.id)
        return {
          name: wizardPipelineDisplayName(card, card.index),
          complete: card.complete,
          input_module_id: card.inputModule?.id || null,
          input: card.inputModule
            ? {
                module_id: card.inputModule.id,
                variables: normalizeWizardDraftValues(
                  pipelineState?.input?.variables,
                  moduleVariablesForWizard(card.inputModule)
                ),
              }
            : null,
          filter_module_ids: card.filterModules.map((item) => item.id),
          filters: (pipelineState?.filters || []).map((instance) => {
            const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
            return {
              module_id: instance.module_id,
              variables: normalizeWizardDraftValues(instance.variables, moduleVariablesForWizard(module)),
            }
          }),
          routes: (pipelineState?.routes || []).map((instance) => {
            const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
            return {
              module_id: instance.module_id,
              variables: normalizeWizardDraftValues(instance.variables, moduleVariablesForWizard(module)),
            }
          }),
          output_targets: card.outputTargets.map((target) => {
            const instance = (pipelineState?.outputs || []).find((output) => output.target_id === target.id)
            return {
              id: target.id,
              name: target.name,
              target_type: target.target_type,
              endpoint: target.endpoint,
              fluent_type: target.fluent_type,
              variables: normalizeWizardDraftValues(
                instance?.variables,
                {
                  ...moduleVariablesForWizard(matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)),
                  ...parseVariablesMap(target?.settings),
                  output_target_name: target?.name || '',
                  output_target_type: target?.target_type || '',
                }
              ),
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
  return wizardPipelines.value.some(
    (p) => p.input || p.filters.length || (p.routes || []).length || p.outputs.length
  )
}

function loadWizardFromTemplate(template) {
  if (!template?.flow_layout) {
    alert(t('configs_page.wizard_load_incompatible'))
    return
  }
  let layout
  try {
    layout = typeof template.flow_layout === 'string' ? JSON.parse(template.flow_layout) : template.flow_layout
  } catch {
    alert(t('configs_page.wizard_load_incompatible'))
    return
  }
  if (layout.builder !== 'wizard') {
    alert(t('configs_page.wizard_load_incompatible'))
    return
  }
  if (wizardHasContent() && !confirm(t('configs_page.wizard_load_overwrite_confirm'))) return

  // Reset current wizard state
  wizardPipelines.value = []
  wizardServiceModuleId.value = null
  wizardParserModuleIds.value = []
  wizardGlobalModuleVariables.value = {}
  renderedConfig.value = null

  // Restore basic settings
  if (layout.runtime) wizardForm.fluent_type = layout.runtime
  if (layout.goal) wizardForm.goal = layout.goal
  wizardForm.name = template.name || ''
  wizardForm.description = template.description || ''

  // Restore global resources
  const restoredServiceModuleID = layout.global?.service_module?.id || layout.global?.service_module_id
  if (restoredServiceModuleID) {
    selectWizardServiceModule(restoredServiceModuleID)
    const module = wizardEligibleModules.value.find((item) => item.id === restoredServiceModuleID && item.module_type === 'service')
    const restoredVariables = layout.global?.service_module?.variables
    if (module && restoredVariables && typeof restoredVariables === 'object') {
      wizardGlobalModuleVariables.value = {
        ...wizardGlobalModuleVariables.value,
        [`service:${restoredServiceModuleID}`]: ensureWizardModuleDraft(
          `service:${restoredServiceModuleID}`,
          module,
          restoredVariables
        ),
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

  // Restore pipelines
  const restored = (layout.pipelines || []).map((p) => {
    const pid = createWizardInstanceID('wizard-pipeline')
    const inputModuleID = p.input?.module_id || p.input_module_id
    const inputModule = wizardEligibleModules.value.find((item) => item.id === inputModuleID && item.module_type === 'input')
    return {
      id: pid,
      name: p.name || '',
      input: inputModuleID
        ? {
            id: createWizardInstanceID('wizard-input'),
            module_id: inputModuleID,
            variables: ensureWizardModuleDraft(`input:${inputModuleID}`, inputModule, p.input?.variables),
          }
        : null,
      filters: (Array.isArray(p.filters) ? p.filters : (p.filter_module_ids || []).map((mid) => ({ module_id: mid })))
        .map((entry) => {
          const moduleID = Number(entry?.module_id || entry)
          const module = wizardEligibleModules.value.find((item) => item.id === moduleID && item.module_type === 'filter')
          if (!moduleID) return null
          return {
            id: createWizardInstanceID('wizard-filter'),
            module_id: moduleID,
            variables: ensureWizardModuleDraft(`filter:${moduleID}`, module, entry?.variables),
          }
        })
        .filter(Boolean),
      routes: (Array.isArray(p.routes) ? p.routes : [])
        .map((entry) => {
          const moduleID = Number(entry?.module_id || entry)
          const module = wizardEligibleModules.value.find((item) => item.id === moduleID && item.module_type === 'route')
          if (!moduleID) return null
          return {
            id: createWizardInstanceID('wizard-route'),
            module_id: moduleID,
            variables: ensureWizardModuleDraft(`route:${moduleID}`, module, entry?.variables),
          }
        })
        .filter(Boolean),
      outputs: (p.output_targets || [])
        .map((ot) => {
          const targetID = Number(ot?.id)
          const target = wizardAvailableOutputTargets.value.find((item) => item.id === targetID)
          if (!targetID) return null
          return {
            id: createWizardInstanceID('wizard-output'),
            target_id: targetID,
            variables: buildWizardOutputDraft(target, wizardForm.fluent_type, {
              input: inputModuleID
                ? {
                    module_id: inputModuleID,
                    variables: p.input?.variables,
                  }
                : null,
            }, ot?.variables),
          }
        })
        .filter(Boolean),
    }
  })
  wizardPipelines.value = restored.length ? restored : [createWizardPipeline()]
  activeWizardPipelineId.value = wizardPipelines.value[0].id
  wizardLoadedFromTemplate.value = template
}

function clearWizardLoadedTemplate() {
  wizardLoadedFromTemplate.value = null
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
      (m) => m.fluent_type === aiAssistantForm.fluent_type || m.fluent_type === 'shared'
    )
    aiAssistantModules.value = mergeAIModules(data.modules || [], existingList)

    setAIAssistantFeedback(
      'success',
      t('configs_page.ai_assistant_success'),
      t('configs_page.ai_assistant_ready'),
      [data.provider, data.account_name].filter(Boolean).join(' / ')
    )
  } catch (error) {
    aiAssistantResult.value = null
    aiAssistantModules.value = []
    setAIAssistantFeedback(
      'danger',
      t('configs_page.ai_assistant_failed'),
      getErrorMessage(error),
      error?.response?.data?.provider || '',
      getProviderErrorMessage(error)
    )
  } finally {
    aiAssistantLoading.value = false
  }
}

// Normalise a single line/segment for comparison: collapse inline whitespace, lowercase.
function normalizeLine(s) {
  return String(s || '').replace(/[ \t]+/g, ' ').trim().toLowerCase()
}

// Split raw content into comparable segments (newline or semicolon delimited), normalising each.
function splitContentLines(raw) {
  return String(raw || '')
    .split(/[;\n]/)
    .map(normalizeLine)
    .filter(Boolean)
}

// Ratio of shared segments between two raw content strings.
function contentSimilarityRatio(rawA, rawB) {
  const aLines = splitContentLines(rawA)
  const bLines = splitContentLines(rawB)
  if (!aLines.length || !bLines.length) return 0
  const bSet = new Set(bLines)
  const matching = aLines.filter((l) => bSet.has(l)).length
  return matching / Math.max(aLines.length, bLines.length)
}

// Match each AI-generated module against the existing catalog and assign a default decision.
function mergeAIModules(generatedModules, existingModules) {
  return generatedModules.map((mod) => {
    const nameLower = (mod.name || '').toLowerCase()
    const typeLower = (mod.module_type || '').toLowerCase()

    // 1. Exact name + type match.
    const exactMatch = existingModules.find(
      (e) => e.name.toLowerCase() === nameLower && e.module_type.toLowerCase() === typeLower
    )
    if (exactMatch) {
      const ratio = contentSimilarityRatio(mod.content, exactMatch.latest_content || '')
      const decision = ratio >= 0.95 ? 'reuse_existing' : 'update_existing'
      return { ...mod, decision, matchedModule: exactMatch }
    }

    // 2. Content similarity scan across same-type modules (≥70% line overlap).
    const sameType = existingModules.filter((e) => e.module_type.toLowerCase() === typeLower)
    const similarMatch = sameType.find(
      (e) => contentSimilarityRatio(mod.content, e.latest_content || '') >= 0.7
    )
    if (similarMatch) {
      return { ...mod, decision: 'reuse_existing', matchedModule: similarMatch }
    }

    return { ...mod, decision: 'create_new', matchedModule: null }
  })
}

async function saveAIModules() {
  if (!aiAssistantModules.value.length) return
  aiAssistantModulesSaving.value = true
  const errors = []
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
          if (!item.matchedModule) {
            throw new Error(`No existing module matched. Change decision to "Create new" or re-run.`)
          }
          await createModuleVersion(item.matchedModule.id, {
            content: item.content,
            variables: item.variables_json || '{}',
            comment: `AI-generated update from assistant`,
          })
        }
      } catch (e) {
        errors.push(`"${item.name}": ${e?.response?.data?.error || e?.message || ''}`)
      }
    }
    // Reload modules and re-run merge so cards reflect current state.
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

async function runImportAnalysis() {
  if (!importForm.content.trim()) {
    importedConfigResult.value = null
    return
  }

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
    const invalidNamedModule = importedConfigResult.value.modules.find((item) =>
      item.module_type !== 'output' &&
      item.import_action !== 'reuse_existing' &&
      !String(item.name || '').trim()
    )
    if (invalidNamedModule) {
      throw new Error(t('configs_page.import_module_name_required').replace('{order}', String(invalidNamedModule.order)))
    }

    const existingModules = await listAllModules()
    const existingOutputTargets = [...outputTargets.value]
    const occupiedIdentities = new Set(
      existingModules.map((item) => importedModuleIdentity(item.name, item.module_type, item.fluent_type))
    )
    const importNameConflict = findImportedModuleNameConflict(importedConfigResult.value.modules, occupiedIdentities)
    if (importNameConflict?.type === 'batch_duplicate') {
      throw new Error(t('configs_page.import_module_name_duplicate_batch').replace('{name}', String(importNameConflict.item?.name || '').trim()))
    }
    if (importNameConflict?.type === 'existing_duplicate') {
      throw new Error(t('configs_page.import_module_name_duplicate_existing').replace('{name}', String(importNameConflict.item?.name || '').trim()))
    }
    const occupiedOutputTargetNames = new Set(existingOutputTargets.map((item) => normalizeName(item.name)).filter(Boolean))
    const reusableImportedOutputTargets = new Map()
    for (const target of existingOutputTargets) {
      reusableImportedOutputTargets.set(createImportedOutputTargetSignature(target), target)
    }
    const created = []
    const assembledModuleRefs = []
    const assembledModules = []
    const ensuredDestinations = Array.isArray(importedConfigResult.value.destinations)
      ? importedConfigResult.value.destinations.map((item) => ({ ...item }))
      : []
    for (const item of importedConfigResult.value.modules) {
      if (item.module_type === 'output') {
        let ensuredTarget = null
        if (item.output_target_id) {
          ensuredTarget = existingOutputTargets.find((target) => target.id === item.output_target_id) || null
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
              name: targetName,
              description: buildImportedModuleDescription(item),
              fluent_type: 'shared',
              target_type: targetDraft.target_type,
              endpoint: targetDraft.endpoint,
              settings: targetDraft.settings,
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
        if (!adapterModule) {
          throw new Error(t('configs_page.import_output_adapter_missing').replace('{type}', ensuredTarget.target_type || 'output'))
        }

        assembledModuleRefs.push({
          module_id: adapterModule.id,
          variables: JSON.stringify(buildImportedOutputRenderVariables(item, ensuredTarget, importForm.fluent_type), null, 2),
        })
        ensuredDestinations.push({
          output_module_name: adapterModule.name,
          output_module_order: item.order,
          output_target_id: ensuredTarget.id,
          name: ensuredTarget.name,
          target_type: ensuredTarget.target_type,
          endpoint: ensuredTarget.endpoint,
          match_type: item.output_target_match_type || 'created',
        })
        continue
      }

      if (item.import_action === 'reuse_existing' && item.existing_module_id) {
        assembledModuleRefs.push({
          module_id: item.existing_module_id,
          variables: item.variables || '{}',
        })
        assembledModules.push({
          id: Number(item.existing_module_id),
          module_type: item.module_type,
        })
        continue
      }

      const name = uniqueImportedModuleName(item.name, item.module_type, item.fluent_type, occupiedIdentities)
      const { data } = await createModule({
        name,
        description: buildImportedModuleDescription(item),
        module_type: item.module_type,
        fluent_type: item.fluent_type,
        content: item.content,
        variables: item.variables || '{}',
        is_builtin: false,
      })
      created.push(data)
      assembledModuleRefs.push({
        module_id: data.id,
        variables: item.variables || '{}',
      })
      assembledModules.push({
        id: Number(data.id),
        module_type: item.module_type,
      })
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
    selectedPreviewModuleIds.value = assembledModules
      .filter((item) => item.module_type !== 'output')
      .map((item) => item.id)

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
      name: previewForm.name,
      fluent_type: previewForm.fluent_type,
      runtime_version: previewForm.runtime_version,
      variables: previewForm.variables,
      modules: assembledModuleRefs,
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
    const templateName = generateUniqueDraftName(
      importedConfigResult.value.suggested_template_name || `${importForm.name_prefix || 'imported-config'}-assembly`,
      templates.value.map((item) => item.name),
      `${importForm.name_prefix || 'imported-config'}-assembly`
    )
    const matchedExistingCount = importedConfigResult.value.modules.filter((item) => item.existing_module_id).length
    const reusedExistingCount = importedConfigResult.value.modules.filter((item) => item.import_action === 'reuse_existing').length
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

function useAIModuleDraft(module) {
  if (!module) return

  resetModuleForm()
  editingModuleId.value = null
  moduleForm.name = module.name || `ai-${module.module_type}`
  moduleForm.description = aiAssistantResult.value?.summary || ''
  moduleForm.module_type = module.module_type || aiAssistantForm.module_type
  moduleForm.fluent_type = aiAssistantForm.fluent_type
  moduleForm.content = module.content || ''
  moduleForm.is_builtin = false
  applyAIModuleVariables(module.variables_json || '{}')
  // Pass a synthetic result so activateAIDraftState gets the right notes/steps.
  const draftResult = {
    ...aiAssistantResult.value,
    notes: module.note ? [module.note, ...(aiAssistantResult.value?.notes || [])] : (aiAssistantResult.value?.notes || []),
    assembly_steps: [],
  }
  activateAIDraftState(aiModuleDraftState, draftResult, [
    t('configs_page.ai_draft_review_name'),
    t('configs_page.ai_draft_review_runtime'),
    t('configs_page.ai_draft_review_variables'),
    t('configs_page.ai_draft_review_module_content'),
  ], [
    t('configs_page.ai_draft_confirm_name'),
    t('configs_page.ai_draft_confirm_variables'),
    t('configs_page.ai_draft_confirm_target'),
    t('configs_page.ai_draft_confirm_module_content'),
  ])
  ensureModuleModal()
  moduleModal.show()
}

function useAITemplateDraft(pipeline) {
  if (!aiAssistantResult.value) return

  resetTemplateForm()
  if (pipeline) {
    templateForm.name = pipeline.name || `ai-${aiAssistantForm.fluent_type}-template`
    templateForm.description = pipeline.description || aiAssistantResult.value.summary || ''
    templateForm.content = pipeline.template_content || ''
  } else {
    templateForm.name = `ai-${aiAssistantForm.fluent_type}-template`
    templateForm.description = aiAssistantResult.value.summary || ''
    templateForm.content = (aiAssistantResult.value.pipelines && aiAssistantResult.value.pipelines[0]?.template_content) || ''
  }
  templateForm.fluent_type = aiAssistantForm.fluent_type
  activateAIDraftState(aiTemplateDraftState, aiAssistantResult.value, [
    t('configs_page.ai_draft_review_name'),
    t('configs_page.ai_draft_review_runtime'),
    t('configs_page.ai_draft_review_template_content'),
    t('configs_page.ai_draft_review_notes'),
  ], [
    t('configs_page.ai_draft_confirm_name'),
    t('configs_page.ai_draft_confirm_target'),
    t('configs_page.ai_draft_confirm_template_content'),
    t('configs_page.ai_draft_confirm_notes'),
  ])
  ensureTemplateModal()
  templateModal.show()
}

// Resolve AI pipeline module_names to catalog module objects.
// Disambiguation key: name + module_type + fluent_type (from AI result + current runtime).
// Returns { resolved: Module[], unsaved: string[] }.
function resolveAIPipelineModules(pipeline) {
  const ft = aiAssistantForm.fluent_type
  // AI result carries name + module_type for each generated module
  const aiMetaByName = new Map(
    (aiAssistantResult.value?.modules || []).map((m) => [m.name, { type: m.module_type }])
  )
  // Catalog keyed by "name\0type\0fluent_type" (most specific), then "name\0type", then "name"
  const key3 = (name, type, ftype) => `${name}\0${type}\0${ftype}`
  const key2 = (name, type) => `${name}\0${type}`
  const byKey3 = new Map(modules.value.map((m) => [key3(m.name, m.module_type, m.fluent_type), m]))
  const byKey2 = new Map(modules.value.map((m) => [key2(m.name, m.module_type), m]))
  const byName = new Map(modules.value.map((m) => [m.name, m]))

  const resolved = []
  const unsaved = []
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
// output modules: try name-match against known OutputTarget records.
// Matched → matchedTargets (can be used as output_target_ids / wizard outputs).
// Unmatched output + all route/filter/parser/service → stageMods (ordered pipeline stages).
function splitAIPipelineModules(resolvedModules) {
  const inputMod = resolvedModules.find((m) => m.module_type === 'input') || null
  const outputMods = resolvedModules.filter((m) => m.module_type === 'output')
  const stageMods = resolvedModules.filter((m) => m.module_type !== 'input' && m.module_type !== 'output')

  const matchedTargets = []
  const unmatchedOutputMods = []
  for (const mod of outputMods) {
    const target = outputTargets.value.find((tgt) => tgt.name === mod.name)
    if (target) matchedTargets.push(target)
    else unmatchedOutputMods.push(mod)
  }

  // Unmatched output modules fall back to stage position (after filters)
  return { inputMod, stageMods: [...stageMods, ...unmatchedOutputMods], matchedTargets, unmatchedOutputMods }
}

async function saveAIPipelineAsConfigPipeline(pipeline) {
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
      output_target_ids: matchedTargets.map((t) => t.id),
    })
    await loadPipelines()
    const msg = t('configs_page.ai_pipeline_saved_as_pipeline').replace('{name}', data.name)
    const detail = unmatchedOutputMods.length
      ? t('configs_page.ai_pipeline_output_stages_note').replace('{names}', unmatchedOutputMods.map((m) => m.name).join(', '))
      : ''
    setAIAssistantFeedback('success', msg, detail)
  } catch (e) {
    setAIAssistantFeedback('danger', t('common.request_failed'), e?.response?.data?.error || e?.message || '')
  }
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

function wizardGoalLabel(goal) {
  const labels = {
    edge_collection: t('configs_page.goal_edge_collection'),
    central_aggregation: t('configs_page.goal_central_aggregation'),
    custom_pipeline: t('configs_page.goal_custom_pipeline'),
  }
  return labels[goal] || goal
}

async function runPreview(options = {}) {
  const outputModuleRefs = buildOutputTargetModuleRefs(
    previewResolvedOutputTargets.value,
    previewEligibleModules.value,
    previewForm.fluent_type
  )
  if (previewUnresolvedOutputTargets.value.length) {
    alert(t('configs_page.output_target_module_missing').replace('{targets}', previewUnresolvedOutputTargets.value.map((item) => item.name).join(', ')))
    return
  }
  if (!selectedPreviewModuleIds.value.length && !outputModuleRefs.length) {
    alert(t('configs_page.choose_modules'))
    return
  }

  try {
    const payload = {
      name: previewForm.name,
      fluent_type: previewForm.fluent_type,
      runtime_version: previewForm.runtime_version,
      variables: previewForm.variables,
      modules: [
        ...buildPreviewSelectedModuleRefs(),
        ...outputModuleRefs,
      ],
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
    if (options.switchTab !== false) {
      activeTab.value = 'preview'
    }
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
    analysisResult.value = await lintConfig({
      fluent_type: renderedConfig.value.fluent_type,
      runtime_version: renderedConfig.value.runtime_version,
      content,
    })
    activeTab.value = 'preview'
  } catch (error) {
    alert(`${t('configs_page.lint_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runCompatibility() {
  const content = requireRenderedConfig(t('configs_page.run_compatibility'))
  if (!content) return

  try {
    const payload = {
      fluent_type: renderedConfig.value.fluent_type,
      runtime_version: renderedConfig.value.runtime_version,
      content,
    }
    if (previewForm.node_id) {
      payload.node_id = Number(previewForm.node_id)
    }
    compatibilityResult.value = await checkCompatibility(payload)
    activeTab.value = 'preview'
  } catch (error) {
    alert(`${t('configs_page.compatibility_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runReplay() {
  const content = requireRenderedConfig(t('configs_page.run_replay'))
  if (!content) return
  if (!previewForm.sample_log) {
    alert(t('configs_page.require_sample_log'))
    return
  }

  try {
    replayResult.value = await replayConfig({
      fluent_type: renderedConfig.value.fluent_type,
      runtime_version: renderedConfig.value.runtime_version,
      content,
      sample_log: previewForm.sample_log,
      sample_tag: previewForm.sample_tag,
    })
    activeTab.value = 'preview'
  } catch (error) {
    alert(`${t('configs_page.replay_failed')}: ${getErrorMessage(error)}`)
  }
}

async function runSemanticDiff() {
  const content = requireRenderedConfig(t('configs_page.run_diff'))
  if (!content) return
  if (!previewForm.diff_content) {
    alert(t('configs_page.require_diff_content'))
    return
  }

  try {
    diffResult.value = await diffConfig({
      fluent_type: renderedConfig.value.fluent_type,
      before_content: previewForm.diff_content,
      after_content: content,
    })
    activeTab.value = 'preview'
  } catch (error) {
    alert(`${t('configs_page.diff_failed')}: ${getErrorMessage(error)}`)
  }
}

function findingBadgeClass(severity) {
  if (severity === 'error') return 'text-bg-danger'
  if (severity === 'warning') return 'text-bg-warning'
  return 'text-bg-info'
}

watch(
  () => pipelineForm.fluent_type,
  () => {
    if (pipelineFormInitializing.value) return
    pipelineForm.input_module_id = null
    pipelineForm.filter_module_ids = []
    pipelineForm.output_target_ids = []
    pipelineFilterPickerValue.value = ''
  },
  { flush: 'sync' }
)

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

watch(
  () => wizardForm.fluent_type,
  (runtime) => {
    if (!wizardForm.name) {
      wizardForm.name = `guided-${runtime}`
    }
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

watch(
  selectedWizardModuleIds,
  () => {
    if (!wizardForm.description) {
      wizardForm.description = t('configs_page.wizard_default_description').replace('{goal}', wizardGoalLabel(wizardForm.goal))
    }
  }
)

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

onMounted(async () => {
  resetWizardForm()
  await Promise.all([loadTemplates(), loadModules(), loadModuleTable(), loadOutputTargets(), loadPipelines()])
  ensureWizardBaselineModules()

  if (route.query.tab === 'wizard') {
    activeTab.value = 'wizard'
  }

  const loadTemplateID = Number(route.query.load_template || 0)
  if (loadTemplateID) {
    const template = templates.value.find((item) => item.id === loadTemplateID)
    if (template) {
      openTemplateInWizard(template, { suppressRouteSync: true })
      await router.replace({
        path: '/configs',
        query: { tab: 'wizard' },
      })
    }
  }
})
</script>
