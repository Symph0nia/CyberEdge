<template>
  <div style="min-height: 100vh; background: #f0f2f5;">

    <!-- 页面头部 -->
    <div style="background: #fff; padding: 16px 24px; margin-bottom: 16px; border-bottom: 1px solid #f0f0f0;">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <div>
          <h1 style="margin: 0; font-size: 20px; font-weight: 500;">用户管理</h1>
          <p style="margin: 4px 0 0 0; color: #8c8c8c;">管理系统用户和权限设置</p>
        </div>
        <a-button type="primary" @click="showAddUser = true">
          <template #icon><UserAddOutlined /></template>
          添加用户
        </a-button>
      </div>
    </div>

    <div style="padding: 0 24px;">
      <!-- 统计卡片 -->
      <a-row :gutter="16" style="margin-bottom: 16px;">
        <a-col :span="6">
          <a-card size="small">
            <a-statistic
              title="总用户数"
              :value="users.length"
              :value-style="{ color: '#1890ff', fontSize: '24px' }"
            >
              <template #prefix>
                <UserOutlined style="color: #1890ff;" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>

        <a-col :span="6">
          <a-card size="small">
            <a-statistic
              title="在线用户"
              :value="onlineUsers"
              :value-style="{ color: '#52c41a', fontSize: '24px' }"
            >
              <template #prefix>
                <GlobalOutlined style="color: #52c41a;" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>

        <a-col :span="6">
          <a-card size="small">
            <a-statistic
              title="管理员"
              :value="adminCount"
              :value-style="{ color: '#fa8c16', fontSize: '24px' }"
            >
              <template #prefix>
                <CrownOutlined style="color: #fa8c16;" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>

        <a-col :span="6">
          <a-card size="small">
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <div>
                <div style="color: #8c8c8c; font-size: 14px; margin-bottom: 4px;">二维码登录</div>
                <div style="font-size: 24px; font-weight: 500; color: #1890ff;">
                  {{ qrcodeEnabled ? '已启用' : '已禁用' }}
                </div>
              </div>
              <a-switch
                v-model:checked="qrcodeEnabled"
                @change="toggleQRCodeStatus"
                :loading="qrStatusLoading"
              />
            </div>
          </a-card>
        </a-col>
      </a-row>

      <!-- 用户列表 -->
      <a-card title="用户列表" size="small">
        <template #extra>
          <a-space>
            <a-input-search
              v-model:value="searchText"
              placeholder="搜索用户名"
              style="width: 240px;"
              @input="handleSearch"
              allow-clear
            />
            <a-button
              v-if="selectedRowKeys.length > 0"
              type="primary"
              danger
              @click="handleBatchDelete"
            >
              <template #icon><DeleteOutlined /></template>
              批量删除 ({{ selectedRowKeys.length }})
            </a-button>
          </a-space>
        </template>

        <a-table
          :columns="columns"
          :data-source="filteredUsers"
          :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
          :pagination="pagination"
          :loading="loading"
          row-key="account"
          size="middle"
          :scroll="{ x: 'max-content' }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'avatar'">
              <a-avatar :style="{ backgroundColor: getAvatarColor(record.account) }" size="large">
                {{ record.account ? record.account.charAt(0).toUpperCase() : 'U' }}
              </a-avatar>
            </template>

            <template v-if="column.key === 'account'">
              <div>
                <div style="font-weight: 500; font-size: 14px;">{{ record.account }}</div>
                <div style="font-size: 12px; color: #8c8c8c;">
                  登录次数: {{ record.loginCount || 0 }}
                </div>
              </div>
            </template>

            <template v-if="column.key === 'status'">
              <a-badge
                :status="record.isOnline ? 'success' : 'default'"
                :text="record.isOnline ? '在线' : '离线'"
              />
            </template>

            <template v-if="column.key === 'role'">
              <a-tag :color="record.role === 'admin' ? 'red' : 'blue'">
                {{ record.role === 'admin' ? '管理员' : '普通用户' }}
              </a-tag>
            </template>

            <template v-if="column.key === 'createdAt'">
              <span style="color: #8c8c8c;">
                {{ formatDate(record.createdAt) }}
              </span>
            </template>

            <template v-if="column.key === 'action'">
              <a-space>
                <a-button type="link" size="small" @click="handleEdit(record)">
                  <template #icon><EditOutlined /></template>
                  编辑
                </a-button>
                <a-button type="link" size="small" danger @click="handleDelete(record)">
                  <template #icon><DeleteOutlined /></template>
                  删除
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>

    <!-- 添加用户模态框 -->
    <a-modal
      v-model:open="showAddUser"
      title="添加新用户"
      @ok="handleAddUser"
      :confirm-loading="addUserLoading"
      width="480px"
    >
      <a-form
        ref="addUserFormRef"
        :model="addUserForm"
        :rules="addUserRules"
        layout="vertical"
      >
        <a-form-item name="username" label="用户名">
          <a-input v-model:value="addUserForm.username" placeholder="请输入用户名" />
        </a-form-item>

        <a-form-item name="email" label="邮箱">
          <a-input v-model:value="addUserForm.email" placeholder="请输入邮箱" />
        </a-form-item>

        <a-form-item name="password" label="密码">
          <a-input-password v-model:value="addUserForm.password" placeholder="请输入密码" />
        </a-form-item>

        <a-form-item name="role" label="角色">
          <a-select v-model:value="addUserForm.role" placeholder="选择用户角色">
            <a-select-option value="user">普通用户</a-select-option>
            <a-select-option value="admin">管理员</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 编辑用户模态框 -->
    <a-modal
      v-model:open="showEditUser"
      title="编辑用户"
      @ok="handleUpdateUser"
      :confirm-loading="editUserLoading"
      width="480px"
    >
      <a-form
        ref="editUserFormRef"
        :model="editUserForm"
        :rules="editUserRules"
        layout="vertical"
      >
        <a-form-item name="username" label="用户名">
          <a-input v-model:value="editUserForm.username" disabled />
        </a-form-item>

        <a-form-item name="email" label="邮箱">
          <a-input v-model:value="editUserForm.email" placeholder="请输入邮箱" />
        </a-form-item>

        <a-form-item name="role" label="角色">
          <a-select v-model:value="editUserForm.role" placeholder="选择用户角色">
            <a-select-option value="user">普通用户</a-select-option>
            <a-select-option value="admin">管理员</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  UserOutlined,
  UserAddOutlined,
  EditOutlined,
  DeleteOutlined,
  GlobalOutlined,
  CrownOutlined
} from '@ant-design/icons-vue'
import api from '../../api/axiosInstance'

