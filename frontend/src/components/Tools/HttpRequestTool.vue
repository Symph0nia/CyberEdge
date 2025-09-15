<template>
  <div class="flex flex-col p-6 text-gray-200 space-y-6 max-w-5xl mx-auto">
    <!-- 标题和说明 -->
    <div class="text-center space-y-4">
      <h2
        class="text-xl font-medium tracking-wide flex items-center justify-center"
      >
        <i class="ri-global-line mr-2 text-cyan-400"></i>网络请求工具
      </h2>
      <div
        class="text-sm text-gray-400 leading-relaxed bg-gray-800/30 p-3 rounded-lg border border-gray-700/30"
      >
        <i class="ri-information-line mr-1 text-yellow-400"></i>
        使用该功能时需要关闭浏览器的 CORS 安全检查
        <div class="text-xs opacity-75 mt-2">
          <span class="mr-2 block sm:inline-block mb-1 sm:mb-0"
            >在 Chrome 中运行：</span
          >
          <div
            class="flex items-center bg-gray-900/50 px-3 py-1.5 rounded-md text-gray-300 group relative"
          >
            <code
              class="flex-1 overflow-x-auto whitespace-nowrap text-xs sm:text-sm"
            >
              chrome.exe --user-data-dir="C:/Chrome dev session"
              --disable-web-security
            </code>
            <button
              @click="copyCommand"
              class="ml-2 text-gray-400 hover:text-cyan-400 transition-colors"
              title="复制命令"
            >
              <i class="ri-clipboard-line"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 表单区域 -->
    <div class="space-y-4">
      <!-- URL输入 -->
      <div class="space-y-2">
        <label
          class="block text-sm font-medium text-gray-300 flex items-center"
        >
          <i class="ri-link mr-2 text-gray-400"></i>请求地址
        </label>
        <div class="flex flex-col sm:flex-row space-y-2 sm:space-y-0">
          <input
            type="text"
            v-model="url"
            class="w-full px-4 py-2.5 rounded-xl sm:rounded-l-xl sm:rounded-r-none bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-2 focus:ring-gray-600/50 transition-all duration-200"
            placeholder="输入请求的 URL"
          />
          <select
            v-model="httpMethod"
            class="px-4 py-2.5 rounded-xl sm:rounded-l-none sm:rounded-r-xl bg-gray-800/70 backdrop-blur-sm border sm:border-l-0 border-gray-700/30 text-sm focus:outline-none focus:ring-2 focus:ring-gray-600/50 transition-all duration-200 appearance-none"
          >
            <option
              v-for="method in [
                'GET',
                'POST',
                'PUT',
                'DELETE',
                'PATCH',
                'OPTIONS',
              ]"
              :key="method"
              :value="method"
            >
              {{ method }}
            </option>
          </select>
        </div>
      </div>

      <!-- 请求头部选项 -->
      <div class="space-y-2">
        <div class="flex justify-between items-center">
          <label
            class="block text-sm font-medium text-gray-300 flex items-center"
          >
            <i class="ri-file-list-line mr-2 text-gray-400"></i>请求头部
          </label>
          <button
            @click="addHeader"
            class="text-xs bg-gray-700/50 px-2 py-1 rounded-md hover:bg-gray-700 transition-colors flex items-center"
          >
            <i class="ri-add-line mr-1"></i>添加
          </button>
        </div>

        <div
          v-for="(header, index) in headers"
          :key="index"
          class="flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-2 items-center mb-2"
        >
          <input
            v-model="header.key"
            class="w-full sm:flex-1 px-3 py-2 rounded-lg sm:rounded-l-lg sm:rounded-r-none bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-1 focus:ring-gray-600/50 transition-all duration-200"
            placeholder="Key"
          />
          <input
            v-model="header.value"
            class="w-full sm:flex-1 px-3 py-2 rounded-lg sm:rounded-l-none sm:rounded-r-lg bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-1 focus:ring-gray-600/50 transition-all duration-200"
            placeholder="Value"
          />
          <button
            @click="removeHeader(index)"
            class="hidden sm:block text-gray-400 hover:text-red-400 transition-colors p-1"
          >
            <i class="ri-delete-bin-line"></i>
          </button>
          <button
            @click="removeHeader(index)"
            class="sm:hidden w-full text-left text-gray-400 hover:text-red-400 transition-colors p-1 mt-1"
          >
            <i class="ri-delete-bin-line mr-1"></i>删除
          </button>
        </div>
      </div>

      <!-- 数据格式选择和请求数据 -->
      <div class="space-y-2" v-if="showRequestBody">
        <div class="flex justify-between items-center">
          <label
            class="block text-sm font-medium text-gray-300 flex items-center"
          >
            <i class="ri-code-box-line mr-2 text-gray-400"></i>请求数据
          </label>
          <select
            v-model="dataFormat"
            class="text-xs px-3 py-1.5 rounded-lg bg-gray-800/50 backdrop-blur-sm border border-gray-700/30 focus:outline-none focus:ring-1 focus:ring-gray-600/50 transition-all duration-200 appearance-none"
          >
            <option value="json">JSON</option>
            <option value="raw">Raw</option>
            <option value="form">Form Data</option>
          </select>
        </div>

        <!-- JSON或Raw数据输入 -->
        <textarea
          v-if="dataFormat !== 'form'"
          v-model="requestData"
          rows="5"
          class="w-full px-4 py-2.5 rounded-xl bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-2 focus:ring-gray-600/50 transition-all duration-200"
          :placeholder="jsonPlaceholder"
        ></textarea>

        <!-- Form数据输入 -->
        <div v-else class="space-y-2">
          <div
            v-for="(formItem, index) in formData"
            :key="index"
            class="flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-2 items-center mb-2"
          >
            <input
              v-model="formItem.key"
              class="w-full sm:flex-1 px-3 py-2 rounded-lg sm:rounded-l-lg sm:rounded-r-none bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-1 focus:ring-gray-600/50 transition-all duration-200"
              placeholder="Key"
            />
            <input
              v-model="formItem.value"
              class="w-full sm:flex-1 px-3 py-2 rounded-lg sm:rounded-l-none sm:rounded-r-lg bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 text-sm focus:outline-none focus:ring-1 focus:ring-gray-600/50 transition-all duration-200"
              placeholder="Value"
            />
            <button
              @click="removeFormItem(index)"
              class="hidden sm:block text-gray-400 hover:text-red-400 transition-colors p-1"
            >
              <i class="ri-delete-bin-line"></i>
            </button>
            <button
              @click="removeFormItem(index)"
              class="sm:hidden w-full text-left text-gray-400 hover:text-red-400 transition-colors p-1 mt-1"
            >
              <i class="ri-delete-bin-line mr-1"></i>删除
            </button>
          </div>
          <button
            @click="addFormItem"
            class="text-xs bg-gray-700/50 px-2 py-1 rounded-md hover:bg-gray-700 transition-colors flex items-center self-start"
          >
            <i class="ri-add-line mr-1"></i>添加字段
          </button>
        </div>
      </div>

      <!-- 发送按钮 -->
      <button
        @click="sendRequest"
        class="w-full px-4 py-3 rounded-xl bg-gray-700/50 hover:bg-gray-600/50 text-sm font-medium transition-all duration-200 flex items-center justify-center space-x-2"
        :disabled="isLoading"
      >
        <i class="ri-send-plane-line mr-2" v-if="!isLoading"></i>
        <i class="ri-loader-4-line animate-spin mr-2" v-else></i>
        <span>{{ isLoading ? "发送中..." : "发送请求" }}</span>
      </button>

      <!-- 响应结果 -->
      <div
        v-if="response"
        class="space-y-3 bg-gray-800/20 p-4 rounded-xl border border-gray-700/30"
      >
        <!-- 状态和时间信息 -->
        <div
          class="flex flex-col sm:flex-row sm:justify-between text-sm space-y-2 sm:space-y-0"
        >
          <div class="flex items-center space-x-2">
            <span class="font-medium text-gray-300">状态码：</span>
            <span
              :class="{
                'text-green-400': responseStatus >= 200 && responseStatus < 300,
                'text-yellow-400':
                  responseStatus >= 300 && responseStatus < 400,
                'text-red-400': responseStatus >= 400,
              }"
              >{{ responseStatus }}</span
            >
          </div>
          <div class="text-gray-400">
            <span>耗时：{{ responseTime }}ms</span>
          </div>
        </div>

        <!-- 响应头部（可折叠） -->
        <div class="border-t border-gray-700/30 pt-3">
          <button
            @click="showHeaders = !showHeaders"
            class="flex items-center justify-between w-full text-left text-sm text-gray-300 mb-2"
          >
            <span class="font-medium flex items-center">
              <i class="ri-file-list-line mr-2"></i>响应头部
            </span>
            <i
              :class="
                showHeaders ? 'ri-arrow-up-s-line' : 'ri-arrow-down-s-line'
              "
              class="text-gray-400"
            ></i>
          </button>

          <div
            v-if="showHeaders"
            class="bg-gray-900/30 rounded-lg p-3 mb-3 text-sm"
          >
            <div
              v-for="(value, key) in responseHeaders"
              :key="key"
              class="flex flex-wrap mb-1"
            >
              <span class="font-medium text-gray-400 mr-2 whitespace-nowrap"
                >{{ key }}:</span
              >
              <span class="text-gray-300 break-all">{{ value }}</span>
            </div>
          </div>
        </div>

        <!-- 响应数据 -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium text-gray-300 flex items-center">
              <i class="ri-file-text-line mr-2"></i>响应数据
            </span>
            <button
              @click="copyResponse"
              class="text-xs bg-gray-700/50 px-2 py-1 rounded-md hover:bg-gray-700 transition-colors flex items-center"
            >
              <i class="ri-clipboard-line mr-1"></i>复制
            </button>
          </div>
          <div
            class="p-3 rounded-xl bg-gray-900/50 backdrop-blur-sm border border-gray-700/30 break-words"
          >
            <div class="max-h-[300px] overflow-y-auto overflow-x-auto">
              <pre class="text-sm whitespace-pre-wrap" v-if="!isJsonResponse">{{
                response
              }}</pre>
              <pre
                v-else
                class="text-sm json-formatter"
                v-html="formattedJsonResponse"
              ></pre>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 复制成功提示 -->
    <div
      v-if="showCopySuccess"
      class="fixed bottom-4 right-4 bg-gray-800/90 backdrop-blur-sm text-white text-sm px-4 py-2 rounded-lg shadow-lg border border-gray-700/30 transition-all duration-300 flex items-center z-[1001]"
    >
      <i class="ri-check-line mr-2 text-green-400"></i>{{ copyMessage }}
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from "vue";

