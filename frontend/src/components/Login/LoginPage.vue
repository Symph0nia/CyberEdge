<template>
  <div style="min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">

    <!-- 主登录表单 -->
    <a-card
      v-if="!show2FA"
      style="width: 400px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15);"
      :bordered="false"
    >
      <template #title>
        <div style="text-align: center;">
          <a-avatar :size="64" style="background-color: #1890ff; margin-bottom: 16px;">
            <template #icon><UserOutlined /></template>
          </a-avatar>
          <div style="font-size: 24px; font-weight: 600; margin-bottom: 8px;">CyberEdge</div>
          <div style="color: #8c8c8c; font-size: 14px;">欢迎回来，请登录您的账户</div>
        </div>
      </template>

      <a-form
        :model="loginForm"
        :rules="loginRules"
        @finish="handleLogin"
        layout="vertical"
        size="large"
        data-testid="login-form"
      >
        <a-form-item name="username" label="用户名">
          <a-input
            v-model:value="loginForm.username"
            placeholder="请输入用户名"
            :prefix="h(UserOutlined)"
            data-testid="username-input"
          />
        </a-form-item>

        <a-form-item name="password" label="密码">
          <a-input-password
            v-model:value="loginForm.password"
            placeholder="请输入密码"
            :prefix="h(LockOutlined)"
            data-testid="password-input"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            block
            :loading="loginLoading"
            style="height: 40px; font-size: 16px;"
            data-testid="login-button"
          >
            登录
          </a-button>
        </a-form-item>

        <a-divider>
          <span style="color: #8c8c8c; font-size: 12px;">没有账户？</span>
        </a-divider>

        <a-form-item>
          <a-button
            block
            @click="showRegister = true"
            style="height: 40px;"
          >
            注册新账户
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 2FA验证表单 -->
    <a-card
      v-if="show2FA"
      style="width: 400px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15);"
      :bordered="false"
    >
      <template #title>
        <div style="text-align: center;">
          <a-avatar :size="64" style="background-color: #52c41a; margin-bottom: 16px;">
            <template #icon><SafetyOutlined /></template>
          </a-avatar>
          <div style="font-size: 24px; font-weight: 600; margin-bottom: 8px;">双重认证</div>
          <div style="color: #8c8c8c; font-size: 14px;">请输入您的6位验证码</div>
        </div>
      </template>

      <a-form
        :model="totpForm"
        :rules="totpRules"
        @finish="handle2FALogin"
        layout="vertical"
        size="large"
      >
        <a-form-item name="totp" label="验证码">
          <a-input
            v-model:value="totpForm.totp"
            placeholder="请输入6位验证码"
            :prefix="h(KeyOutlined)"
            :maxlength="6"
            style="text-align: center; font-size: 18px; letter-spacing: 4px;"
          />
        </a-form-item>

        <a-form-item>
          <a-space style="width: 100%;" direction="vertical" :size="12">
            <a-button
              type="primary"
              html-type="submit"
              block
              :loading="totpLoading"
              style="height: 40px; font-size: 16px;"
            >
              验证并登录
            </a-button>
            <a-button
              block
              @click="goBackToLogin"
              style="height: 40px;"
            >
              返回登录
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 注册表单 -->
    <a-modal
      v-model:open="showRegister"
      title="注册新账户"
      :footer="null"
      width="480px"
    >
      <a-form
        :model="registerForm"
        :rules="registerRules"
        @finish="handleRegister"
        layout="vertical"
        size="large"
      >
        <a-form-item name="username" label="用户名">
          <a-input
            v-model:value="registerForm.username"
            placeholder="请输入用户名"
            :prefix="h(UserOutlined)"
          />
        </a-form-item>

        <a-form-item name="email" label="邮箱">
          <a-input
            v-model:value="registerForm.email"
            placeholder="请输入邮箱"
            :prefix="h(MailOutlined)"
          />
        </a-form-item>

        <a-form-item name="password" label="密码">
          <a-input-password
            v-model:value="registerForm.password"
            placeholder="请输入密码"
            :prefix="h(LockOutlined)"
          />
        </a-form-item>

        <a-form-item name="confirmPassword" label="确认密码">
          <a-input-password
            v-model:value="registerForm.confirmPassword"
            placeholder="请再次输入密码"
            :prefix="h(LockOutlined)"
          />
        </a-form-item>

        <a-form-item>
          <a-space style="width: 100%;" direction="vertical" :size="12">
            <a-button
              type="primary"
              html-type="submit"
              block
              :loading="registerLoading"
              style="height: 40px; font-size: 16px;"
            >
              注册账户
            </a-button>
            <a-button
              block
              @click="showRegister = false"
              style="height: 40px;"
            >
              取消
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 2FA设置模态框 -->
    <a-modal
      v-model:open="showQRSetup"
      title="设置双重认证"
      :footer="null"
      width="480px"
      :closable="false"
      :maskClosable="false"
    >
      <div style="text-align: center;">
        <a-steps :current="qrStep" style="margin-bottom: 24px;">
          <a-step title="扫描二维码" />
          <a-step title="验证设置" />
        </a-steps>

        <!-- 步骤1: 显示二维码 -->
        <div v-if="qrStep === 0">
          <a-alert
            message="请使用Google Authenticator扫描下方二维码"
            type="info"
            style="margin-bottom: 16px;"
          />
          <div style="display: flex; justify-content: center; margin-bottom: 16px;">
            <img v-if="qrCodeUrl" :src="qrCodeUrl" alt="QR Code" style="width: 200px; height: 200px; border: 1px solid #d9d9d9;" />
            <a-spin v-else :size="'large'" />
          </div>
          <a-button type="primary" @click="qrStep = 1" :disabled="!qrCodeUrl">
            下一步
          </a-button>
        </div>

        <!-- 步骤2: 验证设置 -->
        <div v-if="qrStep === 1">
          <a-alert
            message="请输入Google Authenticator中显示的6位验证码"
            type="info"
            style="margin-bottom: 16px;"
          />
          <a-form
            :model="verifyForm"
            @finish="handleVerify2FA"
            layout="vertical"
            size="large"
          >
            <a-form-item name="code">
              <a-input
                v-model:value="verifyForm.code"
                placeholder="请输入6位验证码"
                :maxlength="6"
                style="text-align: center; font-size: 18px; letter-spacing: 4px;"
              />
            </a-form-item>
            <a-form-item>
              <a-space>
                <a-button @click="qrStep = 0">上一步</a-button>
                <a-button type="primary" html-type="submit" :loading="verifyLoading">
                  完成设置
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </div>
      </div>
    </a-modal>

  </div>
