import api from './client'

export const getNodes = (params?: Record<string, any>) => api.get('/nodes', { params })
export const getNode = (id: number) => api.get(`/nodes/${id}`)
export const updateNode = (id: number, data: Record<string, any>) => api.put(`/nodes/${id}`, data)
export const deleteNode = (id: number) => api.delete(`/nodes/${id}`)
export const getNodeStats = () => api.get('/nodes/stats')
export const batchMoveCluster = (data: Record<string, any>) => api.post('/nodes/batch-move', data)
export const getNodeMetrics = (id: number) => api.get(`/nodes/${id}/metrics`)
export const getNodeLogs = (id: number) => api.get(`/nodes/${id}/logs`)
export const sendNodeCommand = (id: number, data: Record<string, any>) => api.post(`/nodes/${id}/commands`, data)
export const getNodeCommands = (id: number) => api.get(`/nodes/${id}/commands`)
export const getNodeCommand = (id: number, cmdId: number) => api.get(`/nodes/${id}/commands/${cmdId}`)
