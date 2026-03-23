import api from './client'
import type { AgentPolicyListResponse, ResolvedAgentPolicy, AgentPolicy } from '../types'

export const getAgentPolicies = () => api.get<AgentPolicyListResponse>('/agent-policies')
export const getAgentPolicy = (id: number) => api.get<AgentPolicy>(`/agent-policies/${id}`)
export const createAgentPolicy = (data: Record<string, any>) => api.post<AgentPolicy>('/agent-policies', data)
export const updateAgentPolicy = (id: number, data: Record<string, any>) => api.put<AgentPolicy>(`/agent-policies/${id}`, data)
export const deleteAgentPolicy = (id: number) => api.delete(`/agent-policies/${id}`)
export const resolveAgentPolicyForNode = (nodeId: number) => api.get<ResolvedAgentPolicy>(`/agent-policies/resolve/${nodeId}`)

