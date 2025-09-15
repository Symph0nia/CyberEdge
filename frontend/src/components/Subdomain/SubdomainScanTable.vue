<template>
  <div >
    <!-- 结果表格 -->
    <div v-if="subdomainScanResults?.length > 0" >
      <!-- 数据表格 -->
      <div
        class="overflow- border"
      >
        <div class="custom-scrollbar">
          <table >
            <thead>
              <tr >
                <th >
                  <input
                    type="checkbox"
                    @change="toggleSelectAll"
                    v-model="selectAll"
                    class="checkbox-input"
                    :disabled="subdomainScanResults.length === 0"
                    id="select-all-header"
                    title="全选/取消全选"
                  />
                  <label for="select-all-header" class="sr-only"
                    >全选/取消全选</label
                  >
                </th>
                <th
                  v-for="header in tableHeaders"
                  :key="header.key"
                  :class="[ ' ', header.width, ]"
                >
                  {{ header.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(result, index) in subdomainScanResults"
                :key="result.id"
                class="duration-200 hover:"
                :class="index % 2 === 0 ? '' : ''"
              >
                <td >
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
                <td >
                  {{ result.id }}
                </td>
                <td >
                  <span >
                    <i class="ri-global-line"></i>
                    {{ result.target }}
                  </span>
                </td>
                <td >
                  <span >
                    <i class="ri-time-line"></i>
                    {{ formatDate(result.timestamp) }}
                  </span>
                </td>
                <td >
                  <span
                    class="border"
                  >
                    <i class="ri-radar-line .5"></i>
                    {{ getSubdomainCount(result) }} 个
                  </span>
                </td>
                <td >
                  <span
                    
                    :class="result.is_read ? ' border ' : ' border '"
                  >
                    <i
                      :class="result.is_read ? 'ri-eye-line' : 'ri-eye-off-line'"
                      class=".5"
                    ></i>
                    {{ result.is_read ? "已读" : "未读" }}
                  </span>
                </td>
                <td >
                  <div >
                    <button
                      @click="handleViewDetails(result.id)"
                      class="action-button border hover:"
                    >
                      <i class="ri-eye-line .5"></i>
                      查看
                    </button>
                    <button
                      @click="handleToggleRead(result)"
                      class="action-button"
                      :class="result.is_read ? ' border ' : ' border hover:'"
                    >
                      <i
                        :class="[ result.is_read ? 'ri-eye-off-line' : 'ri-eye-line', '.5', ]"
                      ></i>
                      {{ result.is_read ? "标为未读" : "标为已读" }}
                    </button>
                    <button
                      @click="handleDelete(result.id)"
                      class="action-button border hover:"
                    >
                      <i class="ri-delete-bin-line .5"></i>
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 批量操作工具栏 -->
      <div
        class="border"
      >
        <div >
          <span >
            <template v-if="hasSelected">
              已选择
              <span >{{
                selectedResults.length
              }}</span>
              项
            </template>
            <template v-else> 请选择要操作的项目 </template>
          </span>
        </div>

        <div >
          <button
            @click="handleBatchRead"
            :disabled="!hasSelected"
            class="batch-button"
            :class="[ !hasSelected ? ' ' : ' hover:', ]"
          >
            <i class="ri-eye-line"></i>
            标记已读
            <span v-if="hasSelected">({{ selectedResults.length }})</span>
          </button>
          <button
            @click="handleBatchDelete"
            :disabled="!hasSelected"
            class="batch-button"
            :class="[ !hasSelected ? ' ' : ' hover:', ]"
          >
            <i class="ri-delete-bin-line"></i>
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
  name: "SubdomainScanTable",
  props: {
    subdomainScanResults: {
      type: Array,
      required: true,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, { emit }) {
    const selectedResults = ref([]);
    const selectAll = ref(false);
    const tableHeaders = [
      { key: "id", label: "ID", width: "w-24" },
      { key: "target", label: "目标地址", width: "w-48" },
      { key: "timestamp", label: "扫描时间", width: "w-44" },
      { key: "count", label: "子域名数量", width: "w-32" },
      { key: "status", label: "状态", width: "w-24" },
      { key: "actions", label: "操作", width: "w-52" },
    ];

    // 计算是否有选中的项目
    const hasSelected = computed(() => selectedResults.value.length > 0);

    // 监听结果变化，重置选择状态
    watch(
      () => props.subdomainScanResults,
      () => {
        selectedResults.value = [];
        selectAll.value = false;
      }
    );

    // 监听全选状态变化
    watch(
      () => selectAll.value,
      (newVal) => {
        if (newVal) {
          // 全选时，将所有结果ID添加到选中数组
          selectedResults.value = props.subdomainScanResults.map(
            (result) => result.id
          );
        } else {
          // 取消全选时，清空选中数组
          selectedResults.value = [];
        }
      }
    );

    // 监听选中结果变化，自动更新全选状态
    watch(
      () => selectedResults.value,
      (newVal) => {
        if (props.subdomainScanResults.length > 0) {
          // 当选择的数量等于总数时，设置全选状态为true
          selectAll.value = newVal.length === props.subdomainScanResults.length;
        }
      }
    );

    // 切换全选状态
    const toggleSelectAll = () => {
      // 由watch处理具体逻辑
      selectAll.value = !selectAll.value;
    };

    // 获取子域名数量
    const getSubdomainCount = (result) => {
      if (!result || !result.data || !Array.isArray(result.data)) {
        return 0;
      }
      const subdomainGroup = result.data.find(
        (group) => group.Key === "subdomains"
      );
      return subdomainGroup?.Value?.length || 0;
    };

    // 格式化日期
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

    // 查看详情
    const handleViewDetails = (id) => emit("view-details", id);

    // 切换已读/未读状态
    const handleToggleRead = (result) =>
      emit("toggle-read-status", result.id, !result.is_read);

    // 删除单个记录
    const handleDelete = (id) => emit("delete-result", id);

    // 批量标记为已读
    const handleBatchRead = () => {
      if (selectedResults.value.length === 0) return;
      emit("mark-selected-read", selectedResults.value);
    };

    // 批量删除
    const handleBatchDelete = () => {
      if (selectedResults.value.length === 0) return;
      emit("delete-selected", selectedResults.value);
    };

    return {
      selectedResults,
      selectAll,
      tableHeaders,
      hasSelected,
      toggleSelectAll,
      getSubdomainCount,
      formatDate,
      handleViewDetails,
      handleToggleRead,
      handleDelete,
      handleBatchRead,
      handleBatchDelete,
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
