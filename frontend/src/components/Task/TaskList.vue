<template>
  <div class="flex flex-col">
    <!-- Header section with title and actions -->
    <div
      class="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4"
    >
      <h2
        class="text-xl font-medium tracking-wide text-gray-200 flex items-center"
      >
        <i class="ri-task-line mr-2 text-blue-400"></i>
        任务管理
      </h2>

      <div class="flex flex-wrap gap-3">
        <!-- Batch operations buttons -->
        <div v-if="selectedTasks.length > 0" class="flex gap-3 animate-fadeIn">
          <button
            @click="handleBatchStart"
            class="action-button bg-blue-500/30 hover:bg-blue-600/40 text-blue-100 border-blue-500/30"
          >
            <i class="ri-play-line mr-2"></i>
            批量启动 ({{ selectedTasks.length }})
          </button>
          <button
            @click="handleBatchDelete"
            class="action-button bg-red-500/30 hover:bg-red-600/40 text-red-100 border-red-500/30"
          >
            <i class="ri-delete-bin-line mr-2"></i>
            批量删除 ({{ selectedTasks.length }})
          </button>
        </div>

        <!-- Refresh button -->
        <button
          @click="$emit('refresh-tasks')"
          class="action-button bg-gray-700/50 hover:bg-gray-600/50 text-gray-200 border-gray-700/50"
        >
          <i class="ri-refresh-line mr-2"></i>
          刷新列表
        </button>
      </div>
    </div>

    <!-- Task data section -->
    <div
      v-if="tasks?.length > 0"
      class="bg-gray-800/30 rounded-xl border border-gray-700/30 overflow-hidden"
    >
      <!-- Responsive table with horizontal scrolling -->
      <div class="overflow-x-auto scrollbar-thin">
        <table class="w-full table-auto">
          <!-- Table header -->
          <thead>
            <tr class="bg-gray-800/50">
              <th class="table-header w-[60px]">
                <div class="flex items-center justify-center">
                  <input
                    type="checkbox"
                    :checked="isAllSelected"
                    @change="toggleSelectAll"
                    class="checkbox-input"
                  />
                </div>
              </th>
              <th class="table-header">任务ID</th>
              <th class="table-header">描述</th>
              <th class="table-header">目标</th>
              <th class="table-header">状态</th>
              <th class="table-header">创建时间</th>
              <th class="table-header">完成时间</th>
              <th class="table-header">结果</th>
              <th class="table-header">操作</th>
            </tr>
          </thead>

          <!-- Table body -->
          <tbody>
            <tr
              v-for="task in tasks"
              :key="task.id"
              class="border-t border-gray-700/30 hover:bg-gray-700/40 transition-all duration-200"
            >
              <td class="table-cell w-[60px]">
                <div class="flex items-center justify-center">
                  <input
                    type="checkbox"
                    v-model="selectedTasks"
                    :value="task.id"
                    class="checkbox-input"
                  />
                </div>
              </td>
              <td
                class="table-cell w-[120px] whitespace-nowrap font-mono text-xs"
              >
                {{ task.id }}
              </td>
              <td class="table-cell w-[140px] whitespace-nowrap">
                <div class="flex items-center">
                  <i
                    :class="getTypeIcon(task.type)"
                    class="mr-2 text-blue-400"
                  ></i>
                  {{ formatDescription(task.type) }}
                </div>
              </td>
              <td
                class="table-cell w-[180px] whitespace-nowrap max-w-[180px] truncate"
              >
                <span class="tooltip" :data-tooltip="task.payload">
                  {{ task.payload }}
                </span>
              </td>
              <td class="table-cell w-[100px]">
                <span class="status-badge" :class="getStatusStyle(task.status)">
                  <i :class="getStatusIcon(task.status)" class="mr-1"></i>
                  {{ formatStatus(task.status) }}
                </span>
              </td>
              <td class="table-cell w-[160px] whitespace-nowrap">
                {{ formatDate(task.created_at) }}
              </td>
              <td class="table-cell w-[160px] whitespace-nowrap text-gray-400">
                {{ task.completed_at ? formatDate(task.completed_at) : "—" }}
              </td>
              <td class="table-cell min-w-[120px] max-w-[200px]">
                <span
                  v-if="task.result"
                  class="tooltip"
                  :data-tooltip="task.result"
                >
                  {{ truncateText(task.result, 20) }}
                </span>
                <span v-else class="text-gray-500">—</span>
              </td>
              <td class="table-cell w-[160px] whitespace-nowrap">
                <div class="flex gap-2">
                  <button
                    @click="$emit('toggle-task', task)"
                    :disabled="task.status === 'running'"
                    class="task-button"
                    :class="
                      task.status === 'running'
                        ? 'bg-gray-700/50 text-gray-400 cursor-not-allowed'
                        : 'bg-blue-500/30 hover:bg-blue-600/40 text-blue-100'
                    "
                  >
                    <i
                      :class="
                        task.status === 'running'
                          ? 'ri-loader-2-line animate-spin'
                          : 'ri-play-line'
                      "
                    ></i>
                    {{ task.status === "running" ? "运行中" : "启动" }}
                  </button>
                  <button
                    @click="handleDelete(task.id)"
                    class="task-button bg-red-500/30 hover:bg-red-600/40 text-red-100"
                  >
                    <i class="ri-delete-bin-line"></i>
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Empty state with improved visuals -->
    <div
      v-else
      class="flex flex-col items-center justify-center py-16 my-4 bg-gray-800/30 rounded-xl border border-gray-700/30 transition-all duration-300"
    >
      <div class="p-6 rounded-full bg-gray-700/30 mb-6">
        <i class="ri-file-list-3-line text-5xl text-gray-500"></i>
      </div>
      <span class="text-xl font-medium text-gray-300 mb-3">暂无任务</span>
      <p class="text-gray-400 mb-6 text-center max-w-md px-4">
        当前还没有创建任何扫描任务，可以在下方创建或前往目标管理添加
      </p>
      <div class="flex gap-4">
        <router-link to="/target-management">
          <button
            class="empty-state-button bg-blue-500/20 hover:bg-blue-600/30 text-blue-300 border-blue-500/30"
          >
            <i class="ri-focus-3-line mr-2"></i>
            前往目标管理
          </button>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "TaskList",
  props: {
    tasks: {
      type: Array,
      required: true,
      default: () => [], // 提供一个默认的空数组
    },
  },
  data() {
    return {
      selectedTasks: [],
    };
  },
  computed: {
    isAllSelected() {
      // 首先检查 tasks 是否存在且是数组
      if (!this.tasks || !Array.isArray(this.tasks)) {
        return false;
      }
      return (
        this.tasks.length > 0 && this.selectedTasks.length === this.tasks.length
      );
    },
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
    getStatusStyle(status) {
      const styleMap = {
        running: "bg-green-500/20 text-green-300 border-green-500/30",
        completed: "bg-blue-500/20 text-blue-300 border-blue-500/30",
        pending: "bg-yellow-500/20 text-yellow-300 border-yellow-500/30",
      };
      return (
        styleMap[status] || "bg-gray-500/20 text-gray-300 border-gray-500/30"
      );
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
    toggleSelectAll() {
      if (this.isAllSelected) {
        this.selectedTasks = [];
      } else {
        this.selectedTasks = this.tasks.map((task) => task.id);
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
/* Action buttons */
.action-button {
  @apply px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200
  focus:outline-none focus:ring-2 shadow-sm border flex items-center;
}

/* Table styling */
.table-header {
  @apply text-left py-3 px-4 text-sm font-medium text-gray-300 sticky top-0;
}

.table-cell {
  @apply py-3 px-4 text-sm text-gray-200;
}

/* Checkbox styling */
.checkbox-input {
  @apply rounded-md border-gray-600 bg-gray-700/50 text-blue-500
  focus:ring-blue-500/30 transition-all duration-200 cursor-pointer h-4 w-4;
}

/* Status badge */
.status-badge {
  @apply px-3 py-1 rounded-md text-xs font-medium flex items-center justify-center
  inline-flex border whitespace-nowrap max-w-fit;
}

/* Task buttons */
.task-button {
  @apply px-2 py-1 rounded-md text-xs font-medium transition-all duration-200
  focus:outline-none focus:ring-1 flex items-center gap-1;
}

/* Empty state button */
.empty-state-button {
  @apply px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200
  focus:outline-none focus:ring-2 border flex items-center;
}

/* Tooltip */
.tooltip {
  @apply relative cursor-help;
}

.tooltip:hover::after {
  content: attr(data-tooltip);
  @apply absolute left-0 top-full mt-1 p-2 bg-gray-800 text-white text-xs rounded-md z-10
  whitespace-normal w-max max-w-xs shadow-lg;
}

/* Animation */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fadeIn {
  animation: fadeIn 0.3s ease-out forwards;
}

/* Custom scrollbar */
.scrollbar-thin {
  scrollbar-width: thin;
  scrollbar-color: rgba(107, 114, 128, 0.3) transparent;
}

.scrollbar-thin::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: rgba(107, 114, 128, 0.3);
  border-radius: 3px;
}

.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background-color: rgba(107, 114, 128, 0.5);
}
</style>
