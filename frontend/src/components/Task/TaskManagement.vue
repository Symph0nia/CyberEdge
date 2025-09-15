<template>
  <div class="task-management">
    <HeaderPage />

    <div class="management-container">
      <!-- 任务管理概览 -->
      <div class="page-header">
        <h1 class="page-title">
          <i class="ri-task-line"></i>
          任务管理中心
        </h1>
        <a-button
          @click="handleRefreshTasks"
          :loading="isLoading"
          class="refresh-btn"
        >
          <i class="ri-refresh-line"></i>
          {{ isLoading ? '加载中' : '刷新任务' }}
        </a-button>
      </div>

      <!-- 卡片布局：左侧任务列表，右侧创建表单 -->
      <a-row :gutter="[24, 24]">
        <!-- 左侧任务列表区域 -->
        <a-col :xs="24" :lg="16">
          <a-card class="task-list-card">
            <TaskList
              :tasks="tasks"
              @toggle-task="toggleTask"
              @delete-task="handleDelete"
              @refresh-tasks="handleRefreshTasks"
              @batch-start="handleBatchStart"
              @batch-delete="handleBatchDelete"
            />
          </a-card>
        </a-col>

        <!-- 右侧任务创建表单 -->
        <a-col :xs="24" :lg="8">
          <a-card class="task-form-card">
            <template #title>
              <div class="form-title">
                <i class="ri-add-circle-line"></i>
                创建新任务
              </div>
            </template>
            <TaskForm @create-task="createTask" />
          </a-card>
        </a-col>
      </a-row>

      <!-- 快捷操作浮动按钮 -->
      <a-float-button-group trigger="click" class="float-buttons">
        <a-float-button @click="scrollToTop" tooltip="返回顶部">
          <template #icon><i class="ri-arrow-up-line"></i></template>
        </a-float-button>
        <a-float-button @click="scrollToForm" tooltip="创建任务" type="primary">
          <template #icon><i class="ri-add-line"></i></template>
        </a-float-button>
      </a-float-button-group>
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

<script>
import { ref, onMounted } from "vue";
import TaskList from "./TaskList.vue";
import TaskForm from "./TaskForm.vue";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import ConfirmDialog from "../Utils/ConfirmDialog.vue";
import { useNotification } from "../../composables/useNotification";
import { useConfirmDialog } from "../../composables/useConfirmDialog";
import api from "../../api/axiosInstance";

