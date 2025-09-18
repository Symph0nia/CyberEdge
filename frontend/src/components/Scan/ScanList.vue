<template>
  <div class="scan-list">
    <div class="page-header">
      <h1>扫描管理</h1>
      <router-link to="/scans/create" class="btn btn-primary">
        <span class="icon">+</span>
        创建扫描
      </router-link>
    </div>

    <!-- 筛选条件 -->
    <div class="filters">
      <div class="filter-group">
        <select v-model="filters.projectId" @change="loadScans" class="filter-select">
          <option value="">所有项目</option>
          <option
            v-for="project in projects"
            :key="project.id"
            :value="project.id"
          >
            {{ project.name }}
          </option>
        </select>

        <select v-model="filters.status" @change="loadScans" class="filter-select">
          <option value="">所有状态</option>
          <option value="running">运行中</option>
          <option value="completed">已完成</option>
          <option value="failed">失败</option>
          <option value="stopped">已停止</option>
        </select>

        <input
          v-model="filters.target"
          @input="debounceSearch"
          placeholder="搜索目标..."
          class="filter-input"
        />
      </div>

      <div class="filter-actions">
        <button @click="refreshScans" class="btn btn-secondary">
          <span class="icon">🔄</span>
          刷新
        </button>
      </div>
    </div>

    <!-- 扫描任务列表 -->
    <div class="scans-container">
      <div v-if="loading" class="loading">
        <div class="loading-spinner">⟳</div>
        <span>加载中...</span>
      </div>

      <div v-else-if="scans.length === 0" class="empty-state">
        <div class="empty-icon">📊</div>
        <h3>暂无扫描任务</h3>
        <p>创建您的第一个扫描任务开始安全检测</p>
        <router-link to="/scans/create" class="btn btn-primary">
          创建扫描任务
        </router-link>
      </div>

      <div v-else class="scans-grid">
        <div
          v-for="scan in scans"
          :key="scan.id"
          class="scan-card"
          :class="getStatusClass(scan.state)"
          @click="navigateToDetail(scan.id)"
        >
          <div class="scan-header">
            <div class="scan-info">
              <h3>{{ scan.target_address || scan.target_id }}</h3>
              <span class="scan-pipeline">{{ scan.service_name }}</span>
            </div>
            <div class="scan-status">
              <span class="status-badge" :class="getStatusClass(scan.state)">
                {{ getStatusText(scan.state) }}
              </span>
            </div>
          </div>

          <div class="scan-details">
            <div class="scan-meta">
              <div class="meta-item">
                <span class="meta-label">项目:</span>
                <span class="meta-value">{{ getProjectName(scan.project_id) }}</span>
              </div>
              <div class="meta-item">
                <span class="meta-label">创建时间:</span>
                <span class="meta-value">{{ formatDate(scan.created_at) }}</span>
              </div>
              <div class="meta-item" v-if="scan.updated_at">
                <span class="meta-label">更新时间:</span>
                <span class="meta-value">{{ formatDate(scan.updated_at) }}</span>
              </div>
            </div>

            <div class="scan-progress" v-if="scan.state === 'running'">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: getProgress(scan) + '%' }"></div>
              </div>
              <span class="progress-text">{{ getProgress(scan) }}%</span>
            </div>
          </div>

          <div class="scan-actions" @click.stop>
            <button
              v-if="scan.state === 'running'"
              @click="stopScan(scan.id)"
              class="btn btn-sm btn-warning"
            >
              停止
            </button>
            <button
              @click="navigateToDetail(scan.id)"
              class="btn btn-sm btn-primary"
            >
              查看详情
            </button>
            <button
              v-if="scan.state !== 'running'"
              @click="deleteScan(scan.id)"
              class="btn btn-sm btn-danger"
            >
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="pagination.total > pagination.pageSize" class="pagination">
      <button
        @click="changePage(pagination.current - 1)"
        :disabled="pagination.current <= 1"
        class="btn btn-secondary"
      >
        上一页
      </button>
      <span class="page-info">
        第 {{ pagination.current }} 页，共 {{ Math.ceil(pagination.total / pagination.pageSize) }} 页
      </span>
      <button
        @click="changePage(pagination.current + 1)"
        :disabled="pagination.current >= Math.ceil(pagination.total / pagination.pageSize)"
        class="btn btn-secondary"
      >
        下一页
      </button>
    </div>
  </div>
</template>

<script>
import { scanApi } from '@/api/scanApi'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'

