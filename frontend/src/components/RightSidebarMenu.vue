<template>
  <!-- 整个侧边栏容器 -->
  <div
    v-if="localVisible || isExiting"
    class="md: z-[1000]"
    :class="{ '': localVisible && !isExiting, '': isExiting, }"
    @click.stop
  >
    <!-- 背景和内容一体化设计 -->
    <div
      
    >
      <!-- 标题栏 -->
      <div
        
      >
        <h2 >
          <i class="ri-global-line"></i>
          网络请求工具箱
        </h2>

        <!-- 关闭按钮 -->
        <button
          @click="closeSidebar"
          class="hover: hover: duration-200"
        >
          <i class="ri-close-line"></i>
        </button>
      </div>

      <!-- 工具内容 - 直接渲染 -->
      <div >
        <HttpRequestTool />
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch, onMounted, onBeforeUnmount } from "vue";
import HttpRequestTool from "./Tools/HttpRequestTool.vue";

export default {
  name: "RightSidebarMenu",
  components: {
    HttpRequestTool,
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
          document.addEventListener("keydown", handleEscKey);
        } else {
          isExiting.value = true;
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
/* 优化后的动画效果 - 从右侧滑入 */
@keyframes slide-in {
  from {
    transform: translateX(100%);
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
    transform: translateX(100%);
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
