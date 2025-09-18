import axiosInstance from './axiosInstance'

export const scanApi = {
  // 项目管理
  async getProjects() {
    return await axiosInstance.get('/api/scan/projects')
  },

  async createProject(data) {
    return await axiosInstance.post('/api/scan/projects', data)
  },

  async getProject(id) {
    return await axiosInstance.get(`/api/scan/projects/${id}`)
  },

  async updateProject(id, data) {
    return await axiosInstance.put(`/api/scan/projects/${id}`, data)
  },

  async deleteProject(id) {
    return await axiosInstance.delete(`/api/scan/projects/${id}`)
  },

  // 扫描数据管理
  async importScanResults(projectId, data) {
    return await axiosInstance.post(`/api/scan/projects/${projectId}/import`, data, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  async createSampleData(projectId) {
    return await axiosInstance.post(`/api/scan/projects/${projectId}/sample`)
  },

  // 统计信息
  async getProjectStats(projectId) {
    return await axiosInstance.get(`/api/scan/projects/${projectId}/stats`)
  },

  // 漏洞管理
  async getVulnerabilities(projectId, params = {}) {
    return await axiosInstance.get(`/api/scan/projects/${projectId}/vulnerabilities`, { params })
  },

  async getVulnerability(projectId, vulnId) {
    return await axiosInstance.get(`/api/scan/projects/${projectId}/vulnerabilities/${vulnId}`)
  },

  async updateVulnerabilityStatus(projectId, vulnId, status) {
    return await axiosInstance.patch(`/api/scan/projects/${projectId}/vulnerabilities/${vulnId}`, { status })
  },

  // 导出功能
  async exportProject(projectId, format = 'json') {
    return await axiosInstance.get(`/api/scan/projects/${projectId}/export`, {
      params: { format },
      responseType: 'blob'
    })
  }
}