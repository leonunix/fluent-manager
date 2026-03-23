import api from './client'

export const getLDAPSettings = () => api.get('/auth-settings/ldap')
export const updateLDAPSettings = (data: Record<string, any>) => api.put('/auth-settings/ldap', data)
export const testLDAPConnection = (data: Record<string, any>) => api.post('/auth-settings/ldap/test', data)
export const getSAMLSettings = () => api.get('/auth-settings/saml')
export const updateSAMLSettings = (data: Record<string, any>) => api.put('/auth-settings/saml', data)
export const getGroupMappings = (source: string) => api.get('/auth-settings/group-mappings', { params: { source } })
export const setGroupMappings = (data: Record<string, any>) => api.put('/auth-settings/group-mappings', data)
