<template>
  <!-- 整个侧边栏容器 -->
  <div
    v-if="localVisible || isExiting"
    class="fixed top-0 left-0 h-full w-80 md:w-96 z-[1000]"
    :class="{
      'animate-slide-in': localVisible && !isExiting,
      'animate-slide-out': isExiting,
    }"
    @click.stop
  >
    <!-- 背景和内容一体化设计 -->
    <div
      class="h-full bg-gray-800/90 backdrop-blur-xl border-r border-gray-700/30 shadow-2xl overflow-auto"
    >
      <!-- 标题栏 -->
      <div
        class="p-6 flex justify-between items-center sticky top-0 bg-gray-800/80 backdrop-blur-sm z-10 border-b border-gray-700/30"
      >
        <h2 class="text-white text-lg font-medium flex items-center">
          <i class="ri-lock-unlock-line mr-2 text-cyan-400"></i>
          加密解密工具箱
        </h2>

        <!-- 关闭按钮 -->
        <button
          @click="closeSidebar"
          class="text-gray-400 hover:text-white p-2 rounded-lg hover:bg-gray-700/30 transition-all duration-200"
        >
          <i class="ri-close-line text-xl"></i>
        </button>
      </div>

      <!-- 工具内容 - 直接渲染 -->
      <div class="p-4">
        <CryptoTools />
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch, onMounted, onBeforeUnmount } from "vue";
import CryptoTools from "./Tools/CryptoTools.vue";

export default {
  name: "LeftSidebarMenu",
  components: {
    CryptoTools,
  },
  props: {
    isVisible: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["close"],
  setup(props, { emit }) {
    const localVisible = ref(props.isVisible);
    const isExiting = ref(false);

    // 关闭侧边栏
    const closeSidebar = () => {
      emit("close");
    };

    // 监听ESC键关闭侧边栏
    const handleEscKey = (event) => {
      if (event.key === "Escape" && localVisible.value) {
        closeSidebar();
      }
    };

    // 监听属性变化
    watch(
      () => props.isVisible,
      (newValue) => {
        if (newValue) {
          localVisible.value = true;
          isExiting.value = false;
          // 添加键盘事件监听
          document.addEventListener("keydown", handleEscKey);
        } else {
          isExiting.value = true;
          // 移除键盘事件监听
          document.removeEventListener("keydown", handleEscKey);
          setTimeout(() => {
            localVisible.value = false;
          }, 300);
        }
      }
    );

    onMounted(() => {
      localVisible.value = props.isVisible;
      if (props.isVisible) {
        document.addEventListener("keydown", handleEscKey);
      }
    });

    onBeforeUnmount(() => {
      document.removeEventListener("keydown", handleEscKey);
    });

    return {
      localVisible,
      isExiting,
      closeSidebar,
    };
  },
};
</script>

<style scoped>
/* 优化后的动画效果 */
@keyframes slide-in {
  from {
    transform: translateX(-100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes slide-out {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(-100%);
    opacity: 0;
  }
}

.animate-slide-in {
  animation: slide-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.animate-slide-out {
  animation: slide-out 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

/* 自定义滚动条样式 */
::-webkit-scrollbar {
  width: 5px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 10px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
</style>
