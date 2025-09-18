<template>
  <div class="project-detail">
    <div class="page-header">
      <div class="header-left">
        <button @click="$router.go(-1)" class="btn-back">
          <span class="icon">←</span>
          返回
        </button>
        <div class="project-info">
          <h1>{{ project?.name }}</h1>
          <p>{{ project?.description || '暂无描述' }}</p>
        </div>
      </div>
      <div class="header-actions">
        <button @click="importScanData" class="btn btn-secondary">
          <span class="icon">📁</span>
          导入扫描数据
        </button>
        <button @click="createSampleData" class="btn btn-primary">
          <span class="icon">🧪</span>
          生成示例数据
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="project" class="project-content">
      <!-- 统计概览 -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon domain">🌐</div>
          <div class="stat-content">
            <div class="stat-value">{{ stats?.domain_count || 0 }}</div>
            <div class="stat-label">域名</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon subdomain">🌍</div>
          <div class="stat-content">
            <div class="stat-value">{{ stats?.subdomain_count || 0 }}</div>
            <div class="stat-label">子域名</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon ip">🖥️</div>
          <div class="stat-content">
            <div class="stat-value">{{ stats?.ip_count || 0 }}</div>
            <div class="stat-label">IP地址</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon port">🔌</div>
          <div class="stat-content">
            <div class="stat-value">{{ stats?.port_count || 0 }}</div>
            <div class="stat-label">开放端口</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon service">⚙️</div>
          <div class="stat-content">
            <div class="stat-value">{{ stats?.service_count || 0 }}</div>
            <div class="stat-label">发现服务</div>
          </div>
        </div>
      </div>

      <!-- 漏洞统计 -->
      <div class="vulnerability-section">
        <h2>漏洞统计</h2>
        <div class="vulnerability-stats">
          <div class="vuln-card critical">
            <div class="vuln-count">{{ stats?.vulnerability_stats?.critical || 0 }}</div>
            <div class="vuln-label">严重</div>
          </div>
          <div class="vuln-card high">
            <div class="vuln-count">{{ stats?.vulnerability_stats?.high || 0 }}</div>
            <div class="vuln-label">高危</div>
          </div>
          <div class="vuln-card medium">
            <div class="vuln-count">{{ stats?.vulnerability_stats?.medium || 0 }}</div>
            <div class="vuln-label">中危</div>
          </div>
          <div class="vuln-card low">
            <div class="vuln-count">{{ stats?.vulnerability_stats?.low || 0 }}</div>
            <div class="vuln-label">低危</div>
          </div>
          <div class="vuln-card info">
            <div class="vuln-count">{{ stats?.vulnerability_stats?.info || 0 }}</div>
            <div class="vuln-label">信息</div>
          </div>
        </div>
      </div>

      <!-- 域名结构树 -->
      <div class="domain-section">
        <h2>域名结构</h2>
        <div v-if="project.domains && project.domains.length > 0" class="domain-tree">
          <div v-for="domain in project.domains" :key="domain.id" class="domain-node">
            <div class="domain-header" @click="toggleDomain(domain.id)">
              <span class="expand-icon" :class="{ expanded: expandedDomains.has(domain.id) }">▶</span>
              <span class="domain-name">{{ domain.name }}</span>
              <span class="domain-count">({{ domain.subdomains?.length || 0 }} 子域名)</span>
            </div>

            <div v-if="expandedDomains.has(domain.id)" class="subdomain-list">
              <div v-for="subdomain in domain.subdomains" :key="subdomain.id" class="subdomain-node">
                <div class="subdomain-header" @click="toggleSubdomain(subdomain.id)">
                  <span class="expand-icon" :class="{ expanded: expandedSubdomains.has(subdomain.id) }">▶</span>
                  <span class="subdomain-name">{{ subdomain.name }}</span>
                  <span class="ip-count">({{ subdomain.ip_addresses?.length || 0 }} IP)</span>
                </div>

                <div v-if="expandedSubdomains.has(subdomain.id)" class="ip-list">
                  <div v-for="ip in subdomain.ip_addresses" :key="ip.id" class="ip-node">
                    <div class="ip-header" @click="toggleIP(ip.id)">
                      <span class="expand-icon" :class="{ expanded: expandedIPs.has(ip.id) }">▶</span>
                      <span class="ip-address">{{ ip.address }}</span>
                      <span class="port-count">({{ ip.ports?.length || 0 }} 端口)</span>
                    </div>

                    <div v-if="expandedIPs.has(ip.id)" class="port-list">
                      <div v-for="port in ip.ports" :key="port.id" class="port-node">
                        <div class="port-info">
                          <span class="port-number">{{ port.number }}/{{ port.protocol }}</span>
                          <span class="port-state" :class="port.state">{{ port.state }}</span>
                          <span v-if="port.service" class="service-name">{{ port.service.name }}</span>
                        </div>

                        <div v-if="port.service && port.service.vulnerabilities?.length > 0" class="service-vulns">
                          <span class="vuln-indicator" :class="getHighestSeverity(port.service.vulnerabilities)">
                            {{ port.service.vulnerabilities.length }} 漏洞
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-domains">
          <div class="empty-icon">🌐</div>
          <p>暂无域名数据</p>
          <button @click="importScanData" class="btn btn-primary">导入扫描数据</button>
        </div>
      </div>
    </div>

    <!-- 导入数据模态框 -->
    <div v-if="showImportModal" class="modal-overlay" @click="closeImportModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>导入扫描数据</h2>
          <button @click="closeImportModal" class="btn-close">×</button>
        </div>
        <div class="modal-content">
          <div class="import-options">
            <div class="option-card" @click="selectImportType('json')">
              <div class="option-icon">📄</div>
              <h3>JSON 文件</h3>
              <p>上传标准格式的扫描结果JSON文件</p>
            </div>
            <div class="option-card" @click="selectImportType('nmap')">
              <div class="option-icon">🗂️</div>
              <h3>Nmap 结果</h3>
              <p>导入Nmap XML格式的扫描结果</p>
            </div>
          </div>

          <div v-if="importType" class="file-upload">
            <input
              ref="fileInput"
              type="file"
              :accept="importType === 'json' ? '.json' : '.xml'"
              @change="handleFileSelect"
              style="display: none"
            />
            <div
              @click="$refs.fileInput.click()"
              @drop.prevent="handleFileDrop"
              @dragover.prevent
              class="upload-area"
              :class="{ 'drag-over': isDragOver }"
            >
              <div class="upload-icon">📁</div>
              <p>点击选择文件或拖拽文件到此处</p>
              <small>支持 {{ importType === 'json' ? 'JSON' : 'XML' }} 格式</small>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { scanApi } from '@/api/scanApi'

