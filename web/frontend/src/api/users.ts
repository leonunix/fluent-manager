import api from './client'

export const getUsers = (params?: Record<string, any>) => api.get('/users', { params })
export const getUser = (id: number) => api.get(`/users/${id}`)
export const createUser = (data: Record<string, any>) => api.post('/users', data)
export const updateUser = (id: number, data: Record<string, any>) => api.put(`/users/${id}`, data)
export const deleteUser = (id: number) => api.delete(`/users/${id}`)
