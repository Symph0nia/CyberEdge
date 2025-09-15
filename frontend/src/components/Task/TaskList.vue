<template>
  <div class="task-list">
    <!-- Header section -->
    <div class="list-header">
      <h2 class="list-title">
        <i class="ri-task-line"></i>
        任务管理
      </h2>

      <a-space>
        <!-- Batch operations -->
        <div v-if="selectedTasks.length > 0" class="batch-actions">
          <a-button
            @click="handleBatchStart"
            type="primary"
            class="batch-btn"
          >
            <i class="ri-play-line"></i>
            批量启动 ({{ selectedTasks.length }})
          </a-button>
          <a-button
            @click="handleBatchDelete"
            danger
            class="batch-btn"
          >
            <i class="ri-delete-bin-line"></i>
            批量删除 ({{ selectedTasks.length }})
          </a-button>
        </div>

        <!-- Refresh button -->
        <a-button @click="$emit('refresh-tasks')" class="refresh-btn">
          <i class="ri-refresh-line"></i>
          刷新列表
        </a-button>
      </a-space>
    </div>

    <!-- Task table -->
    <div v-if="tasks?.length > 0" class="task-table-container">
      <a-table
        :dataSource="tasks"
        :columns="columns"
        :row-selection="{
          selectedRowKeys: selectedTasks,
          onChange: onSelectChange,
          onSelectAll: onSelectAll
        }"
        :scroll="{ x: 1000 }"
        :pagination="false"
        row-key="id"
        class="task-table"
      >
        <!-- 描述列自定义渲染 -->
        <template #description="{ record }">
          <div class="task-description">
            <i :class="getTypeIcon(record.type)" class="type-icon"></i>
            {{ formatDescription(record.type) }}
          </div>
        </template>

        <!-- 目标列自定义渲染 -->
        <template #target="{ record }">
          <a-tooltip :title="record.payload">
            <span class="task-target">{{ record.payload }}</span>
          </a-tooltip>
        </template>

        <!-- 状态列自定义渲染 -->
        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.status)" class="status-tag">
            <i :class="getStatusIcon(record.status)" class="status-icon"></i>
            {{ formatStatus(record.status) }}
          </a-tag>
        </template>

        <!-- 创建时间列自定义渲染 -->
        <template #created_at="{ record }">
          {{ formatDate(record.created_at) }}
        </template>

        <!-- 完成时间列自定义渲染 -->
        <template #completed_at="{ record }">
          {{ record.completed_at ? formatDate(record.completed_at) : "—" }}
        </template>

        <!-- 结果列自定义渲染 -->
        <template #result="{ record }">
          <a-tooltip v-if="record.result" :title="record.result">
            <span class="task-result">{{ truncateText(record.result, 20) }}</span>
          </a-tooltip>
          <span v-else class="no-result">—</span>
        </template>

        <!-- 操作列自定义渲染 -->
        <template #action="{ record }">
          <a-space>
            <a-button
              @click="$emit('toggle-task', record)"
              :disabled="record.status === 'running'"
              :type="record.status === 'running' ? 'default' : 'primary'"
              size="small"
              class="action-btn"
            >
              <i :class="record.status === 'running' ? 'ri-loader-2-line animate-spin' : 'ri-play-line'"></i>
              {{ record.status === "running" ? "运行中" : "启动" }}
            </a-button>
            <a-button
              @click="handleDelete(record.id)"
              danger
              size="small"
              class="action-btn"
            >
              <i class="ri-delete-bin-line"></i>
              删除
            </a-button>
          </a-space>
        </template>
      </a-table>
    </div>

    <!-- Empty state -->
    <div v-else class="empty-state">
      <a-empty
        description="暂无任务"
        :image="Empty.PRESENTED_IMAGE_SIMPLE"
      >
        <template #description>
          <span class="empty-title">暂无任务</span>
          <p class="empty-text">
            当前还没有创建任何扫描任务，可以在下方创建或前往目标管理添加
          </p>
        </template>
        <router-link to="/target-management">
          <a-button type="primary" class="empty-action-btn">
            <i class="ri-focus-3-line"></i>
            前往目标管理
          </a-button>
        </router-link>
      </a-empty>
    </div>
  </div>
</template>

<script>
import { Empty } from 'ant-design-vue';

