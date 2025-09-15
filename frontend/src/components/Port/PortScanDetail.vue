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
            to="/port-scan-results"
            class="hover:"
          >
            <i class="ri-arrow-"></i>
            返回列表
          </router-link>
          <i class="ri-arrow- mx-2"></i>
          <span >端口扫描详情</span>
        </div>

        <!-- 标题和基本信息卡片 -->
        <div
          class="border overflow-"
        >
          <div
            
          >
            <div
              
            >
              <i class="ri-scan-2-line"></i>
            </div>
            <div>
              <h2 >
                {{ scanResult?.target || "加载中..." }}
              </h2>
              <p >端口扫描结果详情</p>
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
            <span >总端口数</span>
            <span >{{
              filteredPorts.length
            }}</span>
          </div>

          <div
            class="border"
          >
            <span >开放端口</span>
            <span >
              {{
                filteredPorts.filter((p) => getPortValue(p, "state") === "open")
                  .length
              }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >已探测HTTP</span>
            <span >
              {{
                filteredPorts.filter((p) => getPortValue(p, "http_status"))
                  .length
              }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >已读状态</span>
            <span >
              {{
                filteredPorts.filter((p) => getPortValue(p, "is_read")).length
              }}
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

            <span  v-if="selectedPorts.length > 0">
              已选择 {{ selectedPorts.length }} 项
            </span>

            <!-- 批量操作按钮组 -->
            <div >
              <!-- HTTP探测按钮 -->
              <button
                @click="probePort(selectedPorts)"
                :disabled="selectedPorts.length === 0 || isProbing"
                class="tool-button hover:"
                :class="{ ' ': selectedPorts.length === 0 || isProbing, }"
              >
                <i class="ri-search-eye-line .5"></i>
                {{ isProbing ? "正在探测..." : "HTTPX探测" }}
              </button>

              <!-- 路径扫描按钮 -->
              <button
                @click="sendToPathScan(selectedPorts)"
                :disabled="selectedPorts.length === 0"
                class="tool-button hover:"
                :class="{ ' ': selectedPorts.length === 0, }"
              >
                <i class="ri-folders-line .5"></i>
                发送到路径扫描
              </button>
            </div>
          </div>
        </div>

        <!-- 端口信息表格 -->
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
                    :key="header"
                    
                  >
                    {{ header }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(port, index) in filteredPorts"
                  :key="getPortValue(port, '_id')"
                  class="duration-200"
                  :class="[ index % 2 === 0 ? '' : '', 'hover:', ]"
                >
                  <td >
                    <input
                      type="checkbox"
                      v-model="selectedPorts"
                      :value="getPortValue(port, '_id')"
                      
                    />
                  </td>
                  <td >
                    {{ getPortValue(port, "_id") }}
                  </td>
                  <td >
                    <span >
                      <i class="ri-door-lock-line"></i>
                      {{ getPortValue(port, "number") }}
                    </span>
                  </td>
                  <td >
                    {{ getPortValue(port, "protocol") }}
                  </td>
                  <td >
                    {{ getPortValue(port, "service") }}
                  </td>
                  <td >
                    <span
                      class="status-badge"
                      :class="{ ' ': getPortValue(port, 'state') === 'open', ' ': getPortValue(port, 'state') === 'closed', ' ': getPortValue(port, 'state') === 'filtered', }"
                    >
                      <i class="ri-checkbox-blank-circle-fill"></i>
                      {{ getPortValue(port, "state") }}
                    </span>
                  </td>
                  <td >
                    <div
                      v-if="getPortValue(port, 'http_status')"
                      
                    >
                      <span
                        :class="[ 'status-badge', getHttpStatusClass(getPortValue(port, 'http_status')), ]"
                      >
                        <i class="ri-earth-line"></i>
                        {{ getPortValue(port, "http_status") }}
                      </span>
                    </div>
                    <button
                      v-else
                      @click="probePort([getPortValue(port, '_id')])"
                      class="status-button"
                    >
                      <i class="ri-search-eye-line"></i>
                      探测
                    </button>
                  </td>
                  <td >
                    <div
                      
                      :title="getPortValue(port, 'http_title')"
                    >
                      {{ getPortValue(port, "http_title") || "-" }}
                    </div>
                  </td>
                  <td >
                    <span
                      class="status-badge"
                      :class="getPortValue(port, 'is_read') ? ' ' : ' '"
                    >
                      <i
                        :class="[ getPortValue(port, 'is_read') ? 'ri-eye-line' : 'ri-eye-off-line', '', ]"
                      ></i>
                      {{ getPortValue(port, "is_read") ? "已读" : "未读" }}
                    </span>
                  </td>
                  <td >
                    <div >
                      <button
                        @click="toggleReadStatus(port)"
                        class="action-button"
                        :class="getPortValue(port, 'is_read') ? ' ' : ' hover:'"
                      >
                        <i
                          :class="[ getPortValue(port, 'is_read') ? 'ri-eye-off-line' : 'ri-eye-line', ]"
                        ></i>
                        {{
                          getPortValue(port, "is_read")
                            ? "标为未读"
                            : "标为已读"
                        }}
                      </button>
                      <button
                        @click="sendToPathScan([getPortValue(port, '_id')])"
                        class="action-button hover:"
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
          
        >
          <div
            
          >
            <i class="ri-search-line"></i>
          </div>
          <h3 >无端口数据</h3>
          <p class="max-">
            该扫描结果中没有发现端口数据，或正在加载中...
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
