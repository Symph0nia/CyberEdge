<template>
  <a-layout class="homepage-layout">
    <HeaderPage />
    <a-layout-content class="homepage-content">
      <div class="content-container">
        <!-- 欢迎信息区域 -->
        <div class="welcome-section">
          <h1 class="welcome-title">CyberEdge 综合扫描器</h1>
          <p class="welcome-description">简单、高效、专业的资产扫描工具</p>

          <div class="action-buttons">
            <a-button type="primary" size="large" class="start-btn">
              <i class="ri-rocket-line"></i>
              开始使用
            </a-button>
            <a-button size="large" class="info-btn">
              <i class="ri-information-line"></i>
              了解更多
            </a-button>
          </div>
        </div>

        <!-- 功能分类 -->
        <div class="feature-header">
          <h2 class="feature-title">
            <i class="ri-apps-line"></i>
            主要功能
          </h2>
        </div>

        <!-- 功能卡片网格 -->
        <a-row :gutter="[24, 24]" class="feature-grid">
          <a-col :xs="24" :sm="12" :lg="6" v-for="(card, index) in cards" :key="index">
            <a-card class="feature-card" hoverable>
              <div class="feature-content">
                <div class="feature-icon">
                  <i :class="card.icon"></i>
                </div>
                <h3 class="card-title">{{ card.title }}</h3>
                <p class="card-description">{{ card.description }}</p>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- 技术优势 -->
        <div class="stats-section">
          <h2 class="stats-title">
            <i class="ri-line-chart-line"></i>
            技术优势
          </h2>
          <a-row :gutter="[24, 24]" class="stats-grid">
            <a-col :xs="12" :sm="6" v-for="(stat, index) in stats" :key="index">
              <div class="stat-card">
                <div class="stat-value">{{ stat.value }}</div>
                <div class="stat-label">{{ stat.label }}</div>
              </div>
            </a-col>
          </a-row>
        </div>
      </div>
    </a-layout-content>
    <FooterPage />
  </a-layout>
</template>

<script>
import { ref } from "vue";
import HeaderPage from "./HeaderPage.vue";
import FooterPage from "./FooterPage.vue";

export default {
  name: "HomePage",
  components: {
    HeaderPage,
    FooterPage,
  },
  setup() {
    const cards = ref([
      {
        icon: "ri-radar-line",
        title: "资产发现",
        description: "快速准确的子域名扫描，全面掌握数字资产分布。",
      },
      {
        icon: "ri-scan-2-line",
        title: "端口扫描",
        description: "高效的端口检测，精准识别开放服务和组件。",
      },
      {
        icon: "ri-folders-line",
        title: "路径扫描",
        description: "深度探索网站结构，发现隐藏资源和目录。",
      },
      {
        icon: "ri-fingerprint-line",
        title: "指纹识别",
        description: "准确识别技术组件和版本，提供详细分析。",
      }
    ]);

    const stats = ref([
      { value: "99%", label: "准确率" },
      { value: "< 1s", label: "响应时间" },
      { value: "24/7", label: "运行时间" },
      { value: "100+", label: "支持工具" }
    ]);

    return {
      cards,
      stats,
    };
  },
};
</script>

<style scoped>
.homepage-layout {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  min-height: 100vh;
}

.homepage-content {
  background: transparent;
  margin-top: 64px;
  padding: 0;
}

.content-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 48px 24px;
}

.welcome-section {
  text-align: center;
  margin-bottom: 80px;
  padding: 60px 0;
}

.welcome-title {
  font-size: 48px;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 16px;
  background: linear-gradient(135deg, #3b82f6, #06b6d4);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.welcome-description {
  font-size: 18px;
  color: #94a3b8;
  margin-bottom: 32px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.action-buttons {
  display: flex;
  gap: 16px;
  justify-content: center;
  flex-wrap: wrap;
}

.feature-header {
  text-align: center;
  margin-bottom: 48px;
}

.feature-title {
  font-size: 32px;
  font-weight: 600;
  color: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.feature-grid {
  margin-bottom: 80px;
}

.feature-card {
  height: 100%;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 16px;
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
  border-color: rgba(59, 130, 246, 0.5);
}

.feature-content {
  text-align: center;
  padding: 24px 16px;
}

.feature-icon {
  font-size: 48px;
  color: #3b82f6;
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.card-title {
  font-size: 20px;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 12px;
}

.card-description {
  color: #94a3b8;
  line-height: 1.6;
  margin: 0;
}

.stats-section {
  text-align: center;
}

.stats-title {
  font-size: 32px;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.stat-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 12px;
  padding: 32px 16px;
  text-align: center;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.3);
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  color: #3b82f6;
  margin-bottom: 8px;
}

.stat-label {
  color: #94a3b8;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .welcome-title {
    font-size: 36px;
  }

  .content-container {
    padding: 32px 16px;
  }

  .welcome-section {
    margin-bottom: 60px;
    padding: 40px 0;
  }

  .feature-title,
  .stats-title {
    font-size: 24px;
    flex-direction: column;
    gap: 8px;
  }

  .stat-value {
    font-size: 28px;
  }
}
</style>