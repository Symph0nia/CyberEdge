<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 管理概览卡片 -->
      <div class="md:">
        <!-- 用户统计 -->
        <div
          class="border"
        >
          <div
            
          >
            <i class="ri-user-line"></i>
          </div>
          <div>
            <h3 >总用户数</h3>
            <p >{{ users.length }}</p>
          </div>
        </div>

        <!-- 二维码接口控制 -->
        <div
          class="border"
        >
          <div >
            <div
              
            >
              <i class="ri-qr-code-line"></i>
            </div>
            <div>
              <h3 >二维码登录</h3>
              <p >
                {{ qrcodeEnabled ? "已启用" : "已禁用" }}
              </p>
            </div>
          </div>
          <button
            @click="toggleQRCodeStatus"
            class="duration-300"
            :class="qrcodeEnabled ? '' : ''"
          >
            <span
              class="duration-300"
              :class="qrcodeEnabled ? ' ' : ''"
            ></span>
          </button>
        </div>
      </div>

      <!-- 用户列表卡片 -->
      <div
        class="border"
      >
        <!-- 列表标题和操作栏 -->
        <div >
          <h2 >
            <i class="ri-user-settings-line"></i>
            用户管理
          </h2>

          <div >
            <!-- 搜索框 -->
            <div >
              <input
                type="text"
                v-model="searchQuery"
                placeholder="搜索用户..."
                class="border focus: duration-200"
              />
              <i
                class="ri-search-line .5"
              ></i>
            </div>

            <!-- 批量删除按钮 -->
            <button
              v-if="selectedUsers.length > 0"
              @click="handleBatchDelete"
              class="hover: duration-200"
            >
              <i class="ri-delete-bin-line"></i>
              批量删除 ({{ selectedUsers.length }})
            </button>
          </div>
        </div>

        <!-- 用户数据表格 -->
        <div class="border">
          <table >
            <thead>
              <tr >
                <th
                  
                >
                  <input
                    type="checkbox"
                    :checked="isAllSelected"
                    @change="toggleSelectAll"
                    
                  />
                </th>
                <th
                  
                >
                  用户名
                </th>
                <th
                  
                >
                  登录次数
                </th>
                <th
                  
                >
                  操作
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in filteredUsers"
                :key="user.account"
                class="hover: duration-200"
              >
                <td >
                  <input
                    type="checkbox"
                    v-model="selectedUsers"
                    :value="user.account"
                    
                  />
                </td>
                <td >
                  <div
                    
                  >
                    {{ user.account.charAt(0).toUpperCase() }}
                  </div>
                  {{ user.account }}
                </td>
                <td >
                  <span
                    
                  >
                    <i class="ri-login-circle-line"></i>
                    {{ user.loginCount }}
                  </span>
                </td>
                <td >
                  <button
                    @click="handleDelete(user.account)"
                    class="hover: duration-200"
                    title="删除用户"
                  >
                    <i class="ri-delete-bin-line"></i>
                  </button>
                </td>
              </tr>
              <!-- 空状态显示 -->
              <tr v-if="filteredUsers.length === 0">
                <td colspan="4" >
                  <div >
                    <i
                      class="ri-user-search-line"
                    ></i>
                    <p v-if="searchQuery">
                      未找到匹配 "{{ searchQuery }}" 的用户
                    </p>
                    <p v-else>暂无用户数据</p>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 表格底部信息 -->
        <div
          
        >
          <p>共 {{ filteredUsers.length }} 个用户</p>
          <div >
            <span>每页显示:</span>
            <select
              v-model="perPage"
              class="border"
            >
              <option>10</option>
              <option>20</option>
              <option>50</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 通知和确认对话框组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />

    <ConfirmDialog
      :show="showDialog"
      :title="dialogTitle"
      :message="dialogMessage"
      :type="dialogType"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </div>
</template>

<script>
import { ref, computed, onMounted } from "vue";
import api from "../../api/axiosInstance";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import ConfirmDialog from "../Utils/ConfirmDialog.vue";
import { useNotification } from "../../composables/useNotification";
import { useConfirmDialog } from "../../composables/useConfirmDialog";

