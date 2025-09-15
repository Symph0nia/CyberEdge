<template>
  <Transition
    enter-active-class="ease-out duration-200"
    enter-from-
    enter-to-
    leave-active-class="ease-in duration-150"
    leave-from-
    leave-to-
  >
    <div class="inset-0" :class="customClass">
      <div class="min-">
        <!-- 背景遮罩 -->
        <div
          class="inset-0"
          @click="closeOnBackdropClick ? $emit('close') : null"
        ></div>

        <!-- 对话框内容 -->
        <Transition
          enter-active-class="ease-out duration-300"
          enter-from-
          enter-to-
          leave-active-class="ease-in duration-200"
          leave-from-
          leave-to-
        >
          <div
            class="/20 border"
            :class="[widthClasses, customDialogClass]"
            ref="dialogRef"
          >
            <!-- 标题栏 -->
            <div >
              <div >
                <div
                  v-if="showIndicator"
                  
                  :class="indicatorColorClass"
                ></div>
                <h3 >{{ title }}</h3>
                <span v-if="subtitle" >{{
                  subtitle
                }}</span>
              </div>

              <button
                v-if="showCloseButton"
                @click="$emit('close')"
                class="hover: duration-200 group"
                aria-label="关闭"
              >
                <i
                  class="ri-close-line group-hover: duration-200"
                ></i>
              </button>
            </div>

            <!-- 内容区域 -->
            <div
              
              :class="{ 'max- custom-scrollbar': scrollable, }"
            >
              <slot></slot>
            </div>

            <!-- 底部操作栏 -->
            <div  v-if="$slots.footer || $slots.actions">
              <slot name="footer">
                <div >
                  <slot name="actions"></slot>
                </div>
              </slot>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import {
  defineProps,
  onMounted,
  onBeforeUnmount,
  defineEmits,
  computed,
  ref,
} from "vue";

const emit = defineEmits(["close"]);
const dialogRef = ref(null);

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  subtitle: {
    type: String,
    default: "",
  },
  width: {
    type: String,
    default: "lg",
  },
  closeOnEsc: {
    type: Boolean,
    default: true,
  },
  closeOnBackdropClick: {
    type: Boolean,
    default: true,
  },
  showCloseButton: {
    type: Boolean,
    default: true,
  },
  showIndicator: {
    type: Boolean,
    default: true,
  },
  indicatorColor: {
    type: String,
    default: "blue",
    validator: (value) =>
      ["blue", "green", "red", "yellow", "purple", "gray"].includes(value),
  },
  scrollable: {
    type: Boolean,
    default: true,
  },
  customClass: {
    type: String,
    default: "",
  },
  customDialogClass: {
    type: String,
    default: "",
  },
  initialFocus: {
    type: String,
    default: "",
  },
});

// 响应式宽度类
const widthClasses = computed(() => {
  const widthMap = {
    sm: "w-full max-w-md",
    md: "w-full max-w-lg",
    lg: "w-full max-w-xl",
    xl: "w-full max-w-2xl",
    "2xl": "w-full max-w-3xl",
    "3xl": "w-full max-w-4xl",
    "4xl": "w-full max-w-5xl",
    "5xl": "w-full max-w-6xl",
    full: "w-full max-w-full mx-4",
    auto: "w-auto",
  };

  return widthMap[props.width] || widthMap.lg;
});

// 指示器颜色类
const indicatorColorClass = computed(() => {
  const colorMap = {
    blue: "bg-blue-400",
    green: "bg-green-400",
    red: "bg-red-400",
    yellow: "bg-yellow-400",
    purple: "bg-purple-400",
    gray: "bg-gray-400",
  };

  return colorMap[props.indicatorColor] || colorMap.blue;
});

// ESC 键关闭
const handleEscKey = (e) => {
  if (props.closeOnEsc && e.key === "Escape") {
    emit("close");
  }
};

// 焦点管理
const focusFirstElement = () => {
  if (props.initialFocus && dialogRef.value) {
    const initialFocusElem = dialogRef.value.querySelector(props.initialFocus);
    if (initialFocusElem) {
      initialFocusElem.focus();
    }
  }
};

// 焦点陷阱
const handleTabKey = (e) => {
  if (e.key !== "Tab" || !dialogRef.value) return;

  const focusableElements = dialogRef.value.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  );

  const firstElement = focusableElements[0];
  const lastElement = focusableElements[focusableElements.length - 1];

  if (e.shiftKey && document.activeElement === firstElement) {
    e.preventDefault();
    lastElement.focus();
  } else if (!e.shiftKey && document.activeElement === lastElement) {
    e.preventDefault();
    firstElement.focus();
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleEscKey);
  document.addEventListener("keydown", handleTabKey);
  document.body.style.overflow = "hidden";

  // 对话框打开时设置焦点
  setTimeout(() => {
    focusFirstElement();
  }, 50);
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleEscKey);
  document.removeEventListener("keydown", handleTabKey);
  document.body.style.overflow = "";
});
</script>

<style scoped>
/* 自定义滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}

/* 确保对话框有最大高度 */
@media (max-height: 640px) {
  .max-h-\[70vh\] {
    max-height: 90vh;
  }
}
</style>
