import api from './client'

export const getTopologyTree = () => api.get('/topology/tree')
export const getEnvironments = () => api.get('/environments')
export const createEnvironment = (data: Record<string, any>) => api.post('/environments', data)
export const updateEnvironment = (id: number, data: Record<string, any>) => api.put(`/environments/${id}`, data)
export const deleteEnvironment = (id: number) => api.delete(`/environments/${id}`)

export const getDataCenters = () => api.get('/datacenters')
export const getDataCenter = (id: number) => api.get(`/datacenters/${id}`)
export const createDataCenter = (data: Record<string, any>) => api.post('/datacenters', data)
export const updateDataCenter = (id: number, data: Record<string, any>) => api.put(`/datacenters/${id}`, data)
export const deleteDataCenter = (id: number) => api.delete(`/datacenters/${id}`)

export const getRegions = (params?: Record<string, any>) => api.get('/regions', { params })
export const getRegion = (id: number) => api.get(`/regions/${id}`)
export const createRegion = (data: Record<string, any>) => api.post('/regions', data)
export const updateRegion = (id: number, data: Record<string, any>) => api.put(`/regions/${id}`, data)
export const deleteRegion = (id: number) => api.delete(`/regions/${id}`)

export const getClusters = (params?: Record<string, any>) => api.get('/clusters', { params })
export const getCluster = (id: number) => api.get(`/clusters/${id}`)
export const createCluster = (data: Record<string, any>) => api.post('/clusters', data)
export const updateCluster = (id: number, data: Record<string, any>) => api.put(`/clusters/${id}`, data)
export const deleteCluster = (id: number) => api.delete(`/clusters/${id}`)

export const getClusterRules = (clusterID: number) => api.get(`/clusters/${clusterID}/rules`)
export const createClusterRule = (clusterID: number, data: Record<string, any>) => api.post(`/clusters/${clusterID}/rules`, data)
export const updateClusterRule = (clusterID: number, ruleID: number, data: Record<string, any>) => api.put(`/clusters/${clusterID}/rules/${ruleID}`, data)
export const deleteClusterRule = (clusterID: number, ruleID: number) => api.delete(`/clusters/${clusterID}/rules/${ruleID}`)

export const getUserScopes = (userID: number) => api.get(`/users/${userID}/scopes`)
export const setUserScopes = (userID: number, data: Record<string, any>) => api.put(`/users/${userID}/scopes`, data)
