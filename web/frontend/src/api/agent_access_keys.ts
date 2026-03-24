import api from './client'
import type { AgentAccessKey, AgentAccessKeyCreateResult, AgentAccessKeyInput } from '../types'

export const getAgentAccessKeys = () =>
  api.get<{ data: AgentAccessKey[] }>('/agent-access-keys').then(({ data }) => data)

export const createAgentAccessKey = (payload: AgentAccessKeyInput) =>
  api.post<AgentAccessKeyCreateResult>('/agent-access-keys', payload).then(({ data }) => data)

export const updateAgentAccessKey = (id: number, payload: AgentAccessKeyInput) =>
  api.put<AgentAccessKey>(`/agent-access-keys/${id}`, payload).then(({ data }) => data)

export const deleteAgentAccessKey = (id: number) =>
  api.delete(`/agent-access-keys/${id}`)