export default {
  name: 'ProjectDetail',
  data() {
    return {
      project: null,
      stats: null,
      loading: true,
      showImportModal: false,
      importType: null,
      isDragOver: false,
      expandedDomains: new Set(),
      expandedSubdomains: new Set(),
      expandedIPs: new Set()
    }
  },
  async mounted() {
    await this.loadProject()
    await this.loadStats()
  },
  methods: {
    async loadProject() {
      try {
        this.loading = true
        const projectId = this.$route.params.id
        const response = await scanApi.getProject(projectId)
        this.project = response.data.data
      } catch (error) {
        console.error('加载项目失败:', error)
        this.$toast.error('加载项目失败')
      } finally {
        this.loading = false
      }
    },

    async loadStats() {
      try {
        const projectId = this.$route.params.id
        const response = await scanApi.getProjectStats(projectId)
        this.stats = response.data.data
      } catch (error) {
        console.error('加载统计数据失败:', error)
      }
    },

    async createSampleData() {
      try {
        const projectId = this.$route.params.id
        await scanApi.createSampleData(projectId)
        this.$toast.success('示例数据已生成')
        await this.loadProject()
        await this.loadStats()
      } catch (error) {
        console.error('生成示例数据失败:', error)
        this.$toast.error('生成示例数据失败')
      }
    },

    importScanData() {
      this.showImportModal = true
    },

    closeImportModal() {
      this.showImportModal = false
      this.importType = null
    },

    selectImportType(type) {
      this.importType = type
    },

    handleFileSelect(event) {
      const file = event.target.files[0]
      if (file) {
        this.processFile(file)
      }
    },

    handleFileDrop(event) {
      this.isDragOver = false
      const file = event.dataTransfer.files[0]
      if (file) {
        this.processFile(file)
      }
    },

    async processFile(file) {
      try {
        const formData = new FormData()
        formData.append('file', file)
        formData.append('type', this.importType)

        const projectId = this.$route.params.id
        await scanApi.importScanResults(projectId, formData)

        this.$toast.success('数据导入成功')
        this.closeImportModal()
        await this.loadProject()
        await this.loadStats()
      } catch (error) {
        console.error('导入数据失败:', error)
        this.$toast.error('导入数据失败')
      }
    },

    toggleDomain(domainId) {
      if (this.expandedDomains.has(domainId)) {
        this.expandedDomains.delete(domainId)
      } else {
        this.expandedDomains.add(domainId)
      }
    },

    toggleSubdomain(subdomainId) {
      if (this.expandedSubdomains.has(subdomainId)) {
        this.expandedSubdomains.delete(subdomainId)
      } else {
        this.expandedSubdomains.add(subdomainId)
      }
    },

    toggleIP(ipId) {
      if (this.expandedIPs.has(ipId)) {
        this.expandedIPs.delete(ipId)
      } else {
        this.expandedIPs.add(ipId)
      }
    },

    getHighestSeverity(vulnerabilities) {
      const severities = vulnerabilities.map(v => v.severity)
      if (severities.includes('critical')) return 'critical'
      if (severities.includes('high')) return 'high'
      if (severities.includes('medium')) return 'medium'
      if (severities.includes('low')) return 'low'
      return 'info'
    }
  }
}
</script>

