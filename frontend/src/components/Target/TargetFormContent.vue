<!-- components/Target/TargetFormContent.vue -->
<template>
  <form @submit.prevent="handleSubmit" >
    <!-- 目标名称 -->
    <div class="form-field">
      <label class="form-label">
        <span class="required-indicator">*</span>
        目标名称
      </label>
      <div class="input-wrapper">
        <i class="ri-file- input-icon"></i>
        <input
          v-model="formData.name"
          type="text"
          placeholder="请输入目标名称"
          class="form-input"
          required
        />
      </div>
    </div>

    <!-- 目标类型 - 使用单选按钮替代下拉菜单 -->
    <div class="form-field">
      <label class="form-label">
        <span class="required-indicator">*</span>
        目标类型
      </label>
      <div >
        <label
          class="type-radio-button"
          :class="{ active: formData.type === 'domain' }"
        >
          <input
            type="radio"
            v-model="formData.type"
            value="domain"
            class="sr-only"
          />
          <i class="ri-global-line"></i>
          域名
        </label>
        <label
          class="type-radio-button"
          :class="{ active: formData.type === 'ip' }"
        >
          <input
            type="radio"
            v-model="formData.type"
            value="ip"
            class="sr-only"
          />
          <i class="ri-server-line"></i>
          IP地址
        </label>
      </div>
    </div>

    <!-- 目标地址 -->
    <div class="form-field">
      <label class="form-label">
        <span class="required-indicator">*</span>
        {{ formData.type === "domain" ? "域名" : "IP地址" }}
      </label>
      <div class="input-wrapper" :class="{ error: validationError }">
        <i
          class="ri-link input-icon"
          :class="{ '': validationError }"
        ></i>
        <input
          v-model="formData.target"
          type="text"
          @input="handleTargetInput"
          :placeholder="
            formData.type === 'domain'
              ? '请输入域名（不含 http:// 或 https://）'
              : '请输入IP地址（如 192.168.1.1）'
          "
          class="form-input"
          :class="{ ' ': validationError, }"
          required
        />
      </div>

      <!-- 错误提示 - 改进样式 -->
      <div v-if="validationError" class="error-message">
        <i class="ri-error-warning-line"></i>
        {{ validationError }}
      </div>

      <!-- 表单帮助文本 -->
      <p class="help-text" v-if="!validationError">
        <i class="ri-information-line"></i>
        {{
          formData.type === "domain"
            ? "请输入不带协议的域名，如: example.com"
            : "请输入有效的IPv4地址，如: 192.168.1.1"
        }}
      </p>
    </div>

    <!-- 目标描述 -->
    <div class="form-field">
      <label class="form-label">
        描述
        <span >(选填)</span>
      </label>
      <div class="input-wrapper">
        <i class="ri-file-list-line input-icon"></i>
        <textarea
          v-model="formData.description"
          rows="3"
          placeholder="请输入目标描述信息，帮助您区分不同目标"
          class="form-input"
        ></textarea>
      </div>
      <div >
        <p class="help-text">详细描述有助于更好地管理目标</p>
        <p >
          {{ formData.description.length }}/200
        </p>
      </div>
    </div>

    <!-- 表单按钮 -->
    <div
      
    >
      <button type="button" @click="$emit('cancel')" class="cancel-button">
        <i class="ri-close-line .5"></i>
        取消
      </button>
      <button
        type="submit"
        :disabled="isSubmitDisabled || isSubmitting"
        class="submit-button"
        :class="{ ' ': isSubmitDisabled || isSubmitting, }"
      >
        <i class="ri-save-line .5"></i>
        {{ isSubmitting ? "提交中..." : "保存" }}
        <span v-if="isSubmitting" class="loading-dots">...</span>
      </button>
    </div>
  </form>
</template>

<script setup>
import { ref, onMounted, defineProps, defineEmits, computed, watch } from "vue";

