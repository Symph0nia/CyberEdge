<template>
  <div class="port-scan-table-container">
    <!-- 批量操作工具栏 -->
    <div class="batch-toolbar" v-if="selectedRowKeys.length > 0">
      <div class="batch-info">
        已选择 <span class="selected-count">{{ selectedRowKeys.length }}</span> 项
      </div>
      <a-space>
        <a-button
          type="primary"
          ghost
          @click="handleBatchRead"
          :icon="h('i', { class: 'ri-eye-line' })"
        >
          标记已读 ({{ selectedRowKeys.length }})
        </a-button>
        <a-button
          danger
          ghost
          @click="handleBatchDelete"
          :icon="h('i', { class: 'ri-delete-bin-line' })"
        >
          批量删除 ({{ selectedRowKeys.length }})
        </a-button>
      </a-space>
    </div>

    <!-- 主表格 -->
    <a-table
      :columns="columns"
      :data-source="portScanResults"
      :row-key="record => record.id"
      :row-selection="rowSelection"
      :pagination="paginationConfig"
      :scroll="{ x: 'max-content' }"
      :row-class-name="getRowClassName"
      size="middle"
      class="cyber-table"
    >
      <!-- 目标地址列 -->
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'target'">
          <div class="target-cell">
            <i class="ri-global-line target-icon"></i>
            <span class="target-text">{{ record.target }}</span>
          </div>
        </template>

        <!-- 端口数量列 -->
        <template v-else-if="column.key === 'port_count'">
          <a-tag color="blue" class="port-count-tag">
            {{ getPortCount(record.ports) }} 个端口
          </a-tag>
        </template>

        <!-- 状态列 -->
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)" class="status-tag">
            {{ record.status }}
          </a-tag>
        </template>

        <!-- 扫描时间列 -->
        <template v-else-if="column.key === 'scan_time'">
          <div class="time-cell">
            <i class="ri-time-line time-icon"></i>
            <span class="time-text">{{ formatDate(record.scan_time) }}</span>
          </div>
        </template>

        <!-- 已读状态列 -->
        <template v-else-if="column.key === 'is_read'">
          <a-tag :color="record.is_read ? 'default' : 'success'">
            <i :class="record.is_read ? 'ri-eye-line' : 'ri-eye-off-line'"></i>
            {{ record.is_read ? '已读' : '未读' }}
          </a-tag>
        </template>

        <!-- 操作列 -->
        <template v-else-if="column.key === 'actions'">
          <a-space size="small">
            <a-tooltip title="查看详情">
              <a-button
                type="text"
                size="small"
                @click="handleViewDetails(record)"
                :icon="h('i', { class: 'ri-eye-line' })"
              />
            </a-tooltip>

            <a-tooltip :title="record.is_read ? '标为未读' : '标为已读'">
              <a-button
                type="text"
                size="small"
                @click="handleToggleRead(record)"
                :icon="h('i', { class: record.is_read ? 'ri-eye-off-line' : 'ri-eye-line' })"
              />
            </a-tooltip>

            <a-popconfirm
              title="确定要删除这个扫描结果吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleDelete(record.id)"
            >
              <a-tooltip title="删除">
                <a-button
                  type="text"
                  size="small"
                  danger
                  :icon="h('i', { class: 'ri-delete-bin-line' })"
                />
              </a-tooltip>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script>
import { ref, h } from "vue";
import { message } from 'ant-design-vue';

