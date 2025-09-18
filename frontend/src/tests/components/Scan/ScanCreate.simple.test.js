import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ScanCreate from '@/components/Scan/ScanCreate.vue'

// Mock the APIs
vi.mock('@/api/scanApi', () => ({
  scanApi: {
    getProjects: vi.fn().mockResolvedValue({ data: [] })
  }
}))

vi.mock('@/api/scanFrameworkApi', () => ({
  scanFrameworkApi: {
    getAvailablePipelines: vi.fn().mockResolvedValue({ data: [] }),
    getAvailableTools: vi.fn().mockResolvedValue({ data: {} }),
    startScan: vi.fn().mockResolvedValue({ data: { scan_id: 'test123' } })
  }
}))

describe('ScanCreate - Simple Tests', () => {
  let wrapper

  const createWrapper = () => {
    return mount(ScanCreate, {
      global: {
        mocks: {
          $router: { push: vi.fn() },
          $message: { success: vi.fn(), error: vi.fn() }
        }
      }
    })
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render the component', () => {
    wrapper = createWrapper()
    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('h1').text()).toBe('创建扫描任务')
  })

  it('should have correct initial data', () => {
    wrapper = createWrapper()
    expect(wrapper.vm.loading).toBe(false)
    expect(wrapper.vm.showAdvanced).toBe(false)
    expect(wrapper.vm.projects).toEqual([])
    expect(wrapper.vm.pipelines).toEqual([])
    expect(wrapper.vm.availableTools).toEqual([])
  })

  it('should have correct form structure', () => {
    wrapper = createWrapper()
    expect(wrapper.vm.scanForm.projectId).toBe('')
    expect(wrapper.vm.scanForm.target).toBe('')
    expect(wrapper.vm.scanForm.pipelineName).toBe('')
    expect(wrapper.vm.scanForm.timeout).toBe(30)
    expect(wrapper.vm.scanForm.concurrent).toBe(3)
  })

  it('should validate form correctly', () => {
    wrapper = createWrapper()

    // Initially invalid
    expect(wrapper.vm.isFormValid).toBe(false)

    // Set required fields
    wrapper.vm.scanForm.projectId = 'project1'
    wrapper.vm.scanForm.target = 'example.com'
    wrapper.vm.scanForm.pipelineName = 'comprehensive'

    // Now should be valid
    expect(wrapper.vm.isFormValid).toBe(true)
  })

  it('should have utility methods', () => {
    wrapper = createWrapper()

    expect(wrapper.vm.getPipelineDisplayName('comprehensive')).toBe('全面扫描')
    expect(wrapper.vm.getCategoryDisplayName('subdomain')).toBe('子域名发现')
    expect(wrapper.vm.flattenTools).toBeDefined()
  })

  it('should toggle advanced options', async () => {
    wrapper = createWrapper()

    expect(wrapper.vm.showAdvanced).toBe(false)

    // Simulate toggle
    wrapper.vm.showAdvanced = true

    expect(wrapper.vm.showAdvanced).toBe(true)
  })
})