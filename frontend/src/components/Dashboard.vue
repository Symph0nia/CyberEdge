<template>
  <a-layout class="dashboard-layout">
    <!-- 顶部导航栏 -->
    <HeaderPage />

    <!-- 主体内容 -->
    <a-layout-content class="dashboard-content">
      <div class="content-container">
        <!-- 统计卡片 -->
        <a-row :gutter="[16, 16]" class="metrics-row">
          <!-- 卡片 1: 全部任务 -->
          <a-col :xs="24" :sm="12" :lg="6">
            <a-card class="metric-card metric-card-primary" hoverable>
              <a-statistic
                title="全部任务"
                :value="metrics.total_tasks"
                :prefix="'📋'"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>

          <!-- 卡片 2: 进行中的任务 -->
          <a-col :xs="24" :sm="12" :lg="6">
            <a-card class="metric-card metric-card-warning" hoverable>
              <a-statistic
                title="进行中的任务"
                :value="metrics.in_progress_tasks"
                :prefix="'⏳'"
                :value-style="{ color: '#faad14' }"
              />
            </a-card>
          </a-col>

          <!-- 卡片 3: 完成的任务 -->
          <a-col :xs="24" :sm="12" :lg="6">
            <a-card class="metric-card metric-card-success" hoverable>
              <a-statistic
                title="完成的任务"
                :value="metrics.completed_tasks"
                :prefix="'✅'"
                :value-style="{ color: '#52c41a' }"
              />
            </a-card>
          </a-col>

          <!-- 卡片 4: 失败的任务 -->
          <a-col :xs="24" :sm="12" :lg="6">
            <a-card class="metric-card metric-card-error" hoverable>
              <a-statistic
                title="失败的任务"
                :value="metrics.failed_tasks"
                :prefix="'🚫'"
                :value-style="{ color: '#ff4d4f' }"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- 图表区域 -->
        <a-row :gutter="[16, 16]" class="chart-row">
          <a-col :span="24">
            <a-card title="最近7天扫描分析" class="chart-card">
              <div class="chart-container">
                <BarChart :chartData="chartData" :chartOptions="chartOptions" />
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- 活动日志 -->
        <a-row :gutter="[16, 16]" class="activity-row">
          <a-col :span="24">
            <a-card title="最近活动" class="activity-card">
              <a-timeline>
                <a-timeline-item
                  v-for="(log, index) in activityLogs"
                  :key="index"
                  :color="getTimelineColor(log.type)"
                >
                  <template #dot>
                    <i class="ri-time-line timeline-icon"></i>
                  </template>
                  <div class="activity-item">
                    <div class="activity-time">{{ log.time }}</div>
                    <div class="activity-message">{{ log.message }}</div>
                  </div>
                </a-timeline-item>
              </a-timeline>
            </a-card>
          </a-col>
        </a-row>
      </div>
    </a-layout-content>

    <!-- 页脚 -->
    <FooterPage />
  </a-layout>
</template>

<script>
import { ref, onMounted, computed } from "vue";
import HeaderPage from "./HeaderPage.vue";
import FooterPage from "./FooterPage.vue";
import BarChart from "./Utils/BarChart.vue";
import api from "../api/axiosInstance";

