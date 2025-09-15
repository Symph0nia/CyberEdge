<template>
  <div class="bg-gray-900 text-white flex flex-col min-h-screen">
    <HeaderPage />

    <!-- 主体内容 -->
    <div class="container mx-auto px-6 py-8 flex-1 mt-16">
      <!-- 卡片式设计 -->
      <div
        class="bg-gray-800/40 backdrop-blur-xl p-8 rounded-2xl shadow-2xl border border-gray-700/30"
      >
        <!-- 顶部操作栏 - 重新设计为两行布局 -->
        <div class="mb-8">
          <div class="flex justify-between items-center mb-6">
            <div class="flex items-center">
              <h2
                class="text-xl font-medium tracking-wide text-gray-200 flex items-center"
              >
                <i class="ri-focus-3-line mr-2"></i>
                目标管理
              </h2>
              <span
                class="ml-4 px-3 py-1.5 rounded-xl bg-gray-700/50 text-gray-200 text-sm"
              >
                共 {{ filteredTargets.length }} 个目标
              </span>
            </div>

            <div class="flex space-x-4">
              <button
                @click="openCreateDialog"
                class="px-4 py-2.5 rounded-xl text-sm font-medium bg-blue-600/70 hover:bg-blue-500/70 text-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 flex items-center shadow-lg shadow-blue-900/20"
              >
                <i class="ri-add-line mr-2"></i>
                新建目标
              </button>

              <button
                @click="refreshTargets"
                class="control-button"
                :disabled="isLoading"
                :class="{ 'opacity-70 cursor-not-allowed': isLoading }"
              >
                <i
                  class="ri-refresh-line mr-2"
                  :class="{ 'animate-spin': isLoading }"
                ></i>
                {{ isLoading ? "加载中..." : "刷新" }}
              </button>
            </div>
          </div>

          <!-- 搜索和过滤栏 - 重新设计 -->
          <div class="flex flex-col md:flex-row md:items-center gap-4">
            <!-- 搜索框 -->
            <div class="flex-1 relative group">
              <i
                class="ri-search-line absolute left-4 top-3 text-gray-400 group-focus-within:text-blue-400 transition-colors duration-200"
              ></i>
              <input
                v-model="searchQuery"
                type="text"
                placeholder="搜索目标名称、地址或描述..."
                class="w-full pl-10 pr-4 py-2.5 rounded-xl bg-gray-700/50 text-gray-100 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all duration-200"
              />
            </div>

            <!-- 筛选器 - 按钮式设计 -->
            <div class="flex space-x-2">
              <button
                @click="statusFilter = ''"
                class="filter-button"
                :class="{ 'filter-active': statusFilter === '' }"
              >
                全部
              </button>
              <button
                @click="statusFilter = 'active'"
                class="filter-button"
                :class="{ 'filter-active': statusFilter === 'active' }"
              >
                <span class="w-2 h-2 rounded-full bg-emerald-400 mr-2"></span>
                活跃
              </button>
              <button
                @click="statusFilter = 'archived'"
                class="filter-button"
                :class="{ 'filter-active': statusFilter === 'archived' }"
              >
                <span class="w-2 h-2 rounded-full bg-gray-400 mr-2"></span>
                已归档
              </button>
            </div>
          </div>
        </div>

        <!-- 视图切换按钮组 -->
        <div class="flex justify-end mb-6">
          <div class="bg-gray-700/50 rounded-lg p-1 flex">
            <button
              @click="viewMode = 'card'"
              class="px-3 py-1.5 rounded-md flex items-center text-sm transition-all duration-200"
              :class="
                viewMode === 'card'
                  ? 'bg-gray-600 text-white'
                  : 'text-gray-400 hover:text-gray-200'
              "
            >
              <i class="ri-layout-grid-fill mr-1.5"></i>
              卡片视图
            </button>
            <button
              @click="viewMode = 'table'"
              class="px-3 py-1.5 rounded-md flex items-center text-sm transition-all duration-200"
              :class="
                viewMode === 'table'
                  ? 'bg-gray-600 text-white'
                  : 'text-gray-400 hover:text-gray-200'
              "
            >
              <i class="ri-table-fill mr-1.5"></i>
              表格视图
            </button>
          </div>
        </div>

        <!-- 卡片视图 -->
        <div
          v-if="viewMode === 'card'"
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
        >
          <div
            v-for="target in filteredTargets"
            :key="target.id"
            class="bg-gray-750 rounded-xl border border-gray-700/50 hover:border-blue-500/30 transition-all duration-300 overflow-hidden shadow-lg hover:shadow-xl transform hover:-translate-y-1"
          >
            <!-- 卡片头部 - 状态指示和名称 -->
            <div class="p-5 border-b border-gray-700/30">
              <div class="flex justify-between items-start">
                <div>
                  <h3 class="text-lg font-medium text-gray-200 mb-1">
                    {{ target.name }}
                  </h3>
                  <div class="text-gray-400 text-xs truncate max-w-xs">
                    {{ target.target }}
                  </div>
                </div>

                <div
                  class="status-indicator"
                  :class="target.status === 'active' ? 'active' : 'archived'"
                >
                  {{ target.status === "active" ? "活跃" : "已归档" }}
                </div>
              </div>
            </div>

            <!-- 卡片内容 -->
            <div class="p-5 text-sm space-y-4">
              <!-- 描述信息 -->
              <p class="text-gray-400 line-clamp-2 min-h-[40px]">
                {{ target.description || "无描述信息" }}
              </p>

              <!-- 创建/更新时间信息 -->
              <div class="flex justify-between text-xs text-gray-500">
                <div class="flex items-center">
                  <i class="ri-time-line mr-1"></i>
                  创建: {{ formatDate(target.createdAt) }}
                </div>
                <div class="flex items-center">
                  <i class="ri-refresh-line mr-1"></i>
                  {{
                    target.updatedAt ? formatDate(target.updatedAt) : "未扫描"
                  }}
                </div>
              </div>
            </div>

            <!-- 操作按钮区 -->
            <div
              class="grid grid-cols-3 divide-x divide-gray-700/30 border-t border-gray-700/30"
            >
              <button @click="viewDetails(target)" class="card-action-button">
                <i class="ri-file-list-line mr-1"></i> 详情
              </button>

              <button
                @click="startScan(target)"
                class="card-action-button text-green-400"
              >
                <i class="ri-scan-line mr-1"></i> 扫描
              </button>

              <div class="relative group">
                <button class="card-action-button text-gray-400 w-full">
                  <i class="ri-more-2-fill"></i>
                </button>

                <!-- 更多选项下拉菜单 -->
                <div class="dropdown-menu origin-bottom-right">
                  <button @click="editTarget(target)" class="dropdown-item">
                    <i class="ri-edit-line mr-2 text-blue-400"></i> 编辑
                  </button>

                  <button @click="archiveTarget(target)" class="dropdown-item">
                    <i class="ri-archive-line mr-2 text-yellow-400"></i>
                    {{ target.status === "active" ? "归档" : "激活" }}
                  </button>

                  <button
                    @click="deleteTarget(target)"
                    class="dropdown-item text-red-400 border-t border-gray-700/30"
                  >
                    <i class="ri-delete-bin-line mr-2"></i> 删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 表格视图 -->
        <div v-else-if="viewMode === 'table'" class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="bg-gray-750">
                <th class="table-header">目标名称</th>
                <th class="table-header">目标地址</th>
                <th class="table-header">状态</th>
                <th class="table-header">创建时间</th>
                <th class="table-header">上次更新</th>
                <th class="table-header text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="target in filteredTargets"
                :key="target.id"
                class="border-b border-gray-700/20 hover:bg-gray-800/30 transition-colors duration-150"
              >
                <td class="table-cell font-medium text-white">
                  {{ target.name }}
                </td>
                <td class="table-cell text-gray-400">{{ target.target }}</td>
                <td class="table-cell">
                  <span
                    class="status-indicator"
                    :class="target.status === 'active' ? 'active' : 'archived'"
                  >
                    {{ target.status === "active" ? "活跃" : "已归档" }}
                  </span>
                </td>
                <td class="table-cell text-gray-400 text-sm">
                  {{ formatDate(target.createdAt) }}
                </td>
                <td class="table-cell text-gray-400 text-sm">
                  {{
                    target.updatedAt ? formatDate(target.updatedAt) : "未扫描"
                  }}
                </td>
                <td class="table-cell text-right space-x-2">
                  <button
                    @click="viewDetails(target)"
                    class="table-action-button bg-gray-700 hover:bg-gray-600"
                    title="查看详情"
                  >
                    <i class="ri-file-list-line"></i>
                  </button>

                  <button
                    @click="startScan(target)"
                    class="table-action-button bg-green-700/50 hover:bg-green-600/50 text-green-100"
                    title="开始扫描"
                  >
                    <i class="ri-scan-line"></i>
                  </button>

                  <button
                    @click="editTarget(target)"
                    class="table-action-button bg-blue-700/50 hover:bg-blue-600/50 text-blue-100"
                    title="编辑目标"
                  >
                    <i class="ri-edit-line"></i>
                  </button>

                  <button
                    @click="archiveTarget(target)"
                    class="table-action-button bg-yellow-700/50 hover:bg-yellow-600/50 text-yellow-100"
                    title="归档/激活"
                  >
                    <i class="ri-archive-line"></i>
                  </button>

                  <button
                    @click="deleteTarget(target)"
                    class="table-action-button bg-red-700/50 hover:bg-red-600/50 text-red-100"
                    title="删除目标"
                  >
                    <i class="ri-delete-bin-line"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 空状态展示 - 重新设计 -->
        <div
          v-if="filteredTargets.length === 0"
          class="flex flex-col items-center justify-center py-16 my-4 bg-gray-800/20 backdrop-blur-xl rounded-2xl border border-dashed border-gray-700/30"
        >
          <div class="empty-illustration mb-6">
            <i class="ri-radar-line text-5xl text-gray-700 animate-pulse"></i>
          </div>
          <span class="text-xl text-gray-300 mb-3">暂无目标数据</span>
          <p class="text-gray-500 mb-6 text-center max-w-md">
            创建你的第一个目标开始使用系统的全部功能
          </p>
          <button
            @click="openCreateDialog"
            class="px-6 py-3 rounded-xl text-sm font-medium bg-blue-600/70 hover:bg-blue-500/70 text-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 flex items-center shadow-lg"
          >
            <i class="ri-add-line mr-2"></i>
            创建第一个目标
          </button>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 创建/编辑目标对话框 -->
    <DialogModal
      v-if="showDialog"
      :title="dialogMode === 'create' ? '新建目标' : '编辑目标'"
      @close="showDialog = false"
      class="bg-gray-800/40 backdrop-blur-xl rounded-2xl border border-gray-700/30"
    >
      <TargetFormContent
        :initial-data="targetForm"
        :is-submitting="isSubmitting"
        @submit="submitTargetForm"
        @cancel="showDialog = false"
      />
    </DialogModal>

    <!-- 通知组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
    />

    <!-- 确认对话框 -->
    <ConfirmDialog
      :show="showConfirmDialog"
      :title="dialogTitle"
      :message="dialogMessage"
      :type="dialogType"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { useTargetManagement } from "@/composables/useTargetManagement";
