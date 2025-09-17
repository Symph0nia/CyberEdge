import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import LoginPage from '@/components/Login/LoginPage.vue'
import { createStore } from 'vuex'
import { createRouter, createWebHistory } from 'vue-router'

// Mock ant-design-vue components
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
}))

// Mock axios
vi.mock('@/api/axiosInstance', () => ({
  default: {
    post: vi.fn(),
  },
}))

const createMockStore = (state = {}) => {
  return createStore({
    state: {
      isAuthenticated: false,
      user: null,
      ...state,
    },
    mutations: {
      SET_AUTH: (state, { isAuthenticated, user }) => {
        state.isAuthenticated = isAuthenticated
        state.user = user
      },
    },
    actions: {
      login: vi.fn(),
      checkAuth: vi.fn(),
    },
  })
}

const createMockRouter = () => {
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/', component: { template: '<div>Home</div>' } },
      { path: '/user-management', component: { template: '<div>User Management</div>' } },
    ],
  })
}

describe('LoginPage.vue', () => {
  let wrapper
  let store
  let router

  beforeEach(() => {
    store = createMockStore()
    router = createMockRouter()

    wrapper = mount(LoginPage, {
      global: {
        plugins: [store, router],
        stubs: {
          'a-card': { template: '<div class="a-card"><slot /></div>' },
          'a-form': { template: '<form class="a-form"><slot /></form>' },
          'a-form-item': { template: '<div class="a-form-item"><slot /></div>' },
          'a-input': {
            template: '<input class="a-input" @input="$emit(\'update:value\', $event.target.value)" />',
            emits: ['update:value'],
          },
          'a-input-password': {
            template: '<input type="password" class="a-input-password" @input="$emit(\'update:value\', $event.target.value)" />',
            emits: ['update:value'],
          },
          'a-button': {
            template: '<button class="a-button" @click="$emit(\'click\')" :loading="loading"><slot /></button>',
            props: ['loading'],
            emits: ['click'],
          },
          'a-checkbox': {
            template: '<input type="checkbox" class="a-checkbox" @change="$emit(\'update:checked\', $event.target.checked)" />',
            emits: ['update:checked'],
          },
          'a-divider': { template: '<hr class="a-divider" />' },
          'a-space': { template: '<div class="a-space"><slot /></div>' },
          'UserOutlined': { template: '<span class="user-icon"></span>' },
          'LockOutlined': { template: '<span class="lock-icon"></span>' },
        },
      },
    })
  })

  it('renders login form correctly', () => {
    expect(wrapper.find('.a-card').exists()).toBe(true)
    expect(wrapper.find('.a-form').exists()).toBe(true)
    expect(wrapper.find('.a-input').exists()).toBe(true)
    expect(wrapper.find('.a-input-password').exists()).toBe(true)
    expect(wrapper.find('.a-button').exists()).toBe(true)
  })

  it('updates form data when inputs change', async () => {
    const usernameInput = wrapper.find('.a-input')
    const passwordInput = wrapper.find('.a-input-password')

    await usernameInput.setValue('testuser')
    await passwordInput.setValue('password123')

    expect(wrapper.vm.form.username).toBe('testuser')
    expect(wrapper.vm.form.password).toBe('password123')
  })

  it('validates required fields', async () => {
    const form = wrapper.find('.a-form')

    // Try to submit with empty fields
    await form.trigger('submit')

    // Should not proceed with login if validation fails
    expect(store.dispatch).not.toHaveBeenCalledWith('login')
  })

  it('toggles between login and register modes', async () => {
    // Should start in login mode
    expect(wrapper.vm.isRegisterMode).toBe(false)

    // Find and click register link
    const registerButton = wrapper.findAll('.a-button').find(btn =>
      btn.text().includes('注册')
    )

    if (registerButton) {
      await registerButton.trigger('click')
      expect(wrapper.vm.isRegisterMode).toBe(true)
    }
  })

  it('shows register form when in register mode', async () => {
    wrapper.vm.isRegisterMode = true
    await wrapper.vm.$nextTick()

    // Should show email field in register mode
    const inputs = wrapper.findAll('.a-input')
    expect(inputs.length).toBeGreaterThan(1) // username and email
  })

  it('handles form submission', async () => {
    // Mock successful login
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockResolvedValue({
      data: {
        success: true,
        token: 'fake-token',
        message: '登录成功',
      },
    })

    wrapper.vm.form.username = 'testuser'
    wrapper.vm.form.password = 'password123'

    const form = wrapper.find('.a-form')
    await form.trigger('submit')

    // Should call axios post
    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/login', {
      username: 'testuser',
      password: 'password123',
    })
  })

  it('handles login error', async () => {
    // Mock failed login
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockRejectedValue({
      response: {
        data: {
          error: '用户名或密码错误',
        },
      },
    })

    const { message } = await import('ant-design-vue')

    wrapper.vm.form.username = 'testuser'
    wrapper.vm.form.password = 'wrongpassword'

    const form = wrapper.find('.a-form')
    await form.trigger('submit')

    // Should show error message
    expect(message.error).toHaveBeenCalledWith('用户名或密码错误')
  })

  it('handles register mode correctly', async () => {
    wrapper.vm.isRegisterMode = true
    wrapper.vm.form.username = 'newuser'
    wrapper.vm.form.email = 'newuser@example.com'
    wrapper.vm.form.password = 'password123'

    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockResolvedValue({
      data: {
        success: true,
        message: '注册成功',
      },
    })

    const form = wrapper.find('.a-form')
    await form.trigger('submit')

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/register', {
      username: 'newuser',
      email: 'newuser@example.com',
      password: 'password123',
    })
  })

  it('disables submit button during loading', async () => {
    wrapper.vm.loading = true
    await wrapper.vm.$nextTick()

    const submitButton = wrapper.findAll('.a-button').find(btn =>
      btn.text().includes('登录')
    )

    if (submitButton) {
      expect(submitButton.props('loading')).toBe(true)
    }
  })
})