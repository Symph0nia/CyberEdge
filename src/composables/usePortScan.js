// usePortScan.js
import { ref, computed, watch } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axiosInstance";
import { useNotification } from "./useNotification";
import { useConfirmDialog } from "./useConfirmDialog";
import { getHttpStatusClass } from "./constants";

export function usePortScan() {
  const route = useRoute();

  const {
    showNotification,
    notificationMessage,
    notificationType,
    showSuccess,
    showError,
    showWarning,
  } = useNotification();

  const {
    confirm,
    showDialog,
    dialogTitle,
    dialogMessage,
    dialogType,
    handleConfirm,
    handleCancel,
  } = useConfirmDialog();

  // 状态管理
  const scanResult = ref(null);
  const errorMessage = ref("");
  const selectedPorts = ref([]);
  const selectAll = ref(false);
  const isProbing = ref(false);

  // 监听选中状态
  watch(selectedPorts, (newVal) => {
    selectAll.value =
      newVal.length === filteredPorts.value.length && newVal.length > 0;
  });

  // 获取扫描结果
  const fetchScanResult = async (id) => {
    try {
      const response = await api.get(`/results/${id}`);
      scanResult.value = response.data;
      errorMessage.value = "";
    } catch (error) {
      errorMessage.value = "获取扫描结果详情失败";
      showError("获取扫描结果详情失败");
    }
  };

  // 工具函数 - 获取端口值
  const getPortValue = (port, key) => {
    if (!port) return "";
    if (Array.isArray(port)) {
      const item = port.find((p) => p.Key === key);
      return item ? item.Value : "";
    }
    if (typeof port === "object") {
      return port[key] || "";
    }
    return "";
  };

  // 计算属性 - 端口列表
  const filteredPorts = computed(() => {
    if (!scanResult.value?.data) return [];
    const portGroup = scanResult.value.data.find(
      (group) => group.Key === "ports"
    );
    return portGroup?.Value || [];
  });

  // 切换全选
  const toggleSelectAll = () => {
    selectedPorts.value = selectAll.value
      ? filteredPorts.value.map((port) => getPortValue(port, "_id"))
      : [];
  };

  // 切换已读状态
  const toggleReadStatus = async (port) => {
    const portID = getPortValue(port, "_id");
    const currentStatus = getPortValue(port, "is_read");
    try {
      await api.put(`/results/${route.params.id}/entries/${portID}/read`, {
        is_read: !currentStatus,
      });
      await fetchScanResult(route.params.id);
      showSuccess(`已${currentStatus ? "标记为未读" : "标记为已读"}`);
    } catch (error) {
      showError("更新状态失败");
    }
  };

  // 发送到路径扫描
  const sendToPathScan = async (input) => {
    const targets = Array.isArray(input)
      ? input
          .map((id) =>
            filteredPorts.value.find((port) => getPortValue(port, "_id") === id)
          )
          .filter((port) => port)
      : [input];

    if (!targets.length) {
      showWarning("请先选择要扫描的端口");
      return;
    }

    if (!scanResult.value?.target_id) {
      showWarning("无法获取目标ID");
      return;
    }

    const confirmed = await confirm({
      title: targets.length > 1 ? "批量发送到路径扫描" : "发送到路径扫描",
      message:
        targets.length > 1
          ? `是否将选中的 ${targets.length} 个端口发送到路径扫描？`
          : `是否将端口 ${getPortValue(targets[0], "number")} 发送到路径扫描？`,
      type: "info",
    });

    if (!confirmed) return;

    try {
      for (const port of targets) {
        const portNumber = getPortValue(port, "number");
        const host = getPortValue(port, "host");
        const protocol =
          getPortValue(port, "service") === "https" ? "https" : "http";

        await api.post("/tasks", {
          type: "ffuf",
          payload: `${protocol}://${host}:${portNumber}`,
          target_id: scanResult.value.target_id,
        });
      }
      showSuccess(
        targets.length > 1
          ? `已发送 ${targets.length} 个端口到路径扫描`
          : "已发送到路径扫描"
      );
    } catch (error) {
      showError(targets.length > 1 ? "批量发送失败" : "发送失败");
    }
  };

  // 探测端口HTTP服务
  const probePort = async (input) => {
    // 找到对应的完整端口数据
    const portDetails = Array.isArray(input)
      ? input.map((id) =>
          filteredPorts.value.find((port) => getPortValue(port, "_id") === id)
        )
      : [input];

    const targets = portDetails.map((port) => getPortValue(port, "_id"));

    if (!targets.length) {
      showWarning("请先选择端口");
      return;
    }

    const confirmed = await confirm({
      title: targets.length > 1 ? "批量HTTP探测" : "HTTP探测",
      message:
        targets.length > 1
          ? `是否对选中的 ${targets.length} 个端口进行HTTP探测？`
          : `是否对端口 ${getPortValue(
              portDetails[0],
              "number"
            )} 进行HTTP探测？`,
      type: "info",
    });

    if (!confirmed) return;

    try {
      isProbing.value = true;
      await api.put(`/results/${route.params.id}/entries/probe`, {
        entryIds: targets, // 修改为 entryIds
      });
      await fetchScanResult(route.params.id);
      showSuccess(targets.length > 1 ? "批量探测成功" : "HTTP探测成功");
    } catch (error) {
      showError(targets.length > 1 ? "批量探测失败" : "HTTP探测失败");
    } finally {
      isProbing.value = false;
    }
  };

  return {
    // 状态数据
    scanResult,
    errorMessage,
    selectedPorts,
    selectAll,
    filteredPorts,
    isProbing,

    // 业务操作方法
    fetchScanResult,
    getPortValue,
    toggleSelectAll,
    toggleReadStatus,
    sendToPathScan,
    probePort,
    getHttpStatusClass,

    // UI控制 - 通知
    showNotification,
    notificationMessage,
    notificationType,

    // UI控制 - 确认对话框
    showDialog,
    dialogTitle,
    dialogMessage,
    dialogType,
    handleConfirm,
    handleCancel,
  };
}
