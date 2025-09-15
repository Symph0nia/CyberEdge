<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 顶部功能栏 -->
      <div >
        <div >
          <button
            @click="goBack"
            class="hover: duration-200"
            title="返回"
          >
            <i class="ri-arrow-"></i>
          </button>
          <h1 >目标详情</h1>
        </div>

        <div >
          <button
            @click="refreshDetails"
            class="action-button hover:"
            :class="{ ' ': isLoading }"
            :disabled="isLoading"
          >
            <i
              class="ri-refresh-line .5"
              :class="{ '': isLoading }"
            ></i>
            刷新
          </button>
          <button class="action-button hover:">
            <i class="ri-scan-line .5"></i>
            扫描
          </button>
        </div>
      </div>

      <!-- 目标基本信息卡片 -->
      <div class="info-card">
        <div
          class="md:"
        >
          <!-- 左侧标题和描述 -->
          <div >
            <div >
              <div
                
              >
                <i class="ri-focus-3-line"></i>
              </div>
              <div>
                <h2 >
                  {{ details?.name || "加载中..." }}
                </h2>
                <div >
                  <span
                    class="status-badge"
                    :class="getStatusStyle(details?.status)"
                  >
                    {{ getStatusText(details?.status) }}
                  </span>
                  <span >
                    <i class="ri-time-line"></i>
                    {{ formatTimeAgo(details?.createdAt) }}创建
                  </span>
                </div>
              </div>
            </div>

            <p >
              {{ details?.description || "暂无描述" }}
            </p>
          </div>

          <!-- 右侧目标详情 -->
          <div >
            <h3
              
            >
              <i class="ri-information-line .5"></i> 目标信息
            </h3>
            <div >
              <div >
                <span >目标类型:</span>
                <span >{{
                  details?.type === "domain" ? "域名" : "IP地址"
                }}</span>
              </div>
              <div >
                <span >目标地址:</span>
                <span
                  class="overflow-"
                >
                  {{ details?.target || "-" }}
                </span>
                <button
                  @click="copyToClipboard(details?.target)"
                  class="hover:"
                  title="复制"
                >
                  <i class="ri-clipboard-line"></i>
                </button>
              </div>
              <div >
                <span >更新时间:</span>
                <span >{{
                  formatDate(details?.updatedAt)
                }}</span>
              </div>
              <div >
                <span >目标ID:</span>
                <span >{{
                  details?.id || "-"
                }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 数据统计卡片网格 -->
      <div class="sm: lg:">
        <div
          v-for="stat in statistics"
          :key="stat.title"
          class="stat-card"
          :class="`border-${stat.color}-900/30 ${stat.color}-800/30`"
        >
          <div >
            <div
              
              :class="`bg-${stat.color}-900/30 text-${stat.color}-400`"
            >
              <i :class="`${stat.icon} `"></i>
            </div>
            <div>
              <h3  :class="`text-${stat.color}-400`">
                {{ stat.value }}
              </h3>
              <p >{{ stat.title }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 图表区域 -->
      <div class="lg:">
        <!-- 端口分布图表 -->
        <div class="chart-card">
          <div >
            <h3 class="chart-title">
              <i class="ri-door-line .5"></i>
              端口分布
            </h3>
            <div
              class=".5"
            >
              显示前 {{ portChartData.labels.length || 0 }} 个
            </div>
          </div>

          <div class="chart-">
            <Bar
              v-if="portChartData.datasets[0].data.length > 0"
              :data="portChartData"
              :options="portChartOptions"
            />
            <div v-else class="empty-chart">
              <i
                class="ri-bar-chart-grouped-line"
              ></i>
              <p>暂无端口数据</p>
            </div>
          </div>
        </div>

        <!-- HTTP状态码分布图表 -->
        <div class="chart-card">
          <div >
            <h3 class="chart-title">
              <i class="ri-http-fill .5"></i>
              HTTP状态码分布
            </h3>
          </div>

          <div class="chart-">
            <Pie
              v-if="httpStatusChartData.datasets[0].data.length > 0"
              :data="httpStatusChartData"
              :options="httpStatusChartOptions"
            />
            <div v-else class="empty-chart">
              <i class="ri-pie-chart-line"></i>
              <p>暂无HTTP状态码数据</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 额外功能卡片区 -->
      <div class="md:">
        <div
          class="info-card hover: duration-200 group"
        >
          <div
            class="group- duration-200"
          >
            <i class="ri-folder-chart-line"></i>
          </div>
          <h3 >子域名报告</h3>
          <p >查看子域名扫描发现的详细结果</p>
        </div>

        <div
          class="info-card hover: duration-200 group"
        >
          <div
            class="group- duration-200"
          >
            <i class="ri-database-2-line"></i>
          </div>
          <h3 >端口服务详情</h3>
          <p >查看开放的端口和服务详细信息</p>
        </div>

        <div
          class="info-card hover: duration-200 group"
        >
          <div
            class="group- duration-200"
          >
            <i class="ri-bug-line"></i>
          </div>
          <h3 >漏洞分析</h3>
          <p >查看已发现的安全漏洞和风险</p>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 复制成功提示 -->
    <div
      v-if="showCopyNotification"
      class="duration-300"
      :class="{ '': showCopyNotification, '': !showCopyNotification, }"
    >
      <i class="ri-check-line"></i>
      已复制到剪贴板
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import HeaderPage from "@/components/HeaderPage.vue";
import FooterPage from "@/components/FooterPage.vue";
import api from "@/api/axiosInstance";
import { useNotification } from "@/composables/useNotification";
import { Bar, Pie } from "vue-chartjs";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement,
} from "chart.js";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement
);

