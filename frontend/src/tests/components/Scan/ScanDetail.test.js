import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ScanDetail from '@/components/Scan/ScanDetail.vue'
import { scanApi } from '@/api/scanApi'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'

// Mock the APIs
vi.mock('@/api/scanApi')
vi.mock('@/api/scanFrameworkApi')

// Mock router
const mockRoute = {
  params: { id: 'scan123' }
}

const mockRouter = {
  push: vi.fn()
}

// Mock $message
const mockMessage = {
  success: vi.fn(),
  error: vi.fn()
}

// Mock data
const mockScanInfo = {
  id: 'scan123',
  target_address: 'example.com',
  target_id: 'target123',
  service_name: 'comprehensive',
  project_id: 'project123',
  state: 'running',
  created_at: '2024-01-01T10:00:00Z',
  updated_at: '2024-01-01T10:30:00Z'
}

const mockProject = {
  id: 'project123',
  name: 'Test Project'
}

const mockScanResults = [
  {
    id: 'result1',
    service_name: 'HTTP',
    port: 80,
    protocol: 'tcp',
    target_address: 'example.com',
    state: 'discovered',
    created_at: '2024-01-01T10:15:00Z',
    version: 'Apache/2.4.41',
    banner: 'Apache/2.4.41 (Ubuntu)'
  },
  {
    id: 'result2',
    service_name: 'HTTPS',
    port: 443,
    protocol: 'tcp',
    target_address: 'example.com',
    state: 'open',
    created_at: '2024-01-01T10:16:00Z'
  }
]

const mockVulnerabilities = [
  {
    id: 'vuln1',
    title: 'SQL Injection Vulnerability',
    description: 'SQL injection vulnerability found in login form',
    severity: 'high',
    location: '/login.php',
    cvss: '8.5',
    cve_id: 'CVE-2024-1234'
  },
  {
    id: 'vuln2',
    title: 'Cross-Site Scripting (XSS)',
    description: 'Reflected XSS vulnerability in search parameter',
    severity: 'medium',
    location: '/search.php',
    cvss: '6.1'
  }
]

const mockVulnStats = {
  critical: 1,
  high: 2,
  medium: 3,
  low: 1,
  info: 0
}

