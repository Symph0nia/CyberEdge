<template>
  <div class="max-">
    <!-- 标题和说明 -->
    <div >
      <h2
        
      >
        <i class="ri-global-line"></i>网络请求工具
      </h2>
      <div
        class="border"
      >
        <i class="ri-information-line"></i>
        使用该功能时需要关闭浏览器的 CORS 安全检查
        <div >
          <span 
            >在 Chrome 中运行：</span
          >
          <div
            class=".5 group"
          >
            <code
              class="sm:"
            >
              chrome.exe --user-data-dir="C:/Chrome dev session"
              --disable-web-security
            </code>
            <button
              @click="copyCommand"
              class="hover:"
              title="复制命令"
            >
              <i class="ri-clipboard-line"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 表单区域 -->
    <div >
      <!-- URL输入 -->
      <div >
        <label
          
        >
          <i class="ri-link"></i>请求地址
        </label>
        <div class="sm: sm:">
          <input
            type="text"
            v-model="url"
            class=".5 sm: sm: border duration-200"
            placeholder="输入请求的 URL"
          />
          <select
            v-model="httpMethod"
            class=".5 sm: sm: border sm: duration-200 appearance-none"
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
      <div >
        <div >
          <label
            
          >
            <i class="ri-file-list-line"></i>请求头部
          </label>
          <button
            @click="addHeader"
            class="hover:"
          >
            <i class="ri-add-line"></i>添加
          </button>
        </div>

        <div
          v-for="(header, index) in headers"
          :key="index"
          class="sm: sm: sm:"
        >
          <input
            v-model="header.key"
            class="sm: sm: sm: border duration-200"
            placeholder="Key"
          />
          <input
            v-model="header.value"
            class="sm: sm: sm: border duration-200"
            placeholder="Value"
          />
          <button
            @click="removeHeader(index)"
            class="hover:"
          >
            <i class="ri-delete-bin-line"></i>
          </button>
          <button
            @click="removeHeader(index)"
            class="hover:"
          >
            <i class="ri-delete-bin-line"></i>删除
          </button>
        </div>
      </div>

      <!-- 数据格式选择和请求数据 -->
      <div  v-if="showRequestBody">
        <div >
          <label
            
          >
            <i class="ri-code-box-line"></i>请求数据
          </label>
          <select
            v-model="dataFormat"
            class=".5 border duration-200 appearance-none"
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
          class=".5 border duration-200"
          :placeholder="jsonPlaceholder"
        ></textarea>

        <!-- Form数据输入 -->
        <div v-else >
          <div
            v-for="(formItem, index) in formData"
            :key="index"
            class="sm: sm: sm:"
          >
            <input
              v-model="formItem.key"
              class="sm: sm: sm: border duration-200"
              placeholder="Key"
            />
            <input
              v-model="formItem.value"
              class="sm: sm: sm: border duration-200"
              placeholder="Value"
            />
            <button
              @click="removeFormItem(index)"
              class="hover:"
            >
              <i class="ri-delete-bin-line"></i>
            </button>
            <button
              @click="removeFormItem(index)"
              class="hover:"
            >
              <i class="ri-delete-bin-line"></i>删除
            </button>
          </div>
          <button
            @click="addFormItem"
            class="hover:"
          >
            <i class="ri-add-line"></i>添加字段
          </button>
        </div>
      </div>

      <!-- 发送按钮 -->
      <button
        @click="sendRequest"
        class="hover: duration-200"
        :disabled="isLoading"
      >
        <i class="ri-send-plane-line" v-if="!isLoading"></i>
        <i class="ri-loader-4-line" v-else></i>
        <span>{{ isLoading ? "发送中..." : "发送请求" }}</span>
      </button>

      <!-- 响应结果 -->
      <div
        v-if="response"
        class="border"
      >
        <!-- 状态和时间信息 -->
        <div
          class="sm: sm:"
        >
          <div >
            <span >状态码：</span>
            <span
              :class="{ '': responseStatus >= 200 && responseStatus < 300, '': responseStatus >= 300 && responseStatus < 400, '': responseStatus >= 400, }"
              >{{ responseStatus }}</span
            >
          </div>
          <div >
            <span>耗时：{{ responseTime }}ms</span>
          </div>
        </div>

        <!-- 响应头部（可折叠） -->
        <div >
          <button
            @click="showHeaders = !showHeaders"
            
          >
            <span >
              <i class="ri-file-list-line"></i>响应头部
            </span>
            <i
              :class="showHeaders ? 'ri-arrow-up-s-line' : 'ri-arrow-down-s-line'"
              
            ></i>
          </button>

          <div
            v-if="showHeaders"
            
          >
            <div
              v-for="(value, key) in responseHeaders"
              :key="key"
              
            >
              <span 
                >{{ key }}:</span
              >
              <span >{{ value }}</span>
            </div>
          </div>
        </div>

        <!-- 响应数据 -->
        <div>
          <div >
            <span >
              <i class="ri-file-"></i>响应数据
            </span>
            <button
              @click="copyResponse"
              class="hover:"
            >
              <i class="ri-clipboard-line"></i>复制
            </button>
          </div>
          <div
            class="border"
          >
            <div class="max-">
              <pre  v-if="!isJsonResponse">{{
                response
              }}</pre>
              <pre
                v-else
                class="json-formatter"
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
      class="border duration-300 z-[1001]"
    >
      <i class="ri-check-line"></i>{{ copyMessage }}
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
