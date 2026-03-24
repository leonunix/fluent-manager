<template>
  <div class="card border-0 shadow-sm">
    <div class="card-body p-0">
      <div class="border-bottom bg-light-subtle p-3">
        <div class="d-flex flex-wrap justify-content-between align-items-start gap-3">
          <div>
            <div class="fw-semibold">{{ t('configs_page.templates_recommended_title') }}</div>
            <div class="small text-muted mt-1">{{ t('configs_page.templates_recommended_body') }}</div>
          </div>
          <div class="d-flex flex-wrap gap-2">
            <span class="badge bg-success-subtle text-success-emphasis">
              {{ t('configs_page.template_source_module_assembly') }} {{ state.assemblyTemplateCount }}
            </span>
            <span class="badge text-bg-light">
              {{ t('configs_page.template_source_manual') }} {{ state.manualTemplateCount }}
            </span>
          </div>
        </div>
      </div>
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>{{ t('common.name') }}</th>
              <th>{{ t('common.runtime') }}</th>
              <th>{{ t('configs_page.template_source') }}</th>
              <th>{{ t('common.description') }}</th>
              <th>{{ t('deploys_page.creator') }}</th>
              <th>{{ t('deploys_page.created_at') }}</th>
              <th>{{ t('actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="tpl in state.templates" :key="tpl.id">
              <td>
                <router-link :to="`/configs/${tpl.id}`" class="text-decoration-none fw-semibold">
                  {{ tpl.name }}
                </router-link>
              </td>
              <td><span class="badge bg-info-subtle text-info-emphasis">{{ helpers.runtimeLabel(tpl.fluent_type) }}</span></td>
              <td>
                <span
                  class="badge"
                  :class="tpl.source_type === 'module_assembly' ? 'bg-success-subtle text-success-emphasis' : 'text-bg-light'"
                >
                  {{ helpers.templateSourceLabel(tpl.source_type) }}
                </span>
                <div v-if="tpl.source_type === 'module_assembly'" class="small text-muted mt-1">
                  {{ t('configs_page.assembly_module_count').replace('{count}', String(helpers.templateAssemblyModules(tpl).length)) }}
                </div>
              </td>
              <td>{{ tpl.description || '-' }}</td>
              <td>{{ tpl.creator?.username || '-' }}</td>
              <td>{{ helpers.formatTime(tpl.created_at) }}</td>
              <td>
                <router-link :to="`/configs/${tpl.id}`" class="btn btn-sm btn-outline-primary me-1">
                  <i class="bi bi-eye"></i>
                </router-link>
                <button class="btn btn-sm btn-outline-danger" @click="actions.handleDeleteTemplate(tpl)">
                  <i class="bi bi-trash"></i>
                </button>
              </td>
            </tr>
            <tr v-if="!state.templates.length">
              <td colspan="7" class="text-center text-muted py-4">{{ t('configs_page.no_templates') }}</td>
            </tr>
          </tbody>
        </table>
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