export default {
  name: "UserManagement",
  components: {
    HeaderPage,
    FooterPage,
    PopupNotification,
    ConfirmDialog,
  },
  setup() {
    const users = ref([]);
    const selectedUsers = ref([]);
    const qrcodeEnabled = ref(false);
    const searchQuery = ref("");
    const perPage = ref(10);

    // 使用通知钩子
    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    // 使用确认对话框钩子
    const {
      showDialog,
      dialogTitle,
      dialogMessage,
      dialogType,
      confirm,
      handleConfirm,
      handleCancel,
    } = useConfirmDialog();

    // 过滤用户列表
    const filteredUsers = computed(() => {
      if (!searchQuery.value) return users.value;

      return users.value.filter((user) =>
        user.account.toLowerCase().includes(searchQuery.value.toLowerCase())
      );
    });

    // 获取用户列表
    const fetchUsers = async () => {
      try {
        const response = await api.get("/users");
        users.value = response.data;
      } catch (error) {
        showError("获取用户列表失败");
      }
    };

    // 批量删除用户
    const handleBatchDelete = async () => {
      if (selectedUsers.value.length === 0) return;

      try {
        const confirmed = await confirm({
          title: "确认批量删除",
          message: `是否确认删除选中的 ${selectedUsers.value.length} 个用户？此操作不可撤销。`,
          type: "danger",
        });

        if (confirmed) {
          const response = await api.delete("/users", {
            data: { accounts: selectedUsers.value },
          });

          // 处理删除结果
          const result = response.data.result;
          if (result.success.length > 0) {
            showSuccess(`成功删除 ${result.success.length} 个用户`);
          }

          if (Object.keys(result.failed).length > 0) {
            const failedCount = Object.keys(result.failed).length;
            showError(`${failedCount} 个用户删除失败`);
          }

          // 清空选择并刷新列表
          selectedUsers.value = [];
          await fetchUsers();
        }
      } catch (error) {
        showError("批量删除用户失败");
      }
    };

    // 单个删除用户
    const handleDelete = async (account) => {
      try {
        const confirmed = await confirm({
          title: "确认删除",
          message: `是否确认删除用户 ${account}？此操作不可撤销。`,
          type: "danger",
        });

        if (confirmed) {
          const response = await api.delete("/users", {
            data: { accounts: [account] },
          });

          const result = response.data.result;
          if (result.success.includes(account)) {
            showSuccess(`已删除用户 ${account}`);
          } else {
            showError(`删除用户 ${account} 失败：${result.failed[account]}`);
          }

          await fetchUsers();
        }
      } catch (error) {
        showError(`删除用户 ${account} 失败`);
      }
    };

    // 切换二维码接口状态
    const toggleQRCodeStatus = async () => {
      try {
        await api.post("/auth/qrcode/status", {
          enabled: !qrcodeEnabled.value,
        });
        qrcodeEnabled.value = !qrcodeEnabled.value;
        showSuccess(`二维码登录已${qrcodeEnabled.value ? "启用" : "禁用"}`);
      } catch (error) {
        showError("更新二维码接口状态失败");
      }
    };

    // 获取二维码接口状态
    const getQRCodeStatus = async () => {
      try {
        const response = await api.get("/auth/qrcode/status");
        qrcodeEnabled.value = response.data.enabled;
      } catch (error) {
        showError("获取二维码接口状态失败");
      }
    };

    const isAllSelected = computed(() => {
      return (
        filteredUsers.value.length > 0 &&
        selectedUsers.value.length === filteredUsers.value.length
      );
    });

    // 切换全选状态
    const toggleSelectAll = () => {
      if (isAllSelected.value) {
        selectedUsers.value = [];
      } else {
        selectedUsers.value = filteredUsers.value.map((user) => user.account);
      }
    };

    onMounted(() => {
      fetchUsers();
      getQRCodeStatus();
    });

    return {
      users,
      filteredUsers,
      selectedUsers,
      isAllSelected,
      qrcodeEnabled,
      searchQuery,
      perPage,
      showNotification,
      notificationMessage,
      notificationType,
      showDialog,
      dialogTitle,
      dialogMessage,
      dialogType,
      handleConfirm,
      handleCancel,
      handleDelete,
      handleBatchDelete,
      toggleSelectAll,
      toggleQRCodeStatus,
    };
  },
};
</script>

<style scoped>
/* 背景模糊效果 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 表格行交替颜色 */
tr:nth-child(odd) {
  background-color: rgba(31, 41, 55, 0.3);
}

/* 切换按钮动画 */
button:active {
  transform: scale(0.98);
}

/* 表格悬停效果 */
tbody tr:hover td {
  background-color: rgba(55, 65, 81, 0.3);
}
</style>
