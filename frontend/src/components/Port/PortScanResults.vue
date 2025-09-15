<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 端口扫描结果面板 -->
      <div
        class="border"
      >
        <!-- 标题区域 -->
        <div
          class="md:"
        >
          <div >
            <div
              
            >
              <i class="ri-scan-2-line"></i>
            </div>
            <div>
              <h2 >
                端口扫描结果
              </h2>
              <p >
                查看和管理端口扫描任务的结果
              </p>
            </div>
            <div
              
              :class="filteredResults.length > 0 ? ' ' : ''"
            >
              <i class="ri-database-2-line .5"></i>
              {{ filteredResults.length }} 个结果
            </div>
          </div>

          <div >
            <button
              @click="handleRefreshTasks"
              class=".5 hover: duration-200"
            >
              <i class="ri-refresh-line"></i>
              刷新列表
            </button>

            <button
              @click="router.push('/target-management')"
              class=".5 hover: duration-200 border"
            >
              <i class="ri-add-line"></i>
              新建扫描
            </button>
          </div>
        </div>

        <!-- 搜索和过滤栏 -->
        <div
          class="border"
        >
          <div class="md:">
            <!-- 搜索框 -->
            <div >
              <i class="ri-search-line"></i>
              <input
                v-model.trim="searchQuery"
                type="text"
                placeholder="搜索目标地址..."
                class=".5 border focus: duration-200"
              />
            </div>

            <div >
              <!-- 状态过滤下拉框 -->
              <div class="min-">
                <i
                  class="ri-filter-line"
                ></i>
                <select
                  v-model="statusFilter"
                  class=".5 border focus: duration-200 appearance-none"
                >
                  <option value="">所有状态</option>
                  <option value="true">已读</option>
                  <option value="false">未读</option>
                </select>
                <i
                  class="ri-arrow-down-s-line"
                ></i>
              </div>

              <!-- 查询按钮 -->
              <button
                @click="handleSearch"
                class=".5 hover: duration-200 min-"
              >
                <i class="ri-search-line"></i>
                查询
              </button>
            </div>
          </div>

          <!-- 活跃过滤器提示 -->
          <div
            v-if="hasActiveFilters"
            
          >
            <i class="ri-filter-3-line"></i>
            <span>已过滤: </span>
            <div
              v-if="searchQuery"
              class=".5"
            >
              搜索 "{{ searchQuery }}"
              <button
                @click="searchQuery = ''"
                class="hover:"
              >
                <i class="ri-close-line"></i>
              </button>
            </div>
            <div
              v-if="statusFilter"
              class=".5"
            >
              {{ statusFilter === "true" ? "已读" : "未读" }}
              <button
                @click="statusFilter = ''"
                class="hover:"
              >
                <i class="ri-close-line"></i>
              </button>
            </div>
            <button
              @click="clearFilters"
              class="hover:"
            >
              清除全部
            </button>
          </div>
        </div>

        <!-- 端口扫描结果表格 -->
        <div
          class="overflow- border"
          :class="{ '': isLoading }"
        >
          <PortScanTable
            :portScanResults="filteredResults"
            :loading="isLoading"
            @view-details="handleViewDetails"
            @delete-result="handleDeleteResult"
            @delete-selected="handleDeleteSelected"
            @toggle-read-status="handleToggleReadStatus"
            @mark-selected-read="handleMarkSelectedRead"
          />
        </div>

        <!-- 空状态展示 -->
        <div
          v-if="!isLoading && filteredResults.length === 0"
          
        >
          <div
            
          >
            <i class="ri-scan-2-line"></i>
          </div>
          <h3 >无扫描结果</h3>
          <p class="max-">
            {{
              hasActiveFilters
                ? "没有符合当前过滤条件的端口扫描结果，请尝试修改过滤条件或清除筛选"
                : "当前还没有任何端口扫描结果。创建一个新的扫描任务来检测开放端口。"
            }}
          </p>
          <button
            v-if="hasActiveFilters"
            @click="clearFilters"
            class=".5 hover: duration-200"
          >
            <i class="ri-filter-off-line"></i>
            清除筛选条件
          </button>
          <button
            v-else
            @click="router.push('/target-management')"
            class=".5 hover: duration-200 border"
          >
            <i class="ri-add-line"></i>
            新建扫描任务
          </button>
        </div>

        <!-- 错误提示 -->
        <div
          v-if="errorMessage"
          class="border"
        >
          <i class="ri-error-warning-line"></i>
          <p >{{ errorMessage }}</p>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 通知组件 -->
    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />

    <!-- 确认对话框 -->
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
import { useRouter } from "vue-router";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PortScanTable from "./PortScanTable.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import ConfirmDialog from "../Utils/ConfirmDialog.vue";
import api from "../../api/axiosInstance";
import { useNotification } from "../../composables/useNotification";
import { useConfirmDialog } from "../../composables/useConfirmDialog";