import HeaderPage from "@/components/HeaderPage.vue";
import DialogModal from "@/components/Target/DialogModal.vue";
import PopupNotification from "@/components/Utils/PopupNotification.vue";
import ConfirmDialog from "@/components/Utils/ConfirmDialog.vue";
import FooterPage from "@/components/FooterPage.vue";
import TargetFormContent from "@/components/Target/TargetFormContent.vue";

// 组合式函数 - 目标管理核心逻辑
const {
  targets,
  isLoading,
  isSubmitting,
  targetForm,
  dialogMode,
  showDialog,
  showNotification,
  showConfirmDialog,
  notificationMessage,
  notificationType,
  dialogTitle,
  dialogMessage,
  dialogType,

  fetchTargets,
  openCreateDialog,
  editTarget,
  deleteTarget,
  archiveTarget,
  startScan,
  submitTargetForm,
  handleConfirm,
  handleCancel,
  viewDetails,
} = useTargetManagement();

// 本地状态
const searchQuery = ref("");
const statusFilter = ref("");
const viewMode = ref("card"); // 默认视图模式: card 或 table

// 过滤后的目标列表
const filteredTargets = computed(() => {
  if (!targets.value || !Array.isArray(targets.value)) return [];

  let filtered = [...targets.value];

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (target) =>
        target.name.toLowerCase().includes(query) ||
        target.target.toLowerCase().includes(query) ||
        target.description?.toLowerCase().includes(query)
    );
  }

  // 状态过滤
  if (statusFilter.value) {
    filtered = filtered.filter(
      (target) => target.status === statusFilter.value
    );
  }

  // 按创建时间排序（最新的在前）
  filtered.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

  return filtered;
});

