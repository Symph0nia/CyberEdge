<template>
  <div class="bg-gray-900 flex items-center justify-center min-h-screen p-4">
    <div
      class="bg-gray-800/40 backdrop-blur-xl p-8 md:p-10 rounded-3xl shadow-2xl w-full max-w-md border border-gray-700/30 register-card"
    >
      <div class="space-y-6">
        <!-- 标题区域 -->
        <div class="flex flex-col items-center text-center space-y-3">
          <div
            class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-700/30 shadow-inner"
          >
            <i class="ri-shield-keyhole-line text-2xl text-emerald-300"></i>
          </div>
          <div>
            <h2 class="text-xl font-medium tracking-wide text-gray-200">
              设置双重认证
            </h2>
            <p class="text-gray-500 text-sm mt-1">
              通过扫描二维码创建您的安全账户
            </p>
          </div>
        </div>

        <!-- 接口关闭状态 -->
        <div
          v-if="interfaceClosed"
          class="text-center p-6 bg-gray-900/30 rounded-xl border border-red-900/20"
        >
          <div
            class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-red-900/20 mb-3"
          >
            <i class="ri-close-circle-line text-xl text-red-400"></i>
          </div>
          <p class="text-red-400/90 text-sm font-medium mb-1">注册通道已关闭</p>
          <p class="text-gray-400 text-sm">系统维护中，请稍后重试</p>
        </div>

        <!-- 正常状态 -->
        <div v-else class="space-y-6">
          <!-- 二维码显示区域 -->
          <div v-if="qrCodeUrl" class="space-y-4">
            <div
              class="bg-gray-900/50 p-6 rounded-2xl border border-gray-700/30 flex flex-col items-center qr-container hover:border-emerald-700/30 transition-colors duration-300"
            >
              <div class="qr-overlay flex items-center justify-center">
                <button @click="fetchQRCode" class="refresh-qr-button">
                  <i class="ri-refresh-line mr-1.5"></i>
                  刷新二维码
                </button>
              </div>
              <img
                :src="qrCodeUrl"
                alt="认证二维码"
                class="mx-auto w-48 h-48 qr-image"
              />
              <p class="text-xs text-gray-500 mt-2">
                <i class="ri-time-line mr-1"></i>
                二维码有效期为10分钟
              </p>
            </div>

            <!-- 账户信息区域 -->
            <div
              class="bg-gray-900/30 p-5 rounded-xl space-y-3 border border-gray-700/30"
            >
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-400">账户名称</span>
                <div class="flex items-center gap-2">
                  <span class="text-sm text-gray-200 font-mono">{{
                    accountName
                  }}</span>
                  <button
                    @click="copyAccountName"
                    class="p-1.5 hover:bg-gray-700/50 rounded-lg transition-colors relative"
                    :class="{ copied: copied }"
                    :title="copied ? '已复制' : '复制账户名'"
                  >
                    <i
                      class="ri-file-copy-line text-gray-400 group-hover:text-gray-200"
                    ></i>
                    <span v-if="copied" class="copy-indicator"></span>
                  </button>
                </div>
              </div>
              <div
                class="flex items-center gap-1.5 text-xs text-gray-500 bg-gray-800/30 p-2 rounded-lg"
              >
                <i class="ri-information-line text-blue-400"></i>
                <span>请妥善保管账户名称，用于登录验证</span>
              </div>
            </div>
          </div>

          <!-- 加载状态 -->
          <div
            v-else-if="loading"
            class="flex flex-col items-center justify-center py-12"
          >
            <div class="loading-spinner mb-4"></div>
            <p class="text-sm text-gray-400">正在生成二维码...</p>
          </div>

          <!-- 初始状态 -->
          <div
            v-else
            class="bg-gray-900/30 p-5 rounded-xl border border-gray-700/30 space-y-4"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-800/50 flex items-center justify-center text-emerald-400"
              >
                <i class="ri-google-line"></i>
              </div>
              <div>
                <h3 class="text-sm font-medium text-gray-300 mb-1">
                  使用 Google Authenticator
                </h3>
                <p class="text-xs text-gray-500 leading-relaxed">
                  扫描二维码以启用双重认证，增强账户安全性
                </p>
              </div>
            </div>

            <div class="flex items-start gap-3">
              <div
                class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-800/50 flex items-center justify-center text-emerald-400"
              >
                <i class="ri-lock-password-line"></i>
              </div>
              <div>
                <h3 class="text-sm font-medium text-gray-300 mb-1">
                  安全登录保障
                </h3>
                <p class="text-xs text-gray-500 leading-relaxed">
                  通过验证码进行二次验证，有效防止未授权访问
                </p>
              </div>
            </div>
          </div>

          <!-- 按钮区域 -->
          <div class="space-y-3 pt-2">
            <button
              v-if="!qrCodeUrl"
              @click="fetchQRCode"
              :disabled="loading"
              class="primary-button"
            >
              <i class="ri-qr-code-line mr-1.5"></i>
              生成二维码
            </button>

            <button @click="goToLogin" class="secondary-button">
              <i class="ri-login-circle-line mr-1.5"></i>
              去登录
            </button>
          </div>
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
import { useNotification } from "../../composables/useNotification";
import api from "../../api/axiosInstance";
import PopupNotification from "../Utils/PopupNotification.vue";