</template>

<script>
import { ref, reactive, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { message } from 'ant-design-vue'
import {
  UserOutlined,
  LockOutlined,
  SafetyOutlined,
  KeyOutlined,
  MailOutlined
} from '@ant-design/icons-vue'
import api from '../../api/axiosInstance'

export default {
  name: 'LoginPage',
  setup() {
    const router = useRouter()
    const store = useStore()

    // 状态管理
    const show2FA = ref(false)
    const showRegister = ref(false)
    const showQRSetup = ref(false)
    const qrStep = ref(0)
    const qrCodeUrl = ref('')
    const pendingUsername = ref('')

    // 加载状态
    const loginLoading = ref(false)
    const totpLoading = ref(false)
    const registerLoading = ref(false)
    const verifyLoading = ref(false)

    // 表单数据
    const loginForm = reactive({
      username: '',
      password: ''
    })

    const totpForm = reactive({
      totp: ''
    })

    const registerForm = reactive({
      username: '',
      email: '',
      password: '',
      confirmPassword: ''
    })

    const verifyForm = reactive({
      code: ''
    })

    // 表单验证规则
    const loginRules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' }
      ]
    }

    const totpRules = {
      totp: [
        { required: true, message: '请输入验证码', trigger: 'blur' },
        { len: 6, message: '验证码必须是6位', trigger: 'blur' },
        { pattern: /^\d{6}$/, message: '验证码必须是6位数字', trigger: 'blur' }
      ]
    }

    const registerRules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
        { min: 3, max: 20, message: '用户名长度应在3-20个字符之间', trigger: 'blur' }
      ],
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度至少6位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请确认密码', trigger: 'blur' },
        {
          validator: (_, value) => {
            if (value !== registerForm.password) {
              return Promise.reject('两次输入的密码不一致')
            }
            return Promise.resolve()
          },
          trigger: 'blur'
        }
      ]
    }

    // 处理普通登录
    const handleLogin = async (values) => {
      loginLoading.value = true
      try {
        const response = await api.post('/auth/login', {
          username: values.username,
          password: values.password
        })

        if (response.data.token) {
          // 登录成功，直接设置token并跳转
          handleLoginSuccess(response.data)
        } else {
          message.error(response.data.message || '登录失败')
        }
      } catch (error) {
        console.error('登录错误:', error)
        if (error.response?.status === 401) {
          message.error('用户名或密码错误')
        } else {
          message.error('登录失败，请检查网络连接')
        }
      } finally {
        loginLoading.value = false
      }
    }

    // 处理2FA登录
    const handle2FALogin = async (values) => {
      totpLoading.value = true
      try {
        const response = await api.post('/auth/validate', {
          username: pendingUsername.value,
          totp_code: values.totp
        })

        if (response.data.success) {
          handleLoginSuccess(response.data)
        } else {
          message.error(response.data.message || '验证码错误')
        }
      } catch (error) {
        console.error('2FA验证错误:', error)
        message.error('验证失败，请重试')
      } finally {
        totpLoading.value = false
      }
    }

    // 处理注册
    const handleRegister = async (values) => {
      registerLoading.value = true
      try {
        const response = await api.post('/auth/register', {
          username: values.username,
          email: values.email,
          password: values.password
        })

        if (response.data.success) {
          message.success('注册成功，正在设置双重认证...')
          showRegister.value = false
          pendingUsername.value = values.username
          await setup2FA()
        } else {
          message.error(response.data.message || '注册失败')
        }
      } catch (error) {
        console.error('注册错误:', error)
        message.error('注册失败，请重试')
      } finally {
        registerLoading.value = false
      }
    }

    // 设置2FA
    const setup2FA = async () => {
      try {
        const response = await api.post('/auth/2fa/setup', {
          username: pendingUsername.value
        })

        if (response.data.success) {
          qrCodeUrl.value = response.data.qr_code
          showQRSetup.value = true
          qrStep.value = 0
        } else {
          message.error('设置失败')
        }
      } catch (error) {
        console.error('设置2FA错误:', error)
        message.error('设置失败，请重试')
      }
    }

    // 验证2FA设置
    const handleVerify2FA = async (values) => {
      verifyLoading.value = true
      try {
        const response = await api.post('/auth/2fa/verify', {
          username: pendingUsername.value,
          code: values.code
        })

        if (response.data.success) {
          message.success('双重认证设置成功！')
          showQRSetup.value = false
          qrStep.value = 0
          verifyForm.code = ''
          // 重置到登录页面
          show2FA.value = false
          loginForm.username = ''
          loginForm.password = ''
        } else {
          message.error(response.data.message || '验证失败')
        }
      } catch (error) {
        console.error('验证2FA错误:', error)
        message.error('验证失败，请重试')
      } finally {
        verifyLoading.value = false
      }
    }

    // 登录成功处理
    const handleLoginSuccess = (data) => {
      // 保存token到localStorage
      localStorage.setItem('token', data.token)

      store.dispatch('login', {
        token: data.token
      })
      message.success('登录成功！')
      router.push('/user-management')
    }

    // 返回登录
    const goBackToLogin = () => {
      show2FA.value = false
      totpForm.totp = ''
    }

    // 检查登录状态
    onMounted(() => {
      if (store.getters.isAuthenticated) {
        router.push('/')
      }
    })

    return {
      h,
      show2FA,
      showRegister,
      showQRSetup,
      qrStep,
      qrCodeUrl,
      loginLoading,
      totpLoading,
      registerLoading,
      verifyLoading,
      loginForm,
      totpForm,
      registerForm,
      verifyForm,
      loginRules,
      totpRules,
      registerRules,
      handleLogin,
      handle2FALogin,
      handleRegister,
      handleVerify2FA,
      goBackToLogin,
      UserOutlined,
      LockOutlined,
      SafetyOutlined,
      KeyOutlined,
      MailOutlined
    }
  }
}
</script>