<template>
  <nav
    class="bg-gradient-to-r from-gray-800 via-gray-900 to-gray-800 backdrop-blur-md p-4 shadow-xl fixed w-full z-10 transition-all duration-500 border-b border-gray-700/30"
  >
    <div class="container mx-auto flex justify-between items-center">
      <!-- Logo区域 - 优化动效但保持原色调 -->
      <div
        class="text-2xl font-medium text-white tracking-tight flex items-center group"
      >
        <i
          class="ri-global-line mr-2 text-gray-300 group-hover:text-cyan-400 transition-colors duration-300"
        ></i>
        <span class="group-hover:text-cyan-300 transition-colors duration-300">
          CyberEdge 综合扫描器
        </span>
      </div>

      <!-- 导航按钮区域 - 改进动效 -->
      <div class="space-x-5 relative">
        <!-- 未登录状态 -->
        <template v-if="!isAuthenticated">
          <router-link to="/login" v-slot="{ navigate }">
            <button
              @click="navigate"
              class="text-sm font-medium text-gray-200 hover:text-white transition-all duration-300 px-3 py-1.5 rounded-lg hover:bg-gray-700/50"
            >
              <i class="ri-login-box-line mr-1"></i>
              登录
            </button>
          </router-link>
          <router-link to="/setup-2fa" v-slot="{ navigate }">
            <button
              @click="navigate"
              class="text-sm font-medium text-gray-200 hover:text-white transition-all duration-300 px-3 py-1.5 rounded-lg hover:bg-gray-700/50"
            >
              <i class="ri-user-add-line mr-1"></i>
              注册
            </button>
          </router-link>
        </template>

        <!-- 登录状态 - 整合攻击面的菜单，保持灰色调 -->
        <template v-else>
          <!-- 主页按钮 -->
          <router-link to="/" v-slot="{ navigate }">
            <button @click="navigate" class="nav-button">
              <i class="ri-home-line mr-1"></i>
              主页
            </button>
          </router-link>

          <!-- 目标管理 -->
          <router-link to="/target-management" v-slot="{ navigate }">
            <button @click="navigate" class="nav-button">
              <i class="ri-focus-3-line mr-1"></i>
              目标管理
            </button>
          </router-link>

          <!-- 攻击面整合下拉菜单 -->
          <div class="relative group inline-block">
            <button
              @click="toggleDropdown('attackSurface')"
              class="nav-button flex items-center"
            >
              <i class="ri-radar-line mr-1"></i>
              攻击面
              <i
                class="ri-arrow-down-s-line ml-1 text-xs transition-transform duration-300"
                :class="{ 'rotate-180': dropdowns.attackSurface }"
              ></i>
            </button>
            <div
              v-show="dropdowns.attackSurface"
              class="dropdown-menu w-48"
              :class="{ 'dropdown-active': dropdowns.attackSurface }"
            >
              <!-- 攻击面搜集 -->
              <div class="dropdown-category">攻击面搜集</div>
              <router-link to="/subdomain-scan-results" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-global-line mr-1"></i>
                  子域名发现
                </button>
              </router-link>
              <router-link to="/port-scan-results" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-scan-2-line mr-1"></i>
                  端口扫描
                </button>
              </router-link>

              <!-- 攻击面刻画 -->
              <div class="dropdown-category">攻击面刻画</div>
              <router-link to="/path-scan-results" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-folders-line mr-1"></i>
                  路径扫描
                </button>
              </router-link>
              <router-link to="/under-development" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-fingerprint-line mr-1"></i>
                  指纹识别
                </button>
              </router-link>

              <!-- 攻击面渗透 -->
              <div class="dropdown-category">攻击面渗透</div>
              <router-link to="/under-development" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-bug-line mr-1"></i>
                  漏洞扫描
                </button>
              </router-link>
              <router-link to="/under-development" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-error-warning-line mr-1"></i>
                  漏洞利用️
                </button>
              </router-link>
            </div>
          </div>

          <!-- 任务管理 -->
          <router-link to="/task-management" v-slot="{ navigate }">
            <button @click="navigate" class="nav-button">
              <i class="ri-task-line mr-1"></i>
              任务管理
            </button>
          </router-link>

          <!-- 系统配置下拉菜单 -->
          <div class="relative group inline-block">
            <button
              @click="toggleDropdown('configuration')"
              class="nav-button flex items-center"
            >
              <i class="ri-settings-3-line mr-1"></i>
              系统配置
              <i
                class="ri-arrow-down-s-line ml-1 text-xs transition-transform duration-300"
                :class="{ 'rotate-180': dropdowns.configuration }"
              ></i>
            </button>
            <div
              v-show="dropdowns.configuration"
              class="dropdown-menu"
              :class="{ 'dropdown-active': dropdowns.configuration }"
            >
              <router-link to="/system-configuration" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-settings-3-line mr-1"></i>
                  系统配置
                </button>
              </router-link>
              <router-link to="/tool-configuration" v-slot="{ navigate }">
                <button @click="navigate" class="dropdown-item">
                  <i class="ri-tools-line mr-1"></i>
                  工具配置
                </button>
              </router-link>
            </div>
          </div>

          <!-- 用户管理 -->
          <router-link to="/user-management" v-slot="{ navigate }">
            <button @click="navigate" class="nav-button">
              <i class="ri-user-settings-line mr-1"></i>
              用户管理
            </button>
          </router-link>

          <!-- 综合扫描 - 特殊样式但保持灰色调 -->
          <router-link to="/under-development" v-slot="{ navigate }">
            <button
              @click="navigate"
              class="text-sm font-medium bg-gray-700 hover:bg-gray-600 text-white transition-all duration-300 px-4 py-1.5 rounded-lg shadow-md hover:shadow-lg"
            >
              <i class="ri-rocket-line mr-1"></i>
              综合扫描
            </button>
          </router-link>

          <!-- 登出按钮 -->
          <button @click="handleLogout" class="nav-button">
            <i class="ri-logout-box-line mr-1"></i>
            登出
          </button>
        </template>
      </div>
    </div>

    <!-- 通知组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :emoji="notificationEmoji"
      :type="notificationType"
      @close="showNotification = false"
    />
  </nav>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "vuex";
