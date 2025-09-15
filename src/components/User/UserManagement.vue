<template>
  <div class="bg-gray-900 text-white flex flex-col min-h-screen">
    <HeaderPage />

    <div class="container mx-auto px-6 py-8 flex-1 mt-16">
      <!-- 管理概览卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <!-- 用户统计 -->
        <div
          class="bg-gray-800/40 backdrop-blur-xl p-6 rounded-2xl shadow-lg border border-gray-700/30 flex items-center"
        >
          <div
            class="h-12 w-12 rounded-xl bg-blue-500/20 flex items-center justify-center mr-4"
          >
            <i class="ri-user-line text-xl text-blue-400"></i>
          </div>
          <div>
            <h3 class="text-gray-400 text-sm font-medium">总用户数</h3>
            <p class="text-2xl font-semibold text-white">{{ users.length }}</p>
          </div>
        </div>

        <!-- 二维码接口控制 -->
        <div
          class="bg-gray-800/40 backdrop-blur-xl p-6 rounded-2xl shadow-lg border border-gray-700/30 flex items-center justify-between"
        >
          <div class="flex items-center">
            <div
              class="h-12 w-12 rounded-xl bg-purple-500/20 flex items-center justify-center mr-4"
            >
              <i class="ri-qr-code-line text-xl text-purple-400"></i>
            </div>
            <div>
              <h3 class="text-gray-400 text-sm font-medium">二维码登录</h3>
              <p class="text-lg font-medium text-white">
                {{ qrcodeEnabled ? "已启用" : "已禁用" }}
              </p>
            </div>
          </div>
          <button
            @click="toggleQRCodeStatus"
            class="relative w-14 h-7 rounded-full transition-colors duration-300 focus:outline-none"
            :class="qrcodeEnabled ? 'bg-purple-500/70' : 'bg-gray-600/50'"
          >
            <span
              class="absolute left-1 top-1 w-5 h-5 rounded-full bg-white shadow-md transition-transform duration-300"
              :class="qrcodeEnabled ? 'transform translate-x-7' : ''"
            ></span>
          </button>
        </div>
      </div>

      <!-- 用户列表卡片 -->
      <div
        class="bg-gray-800/40 backdrop-blur-xl p-8 rounded-2xl shadow-xl border border-gray-700/30"
      >
        <!-- 列表标题和操作栏 -->
        <div class="flex flex-wrap items-center justify-between gap-4 mb-6">
          <h2 class="text-xl font-medium tracking-wide flex items-center">
            <i class="ri-user-settings-line mr-2 text-blue-400"></i>
            用户管理
          </h2>

          <div class="flex items-center gap-3">
            <!-- 搜索框 -->
            <div class="relative">
              <input
                type="text"
                v-model="searchQuery"
                placeholder="搜索用户..."
                class="bg-gray-700/50 border border-gray-600/50 rounded-xl py-2 pl-10 pr-4 w-64 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500/40 transition-all duration-200"
              />
              <i
                class="ri-search-line absolute left-3 top-2.5 text-gray-400"
              ></i>
            </div>

            <!-- 批量删除按钮 -->
            <button
              v-if="selectedUsers.length > 0"
              @click="handleBatchDelete"
              class="px-4 py-2 rounded-xl text-sm font-medium bg-red-500/50 hover:bg-red-600/60 text-red-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-red-500/50 flex items-center"
            >
              <i class="ri-delete-bin-line mr-2"></i>
              批量删除 ({{ selectedUsers.length }})
            </button>
          </div>
        </div>

        <!-- 用户数据表格 -->
        <div class="overflow-x-auto rounded-xl border border-gray-700/30">
          <table class="w-full">
            <thead>
              <tr class="bg-gray-800/70">
                <th
                  class="text-left py-3 px-4 text-sm font-medium text-gray-400"
                >
                  <input
                    type="checkbox"
                    :checked="isAllSelected"
                    @change="toggleSelectAll"
                    class="rounded border-gray-600 text-blue-500 focus:ring-blue-500/50 bg-gray-700/50"
                  />
                </th>
                <th
                  class="text-left py-3 px-4 text-sm font-medium text-gray-400"
                >
                  用户名
                </th>
                <th
                  class="text-left py-3 px-4 text-sm font-medium text-gray-400"
                >
                  登录次数
                </th>
                <th
                  class="text-left py-3 px-4 text-sm font-medium text-gray-400"
                >
                  操作
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in filteredUsers"
                :key="user.account"
                class="border-t border-gray-700/30 hover:bg-gray-700/30 transition-colors duration-200"
              >
                <td class="py-3 px-4">
                  <input
                    type="checkbox"
                    v-model="selectedUsers"
                    :value="user.account"
                    class="rounded border-gray-600 text-blue-500 focus:ring-blue-500/50 bg-gray-700/50"
                  />
                </td>
                <td class="py-3 px-4 text-sm text-gray-200 flex items-center">
                  <div
                    class="w-8 h-8 bg-gray-700 rounded-full mr-3 flex items-center justify-center text-blue-400"
                  >
                    {{ user.account.charAt(0).toUpperCase() }}
                  </div>
                  {{ user.account }}
                </td>
                <td class="py-3 px-4 text-sm">
                  <span
                    class="px-2 py-1 rounded-md bg-blue-500/20 text-blue-300 flex items-center w-fit"
                  >
                    <i class="ri-login-circle-line mr-2"></i>
                    {{ user.loginCount }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <button
                    @click="handleDelete(user.account)"
                    class="p-2 rounded-lg text-sm font-medium bg-red-500/30 hover:bg-red-500/50 text-red-300 transition-all duration-200 focus:outline-none"
                    title="删除用户"
                  >
                    <i class="ri-delete-bin-line"></i>
                  </button>
                </td>
              </tr>
              <!-- 空状态显示 -->
              <tr v-if="filteredUsers.length === 0">
                <td colspan="4" class="py-12 text-center text-gray-400">
                  <div class="flex flex-col items-center justify-center">
                    <i
                      class="ri-user-search-line text-4xl mb-3 text-gray-600"
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
          class="mt-4 text-sm text-gray-400 flex justify-between items-center"
        >
          <p>共 {{ filteredUsers.length }} 个用户</p>
          <div class="flex items-center space-x-2">
            <span>每页显示:</span>
            <select
              v-model="perPage"
              class="bg-gray-700/50 border border-gray-600/30 rounded-md px-2 py-1 text-white text-sm"
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
