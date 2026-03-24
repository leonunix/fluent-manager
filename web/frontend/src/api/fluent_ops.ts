import api from './client'
import type {
  AggregationGroupRuntimeMetric,
  CompatibilityCheckInput,
  CompatibilityCheckResult,
  ConfigAnalysisResult,
  ConfigLintInput,
  ConfigReplayInput,
  ConfigReplayResult,
  ConfigSemanticDiffInput,
  ConfigSemanticDiffResult,
  LogPipeline,
  LogPipelineInput,
  OutputTarget,
  OutputTargetInput,
  PipelineGraph,
  RuntimeDriftItem,
  RuntimeRecommendation,
} from '../types/fluent'

interface ListEnvelope<T> {
  data: T[]
}

export const getPipelines = () =>
  api.get<ListEnvelope<LogPipeline>>('/log-pipelines').then(({ data }) => data.data || [])
export const getPipeline = (id: number) =>
  api.get<LogPipeline>(`/log-pipelines/${id}`).then(({ data }) => data)
export const createPipeline = (data: LogPipelineInput) =>
  api.post<LogPipeline>('/log-pipelines', data).then(({ data: created }) => created)
export const updatePipeline = (id: number, data: LogPipelineInput) =>
  api.put<LogPipeline>(`/log-pipelines/${id}`, data).then(({ data: updated }) => updated)
export const deletePipeline = (id: number) => api.delete(`/log-pipelines/${id}`)
export const getPipelineGraph = () =>
  api.get<PipelineGraph>('/log-pipelines/graph').then(({ data }) => data)

export const getOutputTargets = () =>
  api.get<ListEnvelope<OutputTarget>>('/output-targets').then(({ data }) => data.data || [])
export const getOutputTarget = (id: number) =>
  api.get<OutputTarget>(`/output-targets/${id}`).then(({ data }) => data)
export const createOutputTarget = (data: OutputTargetInput) =>
  api.post<OutputTarget>('/output-targets', data).then(({ data: created }) => created)
export const updateOutputTarget = (id: number, data: OutputTargetInput) =>
  api.put<OutputTarget>(`/output-targets/${id}`, data).then(({ data: updated }) => updated)
export const deleteOutputTarget = (id: number) => api.delete(`/output-targets/${id}`)

export const lintConfig = (data: ConfigLintInput) =>
  api.post<ConfigAnalysisResult>('/config-analysis/lint', data).then(({ data: result }) => result)
export const replayConfig = (data: ConfigReplayInput) =>
  api.post<ConfigReplayResult>('/config-analysis/replay', data).then(({ data: result }) => result)
export const diffConfig = (data: ConfigSemanticDiffInput) =>
  api.post<ConfigSemanticDiffResult>('/config-analysis/diff', data).then(({ data: result }) => result)
export const checkCompatibility = (data: CompatibilityCheckInput) =>
  api.post<CompatibilityCheckResult>('/config-analysis/compatibility', data).then(({ data: result }) => result)
export const getAnalysisResult = (id: number) =>
  api.get<ConfigAnalysisResult>(`/config-analysis/${id}`).then(({ data }) => data)

export const getRuntimeDrift = () =>
  api.get<ListEnvelope<RuntimeDriftItem>>('/runtime/drift').then(({ data }) => data.data || [])
export const getRuntimeHealthGraph = () =>
  api.get<PipelineGraph>('/runtime/health/graph').then(({ data }) => data)
export const getRuntimeRecommendations = () =>
  api.get<ListEnvelope<RuntimeRecommendation>>('/runtime/recommendations').then(({ data }) => data.data || [])
export const getAggregationGroupMetrics = (id: number) =>
  api.get<AggregationGroupRuntimeMetric>(`/aggregation-groups/${id}/metrics`).then(({ data }) => data)
