<template>
  <a-drawer
    v-model:open="localVisible"
    title="系统设置"
    placement="right"
    :width="300"
    :closable="true"
    :mask="true"
    :keyboard="true"
    class="cyber-drawer"
    @close="closeSidebar"
  >
    <template #title>
      <div class="drawer-title">
        <i class="ri-settings-line cyber-icon"></i>
        系统设置
      </div>
    </template>

    <div class="drawer-content">
      <a-menu mode="vertical" theme="dark">
        <a-menu-item key="preferences">
          <i class="ri-settings-3-line"></i>
          偏好设置
        </a-menu-item>
        <a-menu-item key="about">
          <i class="ri-information-line"></i>
          关于系统
        </a-menu-item>
      </a-menu>
    </div>
  </a-drawer>
</template>

<script>
import { ref, watch } from "vue";

export default {
  name: "RightSidebarMenu",
  props: {
    isVisible: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["close"],
  setup(props, { emit }) {
    const localVisible = ref(props.isVisible);

    // 关闭侧边栏
    const closeSidebar = () => {
      emit("close");
    };

    // 监听属性变化
    watch(
      () => props.isVisible,
      (newValue) => {
        localVisible.value = newValue;
      }
    );

    return {
      localVisible,
      closeSidebar,
    };
  },
};
</script>

<style scoped>
/* 网络安全主题样式 */
:deep(.cyber-drawer .ant-drawer-content) {
  background: linear-gradient(135deg, #1f2937 0%, #111827 50%, #1f2937 100%);
  color: #ffffff;
}

:deep(.cyber-drawer .ant-drawer-header) {
  background: linear-gradient(135deg, #374151 0%, #1f2937 100%);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  padding: 16px 24px;
}

:deep(.cyber-drawer .ant-drawer-close) {
  color: #9ca3af;
  background: rgba(75, 85, 99, 0.2);
  border-radius: 6px;
  transition: all 0.3s ease;
}

:deep(.cyber-drawer .ant-drawer-close:hover) {
  color: #ffffff;
  background: rgba(75, 85, 99, 0.5);
}

:deep(.cyber-drawer .ant-drawer-body) {
  padding: 0;
}

.drawer-title {
  display: flex;
  align-items: center;
  color: #ffffff;
  font-size: 16px;
  font-weight: 600;
}

.cyber-icon {
  margin-right: 8px;
  color: #22d3ee;
  font-size: 18px;
}

.drawer-content {
  height: 100%;
  background: transparent;
}
</style>