export default {
  name: "PortScanResults",
  components: {
    HeaderPage,
    FooterPage,
    PortScanTable,
    PopupNotification,
    ConfirmDialog,
  },
  setup() {
    const router = useRouter();
    const portScanResults = ref([]);
    const errorMessage = ref("");
    const searchQuery = ref("");
    const statusFilter = ref("");
    const isLoading = ref(false);

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

    // 检查是否有活跃的过滤条件
    const hasActiveFilters = computed(() => {
      return searchQuery.value.trim() !== "" || statusFilter.value !== "";
    });

    // 过滤后的结果
    const filteredResults = computed(() => {
      if (!portScanResults.value) return [];

      let filtered = [...portScanResults.value];

      // 搜索过滤
      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase();
        filtered = filtered.filter((result) =>
          result.target.toLowerCase().includes(query)
        );
      }

      // 状态过滤
      if (statusFilter.value !== "") {
        filtered = filtered.filter(
          (result) => String(result.is_read) === statusFilter.value
        );
      }

      // 按时间戳排序（最新的在前）
      filtered.sort((a, b) => new Date(b.Timestamp) - new Date(a.Timestamp));

      return filtered;
    });

    // 获取扫描结果
    const fetchPortScanResults = async () => {
      try {
        isLoading.value = true;
        errorMessage.value = "";

        const response = await api.get("/results/type/Port");
        portScanResults.value = response.data;

        showSuccess("已刷新扫描结果");
      } catch (error) {
        errorMessage.value = "获取扫描结果失败，请重试或联系管理员";
        showError("获取扫描结果失败");
      } finally {
        isLoading.value = false;
      }
    };

    // 处理搜索
    const handleSearch = () => {
      // 这里只是触发一次动画效果，实际过滤是通过计算属性实时完成的
      isLoading.value = true;
      setTimeout(() => {
        isLoading.value = false;
      }, 300);
    };

    // 清除所有过滤条件
    const clearFilters = () => {
      searchQuery.value = "";
      statusFilter.value = "";
      handleSearch();
    };

    // 查看详情
    const handleViewDetails = (id) => {
      router.push({ name: "PortScanDetail", params: { id } });
    };

    // 删除单个结果
    const handleDeleteResult = async (id) => {
      try {
        const confirmed = await confirm({
          title: "确认删除",
          message: "是否确认删除此扫描结果？此操作不可撤销。",
          type: "danger",
        });

        if (!confirmed) return;

        isLoading.value = true;
        await api.delete(`/results/${id}`);
        await fetchPortScanResults();
        showSuccess("已删除扫描结果");
      } catch (error) {
        showError("删除扫描结果失败");
        isLoading.value = false;
      }
    };

    // 切换已读状态
    const handleToggleReadStatus = async (id, is_read) => {
      try {
        isLoading.value = true;
        await api.put(`/results/${id}/read`, { is_read });

        // 乐观更新UI，避免完全刷新
        const index = portScanResults.value.findIndex(
          (result) => result.id === id
        );
        if (index !== -1) {
          portScanResults.value[index].is_read = is_read;
        }

        showSuccess("已更新状态");
        isLoading.value = false;
      } catch (error) {
        showError("更新状态失败");
        await fetchPortScanResults(); // 出错时刷新以保持一致性
      }
    };

    // 批量标记已读
    const handleMarkSelectedRead = async (selectedIds) => {
      if (selectedIds.length === 0) {
        showError("请选择要标记的结果");
        return;
      }

      try {
        const confirmed = await confirm({
          title: "批量标记",
          message: `是否将选中的 ${selectedIds.length} 项标记为已读？`,
          type: "info",
        });

        if (!confirmed) return;

        isLoading.value = true;
        await Promise.all(
          selectedIds.map((id) =>
            api.put(`/results/${id}/read`, { is_read: true })
          )
        );
        await fetchPortScanResults();
        showSuccess(`已将 ${selectedIds.length} 个结果标记为已读`);
      } catch (error) {
        showError("批量标记失败");
      } finally {
        isLoading.value = false;
      }
    };

    // 批量删除
    const handleDeleteSelected = async (selectedIds) => {
      if (selectedIds.length === 0) {
        showError("请选择要删除的结果");
        return;
      }

      try {
        const confirmed = await confirm({
          title: "批量删除",
          message: `是否确认删除选中的 ${selectedIds.length} 个结果？此操作不可撤销。`,
          type: "danger",
        });

        if (!confirmed) return;

        isLoading.value = true;
        await Promise.all(
          selectedIds.map((id) => api.delete(`/results/${id}`))
        );
        await fetchPortScanResults();
        showSuccess(`已删除 ${selectedIds.length} 个结果`);
      } catch (error) {
        showError("批量删除失败");
        isLoading.value = false;
      }
    };

    onMounted(fetchPortScanResults);

    return {
      router,
      portScanResults,
      errorMessage,
      // 通知相关
      showNotification,
      notificationMessage,
      notificationType,
      // 确认对话框相关
      showDialog,
      dialogTitle,
      dialogMessage,
      dialogType,
      searchQuery,
      statusFilter,
      filteredResults,
      isLoading,
      hasActiveFilters,

      handleConfirm,
      handleCancel,
      // 方法
      handleRefreshTasks: fetchPortScanResults,
      handleViewDetails,
      handleDeleteResult,
      handleDeleteSelected,
      handleToggleReadStatus,
      handleMarkSelectedRead,
      handleSearch,
      clearFilters,
    };
  },
};
</script>

<style scoped>
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 平滑过渡效果 */
.opacity-50 {
  transition: opacity 0.3s ease;
}

/* 输入框样式增强 */
input:focus,
select:focus {
  box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.1);
}
</style>
