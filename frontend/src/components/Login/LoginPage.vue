<template>
  <div class="bg-gray-900 flex items-center justify-center min-h-screen p-4">
    <!-- 登录卡片 -->
    <div
      class="bg-gray-800/40 backdrop-blur-xl p-8 md:p-10 rounded-3xl shadow-2xl w-full max-w-md border border-gray-700/30 login-card"
    >
      <div class="space-y-7">
        <!-- 标题区域 -->
        <div class="text-center space-y-3">
          <div
            class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-700/30 shadow-inner"
          >
            <i class="ri-shield-user-line text-2xl text-blue-300"></i>
          </div>
          <h2 class="text-xl font-medium tracking-wide text-gray-200">
            登录账户
          </h2>
          <p class="text-gray-500 text-sm max-w-xs mx-auto">
            登录您的账户以访问完整功能
          </p>
        </div>

        <!-- 表单 -->
        <form @submit.prevent="handleLogin" class="space-y-5">
          <!-- 账户输入 -->
          <div class="space-y-1.5">
            <label
              class="block text-sm font-medium text-gray-300"
              for="account"
            >
              <div class="flex items-center gap-2">
                <i class="ri-user-line"></i>
                <span>账户</span>
              </div>
            </label>
            <div class="relative">
              <input
                v-model="account"
                class="input-field"
                id="account"
                type="text"
                placeholder="输入账户名"
                autocomplete="username"
                required
              />
              <i
                class="ri-account-circle-line absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              ></i>
            </div>
          </div>

          <!-- 验证码输入 -->
          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-gray-300" for="code">
              <div class="flex items-center gap-2">
                <i class="ri-key-2-line"></i>
                <span>验证码</span>
              </div>
            </label>
            <div class="relative">
              <input
                v-model="code"
                class="input-field"
                id="code"
                type="text"
                placeholder="输入验证码"
                autocomplete="one-time-code"
                required
              />
              <i
                class="ri-lock-password-line absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              ></i>
            </div>
          </div>

          <!-- 按钮区域 -->
          <div class="space-y-3 pt-4">
            <button
              type="submit"
              class="primary-button"
              :disabled="isLoggingIn"
            >
              <i class="ri-login-circle-line"></i>
              <span>{{ isLoggingIn ? "登录中..." : "登录" }}</span>
              <span v-if="isLoggingIn" class="loading-dots"></span>
            </button>

            <button
              @click="goToRegister"
              type="button"
              class="secondary-button"
            >
              <i class="ri-user-add-line"></i>
              <span>注册账户</span>
            </button>
          </div>
        </form>

        <!-- 提示信息 -->
        <div
          class="flex items-center gap-2 text-xs text-gray-500 justify-center p-3 bg-gray-800/30 rounded-xl"
        >
          <i class="ri-information-line text-gray-400"></i>
          <span>使用 Google Authenticator 或其他2FA应用生成验证码</span>
        </div>
      </div>
    </div>

    <!-- 通知组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />
  </div>
</template>

<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "vuex";
import { useNotification } from "../../composables/useNotification";
import PopupNotification from "../Utils/PopupNotification.vue";

export default {
  name: "LoginPage",
  components: {
    PopupNotification,
  },
  setup() {
    const account = ref("");
    const code = ref("");
    const isLoggingIn = ref(false);
    const router = useRouter();
    const store = useStore();

    // 使用通知钩子
    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    const handleLogin = async () => {
      if (!account.value || !code.value) {
        showError("请填写账户和验证码");
        return;
      }

      try {
        isLoggingIn.value = true;
        const success = await store.dispatch("login", {
          account: account.value,
          code: code.value,
        });

        if (success) {
          showSuccess("登录成功，正在跳转...");
          setTimeout(() => {
            router.push({ name: "Home" });
          }, 1500);
        } else {
          throw new Error("登录失败");
        }
      } catch (error) {
        showError("登录失败，请检查账户名和验证码");
      } finally {
        isLoggingIn.value = false;
      }
    };

    const goToRegister = () => {
      router.push("/setup-2fa");
    };

    return {
      account,
      code,
      isLoggingIn,
      handleLogin,
      goToRegister,
      // 返回通知相关的状态
      showNotification,
      notificationMessage,
      notificationType,
    };
  },
};
</script>

<style scoped>
/* 登录卡片动画 */
.login-card {
  animation: card-appear 0.8s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  transform: translateY(20px);
  opacity: 0;
}

@keyframes card-appear {
  0% {
    opacity: 0;
    transform: translateY(20px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 输入框样式 */
.input-field {
  @apply w-full pl-10 pr-4 py-3 rounded-xl bg-gray-900/60 backdrop-blur-sm border border-gray-700/30 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-500/50 transition-all duration-200;
}

/* 主要按钮样式 */
.primary-button {
  @apply w-full px-4 py-3 rounded-xl bg-blue-600/80 hover:bg-blue-500/80 text-sm font-medium text-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-600/50 flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed shadow-lg shadow-blue-900/20;
}

/* 次要按钮样式 */
.secondary-button {
  @apply w-full px-4 py-3 rounded-xl bg-gray-800/70 hover:bg-gray-700/70 text-sm font-medium text-gray-200 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-600/50 flex items-center justify-center gap-2 border border-gray-700/30;
}

/* 加载动画 */
.loading-dots {
  position: relative;
  width: 10px;
  height: 10px;
  margin-left: 8px;
}

.loading-dots::after {
  content: "...";
  position: absolute;
  animation: dots 1.5s infinite;
  opacity: 0.7;
}

@keyframes dots {
  0%,
  20% {
    content: ".";
  }
  40% {
    content: "..";
  }
  60%,
  100% {
    content: "...";
  }
}

/* 毛玻璃效果 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 优化按钮点击效果 */
button:not(:disabled):active {
  transform: scale(0.98);
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
</style>