export default {
  name: "TaskManagement",
  components: {
    HeaderPage,
    FooterPage,
    TaskList,
    TaskForm,
    PopupNotification,
    ConfirmDialog,
  },
  setup() {
    const tasks = ref([]);
    const isLoading = ref(false);

    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    const {
      showDialog,
      dialogTitle,
      dialogMessage,
      dialogType,
      confirm,
      handleConfirm,
      handleCancel,
    } = useConfirmDialog();

    // 获取任务列表
    const fetchTasks = async () => {
      try {
        isLoading.value = true;
        const response = await api.get("/tasks");
        tasks.value = response.data;
      } catch (error) {
        showError("获取任务列表失败");
      } finally {
        isLoading.value = false;
      }
    };

    // 创建任务
    const createTask = async (taskData) => {
      try {
        const response = await api.post("/tasks", taskData);

        if (response.data && response.data.id) {
          // 避免完全刷新列表，只添加新任务
          tasks.value.unshift(response.data);
          showSuccess("已创建新任务");

          // 滚动到顶部查看新任务
          setTimeout(() => scrollToTop(), 300);
        } else {
          await fetchTasks(); // 备用：如果响应格式不一致则刷新整个列表
          showSuccess("已创建新任务");
        }
      } catch (error) {
        showError(error.response?.data?.message || "创建任务失败");
      }
    };

    // 切换单个任务状态
    const toggleTask = async (task) => {
      try {
        const response = await api.post("/tasks/start", {
          taskIds: [task.id],
        });
        const result = response.data.result;

        // 乐观更新UI
        const taskIndex = tasks.value.findIndex((t) => t.id === task.id);
        if (taskIndex !== -1) {
          tasks.value[taskIndex].status = "running";
        }

        if (result.success.includes(task.id)) {
          showSuccess("已启动任务");
        } else {
          const errorMsg = result.failed[task.id] || "启动失败";
          showError(`启动任务失败: ${errorMsg}`);
          await fetchTasks(); // 恢复实际状态
        }
      } catch (error) {
        showError("启动任务失败");
        await fetchTasks(); // 恢复实际状态
      }
    };

    // 删除单个任务
    const handleDelete = async (taskId) => {
      try {
        const confirmed = await confirm({
          title: "确认删除",
          message: `是否确认删除任务 ${taskId}？此操作不可撤销。`,
          type: "danger",
        });

        if (confirmed) {
          // 乐观更新UI
          tasks.value = tasks.value.filter((task) => task.id !== taskId);

          const response = await api.delete("/tasks", {
            data: { taskIds: [taskId] },
          });
          const result = response.data.result;

          if (result.success.includes(taskId)) {
            showSuccess("已删除任务");
          } else {
            const errorMsg = result.failed[taskId] || "删除失败";
            showError(`删除任务失败: ${errorMsg}`);
            await fetchTasks(); // 恢复实际状态
          }
        }
      } catch (error) {
        showError("删除任务失败");
        await fetchTasks(); // 恢复实际状态
      }
    };

    // 批量启动任务
    const handleBatchStart = async (taskIds) => {
      if (taskIds.length === 0) {
        showError("请选择要启动的任务");
        return;
      }

      try {
        const confirmed = await confirm({
          title: "确认批量启动",
          message: `是否确认启动选中的 ${taskIds.length} 个任务？`,
          type: "warning",
        });

        if (confirmed) {
          // 乐观更新UI
          tasks.value = tasks.value.map((task) => {
            if (taskIds.includes(task.id)) {
              return { ...task, status: "running" };
            }
            return task;
          });

          const response = await api.post("/tasks/start", { taskIds });
          const result = response.data.result;

          if (result.success.length > 0) {
            showSuccess(`成功启动 ${result.success.length} 个任务`);
          }

          if (Object.keys(result.failed).length > 0) {
            showError(`${Object.keys(result.failed).length} 个任务启动失败`);
            await fetchTasks(); // 恢复实际状态
          }
        }
      } catch (error) {
        showError("批量启动任务失败");
        await fetchTasks(); // 恢复实际状态
      }
    };

    // 批量删除任务
    const handleBatchDelete = async (taskIds) => {
      if (taskIds.length === 0) {
        showError("请选择要删除的任务");
        return;
      }

      try {
        const confirmed = await confirm({
          title: "确认批量删除",
          message: `是否确认删除选中的 ${taskIds.length} 个任务？此操作不可撤销。`,
          type: "danger",
        });

        if (confirmed) {
          // 乐观更新UI
          tasks.value = tasks.value.filter(
            (task) => !taskIds.includes(task.id)
          );

          const response = await api.delete("/tasks", {
            data: { taskIds },
          });
          const result = response.data.result;

          if (result.success.length > 0) {
            showSuccess(`成功删除 ${result.success.length} 个任务`);
          }

          if (Object.keys(result.failed).length > 0) {
            showError(`${Object.keys(result.failed).length} 个任务删除失败`);
            await fetchTasks(); // 恢复实际状态
          }
        }
      } catch (error) {
        showError("批量删除任务失败");
        await fetchTasks(); // 恢复实际状态
      }
    };

    // 刷新任务列表
    const handleRefreshTasks = async () => {
      try {
        await fetchTasks();
        showSuccess("已刷新任务列表");
      } catch (error) {
        showError("刷新任务列表失败");
      }
    };

    // 滚动到顶部
    const scrollToTop = () => {
      window.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    };

    // 滚动到表单
    const scrollToForm = () => {
      const formElement = document.querySelector(".sticky");
      if (formElement) {
        formElement.scrollIntoView({ behavior: "smooth" });
      }
    };

    onMounted(fetchTasks);

    return {
      tasks,
      isLoading,
      showNotification,
      notificationMessage,
      notificationType,
      showDialog,
      dialogTitle,
      dialogMessage,
      dialogType,
      handleConfirm,
      handleCancel,
      createTask,
      toggleTask,
      handleDelete,
      handleRefreshTasks,
      handleBatchStart,
      handleBatchDelete,
      scrollToTop,
      scrollToForm,
    };
  },
};
</script>

<style scoped>
/* 网络安全主题样式 */
.task-management {
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

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
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

/* 卡片样式 */
:deep(.task-list-card.ant-card),
:deep(.task-form-card.ant-card) {
  background: rgba(31, 41, 55, 0.4);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 16px;
  backdrop-filter: blur(12px);
  transition: all 0.3s ease;
}

:deep(.task-list-card.ant-card:hover),
:deep(.task-form-card.ant-card:hover) {
  border-color: rgba(34, 211, 238, 0.5);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
  transform: translateY(-2px);
}

:deep(.task-list-card .ant-card-body),
:deep(.task-form-card .ant-card-body) {
  padding: 24px;
}

/* 表单卡片标题 */
.form-title {
  display: flex;
  align-items: center;
  color: #ffffff;
  font-size: 16px;
  font-weight: 600;
}

.form-title i {
  margin-right: 8px;
  color: #22d3ee;
  font-size: 18px;
}

:deep(.task-form-card .ant-card-head) {
  background: rgba(75, 85, 99, 0.2);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 16px 16px 0 0;
}

:deep(.task-form-card .ant-card-head-title) {
  color: #ffffff;
}

/* 按钮样式 */
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

/* 浮动按钮 */
:deep(.float-buttons) {
  position: fixed;
  bottom: 24px;
  right: 24px;
}

/* 滚动平滑效果 */
html {
  scroll-behavior: smooth;
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
}
</style>
