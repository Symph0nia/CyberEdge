<template>
  <a-config-provider :theme="theme">
    <div class="app-container">
      <!-- 侧边栏菜单 -->
      <LeftSidebarMenu
        :isVisible="isMenuVisible"
        @close="isMenuVisible = false"
      />

      <!-- 主内容区域 -->
      <div class="main-content">
        <router-view />
      </div>

      <!-- 右侧边栏菜单 -->
      <RightSidebarMenu
        :isVisible="isRequestToolVisible"
        @close="isRequestToolVisible = false"
      />

      <!-- 左侧工具箱按钮区域 -->
      <div class="left-tool-container" @click.stop>
        <a-button
          v-if="isAuthenticated"
          @click.stop="toggleMenu"
          class="tool-button"
          :class="{ 'active-tool': isMenuVisible }"
          type="default"
          size="large"
        >
          <template #icon>
            <i class="ri-lock-unlock-line"></i>
          </template>
          加密解密工具箱
        </a-button>
      </div>

      <!-- 右侧工具箱按钮区域 -->
      <div class="right-tool-container" @click.stop>
        <a-button
          v-if="isAuthenticated"
          @click.stop="toggleRequestTools"
          class="tool-button"
          :class="{ 'active-tool': isRequestToolVisible }"
          type="default"
          size="large"
        >
          <template #icon>
            <i class="ri-global-line"></i>
          </template>
          网络请求工具箱
        </a-button>
      </div>

      <!-- 全屏遮罩层 -->
      <div
        v-if="isMenuVisible || isRequestToolVisible"
        class="overlay"
        @click.stop="closeAllTools"
      ></div>
    </div>
  </a-config-provider>
</template>

<script>
import LeftSidebarMenu from "./components/LeftSidebarMenu.vue";
import RightSidebarMenu from "./components/RightSidebarMenu.vue";
import { ref, computed, inject } from "vue";
import { useStore } from "vuex";

export default {
  name: "App",
  components: {
    LeftSidebarMenu,
    RightSidebarMenu,
  },
  setup() {
    const store = useStore();
    const theme = inject('antdTheme');
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
      theme,
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
.app-container {
  min-height: 100vh;
  overflow-x: hidden;
}

.main-content {
  transition: all 0.3s ease;
}

.left-tool-container,
.right-tool-container {
  position: fixed;
  z-index: 2000;
}

.left-tool-container {
  left: 20px;
  top: 50%;
  transform: translateY(-50%);
}

.right-tool-container {
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
}

/* 工具按钮样式 */
.tool-button {
  background: rgba(30, 41, 59, 0.8) !important;
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3) !important;
  color: #e2e8f0 !important;
  min-width: 180px;
  border-radius: 16px !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
}

.tool-button:hover {
  background: rgba(51, 65, 85, 0.8) !important;
  border-color: rgba(34, 211, 238, 0.4) !important;
  color: #22d3ee !important;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2);
}

.tool-button.active-tool {
  background: rgba(51, 65, 85, 0.9) !important;
  border-color: rgba(34, 211, 238, 0.6) !important;
  color: #22d3ee !important;
  box-shadow: 0 0 15px rgba(34, 211, 238, 0.2);
}

.tool-button:active {
  transform: scale(0.98);
}

.overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  backdrop-filter: blur(2px);
}

/* 工具按钮图标样式 */
.tool-button i {
  margin-right: 8px;
  font-size: 16px;
  transition: color 0.3s ease;
}
</style>
