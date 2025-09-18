import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ScanCreate from '@/components/Scan/ScanCreate.vue'
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

const mockPipelines = ['comprehensive', 'quick', 'web', 'network']

const mockTools = {
  subdomain: [
    { name: 'subfinder', available: true },
    { name: 'amass', available: false }
  ],
  port: [
    { name: 'nmap', available: true }
  ],
  vulnerability: [
    { name: 'nuclei', available: true }
  ]
}

describe('ScanCreate', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()

    // Setup API mocks
    scanApi.getProjects = vi.fn().mockResolvedValue({ data: mockProjects })
    scanFrameworkApi.getAvailablePipelines = vi.fn().mockResolvedValue({ data: mockPipelines })
    scanFrameworkApi.getAvailableTools = vi.fn().mockResolvedValue({ data: mockTools })
    scanFrameworkApi.startScan = vi.fn().mockResolvedValue({
      data: { scan_id: 'scan123' }
    })
  })

  const createWrapper = (options = {}) => {
    return mount(ScanCreate, {
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

      expect(wrapper.find('h1').text()).toBe('创建扫描任务')
      expect(wrapper.find('.breadcrumb').exists()).toBe(true)
      expect(wrapper.find('form').exists()).toBe(true)
    })

    it('should load initial data on mount', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick() // Wait for async mounted

      expect(scanApi.getProjects).toHaveBeenCalled()
      expect(scanFrameworkApi.getAvailablePipelines).toHaveBeenCalled()
      expect(scanFrameworkApi.getAvailableTools).toHaveBeenCalled()
    })

    it('should populate form data after loading', async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(wrapper.vm.projects).toEqual(mockProjects)
      expect(wrapper.vm.pipelines).toEqual(mockPipelines)
      expect(wrapper.vm.availableTools).toHaveLength(4) // flattened tools
    })

    it('should handle data loading errors gracefully', async () => {
      scanApi.getProjects = vi.fn().mockRejectedValue(new Error('API Error'))

      wrapper = createWrapper()
      await nextTick()
      await nextTick()

      expect(mockMessage.error).toHaveBeenCalledWith('加载数据失败，请刷新页面重试')
    })
  })

  describe('Form Validation', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should validate required fields', () => {
      expect(wrapper.vm.isFormValid).toBe(false)

      wrapper.vm.scanForm.projectId = 'project1'
      expect(wrapper.vm.isFormValid).toBe(false)

      wrapper.vm.scanForm.target = 'example.com'
      expect(wrapper.vm.isFormValid).toBe(false)

      wrapper.vm.scanForm.pipelineName = 'comprehensive'
      expect(wrapper.vm.isFormValid).toBe(true)
    })

    it('should disable submit button when form is invalid', async () => {
      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })

    it('should enable submit button when form is valid', async () => {
      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }
      await nextTick()

      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.attributes('disabled')).toBeUndefined()
    })
  })

  describe('Form Interactions', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should update form data when selecting project', async () => {
      const projectSelect = wrapper.find('select#project')
      await projectSelect.setValue('project1')

      expect(wrapper.vm.scanForm.projectId).toBe('project1')
    })

    it('should update form data when entering target', async () => {
      const targetInput = wrapper.find('input#target')
      await targetInput.setValue('example.com')

      expect(wrapper.vm.scanForm.target).toBe('example.com')
    })

    it('should update form data when selecting pipeline', async () => {
      const pipelineSelect = wrapper.find('select#pipeline')
      await pipelineSelect.setValue('comprehensive')

      expect(wrapper.vm.scanForm.pipelineName).toBe('comprehensive')
    })

    it('should show pipeline info when pipeline is selected', async () => {
      wrapper.vm.scanForm.pipelineName = 'comprehensive'
      await nextTick()

      expect(wrapper.vm.selectedPipelineInfo).toBe('全面扫描 - 包含子域名发现、端口扫描、服务识别、漏洞检测')
    })
  })

  describe('Advanced Configuration', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should toggle advanced options visibility', async () => {
      expect(wrapper.vm.showAdvanced).toBe(false)
      expect(wrapper.find('.advanced-options').exists()).toBe(false)

      await wrapper.find('.advanced-toggle').trigger('click')

      expect(wrapper.vm.showAdvanced).toBe(true)
      await nextTick()
      expect(wrapper.find('.advanced-options').exists()).toBe(true)
    })

    it('should update advanced configuration values', async () => {
      wrapper.vm.showAdvanced = true
      await nextTick()

      const timeoutInput = wrapper.find('input#timeout')
      const concurrentInput = wrapper.find('input#concurrent')

      await timeoutInput.setValue('60')
      await concurrentInput.setValue('5')

      expect(wrapper.vm.scanForm.timeout).toBe(60)
      expect(wrapper.vm.scanForm.concurrent).toBe(5)
    })
  })

  describe('Tools Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should display tools by category', () => {
      const toolsByCategory = wrapper.vm.toolsByCategory

      expect(toolsByCategory.subdomain).toHaveLength(2)
      expect(toolsByCategory.port).toHaveLength(2)
      expect(toolsByCategory.vulnerability).toHaveLength(1)
    })

    it('should show tool availability status', async () => {
      wrapper.vm.scanForm.pipelineName = 'comprehensive'
      await nextTick()

      const toolItems = wrapper.findAll('.tool-item')
      expect(toolItems.length).toBeGreaterThan(0)

      const availableTools = wrapper.findAll('.tool-item.available')
      const unavailableTools = wrapper.findAll('.tool-item.unavailable')

      expect(availableTools.length).toBeGreaterThan(0)
      expect(unavailableTools.length).toBeGreaterThan(0)
    })

    it('should display correct category names', () => {
      expect(wrapper.vm.getCategoryDisplayName('subdomain')).toBe('子域名发现')
      expect(wrapper.vm.getCategoryDisplayName('port')).toBe('端口扫描')
      expect(wrapper.vm.getCategoryDisplayName('vulnerability')).toBe('漏洞检测')
    })

    it('should display correct pipeline names', () => {
      expect(wrapper.vm.getPipelineDisplayName('comprehensive')).toBe('全面扫描')
      expect(wrapper.vm.getPipelineDisplayName('quick')).toBe('快速扫描')
      expect(wrapper.vm.getPipelineDisplayName('web')).toBe('Web扫描')
      expect(wrapper.vm.getPipelineDisplayName('network')).toBe('网络扫描')
    })
  })

  describe('Form Submission', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should submit form with valid data', async () => {
      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.find('form').trigger('submit.prevent')

      expect(scanFrameworkApi.startScan).toHaveBeenCalledWith(
        'project1',
        'example.com',
        'comprehensive'
      )
    })

    it('should show success message and redirect after successful submission', async () => {
      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.vm.createScan()

      expect(mockMessage.success).toHaveBeenCalledWith('扫描任务创建成功！')
      expect(mockRouter.push).toHaveBeenCalledWith('/scans/scan123')
    })

    it('should redirect to scan list if no scan_id returned', async () => {
      scanFrameworkApi.startScan = vi.fn().mockResolvedValue({ data: {} })

      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.vm.createScan()

      expect(mockRouter.push).toHaveBeenCalledWith('/scans')
    })

    it('should handle submission errors', async () => {
      const error = new Error('Submission failed')
      error.response = { data: { error: 'Invalid target' } }
      scanFrameworkApi.startScan = vi.fn().mockRejectedValue(error)

      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'invalid-target',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.vm.createScan()

      expect(mockMessage.error).toHaveBeenCalledWith('Invalid target')
    })

    it('should not submit if form is invalid', async () => {
      wrapper.vm.scanForm = {
        projectId: '',
        target: '',
        pipelineName: '',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.vm.createScan()

      expect(scanFrameworkApi.startScan).not.toHaveBeenCalled()
    })
  })

  describe('Loading States', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should show loading state during data loading', async () => {
      wrapper.vm.loading = true
      await nextTick()

      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.text()).toContain('创建中...')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })

    it('should show normal state when not loading', async () => {
      wrapper.vm.loading = false
      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }
      await nextTick()

      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.text()).toContain('开始扫描')
    })

    it('should disable form during submission', async () => {
      wrapper.vm.loading = true
      await nextTick()

      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })
  })

  describe('Navigation', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should navigate back to scan list on cancel', async () => {
      const cancelButton = wrapper.find('button.btn-secondary')
      await cancelButton.trigger('click')

      expect(mockRouter.push).toHaveBeenCalledWith('/scans')
    })

    it('should have correct breadcrumb link', () => {
      const breadcrumbLink = wrapper.find('.breadcrumb router-link')
      expect(breadcrumbLink.attributes('to')).toBe('/scans')
    })
  })

  describe('Input Validation', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should accept valid domain names', async () => {
      const targetInput = wrapper.find('input#target')

      await targetInput.setValue('example.com')
      expect(wrapper.vm.scanForm.target).toBe('example.com')

      await targetInput.setValue('sub.example.com')
      expect(wrapper.vm.scanForm.target).toBe('sub.example.com')
    })

    it('should accept valid IP addresses', async () => {
      const targetInput = wrapper.find('input#target')

      await targetInput.setValue('192.168.1.1')
      expect(wrapper.vm.scanForm.target).toBe('192.168.1.1')
    })

    it('should accept valid IP ranges', async () => {
      const targetInput = wrapper.find('input#target')

      await targetInput.setValue('192.168.1.0/24')
      expect(wrapper.vm.scanForm.target).toBe('192.168.1.0/24')
    })

    it('should validate timeout range', async () => {
      wrapper.vm.showAdvanced = true
      await nextTick()

      const timeoutInput = wrapper.find('input#timeout')
      expect(timeoutInput.attributes('min')).toBe('1')
      expect(timeoutInput.attributes('max')).toBe('180')
    })

    it('should validate concurrent range', async () => {
      wrapper.vm.showAdvanced = true
      await nextTick()

      const concurrentInput = wrapper.find('input#concurrent')
      expect(concurrentInput.attributes('min')).toBe('1')
      expect(concurrentInput.attributes('max')).toBe('10')
    })
  })

  describe('Data Processing', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should flatten tools data correctly', () => {
      const flattenedTools = wrapper.vm.flattenTools(mockTools)

      expect(flattenedTools).toHaveLength(4)
      expect(flattenedTools[0]).toMatchObject({
        name: 'subfinder',
        available: true,
        category: 'subdomain'
      })
    })

    it('should group tools by category correctly', () => {
      const toolsByCategory = wrapper.vm.toolsByCategory

      expect(Object.keys(toolsByCategory)).toEqual(['subdomain', 'port', 'vulnerability'])
      expect(toolsByCategory.subdomain).toHaveLength(2)
      expect(toolsByCategory.port).toHaveLength(2)
      expect(toolsByCategory.vulnerability).toHaveLength(1)
    })
  })

  describe('Error Scenarios', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await nextTick()
      await nextTick()
    })

    it('should handle empty API responses', async () => {
      scanApi.getProjects = vi.fn().mockResolvedValue({ data: null })
      scanFrameworkApi.getAvailablePipelines = vi.fn().mockResolvedValue({ data: null })
      scanFrameworkApi.getAvailableTools = vi.fn().mockResolvedValue({ data: null })

      await wrapper.vm.loadData()

      expect(wrapper.vm.projects).toEqual([])
      expect(wrapper.vm.pipelines).toEqual([])
      expect(wrapper.vm.availableTools).toEqual([])
    })

    it('should handle network errors during form submission', async () => {
      const networkError = new Error('Network Error')
      scanFrameworkApi.startScan = vi.fn().mockRejectedValue(networkError)

      wrapper.vm.scanForm = {
        projectId: 'project1',
        target: 'example.com',
        pipelineName: 'comprehensive',
        timeout: 30,
        concurrent: 3
      }

      await wrapper.vm.createScan()

      expect(mockMessage.error).toHaveBeenCalledWith('创建扫描失败')
    })
  })
})