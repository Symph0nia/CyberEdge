<template>
  <div class="target-management">
    <HeaderPage />

    <div class="management-container">
      <a-card class="management-card">
        <!-- 页面头部 -->
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">
              <i class="ri-focus-3-line"></i>
              目标管理
            </h1>
            <a-tag color="blue" class="count-tag">
              共 {{ filteredTargets.length }} 个目标
            </a-tag>
          </div>

          <div class="header-actions">
            <a-button
              type="primary"
              @click="openCreateDialog"
              class="create-btn"
            >
              <i class="ri-add-line"></i>
              新建目标
            </a-button>

            <a-button
              @click="refreshTargets"
              :loading="isLoading"
              class="refresh-btn"
            >
              <i class="ri-refresh-line"></i>
              {{ isLoading ? "加载中" : "刷新" }}
            </a-button>
          </div>
        </div>

        <!-- 搜索和筛选 -->
        <div class="search-filters">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="搜索目标名称、地址或描述..."
            size="large"
            class="search-input"
          />

          <div class="filter-buttons">
            <a-radio-group v-model:value="statusFilter" button-style="solid" size="middle">
              <a-radio-button value="">全部</a-radio-button>
              <a-radio-button value="active">
                <span class="status-dot active"></span>
                活跃
              </a-radio-button>
              <a-radio-button value="archived">
                <span class="status-dot archived"></span>
                已归档
              </a-radio-button>
            </a-radio-group>
          </div>
        </div>

        <!-- 视图切换 -->
        <div class="view-controls">
          <a-radio-group v-model:value="viewMode" button-style="solid" size="small">
            <a-radio-button value="card">
              <i class="ri-layout-grid-fill"></i>
              卡片视图
            </a-radio-button>
            <a-radio-button value="table">
              <i class="ri-table-fill"></i>
              表格视图
            </a-radio-button>
          </a-radio-group>
        </div>

        <!-- 卡片视图 -->
        <div v-if="viewMode === 'card'" class="cards-container">
          <a-row :gutter="[16, 16]">
            <a-col
              v-for="target in filteredTargets"
              :key="target.id"
              :xs="24" :sm="12" :md="8" :lg="6"
            >
              <a-card
                size="small"
                class="target-card"
                :class="{ active: target.status === 'active' }"
              >
                <template #title>
                  <div class="card-title">
                    <span class="target-name">{{ target.name }}</span>
                    <a-tag :color="target.status === 'active' ? 'success' : 'default'" size="small">
                      {{ target.status === 'active' ? '活跃' : '已归档' }}
                    </a-tag>
                  </div>
                </template>

                <template #extra>
                  <a-dropdown trigger="click">
                    <a-button type="text" size="small">
                      <i class="ri-more-2-fill"></i>
                    </a-button>
                    <template #overlay>
                      <a-menu>
                        <a-menu-item @click="editTarget(target)">
                          <i class="ri-edit-line"></i>
                          编辑
                        </a-menu-item>
                        <a-menu-item @click="archiveTarget(target)">
                          <i class="ri-archive-line"></i>
                          {{ target.status === 'active' ? '归档' : '激活' }}
                        </a-menu-item>
                        <a-menu-divider />
                        <a-menu-item @click="deleteTarget(target)" class="danger-item">
                          <i class="ri-delete-bin-line"></i>
                          删除
                        </a-menu-item>
                      </a-menu>
                    </template>
                  </a-dropdown>
                </template>

                <div class="card-content">
                  <p class="target-url">{{ target.target }}</p>
                  <p class="target-desc">{{ target.description || '无描述信息' }}</p>

                  <div class="card-meta">
                    <a-space size="small" class="meta-item">
                      <i class="ri-time-line"></i>
                      <span>{{ formatDate(target.createdAt) }}</span>
                    </a-space>
                    <a-space size="small" class="meta-item">
                      <i class="ri-refresh-line"></i>
                      <span>{{ target.updatedAt ? formatDate(target.updatedAt) : '未扫描' }}</span>
                    </a-space>
                  </div>

                  <div class="card-actions">
                    <a-button size="small" @click="viewDetails(target)">
                      <i class="ri-file-list-line"></i>
                      详情
                    </a-button>
                    <a-button size="small" type="primary" @click="startScan(target)">
                      <i class="ri-scan-line"></i>
                      扫描
                    </a-button>
                  </div>
                </div>
              </a-card>
            </a-col>
          </a-row>
        </div>

        <!-- 表格视图 -->
        <a-table
          v-else
          :columns="tableColumns"
          :data-source="filteredTargets"
          :pagination="false"
          row-key="id"
          size="middle"
          class="targets-table"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="record.status === 'active' ? 'success' : 'default'">
                {{ record.status === 'active' ? '活跃' : '已归档' }}
              </a-tag>
            </template>

            <template v-else-if="column.key === 'createdAt'">
              {{ formatDate(record.createdAt) }}
            </template>

            <template v-else-if="column.key === 'updatedAt'">
              {{ record.updatedAt ? formatDate(record.updatedAt) : '未扫描' }}
            </template>

            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button size="small" @click="viewDetails(record)" title="查看详情">
                  <i class="ri-file-list-line"></i>
                </a-button>
                <a-button size="small" type="primary" @click="startScan(record)" title="开始扫描">
                  <i class="ri-scan-line"></i>
                </a-button>
                <a-button size="small" @click="editTarget(record)" title="编辑目标">
                  <i class="ri-edit-line"></i>
                </a-button>
                <a-button size="small" @click="archiveTarget(record)" title="归档/激活">
                  <i class="ri-archive-line"></i>
                </a-button>
                <a-button size="small" danger @click="deleteTarget(record)" title="删除目标">
                  <i class="ri-delete-bin-line"></i>
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>

        <!-- 空状态 -->
        <a-empty
          v-if="filteredTargets.length === 0"
          class="empty-state"
          description="暂无目标数据"
        >
          <template #image>
            <i class="ri-radar-line empty-icon"></i>
          </template>
          <a-button type="primary" @click="openCreateDialog">
            <i class="ri-add-line"></i>
            创建第一个目标
          </a-button>
        </a-empty>
      </a-card>
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

