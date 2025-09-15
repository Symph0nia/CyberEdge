<template>
  <div class="flex flex-col min-h-screen relative overflow-hidden">
    <!-- 侧边栏菜单 -->
    <LeftSidebarMenu
      :isVisible="isMenuVisible"
      @close="isMenuVisible = false"
    />

    <!-- 主内容区域 -->
    <div class="flex-1 transition-all duration-300">
      <router-view />
    </div>

    <!-- 右侧边栏菜单 -->
    <RightSidebarMenu
      :isVisible="isRequestToolVisible"
      @close="isRequestToolVisible = false"
    />

    <!-- 左侧工具箱按钮区域 - 添加事件修饰符 -->
    <div class="fixed left-6 bottom-6 z-[2000]" @click.stop>
      <button
        v-if="isAuthenticated"
        @click.stop="toggleMenu"
        class="tool-button group"
        :class="{ 'active-tool': isMenuVisible }"
        type="button"
      >
        <span class="flex items-center">
          <i
            class="ri-lock-unlock-line mr-2 group-hover:text-cyan-400 transition-colors duration-300"
          ></i>
          加密解密工具箱
        </span>
      </button>
    </div>

    <!-- 右侧工具箱按钮区域 - 添加事件修饰符 -->
    <div class="fixed right-6 bottom-6 z-[2000]" @click.stop>
      <button
        v-if="isAuthenticated"
        @click.stop="toggleRequestTools"
        class="tool-button group"
        :class="{ 'active-tool': isRequestToolVisible }"
        type="button"
      >
        <span class="flex items-center">
          <i
            class="ri-global-line mr-2 group-hover:text-cyan-400 transition-colors duration-300"
          ></i>
          网络请求工具箱
        </span>
      </button>
    </div>

    <!-- 全屏遮罩层，阻止事件穿透 -->
    <div
      v-if="isMenuVisible || isRequestToolVisible"
      class="fixed inset-0 z-[999]"
      @click.stop="closeAllTools"
    ></div>
  </div>
</template>

<script>
import LeftSidebarMenu from "./components/LeftSidebarMenu.vue";
import RightSidebarMenu from "./components/RightSidebarMenu.vue";
import { ref, computed } from "vue";
import { useStore } from "vuex";

export default {
  name: "App",
  components: {
    LeftSidebarMenu,
    RightSidebarMenu,
  },
  setup() {
    const store = useStore();
    const isAuthenticated = computed(() => store.state.isAuthenticated);
    const isMenuVisible = ref(false);
    const isRequestToolVisible = ref(false);

    // 关闭所有工具箱
    const closeAllTools = () => {
      isMenuVisible.value = false;
      isRequestToolVisible.value = false;
    };

    const toggleMenu = (event) => {
      // 停止事件传播
      if (event) {
        event.stopPropagation();
      }

      isMenuVisible.value = !isMenuVisible.value;
      // 如果打开左侧菜单，确保关闭右侧菜单
      if (isMenuVisible.value) {
        isRequestToolVisible.value = false;
      }
    };

    const toggleRequestTools = (event) => {
      // 停止事件传播
      if (event) {
        event.stopPropagation();
      }

      isRequestToolVisible.value = !isRequestToolVisible.value;
      // 如果打开右侧菜单，确保关闭左侧菜单
      if (isRequestToolVisible.value) {
        isMenuVisible.value = false;
      }
    };

    return {
      isAuthenticated,
      isMenuVisible,
      isRequestToolVisible,
      toggleMenu,
      toggleRequestTools,
      closeAllTools,
    };
  },
};
</script>

<style scoped>
/* 工具按钮基础样式 */
.tool-button {
  @apply bg-gray-800/80 backdrop-blur-md text-white text-sm font-medium px-6 py-3 rounded-2xl
  hover:bg-gray-700/80 transition-all duration-300 border border-gray-600/30 focus:outline-none
  shadow-lg hover:shadow-xl tracking-wide flex items-center justify-center;
  min-width: 180px;
}

/* 活跃状态的工具按钮 */
.active-tool {
  @apply bg-gray-700/90 border-cyan-600/30 shadow-cyan-900/10;
  box-shadow: 0 0 15px rgba(8, 145, 178, 0.2);
}

/* 按钮按下效果 */
button:active {
  transform: scale(0.98);
}

/* 添加按钮悬停时的轻微提升效果 */
.tool-button:hover:not(.active-tool) {
  transform: translateY(-2px);
}

/* 创建渐变光晕效果 */
.tool-button::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(
    circle at center,
    rgba(103, 232, 249, 0.1) 0%,
    transparent 70%
  );
  border-radius: inherit;
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.tool-button:hover::before {
  opacity: 1;
}
</style>
