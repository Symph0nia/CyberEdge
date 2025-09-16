<template>
  <div class="app-container">
    <!-- 登录页面 -->
    <div v-if="!isAuthenticated">
      <router-view />
    </div>

    <!-- 认证后的主布局 -->
    <a-layout v-else style="min-height: 100vh; height: 100vh; overflow: hidden;">
      <!-- 顶部导航栏 -->
      <a-layout-header style="background: #fff; padding: 0 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.06); height: 64px; flex-shrink: 0;">
        <div style="display: flex; align-items: center; justify-content: space-between; height: 100%;">
          <!-- 左侧 Logo 和标题 -->
          <div style="display: flex; align-items: center;">
            <a-avatar :size="40" style="background-color: #1890ff; margin-right: 16px;">
              <template #icon><SecurityScanOutlined /></template>
            </a-avatar>
            <h1 style="margin: 0; font-size: 24px; font-weight: 600; color: #1890ff;">CyberEdge</h1>
          </div>

          <!-- 右侧用户信息和操作 -->
          <div style="display: flex; align-items: center;">
            <a-space :size="24">
              <!-- 用户信息 -->
              <a-dropdown>
                <a-button type="text" style="height: auto; padding: 8px 12px;">
                  <a-space>
                    <a-avatar :size="32" style="background-color: #52c41a;">
                      <template #icon><UserOutlined /></template>
                    </a-avatar>
                    <span>{{ currentUser.username || '用户' }}</span>
                    <DownOutlined />
                  </a-space>
                </a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item key="profile">
                      <UserOutlined />
                      个人资料
                    </a-menu-item>
                    <a-menu-item key="settings">
                      <SettingOutlined />
                      设置
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item key="logout" @click="handleLogout">
                      <LogoutOutlined />
                      退出登录
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </div>
        </div>
      </a-layout-header>

      <a-layout style="height: calc(100vh - 64px);">
        <!-- 左侧菜单 -->
        <a-layout-sider
          v-model:collapsed="collapsed"
          :trigger="null"
          collapsible
          :width="200"
          :collapsed-width="80"
          style="background: #fff; box-shadow: 2px 0 8px rgba(0,0,0,0.06); height: 100%;"
        >
          <!-- 折叠按钮 -->
          <div style="padding: 16px; text-align: center;">
            <a-button
              type="text"
              @click="collapsed = !collapsed"
              style="width: 100%;"
            >
              <MenuUnfoldOutlined v-if="collapsed" />
              <MenuFoldOutlined v-else />
            </a-button>
          </div>

          <!-- 菜单项 -->
          <a-menu
            v-model:selectedKeys="selectedKeys"
            mode="inline"
            style="border-right: 0;"
            @click="handleMenuClick"
          >
            <a-menu-item key="user-management">
              <UserOutlined />
              <span>用户管理</span>
            </a-menu-item>
            <a-menu-item key="security-scan">
              <SecurityScanOutlined />
              <span>安全扫描</span>
            </a-menu-item>
            <a-menu-item key="vulnerability">
              <BugOutlined />
              <span>漏洞管理</span>
            </a-menu-item>
            <a-menu-item key="reports">
              <FileTextOutlined />
              <span>报告中心</span>
            </a-menu-item>
            <a-menu-item key="settings">
              <SettingOutlined />
              <span>系统设置</span>
            </a-menu-item>
          </a-menu>
        </a-layout-sider>

        <!-- 主内容区域 -->
        <a-layout-content style="padding: 24px; background: #f0f2f5; overflow: auto; height: 100%;">
          <router-view />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useStore } from 'vuex'
import { message } from 'ant-design-vue'
import {
  UserOutlined,
  SecurityScanOutlined,
  BugOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  DownOutlined
} from '@ant-design/icons-vue'

export default {
  name: 'App',
  components: {
    UserOutlined,
    SecurityScanOutlined,
    BugOutlined,
    FileTextOutlined,
    SettingOutlined,
    LogoutOutlined,
    MenuUnfoldOutlined,
    MenuFoldOutlined,
    DownOutlined
  },
  setup() {
    const router = useRouter()
    const route = useRoute()
    const store = useStore()

    const collapsed = ref(false)
    const selectedKeys = ref(['user-management'])

    const isAuthenticated = computed(() => store.state.isAuthenticated)
    const currentUser = computed(() => store.state.user || {})

    // 根据当前路由设置选中的菜单项
    watch(() => route.path, (newPath) => {
      if (newPath === '/user-management') {
        selectedKeys.value = ['user-management']
      } else if (newPath === '/security-scan') {
        selectedKeys.value = ['security-scan']
      } else if (newPath === '/vulnerability') {
        selectedKeys.value = ['vulnerability']
      } else if (newPath === '/reports') {
        selectedKeys.value = ['reports']
      } else if (newPath === '/settings') {
        selectedKeys.value = ['settings']
      }
    }, { immediate: true })

    // 处理菜单点击
    const handleMenuClick = ({ key }) => {
      switch (key) {
        case 'user-management':
          router.push('/user-management')
          break
        case 'security-scan':
          message.info('安全扫描功能开发中...')
          break
        case 'vulnerability':
          message.info('漏洞管理功能开发中...')
          break
        case 'reports':
          message.info('报告中心功能开发中...')
          break
        case 'settings':
          message.info('系统设置功能开发中...')
          break
      }
    }

    // 处理退出登录
    const handleLogout = () => {
      store.dispatch('logout')
      message.success('已退出登录')
      router.push('/login')
    }

    onMounted(() => {
      // 检查认证状态
      store.dispatch('checkAuth')

      // 修复ResizeObserver循环 - 延迟强制重新计算布局
      setTimeout(() => {
        window.dispatchEvent(new Event('resize'))
      }, 100)
    })

    return {
      collapsed,
      selectedKeys,
      isAuthenticated,
      currentUser,
      handleMenuClick,
      handleLogout
    }
  }
}
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  height: 100vh;
  overflow: hidden;
  position: relative;
}

/* 自定义滚动条 - 修复ResizeObserver循环 */
:deep(.ant-layout-sider-children) {
  overflow-y: auto;
  overflow-x: hidden;
  height: 100%;
}

:deep(.ant-layout-sider-children)::-webkit-scrollbar {
  width: 6px;
}

:deep(.ant-layout-sider-children)::-webkit-scrollbar-track {
  background: #f1f1f1;
}

:deep(.ant-layout-sider-children)::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

:deep(.ant-layout-sider-children)::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 菜单项样式优化 */
:deep(.ant-menu-item) {
  margin: 4px 8px;
  border-radius: 6px;
}

:deep(.ant-menu-item-selected) {
  background-color: #e6f7ff !important;
  border-radius: 6px;
}

/* 头部样式 */
:deep(.ant-layout-header) {
  position: sticky;
  top: 0;
  z-index: 1000;
}

/* 修复ResizeObserver循环 - 稳定布局尺寸 */
:deep(.ant-layout) {
  contain: layout style size;
}

:deep(.ant-layout-sider) {
  will-change: auto;
  transition: none !important;
}

:deep(.ant-layout-content) {
  contain: layout style;
}
</style>