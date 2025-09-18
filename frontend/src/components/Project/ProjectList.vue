<template>
  <div class="project-list">
    <div class="page-header">
      <h1>扫描项目管理</h1>
      <button @click="showCreateModal = true" class="btn btn-primary">
        <span class="icon">+</span>
        创建项目
      </button>
    </div>

    <div class="filters">
      <input
        v-model="searchTerm"
        placeholder="搜索项目..."
        class="search-input"
      />
    </div>

    <div class="projects-grid" v-if="!loading">
      <div
        v-for="project in filteredProjects"
        :key="project.id"
        class="project-card"
        @click="navigateToProject(project.id)"
      >
        <div class="project-header">
          <h3>{{ project.name }}</h3>
          <div class="project-actions">
            <button @click.stop="editProject(project)" class="btn-icon">
              <span class="icon">✏️</span>
            </button>
            <button @click.stop="deleteProject(project.id)" class="btn-icon danger">
              <span class="icon">🗑️</span>
            </button>
          </div>
        </div>

        <p class="project-description">{{ project.description || '暂无描述' }}</p>

        <div class="project-stats">
          <div class="stat">
            <span class="stat-label">域名</span>
            <span class="stat-value">{{ project.domain_count || 0 }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">IP</span>
            <span class="stat-value">{{ project.ip_count || 0 }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">端口</span>
            <span class="stat-value">{{ project.port_count || 0 }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">漏洞</span>
            <span class="stat-value vulnerability">{{ project.vulnerability_count || 0 }}</span>
          </div>
        </div>

        <div class="project-meta">
          <small>创建时间: {{ formatDate(project.created_at) }}</small>
        </div>
      </div>
    </div>

    <div v-else class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-if="!loading && projects.length === 0" class="empty-state">
      <div class="empty-icon">📂</div>
      <h3>还没有扫描项目</h3>
      <p>创建第一个项目开始安全扫描</p>
      <button @click="showCreateModal = true" class="btn btn-primary">
        创建项目
      </button>
    </div>

    <!-- 创建/编辑项目模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingProject ? '编辑项目' : '创建新项目' }}</h2>
          <button @click="closeModal" class="btn-close">×</button>
        </div>

        <form @submit.prevent="submitProject" class="project-form">
          <div class="form-group">
            <label for="name">项目名称 *</label>
            <input
              id="name"
              v-model="projectForm.name"
              type="text"
              required
              maxlength="100"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label for="description">项目描述</label>
            <textarea
              id="description"
              v-model="projectForm.description"
              rows="3"
              maxlength="500"
              class="form-textarea"
              placeholder="简要描述这个扫描项目的目标和范围..."
            ></textarea>
          </div>

          <div class="form-actions">
            <button type="button" @click="closeModal" class="btn btn-secondary">
              取消
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '保存中...' : (editingProject ? '更新' : '创建') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { scanApi } from '@/api/scanApi'

export default {
  name: 'ProjectList',
  data() {
    return {
      projects: [],
      loading: true,
      searchTerm: '',
      showCreateModal: false,
      editingProject: null,
      submitting: false,
      projectForm: {
        name: '',
        description: ''
      }
    }
  },
  computed: {
    filteredProjects() {
      if (!this.searchTerm) return this.projects
      return this.projects.filter(project =>
        project.name.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
        (project.description && project.description.toLowerCase().includes(this.searchTerm.toLowerCase()))
      )
    }
  },
  async mounted() {
    await this.loadProjects()
  },
  methods: {
    async loadProjects() {
      try {
        this.loading = true
        const response = await scanApi.getProjects()
        this.projects = response.data.data
      } catch (error) {
        console.error('加载项目失败:', error)
        this.$toast.error('加载项目失败')
      } finally {
        this.loading = false
      }
    },

    navigateToProject(projectId) {
      this.$router.push({ name: 'ProjectDetail', params: { id: projectId } })
    },

    editProject(project) {
      this.editingProject = project
      this.projectForm = {
        name: project.name,
        description: project.description || ''
      }
      this.showCreateModal = true
    },

    async deleteProject(projectId) {
      if (!confirm('确定要删除这个项目吗？所有相关的扫描数据都将被删除！')) {
        return
      }

      try {
        await scanApi.deleteProject(projectId)
        this.$toast.success('项目已删除')
        await this.loadProjects()
      } catch (error) {
        console.error('删除项目失败:', error)
        this.$toast.error('删除项目失败')
      }
    },

    async submitProject() {
      try {
        this.submitting = true

        if (this.editingProject) {
          await scanApi.updateProject(this.editingProject.id, this.projectForm)
          this.$toast.success('项目已更新')
        } else {
          await scanApi.createProject(this.projectForm)
          this.$toast.success('项目已创建')
        }

        this.closeModal()
        await this.loadProjects()
      } catch (error) {
        console.error('保存项目失败:', error)
        this.$toast.error('保存项目失败')
      } finally {
        this.submitting = false
      }
    },

    closeModal() {
      this.showCreateModal = false
      this.editingProject = null
      this.projectForm = { name: '', description: '' }
    },

    formatDate(timestamp) {
      return new Date(timestamp * 1000).toLocaleDateString('zh-CN')
    }
  }
}
</script>

<style scoped>
.project-list {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  color: #1a1a1a;
  font-size: 28px;
  font-weight: 600;
}

.filters {
  margin-bottom: 24px;
}

.search-input {
  width: 300px;
  padding: 12px 16px;
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  font-size: 14px;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.project-card {
  background: white;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.project-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.project-header h3 {
  margin: 0;
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 600;
}

.project-actions {
  display: flex;
  gap: 8px;
}

.btn-icon {
  background: none;
  border: none;
  padding: 4px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-icon:hover {
  background-color: #f3f4f6;
}

.btn-icon.danger:hover {
  background-color: #fee2e2;
  color: #dc2626;
}

.project-description {
  color: #6b7280;
  margin-bottom: 16px;
  font-size: 14px;
  line-height: 1.5;
}

.project-stats {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.stat {
  text-align: center;
}

.stat-label {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
}

.stat-value {
  display: block;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-value.vulnerability {
  color: #dc2626;
}

.project-meta {
  color: #9ca3af;
  font-size: 12px;
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

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state h3 {
  color: #1a1a1a;
  margin-bottom: 8px;
}

.empty-state p {
  color: #6b7280;
  margin-bottom: 24px;
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
  max-width: 500px;
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

.modal-header h2 {
  margin: 0;
  color: #1a1a1a;
  font-size: 20px;
  font-weight: 600;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6b7280;
}

.project-form {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #374151;
}

.form-input, .form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-input:focus, .form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.icon {
  font-style: normal;
}
</style>