import { createApp } from "vue";
import App from "./App.vue";
import "./main.css";
import "remixicon/fonts/remixicon.css";
import router from "./router";
import store from "./store";

// Ant Design Vue
import Antd, { ConfigProvider } from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';

// Ant Design主题配置
const themeConfig = {
  token: {
    // 主色调
    colorPrimary: '#3b82f6',
    colorSuccess: '#10b981',
    colorWarning: '#f59e0b',
    colorError: '#ef4444',
    colorInfo: '#06b6d4',

    // 背景色
    colorBgBase: '#0f172a',
    colorBgContainer: 'rgba(30, 41, 59, 0.8)',
    colorBgElevated: 'rgba(51, 65, 85, 0.8)',

    // 文字色
    colorText: '#e2e8f0',
    colorTextSecondary: '#94a3b8',
    colorTextTertiary: '#64748b',
    colorTextQuaternary: '#475569',

    // 边框色
    colorBorder: 'rgba(51, 65, 85, 0.3)',
    colorBorderSecondary: 'rgba(51, 65, 85, 0.2)',

    // 圆角
    borderRadius: 12,
    borderRadiusLG: 16,
    borderRadiusSM: 8,

    // 字体
    fontSize: 14,
    fontSizeLG: 16,
    fontSizeSM: 12,
    fontWeightStrong: 600,

    // 阴影
    boxShadow: '0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
    boxShadowSecondary: '0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
  },
  algorithm: 'darkAlgorithm',
  components: {
    Button: {
      borderRadius: 12,
      controlHeight: 40,
      fontWeight: 500,
    },
    Input: {
      borderRadius: 12,
      controlHeight: 40,
      colorBgContainer: 'rgba(15, 23, 42, 0.6)',
      activeBorderColor: '#3b82f6',
      hoverBorderColor: 'rgba(59, 130, 246, 0.5)',
    },
    Card: {
      borderRadius: 16,
      colorBgContainer: 'rgba(30, 41, 59, 0.8)',
      colorBorderSecondary: 'rgba(51, 65, 85, 0.3)',
    },
    Table: {
      borderRadius: 12,
      colorBgContainer: 'rgba(30, 41, 59, 0.8)',
      headerBg: 'rgba(15, 23, 42, 0.8)',
      headerColor: '#94a3b8',
      colorBorderSecondary: 'rgba(51, 65, 85, 0.2)',
    },
    Menu: {
      colorBgContainer: 'rgba(30, 41, 59, 0.8)',
      colorItemText: '#cbd5e1',
      colorItemTextSelected: '#60a5fa',
      colorItemBgSelected: 'rgba(59, 130, 246, 0.2)',
      colorItemTextHover: '#e2e8f0',
      colorItemBgHover: 'rgba(51, 65, 85, 0.3)',
    },
    Dropdown: {
      borderRadius: 12,
      colorBgElevated: 'rgba(17, 24, 39, 0.95)',
      colorBorderSecondary: 'rgba(75, 85, 99, 0.3)',
    },
  },
};

const app = createApp(App);

app.use(router)
   .use(store)
   .use(Antd)
   .provide('antdTheme', themeConfig)
   .mount("#app");