export default {
  name: "GoogleAuthQRCode",
  components: {
    PopupNotification,
  },
  setup() {
    const router = useRouter();
    const qrCodeUrl = ref("");
    const loading = ref(false);
    const interfaceClosed = ref(false);
    const accountName = ref("");
    const copied = ref(false);

    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    const copyAccountName = async () => {
      try {
        await navigator.clipboard.writeText(accountName.value);
        copied.value = true;
        showSuccess("账户名已复制到剪贴板");
        setTimeout(() => {
          copied.value = false;
        }, 2000);
      } catch (err) {
        showError("复制失败，请手动复制");
      }
    };

    const fetchQRCode = async () => {
      loading.value = true;
      try {
        const response = await api.get("/auth/qrcode");
        const { qrcode, account } = response.data;
        qrCodeUrl.value = `data:image/png;base64,${qrcode}`;
        accountName.value = account;
        interfaceClosed.value = false;
        showSuccess("二维码已成功生成");
      } catch (error) {
        console.error("获取二维码失败:", error);
        interfaceClosed.value = true;

        if (error.response?.data?.error === "二维码接口已关闭") {
          showError("注册接口暂时关闭，请稍后再试");
        } else {
          showError("生成二维码失败，请重试");
        }
      } finally {
        loading.value = false;
      }
    };

    const goToLogin = () => {
      router.push("/login");
    };

    return {
      qrCodeUrl,
      loading,
      interfaceClosed,
      accountName,
      copied,
      fetchQRCode,
      copyAccountName,
      showNotification,
      notificationMessage,
      notificationType,
      goToLogin,
    };
  },
};
</script>

<style scoped>
/* 卡片进入动画 */
.register-card {
  animation: slide-up 0.8s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  transform: translateY(20px);
  opacity: 0;
}

@keyframes slide-up {
  0% {
    opacity: 0;
    transform: translateY(20px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 背景模糊效果 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 主要按钮样式 */
.primary-button {
  @apply w-full px-4 py-3 rounded-xl bg-emerald-600/80 hover:bg-emerald-500/80 text-sm font-medium text-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-emerald-600/50 flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed shadow-lg shadow-emerald-900/20;
}

/* 次要按钮样式 */
.secondary-button {
  @apply w-full px-4 py-3 rounded-xl bg-gray-800/70 hover:bg-gray-700/70 text-sm font-medium text-gray-200 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-600/50 flex items-center justify-center gap-2 border border-gray-700/30;
}

/* 二维码区域样式 */
.qr-container {
  position: relative;
  overflow: hidden;
}

.qr-image {
  transition: filter 0.3s ease;
}

.qr-overlay {
  position: absolute;
  inset: 0;
  background: rgba(17, 24, 39, 0.7);
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 10;
  backdrop-filter: blur(2px);
}

.qr-container:hover .qr-overlay {
  opacity: 1;
}

.qr-container:hover .qr-image {
  filter: blur(1px);
}

/* 刷新二维码按钮 */
.refresh-qr-button {
  @apply bg-emerald-600/90 hover:bg-emerald-500/90 text-white px-4 py-2 rounded-lg text-sm font-medium flex items-center justify-center transition-all duration-200 shadow-lg;
}

/* 复制状态指示器 */
.copy-indicator {
  position: absolute;
  width: 5px;
  height: 5px;
  background-color: #10b981;
  border-radius: 50%;
  top: 0;
  right: 0;
}

/* 已复制样式 */
.copied i {
  color: #10b981 !important;
}

/* 加载动画 */
.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(16, 185, 129, 0.2);
  border-radius: 50%;
  border-top: 3px solid rgba(16, 185, 129, 0.8);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
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
