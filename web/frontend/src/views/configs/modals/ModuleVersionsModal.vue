<template>
  <div ref="el" class="modal fade" tabindex="-1">
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
                    <div v-for="version in versions" :key="version.id" class="list-group-item">
                      <div class="d-flex justify-content-between align-items-start mb-1">
                        <strong>v{{ version.version }}</strong>
                        <small class="text-muted">{{ formatTime(version.created_at) }}</small>
                      </div>
                      <div class="small text-muted mb-1">{{ version.comment || t('configs_page.no_version_comment') }}</div>
                      <code class="small">{{ version.hash }}</code>
                    </div>
                    <div v-if="!versions.length" class="list-group-item text-muted">{{ t('configs_page.no_history_versions') }}</div>
                  </div>
                </div>
              </div>
            </div>
            <div class="col-lg-8">
              <h6 class="mb-3">{{ t('configs_page.create_new_version') }}</h6>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.version_comment') }}</label>
                <input v-model="versionForm.comment" type="text" class="form-control">
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.variables_json') }}</label>
                <textarea v-model="versionForm.variables" class="form-control font-monospace" rows="4"></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('configs_page.version_content') }}</label>
                <textarea v-model="versionForm.content" class="form-control font-monospace" rows="14"></textarea>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('common.close') }}</button>
          <button type="button" class="btn btn-primary" @click="$emit('save-version')">{{ t('configs_page.create_new_version') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../../i18n'

defineProps({
  currentModule: { type: Object, default: null },
  versions: { type: Array, default: () => [] },
  versionForm: { type: Object, required: true },
  formatTime: { type: Function, required: true },
})
defineEmits(['save-version'])

const { t } = useI18n()
const el = ref(null)
let modal = null

onMounted(() => {
  modal = new window.bootstrap.Modal(el.value)
})

defineExpose({
  show: () => modal?.show(),
  hide: () => modal?.hide(),
})
</script>
