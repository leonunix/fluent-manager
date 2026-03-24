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
      :actions="{ handleDeleteTemplate }"
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
        wizardOutputSearch,
        wizardStagePages,
        wizardPagedServiceModules,
        wizardPagedParserModules,
        wizardPagedInputModules,
        wizardPagedFilterModules,
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
        addWizardOutputTarget,
        removeWizardOutput,
        moveWizardOutput,
        changeWizardStagePage,
        runWizardPreview,
        saveWizardAsTemplate,
        openAdvancedPreviewFromWizard,
      }"
      :helpers="{ runtimeLabel, wizardGoalLabel, wizardPipelineDisplayName, matchingOutputModuleForTarget }"
    />

    <ConfigsAssistantTab
      v-else-if="activeTab === 'assistant'"
      :state="{ aiAssistantForm, aiAssistantLoading, aiAssistantResult, aiAssistantFeedback, moduleTypes }"
      :actions="{ runAIAssistant, useAIModuleDraft, useAITemplateDraft }"
      :helpers="{ runtimeLabel }"
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
                <span>{{ t('configs_page.version_content') }}</span>
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
import { useRouter } from 'vue-router'
import ConfigsAssistantTab from './tabs/ConfigsAssistantTab.vue'
import ConfigsImportTab from './tabs/ConfigsImportTab.vue'
import ConfigsModulesTab from './tabs/ConfigsModulesTab.vue'
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
  updateModule,
} from '../../api'

