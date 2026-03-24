<template>
  <div>
    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('configs_page.total_modules') }}</div>
            <div class="fs-3 fw-bold">{{ state.moduleCatalogCount }}</div>
            <div class="small text-muted mt-2">{{ t('configs_page.source_runtime_hint') }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('configs_page.shared_modules') }}</div>
            <div class="fs-3 fw-bold">{{ state.sharedModuleCount }}</div>
            <div class="small text-muted mt-2">{{ t('configs_page.source_runtime_hint') }}</div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="text-muted small mb-1">{{ t('configs_page.module_type_coverage') }}</div>
            <div class="fs-3 fw-bold">{{ state.usedModuleTypes.length }}/{{ state.managedModuleTypes.length }}</div>
            <div class="small text-muted mt-2">{{ state.usedModuleTypes.join(' / ') || '-' }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <div class="border-bottom p-3">
          <div class="small text-muted mb-3">{{ t('configs_page.module_search_placeholder') }}</div>
          <div class="row g-3 align-items-end">
            <div class="col-xl-4">
              <label class="form-label">{{ t('common.search') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-search"></i></span>
                <input
                  v-model="state.moduleQuery.search"
                  type="text"
                  class="form-control"
                  :placeholder="t('configs_page.module_search_placeholder')"
                  @keyup.enter="actions.applyModuleQuery"
                >
              </div>
            </div>
            <div class="col-md-3 col-xl-2">
              <label class="form-label">{{ t('configs_page.module_runtime_filter') }}</label>
              <select v-model="state.moduleQuery.fluent_type" class="form-select" @change="actions.applyModuleQuery">
                <option value="">{{ t('all') }}</option>
                <option value="fluentbit">Fluent Bit</option>
                <option value="fluentd">Fluentd</option>
                <option value="shared">Shared</option>
              </select>
            </div>
            <div class="col-md-3 col-xl-2">
              <label class="form-label">{{ t('configs_page.module_type_filter') }}</label>
              <select v-model="state.moduleQuery.module_type" class="form-select" @change="actions.applyModuleQuery">
                <option value="">{{ t('all') }}</option>
                <option v-for="type in state.managedModuleTypes" :key="type" :value="type">{{ type }}</option>
              </select>
            </div>
            <div class="col-md-3 col-xl-2">
              <label class="form-label">{{ t('configs_page.module_page_size') }}</label>
              <select v-model.number="state.moduleQuery.page_size" class="form-select" @change="actions.applyModuleQuery">
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
            </div>
            <div class="col-md-3 col-xl-2">
              <div class="d-grid gap-2">
                <button type="button" class="btn btn-primary" @click="actions.applyModuleQuery">
                  <i class="bi bi-search me-1"></i>{{ t('search') }}
                </button>
                <button type="button" class="btn btn-outline-secondary" @click="actions.resetModuleQuery">
                  {{ t('configs_page.module_clear_filters') }}
                </button>
              </div>
            </div>
          </div>
          <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-3">
            <div class="d-flex flex-wrap align-items-center gap-2">
              <span class="small text-muted">{{ t('configs_page.module_type_filter') }}</span>
              <button
                type="button"
                class="btn btn-sm"
                :class="state.moduleQuery.module_type ? 'btn-outline-secondary' : 'btn-primary'"
                @click="actions.setModuleTypeFilter('')"
              >
                {{ t('all') }}
                <span class="badge rounded-pill text-bg-light ms-1">{{ state.moduleCatalogCount }}</span>
              </button>
              <button
                v-for="item in state.moduleTypeStats"
                :key="item.type"
                type="button"
                class="btn btn-sm"
                :class="state.moduleQuery.module_type === item.type ? 'btn-primary' : 'btn-outline-secondary'"
                @click="actions.setModuleTypeFilter(item.type)"
              >
                {{ item.type }}
                <span class="badge rounded-pill text-bg-light ms-1">{{ item.count }}</span>
              </button>
            </div>
            <div class="small text-muted">
              {{ t('configs_page.module_results')
                .replace('{start}', String(state.moduleTableRangeStart))
                .replace('{end}', String(state.moduleTableRangeEnd))
                .replace('{total}', String(state.moduleTableTotal)) }}
            </div>
            <div class="small text-muted" v-if="state.selectedDeletableModuleCount">
              {{ t('configs_page.module_selected_count').replace('{count}', String(state.selectedDeletableModuleCount)) }}
            </div>
            <div class="btn-group btn-group-sm">
              <button class="btn btn-outline-secondary" :disabled="state.moduleQuery.page <= 1 || state.moduleTableLoading" @click="actions.changeModulePage(state.moduleQuery.page - 1)">
                {{ t('common.previous') }}
              </button>
              <button class="btn btn-outline-secondary" disabled>
                {{ state.moduleQuery.page }} / {{ state.moduleTableTotalPages }}
              </button>
              <button
                class="btn btn-outline-secondary"
                :disabled="state.moduleQuery.page >= state.moduleTableTotalPages || state.moduleTableLoading || !state.moduleTableTotal"
                @click="actions.changeModulePage(state.moduleQuery.page + 1)"
              >
                {{ t('common.next') }}
              </button>
            </div>
          </div>
        </div>
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th class="text-center" style="width: 56px;">
                  <input
                    type="checkbox"
                    class="form-check-input"
                    :checked="state.allVisibleDeletableModulesSelected"
                    :disabled="!state.visibleDeletableModules.length"
                    @change="actions.toggleSelectAllVisibleModules"
                  >
                </th>
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
              <tr v-for="module in state.visibleModules" :key="module.id">
                <td class="text-center">
                  <input
                    type="checkbox"
                    class="form-check-input"
                    :checked="state.selectedModuleIds.includes(module.id)"
                    :disabled="module.is_builtin"
                    @change="actions.toggleModuleSelection(module)"
                  >
                </td>
                <td>
                  <div class="fw-semibold">{{ module.name }}</div>
                  <div class="small text-muted">{{ module.description || t('common.no_description') }}</div>
                </td>
                <td><span class="badge text-bg-secondary">{{ module.module_type }}</span></td>
                <td><span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(module.fluent_type) }}</span></td>
                <td>
                  <span :class="module.is_builtin ? 'badge text-bg-dark' : 'badge text-bg-light'">
                    {{ module.is_builtin ? t('configs_page.builtin') : t('configs_page.custom') }}
                  </span>
                  <div v-if="module.is_builtin" class="small text-muted mt-1">{{ t('configs_page.builtin_module_protected') }}</div>
                </td>
                <td>
                  <code class="small">{{ helpers.shortVariables(module.variables) }}</code>
                </td>
                <td>{{ helpers.formatTime(module.created_at) }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="actions.openEditModule(module)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-secondary me-1" @click="actions.openModuleVersions(module)">
                    <i class="bi bi-clock-history"></i>
                  </button>
                  <button
                    class="btn btn-sm btn-outline-danger"
                    :disabled="module.is_builtin"
                    :title="module.is_builtin ? t('configs_page.builtin_module_protected') : ''"
                    @click="actions.handleDeleteModule(module)"
                  >
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="state.moduleTableLoading">
                <td colspan="8" class="text-center text-muted py-4">{{ t('loading') }}</td>
              </tr>
              <tr v-else-if="!state.visibleModules.length">
                <td colspan="8" class="text-center text-muted py-4">
                  {{ state.moduleQuery.search || state.moduleQuery.fluent_type || state.moduleQuery.module_type ? t('configs_page.module_no_search_results') : t('configs_page.no_modules') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from '../../../i18n'

defineProps({
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
</script>
