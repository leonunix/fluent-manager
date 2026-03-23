import api from './client'

export const getMetricsOverview = () => api.get('/metrics/overview')
export const getMetricsTopNodes = () => api.get('/metrics/top-nodes')
export const getMetricsByDC = () => api.get('/metrics/by-datacenter')
