<template>
  <div class="login-page">
    <div class="login-container">
      <a-card class="login-card" :bordered="false">
        <!-- 标题区域 -->
        <div class="login-header">
          <div class="logo-container">
            <i class="ri-shield-user-line logo-icon"></i>
          </div>
          <h2 class="login-title">登录账户</h2>
          <p class="login-subtitle">登录您的账户以访问完整功能</p>
        </div>

        <!-- 表单 -->
        <a-form
          :model="formState"
          name="login"
          @finish="handleLogin"
          @finish-failed="handleLoginFailed"
          autocomplete="off"
          layout="vertical"
          class="login-form"
          :rules="formRules"
        >
          <!-- 账户输入 -->
          <a-form-item
            name="account"
            class="form-item"
          >
            <template #label>
              <span class="form-label">
                <i class="ri-user-line"></i>
                账户
              </span>
            </template>
            <a-input
              v-model:value="formState.account"
              placeholder="输入账户名"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <i class="ri-account-circle-line input-icon"></i>
              </template>
            </a-input>
          </a-form-item>

          <!-- 验证码输入 -->
          <a-form-item
            name="code"
            class="form-item"
          >
            <template #label>
              <span class="form-label">
                <i class="ri-key-2-line"></i>
                验证码
              </span>
            </template>
            <a-input
              v-model:value="formState.code"
              placeholder="输入验证码"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <i class="ri-lock-line input-icon"></i>
              </template>
            </a-input>
          </a-form-item>

          <!-- 登录按钮 -->
          <a-form-item class="form-item login-button-item">
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              block
              :loading="loading"
              class="login-button"
            >
              <template #icon v-if="!loading">
                <i class="ri-login-box-line"></i>
              </template>
              {{ loading ? '登录中...' : '登录' }}
            </a-button>
          </a-form-item>

          <!-- 注册链接 -->
          <div class="register-link">
            <span class="register-text">还没有账户？</span>
            <router-link to="/setup-2fa" class="register-btn">
              注册账户
            </router-link>
          </div>
        </a-form>

        <!-- 错误提示 -->
        <div v-if="errorMessage" class="error-container">
          <a-alert
            :message="errorMessage"
            type="error"
            show-icon
            closable
            @close="clearError"
            class="error-alert"
          />
        </div>
      </a-card>
    </div>

    <!-- QR码组件 -->
    <GoogleAuthQRCode
      :isVisible="showQRCode"
      @close="showQRCode = false"
      @success="handleQRSuccess"
    />
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { message } from 'ant-design-vue';
import GoogleAuthQRCode from './GoogleAuthQRCode.vue';
import api from '../../api/axiosInstance';

export default {
  name: 'LoginPage',
  components: {
    GoogleAuthQRCode,
  },
  setup() {
    const router = useRouter();
    const store = useStore();

    // 表单状态
    const formState = reactive({
      account: '',
      code: '',
    });

    // 表单验证规则
    const formRules = {
      account: [
        { required: true, message: '请输入账户名', trigger: 'blur' },
        { min: 3, max: 20, message: '账户名长度应在3-20个字符之间', trigger: 'blur' },
      ],
      code: [
        { required: true, message: '请输入验证码', trigger: 'blur' },
        { len: 6, message: '验证码应为6位数字', trigger: 'blur' },
        { pattern: /^\d{6}$/, message: '验证码格式不正确', trigger: 'blur' },
      ],
    };

    // 组件状态
    const loading = ref(false);
    const errorMessage = ref('');
    const showQRCode = ref(false);

    // 清除错误信息
    const clearError = () => {
      errorMessage.value = '';
    };

    // 处理登录成功
    const handleLoginSuccess = (data) => {
      // 存储用户信息
      store.dispatch('login', {
        user: data.user,
        token: data.token,
      });

      message.success('登录成功！');

      // 跳转到主页
      setTimeout(() => {
        router.push('/');
      }, 1000);
    };

    // 处理登录
    const handleLogin = async (values) => {
      loading.value = true;
      clearError();

      try {
        const response = await api.post('/auth/login', {
          username: values.account,
          totp_code: values.code,
        });

        if (response.data && response.data.success) {
          handleLoginSuccess(response.data);
        } else {
          errorMessage.value = response.data?.message || '登录失败，请检查账户信息';
        }
      } catch (error) {
        console.error('登录错误:', error);

        if (error.response?.status === 401) {
          errorMessage.value = '账户名或验证码错误';
        } else if (error.response?.status === 429) {
          errorMessage.value = '登录尝试过于频繁，请稍后再试';
        } else if (error.response?.data?.message) {
          errorMessage.value = error.response.data.message;
        } else {
          errorMessage.value = '登录失败，请检查网络连接后重试';
        }
      } finally {
        loading.value = false;
      }
    };

    // 处理登录失败
    const handleLoginFailed = (errorInfo) => {
      console.error('表单验证失败:', errorInfo);
      message.error('请正确填写所有必填项');
    };

    // 处理QR码成功
    const handleQRSuccess = () => {
      showQRCode.value = false;
      message.success('配置完成，请使用Google Authenticator生成的验证码登录');
    };

    // 组件挂载时的初始化
    onMounted(() => {
      // 检查是否已经登录
      if (store.getters.isAuthenticated) {
        router.push('/');
        return;
      }

      // 自动聚焦到账户输入框
      const accountInput = document.querySelector('input[placeholder="输入账户名"]');
      if (accountInput) {
        setTimeout(() => accountInput.focus(), 100);
      }
    });

    return {
      formState,
      formRules,
      loading,
      errorMessage,
      showQRCode,
      clearError,
      handleLogin,
      handleLoginFailed,
      handleQRSuccess,
    };
  },
};
</script>

