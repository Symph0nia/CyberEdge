<template>
  <div style="padding: 24px; background: #f0f2f5; min-height: 100vh;">
    <!-- 页面头部 -->
    <div style="background: #fff; padding: 16px 24px; margin-bottom: 16px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
      <h1 style="margin: 0; font-size: 20px; font-weight: 500;">个人资料</h1>
      <p style="margin: 4px 0 0 0; color: #8c8c8c;">查看和管理您的个人信息</p>
    </div>

    <a-row :gutter="16">
      <!-- 基本信息 -->
      <a-col :span="16">
        <a-card title="基本信息" style="margin-bottom: 16px;">
          <a-form
            :model="userForm"
            :label-col="{ span: 4 }"
            :wrapper-col="{ span: 20 }"
          >
            <a-form-item label="用户名">
              <a-input v-model:value="userForm.username" disabled />
            </a-form-item>
            <a-form-item label="邮箱">
              <a-input v-model:value="userForm.email" disabled />
            </a-form-item>
            <a-form-item label="角色">
              <a-tag :color="userForm.role === 'admin' ? 'red' : 'blue'">
                {{ userForm.role === 'admin' ? '管理员' : '普通用户' }}
              </a-tag>
            </a-form-item>
            <a-form-item label="注册时间">
              <span>{{ formatDate(userForm.created_at) }}</span>
            </a-form-item>
            <a-form-item label="最后更新">
              <span>{{ formatDate(userForm.updated_at) }}</span>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- 修改密码 -->
        <a-card title="修改密码">
          <a-form
            :model="passwordForm"
            :label-col="{ span: 4 }"
            :wrapper-col="{ span: 20 }"
            @finish="handlePasswordChange"
          >
            <a-form-item
              label="当前密码"
              name="currentPassword"
              :rules="[{ required: true, message: '请输入当前密码' }]"
            >
              <a-input-password v-model:value="passwordForm.currentPassword" />
            </a-form-item>
            <a-form-item
              label="新密码"
              name="newPassword"
              :rules="[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码至少8位' }
              ]"
            >
              <a-input-password v-model:value="passwordForm.newPassword" />
            </a-form-item>
            <a-form-item
              label="确认密码"
              name="confirmPassword"
              :rules="[
                { required: true, message: '请确认新密码' },
                { validator: validateConfirmPassword }
              ]"
            >
              <a-input-password v-model:value="passwordForm.confirmPassword" />
            </a-form-item>
            <a-form-item :wrapper-col="{ offset: 4, span: 20 }">
              <a-button type="primary" html-type="submit" :loading="passwordLoading">
                更新密码
              </a-button>
              <a-button style="margin-left: 8px;" @click="resetPasswordForm">
                重置
              </a-button>
            </a-form-item>
          </a-form>
        </a-card>
      </a-col>

      <!-- 侧边栏 -->
      <a-col :span="8">
        <!-- 头像区域 -->
        <a-card style="margin-bottom: 16px; text-align: center;">
          <a-avatar :size="80" style="background-color: #1890ff; margin-bottom: 16px;">
            <template #icon><UserOutlined /></template>
          </a-avatar>
          <h3 style="margin: 0;">{{ userForm.username }}</h3>
          <p style="color: #8c8c8c; margin: 4px 0 0 0;">{{ userForm.email }}</p>
        </a-card>

        <!-- 双因子认证 -->
        <a-card title="双因子认证">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
            <span>TOTP 验证</span>
            <a-switch
              v-model:checked="is2FAEnabled"
              :loading="totpLoading"
              @change="handle2FAToggle"
            />
          </div>

          <div v-if="!is2FAEnabled">
            <p style="color: #8c8c8c; margin: 0; font-size: 14px;">
              启用双因子认证可以提高账户安全性
            </p>
          </div>

          <div v-else>
            <a-alert
              message="双因子认证已启用"
              description="您的账户已受到双因子认证保护"
              type="success"
              style="margin: 0;"
            />
          </div>

          <a-button
            v-if="is2FAEnabled"
            danger
            style="margin-top: 16px; width: 100%;"
            @click="handleDisable2FA"
            :loading="totpLoading"
          >
            禁用双因子认证
          </a-button>
        </a-card>
      </a-col>
    </a-row>

    <!-- 设置2FA的Modal -->
    <a-modal
      v-model:visible="setup2FAVisible"
      title="设置双因子认证"
      :footer="null"
      width="500px"
    >
      <div style="text-align: center;">
        <div v-if="qrCodeData">
          <p style="margin-bottom: 16px;">请使用认证应用扫描下方二维码：</p>
          <div v-html="qrCodeData" style="margin-bottom: 16px;"></div>
          <p style="color: #8c8c8c; font-size: 12px; margin-bottom: 16px;">
            推荐使用 Google Authenticator 或 Authy
          </p>
        </div>

        <a-form @finish="handleVerify2FA">
          <a-form-item
            label="验证码"
            name="code"
            :rules="[{ required: true, message: '请输入6位验证码' }]"
          >
            <a-input
              v-model:value="verifyCode"
              placeholder="请输入6位验证码"
              style="text-align: center; font-size: 18px; letter-spacing: 2px;"
              maxlength="6"
            />
          </a-form-item>
          <a-form-item style="margin: 0;">
            <a-button type="primary" html-type="submit" style="width: 100%;" :loading="verifyLoading">
              验证并启用
            </a-button>
          </a-form-item>
        </a-form>
      </div>
    </a-modal>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useStore } from 'vuex'
