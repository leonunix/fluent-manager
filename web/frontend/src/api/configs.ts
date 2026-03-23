import api from './client'

export const getTemplates = (params?: Record<string, any>) => api.get('/configs/templates', { params })
export const getTemplate = (id: number) => api.get(`/configs/templates/${id}`)
export const createTemplate = (data: Record<string, any>) => api.post('/configs/templates', data)
export const updateTemplate = (id: number, data: Record<string, any>) => api.put(`/configs/templates/${id}`, data)
export const deleteTemplate = (id: number) => api.delete(`/configs/templates/${id}`)
export const getVersions = (templateId: number) => api.get(`/configs/templates/${templateId}/versions`)
export const createVersion = (templateId: number, data: Record<string, any>) => api.post(`/configs/templates/${templateId}/versions`, data)
export const getVersion = (versionId: number) => api.get(`/configs/versions/${versionId}`)

export const getModules = (params?: Record<string, any>) => api.get('/configs/modules', { params })
export const getModule = (id: number) => api.get(`/configs/modules/${id}`)
export const createModule = (data: Record<string, any>) => api.post('/configs/modules', data)
export const updateModule = (id: number, data: Record<string, any>) => api.put(`/configs/modules/${id}`, data)
export const deleteModule = (id: number) => api.delete(`/configs/modules/${id}`)
export const getModuleVersions = (moduleId: number) => api.get(`/configs/modules/${moduleId}/versions`)
export const createModuleVersion = (moduleId: number, data: Record<string, any>) => api.post(`/configs/modules/${moduleId}/versions`, data)

export const previewRenderedConfig = (data: Record<string, any>) => api.post('/configs/rendered-configs/preview', data)
export const getRenderedConfig = (id: number) => api.get(`/configs/rendered-configs/${id}`)
