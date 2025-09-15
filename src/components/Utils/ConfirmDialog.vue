<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div
        v-if="show"
        class="fixed inset-0 flex items-center justify-center z-50 px-4"
        @click="handleBackdropClick"
        @keydown.esc="onCancel"
        tabindex="-1"
      >
        <!-- 背景遮罩 -->
        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity duration-300"
        ></div>

        <!-- 对话框 -->
        <div
          ref="dialogRef"
          class="bg-gray-800/90 backdrop-blur-xl relative w-full max-w-md p-8 rounded-2xl shadow-2xl border border-gray-700/30 transform transition-all duration-300"
          @click.stop
        >
          <!-- 对话框标题和图标 -->
          <div class="flex items-center mb-4">
            <div
              :class="[
                'w-10 h-10 rounded-lg flex items-center justify-center mr-3',
                type === 'danger'
                  ? 'bg-red-500/20 text-red-400'
                  : type === 'warning'
                  ? 'bg-yellow-500/20 text-yellow-400'
                  : 'bg-blue-500/20 text-blue-400',
              ]"
            >
              <i
                :class="[
                  type === 'danger'
                    ? 'ri-error-warning-line'
                    : type === 'warning'
                    ? 'ri-alert-line'
                    : 'ri-question-line',
                  'text-xl',
                ]"
              ></i>
            </div>
            <h2 class="text-lg font-medium text-gray-200">{{ title }}</h2>
          </div>

          <!-- 消息内容 -->
          <div class="ml-13 mb-6">
            <p class="text-sm text-gray-300 leading-relaxed">
              {{ message }}
            </p>
          </div>

          <!-- 按钮区域 -->
          <div
            class="flex flex-col-reverse sm:flex-row sm:space-x-3 space-y-2 space-y-reverse sm:space-y-0"
          >
            <!-- 取消按钮 -->
            <button
              ref="cancelButtonRef"
              @click="onCancel"
              class="sm:flex-1 px-4 py-2.5 rounded-xl border border-gray-600/30 bg-gray-700/50 hover:bg-gray-600/50 text-sm font-medium text-gray-200 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-600/50"
            >
              <span class="flex items-center justify-center">
                <i class="ri-close-line mr-1.5"></i>
                {{ cancelText }}
              </span>
            </button>

            <!-- 确认按钮 -->
            <button
              ref="confirmButtonRef"
              @click="onConfirm"
              :class="[
                'sm:flex-1 px-4 py-2.5 rounded-xl text-sm font-medium border',
                'transition-all duration-200',
                'focus:outline-none focus:ring-2',
                type === 'danger'
                  ? 'bg-red-500/20 hover:bg-red-500/30 focus:ring-red-500/50 text-red-300 border-red-500/30'
                  : type === 'warning'
                  ? 'bg-yellow-500/20 hover:bg-yellow-500/30 focus:ring-yellow-500/50 text-yellow-300 border-yellow-500/30'
                  : 'bg-blue-500/20 hover:bg-blue-500/30 focus:ring-blue-500/50 text-blue-300 border-blue-500/30',
              ]"
            >
              <span class="flex items-center justify-center">
                <i
                  :class="[
                    type === 'danger'
                      ? 'ri-delete-bin-line'
                      : type === 'warning'
                      ? 'ri-alert-line'
                      : 'ri-check-line',
                    'mr-1.5',
                  ]"
                ></i>
                {{ confirmText }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick, onMounted, onUnmounted } from "vue";

// eslint-disable-next-line no-undef
const props = defineProps({
  show: Boolean,
  title: {
    type: String,
    default: "确认操作",
  },
  message: {
    type: String,
    required: true,
  },
  type: {
    type: String,
    default: "info",
    validator: (value) => ["info", "warning", "danger"].includes(value),
  },
  confirmText: {
    type: String,
    default: "确认",
  },
  cancelText: {
    type: String,
    default: "取消",
  },
  closeOnBackdrop: {
    type: Boolean,
    default: true,
  },
});

// 定义事件
// eslint-disable-next-line no-undef
const emit = defineEmits(["confirm", "cancel"]);

// 引用元素
const dialogRef = ref(null);
const confirmButtonRef = ref(null);
const cancelButtonRef = ref(null);

// 处理确认和取消操作
const onConfirm = () => {
  emit("confirm");
};

const onCancel = () => {
  emit("cancel");
};

// 处理背景点击
const handleBackdropClick = () => {
  if (props.closeOnBackdrop) {
    onCancel();
  }
};

// 监听显示状态，自动聚焦按钮
watch(
  () => props.show,
  async (newVal) => {
    if (newVal) {
      await nextTick();
      // 根据对话框类型决定默认聚焦的按钮
      if (props.type === "danger") {
        cancelButtonRef.value?.focus();
      } else {
        confirmButtonRef.value?.focus();
      }

      // 当对话框显示时，锁定背景滚动
      document.body.style.overflow = "hidden";
    } else {
      // 当对话框隐藏时，恢复背景滚动
      document.body.style.overflow = "";
    }
  },
  { immediate: true }
);

// 键盘事件监听
onMounted(() => {
  const handleKeyDown = (e) => {
    if (!props.show) return;

    if (e.key === "Escape") {
      onCancel();
    } else if (e.key === "Enter") {
      onConfirm();
    }
  };

  window.addEventListener("keydown", handleKeyDown);

  // 清理
  onUnmounted(() => {
    window.removeEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "";
  });
});
</script>

<style scoped>
.dialog-enter-active,
.dialog-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

/* 优化按钮点击效果 */
button:active:not(:disabled) {
  transform: scale(0.98);
}

/* 确保模糊效果在所有浏览器中正常工作 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

.backdrop-blur-sm {
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

@media (max-width: 640px) {
  .ml-13 {
    margin-left: 3.25rem;
  }
}
</style>
