import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ProfilePage from '@/components/Profile/ProfilePage.vue'
import { createStore } from 'vuex'

// Mock ant-design-vue components
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

// Mock axios
vi.mock('@/api/axiosInstance', () => ({
  default: {
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

const createMockStore = (user = {}) => {
  return createStore({
    state: {
      isAuthenticated: true,
      user: {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'user',
        is_2fa_enabled: false,
        created_at: 1234567890,
        updated_at: 1234567890,
        ...user,
      },
    },
    actions: {
      checkAuth: vi.fn(),
      logout: vi.fn(),
    },
  })
}

describe('ProfilePage.vue', () => {
  let wrapper
  let store

  beforeEach(() => {
    store = createMockStore()

    wrapper = mount(ProfilePage, {
      global: {
        plugins: [store],
        stubs: {
          'a-row': { template: '<div class="a-row"><slot /></div>' },
          'a-col': { template: '<div class="a-col"><slot /></div>' },
          'a-card': { template: '<div class="a-card"><slot /></div>' },
          'a-form': { template: '<form class="a-form" @submit.prevent="$emit(\'finish\')"><slot /></form>', emits: ['finish'] },
          'a-form-item': { template: '<div class="a-form-item"><slot /></div>' },
          'a-input': {
            template: '<input class="a-input" @input="$emit(\'update:value\', $event.target.value)" />',
            props: ['disabled'],
            emits: ['update:value'],
          },
          'a-input-password': {
            template: '<input type="password" class="a-input-password" @input="$emit(\'update:value\', $event.target.value)" />',
            emits: ['update:value'],
          },
          'a-button': {
            template: '<button class="a-button" @click="$emit(\'click\')" :loading="loading"><slot /></button>',
            props: ['type', 'loading', 'danger'],
            emits: ['click'],
          },
          'a-tag': {
            template: '<span class="a-tag" :style="{ color }"><slot /></span>',
            props: ['color'],
          },
          'a-avatar': { template: '<div class="a-avatar"><slot /></div>' },
          'a-switch': {
            template: '<input type="checkbox" class="a-switch" :checked="checked" @change="$emit(\'change\', $event.target.checked)" />',
            props: ['checked', 'loading'],
            emits: ['change'],
          },
          'a-modal': {
            template: '<div class="a-modal" v-if="visible"><slot /></div>',
            props: ['visible', 'title'],
          },
          'a-alert': {
            template: '<div class="a-alert"><slot /></div>',
            props: ['message', 'description', 'type'],
          },
          'UserOutlined': { template: '<span class="user-icon"></span>' },
        },
      },
    })
  })

  it('renders profile page correctly', () => {
    expect(wrapper.find('.a-card').exists()).toBe(true)
    expect(wrapper.find('.a-form').exists()).toBe(true)
    expect(wrapper.find('.a-avatar').exists()).toBe(true)
  })

  it('displays user information correctly', async () => {
    await wrapper.vm.$nextTick()

    expect(wrapper.vm.userForm.username).toBe('testuser')
    expect(wrapper.vm.userForm.email).toBe('test@example.com')
    expect(wrapper.vm.userForm.role).toBe('user')
  })

  it('displays correct 2FA status', async () => {
    expect(wrapper.vm.is2FAEnabled).toBe(false)

    // Update store with 2FA enabled user
    store.state.user.is_2fa_enabled = true
    wrapper.vm.initUserInfo()
    await wrapper.vm.$nextTick()

    expect(wrapper.vm.is2FAEnabled).toBe(true)
  })

  it('handles password change', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.post.mockResolvedValue({
      data: { success: true, message: '密码修改成功' },
    })

    wrapper.vm.passwordForm.currentPassword = 'oldPassword'
    wrapper.vm.passwordForm.newPassword = 'NewPassword123'
    wrapper.vm.passwordForm.confirmPassword = 'NewPassword123'

    await wrapper.vm.handlePasswordChange({
      currentPassword: 'oldPassword',
      newPassword: 'NewPassword123',
      confirmPassword: 'NewPassword123',
    })

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/change-password', {
      currentPassword: 'oldPassword',
      newPassword: 'NewPassword123',
    })
    expect(message.success).toHaveBeenCalledWith('密码修改成功，请重新登录')
  })

  it('validates confirm password', async () => {
    wrapper.vm.passwordForm.newPassword = 'NewPassword123'

    // Test matching passwords
    const result1 = await wrapper.vm.validateConfirmPassword({}, 'NewPassword123')
    expect(result1).resolves

    // Test non-matching passwords
    try {
      await wrapper.vm.validateConfirmPassword({}, 'DifferentPassword')
    } catch (error) {
      expect(error).toBe('两次输入的密码不一致')
    }
  })

  it('handles 2FA setup', async () => {
    const mockAxios = await import('@/api/axiosInstance')

    mockAxios.default.post.mockResolvedValue({
      data: {
        secret: 'FAKE_SECRET',
        qr_code: '<svg>fake qr code</svg>',
      },
    })

    await wrapper.vm.setup2FA()

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/2fa/setup', {
      username: 'testuser',
    })
    expect(wrapper.vm.qrCodeData).toBe('<svg>fake qr code</svg>')
    expect(wrapper.vm.setup2FAVisible).toBe(true)
  })

  it('handles 2FA verification', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.post.mockResolvedValue({
      data: { success: true },
    })

    wrapper.vm.verifyCode = '123456'
    await wrapper.vm.handleVerify2FA()

    expect(mockAxios.default.post).toHaveBeenCalledWith('/auth/2fa/verify', {
      username: 'testuser',
      code: '123456',
    })
    expect(message.success).toHaveBeenCalledWith('双因子认证设置成功')
    expect(wrapper.vm.setup2FAVisible).toBe(false)
  })

  it('handles 2FA disable', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.delete.mockResolvedValue({
      data: { success: true },
    })

    await wrapper.vm.handleDisable2FA()

    expect(mockAxios.default.delete).toHaveBeenCalledWith('/auth/2fa', {
      data: { username: 'testuser' },
    })
    expect(message.success).toHaveBeenCalledWith('双因子认证已禁用')
    expect(wrapper.vm.is2FAEnabled).toBe(false)
  })

  it('resets password form', () => {
    wrapper.vm.passwordForm.currentPassword = 'test'
    wrapper.vm.passwordForm.newPassword = 'test'
    wrapper.vm.passwordForm.confirmPassword = 'test'

    wrapper.vm.resetPasswordForm()

    expect(wrapper.vm.passwordForm.currentPassword).toBe('')
    expect(wrapper.vm.passwordForm.newPassword).toBe('')
    expect(wrapper.vm.passwordForm.confirmPassword).toBe('')
  })

  it('formats date correctly', () => {
    const timestamp = 1234567890
    const formatted = wrapper.vm.formatDate(timestamp)

    expect(formatted).toMatch(/\d{4}\/\d{1,2}\/\d{1,2}/)
  })

  it('handles 2FA toggle', async () => {
    const mockAxios = await import('@/api/axiosInstance')

    mockAxios.default.post.mockResolvedValue({
      data: {
        secret: 'FAKE_SECRET',
        qr_code: '<svg>fake qr code</svg>',
      },
    })

    // Enable 2FA
    await wrapper.vm.handle2FAToggle(true)

    expect(wrapper.vm.setup2FAVisible).toBe(true)

    // Disable 2FA (should revert)
    await wrapper.vm.handle2FAToggle(false)
    expect(wrapper.vm.is2FAEnabled).toBe(true) // Should be reverted back
  })

  it('handles API errors gracefully', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.post.mockRejectedValue(new Error('Network error'))

    await wrapper.vm.handlePasswordChange({
      currentPassword: 'old',
      newPassword: 'new',
    })

    expect(message.error).toHaveBeenCalledWith('密码修改失败')
  })

  it('handles loading states correctly', async () => {
    wrapper.vm.passwordLoading = true
    await wrapper.vm.$nextTick()

    const buttons = wrapper.findAll('.a-button')
    const passwordButton = buttons.find(btn => btn.text().includes('更新密码'))
    if (passwordButton) {
      expect(passwordButton.props('loading')).toBe(true)
    }
  })

  it('initializes user info on mount', () => {
    expect(wrapper.vm.userForm.username).toBe('testuser')
    expect(wrapper.vm.userForm.email).toBe('test@example.com')
    expect(wrapper.vm.userForm.role).toBe('user')
    expect(wrapper.vm.is2FAEnabled).toBe(false)
  })
})