// 刷新目标列表
const refreshTargets = async () => {
  try {
    await fetchTargets();
  } catch (error) {
    console.error("Failed to fetch targets:", error);
  }
};

// 格式化日期 - 更简洁的显示
const formatDate = (date) => {
  return new Date(date).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};
</script>

<style scoped>
/* 基础样式 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 控制按钮样式 */
.control-button {
  @apply px-4 py-2.5 rounded-xl text-sm font-medium bg-gray-700/50 hover:bg-gray-600/50 text-gray-200 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-600/50 flex items-center;
}

/* 过滤按钮样式 */
.filter-button {
  @apply px-3 py-2 text-sm rounded-lg flex items-center text-gray-400 bg-gray-800/30 border border-gray-700/30 hover:bg-gray-700/50 hover:text-gray-200 transition-all duration-200;
}

/* 激活的过滤按钮 */
.filter-active {
  @apply text-gray-100 bg-gray-700/70 border-gray-600/50;
}

/* 状态指示器 */
.status-indicator {
  @apply inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium transition-all duration-200;
}

.status-indicator.active {
  @apply bg-green-900/30 text-green-300 border border-green-700/30;
}

.status-indicator.archived {
  @apply bg-gray-700/30 text-gray-400 border border-gray-600/30;
}

/* 卡片操作按钮 */
.card-action-button {
  @apply py-3 text-sm flex items-center justify-center hover:bg-gray-700/50 transition-all duration-200;
}

