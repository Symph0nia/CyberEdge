import axiosInstance from './axiosInstance'

export const scanFrameworkApi = {
  // 扫描任务管理
  async startScan(projectId, target, pipelineName) {
    return await axiosInstance.post('/api/scan-framework/start', {
      project_id: projectId,
      target: target,
      pipeline_name: pipelineName
    })
  },

  async getScanStatus(scanId) {
    return await axiosInstance.get(`/api/scan-framework/scans/${scanId}/status`)
  },

  async getProjectScans(projectId, params = {}) {
    return await axiosInstance.get(`/api/scan-framework/projects/${projectId}/scans`, { params })
  },

  async getScanResults(projectId, params = {}) {
    return await axiosInstance.get(`/api/scan-framework/projects/${projectId}/results`, { params })
  },

  // 扫描工具和流水线配置
  async getAvailableTools() {
    return await axiosInstance.get('/api/scan-framework/tools')
  },

  async getAvailablePipelines() {
    return await axiosInstance.get('/api/scan-framework/pipelines')
  },

  // 漏洞管理
  async getProjectVulnerabilities(projectId, params = {}) {
    return await axiosInstance.get(`/api/scan-framework/projects/${projectId}/vulnerabilities`, { params })
  },

  async getVulnerabilityStats(projectId) {
    return await axiosInstance.get(`/api/scan-framework/projects/${projectId}/vulnerabilities/stats`)
  },

  // 扫描任务操作
  async stopScan(scanId) {
    return await axiosInstance.post(`/api/scan-framework/scans/${scanId}/stop`)
  },

  async deleteScan(scanId) {
    return await axiosInstance.delete(`/api/scan-framework/scans/${scanId}`)
  },

  // 导出扫描结果
  async exportScanResults(projectId, format = 'json') {
    return await axiosInstance.get(`/api/scan-framework/projects/${projectId}/export`, {
      params: { format },
      responseType: 'blob'
    })
  }
}