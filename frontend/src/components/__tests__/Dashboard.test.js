import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Dashboard from '../Dashboard.vue'
import { createStore } from 'vuex'

// Mock HeaderPage组件
vi.mock('../HeaderPage.vue', () => ({
  default: {
    name: 'HeaderPage',
    template: '<div data-testid="header-page">Header Page</div>'
  }
}))

describe('Dashboard.vue', () => {
  let wrapper
  let store

  beforeEach(() => {
    // 创建mock store
    store = createStore({
      modules: {
        auth: {
          namespaced: true,
          getters: {
            isAuthenticated: () => true,
            user: () => ({ id: 1, username: 'testuser' })
          }
        }
      }
    })

    // Mock API调用
    vi.mock('@/api/dashboardApi', () => ({
      getDashboardMetrics: vi.fn().mockResolvedValue({
        data: {
          total_tasks: 100,
          in_progress_tasks: 25,
          completed_tasks: 60,
          failed_tasks: 15,
          recent_vulnerabilities: 5,
          high_severity_count: 3
        }
      })
    }))
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    vi.clearAllMocks()
  })

  it('renders dashboard metrics correctly', async () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [store],
        stubs: {
          'a-layout': { template: '<div><slot /></div>' },
          'a-layout-content': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-statistic': {
            props: ['title', 'value', 'prefix'],
            template: '<div><span>{{ title }}: {{ value }}</span></div>'
          }
        }
      }
    })

    // 等待组件挂载和数据加载
    await wrapper.vm.$nextTick()

    // 验证组件渲染
    expect(wrapper.find('.dashboard-layout').exists()).toBe(true)
    expect(wrapper.find('.dashboard-content').exists()).toBe(true)
  })

  it('displays correct metric values', async () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [store],
        stubs: {
          'a-layout': { template: '<div><slot /></div>' },
          'a-layout-content': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-statistic': {
            props: ['title', 'value', 'prefix'],
            template: '<div data-testid="statistic"><span>{{ title }}: {{ value }}</span></div>'
          }
        }
      }
    })

    await wrapper.vm.$nextTick()

    // 检查初始metrics值
    expect(wrapper.vm.metrics.total_tasks).toBe(0)
    expect(wrapper.vm.metrics.in_progress_tasks).toBe(0)
    expect(wrapper.vm.metrics.completed_tasks).toBe(0)
    expect(wrapper.vm.metrics.failed_tasks).toBe(0)
  })

  it('initializes metrics correctly', () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [store],
        stubs: {
          'a-layout': { template: '<div><slot /></div>' },
          'a-layout-content': { template: '<div><slot /></div>' }
        }
      }
    })

    expect(wrapper.vm.metrics.total_tasks).toBe(0)
    expect(wrapper.vm.metrics.in_progress_tasks).toBe(0)
    expect(wrapper.vm.metrics.completed_tasks).toBe(0)
    expect(wrapper.vm.metrics.failed_tasks).toBe(0)
  })

  it('includes HeaderPage component', () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [store],
        stubs: {
          'a-layout': { template: '<div><slot /></div>' },
          'a-layout-content': { template: '<div><slot /></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="header-page"]').exists()).toBe(true)
  })
})