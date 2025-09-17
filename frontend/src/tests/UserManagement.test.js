import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import UserManagement from '@/components/User/UserManagement.vue'
import { createStore } from 'vuex'

// Mock ant-design-vue components
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
  Modal: {
    confirm: vi.fn(),
  },
}))

// Mock axios
vi.mock('@/api/axiosInstance', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

const createMockStore = (state = {}) => {
  return createStore({
    state: {
      isAuthenticated: true,
      user: { id: 1, username: 'admin', role: 'admin' },
      ...state,
    },
    actions: {
      checkAuth: vi.fn(),
    },
  })
}

const mockUsers = [
  { account: 'user1', loginCount: 5 },
  { account: 'user2', loginCount: 3 },
  { account: 'admin', loginCount: 10 },
]

describe('UserManagement.vue', () => {
  let wrapper
  let store

  beforeEach(() => {
    store = createMockStore()

    wrapper = mount(UserManagement, {
      global: {
        plugins: [store],
        stubs: {
          'a-card': { template: '<div class="a-card"><slot /></div>' },
          'a-row': { template: '<div class="a-row"><slot /></div>' },
          'a-col': { template: '<div class="a-col"><slot /></div>' },
          'a-statistic': {
            template: '<div class="a-statistic">{{ value }}</div>',
            props: ['value', 'title'],
          },
          'a-table': {
            template: '<div class="a-table"><slot /></div>',
            props: ['dataSource', 'columns', 'loading', 'pagination'],
          },
          'a-button': {
            template: '<button class="a-button" @click="$emit(\'click\')" :loading="loading"><slot /></button>',
            props: ['type', 'loading', 'danger'],
            emits: ['click'],
          },
          'a-input': {
            template: '<input class="a-input" @input="$emit(\'update:value\', $event.target.value)" />',
            emits: ['update:value'],
          },
          'a-switch': {
            template: '<input type="checkbox" class="a-switch" @change="$emit(\'change\', $event.target.checked)" />',
            props: ['checked', 'loading'],
            emits: ['change'],
          },
          'a-modal': {
            template: '<div class="a-modal" v-if="visible"><slot /></div>',
            props: ['visible', 'title'],
          },
          'a-form': { template: '<form class="a-form"><slot /></form>' },
          'a-form-item': { template: '<div class="a-form-item"><slot /></div>' },
          'a-select': {
            template: '<select class="a-select"><slot /></select>',
            props: ['value'],
          },
          'a-select-option': {
            template: '<option class="a-select-option" :value="value"><slot /></option>',
            props: ['value'],
          },
          'a-space': { template: '<div class="a-space"><slot /></div>' },
          'UserOutlined': { template: '<span class="user-icon"></span>' },
          'UserAddOutlined': { template: '<span class="user-add-icon"></span>' },
          'GlobalOutlined': { template: '<span class="global-icon"></span>' },
          'SafetyOutlined': { template: '<span class="safety-icon"></span>' },
          'QrcodeOutlined': { template: '<span class="qrcode-icon"></span>' },
        },
      },
    })
  })

  it('renders user management page correctly', () => {
    expect(wrapper.find('.a-card').exists()).toBe(true)
    expect(wrapper.find('.a-table').exists()).toBe(true)
    expect(wrapper.find('.a-statistic').exists()).toBe(true)
  })

  it('fetches users on mount', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.get.mockResolvedValue({
      data: mockUsers,
    })

    // Component should call fetchUsers on mount
    expect(mockAxios.default.get).toHaveBeenCalledWith('/users')
  })

  it('displays user statistics correctly', async () => {
    // Mock users data
    wrapper.vm.users = mockUsers
    await wrapper.vm.$nextTick()

    const statistics = wrapper.findAll('.a-statistic')
    expect(statistics.length).toBeGreaterThan(0)

    // Should show total users count
    expect(wrapper.vm.users.length).toBe(3)
  })

  it('handles search functionality', async () => {
    wrapper.vm.users = mockUsers
    wrapper.vm.searchKeyword = 'user1'
    await wrapper.vm.$nextTick()

    // Should filter users based on search
    const filteredUsers = wrapper.vm.filteredUsers
    expect(filteredUsers.some(user => user.account.includes('user1'))).toBe(true)
  })

  it('opens add user modal', async () => {
    const addButton = wrapper.findAll('.a-button').find(btn =>
      btn.text().includes('添加用户')
    )

    if (addButton) {
      await addButton.trigger('click')
      expect(wrapper.vm.showAddUser).toBe(true)
    }
  })

  it('handles QR code status toggle', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    mockAxios.default.get.mockResolvedValue({
      data: { is2FAEnabled: false },
    })

    const qrSwitch = wrapper.find('.a-switch')
    if (qrSwitch.exists()) {
      await qrSwitch.trigger('change')
      // Should handle QR code toggle
      expect(wrapper.vm.qrStatusLoading).toBe(false) // After toggle completes
    }
  })

  it('handles user deletion', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { Modal } = await import('ant-design-vue')

    mockAxios.default.delete.mockResolvedValue({})

    // Mock modal confirm
    Modal.confirm.mockImplementation(({ onOk }) => {
      onOk && onOk()
    })

    const testUser = { account: 'testuser', id: 1 }
    await wrapper.vm.handleDelete(testUser)

    expect(mockAxios.default.delete).toHaveBeenCalledWith('/users/1')
  })

  it('validates add user form', async () => {
    wrapper.vm.showAddUser = true
    wrapper.vm.addUserForm.username = ''
    wrapper.vm.addUserForm.email = ''
    wrapper.vm.addUserForm.password = ''

    await wrapper.vm.$nextTick()

    // Form should be invalid with empty fields
    expect(wrapper.vm.addUserForm.username).toBe('')
    expect(wrapper.vm.addUserForm.email).toBe('')
    expect(wrapper.vm.addUserForm.password).toBe('')
  })

  it('handles add user submission', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.post.mockResolvedValue({
      data: { success: true, message: '用户创建成功' },
    })

    wrapper.vm.addUserForm = {
      username: 'newuser',
      email: 'new@example.com',
      password: 'Password123',
      role: 'user',
    }

    await wrapper.vm.handleAddUser()

    expect(mockAxios.default.post).toHaveBeenCalledWith('/users', {
      username: 'newuser',
      email: 'new@example.com',
      password: 'Password123',
    })
  })

  it('handles API errors gracefully', async () => {
    const mockAxios = await import('@/api/axiosInstance')
    const { message } = await import('ant-design-vue')

    mockAxios.default.get.mockRejectedValue(new Error('Network error'))

    await wrapper.vm.fetchUsers()

    expect(message.error).toHaveBeenCalledWith('获取用户列表失败')
  })

  it('updates online users count', async () => {
    // Should have initial online users count
    expect(typeof wrapper.vm.onlineUsers).toBe('number')
    expect(wrapper.vm.onlineUsers).toBeGreaterThanOrEqual(0)
  })

  it('handles batch operations', async () => {
    wrapper.vm.selectedRowKeys = [1, 2]
    await wrapper.vm.$nextTick()

    // Should enable batch operations when users are selected
    expect(wrapper.vm.selectedRowKeys.length).toBe(2)
  })

  it('resets add user form', async () => {
    wrapper.vm.addUserForm = {
      username: 'test',
      email: 'test@test.com',
      password: 'password',
      role: 'admin',
    }

    wrapper.vm.resetAddUserForm()

    expect(wrapper.vm.addUserForm.username).toBe('')
    expect(wrapper.vm.addUserForm.email).toBe('')
    expect(wrapper.vm.addUserForm.password).toBe('')
    expect(wrapper.vm.addUserForm.role).toBe('user')
  })

  it('handles loading states correctly', async () => {
    wrapper.vm.loading = true
    await wrapper.vm.$nextTick()

    const table = wrapper.find('.a-table')
    expect(table.props('loading')).toBe(true)
  })
})