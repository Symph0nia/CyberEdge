<template>
  <div
    class="bg-gray-900 text-white flex flex-col min-h-screen animate-fade-in"
  >
    <!-- 顶部导航栏 -->
    <HeaderPage />

    <!-- 主体内容 -->
    <div class="container mx-auto px-4 py-8 flex-1 mt-16">
      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <!-- 卡片 1: 全部任务 -->
        <div
          class="bg-gray-800 p-6 rounded-lg shadow-md transform hover:scale-105 transition duration-500"
        >
          <div class="flex items-center">
            <div class="text-blue-400 text-3xl">📋</div>
            <div class="ml-4">
              <h3 class="text-xl font-bold">全部任务</h3>
              <p
                class="text-2xl animate-number-scroll"
                :data-target="metrics.total_tasks"
              >
                {{ metrics.total_tasks }}
              </p>
            </div>
          </div>
        </div>
        <!-- 卡片 2: 进行中的任务 -->
        <div
          class="bg-gray-800 p-6 rounded-lg shadow-md transform hover:scale-105 transition duration-500"
        >
          <div class="flex items-center">
            <div class="text-yellow-400 text-3xl">⏳</div>
            <div class="ml-4">
              <h3 class="text-xl font-bold">进行中的任务</h3>
              <p
                class="text-2xl animate-number-scroll"
                :data-target="metrics.in_progress_tasks"
              >
                {{ metrics.in_progress_tasks }}
              </p>
            </div>
          </div>
        </div>
        <!-- 卡片 3: 完成的任务 -->
        <div
          class="bg-gray-800 p-6 rounded-lg shadow-md transform hover:scale-105 transition duration-500"
        >
          <div class="flex items-center">
            <div class="text-green-400 text-3xl">✅</div>
            <div class="ml-4">
              <h3 class="text-xl font-bold">完成的任务</h3>
              <p
                class="text-2xl animate-number-scroll"
                :data-target="metrics.completed_tasks"
              >
                {{ metrics.completed_tasks }}
              </p>
            </div>
          </div>
        </div>
        <!-- 卡片 4: 失败的任务 -->
        <div
          class="bg-gray-800 p-6 rounded-lg shadow-md transform hover:scale-105 transition duration-500"
        >
          <div class="flex items-center">
            <div class="text-red-400 text-3xl">🚫</div>
            <div class="ml-4">
              <h3 class="text-xl font-bold">失败的任务</h3>
              <p
                class="text-2xl animate-number-scroll"
                :data-target="metrics.failed_tasks"
              >
                {{ metrics.failed_tasks }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近7天扫描分析图表 -->
      <div class="bg-gray-800 p-6 rounded-lg shadow-md mb-8">
        <h2 class="text-2xl font-bold mb-4">最近7天扫描分析</h2>
        <div class="h-80">
          <!-- 增加高度 -->
          <BarChart :chartData="chartData" :chartOptions="chartOptions" />
        </div>
      </div>

      <!-- 活动日志 -->
      <div class="bg-gray-800 p-6 rounded-lg shadow-md">
        <h2 class="text-2xl font-bold mb-4">最近活动</h2>
        <ul class="space-y-4">
          <li
            v-for="(log, index) in activityLogs"
            :key="index"
            class="flex items-start animate-fade-in-up"
          >
            <span class="text-blue-400 text-2xl mr-4">🕒</span>
            <div>
              <p class="font-bold">{{ log.time }}</p>
              <p>{{ log.message }}</p>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <!-- 页脚 -->
    <FooterPage />
  </div>
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
    };
  },
};
</script>

<style>
@import url("https://fonts.googleapis.com/css2?family=Roboto:wght@400;700&display=swap");

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 1s ease-out;
}

.animate-fade-in-up {
  animation: fade-in-up 0.5s ease-out;
}
</style>
