// src/composables/useTargetManagement.js
import { ref, onMounted } from "vue";
import api from "../api/axiosInstance";
import router from "../router";
import { useNotification } from "./useNotification";
import { useConfirmDialog } from "./useConfirmDialog";

export function useTargetManagement() {
  const {
    showNotification,
    notificationMessage,
    notificationType,
    showSuccess,
    showError,
  } = useNotification();

  const {
    confirm,
    showDialog: showConfirmDialog,
    dialogTitle,
    dialogMessage,
    dialogType,
    handleConfirm,
    handleCancel,
  } = useConfirmDialog();

  // 状态管理
  const targets = ref([]); // 将 null 改为空数组
  const isLoading = ref(false);
  const isSubmitting = ref(false);
  const showDialog = ref(false);
  const dialogMode = ref("create"); // 'create' 或 'edit'

  // 表单数据
  const targetForm = ref({
    name: "",
    description: "",
    type: "domain", // 'domain' 或 'ip'
    target: "", // 具体的域名或IP地址
  });

  // 获取目标列表
  const fetchTargets = async () => {
    isLoading.value = true;
    try {
      const response = await api.get("/targets");
      targets.value = Array.isArray(response.data) ? response.data : []; // 确保是数组
      showSuccess("加载目标列表成功");
    } catch (error) {
      targets.value = []; // 出错时设置为空数组
      showError("加载目标列表失败");
    } finally {
      isLoading.value = false;
    }
  };

  // 打开创建对话框
  const openCreateDialog = () => {
    dialogMode.value = "create";
    targetForm.value = {
      name: "",
      description: "",
      type: "domain",
      target: "",
    };
    showDialog.value = true;
  };

  // 编辑目标
  const editTarget = (target) => {
    console.log("Editing target:", target); // 调试日志
    dialogMode.value = "edit";
    targetForm.value = {
      _id: target._id || target.id, // 保存 ID
      name: target.name,
      description: target.description || "",
      type: target.type,
      target: target.target,
    };
    showDialog.value = true;
  };

  // 删除目标
  const deleteTarget = async (target) => {
    try {
      const confirmed = await confirm({
        title: "确认删除",
        message: `是否确认删除目标 "${target.name}"？此操作不可撤销。`,
        type: "danger",
      });

      if (confirmed) {
        await api.delete(`/targets/${target.id}`);
        await fetchTargets();
        showSuccess("删除目标成功");
      }
    } catch (error) {
      showError("删除目标失败");
    }
  };

  // 归档/激活目标
  const archiveTarget = async (target) => {
    try {
      const newStatus = target.status === "active" ? "archived" : "active";
      const actionName = newStatus === "active" ? "激活" : "归档";

      // 显示确认对话框
      const confirmed = await confirm({
        title: `确认${actionName}`,
        message: `是否确认${actionName}目标 "${target.name}"？`,
        type: "warning",
      });

      if (confirmed) {
        // 准备更新的数据
        const updatedData = {
          ...target,
          status: newStatus,
          updatedAt: new Date().toISOString(),
        };

        // 发送更新请求
        await api.put(`/targets/${target.id}`, updatedData);

        // 刷新目标列表
        await fetchTargets();

        showSuccess(`${actionName}目标成功`);
      }
    } catch (error) {
      showError(`${target.status === "active" ? "归档" : "激活"}目标失败`);
    }
  };

  // 开始扫描目标
  const startScan = async (target) => {
    const scanType = target.type === "domain" ? "subfinder" : "nmap";
    const scanTypeName = target.type === "domain" ? "子域名扫描" : "端口扫描";

    try {
      // 显示确认对话框
      const confirmed = await confirm({
        title: `开始${scanTypeName}`,
        message: `是否对 ${target.name} (${target.target}) 进行${scanTypeName}？`,
        type: "info",
      });

      if (!confirmed) return;

      // 发送扫描请求，使用 target_id 替代 parent_id
      await api.post("/tasks", {
        type: scanType,
        payload: target.target,
        target_id: target.id, // 修改这里
      });

      // 显示成功消息
      showSuccess(`已发送到${scanTypeName}`);
    } catch (error) {
      // 显示错误消息
      showError(`发送到${scanTypeName}失败`);
    }
  };

  // 提交表单
  const submitTargetForm = async (formData) => {
    if (isSubmitting.value) return;

    isSubmitting.value = true;
    try {
      const submitData = {
        name: formData.name,
        description: formData.description,
        type: formData.type,
        target: formData.target,
        status: "active", // 设置默认状态
      };

      if (dialogMode.value === "create") {
        await api.post("/targets", submitData);
      } else {
        // 编辑模式
        const id = targetForm.value._id || targetForm.value.id;
        await api.put(`/targets/${id}`, submitData);
      }

      showDialog.value = false;
      await fetchTargets(); // 刷新列表
      showSuccess(`${dialogMode.value === "create" ? "创建" : "更新"}目标成功`);
    } catch (error) {
      console.error("Submit error:", error.response?.data || error);
      showError(
        `${dialogMode.value === "create" ? "创建" : "更新"}目标失败: ${
          error.response?.data?.message || "未知错误"
        }`
      );
    } finally {
      isSubmitting.value = false;
    }
  };

  // 添加查看详情函数
  const viewDetails = async (target) => {
    try {
      // 使用 router.push 跳转到详情页面，传递目标 ID
      await router.push({
        name: "TargetDetail",
        params: {
          id: target.id || target._id, // 确保兼容两种 ID 字段
        },
      });
    } catch (error) {
      showError("跳转到详情页面失败");
    }
  };

  // 在组件挂载时获取数据
  onMounted(() => {
    fetchTargets();
  });

  return {
    targets,
    isLoading,
    isSubmitting,
    targetForm,
    dialogMode,
    showDialog,
    showNotification,
    notificationMessage,
    notificationType,
    showConfirmDialog,
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
  };
}