export default {
  name: 'ScanList',
  data() {
    return {
      loading: false,
      scans: [],
      projects: [],
      filters: {
        projectId: '',
        status: '',
        target: ''
      },
      pagination: {
        current: 1,
        pageSize: 10,
        total: 0
      },
      searchTimeout: null
    }
  },
  async mounted() {
    await this.loadProjects()
    await this.loadScans()
    this.startPolling()
  },
  beforeUnmount() {
    this.stopPolling()
  },
  methods: {
    async loadProjects() {
      try {
        const response = await scanApi.getProjects()
        this.projects = response.data || []
      } catch (error) {
        console.error('加载项目失败:', error)
      }
    },

    async loadScans() {
      this.loading = true
      try {
        let scans = []

        if (this.filters.projectId) {
          // 如果选择了特定项目，获取该项目的扫描结果
          const response = await scanFrameworkApi.getProjectScans(
            this.filters.projectId,
            {
              status: this.filters.status,
              target: this.filters.target,
              page: this.pagination.current,
              limit: this.pagination.pageSize
            }
          )
          scans = response.data || []
        } else {
          // 否则获取所有项目的扫描结果
          const projectPromises = this.projects.map(project =>
            scanFrameworkApi.getProjectScans(project.id, {
              status: this.filters.status,
              target: this.filters.target
            }).catch(() => ({ data: [] }))
          )

          const results = await Promise.all(projectPromises)
          scans = results.flatMap(result => result.data || [])
        }

        this.scans = scans
        this.pagination.total = scans.length
      } catch (error) {
        console.error('加载扫描任务失败:', error)
        this.$message?.error('加载扫描任务失败')
      } finally {
        this.loading = false
      }
    },

    async refreshScans() {
      await this.loadScans()
    },

    async stopScan(scanId) {
      try {
        await scanFrameworkApi.stopScan(scanId)
        this.$message?.success('扫描任务已停止')
        await this.loadScans()
      } catch (error) {
        console.error('停止扫描失败:', error)
        this.$message?.error('停止扫描失败')
      }
    },

    async deleteScan(scanId) {
      if (!confirm('确定要删除这个扫描任务吗？此操作不可恢复。')) {
        return
      }

      try {
        await scanFrameworkApi.deleteScan(scanId)
        this.$message?.success('扫描任务已删除')
        await this.loadScans()
      } catch (error) {
        console.error('删除扫描失败:', error)
        this.$message?.error('删除扫描失败')
      }
    },

    debounceSearch() {
      clearTimeout(this.searchTimeout)
      this.searchTimeout = setTimeout(() => {
        this.loadScans()
      }, 500)
    },

    changePage(page) {
      this.pagination.current = page
      this.loadScans()
    },

    navigateToDetail(scanId) {
      this.$router.push(`/scans/${scanId}`)
    },

    getProjectName(projectId) {
      const project = this.projects.find(p => p.id === projectId)
      return project ? project.name : '未知项目'
    },

    getStatusClass(status) {
      const classes = {
        'running': 'status-running',
        'completed': 'status-completed',
        'failed': 'status-failed',
        'stopped': 'status-stopped'
      }
      return classes[status] || 'status-unknown'
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

    getProgress(scan) {
      // 简单的进度计算，实际应该从后端获取
      if (scan.state !== 'running') return 100

      const startTime = new Date(scan.created_at).getTime()
      const now = Date.now()
      const elapsed = now - startTime
      const estimatedTotal = 30 * 60 * 1000 // 假设30分钟完成

      return Math.min(Math.round((elapsed / estimatedTotal) * 100), 95)
    },

    formatDate(dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString('zh-CN')
    },

    startPolling() {
      // 每30秒轮询一次更新运行中的扫描状态
      this.pollingInterval = setInterval(() => {
        const hasRunningScans = this.scans.some(scan => scan.state === 'running')
        if (hasRunningScans) {
          this.loadScans()
        }
      }, 30000)
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval)
      }
    }
  }
}
</script>

<style scoped>
.scan-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.filters {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.filter-group {
  display: flex;
  gap: 15px;
  align-items: center;
}

.filter-select,
.filter-input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.filter-select {
  min-width: 150px;
}

.filter-input {
  min-width: 200px;
}

.scans-container {
  min-height: 400px;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #666;
}

.loading-spinner {
  font-size: 24px;
  margin-bottom: 10px;
  animation: spin 1s linear infinite;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 20px;
}

.empty-state h3 {
  margin: 0 0 10px 0;
  color: #333;
}

.scans-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.scan-card {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
  border-left: 4px solid #ddd;
}

.scan-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.scan-card.status-running {
  border-left-color: #007bff;
}

.scan-card.status-completed {
  border-left-color: #28a745;
}

.scan-card.status-failed {
  border-left-color: #dc3545;
}

.scan-card.status-stopped {
  border-left-color: #ffc107;
}

.scan-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 15px;
}

.scan-info h3 {
  margin: 0 0 5px 0;
  color: #333;
  font-size: 16px;
}

.scan-pipeline {
  color: #666;
  font-size: 14px;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
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

.scan-details {
  margin-bottom: 15px;
}

.scan-meta {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
}

.meta-label {
  color: #666;
}

.meta-value {
  color: #333;
  font-weight: 500;
}

.scan-progress {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: #eee;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #007bff;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: #666;
  min-width: 30px;
}

.scan-actions {
  display: flex;
  gap: 10px;
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

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn:hover {
  opacity: 0.9;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  margin-top: 30px;
}

.page-info {
  color: #666;
  font-size: 14px;
}

.icon {
  font-size: 16px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>