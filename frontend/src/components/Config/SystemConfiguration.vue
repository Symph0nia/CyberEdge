<template>
  <a-layout class="config-layout">
    <HeaderPage />
    <a-layout-content class="config-content">
      <div class="content-container">
        <a-card class="config-card" title="系统配置" :loading="loading">
          <template #extra>
            <a-button type="primary" @click="saveConfig" :loading="saving">
              <i class="ri-save-line"></i>
              保存配置
            </a-button>
          </template>

          <a-form
            :model="configForm"
            :label-col="{ span: 6 }"
            :wrapper-col="{ span: 18 }"
            class="config-form"
          >
            <a-divider orientation="left">数据库配置</a-divider>

            <a-form-item label="MongoDB URI">
              <a-input
                v-model:value="configForm.mongoUri"
                placeholder="mongodb://localhost:27017"
                size="large"
              />
            </a-form-item>

            <a-form-item label="Redis地址">
              <a-input
                v-model:value="configForm.redisAddr"
                placeholder="localhost:6379"
                size="large"
              />
            </a-form-item>

            <a-divider orientation="left">安全配置</a-divider>

            <a-form-item label="JWT Secret">
              <a-input-password
                v-model:value="configForm.jwtSecret"
                placeholder="请输入JWT密钥"
                size="large"
              />
            </a-form-item>

            <a-form-item label="Session Secret">
              <a-input-password
                v-model:value="configForm.sessionSecret"
                placeholder="请输入Session密钥"
                size="large"
              />
            </a-form-item>

            <a-form-item label="允许的域名">
              <a-input
                v-model:value="configForm.allowedOrigin"
                placeholder="https://yourdomain.com"
                size="large"
              />
            </a-form-item>

            <a-divider orientation="left">扫描配置</a-divider>

            <a-form-item label="最大并发任务">
              <a-input-number
                v-model:value="configForm.maxConcurrency"
                :min="1"
                :max="50"
                size="large"
                style="width: 100%"
              />
            </a-form-item>

            <a-form-item label="任务超时时间(秒)">
              <a-input-number
                v-model:value="configForm.taskTimeout"
                :min="30"
                :max="3600"
                size="large"
                style="width: 100%"
              />
            </a-form-item>

            <a-divider orientation="left">通知配置</a-divider>

            <a-form-item label="启用邮件通知">
              <a-switch v-model:checked="configForm.emailNotification" />
            </a-form-item>

            <a-form-item label="SMTP服务器" v-if="configForm.emailNotification">
              <a-input
                v-model:value="configForm.smtpServer"
                placeholder="smtp.gmail.com:587"
                size="large"
              />
            </a-form-item>

            <a-form-item label="邮件用户名" v-if="configForm.emailNotification">
              <a-input
                v-model:value="configForm.emailUsername"
                placeholder="your-email@gmail.com"
                size="large"
              />
            </a-form-item>

            <a-form-item label="邮件密码" v-if="configForm.emailNotification">
              <a-input-password
                v-model:value="configForm.emailPassword"
                placeholder="应用专用密码"
                size="large"
              />
            </a-form-item>
          </a-form>
        </a-card>
      </div>
    </a-layout-content>
    <FooterPage />
  </a-layout>
</template>

<script>
import { ref, onMounted } from 'vue';
import { message } from 'ant-design-vue';
import HeaderPage from '../HeaderPage.vue';
import FooterPage from '../FooterPage.vue';

export default {
  name: 'SystemConfiguration',
  components: {
    HeaderPage,
    FooterPage,
  },
  setup() {
    const loading = ref(false);
    const saving = ref(false);

    const configForm = ref({
      mongoUri: '',
      redisAddr: '',
      jwtSecret: '',
      sessionSecret: '',
      allowedOrigin: '',
      maxConcurrency: 10,
      taskTimeout: 300,
      emailNotification: false,
      smtpServer: '',
      emailUsername: '',
      emailPassword: ''
    });

    const loadConfig = async () => {
      loading.value = true;
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));

        // 模拟加载配置数据
        configForm.value = {
          mongoUri: 'mongodb://localhost:27017',
          redisAddr: 'localhost:6379',
          jwtSecret: '',
          sessionSecret: '',
          allowedOrigin: 'http://localhost:8080',
          maxConcurrency: 10,
          taskTimeout: 300,
          emailNotification: false,
          smtpServer: '',
          emailUsername: '',
          emailPassword: ''
        };
      } catch (error) {
        message.error('加载配置失败');
        console.error('加载配置失败:', error);
      } finally {
        loading.value = false;
      }
    };

    const saveConfig = async () => {
      saving.value = true;
      try {
        // 验证必填项
        if (!configForm.value.mongoUri) {
          message.error('MongoDB URI不能为空');
          return;
        }
        if (!configForm.value.redisAddr) {
          message.error('Redis地址不能为空');
          return;
        }

        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));

        message.success('配置保存成功');
      } catch (error) {
        message.error('保存配置失败');
        console.error('保存配置失败:', error);
      } finally {
        saving.value = false;
      }
    };

    onMounted(() => {
      loadConfig();
    });

    return {
      loading,
      saving,
      configForm,
      saveConfig
    };
  }
};
</script>

<style scoped>
.config-layout {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  min-height: 100vh;
}

.config-content {
  background: transparent;
  margin-top: 64px;
  padding: 0;
}

.content-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px;
}

.config-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 16px;
}

/* Ant Design组件样式覆盖 */
.config-card :deep(.ant-card-head) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.config-card :deep(.ant-card-head-title) {
  color: #f1f5f9;
  font-weight: 600;
  font-size: 20px;
}

.config-card :deep(.ant-card-body) {
  background: transparent;
  padding: 32px;
}

.config-card :deep(.ant-form-item-label > label) {
  color: #e2e8f0;
  font-weight: 500;
}

.config-card :deep(.ant-divider-horizontal.ant-divider-with-text) {
  border-color: rgba(51, 65, 85, 0.3);
}

.config-card :deep(.ant-divider-inner-text) {
  color: #f1f5f9;
  font-weight: 600;
}

.config-card :deep(.ant-input),
.config-card :deep(.ant-input-password),
.config-card :deep(.ant-input-number) {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  color: #e2e8f0;
}

.config-card :deep(.ant-input:hover),
.config-card :deep(.ant-input-password:hover),
.config-card :deep(.ant-input-number:hover) {
  border-color: rgba(59, 130, 246, 0.5);
}

.config-card :deep(.ant-input:focus),
.config-card :deep(.ant-input-password:focus),
.config-card :deep(.ant-input-number-focused) {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.config-card :deep(.ant-switch) {
  background: rgba(100, 116, 139, 0.5);
}

.config-card :deep(.ant-switch-checked) {
  background: #3b82f6;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .content-container {
    padding: 16px;
  }

  .config-card :deep(.ant-card-body) {
    padding: 24px;
  }

  .config-card :deep(.ant-form-item-label) {
    text-align: left;
  }

  .config-form :deep(.ant-form-item) {
    flex-direction: column;
  }

  .config-form :deep(.ant-form-item-label) {
    padding-bottom: 8px;
  }
}
</style>