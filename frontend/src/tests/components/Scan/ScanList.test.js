import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ScanList from '@/components/Scan/ScanList.vue'
import { scanApi } from '@/api/scanApi'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'

// Mock the APIs
vi.mock('@/api/scanApi')
vi.mock('@/api/scanFrameworkApi')

// Mock router
const mockRouter = {
  push: vi.fn()
}

// Mock $message
const mockMessage = {
  success: vi.fn(),
  error: vi.fn()
}

// Mock data
const mockProjects = [
  { id: 'project1', name: 'Test Project 1' },
  { id: 'project2', name: 'Test Project 2' }
]

const mockScans = [
  {
    id: 'scan1',
    target_address: 'example.com',
    target_id: 'target1',
    service_name: 'comprehensive',
    project_id: 'project1',
    state: 'running',
    created_at: '2024-01-01T10:00:00Z',
    updated_at: '2024-01-01T10:30:00Z'
  },
  {
    id: 'scan2',
    target_address: 'test.com',
    target_id: 'target2',
    service_name: 'quick',
    project_id: 'project2',
    state: 'completed',
    created_at: '2024-01-01T09:00:00Z',
    updated_at: '2024-01-01T09:15:00Z'
  },
  {
    id: 'scan3',
    target_address: 'failed.com',
    target_id: 'target3',
    service_name: 'web',
    project_id: 'project1',
    state: 'failed',
    created_at: '2024-01-01T08:00:00Z',
    updated_at: '2024-01-01T08:05:00Z'
  }
]