<style scoped>
.login-page {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  position: relative;
}

.login-page::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at 50% 50%, rgba(59, 130, 246, 0.1) 0%, transparent 70%);
  pointer-events: none;
}

.login-container {
  width: 100%;
  max-width: 420px;
  z-index: 1;
}

.login-card {
  background: rgba(30, 41, 59, 0.8) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(51, 65, 85, 0.3) !important;
  border-radius: 24px !important;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3) !important;
  overflow: hidden;
  animation: slideInUp 0.6s ease-out;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-container {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(147, 51, 234, 0.2));
  border: 1px solid rgba(59, 130, 246, 0.3);
  margin-bottom: 16px;
}

.logo-icon {
  font-size: 28px;
  color: #60a5fa;
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 8px;
  letter-spacing: -0.025em;
}

.login-subtitle {
  font-size: 14px;
  color: #94a3b8;
  line-height: 1.5;
  max-width: 300px;
  margin: 0 auto;
}

.login-form {
  margin-top: 24px;
}

.form-item {
  margin-bottom: 20px;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #cbd5e1;
  font-weight: 500;
  font-size: 14px;
}

.form-label i {
  font-size: 16px;
  color: #94a3b8;
}

.form-input :deep(.ant-input) {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 12px;
  color: #e2e8f0;
  padding: 12px 16px 12px 44px;
  font-size: 14px;
  transition: all 0.3s ease;
}

.form-input :deep(.ant-input::placeholder) {
  color: #64748b;
}

.form-input :deep(.ant-input:focus) {
  background: rgba(15, 23, 42, 0.8);
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.form-input :deep(.ant-input-prefix) {
  left: 16px;
}

.input-icon {
  color: #64748b;
  font-size: 16px;
}

.login-button-item {
  margin-bottom: 24px;
  margin-top: 32px;
}

.login-button {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  border-radius: 12px;
  height: 48px;
  font-weight: 600;
  font-size: 16px;
  color: #ffffff;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.login-button:hover {
  background: linear-gradient(135deg, #2563eb, #1e40af);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.4);
}

.login-button:active {
  transform: translateY(0);
}

.register-link {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid rgba(51, 65, 85, 0.3);
}

.register-text {
  color: #94a3b8;
  font-size: 14px;
}

.register-btn {
  color: #60a5fa;
  text-decoration: none;
  font-weight: 500;
  margin-left: 8px;
  transition: color 0.2s ease;
}

.register-btn:hover {
  color: #93c5fd;
  text-decoration: underline;
}

.error-container {
  margin-top: 16px;
}

.error-alert :deep(.ant-alert) {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
}

.error-alert :deep(.ant-alert-message) {
  color: #fca5a5;
}

/* Ant Design表单样式覆盖 */
.login-form :deep(.ant-form-item-label > label) {
  color: #cbd5e1;
  font-weight: 500;
}

.login-form :deep(.ant-form-item-explain-error) {
  color: #fca5a5;
  font-size: 12px;
  margin-top: 4px;
}

.login-form :deep(.ant-form-item-has-error .ant-input) {
  border-color: #ef4444;
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
}

/* 动画效果 */
@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-container {
    max-width: 100%;
  }

  .login-card {
    margin: 0 8px;
    border-radius: 16px !important;
  }

  .login-title {
    font-size: 20px;
  }

  .logo-container {
    width: 56px;
    height: 56px;
  }

  .logo-icon {
    font-size: 24px;
  }
}
</style>