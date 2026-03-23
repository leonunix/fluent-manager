<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('configs_page.title') }}</h4>
        <div class="text-muted">{{ t('configs_page.subtitle') }}</div>
      </div>
      <div class="d-flex gap-2">
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
          class="btn btn-primary"
          @click="openCreateTemplate"
        >
          <i class="bi bi-plus-lg me-1"></i>{{ t('configs_page.create_template') }}
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
            <span class="badge rounded-pill text-bg-light ms-2">{{ modules.length }}</span>
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

    <div v-if="activeTab === 'templates'" class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('common.runtime') }}</th>
                <th>{{ t('common.description') }}</th>
                <th>{{ t('deploys_page.creator') }}</th>
                <th>{{ t('deploys_page.created_at') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tpl in templates" :key="tpl.id">
                <td>
                  <router-link :to="`/configs/${tpl.id}`" class="text-decoration-none fw-semibold">
                    {{ tpl.name }}
                  </router-link>
                </td>
                <td><span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(tpl.fluent_type) }}</span></td>
                <td>{{ tpl.description || '-' }}</td>
                <td>{{ tpl.creator?.username || '-' }}</td>
                <td>{{ formatTime(tpl.created_at) }}</td>
                <td>
                  <router-link :to="`/configs/${tpl.id}`" class="btn btn-sm btn-outline-primary me-1">
                    <i class="bi bi-eye"></i>
                  </router-link>
                  <button class="btn btn-sm btn-outline-danger" @click="handleDeleteTemplate(tpl)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!templates.length">
                <td colspan="6" class="text-center text-muted py-4">{{ t('configs_page.no_templates') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'wizard'" class="row g-4">
      <div class="col-xl-4">
        <div class="card border-0 shadow-sm mb-4">
          <div class="card-header bg-white">
            <h6 class="mb-0">{{ t('configs_page.wizard_basics') }}</h6>
          </div>
          <div class="card-body">
            <div class="alert alert-info py-2">
              {{ t('configs_page.wizard_intro') }}
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.wizard_goal') }}</label>
              <select v-model="wizardForm.goal" class="form-select">
                <option value="edge_collection">{{ t('configs_page.goal_edge_collection') }}</option>
                <option value="central_aggregation">{{ t('configs_page.goal_central_aggregation') }}</option>
                <option value="custom_pipeline">{{ t('configs_page.goal_custom_pipeline') }}</option>
              </select>
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="wizardForm.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('configs_page.target_version') }}</label>
                <input v-model="wizardForm.runtime_version" type="text" class="form-control" placeholder="3.1.0 / 1.16">
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('common.name') }}</label>
              <input v-model="wizardForm.name" type="text" class="form-control" :placeholder="t('configs_page.wizard_name_placeholder')">
            </div>
            <div class="mb-0">
              <label class="form-label">{{ t('common.description') }}</label>
              <textarea v-model="wizardForm.description" class="form-control" rows="3" :placeholder="t('configs_page.wizard_description_placeholder')"></textarea>
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm mb-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">{{ t('configs_page.wizard_variables') }}</h6>
            <span class="badge text-bg-light">{{ wizardVariableFields.length }}</span>
          </div>
          <div class="card-body">
            <div v-if="wizardVariableFields.length" class="row g-3">
              <div v-for="field in wizardVariableFields" :key="field.key" class="col-12">
                <label class="form-label d-flex justify-content-between align-items-center">
                  <span>{{ field.key }}</span>
                  <span class="small text-muted">{{ field.moduleNames.join(', ') }}</span>
                </label>
                <select v-if="field.kind === 'boolean'" v-model="wizardVariableValues[field.key]" class="form-select">
                  <option value="true">true</option>
                  <option value="false">false</option>
                </select>
                <textarea
                  v-else-if="field.kind === 'json'"
                  v-model="wizardVariableValues[field.key]"
                  class="form-control font-monospace"
                  rows="4"
                ></textarea>
                <input
                  v-else
                  v-model="wizardVariableValues[field.key]"
                  type="text"
                  class="form-control"
                >
                <div v-if="field.description" class="small text-muted mt-1">{{ field.description }}</div>
              </div>
            </div>
            <div v-else class="text-center text-muted py-4">
              {{ t('configs_page.wizard_no_variables') }}
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white">
            <h6 class="mb-0">{{ t('configs_page.wizard_summary') }}</h6>
          </div>
          <div class="card-body">
            <div class="d-flex flex-wrap gap-2 mb-3">
              <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(wizardForm.fluent_type) }}</span>
              <span class="badge text-bg-light">{{ wizardSelectedModules.length }} {{ t('configs_page.modules') }}</span>
              <span class="badge text-bg-light">{{ wizardCoverageCount }}/{{ wizardRecommendedTypes.length }}</span>
            </div>
            <div class="small text-muted mb-3">{{ t('configs_page.wizard_summary_hint') }}</div>
            <div v-if="wizardMissingTypes.length" class="alert alert-warning py-2">
              {{ t('configs_page.wizard_missing_types').replace('{types}', wizardMissingTypes.join(', ')) }}
            </div>
            <button class="btn btn-success w-100" @click="runWizardPreview">
              <i class="bi bi-magic me-1"></i>{{ t('configs_page.generate_wizard_preview') }}
            </button>
            <button class="btn btn-outline-primary w-100 mt-2" :disabled="!renderedConfig" @click="saveWizardAsTemplate">
              <i class="bi bi-file-earmark-plus me-1"></i>{{ t('configs_page.save_wizard_template') }}
            </button>
            <button class="btn btn-outline-secondary w-100 mt-2" @click="openAdvancedPreviewFromWizard">
              <i class="bi bi-sliders me-1"></i>{{ t('configs_page.open_advanced_preview') }}
            </button>
          </div>
        </div>
      </div>

      <div class="col-xl-8">
        <div class="card border-0 shadow-sm mb-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.wizard_module_picker') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.runtime_help').replace('{runtime}', runtimeLabel(wizardForm.fluent_type)) }}</div>
            </div>
            <span class="badge text-bg-light">{{ wizardEligibleModules.length }}</span>
          </div>
          <div class="card-body">
            <div class="row g-4">
              <div v-for="group in wizardModulesByType" :key="group.type" class="col-12">
                <div class="d-flex justify-content-between align-items-center mb-2">
                  <div>
                    <h6 class="mb-0 text-capitalize">{{ group.type }}</h6>
                    <div class="small text-muted">{{ t('configs_page.wizard_group_hint').replace('{type}', group.type) }}</div>
                  </div>
                  <span class="badge text-bg-light">{{ group.modules.length }}</span>
                </div>
                <div class="row g-3">
                  <div v-for="module in group.modules" :key="module.id" class="col-lg-6">
                    <label class="fm-module-choice h-100" :class="{ selected: selectedWizardModuleIds.includes(module.id) }">
                      <input
                        :checked="selectedWizardModuleIds.includes(module.id)"
                        type="checkbox"
                        class="form-check-input"
                        @change="toggleWizardModule(module.id)"
                      >
                      <div class="flex-grow-1">
                        <div class="d-flex align-items-center gap-2 mb-2 flex-wrap">
                          <span class="fw-semibold">{{ module.name }}</span>
                          <span class="badge text-bg-secondary">{{ module.module_type }}</span>
                          <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(module.fluent_type) }}</span>
                          <span v-if="module.is_builtin" class="badge text-bg-dark">{{ t('configs_page.builtin') }}</span>
                        </div>
                        <div class="small text-muted mb-2">{{ module.description || t('common.no_description') }}</div>
                        <pre class="fm-module-snippet">{{ module.content }}</pre>
                      </div>
                    </label>
                  </div>
                </div>
              </div>
              <div v-if="!wizardEligibleModules.length" class="col-12 text-center text-muted py-5">
                {{ t('configs_page.current_runtime_no_modules') }}
              </div>
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.wizard_preview') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.wizard_preview_hint') }}</div>
            </div>
            <div v-if="renderedConfig" class="text-end">
              <div class="small text-muted">{{ t('configs_page.render_hash') }}</div>
              <code>{{ renderedConfig.hash }}</code>
            </div>
          </div>
          <div class="card-body">
            <div v-if="renderedConfig">
              <div class="d-flex flex-wrap gap-2 mb-3">
                <span class="badge bg-success-subtle text-success-emphasis">{{ wizardForm.goal }}</span>
                <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(renderedConfig.fluent_type) }}</span>
                <span class="badge text-bg-light">{{ renderedConfig.runtime_version || t('common.unspecified') }}</span>
              </div>
              <pre class="fm-render-preview">{{ renderedConfig.content }}</pre>
            </div>
            <div v-else class="text-center text-muted py-5">
              {{ t('configs_page.wizard_preview_empty') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'assistant'" class="row g-4">
      <div class="col-xl-5">
        <div class="card border-0 shadow-sm mb-4">
          <div class="card-header bg-white">
            <h6 class="mb-0">{{ t('configs_page.ai_assistant') }}</h6>
          </div>
          <div class="card-body">
            <div class="alert alert-info py-2">
              {{ t('configs_page.ai_assistant_intro') }}
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="aiAssistantForm.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('configs_page.module_type_coverage') }}</label>
                <select v-model="aiAssistantForm.module_type" class="form-select">
                  <option v-for="type in moduleTypes" :key="type" :value="type">{{ type }}</option>
                </select>
              </div>
              <div class="col-md-12">
                <label class="form-label">{{ t('configs_page.ai_assistant_goal') }}</label>
                <select v-model="aiAssistantForm.goal" class="form-select">
                  <option value="module">{{ t('configs_page.ai_assistant_goal_module') }}</option>
                  <option value="template">{{ t('configs_page.ai_assistant_goal_template') }}</option>
                  <option value="both">{{ t('configs_page.ai_assistant_goal_both') }}</option>
                </select>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.sample_log') }}</label>
              <textarea
                v-model="aiAssistantForm.sample"
                class="form-control font-monospace"
                rows="12"
                :placeholder="t('configs_page.ai_assistant_sample_placeholder')"
              ></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('common.description') }}</label>
              <textarea
                v-model="aiAssistantForm.extra_requirements"
                class="form-control"
                rows="4"
                :placeholder="t('configs_page.ai_assistant_requirements_placeholder')"
              ></textarea>
            </div>
            <button class="btn btn-success w-100" :disabled="aiAssistantLoading || !aiAssistantForm.sample.trim()" @click="runAIAssistant">
              <i class="bi bi-stars me-1"></i>{{ aiAssistantLoading ? t('configs_page.ai_assistant_running') : t('configs_page.ai_assistant_run') }}
            </button>
          </div>
        </div>
      </div>

      <div class="col-xl-7">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.ai_assistant_result') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.ai_assistant_result_hint') }}</div>
            </div>
            <div v-if="aiAssistantResult" class="d-flex gap-2">
              <button class="btn btn-sm btn-outline-primary" @click="useAIModuleDraft">
                <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.ai_use_module') }}
              </button>
              <button class="btn btn-sm btn-outline-primary" @click="useAITemplateDraft">
                <i class="bi bi-box-arrow-in-down-right me-1"></i>{{ t('configs_page.ai_use_template') }}
              </button>
            </div>
          </div>
          <div class="card-body">
            <div
              v-if="aiAssistantLoading || aiAssistantFeedback.message"
              class="fm-ai-assistant-feedback mb-3"
              :class="{
                'is-success': aiAssistantFeedback.type === 'success',
                'is-danger': aiAssistantFeedback.type === 'danger',
              }"
            >
              <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
                <div>
                  <div class="fw-semibold">
                    {{ aiAssistantLoading ? t('configs_page.ai_assistant_running') : aiAssistantFeedback.message }}
                  </div>
                  <div
                    v-if="!aiAssistantLoading && aiAssistantFeedback.detail && aiAssistantFeedback.detail !== aiAssistantFeedback.message"
                    class="small text-muted mt-1"
                  >
                    {{ aiAssistantFeedback.detail }}
                  </div>
                </div>
                <div v-if="!aiAssistantLoading && aiAssistantFeedback.provider" class="small text-muted text-nowrap">
                  {{ aiAssistantFeedback.provider }}
                </div>
              </div>
              <div
                v-if="!aiAssistantLoading && aiAssistantFeedback.providerDetail"
                class="small text-muted mt-2"
              >
                {{ t('configs_page.ai_provider_feedback') }}: {{ aiAssistantFeedback.providerDetail }}
              </div>
            </div>

            <div v-if="aiAssistantResult">
              <div class="d-flex flex-wrap gap-2 mb-3">
                <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(aiAssistantForm.fluent_type) }}</span>
                <span class="badge text-bg-light">{{ aiAssistantResult.provider }}</span>
                <span class="badge text-bg-light">{{ aiAssistantResult.account_name }}</span>
              </div>

              <div class="row g-3 mb-3">
                <div class="col-md-6">
                  <div class="fm-ai-result-box">
                    <div class="fm-ai-result-box__label">{{ t('configs_page.ai_detected_format') }}</div>
                    <div>{{ aiAssistantResult.detected_format || '-' }}</div>
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="fm-ai-result-box">
                    <div class="fm-ai-result-box__label">{{ t('configs_page.ai_summary') }}</div>
                    <div>{{ aiAssistantResult.summary || '-' }}</div>
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.variables_json') }}</label>
                <textarea class="form-control font-monospace fm-config-textarea" rows="7" readonly :value="aiAssistantResult.variables_json"></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.version_content') }}</label>
                <textarea class="form-control font-monospace fm-config-textarea" rows="10" readonly :value="aiAssistantResult.module_content"></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.template_content') }}</label>
                <textarea class="form-control font-monospace fm-config-textarea" rows="10" readonly :value="aiAssistantResult.template_content"></textarea>
              </div>

              <div class="row g-3">
                <div class="col-md-6">
                  <div class="fm-ai-result-box">
                    <div class="fm-ai-result-box__label">{{ t('configs_page.ai_assembly_steps') }}</div>
                    <ul class="mb-0">
                      <li v-for="(step, index) in aiAssistantResult.assembly_steps || []" :key="index">{{ step }}</li>
                    </ul>
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="fm-ai-result-box">
                    <div class="fm-ai-result-box__label">{{ t('configs_page.ai_notes') }}</div>
                    <ul class="mb-0">
                      <li v-for="(note, index) in aiAssistantResult.notes || []" :key="index">{{ note }}</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
            <div v-else-if="!aiAssistantLoading && !aiAssistantFeedback.message" class="text-center text-muted py-5">
              {{ t('configs_page.ai_assistant_empty') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'modules'">
      <div class="row g-4 mb-4">
        <div class="col-md-4">
          <div class="card border-0 shadow-sm h-100">
            <div class="card-body">
              <div class="text-muted small mb-1">{{ t('configs_page.total_modules') }}</div>
              <div class="fs-3 fw-bold">{{ modules.length }}</div>
              <div class="small text-muted mt-2">{{ t('configs_page.source_runtime_hint') }}</div>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card border-0 shadow-sm h-100">
            <div class="card-body">
              <div class="text-muted small mb-1">{{ t('configs_page.shared_modules') }}</div>
              <div class="fs-3 fw-bold">{{ sharedModuleCount }}</div>
              <div class="small text-muted mt-2">{{ t('configs_page.source_runtime_hint') }}</div>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card border-0 shadow-sm h-100">
            <div class="card-body">
              <div class="text-muted small mb-1">{{ t('configs_page.module_type_coverage') }}</div>
              <div class="fs-3 fw-bold">{{ usedModuleTypes.length }}/6</div>
              <div class="small text-muted mt-2">{{ usedModuleTypes.join(' / ') || '-' }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-body p-0">
          <div class="table-responsive">
            <table class="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>{{ t('common.name') }}</th>
                  <th>{{ t('configs_page.module_type_coverage') }}</th>
                  <th>{{ t('common.runtime') }}</th>
                  <th>{{ t('configs_page.builtin') }}</th>
                  <th>{{ t('configs_page.variables_json') }}</th>
                  <th>{{ t('deploys_page.created_at') }}</th>
                  <th>{{ t('actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="module in modules" :key="module.id">
                  <td>
                    <div class="fw-semibold">{{ module.name }}</div>
                    <div class="small text-muted">{{ module.description || t('common.no_description') }}</div>
                  </td>
                  <td><span class="badge text-bg-secondary">{{ module.module_type }}</span></td>
                  <td><span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(module.fluent_type) }}</span></td>
                  <td>
                    <span :class="module.is_builtin ? 'badge text-bg-dark' : 'badge text-bg-light'">
                      {{ module.is_builtin ? t('configs_page.builtin') : t('configs_page.custom') }}
                    </span>
                  </td>
                  <td>
                    <code class="small">{{ shortVariables(module.variables) }}</code>
                  </td>
                  <td>{{ formatTime(module.created_at) }}</td>
                  <td>
                    <button class="btn btn-sm btn-outline-primary me-1" @click="openEditModule(module)">
                      <i class="bi bi-pencil"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-secondary me-1" @click="openModuleVersions(module)">
                      <i class="bi bi-clock-history"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger" @click="handleDeleteModule(module)">
                      <i class="bi bi-trash"></i>
                    </button>
                  </td>
                </tr>
                <tr v-if="!modules.length">
                  <td colspan="7" class="text-center text-muted py-4">{{ t('configs_page.no_modules') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="row g-4">
      <div class="col-xl-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-header bg-white">
            <h6 class="mb-0">{{ t('configs_page.render_params') }}</h6>
          </div>
          <div class="card-body">
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.preview_name') }}</label>
              <input v-model="previewForm.name" type="text" class="form-control" placeholder="preview-fluentbit-edge">
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('common.runtime') }}</label>
                <select v-model="previewForm.fluent_type" class="form-select">
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('configs_page.target_version') }}</label>
                <input v-model="previewForm.runtime_version" type="text" class="form-control" placeholder="3.1.0 / 1.16">
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.variables_json') }}</label>
              <textarea
                v-model="previewForm.variables"
                class="form-control font-monospace"
                rows="8"
                placeholder='{"path":"/var/log/*.log","match":"*"}'
              ></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.compatibility_node') }}</label>
              <input v-model="previewForm.node_id" type="number" min="1" class="form-control" :placeholder="t('configs_page.compatibility_node')">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.sample_tag') }}</label>
              <input v-model="previewForm.sample_tag" type="text" class="form-control" placeholder="app.logs">
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.sample_log') }}</label>
              <textarea
                v-model="previewForm.sample_log"
                class="form-control font-monospace"
                rows="6"
                placeholder='{"message":"hello fluent","level":"info"}'
              ></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('configs_page.diff_content') }}</label>
              <textarea
                v-model="previewForm.diff_content"
                class="form-control font-monospace"
                rows="6"
                :placeholder="t('configs_page.diff_content')"
              ></textarea>
            </div>
            <button class="btn btn-success w-100" @click="runPreview">
              <i class="bi bi-play-circle me-1"></i>{{ t('configs_page.generate_preview') }}
            </button>
            <button class="btn btn-outline-primary w-100 mt-2" @click="runLint">
              <i class="bi bi-shield-check me-1"></i>{{ t('configs_page.run_lint') }}
            </button>
            <button class="btn btn-outline-secondary w-100 mt-2" @click="runCompatibility">
              <i class="bi bi-patch-check me-1"></i>{{ t('configs_page.run_compatibility') }}
            </button>
            <button class="btn btn-outline-dark w-100 mt-2" @click="runReplay">
              <i class="bi bi-magic me-1"></i>{{ t('configs_page.run_replay') }}
            </button>
            <button class="btn btn-outline-info w-100 mt-2" @click="runSemanticDiff">
              <i class="bi bi-intersect me-1"></i>{{ t('configs_page.run_diff') }}
            </button>
          </div>
        </div>
      </div>

      <div class="col-xl-8">
        <div class="card border-0 shadow-sm mb-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <h6 class="mb-0">{{ t('configs_page.available_modules') }}</h6>
            <span class="text-muted small">{{ t('configs_page.runtime_help').replace('{runtime}', runtimeLabel(previewForm.fluent_type)) }}</span>
          </div>
          <div class="card-body">
            <div class="row g-3">
              <div
                v-for="module in previewEligibleModules"
                :key="module.id"
                class="col-lg-6"
              >
                <label class="fm-module-choice h-100" :class="{ selected: selectedPreviewModuleIds.includes(module.id) }">
                  <input
                    :checked="selectedPreviewModuleIds.includes(module.id)"
                    type="checkbox"
                    class="form-check-input"
                    @change="togglePreviewModule(module.id)"
                  >
                  <div class="flex-grow-1">
                    <div class="d-flex align-items-center gap-2 mb-2">
                      <span class="fw-semibold">{{ module.name }}</span>
                      <span class="badge text-bg-secondary">{{ module.module_type }}</span>
                      <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(module.fluent_type) }}</span>
                    </div>
                    <div class="small text-muted mb-2">{{ module.description || t('common.no_description') }}</div>
                    <pre class="fm-module-snippet">{{ module.content }}</pre>
                  </div>
                </label>
              </div>
              <div v-if="!previewEligibleModules.length" class="col-12 text-center text-muted py-4">
                {{ t('configs_page.current_runtime_no_modules') }}
              </div>
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.render_result') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.render_result') }}</div>
            </div>
            <div v-if="renderedConfig" class="text-end">
              <div class="small text-muted">{{ t('configs_page.render_hash') }}</div>
              <code>{{ renderedConfig.hash }}</code>
            </div>
          </div>
          <div class="card-body">
            <div v-if="renderedConfig">
              <div class="d-flex flex-wrap gap-2 mb-3">
                <span class="badge bg-success-subtle text-success-emphasis">ID {{ renderedConfig.id }}</span>
                <span class="badge bg-info-subtle text-info-emphasis">{{ runtimeLabel(renderedConfig.fluent_type) }}</span>
                <span class="badge text-bg-light">{{ renderedConfig.runtime_version || t('common.unspecified') }}</span>
              </div>
              <pre class="fm-render-preview">{{ renderedConfig.content }}</pre>
            </div>
            <div v-else class="text-center text-muted py-5">
              {{ t('configs_page.no_preview') }}
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm mt-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.analysis') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.analysis') }}</div>
            </div>
            <span v-if="analysisResult" class="badge text-bg-light">{{ analysisResult.summary }}</span>
          </div>
          <div class="card-body">
            <div v-if="analysisResult">
              <div class="list-group list-group-flush">
                <div
                  v-for="finding in analysisResult.findings || []"
                  :key="`${finding.rule_code}-${finding.line}`"
                  class="list-group-item px-0"
                >
                  <div class="d-flex align-items-center gap-2 mb-1">
                    <span class="badge" :class="findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                    <code>{{ finding.rule_code }}</code>
                    <span class="small text-muted">Line {{ finding.line }}</span>
                  </div>
                  <div class="fw-semibold">{{ finding.message }}</div>
                  <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
                </div>
              </div>
            </div>
            <div v-else class="text-center text-muted py-4">
              {{ t('configs_page.no_analysis') }}
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm mt-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.compatibility') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.compatibility') }}</div>
            </div>
            <span
              v-if="compatibilityResult"
              class="badge"
              :class="compatibilityResult.compatible ? 'text-bg-success' : 'text-bg-danger'"
            >
              {{ compatibilityResult.compatible ? 'compatible' : 'needs attention' }}
            </span>
          </div>
          <div class="card-body">
            <div v-if="compatibilityResult">
              <div class="d-flex flex-wrap gap-2 mb-3">
                <span class="badge text-bg-light">{{ compatibilityResult.hot_reload_supported ? t('configs_page.hot_reload_available') : t('configs_page.hot_reload_unavailable') }}</span>
                <span class="badge text-bg-light">{{ t('configs_page.missing_plugins') }} {{ compatibilityResult.missing_plugins?.length || 0 }}</span>
              </div>
              <div v-if="compatibilityResult.missing_plugins?.length" class="alert alert-warning py-2">
                {{ t('configs_page.missing_plugins') }}: {{ compatibilityResult.missing_plugins.join(', ') }}
              </div>
              <div class="list-group list-group-flush">
                <div
                  v-for="finding in compatibilityResult.findings || []"
                  :key="`compat-${finding.rule_code}-${finding.line}`"
                  class="list-group-item px-0"
                >
                  <div class="d-flex align-items-center gap-2 mb-1">
                    <span class="badge" :class="findingBadgeClass(finding.severity)">{{ finding.severity }}</span>
                    <code>{{ finding.rule_code }}</code>
                  </div>
                  <div class="fw-semibold">{{ finding.message }}</div>
                  <div class="small text-muted mt-1">{{ finding.suggestion || t('common.no_description') }}</div>
                </div>
              </div>
            </div>
            <div v-else class="text-center text-muted py-4">
              {{ t('configs_page.no_compatibility') }}
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm mt-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.replay') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.replay') }}</div>
            </div>
            <span v-if="replayResult" class="badge" :class="replayResult.route_matched ? 'text-bg-success' : 'text-bg-warning'">
              {{ replayResult.route_matched ? t('configs_page.route_matched') : t('configs_page.no_route') }}
            </span>
          </div>
          <div class="card-body">
            <div v-if="replayResult">
              <div class="d-flex flex-wrap gap-2 mb-3">
                <span class="badge text-bg-light">Parser {{ replayResult.detected_parser || '-' }}</span>
                <span class="badge text-bg-light">{{ t('common.output') }} {{ replayResult.final_output || '-' }}</span>
              </div>
              <div v-if="replayResult.warnings?.length" class="alert alert-warning py-2">
                {{ replayResult.warnings.join('；') }}
              </div>
              <div class="row g-4">
                <div class="col-lg-6">
                  <h6 class="small text-muted text-uppercase">{{ t('configs_page.steps') }}</h6>
                  <div class="list-group list-group-flush">
                    <div v-for="step in replayResult.steps || []" :key="`${step.stage}-${step.name}`" class="list-group-item px-0">
                      <div class="d-flex align-items-center gap-2 mb-1">
                        <span class="badge text-bg-light">{{ step.stage }}</span>
                        <span class="fw-semibold">{{ step.name }}</span>
                      </div>
                      <div class="small text-muted">{{ step.detail }}</div>
                    </div>
                  </div>
                </div>
                <div class="col-lg-6">
                  <h6 class="small text-muted text-uppercase">{{ t('configs_page.parsed_record') }}</h6>
                  <pre class="fm-render-preview mb-0">{{ formatJson(replayResult.parsed_record) }}</pre>
                </div>
              </div>
            </div>
            <div v-else class="text-center text-muted py-4">
              {{ t('configs_page.no_replay') }}
            </div>
          </div>
        </div>

        <div class="card border-0 shadow-sm mt-4">
          <div class="card-header bg-white d-flex justify-content-between align-items-center">
            <div>
              <h6 class="mb-0">{{ t('configs_page.semantic_diff') }}</h6>
              <div class="small text-muted mt-1">{{ t('configs_page.semantic_diff') }}</div>
            </div>
            <span v-if="diffResult" class="badge text-bg-light">{{ diffResult.summary }}</span>
          </div>
          <div class="card-body">
            <div v-if="diffResult">
              <div class="list-group list-group-flush">
                <div v-for="change in diffResult.changes || []" :key="`${change.category}-${change.change_type}-${change.item}`" class="list-group-item px-0">
                  <div class="d-flex align-items-center gap-2 mb-1">
                    <span class="badge text-bg-light">{{ change.category }}</span>
                    <span class="badge" :class="change.change_type === 'added' ? 'text-bg-success' : 'text-bg-danger'">{{ change.change_type }}</span>
                    <code>{{ change.item }}</code>
                  </div>
                  <div class="small text-muted">{{ change.detail }}</div>
                </div>
              </div>
              <div v-if="!diffResult.changes?.length" class="text-center text-muted py-3">
                {{ t('configs_page.no_changes') }}
              </div>
            </div>
            <div v-else class="text-center text-muted py-4">
              {{ t('configs_page.no_diff') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="templateModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('configs_page.create_template_title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info py-2">
              {{ t('configs_page.template_modal_hint') }}
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import {
  analyzeLogSampleAssistant,
  checkCompatibility,
  createModule,
  createModuleVersion,
  createTemplate,
  deleteModule,
  deleteTemplate,
  diffConfig,
  getModuleVersions,
  getModules,
  getRenderedConfig,
  getTemplates,
  lintConfig,
  previewRenderedConfig,
  replayConfig,
  updateModule,
} from '../api'

const activeTab = ref('templates')
const templates = ref([])
const modules = ref([])
const renderedConfig = ref(null)
const analysisResult = ref(null)
const compatibilityResult = ref(null)
const replayResult = ref(null)
const diffResult = ref(null)
const currentModule = ref(null)
const moduleVersions = ref([])
const selectedPreviewModuleIds = ref([])
const selectedWizardModuleIds = ref([])
const aiAssistantLoading = ref(false)
const aiAssistantResult = ref(null)
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
const wizardVariableValues = ref({})
const moduleVariablesMode = ref('form')
const moduleVariableRows = ref([])
const moduleVariablesFormError = ref('')

const moduleTypes = ['service', 'input', 'parser', 'filter', 'route', 'output']
const wizardRecommendedTypes = ['service', 'input', 'filter', 'output']
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

const wizardForm = reactive({
  goal: 'edge_collection',
  name: '',
  description: '',
  fluent_type: 'fluentbit',
  runtime_version: '',
})
const aiAssistantForm = reactive({
  fluent_type: 'fluentbit',
  goal: 'both',
  module_type: 'input',
  sample: '',
  extra_requirements: '',
})

const sharedModuleCount = computed(() => modules.value.filter((item) => item.fluent_type === 'shared').length)
const usedModuleTypes = computed(() => [...new Set(modules.value.map((item) => item.module_type))])
const previewEligibleModules = computed(() =>
  modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === previewForm.fluent_type)
)
const wizardEligibleModules = computed(() =>
  modules.value.filter((item) => item.fluent_type === 'shared' || item.fluent_type === wizardForm.fluent_type)
)
const wizardSelectedModules = computed(() =>
  wizardEligibleModules.value.filter((item) => selectedWizardModuleIds.value.includes(item.id))
)
const wizardModulesByType = computed(() =>
  moduleTypes
    .map((type) => ({
      type,
      modules: wizardEligibleModules.value.filter((item) => item.module_type === type),
    }))
    .filter((group) => group.modules.length)
)
const wizardSelectedTypes = computed(() => [...new Set(wizardSelectedModules.value.map((item) => item.module_type))])
const wizardMissingTypes = computed(() => wizardRecommendedTypes.filter((type) => !wizardSelectedTypes.value.includes(type)))
const wizardCoverageCount = computed(() => wizardRecommendedTypes.filter((type) => wizardSelectedTypes.value.includes(type)).length)
const wizardVariableFields = computed(() => {
  const merged = new Map()
  for (const module of wizardSelectedModules.value) {
    const variables = parseVariablesMap(module.variables)
    for (const [key, value] of Object.entries(variables)) {
      const existing = merged.get(key)
      const next = {
        key,
        defaultValue: stringifyVariableValue(value),
        kind: inferVariableKind(value),
        description: module.description || '',
        moduleNames: [module.name],
      }
      if (!existing) {
        merged.set(key, next)
      } else {
        existing.moduleNames = Array.from(new Set([...existing.moduleNames, module.name]))
        if (!existing.description && module.description) {
          existing.description = module.description
        }
      }
    }
  }
  return Array.from(merged.values())
})
const currentModuleExample = computed(() => {
  const runtimeExamples = moduleExamples[moduleForm.fluent_type] || moduleExamples.shared
  return runtimeExamples[moduleForm.module_type] || runtimeExamples.input || {
    variables: '{}',
    content: '# Example content',
  }
})
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

function formatJson(value) {
  try {
    return JSON.stringify(value || {}, null, 2)
  } catch {
    return String(value || '{}')
  }
}

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

function resetWizardForm() {
  wizardForm.goal = 'edge_collection'
  wizardForm.name = ''
  wizardForm.description = ''
  wizardForm.fluent_type = 'fluentbit'
  wizardForm.runtime_version = ''
  selectedWizardModuleIds.value = []
  wizardVariableValues.value = {}
}

async function loadTemplates() {
  const { data } = await getTemplates()
  templates.value = data.data || []
}

async function loadModules() {
  const { data } = await getModules()
  modules.value = data.data || []
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
    await loadModules()
  } catch (error) {
    alert(`${editingModuleId.value ? t('configs_page.save_module_failed') : t('configs_page.create_module_failed')}: ${getErrorMessage(error)}`)
  }
}

