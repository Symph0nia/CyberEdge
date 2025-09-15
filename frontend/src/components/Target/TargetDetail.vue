<template>
  <div class="bg-gray-900 text-white flex flex-col min-h-screen">
    <HeaderPage />

    <div class="container mx-auto px-6 py-8 flex-1 mt-16">
      <!-- 顶部功能栏 -->
      <div class="flex justify-between items-center mb-6">
        <div class="flex items-center">
          <button
            @click="goBack"
            class="mr-4 p-2 rounded-lg bg-gray-800/50 hover:bg-gray-700/50 transition-all duration-200"
            title="返回"
          >
            <i class="ri-arrow-left-line text-lg"></i>
          </button>
          <h1 class="text-xl font-medium text-gray-100">目标详情</h1>
        </div>

        <div class="flex space-x-3">
          <button
            @click="refreshDetails"
            class="action-button bg-gray-700/50 hover:bg-gray-600/50"
            :class="{ 'opacity-50 cursor-wait': isLoading }"
            :disabled="isLoading"
          >
            <i
              class="ri-refresh-line mr-1.5"
              :class="{ 'animate-spin': isLoading }"
            ></i>
            刷新
          </button>
          <button class="action-button bg-blue-600/50 hover:bg-blue-500/50">
            <i class="ri-scan-line mr-1.5"></i>
            扫描
          </button>
        </div>
      </div>

      <!-- 目标基本信息卡片 -->
      <div class="info-card mb-6">
        <div
          class="flex flex-col md:flex-row md:items-start md:justify-between gap-4"
        >
          <!-- 左侧标题和描述 -->
          <div class="flex-1">
            <div class="flex items-center gap-3 mb-3">
              <div
                class="flex items-center justify-center w-10 h-10 rounded-xl bg-gray-700/50"
              >
                <i class="ri-focus-3-line text-xl text-blue-400"></i>
              </div>
              <div>
                <h2 class="text-xl font-semibold text-gray-100">
                  {{ details?.name || "加载中..." }}
                </h2>
                <div class="flex items-center mt-1">
                  <span
                    class="status-badge"
                    :class="getStatusStyle(details?.status)"
                  >
                    {{ getStatusText(details?.status) }}
                  </span>
                  <span class="text-sm text-gray-500 ml-3">
                    <i class="ri-time-line mr-1"></i>
                    {{ formatTimeAgo(details?.createdAt) }}创建
                  </span>
                </div>
              </div>
            </div>

            <p class="text-gray-400 mt-2 mb-4">
              {{ details?.description || "暂无描述" }}
            </p>
          </div>

          <!-- 右侧目标详情 -->
          <div class="bg-gray-800/50 rounded-xl p-4 md:min-w-[300px]">
            <h3
              class="text-sm font-medium text-gray-400 mb-3 flex items-center"
            >
              <i class="ri-information-line mr-1.5"></i> 目标信息
            </h3>
            <div class="space-y-3">
              <div class="flex items-center">
                <span class="text-gray-500 text-sm w-20">目标类型:</span>
                <span class="text-gray-200">{{
                  details?.type === "domain" ? "域名" : "IP地址"
                }}</span>
              </div>
              <div class="flex items-center">
                <span class="text-gray-500 text-sm w-20">目标地址:</span>
                <span
                  class="text-gray-200 flex-1 font-mono overflow-hidden text-ellipsis"
                >
                  {{ details?.target || "-" }}
                </span>
                <button
                  @click="copyToClipboard(details?.target)"
                  class="text-gray-400 hover:text-blue-400 ml-2"
                  title="复制"
                >
                  <i class="ri-clipboard-line"></i>
                </button>
              </div>
              <div class="flex items-center">
                <span class="text-gray-500 text-sm w-20">更新时间:</span>
                <span class="text-gray-200">{{
                  formatDate(details?.updatedAt)
                }}</span>
              </div>
              <div class="flex items-center">
                <span class="text-gray-500 text-sm w-20">目标ID:</span>
                <span class="text-gray-400 font-mono text-sm">{{
                  details?.id || "-"
                }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 数据统计卡片网格 -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div
          v-for="stat in statistics"
          :key="stat.title"
          class="stat-card"
          :class="`border-${stat.color}-900/30 hover:border-${stat.color}-800/30`"
        >
          <div class="flex items-center">
            <div
              class="flex items-center justify-center w-12 h-12 rounded-xl mr-4"
              :class="`bg-${stat.color}-900/30 text-${stat.color}-400`"
            >
              <i :class="`${stat.icon} text-xl`"></i>
            </div>
            <div>
              <h3 class="text-2xl font-bold" :class="`text-${stat.color}-400`">
                {{ stat.value }}
              </h3>
              <p class="text-sm text-gray-400">{{ stat.title }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 图表区域 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <!-- 端口分布图表 -->
        <div class="chart-card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="chart-title">
              <i class="ri-door-line mr-1.5"></i>
              端口分布
            </h3>
            <div
              class="bg-gray-800/70 text-xs text-gray-400 px-2.5 py-1 rounded-lg"
            >
              显示前 {{ portChartData.labels.length || 0 }} 个
            </div>
          </div>

          <div class="chart-container">
            <Bar
              v-if="portChartData.datasets[0].data.length > 0"
              :data="portChartData"
              :options="portChartOptions"
            />
            <div v-else class="empty-chart">
              <i
                class="ri-bar-chart-grouped-line text-4xl text-gray-700 mb-3"
              ></i>
              <p>暂无端口数据</p>
            </div>
          </div>
        </div>

        <!-- HTTP状态码分布图表 -->
        <div class="chart-card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="chart-title">
              <i class="ri-http-fill mr-1.5"></i>
              HTTP状态码分布
            </h3>
          </div>

          <div class="chart-container">
            <Pie
              v-if="httpStatusChartData.datasets[0].data.length > 0"
              :data="httpStatusChartData"
              :options="httpStatusChartOptions"
            />
            <div v-else class="empty-chart">
              <i class="ri-pie-chart-line text-4xl text-gray-700 mb-3"></i>
              <p>暂无HTTP状态码数据</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 额外功能卡片区 -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <div
          class="info-card flex flex-col items-center justify-center py-6 hover:bg-gray-800/50 transition-colors duration-200 group cursor-pointer"
        >
          <div
            class="w-14 h-14 flex items-center justify-center rounded-full bg-purple-900/30 text-purple-400 mb-3 group-hover:scale-110 transition-transform duration-200"
          >
            <i class="ri-folder-chart-line text-2xl"></i>
          </div>
          <h3 class="text-lg font-medium text-gray-200 mb-1">子域名报告</h3>
          <p class="text-sm text-gray-500">查看子域名扫描发现的详细结果</p>
        </div>

        <div
          class="info-card flex flex-col items-center justify-center py-6 hover:bg-gray-800/50 transition-colors duration-200 group cursor-pointer"
        >
          <div
            class="w-14 h-14 flex items-center justify-center rounded-full bg-green-900/30 text-green-400 mb-3 group-hover:scale-110 transition-transform duration-200"
          >
            <i class="ri-database-2-line text-2xl"></i>
          </div>
          <h3 class="text-lg font-medium text-gray-200 mb-1">端口服务详情</h3>
          <p class="text-sm text-gray-500">查看开放的端口和服务详细信息</p>
        </div>

        <div
          class="info-card flex flex-col items-center justify-center py-6 hover:bg-gray-800/50 transition-colors duration-200 group cursor-pointer"
        >
          <div
            class="w-14 h-14 flex items-center justify-center rounded-full bg-red-900/30 text-red-400 mb-3 group-hover:scale-110 transition-transform duration-200"
          >
            <i class="ri-bug-line text-2xl"></i>
          </div>
          <h3 class="text-lg font-medium text-gray-200 mb-1">漏洞分析</h3>
          <p class="text-sm text-gray-500">查看已发现的安全漏洞和风险</p>
        </div>
      </div>
    </div>

    <FooterPage />

    <!-- 复制成功提示 -->
    <div
      v-if="showCopyNotification"
      class="fixed bottom-8 right-8 bg-gray-800 text-gray-200 py-2 px-4 rounded-lg shadow-lg flex items-center transition-opacity duration-300"
      :class="{
        'opacity-100': showCopyNotification,
        'opacity-0': !showCopyNotification,
      }"
    >
      <i class="ri-check-line mr-2 text-green-400"></i>
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
