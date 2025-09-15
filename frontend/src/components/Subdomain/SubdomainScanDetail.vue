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
            to="/subdomain-scan-results"
            class="hover:"
          >
            <i class="ri-arrow-"></i>
            返回列表
          </router-link>
          <i class="ri-arrow- mx-2"></i>
          <span >扫描详情</span>
        </div>

        <!-- 标题和基本信息卡片 -->
        <div
          class="border overflow-"
        >
          <div
            
          >
            <div
              
            >
              <i class="ri-radar-line"></i>
            </div>
            <div>
              <h2 >
                {{ scanResult?.target || "加载中..." }}
              </h2>
              <p >子域名扫描结果详情</p>
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
                  {{
                    scanResult
                      ? new Date(scanResult.timestamp).toLocaleString()
                      : "-"
                  }}
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
            <span >总子域名</span>
            <span >{{
              subdomains.length
            }}</span>
          </div>

          <div
            class="border"
          >
            <span >已解析IP</span>
            <span >
              {{ subdomains.filter((s) => s.ip).length }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >已探测HTTP</span>
            <span >
              {{ subdomains.filter((s) => s.httpStatus).length }}
            </span>
          </div>

          <div
            class="border"
          >
            <span >已读状态</span>
            <span >
              {{ subdomains.filter((s) => s.is_read).length }}
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

            <span
              
              v-if="selectedSubdomains.length > 0"
            >
              已选择 {{ selectedSubdomains.length }} 项
            </span>

            <!-- 批量操作按钮组 -->
            <div >
              <!-- IP解析按钮 -->
              <button
                @click="resolveIPs(selectedSubdomains)"
                :disabled="selectedSubdomains.length === 0 || isResolving"
                class="tool-button hover:"
                :class="{ ' ': selectedSubdomains.length === 0 || isResolving, }"
              >
                <i class="ri-radar-line .5"></i>
                {{ isResolving ? "正在解析..." : "解析选中IP" }}
              </button>

              <!-- HTTPX探测按钮 -->
              <button
                @click="probeHosts(selectedSubdomains)"
                :disabled="selectedSubdomains.length === 0 || isProbing"
                class="tool-button hover:"
                :class="{ ' ': selectedSubdomains.length === 0 || isProbing, }"
              >
                <i class="ri-search-eye-line .5"></i>
                {{ isProbing ? "正在探测..." : "HTTPX探测" }}
              </button>

              <!-- 端口扫描按钮 -->
              <button
                @click="sendToPortScan(selectedSubdomains)"
                :disabled="selectedSubdomains.length === 0"
                class="tool-button hover:"
                :class="{ ' ': selectedSubdomains.length === 0, }"
              >
                <i class="ri-scan-2-line .5"></i>
                发送到端口扫描
              </button>
            </div>
          </div>
        </div>

        <!-- 子域名表格 -->
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
                  v-for="(subdomain, index) in subdomains"
                  :key="subdomain.id"
                  class="duration-200"
                  :class="[ subdomain.isFirstIP ? '' : index % 2 === 0 ? '' : '', subdomain.isFirstIP ? '' : '', 'hover:', ]"
                >
                  <td >
                    <input
                      type="checkbox"
                      v-model="selectedSubdomains"
                      :value="subdomain.id"
                      
                    />
                  </td>
                  <td >
                    {{ subdomain.id }}
                  </td>
                  <td >
                    <div >
                      <a
                        :href="`http://${subdomain.domain}`"
                        target="_blank"
                        class="hover: max-"
                        :title="subdomain.domain"
                      >
                        {{ subdomain.domain }}
                      </a>
                      <button
                        @click="copyToClipboard(subdomain.domain)"
                        class="hover:"
                        title="复制域名"
                      >
                        <i class="ri-clipboard-line"></i>
                      </button>
                    </div>
                  </td>

                  <td >
                    <div v-if="subdomain.ip" >
                      <span
                        :class="[ '', subdomain.isFirstIP ? '' : '', ]"
                      >
                        {{ subdomain.ip }}
                      </span>
                      <button
                        @click="copyToClipboard(subdomain.ip)"
                        class="hover:"
                        title="复制IP"
                      >
                        <i class="ri-clipboard-line"></i>
                      </button>
                    </div>
                    <button
                      v-else
                      @click="resolveIPs([subdomain.id])"
                      class="status-button"
                    >
                      <i class="ri-radar-line"></i>
                      解析IP
                    </button>
                  </td>
                  <td >
                    <span
                      v-if="subdomain.httpStatus"
                      :class="[ 'status-badge', getHttpStatusClass(subdomain.httpStatus), ]"
                    >
                      {{ subdomain.httpStatus }}
                    </span>
                    <button
                      v-else
                      @click="probeHosts([subdomain.id])"
                      class="status-button"
                    >
                      <i class="ri-search-eye-line"></i>
                      探测
                    </button>
                  </td>
                  <td >
                    <div  :title="subdomain.httpTitle">
                      {{ subdomain.httpTitle || "-" }}
                    </div>
                  </td>
                  <td >
                    <span
                      class="status-badge"
                      :class="subdomain.is_read ? ' ' : ' '"
                    >
                      <i
                        :class="[ subdomain.is_read ? 'ri-eye-line' : 'ri-eye-off-line', '', ]"
                      ></i>
                      {{ subdomain.is_read ? "已读" : "未读" }}
                    </span>
                  </td>
                  <td >
                    <div >
                      <button
                        @click="toggleReadStatus(subdomain)"
                        class="action-button"
                        :class="subdomain.is_read ? ' ' : ' hover:'"
                      >
                        <i
                          :class="[ subdomain.is_read ? 'ri-eye-off-line' : 'ri-eye-line', ]"
                        ></i>
                        {{ subdomain.is_read ? "标为未读" : "标为已读" }}
                      </button>
                      <button
                        @click="sendToPortScan([subdomain.id])"
                        :disabled="!subdomain.ip"
                        class="action-button"
                        :class="[ subdomain.ip ? ' hover:' : ' ', ]"
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
          v-if="subdomains.length === 0 && !errorMessage"
          
        >
          <div
            
          >
            <i class="ri-search-line"></i>
          </div>
          <h3 >无子域名数据</h3>
          <p class="max-">
            该扫描结果中没有发现子域名数据，或正在加载中...
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
import { useSubdomainScan } from "../../composables/useSubdomainScan";

const route = useRoute();

// 表头配置
const tableHeaders = [
  "ID",
  "子域名",
  "IP地址",
  "HTTP状态",
  "标题",
  "状态",
  "操作",
];

// 从组合式函数中解构所需的状态和方法
const {
  // 状态数据
  scanResult,
  errorMessage,
  subdomains,
  selectedSubdomains,
  selectAll,
  isResolving,
  isProbing,

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
  resolveIPs,
  sendToPortScan,
  probeHosts,
  handleConfirm,
  handleCancel,
  getHttpStatusClass,
  copyToClipboard,
} = useSubdomainScan();

// 初始化加载
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

/* 隔行变色效果 */
tr:nth-child(even) {
  background-color: rgba(31, 41, 55, 0.3);
}
</style>
