import axios from 'axios'

const api = axios.create({
  baseURL: '',
  withCredentials: true,
})

// Response interceptor for 401 handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const isDev = import.meta.env.DEV
      const mainSiteURL = isDev ? 'https://tropical.wavespeed.ai' : 'https://wavespeed.ai'
      const redirectURL = `${mainSiteURL}/portal/`
      const loginURL = `${mainSiteURL}/sign-in?redirect=${encodeURIComponent(redirectURL)}`
      window.location.href = loginURL
    }
    return Promise.reject(error)
  }
)

// User
export const getCurrentUser = () => api.get('/api/v1/user').then(r => r.data)

// Specs
export const getSpecs = (type?: string) => 
  api.get('/api/v1/specs', { params: { type } }).then(r => r.data)

export const estimateCost = (specName: string, hours: number, replicas: number) =>
  api.post('/api/v1/estimate-cost', { spec_name: specName, hours, replicas }).then(r => r.data)

// Endpoints
export const getEndpoints = () => 
  api.get('/api/v1/endpoints').then(r => r.data)

export const getEndpoint = (name: string) => 
  api.get(`/api/v1/endpoints/${name}`).then(r => r.data)

export const createEndpoint = (data: {
  logical_name: string
  spec_name: string
  image: string
  replicas?: number
  min_replicas: number
  max_replicas: number
  task_timeout?: number
  prefer_region?: string
  env?: Record<string, string>
  registry_credential_name?: string
}) => api.post('/api/v1/endpoints', data).then(r => r.data)

export const updateEndpoint = (name: string, data: { replicas?: number; image?: string; env?: Record<string, string> }) =>
  api.put(`/api/v1/endpoints/${name}`, data).then(r => r.data)

export const updateEndpointConfig = (name: string, config: Record<string, unknown>) =>
  api.put(`/api/v1/endpoints/${name}/config`, config).then(r => r.data)

export const deleteEndpoint = (name: string) =>
  api.delete(`/api/v1/endpoints/${name}`).then(r => r.data)

// Endpoint Monitoring
export const getEndpointWorkers = (name: string) =>
  api.get(`/api/v1/endpoints/${name}/workers`).then(r => r.data)

export const getWorkerLogs = (endpoint: string, podName: string) =>
  api.get(`/api/v1/endpoints/${endpoint}/logs`, { params: { pod_name: podName } }).then(r => r.data)

export const getWorkerTasks = (endpoint: string, workerId: string) =>
  api.get(`/api/v1/endpoints/${endpoint}/tasks`, { params: { worker_id: workerId } }).then(r => r.data)

export const getEndpointMetrics = (name: string) =>
  api.get(`/api/v1/endpoints/${name}/metrics`).then(r => r.data)

export const getEndpointStats = (name: string, granularity: string, from?: string, to?: string) =>
  api.get(`/api/v1/endpoints/${name}/stats`, { params: { granularity, from, to } }).then(r => r.data)

export const getEndpointStatistics = (name: string) =>
  api.get(`/api/v1/endpoints/${name}/statistics`).then(r => r.data)

// Billing
export const getUsage = (from?: string, to?: string) =>
  api.get('/api/v1/billing/usage', { params: { from, to } }).then(r => r.data)

export const getWorkerRecords = (limit?: number, offset?: number) =>
  api.get('/api/v1/billing/workers', { params: { limit, offset } }).then(r => r.data)

// Tasks
export const submitTask = (endpoint: string, input: unknown, sync = false) =>
  api.post(`/v1/${endpoint}/${sync ? 'runsync' : 'run'}`, { input }).then(r => r.data)

export const getTaskStatus = (taskId: string) =>
  api.get(`/v1/status/${taskId}`).then(r => r.data)

export const cancelTask = (taskId: string) =>
  api.post(`/v1/cancel/${taskId}`).then(r => r.data)

export const getTaskTimeline = (taskId: string) =>
  api.get(`/api/v1/tasks/${taskId}/timeline`).then(r => r.data)

export const getTaskExecutionHistory = (taskId: string) =>
  api.get(`/api/v1/tasks/${taskId}/execution-history`).then(r => r.data)

export const getEndpointTasks = (endpoint: string, params?: { limit?: number; offset?: number; status?: string; task_id?: string; worker_id?: string }) =>
  api.get(`/api/v1/endpoints/${endpoint}/tasks`, { params }).then(r => r.data)

export const getAllTasks = (params?: { limit?: number; offset?: number; status?: string; task_id?: string }) =>
  api.get('/api/v1/tasks', { params }).then(r => r.data)

// Scaling History
export const getScalingHistory = (endpoint: string, limit?: number) =>
  api.get(`/api/v1/endpoints/${endpoint}/scaling-history`, { params: { limit } }).then(r => r.data)

// Registry Credentials
export const getRegistryCredentials = () =>
  api.get('/api/v1/registry-credentials').then(r => r.data)

export const createRegistryCredential = (data: { name: string; registry?: string; username: string; password: string }) =>
  api.post('/api/v1/registry-credentials', data).then(r => r.data)

export const deleteRegistryCredential = (name: string) =>
  api.delete(`/api/v1/registry-credentials/${name}`).then(r => r.data)

// Clusters (Admin)
export const getClusters = () => api.get('/api/v1/admin/clusters').then(r => r.data)
export const getCluster = (id: string) => api.get(`/api/v1/admin/clusters/${id}`).then(r => r.data)
export const createCluster = (data: object) => api.post('/api/v1/admin/clusters', data).then(r => r.data)
export const updateCluster = (id: string, data: object) => api.put(`/api/v1/admin/clusters/${id}`, data).then(r => r.data)
export const deleteCluster = (id: string) => api.delete(`/api/v1/admin/clusters/${id}`).then(r => r.data)

// Cluster Specs (Admin)
export const getClusterSpecs = (clusterId: string) => api.get(`/api/v1/admin/clusters/${clusterId}/specs`).then(r => r.data)
export const createClusterSpec = (clusterId: string, data: object) => api.post(`/api/v1/admin/clusters/${clusterId}/specs`, data).then(r => r.data)
export const updateClusterSpec = (clusterId: string, data: object) => api.put(`/api/v1/admin/clusters/${clusterId}/specs`, data).then(r => r.data)
export const deleteClusterSpec = (clusterId: string, id: number) => api.delete(`/api/v1/admin/clusters/${clusterId}/specs`, { data: { id } }).then(r => r.data)

// Specs (Admin)
export const getAdminSpecs = () => api.get('/api/v1/admin/specs').then(r => r.data)
export const createSpec = (data: object) => api.post('/api/v1/admin/specs', data).then(r => r.data)
export const updateSpec = (data: object) => api.put('/api/v1/admin/specs', data).then(r => r.data)
export const deleteSpec = (id: number) => api.delete('/api/v1/admin/specs', { data: { id } }).then(r => r.data)

export default api
