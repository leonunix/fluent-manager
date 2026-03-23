import api from './client'

export const getAISettings = () => api.get('/ai-settings')
export const updateAISettings = (data: Record<string, any>) => api.put('/ai-settings', data)
export const testAIAccount = (data: Record<string, any>) => api.post('/ai-settings/test-account', data)
export const analyzeLogSample = (data: Record<string, any>) => api.post('/ai-settings/log-sample-analysis', data)
