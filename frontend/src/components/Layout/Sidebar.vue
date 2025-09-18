<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <div class="logo">
        <span class="logo-icon">🔒</span>
        <span class="logo-text">CyberEdge</span>
      </div>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <div class="nav-section-title">扫描管理</div>
        <router-link to="/projects" class="nav-item" active-class="active">
          <span class="nav-icon">📁</span>
          <span class="nav-text">扫描项目</span>
        </router-link>
        <router-link to="/vulnerabilities" class="nav-item" active-class="active">
          <span class="nav-icon">🚨</span>
          <span class="nav-text">漏洞管理</span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-section-title">系统管理</div>
        <router-link to="/user-management" class="nav-item" active-class="active">
          <span class="nav-icon">👥</span>
          <span class="nav-text">用户管理</span>
        </router-link>
        <router-link to="/settings" class="nav-item" active-class="active">
          <span class="nav-icon">⚙️</span>
          <span class="nav-text">系统设置</span>
        </router-link>
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="user-info">
        <div class="user-avatar">
          <span>{{ userInitial }}</span>
        </div>
        <div class="user-details">
          <div class="user-name">{{ $store.state.user?.username || 'User' }}</div>
          <div class="user-role">管理员</div>
        </div>
      </div>

      <div class="footer-actions">
        <router-link to="/profile" class="footer-link">
          <span class="icon">👤</span>
          个人资料
        </router-link>
        <button @click="logout" class="footer-link logout">
          <span class="icon">🚪</span>
          退出登录
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Sidebar',
  computed: {
    userInitial() {
      const username = this.$store.state.user?.username || 'U'
      return username.charAt(0).toUpperCase()
    }
  },
  methods: {
    async logout() {
      try {
        await this.$store.dispatch('logout')
        this.$router.push('/login')
      } catch (error) {
        console.error('退出登录失败:', error)
      }
    }
  }
}
</script>

<style scoped>
.sidebar {
  width: 260px;
  height: 100vh;
  background: #1f2937;
  color: white;
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 100;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #374151;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  font-size: 24px;
}

.logo-text {
  font-size: 20px;
  font-weight: 600;
  color: #f9fafb;
}

.sidebar-nav {
  flex: 1;
  padding: 20px 0;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: 32px;
}

.nav-section-title {
  padding: 0 20px 12px 20px;
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  color: #d1d5db;
  text-decoration: none;
  transition: all 0.2s ease;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background: #374151;
  color: #f9fafb;
}

.nav-item.active {
  background: #1e40af;
  color: white;
  border-left-color: #3b82f6;
}

.nav-icon {
  font-size: 18px;
  width: 20px;
  text-align: center;
}

.nav-text {
  font-weight: 500;
}

.sidebar-footer {
  border-top: 1px solid #374151;
  padding: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
}

.user-details {
  flex: 1;
}

.user-name {
  font-weight: 500;
  color: #f9fafb;
  margin-bottom: 2px;
}

.user-role {
  font-size: 12px;
  color: #9ca3af;
}

.footer-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.footer-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  color: #d1d5db;
  text-decoration: none;
  font-size: 14px;
  transition: all 0.2s ease;
  border: none;
  background: none;
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.footer-link:hover {
  background: #374151;
  color: #f9fafb;
}

.footer-link.logout:hover {
  background: #dc2626;
  color: white;
}

.icon {
  font-style: normal;
  font-size: 16px;
}
</style>