async function handleDeleteModule(module) {
  if (!confirm(t('configs_page.confirm_delete_module').replace('{name}', module.name))) return

  try {
    await deleteModule(module.id)
    selectedPreviewModuleIds.value = selectedPreviewModuleIds.value.filter((id) => id !== module.id)
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== module.id)
    await loadModules()
  } catch (error) {
    alert(`${t('configs_page.delete_module_failed')}: ${getErrorMessage(error)}`)
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
    await loadModules()
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
  if (selectedWizardModuleIds.value.includes(moduleId)) {
    selectedWizardModuleIds.value = selectedWizardModuleIds.value.filter((id) => id !== moduleId)
    return
  }
  selectedWizardModuleIds.value = [...selectedWizardModuleIds.value, moduleId]
}

function syncPreviewFromWizard() {
  previewForm.name = wizardForm.name || `preview-${wizardForm.goal}`
  previewForm.fluent_type = wizardForm.fluent_type
  previewForm.runtime_version = wizardForm.runtime_version

  const normalizedVariables = {}
  for (const field of wizardVariableFields.value) {
    normalizedVariables[field.key] = normalizeWizardVariableValue(wizardVariableValues.value[field.key], field.kind)
  }
  previewForm.variables = JSON.stringify(normalizedVariables, null, 2)
  selectedPreviewModuleIds.value = [...selectedWizardModuleIds.value]
}

