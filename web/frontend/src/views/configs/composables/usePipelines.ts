// Pipeline management composable – state, computed, and CRUD for config pipelines.
// Deps injected: { modules, outputTargets, pipelines, t }
// pipelines ref is owned by ConfigsPage and passed in.

import { computed, reactive, ref, watch } from 'vue'
import { getConfigPipelines, createConfigPipeline, updateConfigPipeline, deleteConfigPipeline } from '../../../api/configs'

export function usePipelines({ modules, outputTargets, pipelines, t }: any) {
  const editingPipelineId = ref<number | null>(null)
  const pipelineFormInitializing = ref(false)
  const pipelineFilterPickerValue = ref('')

  const pipelineForm = reactive({
    name: '',
    description: '',
    fluent_type: 'fluentbit',
    input_module_id: null as number | null,
    filter_module_ids: [] as number[],
    output_target_ids: [] as number[],
  })

  // --- Computed ---

  const pipelineEligibleModules = computed(() =>
    modules.value.filter((item: any) => item.fluent_type === 'shared' || item.fluent_type === pipelineForm.fluent_type)
  )
  const pipelineInputModules = computed(() =>
    pipelineEligibleModules.value.filter((item: any) => item.module_type === 'input')
  )
  const pipelineFilterModules = computed(() =>
    pipelineEligibleModules.value.filter((item: any) => item.module_type === 'filter')
  )
  const pipelineAvailableOutputTargets = computed(() =>
    outputTargets.value.filter((item: any) => item.fluent_type === 'shared' || item.fluent_type === pipelineForm.fluent_type)
  )

  // --- Watchers ---

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

  // --- Functions ---

  async function loadPipelines() {
    const { data } = await getConfigPipelines()
    pipelines.value = data.data || []
  }

  function preparePipelineCreate() {
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
  }

  function preparePipelineEdit(pipeline: any) {
    pipelineFormInitializing.value = true
    editingPipelineId.value = pipeline.id
    pipelineForm.name = pipeline.name
    pipelineForm.description = pipeline.description || ''
    pipelineForm.fluent_type = pipeline.fluent_type || 'fluentbit'
    pipelineForm.input_module_id = pipeline.input_module_id || null
    pipelineForm.filter_module_ids = (pipeline.filter_modules || []).map((m: any) => m.id)
    pipelineForm.output_target_ids = (pipeline.output_targets || []).map((t: any) => t.id)
    pipelineFilterPickerValue.value = ''
    pipelineFormInitializing.value = false
  }

  async function savePipeline(): Promise<boolean> {
    if (!pipelineForm.name.trim()) { alert('Name is required'); return false }
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
      return true
    } catch (e: any) {
      alert(`${t('common.request_failed')}: ${e?.response?.data?.error || e?.message || ''}`)
      return false
    }
  }

  async function handleDeletePipeline(pipeline: any) {
    if (!confirm(t('configs_page.pipeline_delete_confirm').replace('{name}', pipeline.name))) return
    try {
      await deleteConfigPipeline(pipeline.id)
      await loadPipelines()
    } catch (e: any) {
      alert(`${t('common.request_failed')}: ${e?.response?.data?.error || e?.message || ''}`)
    }
  }

  function addPipelineFilterModule(moduleId: number) {
    if (moduleId && !pipelineForm.filter_module_ids.includes(moduleId)) {
      pipelineForm.filter_module_ids.push(moduleId)
    }
  }

  function removePipelineFilterModule(index: number) {
    pipelineForm.filter_module_ids.splice(index, 1)
  }

  function movePipelineFilterModule(index: number, direction: number) {
    const arr = pipelineForm.filter_module_ids
    const newIndex = index + direction
    if (newIndex < 0 || newIndex >= arr.length) return
    const tmp = arr[index]; arr[index] = arr[newIndex]; arr[newIndex] = tmp
  }

  function togglePipelineOutputTarget(targetId: number) {
    const idx = pipelineForm.output_target_ids.indexOf(targetId)
    if (idx === -1) pipelineForm.output_target_ids.push(targetId)
    else pipelineForm.output_target_ids.splice(idx, 1)
  }

  return {
    editingPipelineId, pipelineForm, pipelineFilterPickerValue,
    pipelineEligibleModules, pipelineInputModules, pipelineFilterModules, pipelineAvailableOutputTargets,
    loadPipelines, preparePipelineCreate, preparePipelineEdit,
    savePipeline, handleDeletePipeline,
    addPipelineFilterModule, removePipelineFilterModule, movePipelineFilterModule, togglePipelineOutputTarget,
  }
}
