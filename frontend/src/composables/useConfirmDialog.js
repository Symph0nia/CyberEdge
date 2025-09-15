// src/composables/useConfirmDialog.js
import { ref } from "vue";

export function useConfirmDialog() {
  // 对话框状态
  const showDialog = ref(false);
  const dialogTitle = ref("");
  const dialogMessage = ref("");
  const dialogType = ref("info");

  // 存储回调函数
  let resolvePromise = null;

  /**
   * 显示确认对话框
   * @param {Object} options - 对话框配置选项
   * @param {string} options.title - 对话框标题
   * @param {string} options.message - 对话框消息
   * @param {('info'|'warning'|'danger')} [options.type='info'] - 对话框类型
   * @returns {Promise<boolean>} - 用户确认返回 true，取消返回 false
   */
  const confirm = ({ title = "确认", message, type = "info" }) => {
    return new Promise((resolve) => {
      showDialog.value = true;
      dialogTitle.value = title;
      dialogMessage.value = message;
      dialogType.value = type;
      resolvePromise = resolve;
    });
  };

  /**
   * 处理确认操作
   */
  const handleConfirm = () => {
    showDialog.value = false;
    if (resolvePromise) {
      resolvePromise(true);
      resolvePromise = null;
    }
  };

  /**
   * 处理取消操作
   */
  const handleCancel = () => {
    showDialog.value = false;
    if (resolvePromise) {
      resolvePromise(false);
      resolvePromise = null;
    }
  };

  return {
    // 状态
    showDialog,
    dialogTitle,
    dialogMessage,
    dialogType,

    // 方法
    confirm,
    handleConfirm,
    handleCancel,
  };
}
