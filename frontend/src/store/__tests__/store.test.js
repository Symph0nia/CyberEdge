import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
})

// Mock axios at the top level
const mockApi = {
  get: vi.fn()
}

vi.mock('@/api/axiosInstance', () => ({
  default: mockApi
}))

describe('Vuex Store Configuration', () => {
  let store

  beforeEach(async () => {
    // Reset localStorage mock
    localStorageMock.getItem.mockReturnValue(null)
    localStorageMock.setItem.mockImplementation(() => {})
    localStorageMock.removeItem.mockImplementation(() => {})

    // Import the actual store module
    const storeModule = await import('@/store')
    store = storeModule.default

    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('has correct initial state', () => {
      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })
  })

  describe('Mutations', () => {
    it('setAuthentication updates isAuthenticated state', () => {
      expect(store.state.isAuthenticated).toBe(false)

      store.commit('setAuthentication', true)
      expect(store.state.isAuthenticated).toBe(true)

      store.commit('setAuthentication', false)
      expect(store.state.isAuthenticated).toBe(false)
    })
  })

  describe('Actions', () => {
    it('login action stores token and sets authentication', async () => {
      const result = await store.dispatch('login', { token: 'test-token' })

      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', 'test-token')
      expect(store.state.isAuthenticated).toBe(true)
      expect(result).toBe(true)
    })

    it('logout action clears state', async () => {
      // Setup authenticated state
      store.commit('setAuthentication', true)

      await store.dispatch('logout')

      expect(store.state.isAuthenticated).toBe(false)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token')
    })

    it('checkAuth action with authenticated response', async () => {
      const mockResponse = {
        data: {
          authenticated: true
        }
      }
      mockApi.get.mockResolvedValue(mockResponse)

      await store.dispatch('checkAuth')

      expect(mockApi.get).toHaveBeenCalledWith('/auth/check')
      expect(store.state.isAuthenticated).toBe(true)
    })

    it('checkAuth action with error response', async () => {
      mockApi.get.mockRejectedValue(new Error('Network error'))

      await store.dispatch('checkAuth')

      expect(store.state.isAuthenticated).toBe(false)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token')
    })
  })

  describe('Store Structure', () => {
    it('has required state properties', () => {
      expect(store.state).toHaveProperty('isAuthenticated')
      expect(store.state).toHaveProperty('user')
      expect(typeof store.state.isAuthenticated).toBe('boolean')
    })

    it('has required mutations', () => {
      expect(store._mutations).toHaveProperty('setAuthentication')
      expect(store._mutations).toHaveProperty('setUser')
      expect(store._mutations).toHaveProperty('SET_AUTH')
      expect(store._mutations).toHaveProperty('CLEAR_AUTH')
    })

    it('has required actions', () => {
      expect(store._actions).toHaveProperty('login')
      expect(store._actions).toHaveProperty('logout')
      expect(store._actions).toHaveProperty('checkAuth')
    })
  })
})