<style scoped>
.project-detail {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.btn-back {
  background: none;
  border: 1px solid #e1e5e9;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  transition: all 0.2s;
}

.btn-back:hover {
  border-color: #3b82f6;
  color: #3b82f6;
}

.project-info h1 {
  margin: 0 0 8px 0;
  color: #1a1a1a;
  font-size: 28px;
  font-weight: 600;
}

.project-info p {
  margin: 0;
  color: #6b7280;
  font-size: 16px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.stat-card {
  background: white;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  font-size: 24px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon.domain { background: #dbeafe; }
.stat-icon.subdomain { background: #dcfce7; }
.stat-icon.ip { background: #fef3c7; }
.stat-icon.port { background: #e0e7ff; }
.stat-icon.service { background: #fce7f3; }

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.stat-label {
  color: #6b7280;
  font-size: 14px;
}

.vulnerability-section {
  background: white;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 32px;
}

.vulnerability-section h2 {
  margin: 0 0 20px 0;
  color: #1a1a1a;
  font-size: 20px;
  font-weight: 600;
}

.vulnerability-stats {
  display: flex;
  gap: 16px;
}

.vuln-card {
  flex: 1;
  text-align: center;
  padding: 16px;
  border-radius: 8px;
  border: 2px solid;
}

.vuln-card.critical {
  background: #fef2f2;
  border-color: #dc2626;
  color: #dc2626;
}

.vuln-card.high {
  background: #fef3ec;
  border-color: #ea580c;
  color: #ea580c;
}

.vuln-card.medium {
  background: #fffbeb;
  border-color: #d97706;
  color: #d97706;
}

.vuln-card.low {
  background: #f0fdf4;
  border-color: #16a34a;
  color: #16a34a;
}

.vuln-card.info {
  background: #f0f9ff;
  border-color: #0284c7;
  color: #0284c7;
}

.vuln-count {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 4px;
}

.vuln-label {
  font-size: 14px;
  font-weight: 500;
}

.domain-section {
  background: white;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  padding: 24px;
}

.domain-section h2 {
  margin: 0 0 20px 0;
  color: #1a1a1a;
  font-size: 20px;
  font-weight: 600;
}

.domain-tree {
  max-height: 600px;
  overflow-y: auto;
}

.domain-node, .subdomain-node, .ip-node {
  margin-bottom: 8px;
}

.domain-header, .subdomain-header, .ip-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.domain-header:hover, .subdomain-header:hover, .ip-header:hover {
  background: #f9fafb;
}

.expand-icon {
  transition: transform 0.2s;
  font-size: 12px;
  color: #6b7280;
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.domain-name {
  font-weight: 600;
  color: #1a1a1a;
}

.subdomain-name {
  font-weight: 500;
  color: #374151;
}

.ip-address {
  font-family: monospace;
  color: #1a1a1a;
}

.domain-count, .ip-count, .port-count {
  color: #6b7280;
  font-size: 14px;
}

.subdomain-list {
  margin-left: 24px;
  border-left: 2px solid #f3f4f6;
  padding-left: 16px;
}

.ip-list {
  margin-left: 24px;
  border-left: 2px solid #f3f4f6;
  padding-left: 16px;
}

.port-list {
  margin-left: 24px;
  border-left: 2px solid #f3f4f6;
  padding-left: 16px;
}

.port-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  margin-bottom: 4px;
  background: #f9fafb;
  border-radius: 6px;
}

.port-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.port-number {
  font-family: monospace;
  font-weight: 600;
  color: #1a1a1a;
}

.port-state {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.port-state.open {
  background: #dcfce7;
  color: #16a34a;
}

.port-state.closed {
  background: #fee2e2;
  color: #dc2626;
}

.service-name {
  color: #6b7280;
  font-size: 14px;
}

.service-vulns {
  display: flex;
  align-items: center;
}

.vuln-indicator {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.vuln-indicator.critical {
  background: #fee2e2;
  color: #dc2626;
}

.vuln-indicator.high {
  background: #fed7aa;
  color: #ea580c;
}

.vuln-indicator.medium {
  background: #fef3c7;
  color: #d97706;
}

.vuln-indicator.low {
  background: #dcfce7;
  color: #16a34a;
}

.empty-domains {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e1e5e9;
}

.modal-content {
  padding: 24px;
}

.import-options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.option-card {
  border: 2px solid #e1e5e9;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.option-card:hover {
  border-color: #3b82f6;
  background: #f8fafc;
}

.option-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.option-card h3 {
  margin: 0 0 8px 0;
  color: #1a1a1a;
}

.option-card p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.upload-area {
  border: 2px dashed #e1e5e9;
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-area:hover,
.upload-area.drag-over {
  border-color: #3b82f6;
  background: #f8fafc;
}

.upload-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f4f6;
  border-top: 4px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6b7280;
}

.icon {
  font-style: normal;
}
</style>