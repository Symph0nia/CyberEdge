<template>
  <!-- 标题和说明 -->
  <div class="mb-6">
    <h2
      class="text-xl font-medium tracking-wide text-gray-200 flex items-center"
    >
      <i class="ri-add-line mr-2 text-blue-400"></i>
      创建任务
    </h2>
    <p class="text-sm text-gray-400 mt-1">填写以下信息以创建新的扫描任务</p>
  </div>

  <div class="space-y-6">
    <!-- 选择任务类型 -->
    <div>
      <label
        class="block text-sm font-medium text-gray-300 mb-2 flex items-center"
      >
        <i class="ri-bar-chart-horizontal-line mr-1.5"></i>
        任务类型
      </label>
      <div class="relative">
        <select
          v-model="newTaskType"
          class="w-full px-4 py-3 rounded-xl bg-gray-700/50 text-gray-100 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all duration-200 appearance-none pl-10"
        >
          <option value="" disabled>选择任务类型</option>
          <option
            v-for="(label, value) in taskTypes"
            :key="value"
            :value="value"
          >
            {{ label }}
          </option>
        </select>
        <div
          class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"
        >
          <i :class="getTaskTypeIcon(newTaskType)" class="text-gray-400"></i>
        </div>
        <div
          class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none"
        >
          <i class="ri-arrow-down-s-line text-gray-400"></i>
        </div>
      </div>
      <div class="mt-1.5 text-xs text-gray-400" v-if="newTaskType">
        {{ getTaskTypeDescription(newTaskType) }}
      </div>
    </div>

    <!-- 输入目标地址 -->
    <div>
      <label
        class="block text-sm font-medium text-gray-300 mb-2 flex items-center"
      >
        <i class="ri-focus-3-line mr-1.5"></i>
        目标地址
      </label>
      <div class="relative">
        <input
          v-model="newTaskAddress"
          type="text"
          :placeholder="getAddressPlaceholder(newTaskType)"
          class="w-full px-4 py-3 rounded-xl bg-gray-700/50 text-gray-100 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all duration-200 pl-10"
        />
        <div
          class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"
        >
          <i class="ri-global-line text-gray-400"></i>
        </div>
      </div>
    </div>

    <!-- 创建按钮 -->
    <button
      @click="handleCreateTask"
      :disabled="!isValidInput"
      class="w-full px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 flex items-center justify-center group"
      :class="
        isValidInput
          ? 'bg-blue-500/30 hover:bg-blue-600/40 text-blue-100 focus:ring-blue-500/30 border border-blue-500/30'
          : 'bg-gray-700/50 text-gray-400 cursor-not-allowed border border-gray-600/30'
      "
    >
      <i
        class="ri-add-circle-line mr-2 group-hover:animate-pulse"
        v-if="isValidInput"
      ></i>
      <i class="ri-error-warning-line mr-2" v-else></i>
      {{ isValidInput ? "创建任务" : "请完善任务信息" }}
    </button>
  </div>

  <!-- 任务类型快速选择 -->
  <div class="mt-8 pt-6 border-t border-gray-700/30">
    <p class="text-sm text-gray-400 mb-4">快速选择任务类型:</p>
    <div class="grid grid-cols-3 gap-3">
      <button
        v-for="(label, value) in taskTypes"
        :key="value"
        @click="selectTaskType(value)"
        class="type-button flex flex-col items-center justify-center p-3 rounded-lg border transition-all duration-200"
        :class="
          newTaskType === value
            ? 'bg-blue-500/20 border-blue-500/30 text-blue-300'
            : 'bg-gray-800/50 border-gray-700/30 text-gray-300 hover:bg-gray-700/50'
        "
      >
        <i :class="getTaskTypeIcon(value)" class="text-xl mb-2"></i>
        <span class="text-xs font-medium">{{ label }}</span>
      </button>
    </div>
  </div>
</template>

<script>
import { ref, computed } from "vue";
import { useNotification } from "../../composables/useNotification";

export default {
  name: "TaskForm",
  setup(props, { emit }) {
    // 使用通知钩子
    const { showSuccess, showError } = useNotification();

    // 表单数据
    const newTaskType = ref("");
    const newTaskAddress = ref("");

    // 任务类型选项
    const taskTypes = {
      subfinder: "子域名扫描",
      nmap: "端口扫描",
      ffuf: "路径扫描",
    };

    // 获取任务类型图标
    const getTaskTypeIcon = (type) => {
      const icons = {
        subfinder: "ri-global-line",
        nmap: "ri-scan-2-line",
        ffuf: "ri-folders-line",
        "": "ri-question-line",
      };
      return icons[type] || "ri-question-line";
    };

    // 获取任务类型描述
    const getTaskTypeDescription = (type) => {
      const descriptions = {
        subfinder: "扫描目标域名的所有子域名，帮助发现攻击面",
        nmap: "扫描目标主机开放的端口和服务信息",
        ffuf: "对Web应用进行路径扫描，发现隐藏资源",
      };
      return descriptions[type] || "";
    };

    // 获取地址输入框的占位文本
    const getAddressPlaceholder = (type) => {
      const placeholders = {
        subfinder: "example.com",
        nmap: "192.168.1.1 或 example.com",
        ffuf: "https://example.com/",
        "": "输入目标地址",
      };
      return placeholders[type] || "输入目标地址";
    };

    // 选择任务类型
    const selectTaskType = (type) => {
      newTaskType.value = type;
    };

    // 输入验证
    const isValidInput = computed(() => {
      return (
        newTaskType.value.trim() !== "" && newTaskAddress.value.trim() !== ""
      );
    });

    // 创建任务处理
    const handleCreateTask = () => {
      if (!isValidInput.value) {
        showError("请填写完整信息");
        return;
      }

      try {
        emit("create-task", {
          type: newTaskType.value,
          payload: newTaskAddress.value,
        });

        // 重置表单
        newTaskType.value = "";
        newTaskAddress.value = "";

        showSuccess("任务已创建");
      } catch (error) {
        showError("创建任务失败");
      }
    };

    return {
      newTaskType,
      newTaskAddress,
      taskTypes,
      isValidInput,
      handleCreateTask,
      getTaskTypeIcon,
      getTaskTypeDescription,
      getAddressPlaceholder,
      selectTaskType,
    };
  },
};
</script>

<style scoped>
.type-button {
  transition: transform 0.2s, box-shadow 0.2s;
}

.type-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.2);
}

.type-button:active {
  transform: translateY(0);
}

/* 输入框焦点效果 */
input:focus,
select:focus {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1),
    0 0 0 4px rgba(59, 130, 246, 0.1);
}
</style>
