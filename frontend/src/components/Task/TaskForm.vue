<template>
  <!-- 标题和说明 -->
  <div >
    <h2
      
    >
      <i class="ri-add-line"></i>
      创建任务
    </h2>
    <p >填写以下信息以创建新的扫描任务</p>
  </div>

  <div >
    <!-- 选择任务类型 -->
    <div>
      <label
        
      >
        <i class="ri-bar-chart-horizontal-line .5"></i>
        任务类型
      </label>
      <div >
        <select
          v-model="newTaskType"
          class="border focus: duration-200 appearance-none"
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
          class="inset-y-0"
        >
          <i :class="getTaskTypeIcon(newTaskType)" ></i>
        </div>
        <div
          class="inset-y-0"
        >
          <i class="ri-arrow-down-s-line"></i>
        </div>
      </div>
      <div class=".5" v-if="newTaskType">
        {{ getTaskTypeDescription(newTaskType) }}
      </div>
    </div>

    <!-- 输入目标地址 -->
    <div>
      <label
        
      >
        <i class="ri-focus-3-line .5"></i>
        目标地址
      </label>
      <div >
        <input
          v-model="newTaskAddress"
          type="text"
          :placeholder="getAddressPlaceholder(newTaskType)"
          class="border focus: duration-200"
        />
        <div
          class="inset-y-0"
        >
          <i class="ri-global-line"></i>
        </div>
      </div>
    </div>

    <!-- 创建按钮 -->
    <button
      @click="handleCreateTask"
      :disabled="!isValidInput"
      class="duration-200 group"
      :class="isValidInput ? ' hover: border ' : ' border '"
    >
      <i
        class="ri-add-circle-line group-"
        v-if="isValidInput"
      ></i>
      <i class="ri-error-warning-line" v-else></i>
      {{ isValidInput ? "创建任务" : "请完善任务信息" }}
    </button>
  </div>

  <!-- 任务类型快速选择 -->
  <div >
    <p >快速选择任务类型:</p>
    <div >
      <button
        v-for="(label, value) in taskTypes"
        :key="value"
        @click="selectTaskType(value)"
        class="type-button border duration-200"
        :class="newTaskType === value ? ' ' : ' hover:'"
      >
        <i :class="getTaskTypeIcon(value)" ></i>
        <span >{{ label }}</span>
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
