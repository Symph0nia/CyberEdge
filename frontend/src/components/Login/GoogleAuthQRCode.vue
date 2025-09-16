<template>
  <div class="qr-container">
    <a-card class="qr-card" title="设置双重认证">
      <template #extra>
        <div class="icon-wrapper">
          <i class="ri-shield-keyhole-line"></i>
        </div>
      </template>

      <div class="qr-content">
        <p class="description">
          通过扫描二维码创建您的安全账户
        </p>

        <div class="qr-section">
          <div class="qr-code">
            <canvas ref="qrCanvas"></canvas>
          </div>
          <p class="qr-text">请使用 Google Authenticator 扫描此二维码</p>
        </div>

        <a-divider />

        <a-form @submit="verifyCode" layout="vertical" class="verify-form">
          <a-form-item label="验证码" required>
            <a-input
              v-model:value="verificationCode"
              placeholder="请输入6位验证码"
              maxlength="6"
              size="large"
              class="code-input"
            />
          </a-form-item>

          <a-form-item>
            <a-space direction="vertical" style="width: 100%">
              <a-button type="primary" html-type="submit" size="large" block>
                验证并启用
              </a-button>
              <a-button @click="$emit('back')" size="large" block>
                返回
              </a-button>
            </a-space>
          </a-form-item>
        </a-form>
      </div>
    </a-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue';
import QRCode from 'qrcode';

export default {
  name: 'GoogleAuthQRCode',
  props: {
    qrCodeUrl: {
      type: String,
      required: true
    }
  },
  emits: ['verify', 'back'],
  setup(props, { emit }) {
    const qrCanvas = ref(null);
    const verificationCode = ref('');

    const generateQRCode = async () => {
      if (qrCanvas.value && props.qrCodeUrl) {
        try {
          await QRCode.toCanvas(qrCanvas.value, props.qrCodeUrl);
        } catch (error) {
          console.error('生成二维码失败:', error);
        }
      }
    };

    const verifyCode = () => {
      if (verificationCode.value && verificationCode.value.length === 6) {
        emit('verify', verificationCode.value);
      }
    };

    onMounted(() => {
      generateQRCode();
    });

    return {
      qrCanvas,
      verificationCode,
      verifyCode
    };
  }
};
</script>

<style scoped>
.qr-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  padding: 24px;
}

.qr-card {
  width: 100%;
  max-width: 480px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

.icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: rgba(59, 130, 246, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-wrapper i {
  font-size: 20px;
  color: #3b82f6;
}

.qr-content {
  text-align: center;
}

.description {
  color: #94a3b8;
  margin-bottom: 32px;
  font-size: 16px;
  line-height: 1.6;
}

.qr-section {
  margin-bottom: 32px;
}

.qr-code {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  display: inline-block;
}

.qr-text {
  color: #cbd5e1;
  font-size: 14px;
  margin: 0;
}

.code-input {
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 4px;
}

/* Ant Design 样式覆盖 */
.qr-card :deep(.ant-card-head) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.qr-card :deep(.ant-card-head-title) {
  color: #f1f5f9;
  font-weight: 600;
  font-size: 20px;
}

.qr-card :deep(.ant-card-body) {
  background: transparent;
  padding: 32px;
}

.qr-card :deep(.ant-form-item-label > label) {
  color: #e2e8f0;
  font-weight: 500;
}

.qr-card :deep(.ant-divider) {
  border-color: rgba(51, 65, 85, 0.3);
}

/* 响应式设计 */
@media (max-width: 640px) {
  .qr-container {
    padding: 16px;
  }

  .qr-card :deep(.ant-card-body) {
    padding: 24px;
  }

  .description {
    font-size: 14px;
  }

  .code-input {
    font-size: 16px;
    letter-spacing: 2px;
  }
}
</style>