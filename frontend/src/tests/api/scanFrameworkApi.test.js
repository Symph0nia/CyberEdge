import { describe, it, expect, vi, beforeEach } from 'vitest'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'
import axiosInstance from '@/api/axiosInstance'

// Mock axios instance
vi.mock('@/api/axiosInstance', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn()
  }
}))

describe('scanFrameworkApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('startScan', () => {
    it('should call POST /api/scan-framework/start with correct parameters', async () => {
      const mockResponse = { data: { scan_id: 'scan123' } }
      axiosInstance.post.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const target = 'example.com'
      const pipelineName = 'comprehensive'

      const result = await scanFrameworkApi.startScan(projectId, target, pipelineName)

      expect(axiosInstance.post).toHaveBeenCalledWith('/api/scan-framework/start', {
        project_id: projectId,
        target: target,
        pipeline_name: pipelineName
      })
      expect(result).toBe(mockResponse)
    })

    it('should handle API errors correctly', async () => {
      const mockError = new Error('Network error')
      axiosInstance.post.mockRejectedValue(mockError)

      await expect(
        scanFrameworkApi.startScan('project123', 'example.com', 'comprehensive')
      ).rejects.toThrow('Network error')
    })

    it('should validate required parameters', async () => {
      const mockResponse = { data: { scan_id: 'scan123' } }
      axiosInstance.post.mockResolvedValue(mockResponse)

      await scanFrameworkApi.startScan('', '', '')

      expect(axiosInstance.post).toHaveBeenCalledWith('/api/scan-framework/start', {
        project_id: '',
        target: '',
        pipeline_name: ''
      })
    })
  })

  describe('getScanStatus', () => {
    it('should call GET /api/scan-framework/scans/:scanId/status', async () => {
      const mockResponse = {
        data: {
          id: 'scan123',
          state: 'running',
          target_address: 'example.com',
          created_at: '2024-01-01T00:00:00Z'
        }
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const scanId = 'scan123'
      const result = await scanFrameworkApi.getScanStatus(scanId)

      expect(axiosInstance.get).toHaveBeenCalledWith(`/api/scan-framework/scans/${scanId}/status`)
      expect(result).toBe(mockResponse)
    })

    it('should handle 404 for non-existent scan', async () => {
      const mockError = { response: { status: 404, data: { error: 'Scan not found' } } }
      axiosInstance.get.mockRejectedValue(mockError)

      await expect(
        scanFrameworkApi.getScanStatus('nonexistent')
      ).rejects.toMatchObject({ response: { status: 404 } })
    })
  })

  describe('getProjectScans', () => {
    it('should call GET /api/scan-framework/projects/:projectId/scans with params', async () => {
      const mockResponse = {
        data: [
          { id: 'scan1', state: 'running' },
          { id: 'scan2', state: 'completed' }
        ]
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const params = { status: 'running', page: 1, limit: 10 }
      const result = await scanFrameworkApi.getProjectScans(projectId, params)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/scans`,
        { params }
      )
      expect(result).toBe(mockResponse)
    })

    it('should work without params', async () => {
      const mockResponse = { data: [] }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      await scanFrameworkApi.getProjectScans(projectId)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/scans`,
        { params: {} }
      )
    })
  })

  describe('getScanResults', () => {
    it('should call GET /api/scan-framework/projects/:projectId/results', async () => {
      const mockResponse = {
        data: [
          { id: 'result1', service_name: 'HTTP', port: 80 },
          { id: 'result2', service_name: 'HTTPS', port: 443 }
        ]
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const params = { port: 80 }
      const result = await scanFrameworkApi.getScanResults(projectId, params)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/results`,
        { params }
      )
      expect(result).toBe(mockResponse)
    })
  })

  describe('getAvailableTools', () => {
    it('should call GET /api/scan-framework/tools', async () => {
      const mockResponse = {
        data: {
          subdomain: [{ name: 'subfinder', available: true }],
          port: [{ name: 'nmap', available: true }]
        }
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const result = await scanFrameworkApi.getAvailableTools()

      expect(axiosInstance.get).toHaveBeenCalledWith('/api/scan-framework/tools')
      expect(result).toBe(mockResponse)
    })
  })

  describe('getAvailablePipelines', () => {
    it('should call GET /api/scan-framework/pipelines', async () => {
      const mockResponse = {
        data: ['comprehensive', 'quick', 'web', 'network']
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const result = await scanFrameworkApi.getAvailablePipelines()

      expect(axiosInstance.get).toHaveBeenCalledWith('/api/scan-framework/pipelines')
      expect(result).toBe(mockResponse)
    })
  })

  describe('getProjectVulnerabilities', () => {
    it('should call GET /api/scan-framework/projects/:projectId/vulnerabilities', async () => {
      const mockResponse = {
        data: [
          {
            id: 'vuln1',
            title: 'SQL Injection',
            severity: 'high',
            description: 'SQL injection vulnerability found',
            location: '/login.php'
          }
        ]
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const params = { severity: 'high' }
      const result = await scanFrameworkApi.getProjectVulnerabilities(projectId, params)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/vulnerabilities`,
        { params }
      )
      expect(result).toBe(mockResponse)
    })
  })

  describe('getVulnerabilityStats', () => {
    it('should call GET /api/scan-framework/projects/:projectId/vulnerabilities/stats', async () => {
      const mockResponse = {
        data: {
          critical: 2,
          high: 5,
          medium: 10,
          low: 15,
          info: 8
        }
      }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const result = await scanFrameworkApi.getVulnerabilityStats(projectId)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/vulnerabilities/stats`
      )
      expect(result).toBe(mockResponse)
    })
  })

  describe('stopScan', () => {
    it('should call POST /api/scan-framework/scans/:scanId/stop', async () => {
      const mockResponse = { data: { message: 'Scan stopped successfully' } }
      axiosInstance.post.mockResolvedValue(mockResponse)

      const scanId = 'scan123'
      const result = await scanFrameworkApi.stopScan(scanId)

      expect(axiosInstance.post).toHaveBeenCalledWith(`/api/scan-framework/scans/${scanId}/stop`)
      expect(result).toBe(mockResponse)
    })

    it('should handle stop failure for already completed scan', async () => {
      const mockError = {
        response: {
          status: 400,
          data: { error: 'Cannot stop completed scan' }
        }
      }
      axiosInstance.post.mockRejectedValue(mockError)

      await expect(
        scanFrameworkApi.stopScan('completed-scan')
      ).rejects.toMatchObject({ response: { status: 400 } })
    })
  })

  describe('deleteScan', () => {
    it('should call DELETE /api/scan-framework/scans/:scanId', async () => {
      const mockResponse = { data: { message: 'Scan deleted successfully' } }
      axiosInstance.delete.mockResolvedValue(mockResponse)

      const scanId = 'scan123'
      const result = await scanFrameworkApi.deleteScan(scanId)

      expect(axiosInstance.delete).toHaveBeenCalledWith(`/api/scan-framework/scans/${scanId}`)
      expect(result).toBe(mockResponse)
    })

    it('should handle delete failure for running scan', async () => {
      const mockError = {
        response: {
          status: 400,
          data: { error: 'Cannot delete running scan' }
        }
      }
      axiosInstance.delete.mockRejectedValue(mockError)

      await expect(
        scanFrameworkApi.deleteScan('running-scan')
      ).rejects.toMatchObject({ response: { status: 400 } })
    })
  })

  describe('exportScanResults', () => {
    it('should call GET /api/scan-framework/projects/:projectId/export with format param', async () => {
      const mockBlob = new Blob(['test data'], { type: 'application/json' })
      const mockResponse = { data: mockBlob }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const format = 'json'
      const result = await scanFrameworkApi.exportScanResults(projectId, format)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/export`,
        {
          params: { format },
          responseType: 'blob'
        }
      )
      expect(result).toBe(mockResponse)
    })

    it('should use default format when not specified', async () => {
      const mockBlob = new Blob(['test data'], { type: 'application/json' })
      const mockResponse = { data: mockBlob }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      await scanFrameworkApi.exportScanResults(projectId)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/export`,
        {
          params: { format: 'json' },
          responseType: 'blob'
        }
      )
    })

    it('should support different export formats', async () => {
      const mockBlob = new Blob(['test data'], { type: 'text/csv' })
      const mockResponse = { data: mockBlob }
      axiosInstance.get.mockResolvedValue(mockResponse)

      const projectId = 'project123'
      const format = 'csv'
      await scanFrameworkApi.exportScanResults(projectId, format)

      expect(axiosInstance.get).toHaveBeenCalledWith(
        `/api/scan-framework/projects/${projectId}/export`,
        {
          params: { format: 'csv' },
          responseType: 'blob'
        }
      )
    })
  })

  describe('error handling', () => {
    it('should handle network errors consistently', async () => {
      const networkError = new Error('Network Error')
      networkError.code = 'NETWORK_ERROR'
      axiosInstance.get.mockRejectedValue(networkError)

      await expect(
        scanFrameworkApi.getAvailableTools()
      ).rejects.toThrow('Network Error')
    })

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout')
      timeoutError.code = 'ECONNABORTED'
      axiosInstance.post.mockRejectedValue(timeoutError)

      await expect(
        scanFrameworkApi.startScan('project123', 'example.com', 'comprehensive')
      ).rejects.toThrow('Request timeout')
    })

    it('should handle server errors (5xx)', async () => {
      const serverError = {
        response: {
          status: 500,
          data: { error: 'Internal server error' }
        }
      }
      axiosInstance.get.mockRejectedValue(serverError)

      await expect(
        scanFrameworkApi.getScanStatus('scan123')
      ).rejects.toMatchObject({ response: { status: 500 } })
    })
  })
})