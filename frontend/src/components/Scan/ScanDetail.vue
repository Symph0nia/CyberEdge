<template>
  <div class="scan-detail">
    <div class="page-header">
      <div class="header-left">
        <router-link to="/scans" class="back-link">
          <span class="icon">←</span>
          返回扫描列表
        </router-link>
        <h1>扫描详情</h1>
      </div>
      <div class="header-actions">
        <button
          v-if="scanInfo.state === 'running'"
          @click="stopScan"
          class="btn btn-warning"
        >
          停止扫描
        </button>
        <button @click="refreshData" class="btn btn-secondary">
          <span class="icon">🔄</span>
          刷新
        </button>
      </div>
    </div>

    <!-- 扫描基本信息 -->
    <div class="scan-info-card">
      <div class="info-header">
        <h2>基本信息</h2>
        <span class="status-badge" :class="getStatusClass(scanInfo.state)">
          {{ getStatusText(scanInfo.state) }}
        </span>
      </div>

      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">扫描目标:</span>
          <span class="info-value">{{ scanInfo.target_address || scanInfo.target_id }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">扫描类型:</span>
          <span class="info-value">{{ scanInfo.service_name }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">项目名称:</span>
          <span class="info-value">{{ projectName }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">创建时间:</span>
          <span class="info-value">{{ formatDate(scanInfo.created_at) }}</span>
        </div>
        <div class="info-item" v-if="scanInfo.updated_at">
          <span class="info-label">更新时间:</span>
          <span class="info-value">{{ formatDate(scanInfo.updated_at) }}</span>
        </div>
        <div class="info-item" v-if="scanInfo.state === 'running'">
          <span class="info-label">运行时长:</span>
          <span class="info-value">{{ getRunningDuration() }}</span>
        </div>
      </div>

      <!-- 进度条 -->
      <div v-if="scanInfo.state === 'running'" class="progress-section">
        <div class="progress-header">
          <span>扫描进度</span>
          <span>{{ getProgress() }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: getProgress() + '%' }"></div>
        </div>
      </div>
    </div>

    <!-- 统计概览 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">🔍</div>
        <div class="stat-content">
          <h3>{{ stats.totalResults || 0 }}</h3>
          <p>扫描结果</p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🚨</div>
        <div class="stat-content">
          <h3>{{ stats.vulnerabilities?.total || 0 }}</h3>
          <p>发现漏洞</p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">⚠️</div>
        <div class="stat-content">
          <h3>{{ stats.vulnerabilities?.critical + stats.vulnerabilities?.high || 0 }}</h3>
          <p>高危漏洞</p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🌐</div>
        <div class="stat-content">
          <h3>{{ stats.assets || 0 }}</h3>
          <p>发现资产</p>
        </div>
      </div>
    </div>

    <!-- 详细结果 -->
    <div class="results-section">
      <div class="tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="tab-button"
          :class="{ active: activeTab === tab.key }"
        >
          {{ tab.label }}
          <span v-if="tab.count" class="tab-count">{{ tab.count }}</span>
        </button>
      </div>

      <div class="tab-content">
        <!-- 漏洞列表 -->
        <div v-if="activeTab === 'vulnerabilities'" class="vulnerabilities-list">
          <div v-if="vulnerabilities.length === 0" class="empty-state">
            <div class="empty-icon">🔒</div>
            <p>暂未发现漏洞</p>
          </div>
          <div v-else>
            <div
              v-for="vuln in vulnerabilities"
              :key="vuln.id"
              class="vulnerability-item"
              :class="getSeverityClass(vuln.severity)"
            >
              <div class="vuln-header">
                <h4>{{ vuln.title }}</h4>
                <span class="severity-badge" :class="getSeverityClass(vuln.severity)">
                  {{ vuln.severity.toUpperCase() }}
                </span>
              </div>
              <div class="vuln-details">
                <p>{{ vuln.description }}</p>
                <div class="vuln-meta">
                  <span><strong>位置:</strong> {{ vuln.location }}</span>
                  <span v-if="vuln.cvss"><strong>CVSS:</strong> {{ vuln.cvss }}</span>
                  <span v-if="vuln.cve_id"><strong>CVE:</strong> {{ vuln.cve_id }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 扫描结果 -->
        <div v-if="activeTab === 'results'" class="scan-results">
          <div v-if="scanResults.length === 0" class="empty-state">
            <div class="empty-icon">📊</div>
            <p>暂无扫描结果</p>
          </div>
          <div v-else>
            <div
              v-for="result in scanResults"
              :key="result.id"
              class="result-item"
            >
              <div class="result-header">
                <h4>{{ result.service_name }}</h4>
                <span class="result-status" :class="getResultStatusClass(result.state)">
                  {{ result.state }}
                </span>
              </div>
              <div class="result-details">
                <div class="result-meta">
                  <span><strong>端口:</strong> {{ result.port }}</span>
                  <span><strong>协议:</strong> {{ result.protocol }}</span>
                  <span v-if="result.target_address"><strong>目标:</strong> {{ result.target_address }}</span>
                  <span><strong>时间:</strong> {{ formatDate(result.created_at) }}</span>
                </div>
                <div v-if="result.version" class="result-version">
                  <strong>版本:</strong> {{ result.version }}
                </div>
                <div v-if="result.banner" class="result-banner">
                  <strong>Banner:</strong>
                  <pre>{{ result.banner }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 日志 -->
        <div v-if="activeTab === 'logs'" class="scan-logs">
          <div class="logs-header">
            <h3>扫描日志</h3>
            <button @click="refreshLogs" class="btn btn-sm btn-secondary">刷新日志</button>
          </div>
          <div class="logs-content">
            <div v-if="logs.length === 0" class="empty-state">
              <div class="empty-icon">📝</div>
              <p>暂无日志记录</p>
            </div>
            <div v-else class="log-lines">
              <div
                v-for="(log, index) in logs"
                :key="index"
                class="log-line"
                :class="getLogLevelClass(log.level)"
              >
                <span class="log-time">{{ formatTime(log.timestamp) }}</span>
                <span class="log-level">{{ log.level }}</span>
                <span class="log-message">{{ log.message }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { scanApi } from '@/api/scanApi'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'

export default {
  name: 'ScanDetail',
  data() {
    return {
      loading: false,
      scanInfo: {},
      scanResults: [],
      vulnerabilities: [],
      logs: [],
      stats: {},
      projectName: '',
      activeTab: 'vulnerabilities',
      pollingInterval: null
    }
  },
  computed: {
    scanId() {
      return this.$route.params.id
    },
    tabs() {
      return [
        {
          key: 'vulnerabilities',
          label: '漏洞',
          count: this.vulnerabilities.length
        },
        {
          key: 'results',
          label: '扫描结果',
          count: this.scanResults.length
        },
        {
          key: 'logs',
          label: '日志',
          count: this.logs.length
        }
      ]
    }
  },
  async mounted() {
    await this.loadScanDetail()
    this.startPolling()
  },
  beforeUnmount() {
    this.stopPolling()
  },
  methods: {
    async loadScanDetail() {
      this.loading = true
      try {
        // 并行加载所有数据
        const [scanResponse] = await Promise.all([
          scanFrameworkApi.getScanStatus(this.scanId),
          this.loadScanResults(),
          this.loadVulnerabilities()
        ])

        this.scanInfo = scanResponse.data || {}
        await this.loadProjectName()
        await this.loadStats()
      } catch (error) {
        console.error('加载扫描详情失败:', error)
        this.$message?.error('加载扫描详情失败')
      } finally {
        this.loading = false
      }
    },

    async loadScanResults() {
      if (!this.scanInfo.project_id) return

      try {
        const response = await scanFrameworkApi.getScanResults(this.scanInfo.project_id)
        this.scanResults = response.data || []
      } catch (error) {
        console.error('加载扫描结果失败:', error)
      }
    },

    async loadVulnerabilities() {
      if (!this.scanInfo.project_id) return

      try {
        const response = await scanFrameworkApi.getProjectVulnerabilities(this.scanInfo.project_id)
        this.vulnerabilities = response.data || []
      } catch (error) {
        console.error('加载漏洞数据失败:', error)
      }
    },

    async loadProjectName() {
      if (!this.scanInfo.project_id) return

      try {
        const response = await scanApi.getProject(this.scanInfo.project_id)
        this.projectName = response.data?.name || '未知项目'
      } catch (error) {
        this.projectName = '未知项目'
      }
    },

    async loadStats() {
      if (!this.scanInfo.project_id) return

      try {
        const [vulnStatsResponse] = await Promise.all([
          scanFrameworkApi.getVulnerabilityStats(this.scanInfo.project_id)
        ])

        const vulnStats = vulnStatsResponse.data || {}
        this.stats = {
          totalResults: this.scanResults.length,
          vulnerabilities: {
            total: this.vulnerabilities.length,
            critical: vulnStats.critical || 0,
            high: vulnStats.high || 0,
            medium: vulnStats.medium || 0,
            low: vulnStats.low || 0,
            info: vulnStats.info || 0
          },
          assets: this.scanResults.filter(r => r.state === 'discovered').length
        }
      } catch (error) {
        console.error('加载统计数据失败:', error)
      }
    },

    async refreshData() {
      await this.loadScanDetail()
    },

    async refreshLogs() {
      // 模拟日志数据，实际应该从后端获取
      this.logs = [
        {
          timestamp: new Date(),
          level: 'INFO',
          message: '开始执行子域名扫描...'
        },
        {
          timestamp: new Date(Date.now() - 60000),
          level: 'INFO',
          message: '发现子域名: api.example.com'
        },
        {
          timestamp: new Date(Date.now() - 120000),
          level: 'INFO',
          message: '开始端口扫描...'
        }
      ]
    },

    async stopScan() {
      try {
        await scanFrameworkApi.stopScan(this.scanId)
        this.$message?.success('扫描已停止')
        await this.loadScanDetail()
      } catch (error) {
        console.error('停止扫描失败:', error)
        this.$message?.error('停止扫描失败')
      }
    },

    getStatusClass(status) {
      return `status-${status}`
    },

    getStatusText(status) {
      const texts = {
        'running': '运行中',
        'completed': '已完成',
        'failed': '失败',
        'stopped': '已停止'
      }
      return texts[status] || '未知'
    },

    getSeverityClass(severity) {
      return `severity-${severity?.toLowerCase()}`
    },

    getResultStatusClass(status) {
      return `result-${status}`
    },

    getLogLevelClass(level) {
      return `log-${level?.toLowerCase()}`
    },

    getProgress() {
      if (this.scanInfo.state !== 'running') return 100

      const startTime = new Date(this.scanInfo.created_at).getTime()
      const now = Date.now()
      const elapsed = now - startTime
      const estimatedTotal = 30 * 60 * 1000 // 假设30分钟完成

      return Math.min(Math.round((elapsed / estimatedTotal) * 100), 95)
    },

    getRunningDuration() {
      if (!this.scanInfo.created_at) return ''

      const startTime = new Date(this.scanInfo.created_at).getTime()
      const now = Date.now()
      const elapsed = now - startTime

      const hours = Math.floor(elapsed / (1000 * 60 * 60))
      const minutes = Math.floor((elapsed % (1000 * 60 * 60)) / (1000 * 60))
      const seconds = Math.floor((elapsed % (1000 * 60)) / 1000)

      if (hours > 0) {
        return `${hours}小时${minutes}分钟`
      } else if (minutes > 0) {
        return `${minutes}分钟${seconds}秒`
      } else {
        return `${seconds}秒`
      }
    },

    formatDate(dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString('zh-CN')
    },

    formatTime(dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleTimeString('zh-CN')
    },

    startPolling() {
      // 如果扫描正在运行，每10秒轮询一次
      if (this.scanInfo.state === 'running') {
        this.pollingInterval = setInterval(() => {
          this.loadScanDetail()
        }, 10000)
      }
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval)
      }
    }
  },
  watch: {
    'scanInfo.state'(newState) {
      if (newState === 'running') {
        this.startPolling()
      } else {
        this.stopPolling()
      }
    }
  }
}
</script>

<style scoped>
.scan-detail {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.back-link {
  color: #007bff;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.back-link:hover {
  text-decoration: underline;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.header-actions {
  display: flex;
  gap: 15px;
}

.scan-info-card {
  background: white;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.info-header h2 {
  margin: 0;
  color: #333;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
}

.status-badge.status-running {
  background: #e3f2fd;
  color: #1976d2;
}

.status-badge.status-completed {
  background: #e8f5e8;
  color: #2e7d32;
}

.status-badge.status-failed {
  background: #ffebee;
  color: #c62828;
}

.status-badge.status-stopped {
  background: #fff3e0;
  color: #ef6c00;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-label {
  color: #666;
  font-weight: 500;
}

.info-value {
  color: #333;
}

.progress-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-weight: 500;
}

.progress-bar {
  height: 8px;
  background: #eee;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #007bff;
  transition: width 0.3s ease;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  font-size: 32px;
}

.stat-content h3 {
  margin: 0 0 4px 0;
  font-size: 24px;
  color: #333;
}

.stat-content p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.results-section {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.tabs {
  display: flex;
  border-bottom: 1px solid #eee;
}

.tab-button {
  background: none;
  border: none;
  padding: 16px 24px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  color: #666;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-button.active {
  color: #007bff;
  border-bottom-color: #007bff;
}

.tab-button:hover {
  background: #f8f9fa;
}

.tab-count {
  background: #e9ecef;
  color: #495057;
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 12px;
}

.tab-button.active .tab-count {
  background: #007bff;
  color: white;
}

.tab-content {
  padding: 24px;
  min-height: 400px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.vulnerability-item {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  border-left: 4px solid #ddd;
}

.vulnerability-item.severity-critical {
  border-left-color: #dc3545;
}

.vulnerability-item.severity-high {
  border-left-color: #fd7e14;
}

.vulnerability-item.severity-medium {
  border-left-color: #ffc107;
}

.vulnerability-item.severity-low {
  border-left-color: #28a745;
}

.vuln-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.vuln-header h4 {
  margin: 0;
  color: #333;
  flex: 1;
}

.severity-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}

.severity-badge.severity-critical {
  background: #ffebee;
  color: #c62828;
}

.severity-badge.severity-high {
  background: #fff3e0;
  color: #ef6c00;
}

.severity-badge.severity-medium {
  background: #fffaef;
  color: #ff8f00;
}

.severity-badge.severity-low {
  background: #e8f5e8;
  color: #2e7d32;
}

.vuln-details p {
  margin: 0 0 12px 0;
  color: #666;
  line-height: 1.5;
}

.vuln-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 14px;
  color: #666;
}

.result-item {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.result-header h4 {
  margin: 0;
  color: #333;
}

.result-status {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.result-status.result-discovered {
  background: #e8f5e8;
  color: #2e7d32;
}

.result-status.result-open {
  background: #e3f2fd;
  color: #1976d2;
}

.result-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.result-version,
.result-banner {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
}

.result-banner pre {
  background: #f8f9fa;
  padding: 8px;
  border-radius: 4px;
  margin: 4px 0 0 0;
  overflow-x: auto;
  font-size: 12px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.logs-header h3 {
  margin: 0;
  color: #333;
}

.logs-content {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.log-lines {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.log-line {
  display: flex;
  gap: 12px;
  padding: 4px 0;
  border-bottom: 1px solid #eee;
}

.log-time {
  color: #666;
  white-space: nowrap;
}

.log-level {
  font-weight: 500;
  min-width: 50px;
}

.log-level.log-info {
  color: #007bff;
}

.log-level.log-warn {
  color: #ffc107;
}

.log-level.log-error {
  color: #dc3545;
}

.log-message {
  flex: 1;
  color: #333;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-warning {
  background: #ffc107;
  color: #212529;
}

.btn:hover {
  opacity: 0.9;
}

.icon {
  font-size: 16px;
}
</style>