<template>
  <div class="crypto-tools">
    <div class="tools-header">
      <h2>
        <i class="ri-lock-line"></i>
        加密解密工具
      </h2>
    </div>

    <!-- 工具列表 -->
    <a-list :data-source="tools" size="small" class="tools-list">
      <template #renderItem="{ item }">
        <a-list-item class="tool-item" @click="selectTool(item)">
          <a-list-item-meta>
            <template #avatar>
              <i :class="item.icon" class="tool-icon"></i>
            </template>
            <template #title>
              <span class="tool-name">{{ item.name }}</span>
            </template>
          </a-list-item-meta>
        </a-list-item>
      </template>
    </a-list>

    <!-- 操作面板 -->
    <a-card
      v-if="selectedTool"
      :title="selectedTool.name"
      size="small"
      class="tool-panel"
    >
      <template #extra>
        <a-button type="text" size="small" @click="closeTool">
          <i class="ri-close-line"></i>
        </a-button>
      </template>

      <a-form layout="vertical" :model="formData" class="tool-form">
        <!-- 输入文本 -->
        <a-form-item label="输入文本">
          <a-textarea
            v-model:value="formData.inputText"
            placeholder="请输入需要处理的文本"
            :rows="4"
            class="input-area"
          />
        </a-form-item>

        <!-- AES密钥输入 -->
        <a-form-item
          v-if="needsKey"
          label="密钥"
        >
          <a-input
            v-model:value="formData.key"
            placeholder="请输入密钥"
            class="key-input"
          />
        </a-form-item>

        <!-- 操作按钮 -->
        <a-form-item>
          <a-space>
            <a-button
              type="primary"
              @click="handleAction"
              :disabled="!formData.inputText"
            >
              <i class="ri-play-line"></i>
              执行
            </a-button>
            <a-button @click="clearForm">
              <i class="ri-delete-bin-line"></i>
              清空
            </a-button>
          </a-space>
        </a-form-item>

        <!-- 结果显示 -->
        <a-form-item v-if="result" label="处理结果">
          <a-textarea
            :value="result"
            readonly
            :rows="6"
            class="result-area"
          />
          <div class="result-actions">
            <a-button
              type="dashed"
              size="small"
              @click="copyResult"
              class="copy-btn"
            >
              <i class="ri-clipboard-line"></i>
              复制结果
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script>
import { ref, computed, reactive } from "vue";
import { message } from "ant-design-vue";
import CryptoJS from "crypto-js";

export default {
  name: "CryptoTools",
  setup() {
    const selectedTool = ref(null);
    const result = ref("");

    const formData = reactive({
      inputText: "",
      key: ""
    });

    const tools = [
      {
        name: "Base64 加密",
        action: "base64Encode",
        icon: "ri-lock-password-line",
      },
      {
        name: "Base64 解密",
        action: "base64Decode",
        icon: "ri-lock-unlock-line",
      },
      { name: "AES 加密", action: "aesEncrypt", icon: "ri-lock-2-line" },
      { name: "AES 解密", action: "aesDecrypt", icon: "ri-key-line" },
      { name: "MD5 加密", action: "md5Hash", icon: "ri-fingerprint-line" },
      {
        name: "SHA-256 加密",
        action: "sha256Hash",
        icon: "ri-shield-keyhole-line",
      },
      { name: "URL 编码", action: "urlEncode", icon: "ri-global-line" },
      { name: "URL 解码", action: "urlDecode", icon: "ri-search-line" },
      { name: "Hex 编码", action: "hexEncode", icon: "ri-code-box-line" },
      { name: "Hex 解码", action: "hexDecode", icon: "ri-braces-line" },
    ];

    // 是否需要密钥
    const needsKey = computed(() => {
      return selectedTool.value && ['aesEncrypt', 'aesDecrypt'].includes(selectedTool.value.action);
    });

    // 选择工具
    const selectTool = (tool) => {
      selectedTool.value = tool;
      clearForm();
    };

    // 关闭工具
    const closeTool = () => {
      selectedTool.value = null;
      clearForm();
    };

    // 清空表单
    const clearForm = () => {
      formData.inputText = "";
      formData.key = "";
      result.value = "";
    };

    // 执行加密解密操作
    const handleAction = () => {
      if (!formData.inputText) {
        message.warning('请输入需要处理的文本');
        return;
      }

      try {
        switch (selectedTool.value.action) {
          case "base64Encode":
            result.value = btoa(formData.inputText);
            break;
          case "base64Decode":
            result.value = atob(formData.inputText);
            break;
          case "aesEncrypt":
            if (!formData.key) {
              message.warning('请输入密钥');
              return;
            }
            result.value = CryptoJS.AES.encrypt(formData.inputText, formData.key).toString();
            break;
          case "aesDecrypt": {
            if (!formData.key) {
              message.warning('请输入密钥');
              return;
            }
            const decrypted = CryptoJS.AES.decrypt(formData.inputText, formData.key).toString(CryptoJS.enc.Utf8);
            result.value = decrypted || "解密失败，可能是密钥错误";
            break;
          }
          case "md5Hash":
            result.value = CryptoJS.MD5(formData.inputText).toString();
            break;
          case "sha256Hash":
            result.value = CryptoJS.SHA256(formData.inputText).toString();
            break;
          case "urlEncode":
            result.value = encodeURIComponent(formData.inputText);
            break;
          case "urlDecode":
            result.value = decodeURIComponent(formData.inputText);
            break;
          case "hexEncode":
            result.value = textToHex(formData.inputText);
            break;
          case "hexDecode":
            result.value = hexToText(formData.inputText);
            break;
        }
        message.success('操作完成');
      } catch (error) {
        result.value = `操作失败: ${error.message}`;
        message.error('操作失败');
      }
    };

    // 复制结果
    const copyResult = () => {
      navigator.clipboard.writeText(result.value).then(() => {
        message.success('已复制到剪贴板');
      }).catch(() => {
        message.error('复制失败');
      });
    };

    // 工具函数
    const textToHex = (text) => {
      return text
        .split("")
        .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
        .join("");
    };

    const hexToText = (hex) => {
      if (!/^[0-9a-fA-F]+$/.test(hex)) {
        throw new Error("Invalid hex string");
      }
      return hex
        .match(/.{1,2}/g)
        .map((byte) => String.fromCharCode(parseInt(byte, 16)))
        .join("");
    };

    return {
      selectedTool,
      formData,
      result,
      tools,
      needsKey,
      selectTool,
      closeTool,
      clearForm,
      handleAction,
      copyResult,
    };
  },
};
</script>

