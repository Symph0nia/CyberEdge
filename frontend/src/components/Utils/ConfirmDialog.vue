<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div
        v-if="show"
        class="inset-0"
        @click="handleBackdropClick"
        @keydown.esc="onCancel"
        tabindex="-1"
      >
        <!-- 背景遮罩 -->
        <div
          class="inset-0 duration-300"
        ></div>

        <!-- 对话框 -->
        <div
          ref="dialogRef"
          class="max- border duration-300"
          @click.stop
        >
          <!-- 对话框标题和图标 -->
          <div >
            <div
              :class="[ ' ', type === 'danger' ? ' ' : type === 'warning' ? ' ' : ' ', ]"
            >
              <i
                :class="[ type === 'danger' ? 'ri-error-warning-line' : type === 'warning' ? 'ri-alert-line' : 'ri-question-line', '', ]"
              ></i>
            </div>
            <h2 >{{ title }}</h2>
          </div>

          <!-- 消息内容 -->
          <div >
            <p >
              {{ message }}
            </p>
          </div>

          <!-- 按钮区域 -->
          <div
            class="sm: sm: sm:"
          >
            <!-- 取消按钮 -->
            <button
              ref="cancelButtonRef"
              @click="onCancel"
              class="sm: .5 border hover: duration-200"
            >
              <span >
                <i class="ri-close-line .5"></i>
                {{ cancelText }}
              </span>
            </button>

            <!-- 确认按钮 -->
            <button
              ref="confirmButtonRef"
              @click="onConfirm"
              :class="[ 'sm: .5 border', ' duration-200', ' ', type === 'danger' ? ' hover: ' : type === 'warning' ? ' hover: ' : ' hover: ', ]"
            >
              <span >
                <i
                  :class="[ type === 'danger' ? 'ri-delete-bin-line' : type === 'warning' ? 'ri-alert-line' : 'ri-check-line', '.5', ]"
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
