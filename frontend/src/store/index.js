// src/store/index.js

import { createStore } from "vuex";
import api from "../api/axiosInstance"; // 导入 Axios 实例

export default createStore({
  state: {
    isAuthenticated: false,
    user: null,
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
      try {
        // 存储 JWT
        localStorage.setItem("token", token);
        commit("setAuthentication", true);
        return true;
      } catch (error) {
        console.error("Login failed:", error);
        return false;
      }
    },
    async logout({ commit }) {
      try {
        // 清除 JWT
        localStorage.removeItem("token");
        commit("setAuthentication", false);
        commit("setUser", null);
      } catch (error) {
        console.error("Logout failed:", error);
      }
    },
    async checkAuth({ commit }) {
      try {
        const response = await api.get("/auth/check"); // 使用 Axios 实例
        commit("setAuthentication", response.data.authenticated);
        if (response.data.authenticated && response.data.user) {
          commit("setUser", response.data.user);
        }
      } catch (error) {
        // 认证失败时清理垃圾token
        localStorage.removeItem("token");
        commit("setAuthentication", false);
        commit("setUser", null);
      }
    },
  },
  getters: {
    isAuthenticated: (state) => state.isAuthenticated,
    currentUser: (state) => state.user,
  },
});
