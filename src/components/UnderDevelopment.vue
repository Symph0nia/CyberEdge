<template>
  <div
    class="min-h-screen bg-gray-900 flex flex-col items-center justify-center p-6 relative overflow-hidden"
  >
    <!-- 背景装饰元素 -->
    <div class="absolute inset-0 overflow-hidden">
      <div class="gear-large"></div>
      <div class="gear-small"></div>
      <div class="code-block"></div>
    </div>

    <!-- 主要内容卡片 -->
    <div
      class="bg-gray-800/40 backdrop-blur-xl p-10 rounded-2xl shadow-2xl border border-gray-700/30 max-w-2xl w-full relative z-10"
    >
      <div class="space-y-8">
        <!-- 标题和图标 -->
        <div class="space-y-4 text-center">
          <div
            class="inline-flex items-center justify-center w-20 h-20 bg-gray-700/50 rounded-2xl mx-auto overflow-hidden border border-gray-600/30 shadow-inner group"
          >
            <span class="text-3xl animate-bounce-slow">🚧</span>
          </div>
          <h1 class="text-2xl font-medium tracking-wide text-gray-200">
            功能开发中
          </h1>
        </div>

        <!-- 说明文本 -->
        <p
          class="text-sm text-gray-400 leading-relaxed text-center max-w-md mx-auto"
        >
          该功能正在开发中，我们正在努力完善这项服务。
          <br />感谢您的耐心等待，敬请期待！
        </p>

        <!-- 开发进度指示器 -->
        <div class="space-y-3 max-w-md mx-auto">
          <div class="flex justify-between text-xs text-gray-500 px-1">
            <span>开发阶段</span>
            <span>75%</span>
          </div>
          <div
            class="w-full bg-gray-900/70 rounded-full h-2 overflow-hidden shadow-inner"
          >
            <div class="progress-bar h-2 rounded-full"></div>
          </div>
          <div class="flex justify-between text-xs text-gray-500 mt-2">
            <span>计划</span>
            <span>设计</span>
            <span>开发</span>
            <span>测试</span>
            <span>发布</span>
          </div>
        </div>

        <!-- 返回按钮 -->
        <div class="flex justify-center pt-6">
          <button
            @click="handleReturn"
            class="return-button px-6 py-2.5 rounded-xl text-sm font-medium bg-gray-700/50 hover:bg-gray-600/50 text-gray-200 transition-all duration-300 focus:outline-none border border-gray-700/30 hover:border-gray-500/50 group"
          >
            <span class="relative z-10 flex items-center">
              <i
                class="ri-arrow-left-line mr-2 transition-transform duration-300 group-hover:-translate-x-1"
              ></i>
              返回上一页
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- 底部信息 -->
    <p class="mt-8 text-xs text-gray-500 relative z-10">
      如有建议或问题，请
      <a
        href="#"
        class="text-gray-400 hover:text-gray-300 underline underline-offset-2"
        >与我们联系</a
      >
    </p>
  </div>

  <!-- 通知组件 -->
  <PopupNotification
    v-if="showNotification"
    :message="notificationMessage"
    :type="notificationType"
    @close="showNotification = false"
  />
</template>

<script>
import { useRouter } from "vue-router";
import { useNotification } from "../composables/useNotification";
import PopupNotification from "./Utils/PopupNotification.vue";

export default {
  name: "UnderDevelopment",
  components: {
    PopupNotification,
  },
  setup() {
    const router = useRouter();
    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
    } = useNotification();

    const handleReturn = () => {
      showSuccess("正在返回上一页");
      router.go(-1);
    };

    return {
      // 通知相关
      showNotification,
      notificationMessage,
      notificationType,
      // 方法
      handleReturn,
    };
  },
};
</script>

<style scoped>
/* 背景模糊效果 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 进度条动画 */
.progress-bar {
  background: linear-gradient(
    90deg,
    rgba(59, 130, 246, 0.5) 0%,
    rgba(147, 197, 253, 0.3) 100%
  );
  width: 0;
  animation: progress 2.5s ease-out forwards;
}

@keyframes progress {
  0% {
    width: 0;
  }
  20% {
    width: 35%;
  }
  50% {
    width: 60%;
  }
  80% {
    width: 70%;
  }
  100% {
    width: 75%;
  }
}

/* 慢速弹跳动画 */
.animate-bounce-slow {
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%,
  100% {
    transform: translateY(-10%);
    animation-timing-function: cubic-bezier(0.8, 0, 1, 1);
  }
  50% {
    transform: translateY(0);
    animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
  }
}

/* 背景装饰动画 */
.gear-large {
  position: absolute;
  width: 300px;
  height: 300px;
  border-radius: 50%;
  border: 15px dashed rgba(107, 114, 128, 0.1);
  top: 10%;
  right: -80px;
  animation: spin 20s linear infinite;
}

.gear-small {
  position: absolute;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  border: 12px dashed rgba(107, 114, 128, 0.1);
  bottom: 15%;
  left: -50px;
  animation: spin 15s linear infinite reverse;
}

.code-block {
  position: absolute;
  width: 150px;
  height: 150px;
  background: repeating-linear-gradient(
    to bottom,
    rgba(75, 85, 99, 0.05) 0px,
    rgba(75, 85, 99, 0.05) 3px,
    transparent 3px,
    transparent 6px
  );
  border-radius: 8px;
  bottom: 20%;
  right: 15%;
  transform: rotate(-15deg);
  opacity: 0.5;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 按钮特效 */
.return-button {
  position: relative;
  overflow: hidden;
}

.return-button::after {
  content: "";
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  transition: width 0.6s, height 0.6s;
}

.return-button:hover::after {
  width: 300%;
  height: 300%;
}

/* 优化按钮点击效果 */
button:active {
  transform: scale(0.98);
}
</style>
