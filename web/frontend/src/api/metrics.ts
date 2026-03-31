import api from './client'

export const getMetricsOverview = () => api.get('/metrics/overview')
export const getMetricsTopNodes = () => api.get('/metrics/top-nodes')
export const getMetricsByDC = () => api.get('/metrics/by-datacenter')
export const getMetricsThroughput = () => api.get('/metrics/throughput')
export const getNodeThroughput24h = (id: number | string) => api.get(`/nodes/${id}/throughput-24h`)
