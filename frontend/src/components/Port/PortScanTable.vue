<template>
  <div class="space-y-6">
    <!-- 结果表格 -->
    <div v-if="portScanResults?.length > 0">
      <div
        class="relative overflow-x-auto rounded-xl border border-gray-700/30 bg-gray-800/30"
      >
        <table class="w-full">
          <thead>
            <tr class="bg-gray-800/60 border-b border-gray-700/50">
              <th class="py-3 px-4 text-left w-10">
                <input
                  type="checkbox"
                  @change="toggleSelectAll"
                  :checked="isAllSelected"
                  class="checkbox-input"
                  id="select-all-header"
                  title="全选/取消全选"
                />
                <label for="select-all-header" class="sr-only"
                  >全选/取消全选</label
                >
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
              v-for="(result, index) in portScanResults"
              :key="result.id"
              class="border-b border-gray-700/30 transition-all duration-200 hover:bg-gray-700/40"
              :class="index % 2 === 0 ? 'bg-gray-800/20' : ''"
            >
              <td class="py-3 px-4">
                <input
                  type="checkbox"
                  v-model="selectedResults"
                  :value="result.id"
                  class="checkbox-input"
                  :id="`select-result-${result.id}`"
                />
                <label :for="`select-result-${result.id}`" class="sr-only"
                  >选择此结果</label
                >
              </td>
              <td class="py-3 px-4 text-sm font-mono text-gray-300">
                {{ result.id }}
              </td>
              <td class="py-3 px-4 text-sm text-gray-200">
                <span class="flex items-center">
                  <i class="ri-global-line mr-2 text-blue-400"></i>
                  {{ result.target }}
                </span>
              </td>
              <td class="py-3 px-4 text-sm text-gray-300">
                <span class="flex items-center">
                  <i class="ri-time-line mr-2 text-gray-500"></i>
                  {{ formatDate(result.timestamp) }}
                </span>
              </td>
              <td class="py-3 px-4 text-sm text-gray-200">
                <span
                  class="px-2 py-1 rounded-md bg-blue-500/10 text-blue-300 border border-blue-500/20 inline-flex items-center"
                >
                  <i class="ri-scan-2-line mr-1.5"></i>
                  {{ getPortCount(result) }} 个
                </span>
              </td>
              <td class="py-3 px-4">
                <span
                  class="px-2 py-1 rounded-md text-xs font-medium inline-flex items-center"
                  :class="
                    result.is_read
                      ? 'bg-green-500/20 text-green-300 border border-green-500/30'
                      : 'bg-yellow-500/20 text-yellow-300 border border-yellow-500/30'
                  "
                >
                  <i
                    :class="result.is_read ? 'ri-eye-line' : 'ri-eye-off-line'"
                    class="mr-1.5"
                  ></i>
                  {{ result.is_read ? "已读" : "未读" }}
                </span>
              </td>
              <td class="py-3 px-4">
                <div class="flex gap-2 flex-wrap">
                  <button
                    @click="handleViewDetails(result.id)"
                    class="action-button bg-blue-500/20 text-blue-300 border border-blue-500/30 hover:bg-blue-500/30"
                  >
                    <i class="ri-eye-line mr-1.5"></i>
                    查看
                  </button>
                  <button
                    @click="handleToggleRead(result)"
                    class="action-button"
                    :class="
                      result.is_read
                        ? 'bg-gray-700/50 text-gray-300 border border-gray-600/30'
                        : 'bg-green-500/20 text-green-300 border border-green-500/30 hover:bg-green-500/30'
                    "
                  >
                    <i
                      :class="[
                        result.is_read ? 'ri-eye-off-line' : 'ri-eye-line',
                        'mr-1.5',
                      ]"
                    ></i>
                    {{ result.is_read ? "标为未读" : "标为已读" }}
                  </button>
                  <button
                    @click="handleDelete(result.id)"
                    class="action-button bg-red-500/20 text-red-300 border border-red-500/30 hover:bg-red-500/30"
                  >
                    <i class="ri-delete-bin-line mr-1.5"></i>
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 批量操作工具栏 -->
      <div
        class="flex items-center justify-between flex-wrap gap-4 p-4 rounded-xl border border-gray-700/30 bg-gray-800/30 mt-6"
      >
        <div class="flex items-center gap-3">
          <span class="text-sm text-gray-400">
            <template v-if="hasSelected">
              已选择
              <span class="text-white font-medium">{{
                selectedResults.length
              }}</span>
              项
            </template>
            <template v-else> 请选择要操作的项目 </template>
          </span>
        </div>

        <div class="flex flex-wrap gap-3">
          <button
            @click="handleBatchRead"
            :disabled="!hasSelected"
            class="batch-button"
            :class="[
              !hasSelected
                ? 'bg-gray-700/50 text-gray-400 border-gray-600/30 cursor-not-allowed'
                : 'bg-green-500/20 text-green-300 border-green-500/30 hover:bg-green-500/30',
            ]"
          >
            <i class="ri-eye-line mr-2"></i>
            标记已读
            <span v-if="hasSelected">({{ selectedResults.length }})</span>
          </button>
          <button
            @click="handleBatchDelete"
            :disabled="!hasSelected"
            class="batch-button"
            :class="[
              !hasSelected
                ? 'bg-gray-700/50 text-gray-400 border-gray-600/30 cursor-not-allowed'
                : 'bg-red-500/20 text-red-300 border-red-500/30 hover:bg-red-500/30',
            ]"
          >
            <i class="ri-delete-bin-line mr-2"></i>
            批量删除
            <span v-if="hasSelected">({{ selectedResults.length }})</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from "vue";

