import api from './client'

export const getDeploys = (params?: Record<string, any>) => api.get('/deploys', { params })
export const getDeploy = (id: number, params?: Record<string, any>) => api.get(`/deploys/${id}`, { params })
export const createDeploy = (data: Record<string, any>) => api.post('/deploys', data)
