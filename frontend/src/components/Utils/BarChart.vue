<template>
  <Bar :data="chartData" :options="mergedOptions" :height="300" />
</template>

<script>
import { computed } from "vue";
import { Bar } from "vue-chartjs";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
} from "chart.js";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale
);

export default {
  name: "BarChart",
  components: { Bar },
  props: {
    chartData: {
      type: Object,
      required: true,
    },
    chartOptions: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props) {
    const defaultOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: "top",
          labels: {
            font: {
              size: 14,
            },
            color: "#FFFFFF", // 白色文字
          },
        },
        tooltip: {
          mode: "index",
          intersect: false,
          backgroundColor: "rgba(0, 0, 0, 0.8)",
          titleColor: "#FFFFFF",
          bodyColor: "#FFFFFF",
          borderColor: "#FFFFFF",
          borderWidth: 1,
        },
      },
      scales: {
        x: {
          grid: {
            color: "rgba(255, 255, 255, 0.1)", // 淡白色网格线
          },
          ticks: {
            color: "#FFFFFF", // 白色文字
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: "rgba(255, 255, 255, 0.1)", // 淡白色网格线
          },
          ticks: {
            color: "#FFFFFF", // 白色文字
            callback: function (value) {
              return value.toLocaleString(); // 格式化大数字
            },
          },
        },
      },
    };

    const mergedOptions = computed(() => ({
      ...defaultOptions,
      ...props.chartOptions,
    }));

    return { mergedOptions };
  },
};
</script>