export default {
  name: "PortScanTable",
  props: {
    portScanResults: {
      type: Array,
      required: true,
      default: () => [],
    },
  },
  emits: [
    "view-details",
    "delete-result",
    "delete-selected",
    "toggle-read-status",
    "mark-selected-read",
  ],
  setup(props, { emit }) {
    const selectedRowKeys = ref([]);

    // 表格列配置
    const columns = [
      {
        title: 'ID',
        dataIndex: 'id',
        key: 'id',
        width: 100,
        sorter: (a, b) => a.id - b.id,
      },
      {
        title: '目标地址',
        dataIndex: 'target',
        key: 'target',
        width: 200,
        sorter: (a, b) => a.target.localeCompare(b.target),
      },
      {
        title: '端口数量',
        key: 'port_count',
        width: 120,
        sorter: (a, b) => getPortCount(a.ports) - getPortCount(b.ports),
      },
      {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 100,
        filters: [
          { text: '完成', value: '完成' },
          { text: '进行中', value: '进行中' },
          { text: '失败', value: '失败' },
        ],
        onFilter: (value, record) => record.status === value,
      },
      {
        title: '扫描时间',
        dataIndex: 'scan_time',
        key: 'scan_time',
        width: 180,
        sorter: (a, b) => new Date(a.scan_time) - new Date(b.scan_time),
      },
      {
        title: '状态',
        key: 'is_read',
        width: 80,
        filters: [
          { text: '已读', value: true },
          { text: '未读', value: false },
        ],
        onFilter: (value, record) => record.is_read === value,
      },
      {
        title: '操作',
        key: 'actions',
        width: 150,
        fixed: 'right',
      },
    ];

    // 行选择配置
    const rowSelection = {
      selectedRowKeys: selectedRowKeys,
      onChange: (keys) => {
        selectedRowKeys.value = keys;
      },
      onSelectAll: () => {
        // 处理全选逻辑
      },
    };

    // 分页配置
    const paginationConfig = {
      showSizeChanger: true,
      showQuickJumper: true,
      showTotal: (total, range) =>
        `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
      pageSizeOptions: ['10', '20', '50', '100'],
    };

    // 工具函数
    const getPortCount = (ports) => {
      return ports ? (Array.isArray(ports) ? ports.length : 0) : 0;
    };

    const formatDate = (dateStr) => {
      if (!dateStr) return '-';
      const date = new Date(dateStr);
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      });
    };

    const getStatusColor = (status) => {
      const colorMap = {
        '完成': 'success',
        '进行中': 'processing',
        '失败': 'error',
        '等待中': 'warning',
      };
      return colorMap[status] || 'default';
    };

    const getRowClassName = (record) => {
      return record.is_read ? 'row-read' : 'row-unread';
    };

    // 事件处理
    const handleViewDetails = (record) => {
      emit('view-details', record);
    };

    const handleDelete = (id) => {
      emit('delete-result', id);
      message.success('删除成功');
    };

    const handleToggleRead = (record) => {
      emit('toggle-read-status', record);
      message.success(record.is_read ? '已标记为未读' : '已标记为已读');
    };

    const handleBatchRead = () => {
      if (selectedRowKeys.value.length === 0) {
        message.warning('请先选择要操作的项目');
        return;
      }
      emit('mark-selected-read', selectedRowKeys.value);
      message.success(`已标记 ${selectedRowKeys.value.length} 项为已读`);
      selectedRowKeys.value = [];
    };

    const handleBatchDelete = () => {
      if (selectedRowKeys.value.length === 0) {
        message.warning('请先选择要删除的项目');
        return;
      }
      emit('delete-selected', selectedRowKeys.value);
      message.success(`已删除 ${selectedRowKeys.value.length} 项`);
      selectedRowKeys.value = [];
    };

    return {
      h,
      selectedRowKeys,
      columns,
      rowSelection,
      paginationConfig,
      getPortCount,
      formatDate,
      getStatusColor,
      getRowClassName,
      handleViewDetails,
      handleDelete,
      handleToggleRead,
      handleBatchRead,
      handleBatchDelete,
    };
  },
};
</script>

<style scoped>
.port-scan-table-container {
  background: transparent;
}

.batch-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  margin-bottom: 16px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
  color: #e2e8f0;
}

.batch-info {
  font-size: 14px;
  color: #94a3b8;
}

.selected-count {
  color: #60a5fa;
  font-weight: 600;
  margin: 0 4px;
}

.target-cell {
  display: flex;
  align-items: center;
}

.target-icon {
  color: #60a5fa;
  margin-right: 8px;
  font-size: 14px;
}

.target-text {
  color: #e2e8f0;
  font-family: 'Monaco', 'Consolas', monospace;
}

.port-count-tag {
  font-weight: 500;
}

.status-tag {
  font-weight: 500;
}

.time-cell {
  display: flex;
  align-items: center;
}

.time-icon {
  color: #94a3b8;
  margin-right: 6px;
  font-size: 12px;
}

.time-text {
  color: #cbd5e1;
  font-size: 13px;
}

/* Ant Design样式覆盖 */
.cyber-table :deep(.ant-table) {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
}

.cyber-table :deep(.ant-table-thead > tr > th) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.5);
  color: #94a3b8;
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.cyber-table :deep(.ant-table-tbody > tr) {
  background: transparent;
  transition: all 0.2s ease;
}

.cyber-table :deep(.ant-table-tbody > tr:hover) {
  background: rgba(51, 65, 85, 0.3) !important;
}

.cyber-table :deep(.ant-table-tbody > tr.row-unread) {
  background: rgba(59, 130, 246, 0.1);
}

.cyber-table :deep(.ant-table-tbody > tr.row-read) {
  background: transparent;
  opacity: 0.8;
}

.cyber-table :deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid rgba(51, 65, 85, 0.2);
  color: #e2e8f0;
  padding: 12px 16px;
}

.cyber-table :deep(.ant-table-selection-column) {
  padding-left: 20px;
}

.cyber-table :deep(.ant-checkbox-wrapper) {
  color: #e2e8f0;
}

.cyber-table :deep(.ant-checkbox-inner) {
  background-color: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.5);
}

.cyber-table :deep(.ant-checkbox-checked .ant-checkbox-inner) {
  background-color: #10b981;
  border-color: #10b981;
}

.cyber-table :deep(.ant-pagination) {
  margin-top: 24px;
  text-align: right;
}

.cyber-table :deep(.ant-pagination-item) {
  background: rgba(30, 41, 59, 0.6);
  border-color: rgba(51, 65, 85, 0.3);
}

.cyber-table :deep(.ant-pagination-item a) {
  color: #94a3b8;
}

.cyber-table :deep(.ant-pagination-item-active) {
  background: rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
}

.cyber-table :deep(.ant-pagination-item-active a) {
  color: #60a5fa;
}

.cyber-table :deep(.ant-btn-text) {
  color: #94a3b8;
  border: none;
}

.cyber-table :deep(.ant-btn-text:hover) {
  color: #e2e8f0;
  background: rgba(51, 65, 85, 0.3);
}

.cyber-table :deep(.ant-btn-dangerous.ant-btn-text) {
  color: #f87171;
}

.cyber-table :deep(.ant-btn-dangerous.ant-btn-text:hover) {
  color: #fca5a5;
  background: rgba(239, 68, 68, 0.1);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .batch-toolbar {
    flex-direction: column;
    gap: 12px;
    text-align: center;
  }

  .cyber-table :deep(.ant-table-thead > tr > th) {
    font-size: 11px;
    padding: 8px 12px;
  }

  .cyber-table :deep(.ant-table-tbody > tr > td) {
    padding: 10px 12px;
    font-size: 13px;
  }
}
</style>