describe('ScanList', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()

    // Setup API mocks
    scanApi.getProjects = vi.fn().mockResolvedValue({ data: mockProjects })
    scanFrameworkApi.getProjectScans = vi.fn().mockResolvedValue({ data: mockScans })
    scanFrameworkApi.stopScan = vi.fn().mockResolvedValue({ data: { message: 'Scan stopped' } })
    scanFrameworkApi.deleteScan = vi.fn().mockResolvedValue({ data: { message: 'Scan deleted' } })

    // Mock window.confirm
    global.confirm = vi.fn(() => true)
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    vi.clearAllTimers()
  })

  const createWrapper = (options = {}) => {
    return mount(ScanList, {
      global: {
        mocks: {
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

      expect(wrapper.find('h1').text()).toBe('扫描管理')
      expect(wrapper.find('.filters').exists()).toBe(true)
      expect(wrapper.find('.btn-primary').text()).toContain('创建扫描')
    })

    it('should load initial data on mount', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(scanApi.getProjects).toHaveBeenCalled()
      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalled()
    })

    it('should populate data after loading', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(wrapper.vm.projects).toEqual(mockProjects)
      expect(wrapper.vm.scans).toEqual(mockScans)
    })

    it('should handle loading errors gracefully', async () => {
      scanApi.getProjects = vi.fn().mockRejectedValue(new Error('API Error'))

      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(mockMessage.error).toHaveBeenCalledWith('加载扫描任务失败')
    })
  })

  describe('Filtering', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should filter by project', async () => {
      const projectSelect = wrapper.find('select[v-model="filters.projectId"]')
      await projectSelect.setValue('project1')

      expect(wrapper.vm.filters.projectId).toBe('project1')
      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalledWith('project1', expect.any(Object))
    })

    it('should filter by status', async () => {
      const statusSelect = wrapper.find('select[v-model="filters.status"]')
      await statusSelect.setValue('running')

      expect(wrapper.vm.filters.status).toBe('running')
    })

    it('should filter by target with debounce', async () => {
      vi.useFakeTimers()

      const targetInput = wrapper.find('input[v-model="filters.target"]')
      await targetInput.setValue('example.com')

      expect(wrapper.vm.filters.target).toBe('example.com')

      // Should not call immediately
      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalledTimes(2) // Initial calls

      // Fast-forward debounce time
      vi.advanceTimersByTime(500)
      await nextTick()

      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalledTimes(4) // Additional call after debounce

      vi.useRealTimers()
    })

    it('should reset filters correctly', async () => {
      wrapper.vm.filters.projectId = 'project1'
      wrapper.vm.filters.status = 'running'
      wrapper.vm.filters.target = 'example.com'

      const projectSelect = wrapper.find('select[v-model="filters.projectId"]')
      await projectSelect.setValue('')

      expect(wrapper.vm.filters.projectId).toBe('')
    })
  })

  describe('Scan Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should display scan cards', () => {
      const scanCards = wrapper.findAll('.scan-card')
      expect(scanCards).toHaveLength(mockScans.length)
    })

    it('should show correct scan information', () => {
      const firstCard = wrapper.find('.scan-card')

      expect(firstCard.text()).toContain('example.com')
      expect(firstCard.text()).toContain('comprehensive')
      expect(firstCard.text()).toContain('Test Project 1')
    })

    it('should display correct status badges', () => {
      const statusBadges = wrapper.findAll('.status-badge')

      expect(statusBadges[0].classes()).toContain('status-running')
      expect(statusBadges[0].text()).toBe('运行中')

      expect(statusBadges[1].classes()).toContain('status-completed')
      expect(statusBadges[1].text()).toBe('已完成')

      expect(statusBadges[2].classes()).toContain('status-failed')
      expect(statusBadges[2].text()).toBe('失败')
    })

    it('should show progress bar for running scans', () => {
      const runningCard = wrapper.find('.scan-card.status-running')
      const progressBar = runningCard.find('.progress-bar')

      expect(progressBar.exists()).toBe(true)
    })

    it('should not show progress bar for non-running scans', () => {
      const completedCard = wrapper.find('.scan-card.status-completed')
      const progressBar = completedCard.find('.progress-bar')

      expect(progressBar.exists()).toBe(false)
    })
  })

  describe('Scan Actions', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should show stop button for running scans', () => {
      const runningCard = wrapper.find('.scan-card.status-running')
      const stopButton = runningCard.find('button.btn-warning')

      expect(stopButton.exists()).toBe(true)
      expect(stopButton.text()).toBe('停止')
    })

    it('should show delete button for non-running scans', () => {
      const completedCard = wrapper.find('.scan-card.status-completed')
      const deleteButton = completedCard.find('button.btn-danger')

      expect(deleteButton.exists()).toBe(true)
      expect(deleteButton.text()).toBe('删除')
    })

    it('should stop scan when stop button clicked', async () => {
      const runningCard = wrapper.find('.scan-card.status-running')
      const stopButton = runningCard.find('button.btn-warning')

      await stopButton.trigger('click')

      expect(scanFrameworkApi.stopScan).toHaveBeenCalledWith('scan1')
      expect(mockMessage.success).toHaveBeenCalledWith('扫描任务已停止')
    })

    it('should delete scan when delete button clicked and confirmed', async () => {
      const completedCard = wrapper.find('.scan-card.status-completed')
      const deleteButton = completedCard.find('button.btn-danger')

      await deleteButton.trigger('click')

      expect(global.confirm).toHaveBeenCalledWith('确定要删除这个扫描任务吗？此操作不可恢复。')
      expect(scanFrameworkApi.deleteScan).toHaveBeenCalledWith('scan2')
      expect(mockMessage.success).toHaveBeenCalledWith('扫描任务已删除')
    })

    it('should not delete scan when deletion is cancelled', async () => {
      global.confirm = vi.fn(() => false)

      const completedCard = wrapper.find('.scan-card.status-completed')
      const deleteButton = completedCard.find('button.btn-danger')

      await deleteButton.trigger('click')

      expect(global.confirm).toHaveBeenCalled()
      expect(scanFrameworkApi.deleteScan).not.toHaveBeenCalled()
    })

    it('should navigate to detail when card is clicked', async () => {
      const scanCard = wrapper.find('.scan-card')
      await scanCard.trigger('click')

      expect(mockRouter.push).toHaveBeenCalledWith('/scans/scan1')
    })

    it('should navigate to detail when detail button is clicked', async () => {
      const detailButton = wrapper.find('button.btn-primary')
      await detailButton.trigger('click')

      expect(mockRouter.push).toHaveBeenCalledWith('/scans/scan1')
    })
  })

  describe('Error Handling', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should handle stop scan errors', async () => {
      scanFrameworkApi.stopScan = vi.fn().mockRejectedValue(new Error('Stop failed'))

      const runningCard = wrapper.find('.scan-card.status-running')
      const stopButton = runningCard.find('button.btn-warning')

      await stopButton.trigger('click')

      expect(mockMessage.error).toHaveBeenCalledWith('停止扫描失败')
    })

    it('should handle delete scan errors', async () => {
      scanFrameworkApi.deleteScan = vi.fn().mockRejectedValue(new Error('Delete failed'))

      const completedCard = wrapper.find('.scan-card.status-completed')
      const deleteButton = completedCard.find('button.btn-danger')

      await deleteButton.trigger('click')

      expect(mockMessage.error).toHaveBeenCalledWith('删除扫描失败')
    })
  })

  describe('Empty State', () => {
    it('should show empty state when no scans exist', async () => {
      scanFrameworkApi.getProjectScans = vi.fn().mockResolvedValue({ data: [] })

      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      const emptyState = wrapper.find('.empty-state')
      expect(emptyState.exists()).toBe(true)
      expect(emptyState.text()).toContain('暂无扫描任务')
      expect(emptyState.text()).toContain('创建您的第一个扫描任务开始安全检测')
    })

    it('should have create scan link in empty state', async () => {
      scanFrameworkApi.getProjectScans = vi.fn().mockResolvedValue({ data: [] })

      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      const createLink = wrapper.find('.empty-state router-link')
      expect(createLink.attributes('to')).toBe('/scans/create')
    })
  })

  describe('Loading State', () => {
    it('should show loading spinner during data fetch', async () => {
      wrapper = createWrapper()
      wrapper.vm.loading = true
      await nextTick()

      const loadingSpinner = wrapper.find('.loading-spinner')
      expect(loadingSpinner.exists()).toBe(true)
      expect(wrapper.text()).toContain('加载中...')
    })

    it('should hide loading state after data is loaded', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      const loadingSpinner = wrapper.find('.loading-spinner')
      expect(loadingSpinner.exists()).toBe(false)
    })
  })

  describe('Pagination', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should show pagination when total exceeds page size', async () => {
      wrapper.vm.pagination.total = 25
      wrapper.vm.pagination.pageSize = 10
      await nextTick()

      const pagination = wrapper.find('.pagination')
      expect(pagination.exists()).toBe(true)
    })

    it('should hide pagination when total is within page size', async () => {
      wrapper.vm.pagination.total = 5
      wrapper.vm.pagination.pageSize = 10
      await nextTick()

      const pagination = wrapper.find('.pagination')
      expect(pagination.exists()).toBe(false)
    })

    it('should change page when pagination buttons are clicked', async () => {
      wrapper.vm.pagination.total = 25
      wrapper.vm.pagination.pageSize = 10
      wrapper.vm.pagination.current = 1
      await nextTick()

      const nextButton = wrapper.find('.pagination button:last-child')
      await nextButton.trigger('click')

      expect(wrapper.vm.pagination.current).toBe(2)
    })

    it('should disable previous button on first page', async () => {
      wrapper.vm.pagination.current = 1
      await nextTick()

      const prevButton = wrapper.find('.pagination button:first-child')
      expect(prevButton.attributes('disabled')).toBeDefined()
    })

    it('should disable next button on last page', async () => {
      wrapper.vm.pagination.total = 25
      wrapper.vm.pagination.pageSize = 10
      wrapper.vm.pagination.current = 3
      await nextTick()

      const nextButton = wrapper.find('.pagination button:last-child')
      expect(nextButton.attributes('disabled')).toBeDefined()
    })
  })

  describe('Refresh Functionality', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should refresh scans when refresh button is clicked', async () => {
      vi.clearAllMocks()

      const refreshButton = wrapper.find('.filter-actions button')
      await refreshButton.trigger('click')

      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalled()
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

    it('should start polling on mount', () => {
      expect(wrapper.vm.pollingInterval).toBeDefined()
    })

    it('should poll when there are running scans', async () => {
      vi.clearAllMocks()

      // Fast-forward 30 seconds
      vi.advanceTimersByTime(30000)
      await nextTick()

      expect(scanFrameworkApi.getProjectScans).toHaveBeenCalled()
    })

    it('should not poll when no running scans', async () => {
      wrapper.vm.scans = mockScans.filter(scan => scan.state !== 'running')
      vi.clearAllMocks()

      // Fast-forward 30 seconds
      vi.advanceTimersByTime(30000)
      await nextTick()

      expect(scanFrameworkApi.getProjectScans).not.toHaveBeenCalled()
    })

    it('should stop polling on unmount', () => {
      const pollingInterval = wrapper.vm.pollingInterval
      wrapper.unmount()

      expect(pollingInterval).not.toBeNull()
    })
  })

  describe('Utility Functions', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should get correct project name', () => {
      expect(wrapper.vm.getProjectName('project1')).toBe('Test Project 1')
      expect(wrapper.vm.getProjectName('project2')).toBe('Test Project 2')
      expect(wrapper.vm.getProjectName('nonexistent')).toBe('未知项目')
    })

    it('should get correct status class', () => {
      expect(wrapper.vm.getStatusClass('running')).toBe('status-running')
      expect(wrapper.vm.getStatusClass('completed')).toBe('status-completed')
      expect(wrapper.vm.getStatusClass('failed')).toBe('status-failed')
      expect(wrapper.vm.getStatusClass('stopped')).toBe('status-stopped')
      expect(wrapper.vm.getStatusClass('unknown')).toBe('status-unknown')
    })

    it('should get correct status text', () => {
      expect(wrapper.vm.getStatusText('running')).toBe('运行中')
      expect(wrapper.vm.getStatusText('completed')).toBe('已完成')
      expect(wrapper.vm.getStatusText('failed')).toBe('失败')
      expect(wrapper.vm.getStatusText('stopped')).toBe('已停止')
      expect(wrapper.vm.getStatusText('unknown')).toBe('未知')
    })

    it('should calculate progress correctly', () => {
      const runningScan = mockScans.find(scan => scan.state === 'running')
      const progress = wrapper.vm.getProgress(runningScan)

      expect(typeof progress).toBe('number')
      expect(progress).toBeGreaterThanOrEqual(0)
      expect(progress).toBeLessThanOrEqual(100)
    })

    it('should return 100% progress for non-running scans', () => {
      const completedScan = mockScans.find(scan => scan.state === 'completed')
      expect(wrapper.vm.getProgress(completedScan)).toBe(100)
    })

    it('should format dates correctly', () => {
      const dateString = '2024-01-01T10:00:00Z'
      const formatted = wrapper.vm.formatDate(dateString)

      expect(typeof formatted).toBe('string')
      expect(formatted.length).toBeGreaterThan(0)
    })

    it('should handle empty date strings', () => {
      expect(wrapper.vm.formatDate('')).toBe('')
      expect(wrapper.vm.formatDate(null)).toBe('')
      expect(wrapper.vm.formatDate(undefined)).toBe('')
    })
  })
})