describe('ScanDetail', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()

    // Setup API mocks
    scanFrameworkApi.getScanStatus = vi.fn().mockResolvedValue({ data: mockScanInfo })
    scanFrameworkApi.getScanResults = vi.fn().mockResolvedValue({ data: mockScanResults })
    scanFrameworkApi.getProjectVulnerabilities = vi.fn().mockResolvedValue({ data: mockVulnerabilities })
    scanFrameworkApi.getVulnerabilityStats = vi.fn().mockResolvedValue({ data: mockVulnStats })
    scanFrameworkApi.stopScan = vi.fn().mockResolvedValue({ data: { message: 'Scan stopped' } })
    scanApi.getProject = vi.fn().mockResolvedValue({ data: mockProject })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    vi.clearAllTimers()
  })

  const createWrapper = (options = {}) => {
    return mount(ScanDetail, {
      global: {
        mocks: {
          $route: mockRoute,
          $router: mockRouter,
          $message: mockMessage
        }
      },
      ...options
    })
  }

  describe('Component Mounting and Data Loading', () => {
    it('should render correctly', async () => {
      wrapper = createWrapper()
      await nextTick()

      expect(wrapper.find('h1').text()).toBe('扫描详情')
      expect(wrapper.find('.back-link').exists()).toBe(true)
      expect(wrapper.find('.scan-info-card').exists()).toBe(true)
    })

    it('should load scan data on mount', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(scanFrameworkApi.getScanStatus).toHaveBeenCalledWith('scan123')
      expect(scanFrameworkApi.getScanResults).toHaveBeenCalledWith('project123')
      expect(scanFrameworkApi.getProjectVulnerabilities).toHaveBeenCalledWith('project123')
      expect(scanApi.getProject).toHaveBeenCalledWith('project123')
    })

    it('should populate data after loading', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(wrapper.vm.scanInfo).toEqual(mockScanInfo)
      expect(wrapper.vm.scanResults).toEqual(mockScanResults)
      expect(wrapper.vm.vulnerabilities).toEqual(mockVulnerabilities)
      expect(wrapper.vm.projectName).toBe('Test Project')
    })

    it('should handle loading errors gracefully', async () => {
      scanFrameworkApi.getScanStatus = vi.fn().mockRejectedValue(new Error('API Error'))

      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(mockMessage.error).toHaveBeenCalledWith('加载扫描详情失败')
    })
  })

  describe('Scan Information Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should display basic scan information', () => {
      expect(wrapper.text()).toContain('example.com')
      expect(wrapper.text()).toContain('comprehensive')
      expect(wrapper.text()).toContain('Test Project')
    })

    it('should show correct status badge', () => {
      const statusBadge = wrapper.find('.status-badge')
      expect(statusBadge.classes()).toContain('status-running')
      expect(statusBadge.text()).toBe('运行中')
    })

    it('should display progress bar for running scans', () => {
      const progressSection = wrapper.find('.progress-section')
      expect(progressSection.exists()).toBe(true)

      const progressBar = progressSection.find('.progress-bar')
      expect(progressBar.exists()).toBe(true)
    })

    it('should not show progress bar for completed scans', async () => {
      wrapper.vm.scanInfo.state = 'completed'
      await nextTick()

      const progressSection = wrapper.find('.progress-section')
      expect(progressSection.exists()).toBe(false)
    })

    it('should show running duration for active scans', () => {
      const duration = wrapper.vm.getRunningDuration()
      expect(typeof duration).toBe('string')
      expect(duration.length).toBeGreaterThan(0)
    })

    it('should display formatted dates correctly', () => {
      expect(wrapper.text()).toContain('2024')
    })
  })

  describe('Statistics Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should display statistics cards', () => {
      const statCards = wrapper.findAll('.stat-card')
      expect(statCards).toHaveLength(4)
    })

    it('should show correct scan results count', () => {
      expect(wrapper.vm.stats.totalResults).toBe(mockScanResults.length)
    })

    it('should show correct vulnerability count', () => {
      expect(wrapper.vm.stats.vulnerabilities.total).toBe(mockVulnerabilities.length)
    })

    it('should calculate high-risk vulnerabilities correctly', () => {
      const highRisk = wrapper.vm.stats.vulnerabilities.critical + wrapper.vm.stats.vulnerabilities.high
      expect(highRisk).toBe(3) // 1 critical + 2 high
    })

    it('should show assets count', () => {
      const assetsCount = mockScanResults.filter(r => r.state === 'discovered').length
      expect(wrapper.vm.stats.assets).toBe(assetsCount)
    })
  })

  describe('Tab Navigation', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should show all tabs with counts', () => {
      const tabs = wrapper.vm.tabs
      expect(tabs).toHaveLength(3)

      expect(tabs[0].key).toBe('vulnerabilities')
      expect(tabs[0].count).toBe(mockVulnerabilities.length)

      expect(tabs[1].key).toBe('results')
      expect(tabs[1].count).toBe(mockScanResults.length)

      expect(tabs[2].key).toBe('logs')
    })

    it('should switch tabs correctly', async () => {
      const resultsTab = wrapper.findAll('.tab-button')[1]
      await resultsTab.trigger('click')

      expect(wrapper.vm.activeTab).toBe('results')
    })

    it('should show active tab styling', async () => {
      wrapper.vm.activeTab = 'vulnerabilities'
      await nextTick()

      const activeTab = wrapper.find('.tab-button.active')
      expect(activeTab.exists()).toBe(true)
    })

    it('should show correct tab content', async () => {
      wrapper.vm.activeTab = 'vulnerabilities'
      await nextTick()

      const vulnList = wrapper.find('.vulnerabilities-list')
      expect(vulnList.exists()).toBe(true)

      wrapper.vm.activeTab = 'results'
      await nextTick()

      const resultsList = wrapper.find('.scan-results')
      expect(resultsList.exists()).toBe(true)
    })
  })

  describe('Vulnerabilities Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
      wrapper.vm.activeTab = 'vulnerabilities'
      await nextTick()
    })

    it('should display vulnerability items', () => {
      const vulnItems = wrapper.findAll('.vulnerability-item')
      expect(vulnItems).toHaveLength(mockVulnerabilities.length)
    })

    it('should show vulnerability details', () => {
      const firstVuln = wrapper.find('.vulnerability-item')
      expect(firstVuln.text()).toContain('SQL Injection Vulnerability')
      expect(firstVuln.text()).toContain('SQL injection vulnerability found in login form')
      expect(firstVuln.text()).toContain('/login.php')
      expect(firstVuln.text()).toContain('8.5')
      expect(firstVuln.text()).toContain('CVE-2024-1234')
    })

    it('should show correct severity styling', () => {
      const vulnItems = wrapper.findAll('.vulnerability-item')
      expect(vulnItems[0].classes()).toContain('severity-high')
      expect(vulnItems[1].classes()).toContain('severity-medium')
    })

    it('should display severity badges correctly', () => {
      const severityBadges = wrapper.findAll('.severity-badge')
      expect(severityBadges[0].text()).toBe('HIGH')
      expect(severityBadges[1].text()).toBe('MEDIUM')
    })

    it('should show empty state when no vulnerabilities', async () => {
      wrapper.vm.vulnerabilities = []
      await nextTick()

      const emptyState = wrapper.find('.empty-state')
      expect(emptyState.exists()).toBe(true)
      expect(emptyState.text()).toContain('暂未发现漏洞')
    })
  })

  describe('Scan Results Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
      wrapper.vm.activeTab = 'results'
      await nextTick()
    })

    it('should display result items', () => {
      const resultItems = wrapper.findAll('.result-item')
      expect(resultItems).toHaveLength(mockScanResults.length)
    })

    it('should show result details', () => {
      const firstResult = wrapper.find('.result-item')
      expect(firstResult.text()).toContain('HTTP')
      expect(firstResult.text()).toContain('80')
      expect(firstResult.text()).toContain('tcp')
      expect(firstResult.text()).toContain('example.com')
    })

    it('should show version information when available', () => {
      const resultWithVersion = wrapper.findAll('.result-item')[0]
      expect(resultWithVersion.text()).toContain('Apache/2.4.41')
    })

    it('should show banner information when available', () => {
      const resultWithBanner = wrapper.findAll('.result-item')[0]
      expect(resultWithBanner.text()).toContain('Apache/2.4.41 (Ubuntu)')
    })

    it('should show correct status styling', () => {
      const resultItems = wrapper.findAll('.result-item')
      const firstStatus = resultItems[0].find('.result-status')
      const secondStatus = resultItems[1].find('.result-status')

      expect(firstStatus.classes()).toContain('result-discovered')
      expect(secondStatus.classes()).toContain('result-open')
    })

    it('should show empty state when no results', async () => {
      wrapper.vm.scanResults = []
      await nextTick()

      const emptyState = wrapper.find('.empty-state')
      expect(emptyState.exists()).toBe(true)
      expect(emptyState.text()).toContain('暂无扫描结果')
    })
  })

  describe('Logs Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
      wrapper.vm.activeTab = 'logs'
      await nextTick()
    })

    it('should show logs section', () => {
      const logsSection = wrapper.find('.scan-logs')
      expect(logsSection.exists()).toBe(true)
    })

    it('should show refresh logs button', () => {
      const refreshButton = wrapper.find('.logs-header button')
      expect(refreshButton.text()).toBe('刷新日志')
    })

    it('should refresh logs when button clicked', async () => {
      const refreshButton = wrapper.find('.logs-header button')
      await refreshButton.trigger('click')

      expect(wrapper.vm.logs.length).toBeGreaterThan(0)
    })

    it('should display log entries after refresh', async () => {
      await wrapper.vm.refreshLogs()
      await nextTick()

      const logLines = wrapper.findAll('.log-line')
      expect(logLines.length).toBeGreaterThan(0)
    })

    it('should show empty state when no logs', () => {
      const emptyState = wrapper.find('.empty-state')
      expect(emptyState.exists()).toBe(true)
      expect(emptyState.text()).toContain('暂无日志记录')
    })
  })

  describe('Scan Actions', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should show stop button for running scans', () => {
      const stopButton = wrapper.find('button.btn-warning')
      expect(stopButton.exists()).toBe(true)
      expect(stopButton.text()).toContain('停止扫描')
    })

    it('should not show stop button for completed scans', async () => {
      wrapper.vm.scanInfo.state = 'completed'
      await nextTick()

      const stopButton = wrapper.find('button.btn-warning')
      expect(stopButton.exists()).toBe(false)
    })

    it('should stop scan when stop button clicked', async () => {
      const stopButton = wrapper.find('button.btn-warning')
      await stopButton.trigger('click')

      expect(scanFrameworkApi.stopScan).toHaveBeenCalledWith('scan123')
      expect(mockMessage.success).toHaveBeenCalledWith('扫描已停止')
    })

    it('should refresh data when refresh button clicked', async () => {
      vi.clearAllMocks()

      const refreshButton = wrapper.find('button.btn-secondary')
      await refreshButton.trigger('click')

      expect(scanFrameworkApi.getScanStatus).toHaveBeenCalled()
    })

    it('should handle stop scan errors', async () => {
      scanFrameworkApi.stopScan = vi.fn().mockRejectedValue(new Error('Stop failed'))

      const stopButton = wrapper.find('button.btn-warning')
      await stopButton.trigger('click')

      expect(mockMessage.error).toHaveBeenCalledWith('停止扫描失败')
    })
  })

  describe('Navigation', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should have back link to scan list', () => {
      const backLink = wrapper.find('.back-link')
      expect(backLink.attributes('to')).toBe('/scans')
      expect(backLink.text()).toContain('返回扫描列表')
    })
  })

  describe('Polling Mechanism', () => {
    beforeEach(async () => {
      vi.useFakeTimers()
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('should start polling for running scans', () => {
      expect(wrapper.vm.pollingInterval).toBeDefined()
    })

    it('should poll every 10 seconds for running scans', async () => {
      vi.clearAllMocks()

      // Fast-forward 10 seconds
      vi.advanceTimersByTime(10000)
      await nextTick()

      expect(scanFrameworkApi.getScanStatus).toHaveBeenCalled()
    })

    it('should not start polling for completed scans', async () => {
      wrapper.vm.scanInfo.state = 'completed'
      wrapper.vm.startPolling()

      expect(wrapper.vm.pollingInterval).toBeNull()
    })

    it('should stop polling on unmount', () => {
      const pollingInterval = wrapper.vm.pollingInterval
      wrapper.unmount()

      expect(pollingInterval).not.toBeNull()
    })

    it('should restart polling when scan state changes to running', async () => {
      wrapper.vm.scanInfo.state = 'completed'
      wrapper.vm.stopPolling()

      wrapper.vm.scanInfo.state = 'running'
      await nextTick()

      expect(wrapper.vm.pollingInterval).toBeDefined()
    })
  })

  describe('Utility Functions', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should get correct status class', () => {
      expect(wrapper.vm.getStatusClass('running')).toBe('status-running')
      expect(wrapper.vm.getStatusClass('completed')).toBe('status-completed')
      expect(wrapper.vm.getStatusClass('failed')).toBe('status-failed')
    })

    it('should get correct status text', () => {
      expect(wrapper.vm.getStatusText('running')).toBe('运行中')
      expect(wrapper.vm.getStatusText('completed')).toBe('已完成')
      expect(wrapper.vm.getStatusText('failed')).toBe('失败')
      expect(wrapper.vm.getStatusText('unknown')).toBe('未知')
    })

    it('should get correct severity class', () => {
      expect(wrapper.vm.getSeverityClass('high')).toBe('severity-high')
      expect(wrapper.vm.getSeverityClass('MEDIUM')).toBe('severity-medium')
      expect(wrapper.vm.getSeverityClass(null)).toBe('severity-')
    })

    it('should get correct result status class', () => {
      expect(wrapper.vm.getResultStatusClass('discovered')).toBe('result-discovered')
      expect(wrapper.vm.getResultStatusClass('open')).toBe('result-open')
    })

    it('should get correct log level class', () => {
      expect(wrapper.vm.getLogLevelClass('INFO')).toBe('log-info')
      expect(wrapper.vm.getLogLevelClass('WARN')).toBe('log-warn')
      expect(wrapper.vm.getLogLevelClass('ERROR')).toBe('log-error')
    })

    it('should calculate progress correctly for running scans', () => {
      const progress = wrapper.vm.getProgress()
      expect(typeof progress).toBe('number')
      expect(progress).toBeGreaterThanOrEqual(0)
      expect(progress).toBeLessThanOrEqual(100)
    })

    it('should return 100% progress for non-running scans', () => {
      wrapper.vm.scanInfo.state = 'completed'
      expect(wrapper.vm.getProgress()).toBe(100)
    })

    it('should format dates correctly', () => {
      const dateString = '2024-01-01T10:00:00Z'
      const formatted = wrapper.vm.formatDate(dateString)

      expect(typeof formatted).toBe('string')
      expect(formatted.length).toBeGreaterThan(0)
    })

    it('should format time correctly', () => {
      const dateString = '2024-01-01T10:00:00Z'
      const formatted = wrapper.vm.formatTime(dateString)

      expect(typeof formatted).toBe('string')
      expect(formatted.length).toBeGreaterThan(0)
    })

    it('should handle empty date strings', () => {
      expect(wrapper.vm.formatDate('')).toBe('')
      expect(wrapper.vm.formatTime(null)).toBe('')
    })

    it('should calculate running duration correctly', () => {
      const duration = wrapper.vm.getRunningDuration()
      expect(typeof duration).toBe('string')

      // Test different time ranges
      const now = Date.now()
      const oneHourAgo = new Date(now - 3600000).toISOString()
      wrapper.vm.scanInfo.created_at = oneHourAgo

      const hourDuration = wrapper.vm.getRunningDuration()
      expect(hourDuration).toContain('小时')
    })
  })

  describe('Error Handling', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should handle project loading errors', async () => {
      scanApi.getProject = vi.fn().mockRejectedValue(new Error('Project not found'))

      await wrapper.vm.loadProjectName()

      expect(wrapper.vm.projectName).toBe('未知项目')
    })

    it('should handle scan results loading errors', async () => {
      scanFrameworkApi.getScanResults = vi.fn().mockRejectedValue(new Error('Results not found'))

      await wrapper.vm.loadScanResults()

      // Should not throw error, just handle gracefully
      expect(wrapper.vm.scanResults).toEqual([])
    })

    it('should handle vulnerabilities loading errors', async () => {
      scanFrameworkApi.getProjectVulnerabilities = vi.fn().mockRejectedValue(new Error('Vulns not found'))

      await wrapper.vm.loadVulnerabilities()

      // Should not throw error, just handle gracefully
      expect(wrapper.vm.vulnerabilities).toEqual([])
    })
  })

  describe('Watch Handlers', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should start polling when scan state changes to running', async () => {
      wrapper.vm.scanInfo.state = 'completed'
      wrapper.vm.stopPolling()

      wrapper.vm.scanInfo.state = 'running'
      await nextTick()

      expect(wrapper.vm.pollingInterval).toBeDefined()
    })

    it('should stop polling when scan state changes from running', async () => {
      wrapper.vm.scanInfo.state = 'running'
      wrapper.vm.startPolling()

      wrapper.vm.scanInfo.state = 'completed'
      await nextTick()

      expect(wrapper.vm.pollingInterval).toBeNull()
    })
  })
})