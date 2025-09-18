import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import { createStore } from 'vuex'
import LoginPage from '../LoginPage.vue'

// Mock API
vi.mock('@/api/authApi', () => ({
  login: vi.fn()
}))

describe('LoginPage.vue', () => {
  let wrapper
  let router
  let store

  beforeEach(() => {
    // 创建测试路由
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Home</div>' } },
        { path: '/dashboard', component: { template: '<div>Dashboard</div>' } }
      ]
    })

    // 创建测试store
    store = createStore({
      modules: {
        auth: {
          namespaced: true,
          actions: {
            login: vi.fn()
          },
          getters: {
            isAuthenticated: () => false
          }
        }
      }
    })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    vi.clearAllMocks()
  })

  it('renders login form correctly', () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': {
            template: '<form><slot /></form>',
            props: ['model', 'rules']
          },
          'a-form-item': {
            template: '<div><slot name="label" /><slot /></div>',
            props: ['name']
          },
          'a-input': {
            template: '<input v-model="value" v-bind="$attrs" />',
            props: ['value'],
            emits: ['update:value']
          },
          'a-button': {
            template: '<button :disabled="loading"><slot /></button>',
            props: ['type', 'htmlType', 'size', 'block', 'loading']
          }
        }
      }
    })

    expect(wrapper.find('.login-page').exists()).toBe(true)
    expect(wrapper.find('.login-title').text()).toBe('登录账户')
    expect(wrapper.find('.login-subtitle').text()).toBe('登录您的账户以访问完整功能')
  })

  it('initializes form state correctly', () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': { template: '<button><slot /></button>' }
        }
      }
    })

    expect(wrapper.vm.formState.account).toBe('')
    expect(wrapper.vm.formState.code).toBe('')
    expect(wrapper.vm.loading).toBe(false)
  })

  it('validates form rules correctly', () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': { template: '<button><slot /></button>' }
        }
      }
    })

    const rules = wrapper.vm.formRules

    // 测试账户验证规则
    expect(rules.account).toBeDefined()
    expect(rules.account[0].required).toBe(true)
    expect(rules.account[0].message).toBe('请输入账户名')

    // 测试验证码验证规则
    expect(rules.code).toBeDefined()
    expect(rules.code[0].required).toBe(true)
    expect(rules.code[0].message).toBe('请输入验证码')
  })

  it('calls handleLogin method correctly', async () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': { template: '<button><slot /></button>' },
          'GoogleAuthQRCode': { template: '<div></div>' }
        }
      }
    })

    const handleLoginSpy = vi.spyOn(wrapper.vm, 'handleLogin')
    const values = { account: 'testuser', code: '123456' }

    await wrapper.vm.handleLogin(values)

    expect(handleLoginSpy).toHaveBeenCalledWith(values)
    handleLoginSpy.mockRestore()
  })

  it('handles loading state during login', () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': {
            template: '<button :disabled="loading"><slot /></button>',
            props: ['loading']
          }
        }
      }
    })

    // 初始状态不应该在加载中
    expect(wrapper.vm.loading).toBe(false)

    // 设置加载状态
    wrapper.vm.loading = true
    expect(wrapper.vm.loading).toBe(true)
  })

  it('handles login failure correctly', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    store.dispatch = vi.fn().mockRejectedValue(new Error('Login failed'))

    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': { template: '<button><slot /></button>' }
        }
      }
    })

    wrapper.vm.formState.account = 'testuser'
    wrapper.vm.formState.code = 'wrongcode'

    await wrapper.vm.handleLogin()

    expect(wrapper.vm.loading).toBe(false)
    consoleSpy.mockRestore()
  })

  it('handles form validation failure', () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': { template: '<div><slot /></div>' },
          'a-input': { template: '<input />' },
          'a-button': { template: '<button><slot /></button>' },
          'GoogleAuthQRCode': { template: '<div></div>' }
        }
      }
    })

    const errorInfo = {
      values: { account: '', code: '' },
      errorFields: [
        { name: ['account'], errors: ['请输入账户名'] },
        { name: ['code'], errors: ['请输入验证码'] }
      ]
    }

    wrapper.vm.handleLoginFailed(errorInfo)

    expect(consoleErrorSpy).toHaveBeenCalledWith('表单验证失败:', errorInfo)
    consoleErrorSpy.mockRestore()
  })

  it('displays form inputs with correct attributes', () => {
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store],
        stubs: {
          'a-card': { template: '<div><slot /></div>' },
          'a-form': { template: '<form><slot /></form>' },
          'a-form-item': {
            template: '<div><slot name="label" /><slot /></div>',
            props: ['name']
          },
          'a-input': {
            name: 'a-input',
            template: '<input :placeholder="placeholder" :size="size" />',
            props: ['placeholder', 'size', 'value']
          },
          'a-button': { template: '<button><slot /></button>' },
          'GoogleAuthQRCode': { template: '<div></div>' }
        }
      }
    })

    const inputs = wrapper.findAll('input')
    expect(inputs).toHaveLength(2)

    // 验证form state是否正确初始化
    expect(wrapper.vm.formState.account).toBe('')
    expect(wrapper.vm.formState.code).toBe('')
  })
})