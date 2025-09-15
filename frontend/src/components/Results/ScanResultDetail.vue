<template>
  <a-layout class="result-layout">
    <HeaderPage />
    <a-layout-content class="result-content">
      <div class="content-container">
        <!-- 返回按钮 -->
        <div class="back-section">
          <a-button @click="goBack" type="text" class="back-btn">
            <i class="ri-arrow-left-line"></i>
            返回结果列表
          </a-button>
        </div>

        <!-- 结果详情卡片 -->
        <a-card class="result-card" :title="cardTitle" :loading="loading">
          <template #extra>
            <a-tag :color="getStatusColor(status)">
              {{ getStatusText(status) }}
            </a-tag>
          </template>

          <a-descriptions :column="2" bordered class="result-descriptions">
            <a-descriptions-item label="任务ID">{{ taskId }}</a-descriptions-item>
            <a-descriptions-item label="扫描类型">{{ scanType }}</a-descriptions-item>
            <a-descriptions-item label="目标地址">{{ target }}</a-descriptions-item>
            <a-descriptions-item label="创建时间">{{ formatDate(createdAt) }}</a-descriptions-item>
            <a-descriptions-item label="完成时间" v-if="completedAt">{{ formatDate(completedAt) }}</a-descriptions-item>
            <a-descriptions-item label="执行时长" v-if="completedAt">{{ getDuration() }}</a-descriptions-item>
          </a-descriptions>

          <!-- 结果数据 -->
          <div class="result-section" v-if="resultData">
            <h3 class="section-title">扫描结果</h3>
            <a-table
              :dataSource="resultData"
              :columns="getColumns()"
              :pagination="{ pageSize: 10 }"
              size="small"
              class="result-table"
            >
            </a-table>
          </div>

          <!-- 原始输出 -->
          <div class="raw-output-section" v-if="rawOutput">
            <h3 class="section-title">原始输出</h3>
            <a-typography-paragraph>
              <pre class="raw-output">{{ rawOutput }}</pre>
            </a-typography-paragraph>
          </div>
        </a-card>
      </div>
    </a-layout-content>
    <FooterPage />
  </a-layout>
</template>

<script>
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import HeaderPage from '../HeaderPage.vue';
import FooterPage from '../FooterPage.vue';

export default {
  name: 'ScanResultDetail',
  components: {
    HeaderPage,
    FooterPage,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const loading = ref(true);

    // 模拟数据，实际应该从API获取
    const taskId = ref(route.params.id);
    const scanType = ref('');
    const target = ref('');
    const status = ref('');
    const createdAt = ref(null);
    const completedAt = ref(null);
    const resultData = ref([]);
    const rawOutput = ref('');

    const cardTitle = computed(() => {
      const typeMap = {
        'port': '端口扫描详情',
        'subdomain': '子域名扫描详情',
        'path': '路径扫描详情'
      };
      return typeMap[scanType.value] || '扫描详情';
    });

    const getStatusColor = (status) => {
      const colorMap = {
        'completed': 'success',
        'running': 'processing',
        'failed': 'error',
        'pending': 'warning'
      };
      return colorMap[status] || 'default';
    };

    const getStatusText = (status) => {
      const textMap = {
        'completed': '已完成',
        'running': '运行中',
        'failed': '失败',
        'pending': '等待中'
      };
      return textMap[status] || status;
    };

    const formatDate = (date) => {
      if (!date) return '-';
      return new Date(date).toLocaleString('zh-CN');
    };

    const getDuration = () => {
      if (!createdAt.value || !completedAt.value) return '-';
      const diff = new Date(completedAt.value) - new Date(createdAt.value);
      return `${Math.round(diff / 1000)}秒`;
    };

    const getColumns = () => {
      // 根据扫描类型返回不同的列配置
      const commonColumns = [
        { title: '#', dataIndex: 'index', width: 60 },
        { title: '地址', dataIndex: 'address', ellipsis: true },
        { title: '状态', dataIndex: 'status', width: 100 }
      ];

      if (scanType.value === 'port') {
        return [
          ...commonColumns,
          { title: '端口', dataIndex: 'port', width: 80 },
          { title: '服务', dataIndex: 'service', width: 100 }
        ];
      } else if (scanType.value === 'subdomain') {
        return [
          ...commonColumns,
          { title: 'IP地址', dataIndex: 'ip', width: 120 }
        ];
      } else if (scanType.value === 'path') {
        return [
          ...commonColumns,
          { title: '响应码', dataIndex: 'code', width: 80 },
          { title: '大小', dataIndex: 'size', width: 80 }
        ];
      }
      return commonColumns;
    };

    const fetchResultDetail = async () => {
      loading.value = true;
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));

        // 模拟数据
        scanType.value = route.path.includes('port') ? 'port' :
                          route.path.includes('subdomain') ? 'subdomain' : 'path';
        target.value = 'example.com';
        status.value = 'completed';
        createdAt.value = new Date(Date.now() - 300000); // 5分钟前
        completedAt.value = new Date();
        resultData.value = [
          { index: 1, address: 'example.com:80', status: 'open', port: 80, service: 'HTTP' },
          { index: 2, address: 'example.com:443', status: 'open', port: 443, service: 'HTTPS' }
        ];
        rawOutput.value = 'Nmap scan results...\n80/tcp open http\n443/tcp open https';
      } catch (error) {
        console.error('获取扫描详情失败:', error);
      } finally {
        loading.value = false;
      }
    };

    const goBack = () => {
      router.go(-1);
    };

    onMounted(() => {
      fetchResultDetail();
    });

    return {
      loading,
      taskId,
      scanType,
      target,
      status,
      createdAt,
      completedAt,
      resultData,
      rawOutput,
      cardTitle,
      getStatusColor,
      getStatusText,
      formatDate,
      getDuration,
      getColumns,
      goBack
    };
  }
};
</script>

<style scoped>
.result-layout {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  min-height: 100vh;
}

.result-content {
  background: transparent;
  margin-top: 64px;
  padding: 0;
}

.content-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.back-section {
  margin-bottom: 24px;
}

.back-btn {
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 8px;
}

.back-btn:hover {
  color: #3b82f6;
}

.result-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 16px;
}

.section-title {
  color: #f1f5f9;
  font-size: 16px;
  font-weight: 600;
  margin: 24px 0 16px 0;
}

.raw-output {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(51, 65, 85, 0.3);
  border-radius: 8px;
  padding: 16px;
  color: #e2e8f0;
  font-family: 'Courier New', monospace;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
}

/* Ant Design组件样式覆盖 */
.result-card :deep(.ant-card-head) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.result-card :deep(.ant-card-head-title) {
  color: #f1f5f9;
  font-weight: 600;
}

.result-card :deep(.ant-card-body) {
  background: transparent;
  color: #e2e8f0;
}

.result-card :deep(.ant-descriptions-item-label) {
  color: #94a3b8;
  font-weight: 500;
}

.result-card :deep(.ant-descriptions-item-content) {
  color: #e2e8f0;
}

.result-card :deep(.ant-table) {
  background: transparent;
}

.result-card :deep(.ant-table-thead > tr > th) {
  background: rgba(15, 23, 42, 0.8);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
  color: #94a3b8;
}

.result-card :deep(.ant-table-tbody > tr > td) {
  background: transparent;
  border-bottom: 1px solid rgba(51, 65, 85, 0.2);
  color: #e2e8f0;
}

.result-card :deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(51, 65, 85, 0.3);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .content-container {
    padding: 16px;
  }

  .result-card :deep(.ant-descriptions) {
    font-size: 14px;
  }
}
</style>