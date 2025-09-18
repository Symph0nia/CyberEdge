// src/api/axiosInstance.js

import axios from "axios";

// 动态确定 API 基础 URL
const getApiBaseURL = () => {
  // 如果有环境变量，优先使用
  if (process.env.VUE_APP_API_BASE_URL) {
    return process.env.VUE_APP_API_BASE_URL;
  }

  // 否则根据当前访问的域名自动确定
  const hostname = window.location.hostname;
  const protocol = window.location.protocol;

  // 如果是 localhost 或 127.0.0.1，使用本地地址
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return `${protocol}//${hostname}:31337`;
  }

  // 否则使用当前域名的 31337 端口
  return `${protocol}//${hostname}:31337`;
};

// 创建一个 axios 实例
const api = axios.create({
  baseURL: getApiBaseURL(),
  withCredentials: true, // 允许跨域请求发送 cookies
});

// 添加请求拦截器，将 JWT 添加到请求头中
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default api;
