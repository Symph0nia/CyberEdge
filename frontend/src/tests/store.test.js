import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createStore } from 'vuex'

// Mock axios
const mockAxios = {
  get: vi.fn(() => Promise.resolve({ data: { authenticated: true, user: { id: 1, username: 'test', role: 'user' } } })),
  defaults: { headers: { common: {} } }
}

vi.mock('@/api/axiosInstance', () => ({ default: mockAxios }))

// Store configuration
const storeConfig = {
  state: {
    isAuthenticated: false,
    user: null,
  },
  mutations: {
    SET_AUTH(state, { isAuthenticated, user = null }) {
      state.isAuthenticated = isAuthenticated;
      state.user = user;
    },
    CLEAR_AUTH(state) {
      state.isAuthenticated = false;
      state.user = null;
    },
  },
  actions: {
    async login({ commit }, { token, user }) {
      localStorage.setItem('token', token);
      mockAxios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      commit('SET_AUTH', { isAuthenticated: true, user });
    },
    async logout({ commit }) {
      localStorage.removeItem('token');
      delete mockAxios.defaults.headers.common['Authorization'];
      commit('CLEAR_AUTH');
    },
    async checkAuth({ commit }) {
      try {
        const response = await mockAxios.get('/auth/check');
        commit('SET_AUTH', {
          isAuthenticated: response.data.authenticated,
          user: response.data.user
        });
      } catch (error) {
        localStorage.removeItem('token');
        commit('CLEAR_AUTH');
      }
    },
  },
  getters: {
    isAuthenticated: (state) => state.isAuthenticated,
    currentUser: (state) => state.user,
    isAdmin: (state) => state.user?.role === 'admin',
  },
}

describe('Vuex Store', () => {
  let store

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    store = createStore(storeConfig)
  })

  describe('Initial State', () => {
    it('has correct initial state', () => {
      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
      expect(store.getters.isAuthenticated).toBe(false)
      expect(store.getters.currentUser).toBe(null)
    })
  })

  describe('Mutations', () => {
    it('SET_AUTH sets authentication state', () => {
      const user = { id: 1, username: 'test', role: 'user' }
      store.commit('SET_AUTH', { isAuthenticated: true, user })

      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toEqual(user)
      expect(store.getters.isAuthenticated).toBe(true)
      expect(store.getters.currentUser).toEqual(user)
    })

    it('CLEAR_AUTH clears authentication state', () => {
      store.commit('SET_AUTH', { isAuthenticated: true, user: { id: 1 } })
      store.commit('CLEAR_AUTH')

      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })
  })

  describe('Actions', () => {
    it('login action sets authentication and stores token', async () => {
      const token = 'test-token'
      const user = { id: 1, username: 'test', role: 'user' }

      await store.dispatch('login', { token, user })

      expect(localStorage.setItem).toHaveBeenCalledWith('token', token)
      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toEqual(user)
    })

    it('logout action clears authentication and removes token', async () => {
      store.commit('SET_AUTH', { isAuthenticated: true, user: { id: 1 } })

      await store.dispatch('logout')

      expect(localStorage.removeItem).toHaveBeenCalledWith('token')
      expect(store.state.isAuthenticated).toBe(false)
      expect(store.state.user).toBe(null)
    })

    it('checkAuth action validates existing token', async () => {
      await store.dispatch('checkAuth')

      expect(mockAxios.get).toHaveBeenCalledWith('/auth/check')
      expect(store.state.isAuthenticated).toBe(true)
      expect(store.state.user).toBeTruthy()
    })
  })

  describe('Getters', () => {
    it('isAdmin getter returns true for admin users', () => {
      store.commit('SET_AUTH', { isAuthenticated: true, user: { role: 'admin' } })
      expect(store.getters.isAdmin).toBe(true)
    })

    it('isAdmin getter returns false for non-admin users', () => {
      store.commit('SET_AUTH', { isAuthenticated: true, user: { role: 'user' } })
      expect(store.getters.isAdmin).toBe(false)
    })
  })
})