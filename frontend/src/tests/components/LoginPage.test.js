import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createStore } from 'vuex'
import LoginPage from '@/components/Login/LoginPage.vue'
import { notification } from 'ant-design-vue'
import storeConfig from '@/store/index.js'

// Mock router
const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', name: 'Dashboard' },
    { path: '/login', name: 'Login' }
  ]
})

// Mock notification
vi.mock('ant-design-vue', async () => {
  const actual = await vi.importActual('ant-design-vue')
  return {
    ...actual,
    notification: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn()
    }
  }
})

// Mock axios
vi.mock('@/api/axiosInstance', () => ({
  default: {
    post: vi.fn()
  }
}))

describe('LoginPage.vue', () => {
  let wrapper
  let store

  beforeEach(() => {
    vi.clearAllMocks()
    store = createStore({
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
    wrapper = mount(LoginPage, {
      global: {
        plugins: [router, store]
      }
    })
  })

  it('renders login form correctly', () => {
    expect(wrapper.find('[data-testid="login-form"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="username-input"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="password-input"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="login-button"]').exists()).toBe(true)
  })

  it('shows validation errors for empty fields', async () => {
    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    // Should show validation errors
    expect(wrapper.text()).toContain('请输入用户名')
    expect(wrapper.text()).toContain('请输入密码')
  })

  it('validates username format', async () => {
    await wrapper.find('[data-testid="username-input"]').setValue('a') // Too short
    await wrapper.find('[data-testid="password-input"]').setValue('validpassword123')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    expect(wrapper.text()).toContain('用户名长度必须在3-50个字符之间')
  })

  it('validates password requirements', async () => {
    await wrapper.find('[data-testid="username-input"]').setValue('validuser')
    await wrapper.find('[data-testid="password-input"]').setValue('weak') // Too weak

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    expect(wrapper.text()).toContain('密码必须包含大小写字母、数字和特殊字符')
  })

  it('handles successful login', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockResolvedValueOnce({
      data: {
        success: true,
        token: 'fake-jwt-token',
        user: {
          id: 1,
          username: 'testuser',
          role: 'user'
        }
      }
    })

    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('ValidPassword123!')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/login', {
      username: 'testuser',
      password: 'ValidPassword123!'
    })
    expect(notification.success).toHaveBeenCalledWith({
      message: '登录成功',
      description: '欢迎回来，testuser!'
    })
  })

  it('handles login failure', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockRejectedValueOnce({
      response: {
        data: {
          error: '用户名或密码错误'
        }
      }
    })

    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('WrongPassword123!')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(notification.error).toHaveBeenCalledWith({
      message: '登录失败',
      description: '用户名或密码错误'
    })
  })

  it('handles network error', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockRejectedValueOnce(new Error('Network Error'))

    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('ValidPassword123!')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(notification.error).toHaveBeenCalledWith({
      message: '登录失败',
      description: '网络错误，请稍后重试'
    })
  })

  it('shows loading state during login', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    // Create a promise that we can control
    let resolveLogin
    const loginPromise = new Promise(resolve => {
      resolveLogin = resolve
    })
    mockAxios.default.post.mockReturnValueOnce(loginPromise)

    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('ValidPassword123!')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    // Should show loading state
    expect(wrapper.find('[data-testid="login-button"]').attributes('loading')).toBeDefined()

    // Resolve the login
    resolveLogin({
      data: {
        success: true,
        token: 'fake-token',
        user: { id: 1, username: 'testuser', role: 'user' }
      }
    })
    await wrapper.vm.$nextTick()

    // Loading state should be gone
    expect(wrapper.find('[data-testid="login-button"]').attributes('loading')).toBeUndefined()
  })

  it('toggles password visibility', async () => {
    const passwordInput = wrapper.find('[data-testid="password-input"] input')
    const toggleButton = wrapper.find('[data-testid="password-toggle"]')

    // Initially password should be hidden
    expect(passwordInput.attributes('type')).toBe('password')

    // Click toggle button
    await toggleButton.trigger('click')
    expect(passwordInput.attributes('type')).toBe('text')

    // Click again to hide
    await toggleButton.trigger('click')
    expect(passwordInput.attributes('type')).toBe('password')
  })

  it('handles remember me functionality', async () => {
    const rememberCheckbox = wrapper.find('[data-testid="remember-checkbox"]')

    // Check remember me
    await rememberCheckbox.setChecked(true)
    expect(wrapper.vm.loginForm.remember).toBe(true)

    // Uncheck remember me
    await rememberCheckbox.setChecked(false)
    expect(wrapper.vm.loginForm.remember).toBe(false)
  })

  it('navigates to register page', async () => {
    const registerLink = wrapper.find('[data-testid="register-link"]')
    await registerLink.trigger('click')

    expect(wrapper.vm.$route.path).toBe('/register')
  })

  it('handles 2FA verification flow', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockResolvedValueOnce({
      data: {
        success: false,
        requires_2fa: true,
        temp_token: 'temp-token'
      }
    })

    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('ValidPassword123!')

    const loginButton = wrapper.find('[data-testid="login-button"]')
    await loginButton.trigger('click')

    await wrapper.vm.$nextTick()

    // Should show 2FA input
    expect(wrapper.find('[data-testid="2fa-input"]').exists()).toBe(true)
    expect(wrapper.vm.show2FA).toBe(true)
  })

  it('validates 2FA code format', async () => {
    // Set up 2FA mode
    await wrapper.setData({ show2FA: true, tempToken: 'temp-token' })

    const twoFAInput = wrapper.find('[data-testid="2fa-input"]')
    await twoFAInput.setValue('123') // Too short

    const verifyButton = wrapper.find('[data-testid="2fa-verify-button"]')
    await verifyButton.trigger('click')

    expect(wrapper.text()).toContain('请输入6位数字验证码')
  })

  it('submits 2FA verification', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.post.mockResolvedValueOnce({
      data: {
        success: true,
        token: 'final-token',
        user: { id: 1, username: 'testuser', role: 'user' }
      }
    })

    // Set up 2FA mode
    await wrapper.setData({ show2FA: true, tempToken: 'temp-token' })

    await wrapper.find('[data-testid="2fa-input"]').setValue('123456')

    const verifyButton = wrapper.find('[data-testid="2fa-verify-button"]')
    await verifyButton.trigger('click')

    await wrapper.vm.$nextTick()

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/2fa/verify', {
      temp_token: 'temp-token',
      code: '123456'
    })
    expect(notification.success).toHaveBeenCalled()
  })

  it('handles keyboard shortcuts', async () => {
    await wrapper.find('[data-testid="username-input"]').setValue('testuser')
    await wrapper.find('[data-testid="password-input"]').setValue('ValidPassword123!')

    const passwordInput = wrapper.find('[data-testid="password-input"] input')

    // Press Enter key
    await passwordInput.trigger('keyup.enter')

    // Should trigger login
    const mockAxios = await import('@/api/axiosInstance')
    expect(mockAxios.default.post).toHaveBeenCalled()
  })
})