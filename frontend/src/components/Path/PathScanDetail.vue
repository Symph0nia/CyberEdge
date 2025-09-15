<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 主要内容卡片 -->
      <div
        class="border"
      >
        <!-- 返回按钮与面包屑导航 -->
        <div >
          <router-link
            to="/path-scan-results"
            class="hover:"
          >
            <i class="ri-arrow-"></i>
            返回列表
          </router-link>
          <i class="ri-arrow- mx-2"></i>
          <span >路径扫描详情</span>
        </div>

        <!-- 标题和基本信息卡片 -->
        <div
          class="border overflow-"
        >
          <div
            
          >
            <div
              
            >
              <i class="ri-folders-line"></i>
            </div>
            <div>
              <h2 >
                {{ scanResult?.target || "加载中..." }}
              </h2>
              <p >路径扫描结果详情</p>
            </div>
          </div>

          <div class="md:">
            <div >
              <span >
                <i class="ri-fingerprint-line .5"></i>
                扫描ID
              </span>
              <div >
                <span >{{
                  scanResult?.id || "-"
                }}</span>
                <button
                  v-if="scanResult?.id"
                  @click="copyToClipboard(scanResult.id)"
                  class="hover:"
                  title="复制ID"
                >
                  <i class="ri-clipboard-line"></i>
                </button>
              </div>
            </div>

            <div >
              <span >
                <i class="ri-global-line .5"></i>
                目标地址
              </span>
              <div >
                <span >{{
                  scanResult?.target || "-"
                }}</span>
                <button
                  v-if="scanResult?.target"
                  @click="copyToClipboard(scanResult.target)"
                  class="hover:"
                  title="复制目标"
                >
                  <i class="ri-clipboard-line"></i>
                </button>
              </div>
            </div>

            <div >
              <span >
                <i class="ri-time-line .5"></i>
                扫描时间
              </span>
              <div >
                <span >
                  {{ scanResult ? formatDate(scanResult.timestamp) : "-" }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 统计数据展示 -->
        <div class="md:">
          <div
            class="border"
          >
            <span >总路径数</span>
            <span >{{
              paths.length
            }}</span>
          </div>

          <div
            class="border"
          >
            <span >可访问路径</span>
            <span >
              {{ paths.filter((p) => p.status === "200").length }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >重定向路径</span>
            <span >
              {{
                paths.filter((p) => p.status && p.status.startsWith("3")).length
              }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >已读状态</span>
            <span >
              {{ paths.filter((p) => p.is_read).length }}
            </span>
          </div>
        </div>

        <!-- 批量操作工具栏 -->
        <div
          class="border"
        >
          <div >
            <div >
              <input
                type="checkbox"
                @change="toggleSelectAll"
                v-model="selectAll"
                
                id="select-all"
              />
              <label
                for="select-all"
                
              >
                全选
              </label>
            </div>

            <span  v-if="selectedPaths.length > 0">
              已选择 {{ selectedPaths.length }} 项
            </span>

            <!-- 批量操作按钮组 -->
            <div >
              <!-- 解析路径按钮 -->
              <button
                @click="resolveSelectedPaths"
                :disabled="selectedPaths.length === 0 || isResolving"
                class="tool-button hover:"
                :class="{ ' ': selectedPaths.length === 0 || isResolving, }"
              >
                <i class="ri-search-eye-line .5"></i>
                {{ isResolving ? "正在解析..." : "解析选中路径" }}
              </button>

              <!-- 端口扫描按钮 -->
              <button
                @click="sendSelectedToPortScan"
                :disabled="selectedPaths.length === 0"
                class="tool-button hover:"
                :class="{ ' ': selectedPaths.length === 0, }"
              >
                <i class="ri-scan-2-line .5"></i>
                发送到端口扫描
              </button>
            </div>
          </div>
        </div>

        <!-- 路径表格 -->
        <div
          class="border overflow-"
        >
          <div class="custom-scrollbar">
            <table >
              <thead>
                <tr >
                  <th >
                    <span class="sr-only">选择</span>
                  </th>
                  <th
                    v-for="header in tableHeaders"
                    :key="header.key"
                    
                  >
                    {{ header.label }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(path, index) in paths"
                  :key="path.id"
                  class="duration-200"
                  :class="[ index % 2 === 0 ? '' : '', 'hover:', ]"
                >
                  <td >
                    <input
                      type="checkbox"
                      v-model="selectedPaths"
                      :value="path.id"
                      
                    />
                  </td>
                  <td >
                    {{ path.id }}
                  </td>
                  <td >
                    <div >
                      <i class="ri-folder-line"></i>
                      <span class="max-" :title="path.path">
                        {{ path.path }}
                      </span>
                      <button
                        @click="copyToClipboard(path.path)"
                        class="hover:"
                        title="复制路径"
                      >
                        <i class="ri-clipboard-line"></i>
                      </button>
                    </div>
                  </td>
                  <td >
                    <span
                      class="status-badge"
                      :class="getStatusClass(path.status)"
                    >
                      <i class="ri-checkbox-blank-circle-fill"></i>
                      {{ path.status || "未知" }}
                    </span>
                  </td>
                  <td >
                    <span
                      class="status-badge"
                      :class="path.is_read ? ' ' : ' '"
                    >
                      <i
                        :class="[ path.is_read ? 'ri-eye-line' : 'ri-eye-off-line', '', ]"
                      ></i>
                      {{ path.is_read ? "已读" : "未读" }}
                    </span>
                  </td>
                  <td >
                    <div >
                      <button
                        @click="toggleReadStatus(path)"
                        class="action-button"
                        :class="path.is_read ? ' ' : ' hover:'"
                      >
                        <i
                          :class="[ path.is_read ? 'ri-eye-off-line' : 'ri-eye-line', '', ]"
                        ></i>
                        {{ path.is_read ? "标为未读" : "标为已读" }}
                      </button>
                      <button
                        @click="sendToPortScan(path)"
                        class="action-button hover:"
                      >
                        <i class="ri-scan-2-line"></i>
                        端口扫描
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 空状态展示 -->
        <div
          v-if="paths.length === 0 && !errorMessage"
          
        >
          <div
            
          >
            <i class="ri-search-line"></i>
          </div>
          <h3 >无路径数据</h3>
          <p class="max-">
            该扫描结果中没有发现路径数据，或正在加载中...
          </p>
        </div>

        <!-- 错误提示 -->
        <div
          v-if="errorMessage"
          class="border"
        >
          <i class="ri-error-warning-line"></i>
          <p >{{ errorMessage }}</p>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 通知组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />

    <!-- 确认对话框 -->
    <ConfirmDialog
      :show="showDialog"
      :title="dialogTitle"
      :message="dialogMessage"
      :type="dialogType"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useRoute } from "vue-router";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import ConfirmDialog from "../Utils/ConfirmDialog.vue";
import { usePathScan } from "../../composables/usePathScan";

const route = useRoute();

// 表头配置
const tableHeaders = [
  { key: "id", label: "ID", width: "w-20" },
  { key: "path", label: "路径", width: "w-64" },
  { key: "status", label: "状态", width: "w-24" },
  { key: "is_read", label: "读取状态", width: "w-24" },
  { key: "actions", label: "操作", width: "w-48" },
];

// 从组合式函数中解构所需的状态和方法
const {
  // 状态数据
  scanResult,
  errorMessage,
  paths,
  selectedPaths,
  selectAll,
  isResolving,

  // UI状态
  showNotification,
  notificationMessage,
  notificationType,
  showDialog,
  dialogTitle,
  dialogMessage,
  dialogType,

  // 方法
  fetchScanResult,
  toggleSelectAll,
  toggleReadStatus,
  resolveSelectedPaths,
  sendToPortScan,
  sendSelectedToPortScan,
  copyToClipboard,
  handleConfirm,
  handleCancel,
} = usePathScan();

// 格式化日期函数
const formatDate = (timestamp) => {
  try {
    return new Date(timestamp).toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch (e) {
    return timestamp || "未知时间";
  }
};

// 获取状态样式类
const getStatusClass = (status) => {
  if (!status) return "bg-gray-500/20 text-gray-300 border-gray-500/30";

  if (status === "200") {
    return "bg-green-500/20 text-green-300 border-green-500/30";
  } else if (status.startsWith("3")) {
    return "bg-blue-500/20 text-blue-300 border-blue-500/30";
  } else if (status.startsWith("4")) {
    return "bg-yellow-500/20 text-yellow-300 border-yellow-500/30";
  } else if (status.startsWith("5")) {
    return "bg-red-500/20 text-red-300 border-red-500/30";
  }

  return "bg-gray-500/20 text-gray-300 border-gray-500/30";
};

// 在组件挂载时获取数据
onMounted(() => {
  fetchScanResult(route.params.id);
});
</script>

<style scoped>
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 自定义工具按钮 */
.tool-button {
  @apply px-4 py-2 rounded-lg text-xs font-medium
  transition-all duration-200 flex items-center
  focus:outline-none focus:ring-1 border;
}

/* 状态标签 */
.status-badge {
  @apply px-2 py-1 rounded-md text-xs font-medium
  whitespace-nowrap inline-flex items-center border;
}

/* 操作按钮 */
.action-button {
  @apply px-2 py-1 text-xs rounded-md flex items-center
  justify-center whitespace-nowrap transition-all duration-200
  border focus:outline-none;
}

/* 自定义滚动条 */
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.3) transparent;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}

/* 优化按钮点击效果 */
button:active:not(:disabled) {
  transform: scale(0.98);
}
</style>