async function runWizardPreview() {
  if (!selectedWizardModuleIds.value.length) {
    alert(t('configs_page.choose_modules'))
    return
  }
  syncPreviewFromWizard()
  await runPreview({ switchTab: false })
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
  syncPreviewFromWizard()
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
  if (!selectedPreviewModuleIds.value.length) {
    alert(t('configs_page.choose_modules'))
    return
  }

  try {
    const payload = {
      name: previewForm.name,
      fluent_type: previewForm.fluent_type,
      runtime_version: previewForm.runtime_version,
      variables: previewForm.variables,
      modules: selectedPreviewModuleIds.value.map((moduleId) => ({ module_id: moduleId })),
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
  wizardVariableFields,
  (fields) => {
    const next = {}
    for (const field of fields) {
      next[field.key] = Object.prototype.hasOwnProperty.call(wizardVariableValues.value, field.key)
        ? wizardVariableValues.value[field.key]
        : field.defaultValue
    }
    wizardVariableValues.value = next
  },
  { immediate: true }
)

onMounted(async () => {
  resetWizardForm()
  await Promise.all([loadTemplates(), loadModules()])
})
</script>

<style scoped>
.fm-config-tabs .nav-link {
  border: 0;
  color: #475569;
  font-weight: 600;
  border-radius: 10px;
}

.fm-config-explainer {
  border-radius: 12px;
  background: linear-gradient(135deg, #ecfeff 0%, #f0fdfa 100%);
  border: 1px solid rgba(13, 148, 136, 0.18);
  color: #0f172a;
}

.fm-config-tabs .nav-link.active {
  background: linear-gradient(135deg, #0f766e 0%, #0d9488 100%);
  color: #fff;
}

.fm-module-choice {
  display: flex;
  gap: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  cursor: pointer;
  transition: all 0.18s ease;
}

.fm-module-choice:hover {
  border-color: #94a3b8;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.06);
}

.fm-module-choice.selected {
  border-color: #0d9488;
  background: linear-gradient(180deg, #f0fdfa 0%, #ecfeff 100%);
  box-shadow: 0 12px 30px rgba(13, 148, 136, 0.12);
}

.fm-module-snippet {
  margin: 0;
  font-size: 0.75rem;
  max-height: 110px;
  overflow: auto;
  background: #0f172a;
  color: #e2e8f0;
  padding: 10px 12px;
  border-radius: 10px;
}

.fm-render-preview {
  margin: 0;
  max-height: 560px;
  overflow: auto;
  background: #0b1120;
  color: #dbeafe;
  padding: 18px;
  border-radius: 14px;
  border: 1px solid rgba(59, 130, 246, 0.15);
}

.fm-config-textarea::placeholder {
  color: #94a3b8;
  opacity: 1;
}

.fm-ai-result-box {
  height: 100%;
  padding: 1rem 1.05rem;
  border-radius: 14px;
  border: 1px solid #d7e2ee;
  background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
}

.fm-ai-result-box__label {
  margin-bottom: 0.45rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
}

.fm-ai-assistant-feedback {
  padding: 0.95rem 1rem;
  border-radius: 14px;
  border: 1px solid #d7e2ee;
  background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
}

.fm-ai-assistant-feedback.is-success {
  border-color: #bbf7d0;
  background: linear-gradient(180deg, #f0fdf4 0%, #ffffff 100%);
}

.fm-ai-assistant-feedback.is-danger {
  border-color: #fecaca;
  background: linear-gradient(180deg, #fef2f2 0%, #ffffff 100%);
}

.fm-ai-draft-panel {
  padding: 1rem 1.05rem;
  border-radius: 14px;
  border: 1px solid #bfdbfe;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
}

.fm-ai-draft-panel__title {
  margin-bottom: 0.35rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
}

.fm-ai-draft-diff-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.fm-ai-draft-diff-card {
  padding: 0.8rem 0.9rem;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(255, 255, 255, 0.72);
}

.fm-ai-draft-diff-card__label {
  margin-bottom: 0.35rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #64748b;
}

.fm-ai-draft-checklist {
  display: grid;
  gap: 0.55rem;
}

.fm-ai-draft-checklist__item {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  padding: 0.65rem 0.75rem;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(255, 255, 255, 0.72);
}

.fm-ai-draft-checklist__item .form-check-input {
  float: none;
  margin-top: 0.15rem;
}

.fm-ai-draft-actions {
  padding: 0.85rem 0.9rem;
  border-radius: 12px;
  border: 1px dashed rgba(59, 130, 246, 0.35);
  background: rgba(255, 255, 255, 0.72);
}

.fm-ai-draft-highlight {
  border-color: #93c5fd;
  background-color: #f8fbff;
  box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.08);
}

@media (max-width: 991.98px) {
  .fm-ai-draft-diff-grid {
    grid-template-columns: 1fr;
  }
}
</style>