import { message } from 'ant-design-vue'
import { UserOutlined } from '@ant-design/icons-vue'
import api from '@/api/axiosInstance'

export default {
  name: 'ProfilePage',
  components: {
    UserOutlined
  },
  setup() {
    const store = useStore()

    const userForm = reactive({
      username: '',
      email: '',
      role: '',
      created_at: '',
      updated_at: ''
    })

    const passwordForm = reactive({
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })

    const passwordLoading = ref(false)
    const totpLoading = ref(false)
    const verifyLoading = ref(false)
    const is2FAEnabled = ref(false)
    const setup2FAVisible = ref(false)
    const qrCodeData = ref('')
    const verifyCode = ref('')

    const currentUser = computed(() => store.state.user || {})

    // 初始化用户信息
    const initUserInfo = () => {
      const user = currentUser.value
      userForm.username = user.username || ''
      userForm.email = user.email || ''
      userForm.role = user.role || ''
      userForm.created_at = user.created_at || ''
      userForm.updated_at = user.updated_at || ''
      is2FAEnabled.value = user.is_2fa_enabled || false
    }

    // 验证确认密码
    const validateConfirmPassword = async (rule, value) => {
      if (value !== passwordForm.newPassword) {
        return Promise.reject('两次输入的密码不一致')
      }
      return Promise.resolve()
    }

    // 修改密码
    const handlePasswordChange = async (values) => {
      passwordLoading.value = true
      try {
        // 这里需要根据实际API调整
        await api.post('/auth/change-password', {
          currentPassword: values.currentPassword,
          newPassword: values.newPassword
        })
        message.success('密码修改成功，请重新登录')
        resetPasswordForm()
        // 可能需要重新登录
        setTimeout(() => {
          store.dispatch('logout')
        }, 2000)
      } catch (error) {
        message.error('密码修改失败')
        console.error('修改密码失败:', error)
      } finally {
        passwordLoading.value = false
      }
    }

    // 重置密码表单
    const resetPasswordForm = () => {
      passwordForm.currentPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
    }

    // 处理2FA切换
    const handle2FAToggle = async (checked) => {
      if (checked) {
        await setup2FA()
      } else {
        // 关闭时直接调用禁用，不需要Modal
        is2FAEnabled.value = true // 回滚状态，等用户确认
      }
    }

    // 设置2FA
    const setup2FA = async () => {
      totpLoading.value = true
      try {
        const response = await api.post('/auth/2fa/setup', {
          username: userForm.username
        })
        qrCodeData.value = response.data.qr_code
        setup2FAVisible.value = true
      } catch (error) {
        message.error('设置双因子认证失败')
        is2FAEnabled.value = false // 回滚状态
        console.error('设置2FA失败:', error)
      } finally {
        totpLoading.value = false
      }
    }

    // 验证2FA
    const handleVerify2FA = async () => {
      if (!verifyCode.value || verifyCode.value.length !== 6) {
        message.error('请输入6位验证码')
        return
      }

      verifyLoading.value = true
      try {
        await api.post('/auth/2fa/verify', {
          username: userForm.username,
          code: verifyCode.value
        })
        message.success('双因子认证设置成功')
        setup2FAVisible.value = false
        is2FAEnabled.value = true
        verifyCode.value = ''
        qrCodeData.value = ''

        // 更新用户信息
        await store.dispatch('checkAuth')
      } catch (error) {
        message.error('验证码错误，请重试')
        console.error('验证2FA失败:', error)
      } finally {
        verifyLoading.value = false
      }
    }

    // 禁用2FA
    const handleDisable2FA = async () => {
      totpLoading.value = true
      try {
        await api.delete('/auth/2fa', {
          data: { username: userForm.username }
        })
        message.success('双因子认证已禁用')
        is2FAEnabled.value = false

        // 更新用户信息
        await store.dispatch('checkAuth')
      } catch (error) {
        message.error('禁用双因子认证失败')
        console.error('禁用2FA失败:', error)
      } finally {
        totpLoading.value = false
      }
    }

    // 格式化日期
    const formatDate = (timestamp) => {
      if (!timestamp) return '-'
      const date = new Date(timestamp * 1000) // Unix timestamp to JS timestamp
      return date.toLocaleString('zh-CN')
    }

    onMounted(() => {
      initUserInfo()
    })

    return {
      userForm,
      passwordForm,
      passwordLoading,
      totpLoading,
      verifyLoading,
      is2FAEnabled,
      setup2FAVisible,
      qrCodeData,
      verifyCode,
      validateConfirmPassword,
      handlePasswordChange,
      resetPasswordForm,
      handle2FAToggle,
      handleVerify2FA,
      handleDisable2FA,
      formatDate
    }
  }
}
</script>

<style scoped>
.ant-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.ant-form-item {
  margin-bottom: 16px;
}

.ant-avatar {
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
</style>