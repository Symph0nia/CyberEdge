import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createStore } from 'vuex'
import storeConfig from '@/store/index.js'

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('Vuex Store', () => {
  let store

  beforeEach(() => {
    vi.clearAllMocks()
    store = createStore(storeConfig)
  })

  describe('User Management', () => {
    it('should have initial user state', () => {
      expect(store.state.user).toEqual({
        isAuthenticated: false,
        userInfo: null,
        token: null,
        permissions: []
      })
    })

    it('should login user successfully', () => {
      const userInfo = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'admin'
      }
      const token = 'fake-jwt-token'

      store.commit('LOGIN_SUCCESS', { user: userInfo, token })

      expect(store.state.user.isAuthenticated).toBe(true)
      expect(store.state.user.userInfo).toEqual(userInfo)
      expect(store.state.user.token).toBe(token)
      expect(localStorage.setItem).toHaveBeenCalledWith('token', token)
      expect(localStorage.setItem).toHaveBeenCalledWith('user', JSON.stringify(userInfo))
    })

    it('should logout user successfully', () => {
      // First login
      store.commit('LOGIN_SUCCESS', {
        user: { id: 1, username: 'testuser' },
        token: 'fake-token'
      })

      // Then logout
      store.commit('LOGOUT')

      expect(store.state.user.isAuthenticated).toBe(false)
      expect(store.state.user.userInfo).toBeNull()
      expect(store.state.user.token).toBeNull()
      expect(localStorage.removeItem).toHaveBeenCalledWith('token')
      expect(localStorage.removeItem).toHaveBeenCalledWith('user')
    })

    it('should update user profile', () => {
      // First login
      store.commit('LOGIN_SUCCESS', {
        user: { id: 1, username: 'testuser', email: 'old@example.com' },
        token: 'token'
      })

      // Update profile
      const updatedInfo = {
        id: 1,
        username: 'testuser',
        email: 'new@example.com',
        profile: { bio: 'Updated bio' }
      }

      store.commit('UPDATE_USER_PROFILE', updatedInfo)

      expect(store.state.user.userInfo).toEqual(updatedInfo)
      expect(localStorage.setItem).toHaveBeenCalledWith('user', JSON.stringify(updatedInfo))
    })

    it('should restore user from localStorage', () => {
      const storedUser = { id: 1, username: 'testuser' }
      const storedToken = 'stored-token'

      localStorage.getItem.mockImplementation((key) => {
        if (key === 'user') return JSON.stringify(storedUser)
        if (key === 'token') return storedToken
        return null
      })

      store.commit('RESTORE_AUTH')

      expect(store.state.user.isAuthenticated).toBe(true)
      expect(store.state.user.userInfo).toEqual(storedUser)
      expect(store.state.user.token).toBe(storedToken)
    })

    it('should handle invalid stored data gracefully', () => {
      localStorage.getItem.mockImplementation((key) => {
        if (key === 'user') return 'invalid-json'
        if (key === 'token') return 'token'
        return null
      })

      store.commit('RESTORE_AUTH')

      expect(store.state.user.isAuthenticated).toBe(false)
      expect(store.state.user.userInfo).toBeNull()
    })
  })

  describe('Permissions', () => {
    it('should set user permissions', () => {
      const permissions = ['create_scan', 'delete_scan', 'view_results']

      store.commit('SET_PERMISSIONS', permissions)

      expect(store.state.user.permissions).toEqual(permissions)
    })

    it('should check if user has permission', () => {
      store.commit('SET_PERMISSIONS', ['create_scan', 'view_results'])

      expect(store.getters.hasPermission('create_scan')).toBe(true)
      expect(store.getters.hasPermission('delete_scan')).toBe(false)
    })

    it('should check if user has admin role', () => {
      // Test admin user
      store.commit('LOGIN_SUCCESS', {
        user: { id: 1, username: 'admin', role: 'admin' },
        token: 'token'
      })
      expect(store.getters.isAdmin).toBe(true)

      // Test regular user
      store.commit('LOGIN_SUCCESS', {
        user: { id: 2, username: 'user', role: 'user' },
        token: 'token'
      })
      expect(store.getters.isAdmin).toBe(false)
    })
  })

  describe('Scan Management', () => {
    it('should have initial scan state', () => {
      expect(store.state.scans).toEqual({
        currentScan: null,
        scanHistory: [],
        runningScans: []
      })
    })

    it('should set current scan', () => {
      const scan = {
        id: 1,
        name: 'Test Scan',
        target: 'example.com',
        status: 'running'
      }

      store.commit('SET_CURRENT_SCAN', scan)

      expect(store.state.scans.currentScan).toEqual(scan)
    })

    it('should add scan to history', () => {
      const scan = {
        id: 1,
        name: 'Test Scan',
        status: 'completed'
      }

      store.commit('ADD_SCAN_TO_HISTORY', scan)

      expect(store.state.scans.scanHistory).toContain(scan)
    })

    it('should update scan status', () => {
      const scan = { id: 1, name: 'Test', status: 'running' }
      store.commit('ADD_SCAN_TO_HISTORY', scan)

      store.commit('UPDATE_SCAN_STATUS', { id: 1, status: 'completed' })

      const updatedScan = store.state.scans.scanHistory.find(s => s.id === 1)
      expect(updatedScan.status).toBe('completed')
    })

    it('should track running scans', () => {
      const scan1 = { id: 1, status: 'running' }
      const scan2 = { id: 2, status: 'running' }

      store.commit('ADD_RUNNING_SCAN', scan1)
      store.commit('ADD_RUNNING_SCAN', scan2)

      expect(store.state.scans.runningScans).toHaveLength(2)

      // Complete one scan
      store.commit('UPDATE_SCAN_STATUS', { id: 1, status: 'completed' })
      store.commit('REMOVE_RUNNING_SCAN', 1)

      expect(store.state.scans.runningScans).toHaveLength(1)
    })
  })

  describe('UI State', () => {
    it('should have initial UI state', () => {
      expect(store.state.ui).toEqual({
        loading: false,
        sidebarCollapsed: false,
        theme: 'light',
        notifications: []
      })
    })

    it('should toggle loading state', () => {
      store.commit('SET_LOADING', true)
      expect(store.state.ui.loading).toBe(true)

      store.commit('SET_LOADING', false)
      expect(store.state.ui.loading).toBe(false)
    })

    it('should toggle sidebar', () => {
      store.commit('TOGGLE_SIDEBAR')
      expect(store.state.ui.sidebarCollapsed).toBe(true)

      store.commit('TOGGLE_SIDEBAR')
      expect(store.state.ui.sidebarCollapsed).toBe(false)
    })

    it('should change theme', () => {
      store.commit('SET_THEME', 'dark')
      expect(store.state.ui.theme).toBe('dark')
      expect(localStorage.setItem).toHaveBeenCalledWith('theme', 'dark')
    })

    it('should manage notifications', () => {
      const notification1 = { id: 1, type: 'success', message: 'Success!' }
      const notification2 = { id: 2, type: 'error', message: 'Error!' }

      store.commit('ADD_NOTIFICATION', notification1)
      store.commit('ADD_NOTIFICATION', notification2)

      expect(store.state.ui.notifications).toHaveLength(2)

      store.commit('REMOVE_NOTIFICATION', 1)

      expect(store.state.ui.notifications).toHaveLength(1)
      expect(store.state.ui.notifications[0].id).toBe(2)
    })
  })

  describe('Actions', () => {
    it('should dispatch login action', async () => {
      const mockLogin = vi.fn().mockResolvedValue({
        data: {
          success: true,
          user: { id: 1, username: 'testuser' },
          token: 'token'
        }
      })

      // Mock the API call
      store._actions.login = [mockLogin]

      const credentials = { username: 'testuser', password: 'password' }
      await store.dispatch('login', credentials)

      expect(mockLogin).toHaveBeenCalledWith(expect.any(Object), credentials)
    })

    it('should dispatch logout action', async () => {
      // Login first
      store.commit('LOGIN_SUCCESS', {
        user: { id: 1, username: 'testuser' },
        token: 'token'
      })

      await store.dispatch('logout')

      expect(store.state.user.isAuthenticated).toBe(false)
    })

    it('should fetch user profile', async () => {
      const mockFetchProfile = vi.fn().mockResolvedValue({
        data: { id: 1, username: 'testuser', email: 'test@example.com' }
      })

      store._actions.fetchUserProfile = [mockFetchProfile]

      await store.dispatch('fetchUserProfile')

      expect(mockFetchProfile).toHaveBeenCalled()
    })
  })

  describe('Getters', () => {
    it('should get authenticated user info', () => {
      const userInfo = { id: 1, username: 'testuser' }
      store.commit('LOGIN_SUCCESS', { user: userInfo, token: 'token' })

      expect(store.getters.currentUser).toEqual(userInfo)
    })

    it('should check authentication status', () => {
      expect(store.getters.isAuthenticated).toBe(false)

      store.commit('LOGIN_SUCCESS', {
        user: { id: 1, username: 'testuser' },
        token: 'token'
      })

      expect(store.getters.isAuthenticated).toBe(true)
    })

    it('should get running scans count', () => {
      store.commit('ADD_RUNNING_SCAN', { id: 1, status: 'running' })
      store.commit('ADD_RUNNING_SCAN', { id: 2, status: 'running' })

      expect(store.getters.runningScansCount).toBe(2)
    })

    it('should get recent scans', () => {
      const scans = [
        { id: 1, created_at: '2024-01-03' },
        { id: 2, created_at: '2024-01-02' },
        { id: 3, created_at: '2024-01-01' }
      ]

      scans.forEach(scan => {
        store.commit('ADD_SCAN_TO_HISTORY', scan)
      })

      const recentScans = store.getters.recentScans(2)
      expect(recentScans).toHaveLength(2)
      expect(recentScans[0].id).toBe(1) // Most recent first
    })
  })

  describe('Error Handling', () => {
    it('should handle login errors', async () => {
      const error = new Error('Login failed')
      const mockLogin = vi.fn().mockRejectedValue(error)

      store._actions.login = [mockLogin]

      try {
        await store.dispatch('login', { username: 'test', password: 'wrong' })
      } catch (e) {
        expect(e).toBe(error)
      }

      expect(store.state.user.isAuthenticated).toBe(false)
    })

    it('should handle network errors gracefully', () => {
      // Simulate network error during auth restore
      localStorage.getItem.mockImplementation(() => {
        throw new Error('Network error')
      })

      store.commit('RESTORE_AUTH')

      // Should not crash and should maintain clean state
      expect(store.state.user.isAuthenticated).toBe(false)
      expect(store.state.user.userInfo).toBeNull()
    })
  })
})