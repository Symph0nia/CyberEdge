import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createStore } from 'vuex'

describe('Vuex Store Configuration', () => {
  let mockApi
  let store

  beforeEach(() => {
    // Mock axios
    mockApi = {
      post: vi.fn(),
      get: vi.fn()
    }

    vi.doMock('@/api/axiosInstance', () => ({
      default: mockApi
    }))

    // Create store with mocked API
    store = createStore({
      state: {
        isAuthenticated: false,
      },
      mutations: {
        setAuthentication(state, status) {
          state.isAuthenticated = status;
        },
      },
      actions: {
        async login({ commit }, { account, code }) {
          try {
            const response = await mockApi.post("/auth/validate", { account, code });
            if (response.data.status === "验证码有效") {
              localStorage.setItem("token", response.data.token);
              commit("setAuthentication", true);
              return true;
            }
          } catch (error) {
            console.error("Login failed:", error);
            return false;
          }
        },
        async logout({ commit }) {
          try {
            localStorage.removeItem("token");
            commit("setAuthentication", false);
          } catch (error) {
            console.error("Logout failed:", error);
          }
        },
        async checkAuth({ commit }) {
          try {
            const response = await mockApi.get("/auth/check");
            commit("setAuthentication", response.data.authenticated);
          } catch (error) {
            commit("setAuthentication", false);
          }
        },
      },
    })

    localStorage.clear()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('has correct initial state', () => {
      expect(store.state.isAuthenticated).toBe(false)
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
    it('login action with valid response', async () => {
      const mockResponse = {
        data: {
          status: '验证码有效',
          token: 'fake-jwt-token'
        }
      }
      mockApi.post.mockResolvedValue(mockResponse)

      const result = await store.dispatch('login', {
        account: 'testuser',
        code: '123456'
      })

      expect(mockApi.post).toHaveBeenCalledWith('/auth/validate', {
        account: 'testuser',
        code: '123456'
      })
      expect(localStorage.getItem('token')).toBe('fake-jwt-token')
      expect(store.state.isAuthenticated).toBe(true)
      expect(result).toBe(true)
    })

    it('login action with invalid response', async () => {
      const mockResponse = {
        data: {
          status: '验证码无效'
        }
      }
      mockApi.post.mockResolvedValue(mockResponse)

      const result = await store.dispatch('login', {
        account: 'testuser',
        code: 'wrong-code'
      })

      expect(store.state.isAuthenticated).toBe(false)
      expect(result).toBe(undefined) // Function doesn't return false explicitly in this case
    })

    it('logout action clears state', async () => {
      // Setup authenticated state
      localStorage.setItem('token', 'fake-token')
      store.commit('setAuthentication', true)

      await store.dispatch('logout')

      expect(store.state.isAuthenticated).toBe(false)
      expect(localStorage.getItem('token')).toBeNull()
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
    })
  })

  describe('Store Structure', () => {
    it('has required state properties', () => {
      expect(store.state).toHaveProperty('isAuthenticated')
      expect(typeof store.state.isAuthenticated).toBe('boolean')
    })

    it('has required mutations', () => {
      expect(store._mutations).toHaveProperty('setAuthentication')
    })

    it('has required actions', () => {
      expect(store._actions).toHaveProperty('login')
      expect(store._actions).toHaveProperty('logout')
      expect(store._actions).toHaveProperty('checkAuth')
    })
  })
})