export default {
  name: "CyberEdgeDashboard",
  components: {
    HeaderPage,
    FooterPage,
    BarChart,
  },
  setup() {
    const metrics = ref({
      total_tasks: 0,
      in_progress_tasks: 0,
      completed_tasks: 0,
      failed_tasks: 0,
    });

    const activityLogs = ref([]);

    // 时间轴颜色函数
    const getTimelineColor = (type) => {
      const colorMap = {
        success: '#52c41a',
        warning: '#faad14',
        error: '#ff4d4f',
        info: '#1890ff',
        default: '#d9d9d9'
      };
      return colorMap[type] || colorMap.default;
    };

    const fetchMetrics = async () => {
      try {
        const response = await api.get("/scanner/metrics");
        if (Array.isArray(response.data) && response.data.length > 0) {
          const latestMetrics = response.data[0];
          metrics.value = {
            total_tasks: latestMetrics.total_tasks || 0,
            in_progress_tasks: latestMetrics.in_progress_tasks || 0,
            completed_tasks: latestMetrics.completed_tasks || 0,
            failed_tasks: latestMetrics.failed_tasks || 0,
          };
        }
      } catch (error) {
        console.error("获取扫描器指标失败:", error);
      }
    };

    const animateNumbers = () => {
      const elements = document.querySelectorAll(".animate-number-scroll");
      elements.forEach((el) => {
        const target = parseInt(el.getAttribute("data-target"), 10);
        let count = 0;
        const increment = target / 100;
        const updateCount = () => {
          count += increment;
          if (count < target) {
            el.textContent = Math.floor(count);
            requestAnimationFrame(updateCount);
          } else {
            el.textContent = target;
          }
        };
        updateCount();
      });
    };

    const fetchActivityLogs = async () => {
      activityLogs.value = [
        { time: "10:30 AM", message: "检测到漏洞，已记录。🔍" },
        { time: "09:45 AM", message: "成功完成一次全面扫描。✅" },
        { time: "08:20 AM", message: "异常流量监测中。👀" },
      ];
    };

    const weeklyMetrics = ref([]);

    const fetchWeeklyMetrics = async () => {
      try {
        const endDate = new Date();
        endDate.setHours(endDate.getHours() + 8); // 调整为北京时间
        const startDate = new Date(endDate);
        startDate.setDate(startDate.getDate() - 6);

        const response = await api.get("/scanner/metrics", {
          params: {
            start_date: startDate.toISOString().split("T")[0],
            end_date: endDate.toISOString().split("T")[0],
          },
        });
        weeklyMetrics.value = response.data;
      } catch (error) {
        console.error("获取每周指标失败:", error);
      }
    };

    const chartData = computed(() => {
      if (weeklyMetrics.value.length === 0) {
        return {
          labels: [],
          datasets: [],
        };
      }

      const labels = weeklyMetrics.value.map((m) => m.date);
      const datasets = [
        // 可以根据需要添加数据集
      ];

      return { labels, datasets };
    });

    onMounted(() => {
      fetchMetrics().then(() => {
        animateNumbers();
      });
      fetchWeeklyMetrics();
      fetchActivityLogs();
    });

    return {
      metrics,
      activityLogs,
      chartData,
      chartOptions,
      getTimelineColor,
    };
  },
};
</script>

<style scoped>
.dashboard-layout {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  min-height: 100vh;
}

.dashboard-content {
  background: transparent;
  margin-top: 64px; /* Header height */
  padding: 0;
}

.content-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.metrics-row {
  margin-bottom: 24px;
}

.metric-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.metric-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.3);
}

.metric-card-primary {
  border-color: rgba(24, 144, 255, 0.3);
}

.metric-card-warning {
  border-color: rgba(250, 173, 20, 0.3);
}

.metric-card-success {
  border-color: rgba(82, 196, 26, 0.3);
}

.metric-card-error {
  border-color: rgba(255, 77, 79, 0.3);
}

.chart-row {
  margin-bottom: 24px;
}

.chart-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
}

.chart-container {
  height: 320px;
  width: 100%;
}

.activity-row {
  margin-bottom: 24px;
}

.activity-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
}

.timeline-icon {
  font-size: 14px;
  color: #94a3b8;
}

.activity-item {
  padding-left: 12px;
}

.activity-time {
  font-weight: 600;
  color: #e2e8f0;
  font-size: 14px;
  margin-bottom: 4px;
}

.activity-message {
  color: #cbd5e1;
  font-size: 13px;
  line-height: 1.4;
}

/* Ant Design组件样式覆盖 */
.dashboard-layout :deep(.ant-card) {
  background: transparent;
  border: none;
}

.dashboard-layout :deep(.ant-card-head) {
  background: transparent;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.dashboard-layout :deep(.ant-card-head-title) {
  color: #f1f5f9;
  font-weight: 600;
  font-size: 16px;
}

.dashboard-layout :deep(.ant-card-body) {
  background: transparent;
  color: #e2e8f0;
}

.dashboard-layout :deep(.ant-statistic-title) {
  color: #94a3b8;
  font-size: 14px;
  margin-bottom: 8px;
}

.dashboard-layout :deep(.ant-statistic-content) {
  font-size: 32px;
  font-weight: 700;
}

.dashboard-layout :deep(.ant-timeline) {
  color: #e2e8f0;
}

.dashboard-layout :deep(.ant-timeline-item-content) {
  color: #e2e8f0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .content-container {
    padding: 16px;
  }

  .chart-container {
    height: 240px;
  }

  .dashboard-layout :deep(.ant-statistic-content) {
    font-size: 24px;
  }
}

/* 动画效果 */
@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.metrics-row,
.chart-row,
.activity-row {
  animation: slideInUp 0.6s ease-out;
}

.chart-row {
  animation-delay: 0.2s;
}

.activity-row {
  animation-delay: 0.4s;
}
</style>
