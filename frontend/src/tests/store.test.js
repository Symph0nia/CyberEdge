import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createStore } from 'vuex'
import storeConfig from '@/store'

// Mock axios
vi.mock('@/api/axiosInstance', () => ({
  default: {
    get: vi.fn(),
    defaults: {
      headers: {
        common: {},
      },
    },
  },
}))

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}

global.localStorage = localStorageMock

describe('Vuex Store', () => {
  let store

  beforeEach(() => {
    vi.clearAllMocks()
    store = createStore(storeConfig)
  })

  describe('Initial State', () => {
    it('has correct initial state', () => {
      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })
  })

  describe('Mutations', () => {
    it('SET_AUTH sets authentication state', () => {
      const user = { id: 1, username: 'testuser', email: 'test@example.com' }

      store.commit('SET_AUTH', { isAuthenticated: true, user })

      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toEqual(user)
    })

    it('CLEAR_AUTH clears authentication state', () => {
      // First set auth
      store.commit('SET_AUTH', {
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      // Then clear it
      store.commit('CLEAR_AUTH')

      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })
  })

  describe('Actions', () => {
    it('login action sets authentication and stores token', async () => {
      const mockUser = { id: 1, username: 'testuser', email: 'test@example.com' }
      const mockToken = 'fake-jwt-token'

      await store.dispatch('login', { user: mockUser, token: mockToken })

      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toEqual(mockUser)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockToken)
    })

    it('logout action clears authentication and removes token', async () => {
      // First login
      await store.dispatch('login', {
        user: { id: 1, username: 'testuser' },
        token: 'fake-token',
      })

      // Then logout
      await store.dispatch('logout')

      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token')
    })

    it('checkAuth action validates existing token', async () => {
      const mockAxios = await import('@/api/axiosInstance')
      const mockUser = { id: 1, username: 'testuser', email: 'test@example.com' }

      localStorageMock.getItem.mockReturnValue('existing-token')
      mockAxios.default.get.mockResolvedValue({
        data: {
          authenticated: true,
          user: mockUser,
        },
      })

      await store.dispatch('checkAuth')

      expect(mockAxios.default.get).toHaveBeenCalledWith('/auth/check')
      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toEqual(mockUser)
    })

    it('checkAuth action handles invalid token', async () => {
      const mockAxios = await import('@/api/axiosInstance')

      localStorageMock.getItem.mockReturnValue('invalid-token')
      mockAxios.default.get.mockRejectedValue(new Error('Unauthorized'))

      await store.dispatch('checkAuth')

      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token')
    })

    it('checkAuth action handles no token', async () => {
      localStorageMock.getItem.mockReturnValue(null)

      await store.dispatch('checkAuth')

      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })

    it('sets axios authorization header on login', async () => {
      const mockAxios = await import('@/api/axiosInstance')
      const token = 'test-token'

      await store.dispatch('login', {
        user: { id: 1, username: 'testuser' },
        token,
      })

      expect(mockAxios.default.defaults.headers.common['Authorization']).toBe(`Bearer ${token}`)
    })

    it('removes axios authorization header on logout', async () => {
      const mockAxios = await import('@/api/axiosInstance')

      // First login to set header
      await store.dispatch('login', {
        user: { id: 1, username: 'testuser' },
        token: 'test-token',
      })

      // Then logout
      await store.dispatch('logout')

      expect(mockAxios.default.defaults.headers.common['Authorization']).toBeUndefined()
    })
  })

  describe('Getters', () => {
    it('isAdmin getter returns true for admin users', () => {
      store.commit('SET_AUTH', {
        isAuthenticated: true,
        user: { id: 1, username: 'admin', role: 'admin' },
      })

      expect(store.getters.isAdmin).toBe(true)
    })

    it('isAdmin getter returns false for non-admin users', () => {
      store.commit('SET_AUTH', {
        isAuthenticated: true,
        user: { id: 1, username: 'user', role: 'user' },
      })

      expect(store.getters.isAdmin).toBe(false)
    })

    it('isAdmin getter returns false when not authenticated', () => {
      expect(store.getters.isAdmin).toBe(false)
    })

    it('currentUser getter returns user when authenticated', () => {
      const user = { id: 1, username: 'testuser', role: 'user' }
      store.commit('SET_AUTH', { isAuthenticated: true, user })

      expect(store.getters.currentUser).toEqual(user)
    })

    it('currentUser getter returns null when not authenticated', () => {
      expect(store.getters.currentUser).toBe(null)
    })
  })

  describe('Token Management', () => {
    it('initializes with token from localStorage', async () => {
      const mockAxios = await import('@/api/axiosInstance')
      const existingToken = 'existing-token'

      localStorageMock.getItem.mockReturnValue(existingToken)

      // Create new store instance to trigger initialization
      const newStore = createStore(storeConfig)

      expect(mockAxios.default.defaults.headers.common['Authorization']).toBe(`Bearer ${existingToken}`)
    })

    it('handles concurrent checkAuth calls gracefully', async () => {
      const mockAxios = await import('@/api/axiosInstance')

      localStorageMock.getItem.mockReturnValue('token')
      mockAxios.default.get.mockResolvedValue({
        data: { authenticated: true, user: { id: 1, username: 'test' } },
      })

      // Make multiple concurrent checkAuth calls
      const promises = [
        store.dispatch('checkAuth'),
        store.dispatch('checkAuth'),
        store.dispatch('checkAuth'),
      ]

      await Promise.all(promises)

      // Should only make one API call
      expect(mockAxios.default.get).toHaveBeenCalledTimes(1)
    })
  })
})