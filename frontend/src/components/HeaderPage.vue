<template>
  <a-layout-header class="cyber-header">
    <div class="header-content">
      <!-- Logo区域 -->
      <div class="logo-section">
        <i class="ri-global-line logo-icon"></i>
        <span class="logo-text">CyberEdge 综合扫描器</span>
      </div>

      <!-- 导航按钮区域 -->
      <div class="nav-section">
        <!-- 未登录状态 -->
        <template v-if="!isAuthenticated">
          <a-space size="middle">
            <router-link to="/login">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-login-box-line"></i></template>
                登录
              </a-button>
            </router-link>
            <router-link to="/setup-2fa">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-user-add-line"></i></template>
                注册
              </a-button>
            </router-link>
          </a-space>
        </template>

        <!-- 登录状态 -->
        <template v-else>
          <a-space size="middle">
            <!-- 主页按钮 -->
            <router-link to="/">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-home-line"></i></template>
                主页
              </a-button>
            </router-link>

            <!-- 目标管理 -->
            <router-link to="/target-management">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-focus-3-line"></i></template>
                目标管理
              </a-button>
            </router-link>

            <!-- 攻击面下拉菜单 -->
            <a-dropdown placement="bottomRight" class="nav-dropdown">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-radar-line"></i></template>
                攻击面
                <template #suffix><i class="ri-arrow-down-s-line"></i></template>
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item-group title="攻击面搜集">
                    <a-menu-item key="subdomain">
                      <router-link to="/subdomain-scan-results">
                        <i class="ri-global-line"></i> 子域名发现
                      </router-link>
                    </a-menu-item>
                    <a-menu-item key="port">
                      <router-link to="/port-scan-results">
                        <i class="ri-scan-2-line"></i> 端口扫描
                      </router-link>
                    </a-menu-item>
                  </a-menu-item-group>
                  <a-menu-divider />
                  <a-menu-item-group title="攻击面刻画">
                    <a-menu-item key="path">
                      <router-link to="/path-scan-results">
                        <i class="ri-folders-line"></i> 路径扫描
                      </router-link>
                    </a-menu-item>
                    <a-menu-item key="fingerprint">
                      <router-link to="/under-development">
                        <i class="ri-fingerprint-line"></i> 指纹识别
                      </router-link>
                    </a-menu-item>
                  </a-menu-item-group>
                  <a-menu-divider />
                  <a-menu-item-group title="攻击面渗透">
                    <a-menu-item key="vuln-scan">
                      <router-link to="/under-development">
                        <i class="ri-bug-line"></i> 漏洞扫描
                      </router-link>
                    </a-menu-item>
                    <a-menu-item key="exploit">
                      <router-link to="/under-development">
                        <i class="ri-error-warning-line"></i> 漏洞利用
                      </router-link>
                    </a-menu-item>
                  </a-menu-item-group>
                </a-menu>
              </template>
            </a-dropdown>

            <!-- 任务管理 -->
            <router-link to="/task-management">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-task-line"></i></template>
                任务管理
              </a-button>
            </router-link>

            <!-- 系统配置下拉菜单 -->
            <a-dropdown placement="bottomRight" class="nav-dropdown">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-settings-3-line"></i></template>
                系统配置
                <template #suffix><i class="ri-arrow-down-s-line"></i></template>
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="system">
                    <router-link to="/system-configuration">
                      <i class="ri-settings-2-line"></i> 系统配置
                    </router-link>
                  </a-menu-item>
                  <a-menu-item key="tools">
                    <router-link to="/tool-configuration">
                      <i class="ri-tools-line"></i> 工具配置
                    </router-link>
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>

            <!-- 工具集合 -->
            <router-link to="/tools">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-hammer-line"></i></template>
                工具
              </a-button>
            </router-link>

            <!-- 用户菜单 -->
            <a-dropdown placement="bottomRight" class="nav-dropdown">
              <a-button type="text" class="nav-btn">
                <template #icon><i class="ri-user-line"></i></template>
                {{ currentUser }}
                <template #suffix><i class="ri-arrow-down-s-line"></i></template>
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="profile">
                    <router-link to="/user-management">
                      <i class="ri-user-settings-line"></i> 用户管理
                    </router-link>
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="logout" @click="logout">
                    <i class="ri-logout-box-line"></i> 退出登录
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </a-space>
        </template>
      </div>
    </div>
  </a-layout-header>
</template>

<script>
import { computed } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'

export default {
  name: 'HeaderPage',
  setup() {
    const store = useStore()
    const router = useRouter()

    const isAuthenticated = computed(() => store.getters.isAuthenticated)
    const currentUser = computed(() => store.getters.currentUser || 'Admin')

    const logout = () => {
      store.dispatch('logout')
      router.push('/login')
    }

    return {
      isAuthenticated,
      currentUser,
      logout
    }
  }
}
</script>

<style scoped>
.cyber-header {
  background: linear-gradient(135deg, #1f2937 0%, #111827 50%, #1f2937 100%);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  position: fixed;
  top: 0;
  width: 100%;
  z-index: 1000;
  height: auto;
  line-height: normal;
  padding: 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 64px;
}

.logo-section {
  display: flex;
  align-items: center;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: -0.025em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.logo-section:hover {
  color: #06b6d4;
}

.logo-icon {
  margin-right: 8px;
  color: #9ca3af;
  transition: color 0.3s ease;
}

.logo-section:hover .logo-icon {
  color: #22d3ee;
}

.logo-text {
  transition: color 0.3s ease;
}

.nav-section {
  display: flex;
  align-items: center;
}

.nav-btn {
  color: #d1d5db !important;
  font-weight: 500;
  border: none !important;
  height: 40px;
  padding: 0 16px;
  border-radius: 8px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-btn:hover {
  background-color: rgba(75, 85, 99, 0.5) !important;
  color: #ffffff !important;
  transform: translateY(-1px);
}

.nav-btn:focus {
  background-color: rgba(75, 85, 99, 0.3) !important;
  color: #ffffff !important;
}

/* 下拉菜单样式覆盖 */
.nav-dropdown .ant-dropdown-menu {
  background: rgba(17, 24, 39, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 12px;
  padding: 8px 0;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
  min-width: 180px;
}

.nav-dropdown .ant-dropdown-menu-item-group-title {
  color: #9ca3af;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 8px 16px 4px;
  margin: 0;
}

.nav-dropdown .ant-dropdown-menu-item {
  color: #d1d5db;
  padding: 8px 16px;
  margin: 2px 8px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.nav-dropdown .ant-dropdown-menu-item:hover {
  background-color: rgba(59, 130, 246, 0.2);
  color: #ffffff;
}

.nav-dropdown .ant-dropdown-menu-item a {
  color: inherit;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-dropdown .ant-dropdown-menu-item i {
  font-size: 16px;
  width: 16px;
  text-align: center;
}

.nav-dropdown .ant-dropdown-menu-divider {
  background-color: rgba(75, 85, 99, 0.3);
  margin: 8px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }

  .logo-text {
    font-size: 18px;
  }

  .nav-btn {
    padding: 0 12px;
    font-size: 14px;
  }
}

/* 路由激活状态 */
.router-link-active .nav-btn {
  background-color: rgba(59, 130, 246, 0.2) !important;
  color: #60a5fa !important;
}
</style>