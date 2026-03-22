import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth
export const login = (data) => api.post('/auth/login', data)
export const getProfile = () => api.get('/auth/profile')
export const changePassword = (data) => api.put('/auth/password', data)

// Users
export const getUsers = (params) => api.get('/users', { params })
export const getUser = (id) => api.get(`/users/${id}`)
export const createUser = (data) => api.post('/users', data)
export const updateUser = (id, data) => api.put(`/users/${id}`, data)
export const deleteUser = (id) => api.delete(`/users/${id}`)

// Roles
export const getRoles = () => api.get('/roles')
export const getRole = (id) => api.get(`/roles/${id}`)
export const createRole = (data) => api.post('/roles', data)
export const updateRole = (id, data) => api.put(`/roles/${id}`, data)
export const deleteRole = (id) => api.delete(`/roles/${id}`)
export const getPermissions = () => api.get('/permissions')

// Topology
export const getTopologyTree = () => api.get('/topology/tree')
export const getEnvironments = () => api.get('/environments')
export const createEnvironment = (data) => api.post('/environments', data)
export const updateEnvironment = (id, data) => api.put(`/environments/${id}`, data)
export const deleteEnvironment = (id) => api.delete(`/environments/${id}`)

// DataCenters
export const getDataCenters = () => api.get('/datacenters')
export const getDataCenter = (id) => api.get(`/datacenters/${id}`)
export const createDataCenter = (data) => api.post('/datacenters', data)
export const updateDataCenter = (id, data) => api.put(`/datacenters/${id}`, data)
export const deleteDataCenter = (id) => api.delete(`/datacenters/${id}`)

// Regions
export const getRegions = (params) => api.get('/regions', { params })
export const getRegion = (id) => api.get(`/regions/${id}`)
export const createRegion = (data) => api.post('/regions', data)
export const updateRegion = (id, data) => api.put(`/regions/${id}`, data)
export const deleteRegion = (id) => api.delete(`/regions/${id}`)

// Clusters
export const getClusters = (params) => api.get('/clusters', { params })
export const getCluster = (id) => api.get(`/clusters/${id}`)
export const createCluster = (data) => api.post('/clusters', data)
export const updateCluster = (id, data) => api.put(`/clusters/${id}`, data)
export const deleteCluster = (id) => api.delete(`/clusters/${id}`)

// Nodes
export const getNodes = (params) => api.get('/nodes', { params })
export const getNode = (id) => api.get(`/nodes/${id}`)
export const updateNode = (id, data) => api.put(`/nodes/${id}`, data)
export const deleteNode = (id) => api.delete(`/nodes/${id}`)
export const getNodeStats = () => api.get('/nodes/stats')
export const batchMoveCluster = (data) => api.post('/nodes/batch-move', data)
export const getNodeMetrics = (id) => api.get(`/nodes/${id}/metrics`)
export const getNodeLogs = (id) => api.get(`/nodes/${id}/logs`)
export const sendNodeCommand = (id, data) => api.post(`/nodes/${id}/commands`, data)
export const getNodeCommands = (id) => api.get(`/nodes/${id}/commands`)

// Configs
export const getTemplates = (params) => api.get('/configs/templates', { params })
export const getTemplate = (id) => api.get(`/configs/templates/${id}`)
export const createTemplate = (data) => api.post('/configs/templates', data)
export const updateTemplate = (id, data) => api.put(`/configs/templates/${id}`, data)
export const deleteTemplate = (id) => api.delete(`/configs/templates/${id}`)
export const getVersions = (templateId) => api.get(`/configs/templates/${templateId}/versions`)
export const createVersion = (templateId, data) => api.post(`/configs/templates/${templateId}/versions`, data)
export const getVersion = (versionId) => api.get(`/configs/versions/${versionId}`)

// Deployments
export const getDeploys = (params) => api.get('/deploys', { params })
export const getDeploy = (id) => api.get(`/deploys/${id}`)
export const createDeploy = (data) => api.post('/deploys', data)

// Audit Logs
export const getAuditLogs = (params) => api.get('/audit-logs', { params })

export default api
