import api from './client'

export const getRoles = () => api.get('/roles')
export const getRole = (id: number) => api.get(`/roles/${id}`)
export const createRole = (data: Record<string, any>) => api.post('/roles', data)
export const updateRole = (id: number, data: Record<string, any>) => api.put(`/roles/${id}`, data)
export const deleteRole = (id: number) => api.delete(`/roles/${id}`)
export const getPermissions = () => api.get('/permissions')
