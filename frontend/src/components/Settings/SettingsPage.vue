<template>
  <div style="padding: 24px; background: #f0f2f5; min-height: 100vh;">
    <!-- 页面头部 -->
    <div style="background: #fff; padding: 16px 24px; margin-bottom: 16px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
      <h1 style="margin: 0; font-size: 20px; font-weight: 500;">系统设置</h1>
      <p style="margin: 4px 0 0 0; color: #8c8c8c;">管理系统配置和安全设置</p>
    </div>

    <a-row :gutter="16">
      <!-- 左侧设置菜单 -->
      <a-col :span="6">
        <a-card title="设置分类" style="margin-bottom: 16px;">
          <a-menu
            v-model:selectedKeys="selectedKeys"
            mode="vertical"
            style="border: none;"
            @click="handleMenuClick"
          >
            <a-menu-item key="general">
              <SettingOutlined />
              常规设置
            </a-menu-item>
            <a-menu-item key="security">
              <SafetyOutlined />
              安全设置
            </a-menu-item>
            <a-menu-item key="notification">
              <BellOutlined />
              通知设置
            </a-menu-item>
            <a-menu-item key="system">
              <DesktopOutlined />
              系统信息
            </a-menu-item>
          </a-menu>
        </a-card>
      </a-col>

      <!-- 右侧设置内容 -->
      <a-col :span="18">
        <!-- 常规设置 -->
        <div v-if="activeTab === 'general'">
          <a-card title="常规设置" style="margin-bottom: 16px;">
            <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
              <a-form-item label="系统名称">
                <a-input v-model:value="settings.systemName" placeholder="CyberEdge" />
              </a-form-item>
              <a-form-item label="系统描述">
                <a-textarea
                  v-model:value="settings.systemDescription"
                  placeholder="用户管理系统"
                  :rows="3"
                />
              </a-form-item>
              <a-form-item label="默认语言">
                <a-select v-model:value="settings.language" style="width: 200px;">
                  <a-select-option value="zh-CN">简体中文</a-select-option>
                  <a-select-option value="en-US">English</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="时区">
                <a-select v-model:value="settings.timezone" style="width: 200px;">
                  <a-select-option value="Asia/Shanghai">Asia/Shanghai</a-select-option>
                  <a-select-option value="UTC">UTC</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
                <a-button type="primary" @click="saveGeneralSettings" :loading="saveLoading">
                  保存设置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-card title="界面设置">
            <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
              <a-form-item label="主题模式">
                <a-radio-group v-model:value="settings.theme">
                  <a-radio value="light">浅色模式</a-radio>
                  <a-radio value="dark">深色模式</a-radio>
                  <a-radio value="auto">跟随系统</a-radio>
                </a-radio-group>
              </a-form-item>
              <a-form-item label="紧凑布局">
                <a-switch v-model:checked="settings.compactLayout" />
              </a-form-item>
              <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
                <a-button type="primary" @click="saveGeneralSettings" :loading="saveLoading">
                  保存设置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>
        </div>

        <!-- 安全设置 -->
        <div v-if="activeTab === 'security'">
          <a-card title="认证设置" style="margin-bottom: 16px;">
            <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
              <a-form-item label="密码策略">
                <a-space direction="vertical" style="width: 100%;">
                  <div>
                    <label>最小长度：</label>
                    <a-input-number
                      v-model:value="security.passwordMinLength"
                      :min="6"
                      :max="32"
                      style="width: 100px; margin-left: 8px;"
                    />
                    <span style="margin-left: 8px; color: #8c8c8c;">字符</span>
                  </div>
                  <a-checkbox v-model:checked="security.requireUppercase">
                    需要大写字母
                  </a-checkbox>
                  <a-checkbox v-model:checked="security.requireLowercase">
                    需要小写字母
                  </a-checkbox>
                  <a-checkbox v-model:checked="security.requireNumbers">
                    需要数字
                  </a-checkbox>
                  <a-checkbox v-model:checked="security.requireSpecialChars">
                    需要特殊字符
                  </a-checkbox>
                </a-space>
              </a-form-item>

              <a-form-item label="会话设置">
                <a-space direction="vertical" style="width: 100%;">
                  <div>
                    <label>会话超时：</label>
                    <a-input-number
                      v-model:value="security.sessionTimeout"
                      :min="30"
                      :max="1440"
                      style="width: 100px; margin-left: 8px;"
                    />
                    <span style="margin-left: 8px; color: #8c8c8c;">分钟</span>
                  </div>
                  <a-checkbox v-model:checked="security.rememberLogin">
                    允许记住登录状态
                  </a-checkbox>
                </a-space>
              </a-form-item>

              <a-form-item label="双因子认证">
                <a-space direction="vertical" style="width: 100%;">
                  <a-checkbox v-model:checked="security.force2FA">
                    强制启用双因子认证
                  </a-checkbox>
                  <a-checkbox v-model:checked="security.allow2FADisable">
                    允许用户禁用双因子认证
                  </a-checkbox>
                </a-space>
              </a-form-item>

              <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
                <a-button type="primary" @click="saveSecuritySettings" :loading="saveLoading">
                  保存设置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-card title="访问控制">
            <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
              <a-form-item label="登录限制">
                <a-space direction="vertical" style="width: 100%;">
                  <div>
                    <label>最大失败次数：</label>
                    <a-input-number
                      v-model:value="security.maxLoginAttempts"
                      :min="3"
                      :max="10"
                      style="width: 100px; margin-left: 8px;"
                    />
                    <span style="margin-left: 8px; color: #8c8c8c;">次</span>
                  </div>
                  <div>
                    <label>锁定时间：</label>
                    <a-input-number
                      v-model:value="security.lockoutDuration"
                      :min="5"
                      :max="60"
                      style="width: 100px; margin-left: 8px;"
                    />
                    <span style="margin-left: 8px; color: #8c8c8c;">分钟</span>
                  </div>
                </a-space>
              </a-form-item>

              <a-form-item label="IP白名单">
                <a-textarea
                  v-model:value="security.ipWhitelist"
                  placeholder="每行一个IP地址或CIDR格式"
                  :rows="4"
                />
              </a-form-item>

              <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
                <a-button type="primary" @click="saveSecuritySettings" :loading="saveLoading">
                  保存设置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>
        </div>

        <!-- 通知设置 -->
        <div v-if="activeTab === 'notification'">
          <a-card title="通知设置">
            <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
              <a-form-item label="邮件通知">
                <a-space direction="vertical" style="width: 100%;">
                  <a-checkbox v-model:checked="notification.emailEnabled">
                    启用邮件通知
                  </a-checkbox>
                  <a-checkbox v-model:checked="notification.loginNotification">
                    登录通知
                  </a-checkbox>
                  <a-checkbox v-model:checked="notification.securityAlert">
                    安全警告
                  </a-checkbox>
                  <a-checkbox v-model:checked="notification.systemUpdate">
                    系统更新
                  </a-checkbox>
                </a-space>
              </a-form-item>

              <a-form-item label="SMTP配置" v-if="notification.emailEnabled">
                <a-space direction="vertical" style="width: 100%;">
                  <a-input
                    v-model:value="notification.smtpHost"
                    placeholder="SMTP服务器地址"
                    addonBefore="服务器"
                  />
                  <a-input-number
                    v-model:value="notification.smtpPort"
                    placeholder="端口"
                    style="width: 100%;"
                    addonBefore="端口"
                  />
                  <a-input
                    v-model:value="notification.smtpUser"
                    placeholder="用户名"
                    addonBefore="用户名"
                  />
                  <a-input-password
                    v-model:value="notification.smtpPassword"
                    placeholder="密码"
                    addonBefore="密码"
                  />
                  <a-checkbox v-model:checked="notification.smtpSSL">
                    使用SSL加密
                  </a-checkbox>
                </a-space>
              </a-form-item>

              <a-form-item :wrapper-col="{ offset: 6, span: 18 }">
                <a-button type="primary" @click="saveNotificationSettings" :loading="saveLoading">
                  保存设置
                </a-button>
                <a-button style="margin-left: 8px;" @click="testEmailSettings" :loading="testLoading">
                  测试邮件
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>
        </div>

        <!-- 系统信息 -->
        <div v-if="activeTab === 'system'">
          <a-card title="系统信息" style="margin-bottom: 16px;">
            <a-descriptions :column="2" bordered>
              <a-descriptions-item label="系统版本">{{ systemInfo.version }}</a-descriptions-item>
              <a-descriptions-item label="构建时间">{{ systemInfo.buildTime }}</a-descriptions-item>
              <a-descriptions-item label="Go版本">{{ systemInfo.goVersion }}</a-descriptions-item>
              <a-descriptions-item label="运行时间">{{ systemInfo.uptime }}</a-descriptions-item>
              <a-descriptions-item label="数据库">{{ systemInfo.database }}</a-descriptions-item>
              <a-descriptions-item label="当前连接数">{{ systemInfo.connections }}</a-descriptions-item>
            </a-descriptions>
          </a-card>

          <a-card title="系统状态">
            <a-row :gutter="16">
              <a-col :span="8">
                <a-statistic
                  title="CPU使用率"
                  :value="systemStatus.cpu"
                  suffix="%"
                  :value-style="{ color: systemStatus.cpu > 80 ? '#f5222d' : '#3f8600' }"
                />
              </a-col>
              <a-col :span="8">
                <a-statistic
                  title="内存使用率"
                  :value="systemStatus.memory"
                  suffix="%"
                  :value-style="{ color: systemStatus.memory > 80 ? '#f5222d' : '#3f8600' }"
                />
              </a-col>
              <a-col :span="8">
                <a-statistic
                  title="磁盘使用率"
                  :value="systemStatus.disk"
                  suffix="%"
                  :value-style="{ color: systemStatus.disk > 80 ? '#f5222d' : '#3f8600' }"
                />
              </a-col>
            </a-row>

            <a-divider />

            <a-space>
              <a-button @click="refreshSystemInfo" :loading="refreshLoading">
                刷新信息
              </a-button>
              <a-button @click="downloadLogs">
                下载日志
              </a-button>
              <a-button danger @click="showRestartConfirm">
                重启系统
              </a-button>
            </a-space>
          </a-card>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  SettingOutlined,
  SafetyOutlined,
  BellOutlined,
  DesktopOutlined
} from '@ant-design/icons-vue'

