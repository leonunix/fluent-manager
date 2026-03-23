import api from './client'

export const getDeploys = (params?: Record<string, any>) => api.get('/deploys', { params })
export const getDeploy = (id: number) => api.get(`/deploys/${id}`)
export const createDeploy = (data: Record<string, any>) => api.post('/deploys', data)
