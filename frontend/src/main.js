import { createApp } from "vue";
import App from "./App.vue";
import "./main.css";
import "remixicon/fonts/remixicon.css";
import router from "./router";
import store from "./store";

// Ant Design Vue
import Antd from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';

// 全局ResizeObserver错误拦截 - 在应用启动前设置
const suppressResizeObserverErrors = () => {
  // 拦截全局错误
  const originalError = console.error;
  console.error = (...args) => {
    if (args[0] && typeof args[0] === 'string' &&
        args[0].includes('ResizeObserver loop completed with undelivered notifications')) {
      return;
    }
    originalError.apply(console, args);
  };

  // 拦截window错误事件
  window.addEventListener('error', (event) => {
    if (event.message && event.message.includes('ResizeObserver loop completed with undelivered notifications')) {
      event.preventDefault();
      event.stopPropagation();
      return false;
    }
  }, true);

  // 拦截Promise错误
  window.addEventListener('unhandledrejection', (event) => {
    if (event.reason && event.reason.message &&
        event.reason.message.includes('ResizeObserver loop completed with undelivered notifications')) {
      event.preventDefault();
      return false;
    }
  });

  // 劫持ResizeObserver构造函数
  if (window.ResizeObserver) {
    const OriginalResizeObserver = window.ResizeObserver;
    window.ResizeObserver = class extends OriginalResizeObserver {
      constructor(callback) {
        super((entries, observer) => {
          try {
            requestAnimationFrame(() => {
              callback(entries, observer);
            });
          } catch (error) {
            if (error.message && error.message.includes('ResizeObserver loop completed with undelivered notifications')) {
              return;
            }
            throw error;
          }
        });
      }
    };
  }
};

// 立即执行错误拦截
suppressResizeObserverErrors();


const app = createApp(App);

app.use(router)
   .use(store)
   .use(Antd)
   .mount("#app");
