<template>
  <div class="bg-gray-900 text-white flex flex-col min-h-screen">
    <HeaderPage />

    <div class="container mx-auto px-6 py-8 flex-1 mt-16">
      <!-- 主要内容卡片 -->
      <div
        class="bg-gray-800/40 backdrop-blur-xl p-8 rounded-2xl shadow-2xl border border-gray-700/30"
      >
        <!-- 返回按钮与面包屑导航 -->
        <div class="flex items-center text-sm text-gray-400 mb-6">
          <router-link
            to="/path-scan-results"
            class="hover:text-blue-400 transition-colors flex items-center"
          >
            <i class="ri-arrow-left-line mr-1"></i>
            返回列表
          </router-link>
          <i class="ri-arrow-right-s-line mx-2"></i>
          <span class="text-gray-200">路径扫描详情</span>
        </div>

        <!-- 标题和基本信息卡片 -->
        <div
          class="bg-gray-800/60 rounded-xl border border-gray-700/30 mb-6 overflow-hidden"
        >
          <div
            class="p-5 border-b border-gray-700/30 flex items-center space-x-3"
          >
            <div
              class="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center"
            >
              <i class="ri-folders-line text-blue-400 text-xl"></i>
            </div>
            <div>
              <h2 class="text-xl font-medium tracking-wide text-gray-200">
                {{ scanResult?.target || "加载中..." }}
              </h2>
              <p class="text-sm text-gray-400 mt-1">路径扫描结果详情</p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6 p-5">
            <div class="flex flex-col">
              <span class="text-sm text-gray-400 mb-1 flex items-center">
                <i class="ri-fingerprint-line mr-1.5"></i>
                扫描ID
              </span>
              <div class="flex items-center">
                <span class="text-sm font-mono text-gray-200">{{
                  scanResult?.id || "-"
                }}</span>
                <button
                  v-if="scanResult?.id"
                  @click="copyToClipboard(scanResult.id)"
                  class="ml-2 text-gray-500 hover:text-gray-300 transition-colors"
                  title="复制ID"
                >
                  <i class="ri-clipboard-line text-xs"></i>
                </button>
              </div>
            </div>

            <div class="flex flex-col">
              <span class="text-sm text-gray-400 mb-1 flex items-center">
                <i class="ri-global-line mr-1.5"></i>
                目标地址
              </span>
              <div class="flex items-center">
                <span class="text-sm text-gray-200">{{
                  scanResult?.target || "-"
                }}</span>
                <button
                  v-if="scanResult?.target"
                  @click="copyToClipboard(scanResult.target)"
                  class="ml-2 text-gray-500 hover:text-gray-300 transition-colors"
                  title="复制目标"
                >
                  <i class="ri-clipboard-line text-xs"></i>
                </button>
              </div>
            </div>

            <div class="flex flex-col">
              <span class="text-sm text-gray-400 mb-1 flex items-center">
                <i class="ri-time-line mr-1.5"></i>
                扫描时间
              </span>
              <div class="flex items-center">
                <span class="text-sm text-gray-200">
                  {{ scanResult ? formatDate(scanResult.timestamp) : "-" }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 统计数据展示 -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">总路径数</span>
            <span class="text-2xl font-medium text-gray-200">{{
              paths.length
            }}</span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">可访问路径</span>
            <span class="text-2xl font-medium text-green-300">
              {{ paths.filter((p) => p.status === "200").length }}
            </span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">重定向路径</span>
            <span class="text-2xl font-medium text-blue-300">
              {{
                paths.filter((p) => p.status && p.status.startsWith("3")).length
              }}
            </span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">已读状态</span>
            <span class="text-2xl font-medium text-yellow-300">
              {{ paths.filter((p) => p.is_read).length }}
            </span>
          </div>
        </div>

        <!-- 批量操作工具栏 -->
        <div
          class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 mb-6"
        >
          <div class="flex flex-wrap items-center gap-3">
            <div class="flex items-center mr-2">
              <input
                type="checkbox"
                @change="toggleSelectAll"
                v-model="selectAll"
                class="rounded border-gray-700/50 bg-gray-800/50 text-blue-500 focus:ring-blue-500/30 mr-2"
                id="select-all"
              />
              <label
                for="select-all"
                class="text-sm text-gray-300 cursor-pointer"
              >
                全选
              </label>
            </div>

            <span class="text-sm text-gray-400" v-if="selectedPaths.length > 0">
              已选择 {{ selectedPaths.length }} 项
            </span>

            <!-- 批量操作按钮组 -->
            <div class="flex flex-wrap gap-3 ml-auto">
              <!-- 解析路径按钮 -->
              <button
                @click="resolveSelectedPaths"
                :disabled="selectedPaths.length === 0 || isResolving"
                class="tool-button bg-blue-500/20 border-blue-500/30 text-blue-300 hover:bg-blue-500/30"
                :class="{
                  'opacity-60 cursor-not-allowed':
                    selectedPaths.length === 0 || isResolving,
                }"
              >
                <i class="ri-search-eye-line mr-1.5"></i>
                {{ isResolving ? "正在解析..." : "解析选中路径" }}
              </button>

              <!-- 端口扫描按钮 -->
              <button
                @click="sendSelectedToPortScan"
                :disabled="selectedPaths.length === 0"
                class="tool-button bg-yellow-500/20 border-yellow-500/30 text-yellow-300 hover:bg-yellow-500/30"
                :class="{
                  'opacity-60 cursor-not-allowed': selectedPaths.length === 0,
                }"
              >
                <i class="ri-scan-2-line mr-1.5"></i>
                发送到端口扫描
              </button>
            </div>
          </div>
        </div>

        <!-- 路径表格 -->
        <div
          class="bg-gray-800/60 rounded-xl border border-gray-700/30 overflow-hidden mb-4"
        >
          <div class="relative overflow-x-auto custom-scrollbar">
            <table class="w-full">
              <thead>
                <tr class="bg-gray-800/80">
                  <th class="py-3 px-4 text-left">
                    <span class="sr-only">选择</span>
                  </th>
                  <th
                    v-for="header in tableHeaders"
                    :key="header.key"
                    class="py-3 px-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider"
                  >
                    {{ header.label }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(path, index) in paths"
                  :key="path.id"
                  class="border-t border-gray-700/30 transition-colors duration-200"
                  :class="[
                    index % 2 === 0 ? 'bg-gray-800/30' : 'bg-transparent',
                    'hover:bg-gray-700/40',
                  ]"
                >
                  <td class="py-3 px-4 w-10">
                    <input
                      type="checkbox"
                      v-model="selectedPaths"
                      :value="path.id"
                      class="rounded border-gray-700/50 bg-gray-800/50 text-blue-500 focus:ring-blue-500/30"
                    />
                  </td>
                  <td class="py-3 px-4 text-xs font-mono text-gray-300 w-20">
                    {{ path.id }}
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-200">
                    <div class="flex items-center">
                      <i class="ri-folder-line mr-2 text-blue-400"></i>
                      <span class="truncate max-w-xs" :title="path.path">
                        {{ path.path }}
                      </span>
                      <button
                        @click="copyToClipboard(path.path)"
                        class="ml-2 text-gray-500 hover:text-gray-300 transition-colors"
                        title="复制路径"
                      >
                        <i class="ri-clipboard-line text-xs"></i>
                      </button>
                    </div>
                  </td>
                  <td class="py-3 px-4 w-24">
                    <span
                      class="status-badge"
                      :class="getStatusClass(path.status)"
                    >
                      <i class="ri-checkbox-blank-circle-fill mr-1 text-xs"></i>
                      {{ path.status || "未知" }}
                    </span>
                  </td>
                  <td class="py-3 px-4 w-20">
                    <span
                      class="status-badge"
                      :class="
                        path.is_read
                          ? 'bg-green-500/20 text-green-300 border-green-500/30'
                          : 'bg-yellow-500/20 text-yellow-300 border-yellow-500/30'
                      "
                    >
                      <i
                        :class="[
                          path.is_read ? 'ri-eye-line' : 'ri-eye-off-line',
                          'mr-1',
                        ]"
                      ></i>
                      {{ path.is_read ? "已读" : "未读" }}
                    </span>
                  </td>
                  <td class="py-3 px-4 w-48">
                    <div class="flex gap-2">
                      <button
                        @click="toggleReadStatus(path)"
                        class="action-button"
                        :class="
                          path.is_read
                            ? 'bg-gray-700/50 text-gray-300 border-gray-600/30'
                            : 'bg-green-500/20 text-green-300 border-green-500/30 hover:bg-green-500/30'
                        "
                      >
                        <i
                          :class="[
                            path.is_read ? 'ri-eye-off-line' : 'ri-eye-line',
                            'mr-1',
                          ]"
                        ></i>
                        {{ path.is_read ? "标为未读" : "标为已读" }}
                      </button>
                      <button
                        @click="sendToPortScan(path)"
                        class="action-button bg-yellow-500/20 text-yellow-300 border-yellow-500/30 hover:bg-yellow-500/30"
                      >
                        <i class="ri-scan-2-line mr-1"></i>
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
          class="flex flex-col items-center justify-center py-12 text-center"
        >
          <div
            class="w-16 h-16 rounded-full bg-gray-800/50 flex items-center justify-center mb-4"
          >
            <i class="ri-search-line text-gray-500 text-3xl"></i>
          </div>
          <h3 class="text-lg font-medium text-gray-300 mb-2">无路径数据</h3>
          <p class="text-gray-400 max-w-md mb-6">
            该扫描结果中没有发现路径数据，或正在加载中...
          </p>
        </div>

        <!-- 错误提示 -->
        <div
          v-if="errorMessage"
          class="mt-4 px-4 py-3 rounded-xl bg-red-500/20 border border-red-500/30 flex items-center"
        >
          <i class="ri-error-warning-line text-red-400 mr-2 text-lg"></i>
          <p class="text-sm text-red-400">{{ errorMessage }}</p>
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
