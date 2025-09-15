// useSubdomainScan.js
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axiosInstance";
import { useNotification } from "./useNotification";
import { useConfirmDialog } from "./useConfirmDialog";
import { getHttpStatusClass } from "./constants";
import { transformSubdomainData, sortByIpAndDomain } from "./utils";

export function useSubdomainScan() {
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
  const selectedSubdomains = ref([]);
  const selectAll = ref(false);
  const isResolving = ref(false);
  const isProbing = ref(false);

  // 监听选中状态
  watch(selectedSubdomains, (newVal) => {
    selectAll.value =
      newVal.length === subdomains.value.length && newVal.length > 0;
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

  // 计算属性 - 子域名列表
  const subdomains = computed(() => {
    if (!scanResult.value?.data) return [];

    const subdomainGroup = scanResult.value.data.find(
      (group) => group.Key === "subdomains"
    );
    if (!subdomainGroup?.Value?.length) return [];

    const domainList = subdomainGroup.Value.map(transformSubdomainData);
    const sorted = domainList.sort(sortByIpAndDomain);

    const seenIPs = new Set();
    return sorted.map((item) => ({
      ...item,
      isFirstIP: item.ip && !seenIPs.has(item.ip) && seenIPs.add(item.ip),
    }));
  });

  // 切换全选
  const toggleSelectAll = () => {
    selectedSubdomains.value = selectAll.value
      ? subdomains.value.map((s) => s.id)
      : [];
  };

  // 切换已读状态
  const toggleReadStatus = async (subdomain) => {
    try {
      await api.put(
        `/results/${route.params.id}/entries/${subdomain.id}/read`,
        { is_read: !subdomain.is_read }
      );
      await fetchScanResult(route.params.id);
      showSuccess(`已${subdomain.is_read ? "标记为未读" : "标记为已读"}`);
    } catch (error) {
      showError("更新状态失败");
    }
  };

  // 解析IP
  const resolveIPs = async (input) => {
    const targets = Array.isArray(input)
      ? input.map((item) => item.id || item)
      : [input.id || input];

    if (!targets.length) {
      showWarning("请先选择子域名");
      return;
    }

    const confirmed = await confirm({
      title: targets.length > 1 ? "批量解析IP" : "解析IP",
      message:
        targets.length > 1
          ? `是否解析选中的 ${targets.length} 个子域名的IP？`
          : `是否解析 ${input.domain} 的IP地址？`,
      type: "info",
    });

    if (!confirmed) return;

    try {
      isResolving.value = true;
      await api.put(`/results/${route.params.id}/entries/resolve`, {
        entryIds: targets,
      });
      await fetchScanResult(route.params.id);
      showSuccess(targets.length > 1 ? "批量解析成功" : "IP解析成功");
    } catch (error) {
      showError(targets.length > 1 ? "批量解析失败" : "IP解析失败");
    } finally {
      isResolving.value = false;
    }
  };

  // 发送到端口扫描
  const sendToPortScan = async (input) => {
    const targets = Array.isArray(input)
      ? input
          .map((id) => subdomains.value.find((sub) => sub.id === id))
          .filter((subdomain) => subdomain?.ip)
      : [input];

    const uniqueIPs = [...new Set(targets.map((subdomain) => subdomain.ip))];

    if (!uniqueIPs.length) {
      showWarning("没有可用的IP");
      return;
    }

    if (!scanResult.value?.target_id) {
      showWarning("无法获取目标ID");
      return;
    }

    const confirmed = await confirm({
      title: targets.length > 1 ? "批量发送到端口扫描" : "发送到端口扫描",
      message:
        targets.length > 1
          ? `是否将选中的 ${uniqueIPs.length} 个IP发送到端口扫描？`
          : `是否将 ${targets[0].domain} (${targets[0].ip}) 发送到端口扫描？`,
      type: "info",
    });

    if (!confirmed) return;

    try {
      for (const ip of uniqueIPs) {
        await api.post("/tasks", {
          type: "nmap",
          payload: ip,
          target_id: scanResult.value.target_id,
        });
      }
      showSuccess(
        targets.length > 1
          ? `已发送 ${uniqueIPs.length} 个IP到端口扫描`
          : "已发送到端口扫描"
      );
    } catch (error) {
      showError(targets.length > 1 ? "批量发送失败" : "发送失败");
    }
  };

  const probeHosts = async (input) => {
    const targets = Array.isArray(input)
      ? input.map((item) => item.id || item)
      : [input.id || input];

    if (!targets.length) {
      showWarning("请先选择子域名");
      return;
    }

    const confirmed = await confirm({
      title: targets.length > 1 ? "批量HTTPX探测" : "HTTPX探测",
      message:
        targets.length > 1
          ? `是否对选中的 ${targets.length} 个子域名进行HTTPX探测？`
          : `是否对 ${input.domain} 进行HTTPX探测？`,
      type: "info",
    });

    if (!confirmed) return;

    try {
      isProbing.value = true;
      await api.put(`/results/${route.params.id}/entries/probe`, {
        entryIds: targets,
      });
      await fetchScanResult(route.params.id);
      showSuccess(targets.length > 1 ? "批量探测成功" : "HTTPX探测成功");
    } catch (error) {
      showError(targets.length > 1 ? "批量探测失败" : "HTTPX探测失败");
    } finally {
      isProbing.value = false;
    }
  };

  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      showSuccess("已复制到剪贴板");
    } catch (err) {
      showError("复制失败");
    }
  };

  return {
    // 状态数据
    scanResult,
    subdomains,
    selectedSubdomains,
    selectAll,
    errorMessage,
    isResolving,
    isProbing,

    // 业务操作方法
    fetchScanResult,
    toggleSelectAll,
    toggleReadStatus,
    resolveIPs,
    sendToPortScan,
    probeHosts,
    getHttpStatusClass,
    copyToClipboard,
    confirm,

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
