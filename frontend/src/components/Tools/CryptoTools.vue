<template>
  <div class="flex flex-col p-6 text-gray-200">
    <h2 class="text-xl font-medium mb-6 tracking-wide flex items-center">
      <i class="ri-lock-line mr-2 text-cyan-400"></i>加密解密工具
    </h2>

    <!-- 工具列表 -->
    <div class="space-y-1.5">
      <button
        v-for="tool in tools"
        :key="tool.action"
        @click="showModal(tool.action)"
        class="w-full text-left px-4 py-2.5 rounded-xl text-sm font-medium tracking-wide transition-all duration-200 hover:bg-gray-700/50 focus:bg-gray-700/50 flex items-center"
      >
        <i
          :class="tool.icon"
          class="text-lg mr-3 text-gray-400 group-hover:text-cyan-400"
        ></i>
        <span>{{ tool.name }}</span>
      </button>
    </div>

    <!-- 操作面板 -->
    <div
      v-if="isModalVisible"
      class="mt-6 rounded-2xl bg-gray-800/30 backdrop-blur-sm border border-gray-700/30 p-6 max-h-[400px] overflow-y-auto relative"
    >
      <!-- 标题和关闭按钮 -->
      <div class="flex justify-between items-center mb-4">
        <h3 class="text-base font-medium">{{ currentToolName }}</h3>
        <button
          @click="closeModal"
          class="text-gray-400 hover:text-white p-1 rounded-lg hover:bg-gray-700/30 transition-all duration-200"
        >
          <i class="ri-close-line text-lg"></i>
        </button>
      </div>

      <!-- 输入区域 -->
      <div class="space-y-4">
        <div>
          <label
            class="block text-sm font-medium mb-2 text-gray-300 flex items-center"
          >
            <i class="ri-text-line mr-2 text-gray-400"></i>输入文本
          </label>
          <textarea
            v-model="inputText"
            class="w-full px-4 py-2.5 rounded-xl bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-2 focus:ring-gray-600/50 transition-all duration-200 min-h-[80px]"
            placeholder="请输入需要处理的文本"
          ></textarea>
        </div>

        <!-- AES加密解密的密钥输入 -->
        <div v-if="['aesEncrypt', 'aesDecrypt'].includes(currentAction)">
          <label
            class="block text-sm font-medium mb-2 text-gray-300 flex items-center"
          >
            <i class="ri-key-2-line mr-2 text-gray-400"></i>密钥
          </label>
          <input
            v-model="key"
            type="text"
            class="w-full px-4 py-2.5 rounded-xl bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-2 focus:ring-gray-600/50 transition-all duration-200"
            placeholder="请输入密钥"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="flex space-x-3">
          <button
            @click="handleAction"
            class="flex-1 px-4 py-2.5 rounded-xl bg-gray-700/50 hover:bg-gray-600/50 text-sm font-medium transition-all duration-200 flex items-center justify-center"
          >
            <i class="ri-play-line mr-2"></i>执行
          </button>
          <button
            @click="clearInputs"
            class="flex-1 px-4 py-2.5 rounded-xl bg-gray-800/50 hover:bg-gray-700/50 text-sm font-medium transition-all duration-200 flex items-center justify-center"
          >
            <i class="ri-delete-bin-line mr-2"></i>清空
          </button>
        </div>

        <!-- 结果显示 -->
        <div v-if="outputText" class="space-y-3 mt-2">
          <div
            class="p-4 rounded-xl bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 break-words"
          >
            <p class="text-sm text-gray-400 mb-2 flex items-center">
              <i class="ri-file-list-line mr-2"></i>处理结果：
            </p>
            <div class="max-h-[200px] overflow-y-auto overflow-x-auto">
              <p class="text-sm p-1">{{ outputText }}</p>
            </div>
          </div>

          <button
            @click="copyToClipboard"
            class="w-full px-4 py-2.5 rounded-xl bg-gray-700/50 hover:bg-gray-600/50 text-sm font-medium transition-all duration-200 flex items-center justify-center"
          >
            <i class="ri-clipboard-line mr-2"></i>
            <span>{{ copyButtonText }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 复制成功提示 -->
    <div
      v-if="showCopySuccess"
      class="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-gray-800/90 backdrop-blur-sm text-white text-sm px-4 py-2 rounded-full shadow-lg border border-gray-700/30 transition-all duration-300 flex items-center"
    >
      <i class="ri-check-line mr-2 text-green-400"></i>已复制到剪贴板
    </div>
  </div>
</template>

<script>
import { ref, computed } from "vue";
import CryptoJS from "crypto-js";

export default {
  name: "CryptoTools",
  setup() {
    const isModalVisible = ref(false);
    const inputText = ref("");
    const outputText = ref("");
    const currentAction = ref("");
    const key = ref("");
    const showCopySuccess = ref(false);
    const copyButtonText = ref("复制结果");

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

    // 获取当前工具名称
    const currentToolName = computed(() => {
      const tool = tools.find((t) => t.action === currentAction.value);
      return tool ? tool.name : "";
    });

    // 显示模态框
    const showModal = (action) => {
      currentAction.value = action;
      isModalVisible.value = true;
      inputText.value = "";
      outputText.value = "";
      key.value = "";
    };

    // 关闭模态框
    const closeModal = () => {
      isModalVisible.value = false;
    };

    // 清空输入
    const clearInputs = () => {
      inputText.value = "";
      outputText.value = "";
      key.value = "";
    };

    // 处理加密解密操作
    const handleAction = () => {
      if (!inputText.value) return;

      switch (currentAction.value) {
        case "base64Encode":
          outputText.value = btoa(inputText.value);
          break;
        case "base64Decode":
          try {
            outputText.value = atob(inputText.value);
          } catch (e) {
            outputText.value = "无效的 Base64 输入";
          }
          break;
        case "aesEncrypt":
          if (inputText.value && key.value) {
            outputText.value = CryptoJS.AES.encrypt(
              inputText.value,
              key.value
            ).toString();
          } else {
            outputText.value = "请输入文本和密钥";
          }
          break;
        case "aesDecrypt":
          if (inputText.value && key.value) {
            try {
              const decrypted = CryptoJS.AES.decrypt(
                inputText.value,
                key.value
              ).toString(CryptoJS.enc.Utf8);
              outputText.value = decrypted || "解密失败，可能是密钥错误";
            } catch (e) {
              outputText.value = "无效的 AES 输入或密钥错误";
            }
          } else {
            outputText.value = "请输入密文和密钥";
          }
          break;
        case "md5Hash":
          outputText.value = CryptoJS.MD5(inputText.value).toString();
          break;
        case "sha256Hash":
          outputText.value = CryptoJS.SHA256(inputText.value).toString();
          break;
        case "urlEncode":
          outputText.value = encodeURIComponent(inputText.value);
          break;
        case "urlDecode":
          try {
            outputText.value = decodeURIComponent(inputText.value);
          } catch (e) {
            outputText.value = "无效的 URL 编码";
          }
          break;
        case "hexEncode":
          outputText.value = textToHex(inputText.value);
          break;
        case "hexDecode":
          try {
            outputText.value = hexToText(inputText.value);
          } catch (e) {
            outputText.value = "无效的 Hex 编码";
          }
          break;
        default:
          outputText.value = "";
      }
    };

    // 文本转 Hex
    const textToHex = (text) => {
      return text
        .split("")
        .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
        .join("");
    };

    // Hex 转文本
    const hexToText = (hex) => {
      if (!/^[0-9a-fA-F]+$/.test(hex)) {
        throw new Error("Invalid hex string");
      }

      return hex
        .match(/.{1,2}/g)
        .map((byte) => String.fromCharCode(parseInt(byte, 16)))
        .join("");
    };

    // 复制到剪贴板
    const copyToClipboard = () => {
      navigator.clipboard
        .writeText(outputText.value)
        .then(() => {
          // 更改按钮文字
          copyButtonText.value = "已复制";
          // 显示提示
          showCopySuccess.value = true;

          // 2秒后恢复按钮文字
          setTimeout(() => {
            copyButtonText.value = "复制结果";
            showCopySuccess.value = false;
          }, 2000);
        })
        .catch(() => {
          copyButtonText.value = "复制失败";
          setTimeout(() => {
            copyButtonText.value = "复制结果";
          }, 2000);
        });
    };

    return {
      isModalVisible,
      inputText,
      outputText,
      currentAction,
      currentToolName,
      key,
      showCopySuccess,
      copyButtonText,
      tools,
      showModal,
      closeModal,
      clearInputs,
      handleAction,
      copyToClipboard,
    };
  },
};
</script>

<style scoped>
/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 10px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}

/* 确保文本正确换行 */
.break-words {
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* 优化滚动容器的样式 */
.overflow-y-auto {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.3) transparent;
}

/* 按钮按下效果 */
button:active {
  transform: scale(0.98);
}

/* 工具列表项悬停效果 */
button:hover i {
  color: #22d3ee; /* 青色 */
}
</style>
