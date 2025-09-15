<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 系统运行信息卡片 -->
      <div
        class="border"
      >
        <!-- 标题和刷新按钮 -->
        <div >
          <div >
            <div
              
            >
              <i class="ri-dashboard-3-line"></i>
            </div>
            <h2 >
              系统状态
            </h2>
          </div>
          <button
            @click="fetchSystemInfo"
            class=".5 hover: duration-200 group"
          >
            <i
              class="ri-refresh-line group- duration-500"
            ></i>
            刷新信息
          </button>
        </div>

        <!-- 系统信息概览 -->
        <div
          class="border"
        >
          <div >
            <div
              
            >
              <i class="ri-server-line"></i>
            </div>
            <div>
              <p >系统概览</p>
              <h3 >
                {{ systemInfo?.osDistribution || "加载中..." }}
              </h3>
            </div>
          </div>

          <div >
            <div>
              <p >内核版本</p>
              <p >
                {{ systemInfo?.kernelVersion || "加载中..." }}
              </p>
            </div>
            <div>
              <p >运行权限</p>
              <p
                
                :class="{ '': systemInfo?.privileges === 'root' }"
              >
                {{ systemInfo?.privileges || "加载中..." }}
              </p>
            </div>
          </div>
        </div>

        <!-- 系统信息卡片 -->
        <div v-if="systemInfo">
          <div class="md: lg:">
            <!-- 当前路径卡片 -->
            <div class="info-card">
              <div >
                <div
                  
                >
                  <i class="ri-folder-line"></i>
                </div>
                <h3 >程序运行目录</h3>
              </div>
              <div >
                <p class="overflow-">
                  {{ systemInfo.currentDirectory }}
                </p>
              </div>
              <button
                @click="copyToClipboard(systemInfo.currentDirectory)"
                class="hover:"
              >
                <i class="ri-file-copy-line"></i> 复制路径
              </button>
            </div>

            <!-- 本机IP卡片 -->
            <div class="info-card">
              <div >
                <div
                  
                >
                  <i class="ri-computer-line"></i>
                </div>
                <h3 >本机 IP</h3>
              </div>
              <div >
                <p >
                  {{ systemInfo.localIP }}
                </p>
              </div>
              <div >内部网络地址</div>
            </div>

            <!-- 外网IP卡片 -->
            <div class="info-card">
              <div >
                <div
                  
                >
                  <i class="ri-global-line"></i>
                </div>
                <h3 >外网 IP</h3>
              </div>
              <div >
                <p >
                  {{ systemInfo.publicIP }}
                </p>
              </div>
              <div >公网访问地址</div>
            </div>
          </div>

          <!-- 系统信息图表部分 -->
          <div class="md:">
            <!-- CPU使用率图表（示例） -->
            <div
              class="border duration-200"
            >
              <div >
                <h3 >CPU 使用率</h3>
                <span >最近24小时</span>
              </div>
              <div >
                <div
                  v-for="i in 24"
                  :key="i"
                  class="%]"
                  :style="`height: ${Math.floor(Math.random() * 80 + 10)}%`"
                ></div>
              </div>
              <div >
                <span>24h前</span>
                <span >18h前</span>
                <span >12h前</span>
                <span >现在</span>
              </div>
            </div>

            <!-- 内存使用图表（示例） -->
            <div
              class="border duration-200"
            >
              <div >
                <h3 >内存使用情况</h3>
                <span >68% 可用</span>
              </div>
              <div >
                <div class="inset-0">
                  <svg >
                    <circle
                      cx="64"
                      cy="64"
                      r="60"
                      fill="none"
                      stroke="#1f2937"
                      stroke-width="8"
                    />
                    <circle
                      cx="64"
                      cy="64"
                      r="60"
                      fill="none"
                      stroke="#3b82f6"
                      stroke-width="8"
                      stroke-dasharray="377"
                      stroke-dashoffset="120"
                      stroke-linecap="round"
                    />
                  </svg>
                </div>
                <div >
                  <p >68%</p>
                  <p >可用内存</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 加载状态 -->
        <div
          v-else
          
        >
          <svg
            
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <p>获取系统信息中...</p>
          <p >这可能需要几秒钟时间</p>
        </div>
      </div>
    </div>

    <FooterPage />

    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />
  </div>
</template>

<script>
import { ref, onMounted } from "vue";
import api from "../../api/axiosInstance";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import { useNotification } from "../../composables/useNotification";

export default {
  name: "SystemStatus",
  components: {
    HeaderPage,
    FooterPage,
    PopupNotification,
  },
  setup() {
    const systemInfo = ref(null);

    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    const systemInfoCards = {
      currentDirectory: { title: "程序运行目录", key: "currentDirectory" },
      localIP: { title: "本机 IP", key: "localIP" },
      publicIP: { title: "外网 IP", key: "publicIP" },
      kernelVersion: { title: "系统内核", key: "kernelVersion" },
      osDistribution: { title: "系统版本", key: "osDistribution" },
      privileges: { title: "运行权限", key: "privileges" },
    };

    const fetchSystemInfo = async () => {
      // 设置为null以显示加载状态
      systemInfo.value = null;

      try {
        const response = await api.get("/system/info");
        if (response.data?.data?.systemInfo) {
          // 延迟一点显示，让加载动画看起来更自然
          setTimeout(() => {
            systemInfo.value = response.data.data.systemInfo;
            showSuccess("系统信息已更新");
          }, 600);
        }
      } catch (error) {
        showError("获取系统信息失败");
      }
    };

    const copyToClipboard = (text) => {
      navigator.clipboard
        .writeText(text)
        .then(() => showSuccess("已复制到剪贴板"))
        .catch(() => showError("复制失败"));
    };

    onMounted(() => {
      fetchSystemInfo();
    });

    return {
      systemInfo,
      systemInfoCards,
      fetchSystemInfo,
      copyToClipboard,
      showNotification,
      notificationMessage,
      notificationType,
    };
  },
};
</script>

<style scoped>
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 信息卡片通用样式 */
.info-card {
  @apply bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 rounded-xl p-6 transition-all duration-200 hover:shadow-lg hover:border-gray-600/40;
}

/* 优化按钮点击效果 */
button:active {
  transform: scale(0.98);
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
</style>
