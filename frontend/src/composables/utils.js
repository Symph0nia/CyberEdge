// utils.js
import { useNotification } from "../composables/useNotification";
import { useConfirmDialog } from "../composables/useConfirmDialog";

export const transformSubdomainData = (subdomainData) => ({
  id: subdomainData.find((item) => item.Key === "_id")?.Value || "",
  domain: subdomainData.find((item) => item.Key === "domain")?.Value || "",
  is_read: subdomainData.find((item) => item.Key === "is_read")?.Value || false,
  ip: subdomainData.find((item) => item.Key === "ip")?.Value || "",
  httpStatus:
    subdomainData.find((item) => item.Key === "http_status")?.Value || null,
  httpTitle:
    subdomainData.find((item) => item.Key === "http_title")?.Value || "",
});

export const sortByIpAndDomain = (a, b) => {
  if (a.ip && b.ip) {
    const aIpParts = a.ip.split(".").map(Number);
    const bIpParts = b.ip.split(".").map(Number);

    for (let i = 0; i < 4; i++) {
      if (aIpParts[i] !== bIpParts[i]) {
        return aIpParts[i] - bIpParts[i];
      }
    }
    return a.domain.localeCompare(b.domain);
  }
  if (a.ip) return -1;
  if (b.ip) return 1;
  return a.domain.localeCompare(b.domain);
};

export const handleBatchOperation = async ({
  targets,
  batchTitle,
  singleTitle,
  batchMessage,
  singleMessage,
  apiCall,
  successMessage,
  errorMessage,
  loadingRef = null,
  resetSelection = true,
  selectionRefs = {},
  formatTarget = null,
  beforeOperation = null,
  afterOperation = null,
}) => {
  const isBatch = targets.length > 1;
  const { showSuccess, showError } = useNotification();
  const { confirm } = useConfirmDialog();

  // 格式化目标信息
  const getFormattedTarget = (target) => {
    if (formatTarget) return formatTarget(target);
    return typeof target === "object" ? target.id || target : target;
  };

  try {
    // 操作前检查
    if (beforeOperation) {
      const shouldContinue = await beforeOperation(targets);
      if (!shouldContinue) return;
    }

    // 确认对话框
    const confirmed = await confirm({
      title: isBatch ? batchTitle : singleTitle,
      message: isBatch
        ? typeof batchMessage === "function"
          ? batchMessage(targets)
          : batchMessage
        : typeof singleMessage === "function"
        ? singleMessage(targets[0])
        : singleMessage,
      type: "info",
    });

    if (!confirmed) return;

    // 设置加载状态
    if (loadingRef) loadingRef.value = true;

    // 执行API调用
    const formattedTargets = targets.map(getFormattedTarget);
    await apiCall(formattedTargets);

    // 重置选择状态
    if (isBatch && resetSelection && selectionRefs.selectedItems) {
      selectionRefs.selectedItems.value = [];
      if (selectionRefs.selectAll) {
        selectionRefs.selectAll.value = false;
      }
    }

    // 操作后处理
    if (afterOperation) {
      await afterOperation(formattedTargets);
    }

    // 显示成功消息
    showSuccess(
      typeof successMessage === "function"
        ? successMessage(formattedTargets)
        : successMessage
    );
  } catch (error) {
    // 显示错误消息
    showError(
      typeof errorMessage === "function" ? errorMessage(targets) : errorMessage
    );
    console.error("Batch operation error:", error);
  } finally {
    // 重置加载状态
    if (loadingRef) loadingRef.value = false;
  }
};