export default {
  name: "PortScanTable",
  props: {
    portScanResults: {
      type: Array,
      required: true,
      default: () => [], // 设置默认值为空数组
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
    // 表格头部配置
    const tableHeaders = [
      "ID",
      "目标地址",
      "扫描时间",
      "端口数量",
      "状态",
      "操作",
    ];

    // 选中的结果ID列表
    const selectedResults = ref([]);

    // 计算属性：是否有选中项
    const hasSelected = computed(() => selectedResults.value.length > 0);

    // 计算属性：是否全选
    const isAllSelected = computed(() => {
      if (!props.portScanResults?.length) return false;
      return selectedResults.value.length === props.portScanResults.length;
    });

    // 格式化日期时间
    const formatDate = (timestamp) => {
      if (!timestamp) return "-";
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

    // 获取端口数量
    const getPortCount = (result) => {
      if (!result || !result.data || !Array.isArray(result.data)) {
        return 0;
      }
      const portGroup = result.data.find((group) => group.Key === "ports");
      return portGroup?.Value?.length || 0;
    };

    // 全选/取消全选
    const toggleSelectAll = () => {
      if (isAllSelected.value) {
        // 如果当前是全选状态，则取消全选
        selectedResults.value = [];
      } else {
        // 如果当前不是全选状态，则全选
        selectedResults.value = props.portScanResults.map(
          (result) => result.id
        );
      }
    };

    // 查看详情处理函数
    const handleViewDetails = (id) => emit("view-details", id);

    // 删除单个结果处理函数
    const handleDelete = (id) => emit("delete-result", id);

    // 切换已读状态处理函数
    const handleToggleRead = (result) =>
      emit("toggle-read-status", result.id, !result.is_read);

    // 批量删除处理函数
    const handleBatchDelete = () => {
      if (selectedResults.value.length === 0) return;
      emit("delete-selected", selectedResults.value);
      // 操作后清空选择
      selectedResults.value = [];
    };

    // 批量标记已读处理函数
    const handleBatchRead = () => {
      if (selectedResults.value.length === 0) return;
      emit("mark-selected-read", selectedResults.value);
      // 操作后清空选择
      selectedResults.value = [];
    };

    // 监听结果变化，重置选择状态
    watch(
      () => props.portScanResults,
      () => {
        selectedResults.value = [];
      }
    );

    return {
      selectedResults,
      tableHeaders,
      hasSelected,
      isAllSelected,
      formatDate,
      getPortCount,
      toggleSelectAll,
      handleViewDetails,
      handleDelete,
      handleToggleRead,
      handleBatchDelete,
      handleBatchRead,
    };
  },
};
</script>

<style scoped>
/* 操作按钮样式 */
.action-button {
  @apply px-3 py-1.5 rounded-md text-xs font-medium
  transition-all duration-200 flex items-center
  focus:outline-none focus:ring-1 focus:ring-opacity-50
  disabled:opacity-50 disabled:cursor-not-allowed;
}

/* 批量操作按钮样式 */
.batch-button {
  @apply px-4 py-2 rounded-lg text-sm font-medium
  transition-all duration-200 flex items-center border
  focus:outline-none focus:ring-1 focus:ring-opacity-50
  disabled:opacity-50 disabled:cursor-not-allowed;
}

/* 优化按钮点击效果 */
.batch-button:active:not(:disabled),
.action-button:active:not(:disabled) {
  transform: scale(0.98);
}

/* 自定义复选框样式 */
.checkbox-input {
  @apply rounded-md border-gray-700/50 bg-gray-800/50
  text-blue-500 focus:ring-blue-500/30 h-4 w-4 cursor-pointer;
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
  background-color: rgba(107, 114, 128, 0.3);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: rgba(107, 114, 128, 0.5);
}
</style>
