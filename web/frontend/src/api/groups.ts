import api from './client'

export const getGroups = (params?: Record<string, any>) => api.get('/groups', { params })
export const getGroup = (id: number) => api.get(`/groups/${id}`)
export const createGroup = (data: Record<string, any>) => api.post('/groups', data)
export const updateGroup = (id: number, data: Record<string, any>) => api.put(`/groups/${id}`, data)
export const deleteGroup = (id: number) => api.delete(`/groups/${id}`)
export const setGroupUsers = (id: number, data: Record<string, any>) => api.put(`/groups/${id}/users`, data)