<style scoped>
/* 网络安全主题样式 */
.crypto-tools {
  padding: 24px;
  background: transparent;
}

.tools-header h2 {
  display: flex;
  align-items: center;
  color: #ffffff;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  padding-bottom: 12px;
}

.tools-header h2 i {
  margin-right: 8px;
  color: #22d3ee;
  font-size: 20px;
}

/* 工具列表样式 */
:deep(.tools-list .ant-list-item) {
  border-bottom: 1px solid rgba(75, 85, 99, 0.2);
  padding: 8px 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 8px;
  margin-bottom: 4px;
}

:deep(.tools-list .ant-list-item:hover) {
  background: rgba(75, 85, 99, 0.3);
  border-color: rgba(34, 211, 238, 0.4);
}

.tool-icon {
  color: #9ca3af;
  font-size: 16px;
  transition: color 0.3s ease;
}

:deep(.tools-list .ant-list-item:hover) .tool-icon {
  color: #22d3ee;
}

.tool-name {
  color: #d1d5db;
  font-size: 14px;
  font-weight: 500;
}

/* 工具面板样式 */
:deep(.tool-panel.ant-card) {
  background: rgba(31, 41, 55, 0.6);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 12px;
  margin-top: 16px;
}

:deep(.tool-panel .ant-card-head) {
  background: rgba(55, 65, 81, 0.5);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 12px 12px 0 0;
}

:deep(.tool-panel .ant-card-head-title) {
  color: #ffffff;
  font-weight: 600;
}

:deep(.tool-panel .ant-card-body) {
  padding: 16px;
}

/* 表单样式 */
:deep(.tool-form .ant-form-item-label > label) {
  color: #d1d5db;
  font-weight: 500;
}

:deep(.input-area.ant-input),
:deep(.key-input.ant-input),
:deep(.result-area.ant-input) {
  background: rgba(17, 24, 39, 0.8);
  border: 1px solid rgba(75, 85, 99, 0.4);
  color: #ffffff;
}

:deep(.input-area.ant-input:focus),
:deep(.key-input.ant-input:focus),
:deep(.result-area.ant-input:focus) {
  border-color: #22d3ee;
  box-shadow: 0 0 0 2px rgba(34, 211, 238, 0.1);
}

:deep(.input-area.ant-input::placeholder),
:deep(.key-input.ant-input::placeholder) {
  color: #6b7280;
}

/* 结果区域 */
.result-actions {
  margin-top: 8px;
  text-align: right;
}

.copy-btn {
  background: rgba(75, 85, 99, 0.4) !important;
  border-color: rgba(75, 85, 99, 0.6) !important;
  color: #d1d5db !important;
}

.copy-btn:hover {
  background: rgba(75, 85, 99, 0.6) !important;
  border-color: #22d3ee !important;
  color: #22d3ee !important;
}

/* 按钮样式 */
:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  color: #ffffff;
}

:deep(.ant-btn-primary:hover) {
  background: linear-gradient(135deg, #2563eb, #1e40af);
}

:deep(.ant-btn-primary:disabled) {
  background: rgba(75, 85, 99, 0.4);
  color: #6b7280;
}

/* 图标样式 */
:deep(.ant-btn) i {
  margin-right: 4px;
}
</style>