/* 下拉菜单 */
.dropdown-menu {
  @apply absolute right-0 bottom-full mb-1 bg-gray-800 rounded-lg shadow-xl border border-gray-700/50 min-w-[140px] invisible opacity-0 transform scale-95 transition-all duration-200 z-10;
}

/* 显示下拉菜单 */
.group:hover .dropdown-menu {
  @apply visible opacity-100 transform scale-100;
}

/* 下拉菜单项 */
.dropdown-item {
  @apply flex items-center w-full px-4 py-2.5 text-sm text-left text-gray-300 hover:bg-gray-700/50 transition-all duration-200 first:rounded-t-lg last:rounded-b-lg;
}

/* 表格样式 */
.table-header {
  @apply py-3 px-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider;
}

.table-cell {
  @apply py-3 px-4 text-sm whitespace-nowrap;
}

/* 表格操作按钮 */
.table-action-button {
  @apply inline-flex items-center justify-center w-8 h-8 rounded-lg text-sm transition-all duration-200 focus:outline-none;
}

/* 背景色 - 稍深一点 */
.bg-gray-750 {
  @apply bg-gray-800/60;
}

/* 空状态插图 */
.empty-illustration {
  @apply flex items-center justify-center w-24 h-24 rounded-full bg-gray-800/30 border border-gray-700/30;
}
</style>
