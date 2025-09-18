<template>
  <div class="scan-create">
    <div class="page-header">
      <h1>创建扫描任务</h1>
      <div class="breadcrumb">
        <router-link to="/scans">扫描管理</router-link>
        <span class="separator">></span>
        <span>创建扫描</span>
      </div>
    </div>

    <div class="scan-form">
      <form @submit.prevent="createScan">
        <!-- 项目选择 -->
        <div class="form-group">
          <label for="project">选择项目 *</label>
          <select
            id="project"
            v-model="scanForm.projectId"
            required
            class="form-control"
          >
            <option value="">请选择项目...</option>
            <option
              v-for="project in projects"
              :key="project.id"
              :value="project.id"
            >
              {{ project.name }}
            </option>
          </select>
        </div>

        <!-- 扫描目标 -->
        <div class="form-group">
          <label for="target">扫描目标 *</label>
          <input
            id="target"
            v-model="scanForm.target"
            type="text"
            required
            placeholder="例如: example.com 或 192.168.1.1"
            class="form-control"
          />
          <small class="form-hint">
            支持域名、IP地址或IP段（例如：192.168.1.0/24）
          </small>
        </div>

        <!-- 扫描流水线 -->
        <div class="form-group">
          <label for="pipeline">扫描流水线 *</label>
          <select
            id="pipeline"
            v-model="scanForm.pipelineName"
            required
            class="form-control"
            @change="onPipelineChange"
          >
            <option value="">请选择扫描流水线...</option>
            <option
              v-for="pipeline in pipelines"
              :key="pipeline"
              :value="pipeline"
            >
              {{ getPipelineDisplayName(pipeline) }}
            </option>
          </select>
          <small class="form-hint" v-if="selectedPipelineInfo">
            {{ selectedPipelineInfo }}
          </small>
        </div>

        <!-- 可用工具展示 -->
        <div class="form-group" v-if="availableTools.length > 0">
          <label>可用扫描工具</label>
          <div class="tools-grid">
            <div
              v-for="category in Object.keys(toolsByCategory)"
              :key="category"
              class="tool-category"
            >
              <h4>{{ getCategoryDisplayName(category) }}</h4>
              <div class="tools-list">
                <div
                  v-for="tool in toolsByCategory[category]"
                  :key="tool.name"
                  class="tool-item"
                  :class="{ 'available': tool.available, 'unavailable': !tool.available }"
                >
                  <span class="tool-name">{{ tool.name }}</span>
                  <span class="tool-status">
                    {{ tool.available ? '✓ 可用' : '✗ 不可用' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 高级配置 -->
        <div class="form-group">
          <div class="advanced-toggle" @click="showAdvanced = !showAdvanced">
            <span>高级配置</span>
            <span class="toggle-icon">{{ showAdvanced ? '▼' : '▶' }}</span>
          </div>

          <div v-if="showAdvanced" class="advanced-options">
            <div class="form-group">
              <label for="timeout">扫描超时时间（分钟）</label>
              <input
                id="timeout"
                v-model.number="scanForm.timeout"
                type="number"
                min="1"
                max="180"
                class="form-control"
              />
            </div>

            <div class="form-group">
              <label for="concurrent">并发数</label>
              <input
                id="concurrent"
                v-model.number="scanForm.concurrent"
                type="number"
                min="1"
                max="10"
                class="form-control"
              />
            </div>
          </div>
        </div>

        <!-- 提交按钮 -->
        <div class="form-actions">
          <button
            type="button"
            @click="$router.push('/scans')"
            class="btn btn-secondary"
          >
            取消
          </button>
          <button
            type="submit"
            :disabled="loading || !isFormValid"
            class="btn btn-primary"
          >
            <span v-if="loading" class="loading-spinner">⟳</span>
            {{ loading ? '创建中...' : '开始扫描' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { scanApi } from '@/api/scanApi'
import { scanFrameworkApi } from '@/api/scanFrameworkApi'

export default {
  name: 'ScanCreate',
  data() {
    return {
      loading: false,
      showAdvanced: false,
      projects: [],
      pipelines: [],
      availableTools: [],
      scanForm: {
        projectId: '',
        target: '',
        pipelineName: '',
        timeout: 30,
        concurrent: 3
      }
    }
  },
  computed: {
    isFormValid() {
      return !!(this.scanForm.projectId &&
                this.scanForm.target &&
                this.scanForm.pipelineName)
    },
    toolsByCategory() {
      return this.availableTools.reduce((acc, tool) => {
        const category = tool.category
        if (!acc[category]) {
          acc[category] = []
        }
        acc[category].push(tool)
        return acc
      }, {})
    },
    selectedPipelineInfo() {
      const pipelineDescriptions = {
        'comprehensive': '全面扫描 - 包含子域名发现、端口扫描、服务识别、漏洞检测',
        'quick': '快速扫描 - 基础端口扫描和服务识别',
        'web': 'Web扫描 - 专注于Web应用安全检测',
        'network': '网络扫描 - 深度网络资产发现'
      }
      return pipelineDescriptions[this.scanForm.pipelineName] || ''
    }
  },
  async mounted() {
    await this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        // 并行加载数据
        const [projectsRes, pipelinesRes, toolsRes] = await Promise.all([
          scanApi.getProjects(),
          scanFrameworkApi.getAvailablePipelines(),
          scanFrameworkApi.getAvailableTools()
        ])

        this.projects = projectsRes.data || []
        this.pipelines = pipelinesRes.data || []
        this.availableTools = this.flattenTools(toolsRes.data || {})
      } catch (error) {
        console.error('加载数据失败:', error)
        this.$message?.error('加载数据失败，请刷新页面重试')
      } finally {
        this.loading = false
      }
    },

    flattenTools(toolsData) {
      const tools = []
      Object.keys(toolsData).forEach(category => {
        toolsData[category].forEach(tool => {
          tools.push({
            ...tool,
            category
          })
        })
      })
      return tools
    },

    async createScan() {
      if (!this.isFormValid) return

      this.loading = true
      try {
        const response = await scanFrameworkApi.startScan(
          this.scanForm.projectId,
          this.scanForm.target,
          this.scanForm.pipelineName
        )

        this.$message?.success('扫描任务创建成功！')

        // 跳转到扫描详情页面
        const scanId = response.data?.scan_id
        if (scanId) {
          this.$router.push(`/scans/${scanId}`)
        } else {
          this.$router.push('/scans')
        }
      } catch (error) {
        console.error('创建扫描失败:', error)
        this.$message?.error(error.response?.data?.error || '创建扫描失败')
      } finally {
        this.loading = false
      }
    },

    onPipelineChange() {
      // 当流水线改变时可以加载相应的配置
      console.log('Pipeline changed to:', this.scanForm.pipelineName)
    },

    getPipelineDisplayName(pipeline) {
      const names = {
        'comprehensive': '全面扫描',
        'quick': '快速扫描',
        'web': 'Web扫描',
        'network': '网络扫描'
      }
      return names[pipeline] || pipeline
    },

    getCategoryDisplayName(category) {
      const names = {
        'subdomain': '子域名发现',
        'port': '端口扫描',
        'service': '服务识别',
        'webtech': 'Web技术探测',
        'webpath': '目录扫描',
        'vulnerability': '漏洞检测'
      }
      return names[category] || category
    }
  }
}
</script>

<style scoped>
.scan-create {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #333;
}

.breadcrumb {
  color: #666;
  font-size: 14px;
}

.breadcrumb a {
  color: #007bff;
  text-decoration: none;
}

.breadcrumb .separator {
  margin: 0 8px;
}

.scan-form {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #333;
}

.form-control {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-control:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.form-hint {
  color: #666;
  font-size: 12px;
  margin-top: 5px;
  display: block;
}

.tools-grid {
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 15px;
  background: #f9f9f9;
}

.tool-category {
  margin-bottom: 15px;
}

.tool-category h4 {
  margin: 0 0 10px 0;
  color: #555;
  font-size: 14px;
  font-weight: 600;
}

.tools-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 8px;
}

.tool-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 13px;
}

.tool-item.available {
  background: #d4edda;
  color: #155724;
}

.tool-item.unavailable {
  background: #f8d7da;
  color: #721c24;
}

.tool-name {
  font-weight: 500;
}

.tool-status {
  font-size: 12px;
}

.advanced-toggle {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  user-select: none;
}

.advanced-toggle:hover {
  background: #e9ecef;
}

.toggle-icon {
  color: #666;
}

.advanced-options {
  margin-top: 15px;
  padding: 15px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  background: #f8f9fa;
}

.form-actions {
  display: flex;
  gap: 15px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #0056b3;
}

.btn-primary:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #545b62;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>