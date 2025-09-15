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
            to="/port-scan-results"
            class="hover:text-blue-400 transition-colors flex items-center"
          >
            <i class="ri-arrow-left-line mr-1"></i>
            返回列表
          </router-link>
          <i class="ri-arrow-right-s-line mx-2"></i>
          <span class="text-gray-200">端口扫描详情</span>
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
              <i class="ri-scan-2-line text-blue-400 text-xl"></i>
            </div>
            <div>
              <h2 class="text-xl font-medium tracking-wide text-gray-200">
                {{ scanResult?.target || "加载中..." }}
              </h2>
              <p class="text-sm text-gray-400 mt-1">端口扫描结果详情</p>
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
            <span class="text-sm text-gray-400 mb-1">总端口数</span>
            <span class="text-2xl font-medium text-gray-200">{{
              filteredPorts.length
            }}</span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">开放端口</span>
            <span class="text-2xl font-medium text-green-300">
              {{
                filteredPorts.filter((p) => getPortValue(p, "state") === "open")
                  .length
              }}
            </span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">已探测HTTP</span>
            <span class="text-2xl font-medium text-purple-300">
              {{
                filteredPorts.filter((p) => getPortValue(p, "http_status"))
                  .length
              }}
            </span>
          </div>

          <div
            class="bg-gray-800/60 rounded-xl border border-gray-700/30 p-4 flex flex-col"
          >
            <span class="text-sm text-gray-400 mb-1">已读状态</span>
            <span class="text-2xl font-medium text-blue-300">
              {{
                filteredPorts.filter((p) => getPortValue(p, "is_read")).length
              }}
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

            <span class="text-sm text-gray-400" v-if="selectedPorts.length > 0">
              已选择 {{ selectedPorts.length }} 项
            </span>

            <!-- 批量操作按钮组 -->
            <div class="flex flex-wrap gap-3 ml-auto">
              <!-- HTTP探测按钮 -->
              <button
                @click="probePort(selectedPorts)"
                :disabled="selectedPorts.length === 0 || isProbing"
                class="tool-button bg-purple-500/20 border-purple-500/30 text-purple-300 hover:bg-purple-500/30"
                :class="{
                  'opacity-60 cursor-not-allowed':
                    selectedPorts.length === 0 || isProbing,
                }"
              >
                <i class="ri-search-eye-line mr-1.5"></i>
                {{ isProbing ? "正在探测..." : "HTTPX探测" }}
              </button>

              <!-- 路径扫描按钮 -->
              <button
                @click="sendToPathScan(selectedPorts)"
                :disabled="selectedPorts.length === 0"
                class="tool-button bg-blue-500/20 border-blue-500/30 text-blue-300 hover:bg-blue-500/30"
                :class="{
                  'opacity-60 cursor-not-allowed': selectedPorts.length === 0,
                }"
              >
                <i class="ri-folders-line mr-1.5"></i>
                发送到路径扫描
              </button>
            </div>
          </div>
        </div>

        <!-- 端口信息表格 -->
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
                    :key="header"
                    class="py-3 px-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider"
                  >
                    {{ header }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(port, index) in filteredPorts"
                  :key="getPortValue(port, '_id')"
                  class="border-t border-gray-700/30 transition-colors duration-200"
                  :class="[
                    index % 2 === 0 ? 'bg-gray-800/30' : 'bg-transparent',
                    'hover:bg-gray-700/40',
                  ]"
                >
                  <td class="py-3 px-4 w-10">
                    <input
                      type="checkbox"
                      v-model="selectedPorts"
                      :value="getPortValue(port, '_id')"
                      class="rounded border-gray-700/50 bg-gray-800/50 text-blue-500 focus:ring-blue-500/30"
                    />
                  </td>
                  <td class="py-3 px-4 text-xs font-mono text-gray-300 w-20">
                    {{ getPortValue(port, "_id") }}
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-200 w-16">
                    <span class="flex items-center">
                      <i class="ri-door-lock-line mr-2 text-blue-400"></i>
                      {{ getPortValue(port, "number") }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-200 w-20">
                    {{ getPortValue(port, "protocol") }}
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-200 w-24">
                    {{ getPortValue(port, "service") }}
                  </td>
                  <td class="py-3 px-4 w-20">
                    <span
                      class="status-badge"
                      :class="{
                        'bg-green-500/20 text-green-300 border-green-500/30':
                          getPortValue(port, 'state') === 'open',
                        'bg-red-500/20 text-red-300 border-red-500/30':
                          getPortValue(port, 'state') === 'closed',
                        'bg-yellow-500/20 text-yellow-300 border-yellow-500/30':
                          getPortValue(port, 'state') === 'filtered',
                      }"
                    >
                      <i class="ri-checkbox-blank-circle-fill mr-1 text-xs"></i>
                      {{ getPortValue(port, "state") }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-sm w-24">
                    <div
                      v-if="getPortValue(port, 'http_status')"
                      class="flex items-center"
                    >
                      <span
                        :class="[
                          'status-badge',
                          getHttpStatusClass(getPortValue(port, 'http_status')),
                        ]"
                      >
                        <i class="ri-earth-line mr-1"></i>
                        {{ getPortValue(port, "http_status") }}
                      </span>
                    </div>
                    <button
                      v-else
                      @click="probePort([getPortValue(port, '_id')])"
                      class="status-button bg-purple-500/20 text-purple-300 border-purple-500/30"
                    >
                      <i class="ri-search-eye-line mr-1"></i>
                      探测
                    </button>
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-200 w-48">
                    <div
                      class="truncate"
                      :title="getPortValue(port, 'http_title')"
                    >
                      {{ getPortValue(port, "http_title") || "-" }}
                    </div>
                  </td>
                  <td class="py-3 px-4 w-20">
                    <span
                      class="status-badge"
                      :class="
                        getPortValue(port, 'is_read')
                          ? 'bg-green-500/20 text-green-300 border-green-500/30'
                          : 'bg-yellow-500/20 text-yellow-300 border-yellow-500/30'
                      "
                    >
                      <i
                        :class="[
                          getPortValue(port, 'is_read')
                            ? 'ri-eye-line'
                            : 'ri-eye-off-line',
                          'mr-1',
                        ]"
                      ></i>
                      {{ getPortValue(port, "is_read") ? "已读" : "未读" }}
                    </span>
                  </td>
                  <td class="py-3 px-4 w-48">
                    <div class="flex gap-2">
                      <button
                        @click="toggleReadStatus(port)"
                        class="action-button"
                        :class="
                          getPortValue(port, 'is_read')
                            ? 'bg-gray-700/50 text-gray-300 border-gray-600/30'
                            : 'bg-green-500/20 text-green-300 border-green-500/30 hover:bg-green-500/30'
                        "
                      >
                        <i
                          :class="[
                            getPortValue(port, 'is_read')
                              ? 'ri-eye-off-line'
                              : 'ri-eye-line',
                          ]"
                        ></i>
                        {{
                          getPortValue(port, "is_read")
                            ? "标为未读"
                            : "标为已读"
                        }}
                      </button>
                      <button
                        @click="sendToPathScan([getPortValue(port, '_id')])"
                        class="action-button bg-blue-500/20 text-blue-300 border-blue-500/30 hover:bg-blue-500/30"
                      >
                        <i class="ri-folders-line"></i>
                        路径扫描
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
          v-if="filteredPorts.length === 0 && !errorMessage"
          class="flex flex-col items-center justify-center py-12 text-center"
        >
          <div
            class="w-16 h-16 rounded-full bg-gray-800/50 flex items-center justify-center mb-4"
          >
            <i class="ri-search-line text-gray-500 text-3xl"></i>
          </div>
          <h3 class="text-lg font-medium text-gray-300 mb-2">无端口数据</h3>
          <p class="text-gray-400 max-w-md mb-6">
            该扫描结果中没有发现端口数据，或正在加载中...
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
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { usePortScan } from "../../composables/usePortScan";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import ConfirmDialog from "../Utils/ConfirmDialog.vue";

const route = useRoute();
const isProbing = ref(false); // 用于跟踪探测状态

// 表头配置
const tableHeaders = [
  "ID",
  "端口",
  "协议",
  "服务",
  "状态",
  "HTTP状态",
  "HTTP标题",
  "读取状态",
  "操作",
];

// 从组合式函数中解构所需的状态和方法
const {
  // 基础数据
  scanResult,
  errorMessage,
  selectedPorts,
  selectAll,
  filteredPorts,

  // 方法
  getPortValue,
  toggleReadStatus,
  toggleSelectAll,
  sendToPathScan,
  fetchScanResult,
  probePort,
  getHttpStatusClass,
  copyToClipboard,

  // 通知状态和方法
  showNotification,
  notificationMessage,
  notificationType,

  // 确认对话框状态和方法
  showDialog,
  dialogTitle,
  dialogMessage,
  dialogType,
  handleConfirm,
  handleCancel,
} = usePortScan();

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

/* 小型状态按钮 */
.status-button {
  @apply px-2 py-1 rounded-md text-xs font-medium
  transition-all duration-200 flex items-center justify-center
  whitespace-nowrap border focus:outline-none;
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