const props = defineProps({
  initialData: {
    type: Object,
    default: () => ({
      name: "",
      type: "domain",
      target: "",
      description: "",
    }),
  },
  isSubmitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["submit", "cancel"]);

const formData = ref({ ...props.initialData });
const validationError = ref("");

// 验证函数
const validateDomain = (domain) => {
  if (!domain) return false;
  // 更精确的域名验证逻辑
  const domainRegex =
    /^(?!http(s)?:\/\/)((?!-)[A-Za-z0-9-]{1,63}(?<!-)\.)+[A-Za-z]{2,}$/;
  return domainRegex.test(domain);
};

const validateIP = (ip) => {
  if (!ip) return false;
  // IPv4验证规则
  const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/;
  if (!ipv4Regex.test(ip)) return false;

  const parts = ip.split(".");
  return parts.every((part) => {
    const num = parseInt(part, 10);
    return num >= 0 && num <= 255;
  });
};

// 验证状态计算
const validation = computed(() => {
  if (!formData.value.name) {
    return {
      isValid: false,
      error: "请填写目标名称",
    };
  }

  if (!formData.value.target) {
    return {
      isValid: false,
      error: "请填写目标地址",
    };
  }

  if (formData.value.type === "domain") {
    return {
      isValid: validateDomain(formData.value.target),
      error: "请输入有效的域名格式 (如: example.com)",
    };
  } else if (formData.value.type === "ip") {
    return {
      isValid: validateIP(formData.value.target),
      error: "请输入有效的 IPv4 地址格式 (如: 192.168.1.1)",
    };
  }

  return { isValid: true, error: "" };
});

// 使用 watch 来更新验证错误信息
watch(
  () => [formData.value.target, formData.value.type],
  () => {
    validationError.value = !validation.value.isValid
      ? validation.value.error
      : "";
  },
  { immediate: true }
);

// 计算提交按钮是否禁用
const isSubmitDisabled = computed(() => {
  return !validation.value.isValid;
});

// 处理目标值变化
const handleTargetInput = () => {
  if (formData.value.type === "domain") {
    // 自动去除http://和https://
    formData.value.target = formData.value.target.replace(/^https?:\/\//i, "");
  }
};

const handleSubmit = async () => {
  if (isSubmitDisabled.value) return;

  // 构造提交数据
  const submitData = {
    name: formData.value.name.trim(),
    type: formData.value.type,
    target: formData.value.target.trim(),
    description: formData.value.description.trim(),
    status: "active", // 添加默认状态
  };

  try {
    emit("submit", submitData);
  } catch (error) {
    console.error("Form submission error:", error);
    validationError.value = "提交失败，请重试";
  }
};

onMounted(() => {
  formData.value = { ...props.initialData };
});
</script>

<style scoped>
/* 表单组件样式 */
.form-field {
  @apply mb-5;
}

.form-label {
  @apply block text-sm font-medium text-gray-300 mb-1.5;
}

.input-wrapper {
  @apply relative;
}

.input-icon {
  @apply absolute left-3.5 top-3 text-gray-400;
}

.form-input {
  @apply w-full pl-10 pr-4 py-2.5 rounded-xl bg-gray-700/50 text-gray-100 border border-gray-600/30 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all duration-200;
}

/* 必填项指示器 */
.required-indicator {
  @apply text-red-400 mr-1;
}

/* 错误消息 */
.error-message {
  @apply mt-1.5 text-sm text-red-400 flex items-start;
}

/* 帮助文本 */
.help-text {
  @apply text-xs text-gray-500 mt-1.5 flex items-center;
}

/* 类型选择按钮 */
.type-radio-button {
  @apply px-4 py-2.5 rounded-xl text-sm bg-gray-700/50 text-gray-300 border border-gray-600/30 flex items-center cursor-pointer transition-all duration-200 hover:bg-gray-600/50;
}

.type-radio-button.active {
  @apply bg-blue-700/30 text-blue-200 border-blue-500/40;
}

/* 取消按钮 */
.cancel-button {
  @apply px-4 py-2.5 rounded-xl text-sm font-medium bg-gray-700/50 hover:bg-gray-600/50 text-gray-300 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-600/50 flex items-center;
}

/* 提交按钮 */
.submit-button {
  @apply px-5 py-2.5 rounded-xl text-sm font-medium bg-blue-600/70 hover:bg-blue-500/70 text-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 flex items-center shadow-md;
}

/* 提交中动画 */
@keyframes loading {
  0%,
  100% {
    opacity: 0.2;
  }
  50% {
    opacity: 1;
  }
}

.loading-dots {
  display: inline-block;
  animation: loading 1.5s infinite;
}
</style>