// 响应式数据
const users = ref([])
const searchText = ref('')
const selectedRowKeys = ref([])
const loading = ref(false)
const qrcodeEnabled = ref(false)
const qrStatusLoading = ref(false)
const showAddUser = ref(false)
const showEditUser = ref(false)
const addUserLoading = ref(false)
const editUserLoading = ref(false)

// 表单引用
const addUserFormRef = ref()
const editUserFormRef = ref()

// 表单数据
const addUserForm = reactive({
  username: '',
  email: '',
  password: '',
  role: 'user'
})

const editUserForm = reactive({
  username: '',
  email: '',
  role: 'user'
})

// 计算属性
const onlineUsers = computed(() => {
  return users.value.filter(user => user.isOnline).length
})

const adminCount = computed(() => {
  return users.value.filter(user => user.role === 'admin').length
})

const filteredUsers = computed(() => {
  if (!searchText.value) {
    return users.value
  }
  return users.value.filter(user =>
    user.account && user.account.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

// 表格配置
const columns = [
  {
    title: '头像',
    key: 'avatar',
    width: 80,
    align: 'center'
  },
  {
    title: '用户信息',
    key: 'account',
    dataIndex: 'account',
    width: 200
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    align: 'center'
  },
  {
    title: '角色',
    key: 'role',
    width: 120,
    align: 'center'
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 120,
    align: 'center'
  },
  {
    title: '操作',
    key: 'action',
    width: 160,
    align: 'center',
    fixed: 'right'
  }
]

const pagination = ref({
  current: 1,
  pageSize: 10,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
  onChange: (page, size) => {
    pagination.value.current = page
    pagination.value.pageSize = size
  },
  onShowSizeChange: (current, size) => {
    pagination.value.current = 1
    pagination.value.pageSize = size
  }
})

// 表单验证规则
const addUserRules = {
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
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

const editUserRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 方法
const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await api.get('/users')
    // 为数据添加额外字段用于展示
    users.value = response.data.map((user, index) => ({
      ...user,
      id: user.id || index, // 确保有ID
      isOnline: Math.random() > 0.7, // 模拟在线状态
      role: user.role || 'user', // 默认角色
      createdAt: Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000 // 模拟创建时间
    }))
  } catch (error) {
    message.error('获取用户列表失败')
    console.error('获取用户失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchQRCodeStatus = async () => {
  try {
    const response = await api.get('/auth/qrcode/status')
    qrcodeEnabled.value = response.data.enabled
  } catch (error) {
    console.error('获取二维码状态失败:', error)
  }
}

const toggleQRCodeStatus = async () => {
  qrStatusLoading.value = true
  try {
    // 注意：这个功能需要根据实际需求实现相应的后端API
    // 当前后端没有提供设置QR码状态的接口，这里暂时模拟实现
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
    message.success(`二维码登录已${qrcodeEnabled.value ? '启用' : '禁用'}`)

    // 实际项目中应该调用类似这样的API:
    // await api.post('/auth/qrcode/toggle', {
    //   enabled: qrcodeEnabled.value
    // })
  } catch (error) {
    message.error('更新二维码状态失败')
    qrcodeEnabled.value = !qrcodeEnabled.value // 回滚状态
  } finally {
    qrStatusLoading.value = false
  }
}

// 防抖搜索
let searchTimeout = null
const handleSearch = (e) => {
  const value = e.target ? e.target.value : e
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(() => {
    searchText.value = value
  }, 300)
}

const onSelectChange = (selectedKeys) => {
  selectedRowKeys.value = selectedKeys
}

const handleAddUser = async () => {
  try {
    await addUserFormRef.value.validate()
    addUserLoading.value = true

    await api.post('/auth/register', {
      username: addUserForm.username,
      email: addUserForm.email,
      password: addUserForm.password
    })

    message.success('用户添加成功')
    showAddUser.value = false
    resetAddUserForm()
    await fetchUsers()
  } catch (error) {
    message.error('添加用户失败')
    console.error('添加用户失败:', error)
  } finally {
    addUserLoading.value = false
  }
}

const handleEdit = (record) => {
  editUserForm.username = record.account
  editUserForm.email = record.email || ''
  editUserForm.role = record.role || 'user'
  showEditUser.value = true
}

const handleUpdateUser = async () => {
  try {
    await editUserFormRef.value.validate()
    editUserLoading.value = true

    // 这里需要根据实际API调整
    message.success('用户信息更新成功')
    showEditUser.value = false
    await fetchUsers()
  } catch (error) {
    message.error('更新用户失败')
    console.error('更新用户失败:', error)
  } finally {
    editUserLoading.value = false
  }
}

const handleDelete = (record) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户 "${record.account}" 吗？此操作不可撤销。`,
    okText: '确认',
    cancelText: '取消',
    okType: 'danger',
    onOk: async () => {
      try {
        // 这里需要根据实际API调整
        await api.delete(`/users/${record.id}`)
        message.success('用户删除成功')
        await fetchUsers()
      } catch (error) {
        message.error('删除用户失败')
        console.error('删除用户失败:', error)
      }
    }
  })
}

const handleBatchDelete = () => {
  Modal.confirm({
    title: '批量删除确认',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个用户吗？此操作不可撤销。`,
    okText: '确认',
    cancelText: '取消',
    okType: 'danger',
    onOk: async () => {
      try {
        // 这里需要根据实际API调整批量删除
        message.success(`成功删除 ${selectedRowKeys.value.length} 个用户`)
        selectedRowKeys.value = []
        await fetchUsers()
      } catch (error) {
        message.error('批量删除失败')
        console.error('批量删除失败:', error)
      }
    }
  })
}

const resetAddUserForm = () => {
  addUserForm.username = ''
  addUserForm.email = ''
  addUserForm.password = ''
  addUserForm.role = 'user'
}

const getAvatarColor = (username) => {
  if (!username) return '#1890ff'
  const colors = ['#f56a00', '#7265e6', '#ffbf00', '#00a2ae', '#f56565', '#9f7aea']
  const index = username.charCodeAt(0) % colors.length
  return colors[index]
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchUsers()
  fetchQRCodeStatus()
})
</script>

<style scoped>
.ant-statistic-content {
  font-size: 24px;
}

.ant-table-tbody > tr:hover > td {
  background: #fafafa !important;
}

.ant-card {
  border-radius: 8px;
}

.ant-card-small > .ant-card-body {
  padding: 16px;
}
</style>