const activeTab = ref('templates')
const templates = ref([])
const modules = ref([])
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
const wizardPipelineModuleTypes = ['input', 'filter']
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
const wizardServiceSearch = ref('')
const wizardParserSearch = ref('')
const wizardInputSearch = ref('')
const wizardFilterSearch = ref('')
const wizardOutputSearch = ref('')
const wizardStagePages = reactive({
  service: 1,
  parser: 1,
  input: 1,
  filter: 1,
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
const assemblyTemplateCount = computed(() => templates.value.filter((item) => item.source_type === 'module_assembly').length)
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
const previewAvailableOutputTargets = computed(() =>
  outputTargets.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
)
const wizardPagedServiceModules = computed(() => paginateItems(wizardServiceModules.value, wizardStagePages.service))
const wizardPagedParserModules = computed(() => paginateItems(wizardParserModules.value, wizardStagePages.parser))
const wizardPagedInputModules = computed(() => paginateItems(wizardInputModules.value, wizardStagePages.input))
const wizardPagedFilterModules = computed(() => paginateItems(wizardFilterModules.value, wizardStagePages.filter))
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
    const outputTargetsForPipeline = pipeline.outputs
      .map((instance) => wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id) || null)
      .filter(Boolean)
    const outputModulesForPipeline = outputTargetsForPipeline
      .map((target) => matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type))
      .filter(Boolean)
    const summary = buildConfigFlowSummary(
      [inputModule, ...filterModules, ...outputModulesForPipeline].filter(Boolean),
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
  if (wizardServiceModule.value) {
    groups.push(buildWizardModuleGroup(
      `wizard-service-${wizardServiceModule.value.id}`,
      wizardServiceModule.value.name,
      t('configs_page.wizard_service_baseline'),
      wizardServiceModule.value,
      wizardGlobalModuleVariables.value[`service:${wizardServiceModule.value.id}`]
    ))
  }
  for (const module of wizardSelectedParserModules.value) {
    groups.push(buildWizardModuleGroup(
      `wizard-parser-${module.id}`,
      module.name,
      t('configs_page.wizard_parser_assets'),
      module,
      wizardGlobalModuleVariables.value[`parser:${module.id}`]
    ))
  }
  return groups.filter(Boolean)
})
const wizardPipelineVariableGroups = computed(() => {
  if (!activeWizardPipeline.value) return []
  const groups = []
  if (activeWizardPipeline.value.input) {
    const module = wizardEligibleModules.value.find((item) => item.id === activeWizardPipeline.value.input.module_id)
    groups.push(buildWizardModuleGroup(
      activeWizardPipeline.value.input.id,
      module?.name || t('configs_page.pipeline_stage_input'),
      t('configs_page.pipeline_stage_input'),
      module,
      activeWizardPipeline.value.input.variables
    ))
  }
  for (const instance of activeWizardPipeline.value.filters) {
    const module = wizardEligibleModules.value.find((item) => item.id === instance.module_id)
    groups.push(buildWizardModuleGroup(
      instance.id,
      module?.name || t('configs_page.pipeline_stage_filter'),
      t('configs_page.pipeline_stage_filter'),
      module,
      instance.variables
    ))
  }
  for (const instance of activeWizardPipeline.value.outputs) {
    const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
    const module = matchingOutputModuleForTarget(target, wizardEligibleModules.value, wizardForm.fluent_type)
    const defaults = {
      ...moduleVariablesForWizard(module),
      ...parseVariablesMap(target?.settings),
      output_target_name: target?.name || '',
      output_target_type: target?.target_type || '',
    }
    groups.push(buildWizardModuleGroup(
      instance.id,
      target?.name || t('configs_page.pipeline_stage_output'),
      t('configs_page.pipeline_stage_output'),
      module,
      instance.variables,
      defaults
    ))
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

function normalizeWizardDraftValues(draft, defaults) {
  const normalized = {}
  const mergedDefaults = { ...(defaults || {}) }
  const keys = new Set([...Object.keys(mergedDefaults), ...Object.keys(draft || {})])
  for (const key of keys) {
    normalized[key] = normalizeWizardVariableValue(draft?.[key], inferVariableKind(mergedDefaults[key]))
  }
  return normalized
}

function buildWizardModuleGroup(key, title, subtitle, module, model, extraDefaults = {}) {
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
  updateWizardPipeline(pipelineId, (pipeline) => ({
    ...pipeline,
    input: {
      id: createWizardInstanceID('wizard-input'),
      module_id: module.id,
      variables: ensureWizardModuleDraft(`input:${module.id}`, module),
    },
  }))
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
        variables: ensureWizardModuleDraft(`filter:${module.id}`, module),
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

function buildWizardOutputDraft(target, fluentType, existingDraft = null) {
  const outputModule = matchingOutputModuleForTarget(target, wizardEligibleModules.value, fluentType)
  const defaults = {
    ...moduleVariablesForWizard(outputModule),
    ...parseVariablesMap(target?.settings),
    output_target_name: target?.name || '',
    output_target_type: target?.target_type || '',
  }
  return ensureWizardModuleDraft(`output:${target?.id}`, outputModule, existingDraft, defaults)
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
        variables: buildWizardOutputDraft(target, wizardForm.fluent_type),
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
    outputs: pipeline.outputs
      .filter((instance) => availableTargetIDs.has(instance.target_id))
      .map((instance) => {
        const target = wizardAvailableOutputTargets.value.find((item) => item.id === instance.target_id)
        return {
          ...instance,
          variables: buildWizardOutputDraft(target, wizardForm.fluent_type, instance.variables),
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
  wizardOutputSearch.value = ''
  wizardStagePages.service = 1
  wizardStagePages.parser = 1
  wizardStagePages.input = 1
  wizardStagePages.filter = 1
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
    alert(t('configs_page.require_preview').replace('{action}', t('configs_page.save_wizard_template')))
    return
  }

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
        parser_module_ids: wizardSelectedParserModules.value.map((item) => item.id),
      },
      pipelines: wizardPipelineCards.value.map((card) => ({
        name: wizardPipelineDisplayName(card, card.index),
        complete: card.complete,
        input_module_id: card.inputModule?.id || null,
        filter_module_ids: card.filterModules.map((item) => item.id),
        output_targets: card.outputTargets.map((target) => ({
          id: target.id,
          name: target.name,
          target_type: target.target_type,
          endpoint: target.endpoint,
          fluent_type: target.fluent_type,
        })),
      })),
    }),
  }

  try {
    await createTemplate(payload)
    await loadTemplates()
    activeTab.value = 'templates'
  } catch (error) {
    alert(`${t('configs_page.create_template_failed')}: ${getErrorMessage(error)}`)
  }
}

function openAdvancedPreviewFromWizard() {
  preparePreviewMetaFromWizard()
  activeTab.value = 'preview'
}

async function runAIAssistant() {
  if (!aiAssistantForm.sample.trim()) {
    aiAssistantResult.value = null
    setAIAssistantFeedback('danger', t('configs_page.require_sample_log'))
    return
  }

  aiAssistantLoading.value = true
  aiAssistantResult.value = null
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
    setAIAssistantFeedback(
      'success',
      t('configs_page.ai_assistant_success'),
      t('configs_page.ai_assistant_ready'),
      [data.provider, data.account_name].filter(Boolean).join(' / ')
    )
  } catch (error) {
    aiAssistantResult.value = null
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

function useAIModuleDraft() {
  if (!aiAssistantResult.value) return

  resetModuleForm()
  editingModuleId.value = null
  moduleForm.name = aiAssistantResult.value.recommended_module_name || `ai-${aiAssistantForm.module_type}`
  moduleForm.description = aiAssistantResult.value.summary || ''
  moduleForm.module_type = aiAssistantResult.value.module_type || aiAssistantForm.module_type
  moduleForm.fluent_type = aiAssistantForm.fluent_type
  moduleForm.content = aiAssistantResult.value.module_content || ''
  moduleForm.is_builtin = false
  applyAIModuleVariables(aiAssistantResult.value.variables_json || '{}')
  activateAIDraftState(aiModuleDraftState, aiAssistantResult.value, [
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

function useAITemplateDraft() {
  if (!aiAssistantResult.value) return

  resetTemplateForm()
  templateForm.name = aiAssistantResult.value.recommended_template_name || `ai-${aiAssistantForm.fluent_type}-template`
  templateForm.description = aiAssistantResult.value.summary || ''
  templateForm.fluent_type = aiAssistantForm.fluent_type
  templateForm.content = aiAssistantResult.value.template_content || ''
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
  await Promise.all([loadTemplates(), loadModules(), loadModuleTable(), loadOutputTargets()])
  ensureWizardBaselineModules()
})
</script>
