import api from './client'

export const login = (data: { username: string; password: string; auth_source?: string }) => api.post('/auth/login', data)
export const getProfile = () => api.get('/auth/profile')
export const changePassword = (data: { old_password: string; new_password: string }) => api.put('/auth/password', data)
export const getAuthMethods = () => api.get('/auth/methods')
export const exchangeSAMLCode = (code: string) => api.post('/auth/saml/exchange', { code })