export default {
  name: "HttpRequestTool",
  setup() {
    const url = ref("");
    const httpMethod = ref("GET");
    const requestData = ref("");
    const response = ref(null);
    const responseStatus = ref(null);
    const dataFormat = ref("json");
    const isLoading = ref(false);
    const responseTime = ref(0);
    const responseHeaders = ref({});
    const showHeaders = ref(false);
    const showCopySuccess = ref(false);
    const copyMessage = ref("");
    const headers = ref([{ key: "Content-Type", value: "application/json" }]);
    const formData = ref([{ key: "", value: "" }]);

    // 添加这个计算属性用于生成正确的占位符文本
    const jsonPlaceholder = computed(() => {
      return dataFormat.value === "json"
        ? '{"key": "value"}'
        : "Raw request data";
    });

    // 计算属性：是否显示请求体
    const showRequestBody = computed(() => {
      return httpMethod.value !== "GET" && httpMethod.value !== "OPTIONS";
    });

    // 计算属性：响应是否为JSON
    const isJsonResponse = computed(() => {
      if (!response.value) return false;
      try {
        JSON.parse(response.value);
        return true;
      } catch (e) {
        return false;
      }
    });

    // 计算属性：格式化的JSON响应
    const formattedJsonResponse = computed(() => {
      if (!isJsonResponse.value) return "";
      try {
        const parsed = JSON.parse(response.value);
        return syntaxHighlight(JSON.stringify(parsed, null, 2));
      } catch (e) {
        return response.value;
      }
    });

    // JSON语法高亮函数
    const syntaxHighlight = (json) => {
      return json
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(
          /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g,
          function (match) {
            let cls = "json-number";
            if (/^"/.test(match)) {
              if (/:$/.test(match)) {
                cls = "json-key";
              } else {
                cls = "json-string";
              }
            } else if (/true|false/.test(match)) {
              cls = "json-boolean";
            } else if (/null/.test(match)) {
              cls = "json-null";
            }
            return '<span class="' + cls + '">' + match + "</span>";
          }
        );
    };

    // 监视HTTP方法变化，如果是GET，自动更新Content-Type
    watch(httpMethod, (newMethod) => {
      if (newMethod === "GET") {
        const contentTypeHeader = headers.value.find(
          (h) => h.key.toLowerCase() === "content-type"
        );
        if (contentTypeHeader) {
          contentTypeHeader.value = "application/json";
        }
      }
    });

    // 添加请求头
    const addHeader = () => {
      headers.value.push({ key: "", value: "" });
    };

    // 移除请求头
    const removeHeader = (index) => {
      headers.value.splice(index, 1);
    };

    // 添加表单项
    const addFormItem = () => {
      formData.value.push({ key: "", value: "" });
    };

    // 移除表单项
    const removeFormItem = (index) => {
      formData.value.splice(index, 1);
    };

    // 复制命令
    const copyCommand = () => {
      const command =
        'chrome.exe --user-data-dir="C:/Chrome dev session" --disable-web-security';
      navigator.clipboard.writeText(command).then(() => {
        showCopyNotification("命令已复制到剪贴板");
      });
    };

    // 复制响应
    const copyResponse = () => {
      navigator.clipboard.writeText(response.value).then(() => {
        showCopyNotification("响应已复制到剪贴板");
      });
    };

    // 显示复制成功通知
    const showCopyNotification = (message) => {
      copyMessage.value = message;
      showCopySuccess.value = true;
      setTimeout(() => {
        showCopySuccess.value = false;
      }, 2000);
    };

    // 发送请求
    const sendRequest = async () => {
      if (!url.value) return;

      isLoading.value = true;
      const startTime = performance.now();

      try {
        // 构建请求头
        const headersObj = {};
        headers.value.forEach((h) => {
          if (h.key && h.value) {
            headersObj[h.key] = h.value;
          }
        });

        // 构建请求体
        let body = null;
        if (showRequestBody.value) {
          if (dataFormat.value === "json") {
            try {
              // 验证是有效的JSON
              body = JSON.stringify(JSON.parse(requestData.value));
            } catch (e) {
              throw new Error("请求数据不是有效的JSON格式");
            }
          } else if (dataFormat.value === "raw") {
            body = requestData.value;
          } else if (dataFormat.value === "form") {
            const formDataObj = new FormData();
            formData.value.forEach((item) => {
              if (item.key) {
                formDataObj.append(item.key, item.value);
              }
            });
            body = formDataObj;
            // 对于FormData，不要手动设置Content-Type，浏览器会自动添加
            delete headersObj["Content-Type"];
          }
        }

        const options = {
          method: httpMethod.value,
          headers: headersObj,
          body: showRequestBody.value ? body : null,
        };

        const res = await fetch(url.value, options);
        responseStatus.value = res.status;

        // 处理响应头
        responseHeaders.value = {};
        res.headers.forEach((value, key) => {
          responseHeaders.value[key] = value;
        });

        response.value = await res.text();
        responseTime.value = Math.round(performance.now() - startTime);
      } catch (error) {
        responseStatus.value = 0;
        response.value = `请求错误: ${error.message}`;
        responseTime.value = Math.round(performance.now() - startTime);
        responseHeaders.value = {};
      } finally {
        isLoading.value = false;
      }
    };

    return {
      url,
      httpMethod,
      requestData,
      response,
      responseStatus,
      dataFormat,
      isLoading,
      responseTime,
      responseHeaders,
      showHeaders,
      showCopySuccess,
      copyMessage,
      headers,
      formData,
      showRequestBody,
      isJsonResponse,
      formattedJsonResponse,
      jsonPlaceholder,
      addHeader,
      removeHeader,
      addFormItem,
      removeFormItem,
      copyCommand,
      copyResponse,
      sendRequest,
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

/* JSON高亮样式 */
.json-formatter .json-key {
  color: #9cdcfe;
}
.json-formatter .json-string {
  color: #ce9178;
}
.json-formatter .json-number {
  color: #b5cea8;
}
.json-formatter .json-boolean {
  color: #569cd6;
}
.json-formatter .json-null {
  color: #569cd6;
}

/* 移除select的默认箭头 */
select {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 0.5rem center;
  background-repeat: no-repeat;
  background-size: 1.5em 1.5em;
  padding-right: 2.5rem;
}

/* 按钮禁用状态 */
button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* 按钮按下效果 */
button:not(:disabled):active {
  transform: scale(0.98);
}

/* 响应式调整 */
@media (max-width: 640px) {
  .form-row {
    flex-direction: column;
  }
}
</style>
