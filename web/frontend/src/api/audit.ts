import api from './client'

export const getAuditLogs = (params?: Record<string, any>) => api.get('/audit-logs', { params })