export default {
  name: "TaskList",
  props: {
    tasks: {
      type: Array,
      required: true,
      default: () => [],
    },
  },
  data() {
    return {
      selectedTasks: [],
      Empty,
      columns: [
        {
          title: '任务ID',
          dataIndex: 'id',
          key: 'id',
          width: 100,
          fixed: 'left',
        },
        {
          title: '描述',
          dataIndex: 'type',
          key: 'description',
          width: 150,
          slots: { customRender: 'description' },
        },
        {
          title: '目标',
          dataIndex: 'payload',
          key: 'target',
          width: 200,
          ellipsis: true,
          slots: { customRender: 'target' },
        },
        {
          title: '状态',
          dataIndex: 'status',
          key: 'status',
          width: 120,
          slots: { customRender: 'status' },
        },
        {
          title: '创建时间',
          dataIndex: 'created_at',
          key: 'created_at',
          width: 160,
          slots: { customRender: 'created_at' },
        },
        {
          title: '完成时间',
          dataIndex: 'completed_at',
          key: 'completed_at',
          width: 160,
          slots: { customRender: 'completed_at' },
        },
        {
          title: '结果',
          dataIndex: 'result',
          key: 'result',
          width: 150,
          ellipsis: true,
          slots: { customRender: 'result' },
        },
        {
          title: '操作',
          key: 'action',
          width: 180,
          fixed: 'right',
          slots: { customRender: 'action' },
        },
      ],
    };
  },
  methods: {
    formatStatus(status) {
      const statusMap = {
        running: "运行中",
        completed: "已完成",
        pending: "等待中",
      };
      return statusMap[status] || "未知状态";
    },
    getStatusColor(status) {
      const colorMap = {
        running: "processing",
        completed: "success",
        pending: "warning",
      };
      return colorMap[status] || "default";
    },
    getStatusIcon(status) {
      const iconMap = {
        running: "ri-loader-2-line animate-spin",
        completed: "ri-check-line",
        pending: "ri-time-line",
      };
      return iconMap[status] || "ri-question-line";
    },
    getTypeIcon(type) {
      const iconMap = {
        subfinder: "ri-global-line",
        nmap: "ri-scan-2-line",
        ffuf: "ri-folders-line",
      };
      return iconMap[type] || "ri-question-line";
    },
    formatDescription(type) {
      const descriptions = {
        subfinder: "子域名扫描",
        nmap: "端口扫描",
        ffuf: "路径扫描",
      };
      return descriptions[type] || "未知任务";
    },
    formatDate(date) {
      return new Date(date).toLocaleString("zh-CN", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      });
    },
    truncateText(text, maxLength) {
      if (!text) return "";
      return text.length > maxLength
        ? text.substring(0, maxLength) + "..."
        : text;
    },
    handleDelete(taskId) {
      this.$emit("delete-task", taskId);
    },
    onSelectChange(selectedRowKeys) {
      this.selectedTasks = selectedRowKeys;
    },
    onSelectAll(selected, selectedRows, changeRows) {
      if (selected) {
        this.selectedTasks = this.tasks.map((task) => task.id);
      } else {
        this.selectedTasks = [];
      }
    },
    handleBatchStart() {
      this.$emit("batch-start", this.selectedTasks);
    },
    handleBatchDelete() {
      this.$emit("batch-delete", this.selectedTasks);
      this.selectedTasks = [];
    },
  },
};
</script>

<style scoped>
.task-list {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 16px;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
  background: rgba(15, 23, 42, 0.8);
}

.list-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 600;
  color: #f1f5f9;
  margin: 0;
}

.list-title i {
  font-size: 24px;
  color: #3b82f6;
}

.batch-actions {
  display: flex;
  gap: 8px;
  margin-right: 16px;
}

.batch-btn,
.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
}

.task-table-container {
  overflow: auto;
}

.task-description {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-icon {
  font-size: 16px;
  color: #3b82f6;
}

.task-target {
  display: block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 0;
}

.status-icon {
  font-size: 14px;
}

.task-result {
  display: block;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.no-result {
  color: #64748b;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
}

.empty-state {
  padding: 48px 24px;
  text-align: center;
}

.empty-title {
  color: #f1f5f9;
  font-size: 18px;
  font-weight: 600;
}

.empty-text {
  color: #94a3b8;
  margin: 8px auto 24px;
  max-width: 400px;
  line-height: 1.6;
}

.empty-action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Ant Design组件样式覆盖 */
.task-list :deep(.ant-table) {
  background: transparent;
}

.task-list :deep(.ant-table-thead > tr > th) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
  color: #94a3b8;
  font-weight: 600;
}

.task-list :deep(.ant-table-tbody > tr > td) {
  background: transparent;
  border-bottom: 1px solid rgba(51, 65, 85, 0.2);
  color: #e2e8f0;
}

.task-list :deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(51, 65, 85, 0.3);
}

.task-list :deep(.ant-table-selection-column) {
  width: 60px;
}

.task-list :deep(.ant-empty) {
  color: #e2e8f0;
}

.task-list :deep(.ant-empty-description) {
  color: #94a3b8;
}

/* 动画 */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .list-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .batch-actions {
    margin-right: 0;
    justify-content: center;
  }

  .empty-state {
    padding: 32px 16px;
  }
}
</style>