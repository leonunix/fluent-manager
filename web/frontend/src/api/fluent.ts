import api from './client'
import type { AggregationGroup, NodeFluentProfile } from '../types/node'
import type { AggregationGroupInput, NodeFluentProfileInput } from '../types/fluent'

interface ListEnvelope<T> {
  data: T[]
}

export const getAggregationGroups = () =>
  api.get<ListEnvelope<AggregationGroup>>('/aggregation-groups').then(({ data }) => data.data || [])
export const getDeletedAggregationGroups = () =>
  api.get<ListEnvelope<AggregationGroup>>('/aggregation-groups/deleted').then(({ data }) => data.data || [])
export const getAggregationGroup = (id: number) =>
  api.get<AggregationGroup>(`/aggregation-groups/${id}`).then(({ data }) => data)
export const createAggregationGroup = (data: AggregationGroupInput) =>
  api.post<AggregationGroup>('/aggregation-groups', data).then(({ data: created }) => created)
export const updateAggregationGroup = (id: number, data: AggregationGroupInput) =>
  api.put<AggregationGroup>(`/aggregation-groups/${id}`, data).then(({ data: updated }) => updated)
export const deleteAggregationGroup = (id: number) => api.delete(`/aggregation-groups/${id}`)
export const restoreAggregationGroup = (id: number) =>
  api.post<AggregationGroup>(`/aggregation-groups/${id}/restore`).then(({ data }) => data)

export const getNodeFluentProfile = (id: number) =>
  api.get<NodeFluentProfile>(`/nodes/${id}/fluent-profile`).then(({ data }) => data)
export const updateNodeFluentProfile = (id: number, data: NodeFluentProfileInput) =>
  api.put<NodeFluentProfile>(`/nodes/${id}/fluent-profile`, data).then(({ data: updated }) => updated)
