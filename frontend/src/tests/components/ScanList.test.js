import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createStore } from 'vuex'
import ScanList from '@/components/Scan/ScanList.vue'
import { notification } from 'ant-design-vue'
import storeConfig from '@/store/index.js'

// Mock router
const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/scans', name: 'ScanList' },
    { path: '/scans/:id', name: 'ScanDetail' }
  ]
})

// Mock notifications
vi.mock('ant-design-vue', async () => {
  const actual = await vi.importActual('ant-design-vue')
  return {
    ...actual,
    notification: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn()
    }
  }
})

// Mock scan API
vi.mock('@/api/scanApi', () => ({
  getScans: vi.fn(),
  deleteScan: vi.fn(),
  getScanResults: vi.fn(),
  updateScanStatus: vi.fn()
}))

// Mock scan framework API
vi.mock('@/api/scanFrameworkApi', () => ({
  getProjects: vi.fn(),
  createProject: vi.fn(),
  startScan: vi.fn(),
  getScanStatus: vi.fn()
}))

describe('ScanList.vue', () => {
  let wrapper

  const mockScans = [
    {
      id: 1,
      name: 'Test Scan 1',
      target: 'example.com',
      status: 'completed',
      scan_type: 'port_scan',
      created_at: '2024-01-01T10:00:00Z',
      updated_at: '2024-01-01T11:00:00Z',
      progress: 100,
      results_count: 25,
      vulnerabilities_count: 3
    },
    {
      id: 2,
      name: 'Test Scan 2',
      target: '192.168.1.1',
      status: 'running',
      scan_type: 'vulnerability_scan',
      created_at: '2024-01-02T10:00:00Z',
      updated_at: '2024-01-02T10:30:00Z',
      progress: 65,
      results_count: 15,
      vulnerabilities_count: 1
    },
    {
      id: 3,
      name: 'Test Scan 3',
      target: 'test.example.com',
      status: 'failed',
      scan_type: 'subdomain_scan',
      created_at: '2024-01-03T10:00:00Z',
      updated_at: '2024-01-03T10:05:00Z',
      progress: 0,
      results_count: 0,
      vulnerabilities_count: 0
    }
  ]

  beforeEach(async () => {
    vi.clearAllMocks()

    const { getScans } = await import('@/api/scanApi')
    getScans.mockResolvedValue({ data: mockScans })

    const store = createStore({
      state: {
        isAuthenticated: false,
        user: null,
      },
      getters: {
        isAuthenticated: (state) => state.isAuthenticated,
        currentUser: (state) => state.user,
        isAdmin: (state) => state.user?.role === 'admin',
      },
      mutations: {
        setAuthentication(state, status) {
          state.isAuthenticated = status;
        },
        setUser(state, user) {
          state.user = user;
        },
      },
      actions: {
        async login({ commit }, { token }) {
          localStorage.setItem("token", token);
          commit("setAuthentication", true);
          return true;
        },
        async logout({ commit }) {
          localStorage.removeItem("token");
          commit("setAuthentication", false);
          commit("setUser", null);
        },
      }
    })
    wrapper = mount(ScanList, {
      global: {
        plugins: [router, store]
      }
    })

    await wrapper.vm.$nextTick()
  })

  it('renders scan list correctly', () => {
    expect(wrapper.find('[data-testid="scan-list-table"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="create-scan-button"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="refresh-button"]').exists()).toBe(true)
  })

  it('displays scan data in table', async () => {
    // Wait for data to load
    await wrapper.vm.$nextTick()

    const tableRows = wrapper.findAll('[data-testid^="scan-row-"]')
    expect(tableRows).toHaveLength(3)

    // Check first scan data
    const firstRow = wrapper.find('[data-testid="scan-row-1"]')
    expect(firstRow.text()).toContain('Test Scan 1')
    expect(firstRow.text()).toContain('example.com')
    expect(firstRow.text()).toContain('completed')
  })

  it('filters scans by status', async () => {
    const statusFilter = wrapper.find('[data-testid="status-filter"]')

    // Filter by running status
    await statusFilter.trigger('click')
    const runningOption = wrapper.find('[data-testid="status-filter-running"]')
    await runningOption.trigger('click')

    // Should only show running scans
    const visibleRows = wrapper.findAll('[data-testid^="scan-row-"]:not(.ant-table-row-hidden)')
    expect(visibleRows).toHaveLength(1)
    expect(wrapper.find('[data-testid="scan-row-2"]').isVisible()).toBe(true)
  })

  it('filters scans by type', async () => {
    const typeFilter = wrapper.find('[data-testid="type-filter"]')

    // Filter by port scan type
    await typeFilter.trigger('click')
    const portScanOption = wrapper.find('[data-testid="type-filter-port_scan"]')
    await portScanOption.trigger('click')

    // Should only show port scans
    expect(wrapper.find('[data-testid="scan-row-1"]').isVisible()).toBe(true)
  })

  it('searches scans by name and target', async () => {
    const searchInput = wrapper.find('[data-testid="scan-search"]')

    // Search for specific scan name
    await searchInput.setValue('Test Scan 1')
    await searchInput.trigger('input')

    // Should show only matching scans
    expect(wrapper.vm.filteredScans).toHaveLength(1)
    expect(wrapper.vm.filteredScans[0].name).toBe('Test Scan 1')

    // Search by target
    await searchInput.setValue('192.168.1.1')
    await searchInput.trigger('input')

    expect(wrapper.vm.filteredScans).toHaveLength(1)
    expect(wrapper.vm.filteredScans[0].target).toBe('192.168.1.1')
  })

  it('navigates to scan detail page', async () => {
    const viewButton = wrapper.find('[data-testid="view-scan-1"]')
    await viewButton.trigger('click')

    expect(wrapper.vm.$route.path).toBe('/scans/1')
  })

  it('opens create scan modal', async () => {
    const createButton = wrapper.find('[data-testid="create-scan-button"]')
    await createButton.trigger('click')

    expect(wrapper.find('[data-testid="create-scan-modal"]').exists()).toBe(true)
    expect(wrapper.vm.showCreateModal).toBe(true)
  })

  it('creates new scan successfully', async () => {
    const { startScan } = await import('@/api/scanFrameworkApi')
    startScan.mockResolvedValue({
      data: { id: 4, name: 'New Scan', status: 'pending' }
    })

    // Open create modal
    await wrapper.setData({ showCreateModal: true })

    // Fill scan form
    await wrapper.find('[data-testid="scan-name-input"]').setValue('New Test Scan')
    await wrapper.find('[data-testid="scan-target-input"]').setValue('newtest.com')
    await wrapper.find('[data-testid="scan-type-select"]').setValue('port_scan')

    // Submit form
    const submitButton = wrapper.find('[data-testid="create-scan-submit"]')
    await submitButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(startScan).toHaveBeenCalledWith({
      name: 'New Test Scan',
      target: 'newtest.com',
      scan_type: 'port_scan'
    })
    expect(notification.success).toHaveBeenCalledWith({
      message: '扫描创建成功',
      description: '扫描任务已开始执行'
    })
  })

  it('validates scan creation form', async () => {
    await wrapper.setData({ showCreateModal: true })

    // Try to submit empty form
    const submitButton = wrapper.find('[data-testid="create-scan-submit"]')
    await submitButton.trigger('click')

    // Should show validation errors
    expect(wrapper.text()).toContain('请输入扫描名称')
    expect(wrapper.text()).toContain('请输入扫描目标')
  })

  it('deletes scan after confirmation', async () => {
    const { deleteScan } = await import('@/api/scanApi')
    deleteScan.mockResolvedValue({ data: { success: true } })

    // Mock confirm dialog
    window.confirm = vi.fn().mockReturnValue(true)

    const deleteButton = wrapper.find('[data-testid="delete-scan-1"]')
    await deleteButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(window.confirm).toHaveBeenCalledWith('确定要删除这个扫描任务吗？')
    expect(deleteScan).toHaveBeenCalledWith(1)
    expect(notification.success).toHaveBeenCalledWith({
      message: '删除成功',
      description: '扫描任务已删除'
    })
  })

  it('cancels scan deletion', async () => {
    const { deleteScan } = await import('@/api/scanApi')

    // Mock confirm dialog to return false
    window.confirm = vi.fn().mockReturnValue(false)

    const deleteButton = wrapper.find('[data-testid="delete-scan-1"]')
    await deleteButton.trigger('click')

    expect(window.confirm).toHaveBeenCalled()
    expect(deleteScan).not.toHaveBeenCalled()
  })

  it('refreshes scan list', async () => {
    const { getScans } = await import('@/api/scanApi')
    getScans.mockClear()

    const refreshButton = wrapper.find('[data-testid="refresh-button"]')
    await refreshButton.trigger('click')

    expect(getScans).toHaveBeenCalled()
    expect(wrapper.vm.loading).toBe(false)
  })

  it('displays correct status badges', () => {
    // Completed scan should have success badge
    const completedBadge = wrapper.find('[data-testid="status-badge-1"]')
    expect(completedBadge.classes()).toContain('ant-tag-success')

    // Running scan should have processing badge
    const runningBadge = wrapper.find('[data-testid="status-badge-2"]')
    expect(runningBadge.classes()).toContain('ant-tag-processing')

    // Failed scan should have error badge
    const failedBadge = wrapper.find('[data-testid="status-badge-3"]')
    expect(failedBadge.classes()).toContain('ant-tag-error')
  })

  it('shows progress bars for running scans', () => {
    // Running scan should show progress bar
    const progressBar = wrapper.find('[data-testid="progress-bar-2"]')
    expect(progressBar.exists()).toBe(true)
    expect(progressBar.attributes('percent')).toBe('65')

    // Completed scan should not show progress bar
    const completedProgress = wrapper.find('[data-testid="progress-bar-1"]')
    expect(completedProgress.exists()).toBe(false)
  })

  it('handles loading state', async () => {
    await wrapper.setData({ loading: true })

    const table = wrapper.find('[data-testid="scan-list-table"]')
    expect(table.attributes('loading')).toBeDefined()
  })

  it('handles empty scan list', async () => {
    const { getScans } = await import('@/api/scanApi')
    getScans.mockResolvedValue({ data: [] })

    await wrapper.vm.loadScans()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="empty-state"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('暂无扫描任务')
  })

  it('handles API errors', async () => {
    const { getScans } = await import('@/api/scanApi')
    getScans.mockRejectedValue(new Error('API Error'))

    await wrapper.vm.loadScans()
    await wrapper.vm.$nextTick()

    expect(notification.error).toHaveBeenCalledWith({
      message: '加载失败',
      description: '无法获取扫描列表，请稍后重试'
    })
  })

  it('formats dates correctly', () => {
    // Should display formatted dates
    const timeCell = wrapper.find('[data-testid="created-time-1"]')
    expect(timeCell.text()).toMatch(/\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/)
  })

  it('sorts scans by different columns', async () => {
    // Click on name column to sort
    const nameHeader = wrapper.find('[data-testid="name-column-header"]')
    await nameHeader.trigger('click')

    // Should sort scans by name
    expect(wrapper.vm.sortBy).toBe('name')
    expect(wrapper.vm.sortOrder).toBe('asc')

    // Click again to reverse sort
    await nameHeader.trigger('click')
    expect(wrapper.vm.sortOrder).toBe('desc')
  })

  it('updates scan status automatically for running scans', async () => {
    const { getScanStatus } = await import('@/api/scanFrameworkApi')
    getScanStatus.mockResolvedValue({
      data: { id: 2, status: 'completed', progress: 100 }
    })

    // Trigger status update
    await wrapper.vm.updateRunningScans()

    expect(getScanStatus).toHaveBeenCalledWith(2)
    expect(wrapper.vm.scans[1].status).toBe('completed')
    expect(wrapper.vm.scans[1].progress).toBe(100)
  })
})