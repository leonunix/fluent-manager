<template>
  <div class="row g-4">
    <div class="col-12">
      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <h6 class="mb-0">{{ t('configs_page.wizard_basics') }}</h6>
          <div v-if="state.wizardLoadedFromTemplate" class="d-flex align-items-center gap-2">
            <span class="badge bg-success-subtle text-success-emphasis">
              <i class="bi bi-check-circle me-1"></i>{{ t('configs_page.wizard_loaded_from') }}: {{ state.wizardLoadedFromTemplate.name }}
            </span>
            <button class="btn btn-sm btn-outline-secondary py-0" @click="actions.clearWizardLoadedTemplate">
              {{ t('configs_page.wizard_clear_loaded') }}
            </button>
          </div>
        </div>
        <div class="card-body">
          <div class="alert alert-info py-2 mb-3">
            <div class="fw-semibold">{{ t('configs_page.wizard_intro') }}</div>
            <div class="small mt-1">{{ t('configs_page.wizard_architecture_hint') }}</div>
          </div>

          <!-- Load from existing template -->
          <div class="border rounded-4 p-3 mb-3 bg-light-subtle fm-template-picker">
            <div class="d-flex justify-content-between align-items-start gap-3 flex-wrap mb-3">
              <div>
                <div class="fw-semibold">{{ t('configs_page.wizard_load_from_template') }}</div>
                <div class="small text-muted mt-1">{{ t('configs_page.wizard_load_templates_hint') }}</div>
              </div>
              <span class="badge text-bg-light">{{ filteredLoadTemplates.length }} / {{ loadableTemplateTotal }}</span>
            </div>

            <div v-if="!loadableTemplateTotal" class="small text-muted">
              {{ t('configs_page.wizard_no_assembly_templates') }}
            </div>

            <div v-else class="row g-3">
              <div class="col-lg-5">
                <div class="input-group input-group-sm mb-3">
                  <span class="input-group-text"><i class="bi bi-search"></i></span>
                  <input
                    v-model="loadTemplateSearch"
                    type="text"
                    class="form-control"
                    :placeholder="t('configs_page.wizard_template_search_placeholder')"
                  >
                </div>

                <div class="d-flex flex-wrap gap-2 mb-3">
                  <button
                    type="button"
                    class="btn btn-sm"
                    :class="loadRuntimeFilter === 'current' ? 'btn-primary' : 'btn-outline-secondary'"
                    @click="loadRuntimeFilter = 'current'"
                  >
                    {{ t('configs_page.wizard_template_filter_current') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-sm"
                    :class="loadRuntimeFilter === 'all' ? 'btn-primary' : 'btn-outline-secondary'"
                    @click="loadRuntimeFilter = 'all'"
                  >
                    {{ t('configs_page.wizard_template_filter_all') }}
                  </button>
                  <button
                    v-for="runtime in runtimeFilters"
                    :key="runtime"
                    type="button"
                    class="btn btn-sm"
                    :class="loadRuntimeFilter === runtime ? 'btn-primary' : 'btn-outline-secondary'"
                    @click="loadRuntimeFilter = runtime"
                  >
                    {{ helpers.runtimeLabel(runtime) }}
                  </button>
                </div>

                <div v-if="!filteredLoadTemplates.length" class="fm-template-picker__empty border rounded-3 bg-white">
                  {{ t('configs_page.wizard_template_empty_search') }}
                </div>

                <div v-else class="fm-template-picker__list d-grid gap-2">
                  <button
                    v-for="tpl in filteredLoadTemplates"
                    :key="tpl.id"
                    type="button"
                    class="btn text-start fm-template-picker__item"
                    :class="String(selectedLoadTemplateId) === String(tpl.id) ? 'btn-primary' : 'btn-outline-secondary'"
                    @click="selectedLoadTemplateId = tpl.id"
                  >
                    <div class="d-flex justify-content-between align-items-start gap-2">
                      <div class="min-w-0">
                        <div class="fw-semibold text-truncate">{{ tpl.name }}</div>
                        <div class="small opacity-75 fm-template-picker__description">
                          {{ tpl.description || t('common.no_description') }}
                        </div>
                      </div>
                      <span class="badge" :class="String(selectedLoadTemplateId) === String(tpl.id) ? 'text-bg-light' : 'text-bg-secondary'">
                        {{ helpers.runtimeLabel(tpl.fluent_type) }}
                      </span>
                    </div>
                  </button>
                </div>
              </div>

              <div class="col-lg-7">
                <div v-if="selectedLoadTemplate" class="fm-template-picker__detail border rounded-3 bg-white p-3 h-100">
                  <div class="d-flex justify-content-between align-items-start gap-2 mb-3">
                    <div>
                      <div class="fw-semibold">{{ selectedLoadTemplate.name }}</div>
                      <div class="small text-muted mt-1">{{ selectedLoadTemplate.description || t('common.no_description') }}</div>
                    </div>
                    <span class="badge text-bg-light">{{ helpers.runtimeLabel(selectedLoadTemplate.fluent_type) }}</span>
                  </div>
                  <div class="small text-muted mb-3">
                    {{ t('configs_page.wizard_loaded_from') }} ID #{{ selectedLoadTemplate.id }}
                  </div>
                  <button
                    class="btn btn-sm btn-primary text-nowrap"
                    @click="doLoadFromTemplate"
                  >
                    <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.wizard_load_button') }}
                  </button>
                </div>

                <div v-else class="fm-template-picker__empty border rounded-3 bg-white">
                  {{ t('configs_page.wizard_template_select_prompt') }}
                </div>
              </div>
            </div>
          </div>

          <div class="row g-3">
            <div class="col-lg-3">
              <label class="form-label">{{ t('configs_page.wizard_goal') }}</label>
              <select v-model="state.wizardForm.goal" class="form-select">
                <option value="edge_collection">{{ t('configs_page.goal_edge_collection') }}</option>
                <option value="central_aggregation">{{ t('configs_page.goal_central_aggregation') }}</option>
                <option value="custom_pipeline">{{ t('configs_page.goal_custom_pipeline') }}</option>
              </select>
            </div>
            <div class="col-lg-3">
              <label class="form-label">{{ t('common.runtime') }}</label>
              <select v-model="state.wizardForm.fluent_type" class="form-select">
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
              </select>
            </div>
            <div class="col-lg-3">
              <label class="form-label">{{ t('configs_page.target_version') }}</label>
              <input v-model="state.wizardForm.runtime_version" type="text" class="form-control" placeholder="3.1.0 / 1.16">
            </div>
            <div class="col-lg-3">
              <label class="form-label">{{ t('common.name') }}</label>
              <input v-model="state.wizardForm.name" type="text" class="form-control" :placeholder="t('configs_page.wizard_name_placeholder')">
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('common.description') }}</label>
              <textarea v-model="state.wizardForm.description" class="form-control" rows="2" :placeholder="t('configs_page.wizard_description_placeholder')"></textarea>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="col-xl-4">
      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.wizard_global_resources') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.wizard_global_resources_hint') }}</div>
          </div>
          <span class="badge text-bg-light">{{ globalResourceCount }}</span>
        </div>
        <div class="card-body">
          <div class="mb-4">
            <div class="d-flex justify-content-between align-items-center mb-2">
              <div>
                <div class="fw-semibold">{{ t('configs_page.wizard_service_baseline') }}</div>
                <div class="small text-muted">{{ t('configs_page.wizard_service_baseline_hint') }}</div>
              </div>
              <span class="badge text-bg-light">{{ state.wizardPagedServiceModules.total }}</span>
            </div>
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input
                v-model="state.wizardServiceSearch"
                type="text"
                class="form-control"
                :placeholder="t('configs_page.module_picker_search_placeholder')"
                @input="actions.changeWizardStagePage('service', 1)"
              >
            </div>
            <div class="d-grid gap-2">
              <button
                v-for="module in state.wizardPagedServiceModules.items"
                :key="module.id"
                type="button"
                class="btn text-start"
                :class="state.wizardServiceModuleId === module.id ? 'btn-primary' : 'btn-outline-secondary'"
                @click="actions.selectWizardServiceModule(module.id)"
              >
                <div class="fw-semibold">{{ module.name }}</div>
                <div class="small opacity-75">{{ module.description || t('common.no_description') }}</div>
              </button>
            </div>
            <div v-if="!state.wizardPagedServiceModules.total" class="text-center text-muted small py-3">
              {{ t('configs_page.current_runtime_no_modules') }}
            </div>
            <div v-else class="d-flex justify-content-between align-items-center mt-3">
              <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedServiceModules.currentPage <= 1" @click="actions.changeWizardStagePage('service', state.wizardPagedServiceModules.currentPage - 1)">
                {{ t('common.previous') }}
              </button>
              <span class="small text-muted">{{ state.wizardPagedServiceModules.currentPage }} / {{ state.wizardPagedServiceModules.totalPages }}</span>
              <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedServiceModules.currentPage >= state.wizardPagedServiceModules.totalPages" @click="actions.changeWizardStagePage('service', state.wizardPagedServiceModules.currentPage + 1)">
                {{ t('common.next') }}
              </button>
            </div>
          </div>

          <div>
            <div class="d-flex justify-content-between align-items-center mb-2">
              <div>
                <div class="fw-semibold">{{ t('configs_page.wizard_parser_assets') }}</div>
                <div class="small text-muted">{{ t('configs_page.wizard_parser_assets_hint') }}</div>
              </div>
              <span class="badge text-bg-light">{{ state.wizardParserModuleIds.length }}</span>
            </div>
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input
                v-model="state.wizardParserSearch"
                type="text"
                class="form-control"
                :placeholder="t('configs_page.module_picker_search_placeholder')"
                @input="actions.changeWizardStagePage('parser', 1)"
              >
            </div>
            <div class="d-grid gap-2">
              <button
                v-for="module in state.wizardPagedParserModules.items"
                :key="module.id"
                type="button"
                class="btn text-start"
                :class="state.wizardParserModuleIds.includes(module.id) ? 'btn-primary' : 'btn-outline-secondary'"
                @click="actions.toggleWizardParserModule(module.id)"
              >
                <div class="fw-semibold">{{ module.name }}</div>
                <div class="small opacity-75">{{ module.description || t('common.no_description') }}</div>
              </button>
            </div>
            <div v-if="!state.wizardPagedParserModules.total" class="text-center text-muted small py-3">
              {{ t('configs_page.current_runtime_no_modules') }}
            </div>
            <div v-else class="d-flex justify-content-between align-items-center mt-3">
              <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedParserModules.currentPage <= 1" @click="actions.changeWizardStagePage('parser', state.wizardPagedParserModules.currentPage - 1)">
                {{ t('common.previous') }}
              </button>
              <span class="small text-muted">{{ state.wizardPagedParserModules.currentPage }} / {{ state.wizardPagedParserModules.totalPages }}</span>
              <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedParserModules.currentPage >= state.wizardPagedParserModules.totalPages" @click="actions.changeWizardStagePage('parser', state.wizardPagedParserModules.currentPage + 1)">
                {{ t('common.next') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <h6 class="mb-0">{{ t('configs_page.wizard_variables') }}</h6>
          <span class="badge text-bg-light">{{ variableGroupCount }}</span>
        </div>
        <div class="card-body">
          <div class="alert alert-info py-2">
            {{ t('configs_page.wizard_variable_scope_hint') }}
          </div>
          <div v-if="variableSections.length" class="d-grid gap-3">
            <details
              v-for="section in variableSections"
              :key="section.key"
              class="fm-variable-section border rounded-3 bg-light-subtle"
            >
              <summary class="fm-variable-section__summary">
                <div>
                  <div class="fw-semibold">{{ section.title }}</div>
                  <div class="small text-muted">{{ section.groups.length }}</div>
                </div>
                <span class="badge text-bg-light">{{ section.fieldCount }}</span>
              </summary>

              <div class="fm-variable-section__body d-grid gap-3">
                <details
                  v-for="(group, groupIndex) in section.groups"
                  :key="group.key"
                  class="fm-variable-group border rounded-3 bg-white"
                >
                  <summary class="fm-variable-group__summary">
                    <div>
                      <div class="fw-semibold">{{ group.title }}</div>
                      <div class="small text-muted">{{ group.subtitle }}</div>
                    </div>
                    <span class="badge text-bg-light">{{ group.fields.length }}</span>
                  </summary>

                  <div class="p-3 pt-0">
                    <div class="row g-3">
                      <div v-for="field in group.fields" :key="`${group.key}-${field.key}`" class="col-12">
                        <label class="form-label">{{ field.key }}</label>
                        <select v-if="field.kind === 'boolean'" v-model="group.model[field.key]" class="form-select">
                          <option value="true">true</option>
                          <option value="false">false</option>
                        </select>
                        <textarea
                          v-else-if="field.kind === 'json'"
                          v-model="group.model[field.key]"
                          class="form-control font-monospace"
                          rows="4"
                        ></textarea>
                        <input
                          v-else
                          v-model="group.model[field.key]"
                          type="text"
                          class="form-control"
                        >
                        <div v-if="field.description" class="small text-muted mt-1">{{ field.description }}</div>
                      </div>
                    </div>
                  </div>
                </details>
              </div>
            </details>
          </div>
          <div v-else class="text-center text-muted py-4">
            {{ t('configs_page.wizard_no_variables') }}
          </div>
        </div>
      </div>
    </div>

    <div class="col-xl-8">
      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header bg-white d-flex flex-wrap justify-content-between align-items-center gap-2">
          <div>
            <h6 class="mb-0">{{ t('configs_page.wizard_pipeline_workspace') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.wizard_pipeline_workspace_hint') }}</div>
          </div>
          <div class="d-flex flex-wrap gap-2 align-items-center">
            <div v-if="state.wizardCompatiblePipelines && state.wizardCompatiblePipelines.length" class="input-group input-group-sm" style="max-width:300px">
              <select v-model="selectedFromSavedPipelineId" class="form-select form-select-sm">
                <option value="">{{ t('configs_page.wizard_from_pipeline') }}…</option>
                <option v-for="p in state.wizardCompatiblePipelines" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
              <button
                class="btn btn-sm btn-outline-secondary"
                :disabled="!selectedFromSavedPipelineId"
                @click="doAddFromSavedPipeline"
              >{{ t('configs_page.wizard_add_pipeline_from') }}</button>
            </div>
            <button class="btn btn-primary btn-sm" @click="actions.addWizardPipeline">
              <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.wizard_add_pipeline') }}
            </button>
          </div>
        </div>
        <div class="card-body">
          <div class="row g-4">
            <div class="col-lg-4">
              <div class="fm-wizard-pipeline-list">
                <button
                  v-for="card in state.wizardPipelineCards"
                  :key="card.id"
                  type="button"
                  class="fm-wizard-pipeline-card text-start"
                  :class="{ active: state.activeWizardPipelineId === card.id }"
                  @click="actions.selectWizardPipeline(card.id)"
                >
                  <div class="d-flex justify-content-between align-items-start gap-2 mb-2">
                    <div>
                      <div class="fw-semibold">{{ pipelineLabel(card) }}</div>
                      <div class="small text-muted">{{ card.summary.path.length ? card.summary.path.join(' -> ') : t('configs_page.no_solution_path') }}</div>
                    </div>
                    <span class="badge" :class="card.complete ? 'bg-success-subtle text-success-emphasis' : 'bg-warning-subtle text-warning-emphasis'">
                      {{ card.complete ? t('configs_page.wizard_pipeline_complete') : t('configs_page.wizard_pipeline_incomplete') }}
                    </span>
                  </div>
                  <div class="small text-muted">
                    {{ t('configs_page.pipeline_stage_input') }} {{ card.inputModule ? 1 : 0 }}
                    · {{ t('configs_page.pipeline_stage_filter') }} {{ card.filterModules.length }}
                    <template v-if="state.wizardForm.fluent_type === 'fluentd'">· {{ t('configs_page.pipeline_stage_route', 'Route') }} {{ card.routeModules.length }}</template>
                    · {{ t('configs_page.pipeline_stage_output') }} {{ card.outputTargets.length }}
                  </div>
                </button>
              </div>
            </div>

            <div class="col-lg-8">
              <div v-if="state.activeWizardPipeline" class="fm-wizard-editor">
                <div class="d-flex flex-wrap justify-content-between align-items-start gap-2 mb-3">
                  <div class="flex-grow-1">
                    <label class="form-label">{{ t('configs_page.wizard_pipeline_name') }}</label>
                    <input
                      v-model="state.activeWizardPipeline.name"
                      type="text"
                      class="form-control"
                      :placeholder="t('configs_page.wizard_pipeline_name_placeholder')"
                    >
                  </div>
                  <div class="d-flex gap-2 align-self-end">
                    <button class="btn btn-outline-secondary btn-sm" @click="actions.duplicateWizardPipeline(state.activeWizardPipeline.id)">
                      <i class="bi bi-copy me-1"></i>{{ t('configs_page.wizard_duplicate_pipeline') }}
                    </button>
                    <button class="btn btn-outline-danger btn-sm" @click="actions.removeWizardPipeline(state.activeWizardPipeline.id)">
                      <i class="bi bi-trash me-1"></i>{{ t('configs_page.wizard_remove_pipeline') }}
                    </button>
                  </div>
                </div>

                <div class="card border-0 bg-light-subtle mb-3">
                  <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <div>
                        <div class="fw-semibold">{{ t('configs_page.pipeline_stage_input') }}</div>
                        <div class="small text-muted">{{ t('configs_page.wizard_input_group_hint') }}</div>
                      </div>
                      <span class="badge text-bg-light">{{ state.wizardPagedInputModules.total }}</span>
                    </div>
                    <div class="input-group mb-3">
                      <span class="input-group-text"><i class="bi bi-search"></i></span>
                      <input
                        v-model="state.wizardInputSearch"
                        type="text"
                        class="form-control"
                        :placeholder="t('configs_page.module_picker_search_placeholder')"
                        @input="actions.changeWizardStagePage('input', 1)"
                      >
                    </div>
                    <div class="row g-2">
                      <div v-for="module in state.wizardPagedInputModules.items" :key="module.id" class="col-md-6">
                        <button
                          type="button"
                          class="btn w-100 text-start"
                          :class="state.activeWizardPipeline.input?.module_id === module.id ? 'btn-primary' : 'btn-outline-secondary'"
                          @click="actions.setWizardPipelineInput(state.activeWizardPipeline.id, module.id)"
                        >
                          <div class="fw-semibold">{{ module.name }}</div>
                          <div class="small opacity-75">{{ module.description || t('common.no_description') }}</div>
                        </button>
                      </div>
                    </div>
                    <div class="d-flex justify-content-between align-items-center mt-3">
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedInputModules.currentPage <= 1" @click="actions.changeWizardStagePage('input', state.wizardPagedInputModules.currentPage - 1)">
                        {{ t('common.previous') }}
                      </button>
                      <span class="small text-muted">{{ state.wizardPagedInputModules.currentPage }} / {{ state.wizardPagedInputModules.totalPages }}</span>
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedInputModules.currentPage >= state.wizardPagedInputModules.totalPages" @click="actions.changeWizardStagePage('input', state.wizardPagedInputModules.currentPage + 1)">
                        {{ t('common.next') }}
                      </button>
                    </div>
                  </div>
                </div>

                <div class="card border-0 bg-light-subtle mb-3">
                  <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <div>
                        <div class="fw-semibold">{{ t('configs_page.pipeline_stage_filter') }}</div>
                        <div class="small text-muted">{{ t('configs_page.wizard_filter_group_hint') }}</div>
                      </div>
                      <span class="badge text-bg-light">{{ state.activeWizardPipeline.filters.length }}</span>
                    </div>
                    <div class="input-group mb-3">
                      <span class="input-group-text"><i class="bi bi-search"></i></span>
                      <input
                        v-model="state.wizardFilterSearch"
                        type="text"
                        class="form-control"
                        :placeholder="t('configs_page.module_picker_search_placeholder')"
                        @input="actions.changeWizardStagePage('filter', 1)"
                      >
                    </div>
                    <div class="row g-2 mb-3">
                      <div v-for="module in state.wizardPagedFilterModules.items" :key="module.id" class="col-md-6">
                        <button
                          type="button"
                          class="btn btn-outline-secondary w-100 text-start"
                          @click="actions.addWizardFilter(state.activeWizardPipeline.id, module.id)"
                        >
                          <div class="fw-semibold">{{ module.name }}</div>
                          <div class="small text-muted">{{ module.description || t('common.no_description') }}</div>
                          <div class="small text-primary mt-1">{{ t('configs_page.wizard_add_filter') }}</div>
                        </button>
                      </div>
                    </div>
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedFilterModules.currentPage <= 1" @click="actions.changeWizardStagePage('filter', state.wizardPagedFilterModules.currentPage - 1)">
                        {{ t('common.previous') }}
                      </button>
                      <span class="small text-muted">{{ state.wizardPagedFilterModules.currentPage }} / {{ state.wizardPagedFilterModules.totalPages }}</span>
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedFilterModules.currentPage >= state.wizardPagedFilterModules.totalPages" @click="actions.changeWizardStagePage('filter', state.wizardPagedFilterModules.currentPage + 1)">
                        {{ t('common.next') }}
                      </button>
                    </div>
                    <div v-if="state.activeWizardPipeline.filters.length" class="d-grid gap-2">
                      <div
                        v-for="(instance, index) in state.activeWizardPipeline.filters"
                        :key="instance.id"
                        class="border rounded-3 bg-white p-3"
                      >
                        <div class="d-flex justify-content-between align-items-start gap-2">
                          <div>
                            <div class="fw-semibold">{{ filterName(instance.module_id) }}</div>
                            <div class="small text-muted">{{ t('configs_page.wizard_filter_instance').replace('{index}', String(index + 1)) }}</div>
                          </div>
                          <div class="btn-group btn-group-sm">
                            <button class="btn btn-outline-secondary" :disabled="index === 0" @click="actions.moveWizardFilter(state.activeWizardPipeline.id, instance.id, 'up')">
                              <i class="bi bi-arrow-up"></i>
                            </button>
                            <button class="btn btn-outline-secondary" :disabled="index === state.activeWizardPipeline.filters.length - 1" @click="actions.moveWizardFilter(state.activeWizardPipeline.id, instance.id, 'down')">
                              <i class="bi bi-arrow-down"></i>
                            </button>
                            <button class="btn btn-outline-danger" @click="actions.removeWizardFilter(state.activeWizardPipeline.id, instance.id)">
                              <i class="bi bi-trash"></i>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-center text-muted py-3">
                      {{ t('configs_page.wizard_no_filters') }}
                    </div>
                  </div>
                </div>

                <div v-if="state.wizardForm.fluent_type === 'fluentd'" class="card border-0 bg-light-subtle mb-3">
                  <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <div>
                        <div class="fw-semibold">{{ t('configs_page.pipeline_stage_route', 'Route') }}</div>
                        <div class="small text-muted">{{ t('configs_page.wizard_route_group_hint', 'Optional Fluentd label routing / relabel / copy modules.') }}</div>
                      </div>
                      <span class="badge text-bg-light">{{ (state.activeWizardPipeline.routes || []).length }}</span>
                    </div>
                    <div class="input-group mb-3">
                      <span class="input-group-text"><i class="bi bi-search"></i></span>
                      <input
                        v-model="state.wizardRouteSearch"
                        type="text"
                        class="form-control"
                        :placeholder="t('configs_page.module_picker_search_placeholder')"
                        @input="actions.changeWizardStagePage('route', 1)"
                      >
                    </div>
                    <div class="row g-2 mb-3">
                      <div v-for="module in state.wizardPagedRouteModules.items" :key="module.id" class="col-md-6">
                        <button
                          type="button"
                          class="btn btn-outline-secondary w-100 text-start"
                          @click="actions.addWizardRoute(state.activeWizardPipeline.id, module.id)"
                        >
                          <div class="fw-semibold">{{ module.name }}</div>
                          <div class="small text-muted">{{ module.description || t('common.no_description') }}</div>
                          <div class="small text-primary mt-1">{{ t('configs_page.wizard_add_route') }}</div>
                        </button>
                      </div>
                    </div>
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedRouteModules.currentPage <= 1" @click="actions.changeWizardStagePage('route', state.wizardPagedRouteModules.currentPage - 1)">
                        {{ t('common.previous') }}
                      </button>
                      <span class="small text-muted">{{ state.wizardPagedRouteModules.currentPage }} / {{ state.wizardPagedRouteModules.totalPages }}</span>
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedRouteModules.currentPage >= state.wizardPagedRouteModules.totalPages" @click="actions.changeWizardStagePage('route', state.wizardPagedRouteModules.currentPage + 1)">
                        {{ t('common.next') }}
                      </button>
                    </div>
                    <div v-if="(state.activeWizardPipeline.routes || []).length" class="d-grid gap-2">
                      <div
                        v-for="(instance, index) in (state.activeWizardPipeline.routes || [])"
                        :key="instance.id"
                        class="border rounded-3 bg-white p-3"
                      >
                        <div class="d-flex justify-content-between align-items-start gap-2">
                          <div>
                            <div class="fw-semibold">{{ routeName(instance.module_id) }}</div>
                            <div class="small text-muted">{{ t('configs_page.wizard_route_instance').replace('{index}', String(index + 1)) }}</div>
                          </div>
                          <div class="btn-group btn-group-sm">
                            <button class="btn btn-outline-secondary" :disabled="index === 0" @click="actions.moveWizardRoute(state.activeWizardPipeline.id, instance.id, 'up')">
                              <i class="bi bi-arrow-up"></i>
                            </button>
                            <button class="btn btn-outline-secondary" :disabled="index === (state.activeWizardPipeline.routes || []).length - 1" @click="actions.moveWizardRoute(state.activeWizardPipeline.id, instance.id, 'down')">
                              <i class="bi bi-arrow-down"></i>
                            </button>
                            <button class="btn btn-outline-danger" @click="actions.removeWizardRoute(state.activeWizardPipeline.id, instance.id)">
                              <i class="bi bi-trash"></i>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-center text-muted py-3">
                      {{ t('configs_page.wizard_no_routes') }}
                    </div>
                  </div>
                </div>

                <div class="card border-0 bg-light-subtle">
                  <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <div>
                        <div class="fw-semibold">{{ t('configs_page.pipeline_stage_output') }}</div>
                        <div class="small text-muted">{{ t('configs_page.wizard_output_group_hint') }}</div>
                      </div>
                      <span class="badge text-bg-light">{{ state.activeWizardPipeline.outputs.length }}</span>
                    </div>
                    <div class="input-group mb-3">
                      <span class="input-group-text"><i class="bi bi-search"></i></span>
                      <input
                        v-model="state.wizardOutputSearch"
                        type="text"
                        class="form-control"
                        :placeholder="t('configs_page.module_picker_search_placeholder')"
                        @input="actions.changeWizardStagePage('output', 1)"
                      >
                    </div>
                    <div class="row g-2 mb-3">
                      <div v-for="target in state.wizardPagedOutputTargets.items" :key="target.id" class="col-md-6">
                        <button
                          type="button"
                          class="btn btn-outline-secondary w-100 text-start"
                          @click="actions.addWizardOutputTarget(state.activeWizardPipeline.id, target.id)"
                        >
                          <div class="d-flex justify-content-between align-items-start gap-1 mb-1">
                            <div class="fw-semibold">{{ target.name }}</div>
                            <span v-if="target._is_aggregation_group" class="badge bg-info-subtle text-info-emphasis flex-shrink-0" style="font-size:0.65em">{{ t('configs_page.aggregation_group', 'Aggregation Group') }}</span>
                          </div>
                          <div class="small text-muted">{{ target.target_type }} · {{ target.endpoint || t('common.unspecified') }}</div>
                          <div class="small text-primary mt-1">{{ t('configs_page.wizard_add_output') }}</div>
                        </button>
                      </div>
                    </div>
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedOutputTargets.currentPage <= 1" @click="actions.changeWizardStagePage('output', state.wizardPagedOutputTargets.currentPage - 1)">
                        {{ t('common.previous') }}
                      </button>
                      <span class="small text-muted">{{ state.wizardPagedOutputTargets.currentPage }} / {{ state.wizardPagedOutputTargets.totalPages }}</span>
                      <button class="btn btn-sm btn-outline-secondary" :disabled="state.wizardPagedOutputTargets.currentPage >= state.wizardPagedOutputTargets.totalPages" @click="actions.changeWizardStagePage('output', state.wizardPagedOutputTargets.currentPage + 1)">
                        {{ t('common.next') }}
                      </button>
                    </div>
                    <div v-if="state.activeWizardPipeline.outputs.length" class="d-grid gap-2">
                      <div
                        v-for="(instance, index) in state.activeWizardPipeline.outputs"
                        :key="instance.id"
                        class="border rounded-3 bg-white p-3"
                      >
                        <div class="d-flex justify-content-between align-items-start gap-2">
                          <div>
                            <div class="d-flex align-items-center gap-2">
                              <div class="fw-semibold">{{ outputName(instance.target_id) }}</div>
                              <span v-if="outputIsAggGroup(instance.target_id)" class="badge bg-info-subtle text-info-emphasis" style="font-size:0.65em">{{ t('configs_page.aggregation_group', 'Aggregation Group') }}</span>
                            </div>
                            <div class="small text-muted">{{ t('configs_page.wizard_output_instance').replace('{index}', String(index + 1)) }}</div>
                          </div>
                          <div class="d-flex align-items-center gap-1">
                            <button
                              v-if="outputSettingsFields(instance.target_id, instance).length"
                              class="btn btn-sm btn-outline-secondary"
                              :title="t('configs_page.wizard_output_params')"
                              @click="toggleOutputExpand(instance.id)"
                            >
                              <i class="bi" :class="expandedOutputIds.has(instance.id) ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
                            </button>
                            <div class="btn-group btn-group-sm">
                              <button class="btn btn-outline-secondary" :disabled="index === 0" @click="actions.moveWizardOutput(state.activeWizardPipeline.id, instance.id, 'up')">
                                <i class="bi bi-arrow-up"></i>
                              </button>
                              <button class="btn btn-outline-secondary" :disabled="index === state.activeWizardPipeline.outputs.length - 1" @click="actions.moveWizardOutput(state.activeWizardPipeline.id, instance.id, 'down')">
                                <i class="bi bi-arrow-down"></i>
                              </button>
                              <button class="btn btn-outline-danger" @click="actions.removeWizardOutput(state.activeWizardPipeline.id, instance.id)">
                                <i class="bi bi-trash"></i>
                              </button>
                            </div>
                          </div>
                        </div>
                        <!-- Inline destination parameter editor -->
                        <div v-if="expandedOutputIds.has(instance.id) && outputSettingsFields(instance.target_id, instance).length" class="mt-3 pt-3 border-top">
                          <div class="small fw-semibold text-muted mb-2">{{ t('configs_page.wizard_output_params') }}</div>
                          <div class="row g-2">
                            <div
                              v-for="[key] in outputSettingsFields(instance.target_id, instance)"
                              :key="key"
                              class="col-md-6"
                            >
                              <label class="form-label small mb-1 text-truncate d-block" :title="key">{{ key }}</label>
                              <input
                                :value="instance.variables != null && Object.prototype.hasOwnProperty.call(instance.variables, key) ? instance.variables[key] : ''"
                                type="text"
                                class="form-control form-control-sm font-monospace"
                                @input="actions.updateWizardOutputVariable(state.activeWizardPipeline.id, instance.id, key, $event.target.value)"
                              >
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-center text-muted py-3">
                      {{ t('configs_page.wizard_no_outputs') }}
                    </div>
                  </div>
                </div>
              </div>

              <div v-else class="text-center text-muted py-5">
                {{ t('configs_page.wizard_pipeline_empty') }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header bg-white">
          <h6 class="mb-0">{{ t('configs_page.wizard_summary') }}</h6>
        </div>
        <div class="card-body">
          <div class="d-flex flex-wrap gap-2 mb-3">
            <span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(state.wizardForm.fluent_type) }}</span>
            <span class="badge text-bg-light">{{ t('configs_page.wizard_pipeline_count').replace('{count}', String(state.wizardPipelines.length)) }}</span>
            <span class="badge text-bg-light">{{ t('configs_page.wizard_renderable_pipeline_count').replace('{count}', String(state.wizardRenderSummary.pipelineCount)) }}</span>
            <span class="badge text-bg-light">{{ t('configs_page.wizard_output_count').replace('{count}', String(state.wizardRenderSummary.outputCount)) }}</span>
          </div>
          <div class="small text-muted mb-3">{{ t('configs_page.wizard_summary_hint') }}</div>
          <div v-if="state.wizardIncompletePipelineLabels.length" class="alert alert-warning py-2">
            {{ t('configs_page.wizard_incomplete_pipelines').replace('{items}', state.wizardIncompletePipelineLabels.join(', ')) }}
          </div>
          <div v-if="state.wizardOutputResolutionWarnings.length" class="alert alert-warning py-2">
            <div class="fw-semibold mb-1">{{ t('configs_page.output_target_module_missing_title') }}</div>
            <div v-for="item in state.wizardOutputResolutionWarnings" :key="`${item.pipeline}-${item.target}`">
              {{ item.pipeline }}: {{ item.target }}
            </div>
          </div>
          <div class="d-grid gap-2">
            <button class="btn btn-success" @click="actions.runWizardPreview">
              <i class="bi bi-magic me-1"></i>{{ t('configs_page.generate_wizard_preview') }}
            </button>
            <button class="btn btn-outline-primary" :disabled="!state.renderedConfig" @click="actions.saveWizardAsTemplate">
              <i class="bi bi-file-earmark-plus me-1"></i>{{ state.wizardSaveButtonLabel || t('configs_page.save_wizard_template') }}
            </button>
            <button class="btn btn-outline-secondary" @click="actions.openAdvancedPreviewFromWizard">
              <i class="bi bi-sliders me-1"></i>{{ t('configs_page.open_advanced_preview') }}
            </button>
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-0">{{ t('configs_page.wizard_preview') }}</h6>
            <div class="small text-muted mt-1">{{ t('configs_page.wizard_preview_hint') }}</div>
          </div>
          <div v-if="state.renderedConfig" class="text-end">
            <div class="small text-muted">{{ t('configs_page.render_hash') }}</div>
            <code>{{ state.renderedConfig.hash }}</code>
          </div>
        </div>
        <div class="card-body">
          <div v-if="state.renderedConfig">
            <div class="row g-3 mb-3">
              <div v-for="card in state.wizardPipelineCards" :key="card.id" class="col-lg-6">
                <div class="border rounded-3 p-3 bg-light-subtle h-100">
                  <div class="d-flex justify-content-between align-items-start gap-2 mb-2">
                    <div>
                      <div class="fw-semibold">{{ pipelineLabel(card) }}</div>
                      <div class="small text-muted">{{ card.summary.path.length ? card.summary.path.join(' -> ') : t('configs_page.no_solution_path') }}</div>
                    </div>
                    <span class="badge" :class="card.complete ? 'bg-success-subtle text-success-emphasis' : 'bg-warning-subtle text-warning-emphasis'">
                      {{ card.complete ? t('configs_page.wizard_pipeline_complete') : t('configs_page.wizard_pipeline_incomplete') }}
                    </span>
                  </div>
                  <div class="small text-muted">
                    {{ t('configs_page.pipeline_stage_input') }}: {{ card.inputModule?.name || '-' }}
                  </div>
                  <div class="small text-muted">
                    {{ t('configs_page.pipeline_stage_filter') }}: {{ card.filterModules.length ? card.filterModules.map((item) => item.name).join(' -> ') : t('configs_page.no_processors') }}
                  </div>
                  <div v-if="state.wizardForm.fluent_type === 'fluentd'" class="small text-muted">
                    {{ t('configs_page.pipeline_stage_route', 'Route') }}: {{ card.routeModules.length ? card.routeModules.map((item) => item.name).join(' -> ') : '-' }}
                  </div>
                  <div class="small text-muted">
                    {{ t('configs_page.pipeline_stage_output') }}: {{ card.outputTargets.length ? card.outputTargets.map((item) => item.name).join(' + ') : '-' }}
                  </div>
                </div>
              </div>
            </div>
            <pre class="fm-render-preview">{{ state.renderedConfig.content }}</pre>
          </div>
          <div v-else class="text-center text-muted py-5">
            {{ t('configs_page.wizard_preview_empty') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, unref, watch } from 'vue'
import { useI18n } from '../../../i18n'

const props = defineProps({
  state: {
    type: Object,
    required: true,
  },
  actions: {
    type: Object,
    required: true,
  },
  helpers: {
    type: Object,
    required: true,
  },
})

const { t } = useI18n()

const selectedLoadTemplateId = ref('')
const selectedFromSavedPipelineId = ref('')
const loadTemplateSearch = ref('')
const loadRuntimeFilter = ref('current')

const runtimeFilters = computed(() => {
  const runtimes = new Set(
    (unref(props.state.wizardAssemblyTemplates) || [])
      .map((tpl) => String(tpl?.fluent_type || '').trim())
      .filter(Boolean)
  )
  return Array.from(runtimes)
})

const loadableTemplateTotal = computed(() => (unref(props.state.wizardAssemblyTemplates) || []).length)

const filteredLoadTemplates = computed(() => {
  const templates = unref(props.state.wizardAssemblyTemplates) || []
  const keyword = String(loadTemplateSearch.value || '').trim().toLowerCase()
  const currentRuntime = String(unref(props.state.wizardForm)?.fluent_type || '').trim().toLowerCase()

  return templates.filter((tpl) => {
    const runtime = String(tpl?.fluent_type || '').trim().toLowerCase()
    if (loadRuntimeFilter.value === 'current' && currentRuntime && runtime !== currentRuntime) {
      return false
    }
    if (loadRuntimeFilter.value !== 'current' && loadRuntimeFilter.value !== 'all' && runtime !== String(loadRuntimeFilter.value).trim().toLowerCase()) {
      return false
    }
    if (!keyword) return true
    const haystack = [
      tpl?.name,
      tpl?.description,
      tpl?.fluent_type,
    ]
      .filter(Boolean)
      .join('\n')
      .toLowerCase()
    return haystack.includes(keyword)
  })
})

const selectedLoadTemplate = computed(() =>
  (unref(props.state.wizardAssemblyTemplates) || []).find((item) => String(item.id) === String(selectedLoadTemplateId.value)) || null
)

function doLoadFromTemplate() {
  const tpl = selectedLoadTemplate.value
  if (!tpl) return
  props.actions.loadWizardFromTemplate(tpl)
  selectedLoadTemplateId.value = ''
  loadTemplateSearch.value = ''
}

function doAddFromSavedPipeline() {
  if (!selectedFromSavedPipelineId.value) return
  props.actions.addWizardPipelineFromSaved(selectedFromSavedPipelineId.value)
  selectedFromSavedPipelineId.value = ''
}

const allVariableGroups = computed(() => [
  ...(unref(props.state.wizardGlobalVariableGroups) || []),
  ...(unref(props.state.wizardPipelineVariableGroups) || []),
])

const variableSections = computed(() => {
  const sections = []
  const sectionMap = new Map()

  for (const group of allVariableGroups.value) {
    const key = group.sectionKey || 'default'
    if (!sectionMap.has(key)) {
      const section = {
        key,
        title: group.sectionTitle || t('configs_page.wizard_variables'),
        kind: group.sectionKind || 'default',
        ref: group.sectionRef || '',
        groups: [],
        fieldCount: 0,
      }
      sectionMap.set(key, section)
      sections.push(section)
    }
    const section = sectionMap.get(key)
    section.groups.push(group)
    section.fieldCount += group.fields.length
  }

  return sections
})

const variableGroupCount = computed(() =>
  allVariableGroups.value.reduce((total, group) => total + group.fields.length, 0)
)

const globalResourceCount = computed(() => {
  let total = 0
  if (unref(props.state.wizardServiceModuleId)) total += 1
  total += (unref(props.state.wizardParserModuleIds) || []).length
  return total
})

function pipelineLabel(card) {
  return props.helpers.wizardPipelineDisplayName(card, card.index)
}

function filterName(moduleId) {
  const module = (unref(props.state.wizardPagedFilterModules)?.items || [])
    .concat((unref(props.state.wizardPipelineCards) || []).flatMap((card) => card.filterModules || []))
    .find((item) => item.id === moduleId)
  return module?.name || t('configs_page.pipeline_stage_filter')
}

function routeName(moduleId) {
  const module = (unref(props.state.wizardPagedRouteModules)?.items || [])
    .concat((unref(props.state.wizardPipelineCards) || []).flatMap((card) => card.routeModules || []))
    .find((item) => item.id === moduleId)
  return module?.name || t('configs_page.pipeline_stage_route', 'Route')
}

function outputName(targetId) {
  const target = (unref(props.state.wizardPagedOutputTargets)?.items || [])
    .concat((unref(props.state.wizardPipelineCards) || []).flatMap((card) => card.outputTargets || []))
    .find((item) => item.id === targetId)
  return target?.name || t('configs_page.pipeline_stage_output')
}

function outputIsAggGroup(targetId) {
  const target = (unref(props.state.wizardPagedOutputTargets)?.items || [])
    .concat((unref(props.state.wizardPipelineCards) || []).flatMap((card) => card.outputTargets || []))
    .find((item) => item.id === targetId)
  return Boolean(target?._is_aggregation_group)
}

// --- Inline output parameter editor ---

const expandedOutputIds = ref(new Set())

function toggleOutputExpand(instanceId) {
  const next = new Set(expandedOutputIds.value)
  if (next.has(instanceId)) {
    next.delete(instanceId)
  } else {
    next.add(instanceId)
  }
  expandedOutputIds.value = next
}

// Per output type: which settings fields vary per environment.
// Host/port/credentials stay with the OutputTarget definition and are not shown here.
const OUTPUT_ENV_FIELDS = {
  opensearch:    ['index', 'index_name', 'logstash_prefix', 'logstash_dateformat'],
  elasticsearch: ['index', 'index_name', 'logstash_prefix', 'logstash_dateformat'],
  kafka:         ['topics', 'topic_key'],
  loki:          ['labels', 'tenant_id', 'line_format'],
  s3:            ['bucket', 's3_bucket', 's3_key_format'],
  splunk:        ['splunk_index', 'splunk_sourcetype', 'splunk_source'],
  datadog:       ['dd_service', 'dd_source', 'dd_tags'],
  http:          ['uri'],
}

function outputSettingsFields(targetId, instance) {
  const target = (unref(props.state.wizardPagedOutputTargets)?.items || [])
    .concat((unref(props.state.wizardPipelineCards) || []).flatMap((card) => card.outputTargets || []))
    .find((item) => item.id === targetId)
  if (!target) return []
  const targetType = String(target.target_type || '').toLowerCase()
  const allowedKeys = OUTPUT_ENV_FIELDS[targetType]
  // Source: instance.variables (module defaults + target settings already merged)
  const variables = instance?.variables || {}
  const entries = Object.entries(variables).filter(([key]) => !key.startsWith('output_target_'))
  if (allowedKeys !== undefined) {
    return entries.filter(([key]) => allowedKeys.includes(key))
  }
  // Unknown type: show non-boolean, non-match fields as fallback
  return entries.filter(([key, value]) => {
    if (key === 'match') return false
    const lower = String(value).trim().toLowerCase()
    return !['on', 'off', 'true', 'false', 'yes', 'no'].includes(lower)
  })
}

// Auto-expand newly added output instances
watch(
  () => (unref(props.state.activeWizardPipeline)?.outputs || []).map((o) => o.id),
  (newIds, oldIds) => {
    const oldSet = new Set(oldIds || [])
    const next = new Set(expandedOutputIds.value)
    for (const id of newIds) {
      if (!oldSet.has(id)) next.add(id)
    }
    expandedOutputIds.value = next
  }
)
</script>

<style scoped>
.fm-template-picker__list {
  max-height: 18rem;
  overflow: auto;
}

.fm-template-picker__item {
  border-width: 1px;
}

.fm-template-picker__description {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.fm-template-picker__detail,
.fm-template-picker__empty {
  min-height: 100%;
}

.fm-template-picker__empty {
  min-height: 10rem;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: var(--bs-secondary-color);
  padding: 1rem;
}

.fm-variable-section__summary,
.fm-variable-group__summary {
  list-style: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  cursor: pointer;
  padding: 1rem;
}

.fm-variable-section__summary::-webkit-details-marker,
.fm-variable-group__summary::-webkit-details-marker {
  display: none;
}

.fm-variable-section__body {
  padding: 0 1rem 1rem;
}
</style>
