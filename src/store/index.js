// src/store/index.js

import { createStore } from "vuex";
import api from "../api/axiosInstance"; // 导入 Axios 实例

export default createStore({
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
        const response = await api.post("/auth/validate", { account, code });
        if (response.data.status === "验证码有效") {
          // 存储 JWT
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
        // 清除 JWT
        localStorage.removeItem("token");
        commit("setAuthentication", false);
      } catch (error) {
        console.error("Logout failed:", error);
      }
    },
    async checkAuth({ commit }) {
      try {
        const response = await api.get("/auth/check"); // 使用 Axios 实例
        commit("setAuthentication", response.data.authenticated);
      } catch (error) {
        commit("setAuthentication", false);
      }
    },
  },
});