export default {
  name: 'SettingsPage',
  components: {
    SettingOutlined,
    SafetyOutlined,
    BellOutlined,
    DesktopOutlined
  },
  setup() {
    const selectedKeys = ref(['general'])
    const activeTab = ref('general')
    const saveLoading = ref(false)
    const testLoading = ref(false)
    const refreshLoading = ref(false)

    const settings = reactive({
      systemName: 'CyberEdge',
      systemDescription: '用户管理系统',
      language: 'zh-CN',
      timezone: 'Asia/Shanghai',
      theme: 'light',
      compactLayout: false
    })

    const security = reactive({
      passwordMinLength: 8,
      requireUppercase: true,
      requireLowercase: true,
      requireNumbers: true,
      requireSpecialChars: false,
      sessionTimeout: 120,
      rememberLogin: true,
      force2FA: false,
      allow2FADisable: true,
      maxLoginAttempts: 5,
      lockoutDuration: 15,
      ipWhitelist: ''
    })

    const notification = reactive({
      emailEnabled: false,
      loginNotification: true,
      securityAlert: true,
      systemUpdate: false,
      smtpHost: '',
      smtpPort: 587,
      smtpUser: '',
      smtpPassword: '',
      smtpSSL: true
    })

    const systemInfo = reactive({
      version: '1.0.0',
      buildTime: '2024-01-01 12:00:00',
      goVersion: 'go1.22.2',
      uptime: '1天2小时30分钟',
      database: 'MySQL 8.0',
      connections: 5
    })

    const systemStatus = reactive({
      cpu: 25,
      memory: 45,
      disk: 60
    })

    const handleMenuClick = ({ key }) => {
      activeTab.value = key
      selectedKeys.value = [key]
    }

    const saveGeneralSettings = async () => {
      saveLoading.value = true
      try {
        // 这里调用实际的API保存设置
        await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
        message.success('常规设置保存成功')
      } catch (error) {
        message.error('保存设置失败')
        console.error('保存设置失败:', error)
      } finally {
        saveLoading.value = false
      }
    }

    const saveSecuritySettings = async () => {
      saveLoading.value = true
      try {
        // 这里调用实际的API保存设置
        await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
        message.success('安全设置保存成功')
      } catch (error) {
        message.error('保存设置失败')
        console.error('保存设置失败:', error)
      } finally {
        saveLoading.value = false
      }
    }

    const saveNotificationSettings = async () => {
      saveLoading.value = true
      try {
        // 这里调用实际的API保存设置
        await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
        message.success('通知设置保存成功')
      } catch (error) {
        message.error('保存设置失败')
        console.error('保存设置失败:', error)
      } finally {
        saveLoading.value = false
      }
    }

    const testEmailSettings = async () => {
      testLoading.value = true
      try {
        // 这里调用实际的API测试邮件
        await new Promise(resolve => setTimeout(resolve, 2000)) // 模拟API调用
        message.success('测试邮件发送成功')
      } catch (error) {
        message.error('测试邮件发送失败')
        console.error('测试邮件失败:', error)
      } finally {
        testLoading.value = false
      }
    }

    const refreshSystemInfo = async () => {
      refreshLoading.value = true
      try {
        // 这里调用实际的API获取系统信息
        await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用

        // 模拟更新系统状态
        systemStatus.cpu = Math.floor(Math.random() * 100)
        systemStatus.memory = Math.floor(Math.random() * 100)
        systemStatus.disk = Math.floor(Math.random() * 100)

        message.success('系统信息已刷新')
      } catch (error) {
        message.error('刷新系统信息失败')
        console.error('刷新系统信息失败:', error)
      } finally {
        refreshLoading.value = false
      }
    }

    const downloadLogs = () => {
      // 这里实现日志下载
      message.info('日志下载功能开发中...')
    }

    const showRestartConfirm = () => {
      Modal.confirm({
        title: '重启系统确认',
        content: '确定要重启系统吗？这将断开所有用户连接。',
        okText: '确认重启',
        cancelText: '取消',
        okType: 'danger',
        onOk: () => {
          message.info('系统重启功能开发中...')
        }
      })
    }

    onMounted(() => {
      // 这里可以加载设置数据
    })

    return {
      selectedKeys,
      activeTab,
      saveLoading,
      testLoading,
      refreshLoading,
      settings,
      security,
      notification,
      systemInfo,
      systemStatus,
      handleMenuClick,
      saveGeneralSettings,
      saveSecuritySettings,
      saveNotificationSettings,
      testEmailSettings,
      refreshSystemInfo,
      downloadLogs,
      showRestartConfirm
    }
  }
}
</script>

<style scoped>
.ant-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.ant-menu {
  background: transparent;
}

.ant-menu-item-selected {
  background-color: #e6f7ff !important;
  border-radius: 6px;
}

.ant-form-item {
  margin-bottom: 16px;
}

.ant-descriptions-item {
  padding: 12px 16px;
}

.ant-statistic {
  text-align: center;
}
</style>