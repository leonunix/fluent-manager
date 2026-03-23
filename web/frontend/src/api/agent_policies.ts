import api from './client'
import type { AgentPolicyListResponse, ResolvedAgentPolicy, AgentPolicy, AgentPolicyInput } from '../types'

export const getAgentPolicies = () =>
  api.get<AgentPolicyListResponse>('/agent-policies').then(({ data }) => data)
export const getAgentPolicy = (id: number) =>
  api.get<AgentPolicy>(`/agent-policies/${id}`).then(({ data }) => data)
export const createAgentPolicy = (data: AgentPolicyInput) =>
  api.post<AgentPolicy>('/agent-policies', data).then(({ data: created }) => created)
export const updateAgentPolicy = (id: number, data: AgentPolicyInput) =>
  api.put<AgentPolicy>(`/agent-policies/${id}`, data).then(({ data: updated }) => updated)
export const deleteAgentPolicy = (id: number) => api.delete(`/agent-policies/${id}`)
export const resolveAgentPolicyForNode = (nodeId: number) =>
  api.get<ResolvedAgentPolicy>(`/agent-policies/resolve/${nodeId}`).then(({ data }) => data)