export default {
  name: "TargetDetail",
  components: {
    HeaderPage,
    FooterPage,
    Bar,
    Pie,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const targetData = ref(null);
    const isLoading = ref(false);
    const showCopyNotification = ref(false);
    const { showError, showSuccess } = useNotification();

    const details = computed(() => targetData.value?.target || {});
    const stats = computed(() => targetData.value?.stats || {});

    const statistics = computed(() => [
      {
        title: "子域名数量",
        value: stats.value.subdomain_count || 0,
        icon: "ri-global-line",
        color: "blue",
      },
      {
        title: "端口数量",
        value: stats.value.port_count || 0,
        icon: "ri-door-open-line",
        color: "green",
      },
      {
        title: "路径数量",
        value: stats.value.path_count || 0,
        icon: "ri-route-line",
        color: "purple",
      },
      {
        title: "漏洞数量",
        value: stats.value.vulnerability_count || 0,
        icon: "ri-bug-line",
        color: "red",
      },
    ]);

    const fetchTargetDetails = async () => {
      try {
        isLoading.value = true;
        const response = await api.get(`/targets/${route.params.id}/details`);
        targetData.value = response.data;
      } catch (error) {
        showError("获取目标详情失败");
        console.error("获取目标详情失败:", error);
      } finally {
        isLoading.value = false;
      }
    };

    const refreshDetails = async () => {
      await fetchTargetDetails();
      showSuccess("数据已刷新");
    };

    const goBack = () => {
      router.go(-1);
    };

    const copyToClipboard = (text) => {
      if (!text) return;

      navigator.clipboard.writeText(text).then(() => {
        showCopyNotification.value = true;
        setTimeout(() => {
          showCopyNotification.value = false;
        }, 2000);
      });
    };

    const formatDate = (date) => {
      if (!date) return "未知";
      return new Date(date).toLocaleString("zh-CN", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      });
    };

    const formatTimeAgo = (date) => {
      if (!date) return "";

      const now = new Date();
      const past = new Date(date);
      const diffInSeconds = Math.floor((now - past) / 1000);

      if (diffInSeconds < 60) return "刚刚";
      if (diffInSeconds < 3600)
        return `${Math.floor(diffInSeconds / 60)}分钟前`;
      if (diffInSeconds < 86400)
        return `${Math.floor(diffInSeconds / 3600)}小时前`;
      if (diffInSeconds < 2592000)
        return `${Math.floor(diffInSeconds / 86400)}天前`;

      return `${past.getFullYear()}年${
        past.getMonth() + 1
      }月${past.getDate()}日`;
    };

    const getStatusText = (status) => {
      const statusMap = {
        active: "活跃",
        archived: "已归档",
      };
      return statusMap[status] || "未知";
    };

    const getStatusStyle = (status) => {
      const styleMap = {
        active: "bg-green-900/30 text-green-400 border-green-700/30",
        archived: "bg-gray-800/30 text-gray-400 border-gray-700/30",
      };
      return (
        styleMap[status] || "bg-gray-800/30 text-gray-400 border-gray-700/30"
      );
    };

    const chartColors = {
      backgrounds: [
        "rgba(59, 130, 246, 0.3)", // 蓝色
        "rgba(34, 197, 94, 0.3)", // 绿色
        "rgba(239, 68, 68, 0.3)", // 红色
        "rgba(168, 85, 247, 0.3)", // 紫色
        "rgba(234, 179, 8, 0.3)", // 黄色
        "rgba(14, 165, 233, 0.3)", // 天蓝色
        "rgba(249, 115, 22, 0.3)", // 橙色
        "rgba(236, 72, 153, 0.3)", // 粉色
        "rgba(45, 212, 191, 0.3)", // 青色
        "rgba(139, 92, 246, 0.3)", // 靛蓝色
      ],
      borders: [
        "rgb(59, 130, 246)", // 蓝色
        "rgb(34, 197, 94)", // 绿色
        "rgb(239, 68, 68)", // 红色
        "rgb(168, 85, 247)", // 紫色
        "rgb(234, 179, 8)", // 黄色
        "rgb(14, 165, 233)", // 天蓝色
        "rgb(249, 115, 22)", // 橙色
        "rgb(236, 72, 153)", // 粉色
        "rgb(45, 212, 191)", // 青色
        "rgb(139, 92, 246)", // 靛蓝色
      ],
    };

    // 端口排名图表数据
    const portChartData = computed(() => ({
      labels: stats.value.top_ports?.map((p) => `端口 ${p.port}`) || [],
      datasets: [
        {
          label: "端口数量",
          data: stats.value.top_ports?.map((p) => p.count) || [],
          backgroundColor: chartColors.backgrounds,
          borderColor: chartColors.borders,
          borderWidth: 1,
        },
      ],
    }));

    // 端口排名图表配置
    const portChartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          backgroundColor: "rgba(31, 41, 55, 0.8)",
          titleColor: "rgb(209, 213, 219)",
          bodyColor: "rgb(229, 231, 235)",
          borderColor: "rgba(75, 85, 99, 0.3)",
          borderWidth: 1,
          padding: 10,
          boxPadding: 5,
          callbacks: {
            label: function (context) {
              return `数量: ${context.raw}`;
            },
          },
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          ticks: {
            color: "rgba(156, 163, 175, 0.8)",
            font: {
              size: 11,
            },
          },
          grid: {
            color: "rgba(75, 85, 99, 0.15)",
            drawBorder: false,
          },
        },
        x: {
          ticks: {
            color: "rgba(156, 163, 175, 0.8)",
            maxRotation: 45,
            minRotation: 45,
            font: {
              size: 11,
            },
          },
          grid: {
            display: false,
            drawBorder: false,
          },
        },
      },
    };

    // HTTP状态码图表数据
    const httpStatusChartData = computed(() => ({
      labels: stats.value.http_status_stats?.map((s) => s.label) || [],
      datasets: [
        {
          data: stats.value.http_status_stats?.map((s) => s.count) || [],
          backgroundColor: [
            "rgba(239, 68, 68, 0.7)", // 500系列
            "rgba(234, 179, 8, 0.7)", // 400系列
            "rgba(59, 130, 246, 0.7)", // 300系列
            "rgba(34, 197, 94, 0.7)", // 200系列
            "rgba(168, 85, 247, 0.7)", // 100系列
            "rgba(107, 114, 128, 0.7)", // 其他
          ],
          borderColor: [
            "rgb(239, 68, 68)",
            "rgb(234, 179, 8)",
            "rgb(59, 130, 246)",
            "rgb(34, 197, 94)",
            "rgb(168, 85, 247)",
            "rgb(107, 114, 128)",
          ],
          borderWidth: 1,
        },
      ],
    }));

    // HTTP状态码图表配置
    const httpStatusChartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: "right",
          labels: {
            color: "rgba(156, 163, 175, 0.8)",
            usePointStyle: true,
            padding: 15,
            font: {
              size: 11,
            },
          },
        },
        tooltip: {
          backgroundColor: "rgba(31, 41, 55, 0.8)",
          titleColor: "rgb(209, 213, 219)",
          bodyColor: "rgb(229, 231, 235)",
          borderColor: "rgba(75, 85, 99, 0.3)",
          borderWidth: 1,
          padding: 10,
          boxPadding: 5,
        },
      },
    };

    onMounted(() => {
      fetchTargetDetails();
    });

    return {
      details,
      statistics,
      isLoading,
      showCopyNotification,
      formatDate,
      formatTimeAgo,
      getStatusText,
      getStatusStyle,
      portChartData,
      portChartOptions,
      httpStatusChartData,
      httpStatusChartOptions,
      refreshDetails,
      goBack,
      copyToClipboard,
    };
  },
};
</script>

<style scoped>
/* 基础样式 */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 图表容器样式 */
.chart-container {
  height: 350px;
  position: relative;
}

/* 统一卡片样式 */
.info-card {
  @apply bg-gray-800/40 backdrop-blur-xl p-6 rounded-xl shadow-md border border-gray-700/30;
}

.stat-card {
  @apply bg-gray-800/40 backdrop-blur-xl p-4 rounded-xl shadow-md border transition-all duration-200;
}

.chart-card {
  @apply bg-gray-800/40 backdrop-blur-xl p-5 rounded-xl shadow-md border border-gray-700/30;
}

.chart-title {
  @apply text-lg font-medium text-gray-200 flex items-center;
}

/* 状态标签样式 */
.status-badge {
  @apply text-xs font-medium px-3 py-1 rounded-full border inline-flex items-center;
}

/* 操作按钮 */
.action-button {
  @apply px-4 py-2 rounded-lg text-sm font-medium flex items-center text-gray-200 transition-all duration-200 focus:outline-none;
}

/* 空图表状态 */
.empty-chart {
  @apply flex flex-col items-center justify-center h-full text-gray-500 text-sm;
}
</style>