// 表格列定义
const tableColumns = [
  {
    title: '目标名称',
    dataIndex: 'name',
    key: 'name',
    width: 150,
  },
  {
    title: '目标地址',
    dataIndex: 'target',
    key: 'target',
    ellipsis: true,
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    align: 'center',
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 150,
  },
  {
    title: '上次更新',
    key: 'updatedAt',
    width: 150,
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    align: 'center',
  },
];

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

// 导出所有需要的数据和方法
return {
  searchQuery,
  statusFilter,
  viewMode,
  tableColumns,
  filteredTargets,
  refreshTargets,
  formatDate,
};
</script>

<style scoped>
/* 网络安全主题样式 */
.target-management {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  min-height: 100vh;
  color: #ffffff;
}

.management-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
  padding-top: 88px; /* 为HeaderPage留空间 */
}

/* 主卡片样式 */
:deep(.management-card.ant-card) {
  background: rgba(31, 41, 55, 0.4);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 16px;
  backdrop-filter: blur(12px);
}

:deep(.management-card .ant-card-body) {
  padding: 32px;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-title {
  display: flex;
  align-items: center;
  color: #ffffff;
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.page-title i {
  margin-right: 12px;
  color: #22d3ee;
  font-size: 28px;
}

.count-tag {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.4);
  color: #60a5fa;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 搜索和筛选区域 */
.search-filters {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.search-input {
  flex: 1;
  min-width: 300px;
}

:deep(.search-input .ant-input) {
  background: rgba(17, 24, 39, 0.6);
  border-color: rgba(75, 85, 99, 0.4);
  color: #ffffff;
}

:deep(.search-input .ant-input:focus) {
  border-color: #22d3ee;
  box-shadow: 0 0 0 2px rgba(34, 211, 238, 0.1);
}

/* 筛选按钮 */
.filter-buttons .status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
}

.status-dot.active {
  background: #10b981;
}

.status-dot.archived {
  background: #6b7280;
}

/* 视图控制 */
.view-controls {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 24px;
}

/* 卡片视图 */
.cards-container {
  margin-top: 16px;
}

:deep(.target-card.ant-card) {
  background: rgba(55, 65, 81, 0.4);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 12px;
  transition: all 0.3s ease;
}

:deep(.target-card.ant-card:hover) {
  border-color: rgba(34, 211, 238, 0.5);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
  transform: translateY(-2px);
}

:deep(.target-card .ant-card-head) {
  background: rgba(75, 85, 99, 0.2);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 12px 12px 0 0;
}

:deep(.target-card .ant-card-head-title) {
  color: #ffffff;
  font-weight: 600;
}

.card-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.target-name {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.card-content {
  color: #d1d5db;
}

.target-url {
  color: #9ca3af;
  font-size: 13px;
  margin-bottom: 8px;
  word-break: break-all;
}

.target-desc {
  color: #d1d5db;
  font-size: 14px;
  margin-bottom: 16px;
  min-height: 40px;
  line-height: 1.4;
}

.card-meta {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
  font-size: 12px;
  color: #9ca3af;
}

.meta-item i {
  color: #6b7280;
}

.card-actions {
  display: flex;
  gap: 8px;
}

/* 表格样式 */
:deep(.targets-table.ant-table) {
  background: transparent;
  color: #ffffff;
}

:deep(.targets-table .ant-table-thead > tr > th) {
  background: rgba(55, 65, 81, 0.5);
  color: #d1d5db;
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  font-weight: 600;
}

:deep(.targets-table .ant-table-tbody > tr > td) {
  background: transparent;
  color: #d1d5db;
  border-bottom: 1px solid rgba(75, 85, 99, 0.2);
}

:deep(.targets-table .ant-table-tbody > tr:hover > td) {
  background: rgba(75, 85, 99, 0.2);
}

/* 下拉菜单 */
:deep(.danger-item) {
  color: #f87171 !important;
}

:deep(.danger-item:hover) {
  background: rgba(248, 113, 113, 0.1) !important;
}

/* 空状态 */
.empty-state {
  margin: 48px 0;
}

.empty-icon {
  font-size: 64px;
  color: #6b7280;
}

/* 按钮样式 */
:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
}

:deep(.ant-btn-primary:hover) {
  background: linear-gradient(135deg, #2563eb, #1e40af);
}

:deep(.create-btn) {
  background: linear-gradient(135deg, #10b981, #059669);
  border: none;
}

:deep(.create-btn:hover) {
  background: linear-gradient(135deg, #059669, #047857);
}

:deep(.refresh-btn) {
  background: rgba(75, 85, 99, 0.4);
  border-color: rgba(75, 85, 99, 0.6);
  color: #d1d5db;
}

:deep(.refresh-btn:hover) {
  background: rgba(75, 85, 99, 0.6);
  border-color: #22d3ee;
  color: #22d3ee;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .management-container {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .search-filters {
    flex-direction: column;
  }

  .search-input {
    min-width: auto;
    width: 100%;
  }
}
</style>