import PopupNotification from "./Utils/PopupNotification.vue";

export default {
  name: "HeaderPage",
  components: {
    PopupNotification,
  },
  setup() {
    const router = useRouter();
    const store = useStore();

    // 通知相关的状态
    const showNotification = ref(false);
    const notificationMessage = ref("");
    const notificationEmoji = ref("");
    const notificationType = ref("success");

    // 下拉菜单的状态 - 整合攻击面菜单
    const dropdowns = ref({
      attackSurface: false, // 整合后的攻击面菜单
      configuration: false, // 系统配置菜单
    });

    // 切换下拉菜单
    const toggleDropdown = (menu) => {
      // 阻止事件冒泡
      event?.stopPropagation();

      // 关闭其他菜单，只保持当前菜单的状态切换
      Object.keys(dropdowns.value).forEach((key) => {
        if (key !== menu) {
          dropdowns.value[key] = false;
        }
      });
      dropdowns.value[menu] = !dropdowns.value[menu];
    };

    // 关闭所有下拉菜单
    const closeAllDropdowns = () => {
      Object.keys(dropdowns.value).forEach((key) => {
        dropdowns.value[key] = false;
      });
    };

    // 登出处理
    const handleLogout = async () => {
      await store.dispatch("logout");
      notificationMessage.value = "登出成功！期待您的再次访问！";
      notificationEmoji.value = "👋";
      notificationType.value = "success";
      showNotification.value = true;

      // 延迟跳转到首页
      setTimeout(() => {
        router.push({ name: "Home" });
      }, 1500);
    };

    // 点击外部区域处理函数
    const handleClickOutside = (e) => {
      // 如果点击的是按钮本身，不处理
      if (e.target.closest("button")) return;

      // 如果点击的不是下拉菜单区域，则关闭所有下拉菜单
      if (!e.target.closest(".relative.group")) {
        closeAllDropdowns();
      }
    };

    // 组件挂载时添加事件监听
    onMounted(() => {
      document.addEventListener("click", handleClickOutside);
    });

    // 组件卸载前移除事件监听
    onBeforeUnmount(() => {
      document.removeEventListener("click", handleClickOutside);
    });

    return {
      isAuthenticated: computed(() => store.state.isAuthenticated),
      handleLogout,
      showNotification,
      notificationMessage,
      notificationEmoji,
      notificationType,
      dropdowns,
      toggleDropdown,
    };
  },
};
</script>

<style scoped>
/* 导航栏的玻璃态效果 */
nav {
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.3);
}

/* 通用导航按钮样式 */
.nav-button {
  @apply text-sm font-medium text-gray-200 hover:text-white transition-all duration-300 px-3 py-1.5 rounded-lg hover:bg-gray-700/50 relative overflow-hidden;
}

/* 下拉菜单基础样式 */
.dropdown-menu {
  @apply absolute left-0 bg-gray-800/90 backdrop-blur-md text-white rounded-lg shadow-xl mt-2 transition-all duration-300 border border-gray-700/30 opacity-0 transform -translate-y-2 pointer-events-none overflow-hidden w-40;
}

/* 活跃状态的下拉菜单 */
.dropdown-active {
  @apply opacity-100 transform translate-y-0 pointer-events-auto;
}

/* 下拉菜单中的分类标题 */
.dropdown-category {
  @apply px-3 py-2 text-xs text-gray-300 font-semibold border-b border-gray-700/30 bg-gray-800/50;
}

/* 下拉菜单中的项目 */
.dropdown-item {
  @apply block w-full text-left px-4 py-2 text-sm hover:bg-gray-700/50 transition-all duration-200 hover:pl-5;
}

/* 通用的按钮悬停波纹效果 */
.nav-button::after {
  content: "";
  @apply absolute rounded-full w-0 h-0 opacity-30 bg-gray-500;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  transition: width 0.6s, height 0.6s;
}

.nav-button:hover::after {
  @apply w-[150%] h-[150%];
}

/* 按钮点击效果 */
button:active {
  transform: scale(0.97);
}

/* 下拉菜单动画效果 - 优化 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeOut {
  from {
    opacity: 1;
    transform: translateY(0);
  }
  to {
    opacity: 0;
    transform: translateY(-10px);
  }
}
</style>
