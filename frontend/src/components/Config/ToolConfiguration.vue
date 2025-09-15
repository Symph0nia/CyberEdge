<template>
  <div class="min-">
    <HeaderPage />

    <div >
      <!-- 配置与工具状态总览 -->
      <div >
        <div
          class="border"
        >
          <div >
            <div
              
            >
              <i class="ri-settings-4-line"></i>
            </div>
            <div>
              <div >总配置数</div>
              <div >{{ toolConfigs.length }}</div>
            </div>
          </div>
          <button
            @click="openCreateModal"
            class=".5 hover: duration-200"
          >
            <i class="ri-add-line"></i>
            新建配置
          </button>
        </div>

        <div
          class="border"
        >
          <div >
            <div
              
            >
              <i class="ri-tools-line"></i>
            </div>
            <div>
              <div >工具状态</div>
              <div >
                {{
                  toolsInfo
                    ? Object.values(toolsInfo.installedStatus).filter(Boolean)
                        .length
                    : "加载中"
                }}/{{
                  toolsInfo
                    ? Object.keys(toolsInfo.installedStatus).length
                    : "..."
                }}
              </div>
            </div>
          </div>
          <button
            @click="fetchToolsStatus"
            class=".5 hover: duration-200 group"
          >
            <i
              class="ri-refresh-line group- duration-500"
            ></i>
            刷新状态
          </button>
        </div>
      </div>

      <!-- 主面板 - 使用标签页分离不同功能 -->
      <div
        class="border overflow-"
      >
        <!-- 标签导航 -->
        <div >
          <button
            @click="activeMainTab = 'configs'"
            class="duration-200"
            :class="activeMainTab === 'configs' ? '' : ' hover:'"
          >
            <i class="ri-settings-4-line"></i>
            工具配置管理
            <div
              v-if="activeMainTab === 'configs'"
              class=".5"
            ></div>
          </button>
          <button
            @click="activeMainTab = 'tools'"
            class="duration-200"
            :class="activeMainTab === 'tools' ? '' : ' hover:'"
          >
            <i class="ri-tools-line"></i>
            工具支持状态
            <div
              v-if="activeMainTab === 'tools'"
              class=".5"
            ></div>
          </button>
        </div>

        <!-- 工具配置管理面板 -->
        <div v-if="activeMainTab === 'configs'" >
          <!-- 搜索和操作区 -->
          <div >
            <div >
              <input
                v-model="configSearchQuery"
                type="text"
                placeholder="搜索配置..."
                class="border"
              />
              <i
                class="ri-search-line .5"
              ></i>
            </div>
            <div >
              <button
                @click="fetchToolConfigs"
                class="hover: duration-200 group"
              >
                <i
                  class="ri-refresh-line group- duration-500"
                ></i>
                刷新配置
              </button>
              <button
                @click="openCreateModal"
                class="hover: duration-200"
              >
                <i class="ri-add-line"></i>
                新建配置
              </button>
            </div>
          </div>

          <!-- 配置卡片列表 -->
          <div
            v-if="filteredConfigs.length > 0"
            class="md: xl:"
          >
            <div
              v-for="config in filteredConfigs"
              :key="config.id"
              class="border overflow- hover: hover: duration-200"
            >
              <div >
                <div>
                  <div >
                    <h3 >
                      {{ config.name }}
                    </h3>
                    <span
                      v-if="config.is_default"
                      class=".5 .5 border"
                    >
                      默认
                    </span>
                  </div>
                  <p >
                    {{ formatDate(config.created_at) }}
                  </p>
                </div>
                <div >
                  <button
                    @click="viewConfigDetails(config)"
                    class=".5 hover: duration-200"
                    title="查看详情"
                  >
                    <i class="ri-eye-line"></i>
                  </button>
                  <button
                    @click="editConfig(config)"
                    class=".5 hover: duration-200"
                    title="编辑"
                  >
                    <i class="ri-edit-line"></i>
                  </button>
                  <button
                    v-if="!config.is_default"
                    @click="setAsDefault(config.id)"
                    class=".5 hover: duration-200"
                    title="设为默认"
                  >
                    <i class="ri-star-line"></i>
                  </button>
                  <button
                    v-if="!config.is_default"
                    @click="confirmDelete(config)"
                    class=".5 hover: duration-200"
                    title="删除"
                  >
                    <i class="ri-delete-bin-line"></i>
                  </button>
                </div>
              </div>

              <!-- 工具启用状态指示器 -->
              <div >
                <div >
                  <span
                    v-for="tool in [
                      'nmap',
                      'ffuf',
                      'subfinder',
                      'httpx',
                      'fscan',
                      'afrog',
                      'nuclei',
                    ]"
                    :key="tool"
                    :class="config[`${tool}_config`].enabled ? ' ' : ' '"
                    class=".5 border"
                  >
                    {{ tool }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 加载中状态 -->
          <div
            v-else-if="loading"
            
          >
            <svg
              class="-"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            加载中...
          </div>

          <!-- 无数据状态 -->
          <div
            v-else-if="toolConfigs.length === 0"
            
          >
            <div
              
            >
              <i class="ri-file-list-3-line"></i>
            </div>
            <p>暂无配置数据</p>
            <button
              @click="openCreateModal"
              class="hover: duration-200"
            >
              创建第一个配置
            </button>
          </div>

          <!-- 搜索无结果 -->
          <div
            v-else
            
          >
            <div
              
            >
              <i class="ri-search-line"></i>
            </div>
            <p>没有找到匹配 "{{ configSearchQuery }}" 的配置</p>
            <button
              @click="configSearchQuery = ''"
              class="hover: duration-200"
            >
              清除搜索
            </button>
          </div>
        </div>

        <!-- 工具支持状态面板 -->
        <div v-if="activeMainTab === 'tools'" >
          <!-- 工具状态网格 -->
          <div
            v-if="toolsInfo"
            class="sm: lg:"
          >
            <div
              v-for="(status, tool) in toolsInfo.installedStatus"
              :key="tool"
              :class="status ? '' : ''"
              class="border duration-200 overflow-"
            >
              <!-- 背景指示器 -->
              <div
                :class="status ? '' : ''"
                class="inset-0"
              ></div>

              <div >
                <div >
                  <div
                    :class="status ? ' ' : ' '"
                    
                  >
                    <i class="ri-terminal-box-line"></i>
                  </div>
                  <div>
                    <h3 >{{ tool }}</h3>
                    <div
                      :class="status ? '' : ''"
                      class=".5"
                    >
                      <i
                        :class="status ? 'ri-checkbox-circle-line' : 'ri-close-circle-line'"
                        
                      ></i>
                      {{ status ? "已安装" : "未安装" }}
                    </div>
                  </div>
                </div>

                <div
                  v-if="
                    toolsInfo.versions && toolsInfo.versions[tool] && status
                  "
                  class=".5"
                >
                  v{{ toolsInfo.versions[tool] }}
                </div>
              </div>
            </div>
          </div>

          <!-- 加载中状态 -->
          <div
            v-else
            
          >
            <svg
              class="-"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            加载中...
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑配置模态框 -->
    <div
      v-if="showConfigModal"
      class="inset-0 duration-300"
      @click="showConfigModal = false"
    >
      <div
        class="border max- max- duration-300"
        @click.stop
      >
        <div
          
        >
          <h3 >
            <i
              :class="isEditing ? 'ri-edit-line' : 'ri-add-line'"
              
            ></i>
            {{ isEditing ? "编辑配置" : "创建新配置" }}
          </h3>
          <button
            @click="showConfigModal = false"
            class="hover: hover: duration-200"
          >
            <i class="ri-close-line"></i>
          </button>
        </div>

        <div >
          <form @submit.prevent="saveConfig">
            <!-- 基础配置 -->
            <div >
              <div
                
              >
                <i class="ri-information-line"></i>
                基本信息
              </div>

              <div class="md:">
                <div>
                  <label 
                    >配置名称</label
                  >
                  <input
                    v-model="currentConfig.name"
                    type="text"
                    class="border"
                    placeholder="输入配置名称"
                    required
                  />
                </div>

                <div >
                  <label
                    class="border"
                  >
                    <input
                      v-model="currentConfig.is_default"
                      type="checkbox"
                      
                    />
                    <span >设为默认配置</span>
                  </label>
                </div>
              </div>
            </div>

            <!-- 工具配置选项卡 -->
            <div >
              <div
                
              >
                <i class="ri-tools-line"></i>
                工具配置
              </div>

              <div
                class="scrollbar-"
              >
                <button
                  v-for="tool in [
                    'nmap',
                    'ffuf',
                    'subfinder',
                    'httpx',
                    'fscan',
                    'afrog',
                    'nuclei',
                  ]"
                  :key="tool"
                  type="button"
                  :class="{ ' ': activeTab === tool, ' hover: hover:': activeTab !== tool, }"
                  class="border duration-200"
                  @click="activeTab = tool"
                >
                  {{ tool.charAt(0).toUpperCase() + tool.slice(1) }}
                </button>
              </div>

              <!-- 工具启用状态 -->
              <div >
                <label
                  class="border"
                >
                  <input
                    v-model="currentConfig[`${activeTab}_config`].enabled"
                    type="checkbox"
                    
                  />
                  <span >
                    启用
                    {{ activeTab.charAt(0).toUpperCase() + activeTab.slice(1) }}
                  </span>
                </label>
              </div>

              <!-- Nmap 配置 -->
              <div v-if="activeTab === 'nmap'" >
                <div>
                  <label 
                    >端口范围</label
                  >
                  <input
                    v-model="currentConfig.nmap_config.ports"
                    type="text"
                    class="border"
                    placeholder="例如: 80,443,8080-8090"
                  />
                  <p >
                    支持单个端口，多个端口（逗号分隔）或端口范围（使用横线）
                  </p>
                </div>

                <div class="md:">
                  <div>
                    <label 
                      >扫描超时（秒）</label
                    >
                    <input
                      v-model.number="currentConfig.nmap_config.scan_timeout"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                  <div>
                    <label 
                      >并发数</label
                    >
                    <input
                      v-model.number="currentConfig.nmap_config.concurrency"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                </div>
              </div>

              <!-- Ffuf 配置 -->
              <div v-if="activeTab === 'ffuf'" >
                <div>
                  <label 
                    >字典文件路径</label
                  >
                  <input
                    v-model="currentConfig.ffuf_config.wordlist_path"
                    type="text"
                    class="border"
                    placeholder="例如: /usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt"
                  />
                </div>

                <div class="md:">
                  <div>
                    <label 
                      >扩展名</label
                    >
                    <input
                      v-model="currentConfig.ffuf_config.extensions"
                      type="text"
                      class="border"
                      placeholder="例如: php,asp,aspx,jsp"
                    />
                  </div>
                  <div>
                    <label 
                      >线程数</label
                    >
                    <input
                      v-model.number="currentConfig.ffuf_config.threads"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                </div>

                <div>
                  <label 
                    >匹配HTTP状态码</label
                  >
                  <input
                    v-model="currentConfig.ffuf_config.match_http_code"
                    type="text"
                    class="border"
                    placeholder="例如: 200,204,301,302,307,401,403"
                  />
                </div>
              </div>

              <!-- Subfinder 配置 -->
              <div v-if="activeTab === 'subfinder'" >
                <div>
                  <label 
                    >配置文件路径</label
                  >
                  <input
                    v-model="currentConfig.subfinder_config.config_path"
                    type="text"
                    class="border"
                    placeholder="例如: /etc/subfinder/config.yaml"
                  />
                </div>

                <div class="md:">
                  <div>
                    <label 
                      >线程数</label
                    >
                    <input
                      v-model.number="currentConfig.subfinder_config.threads"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                  <div>
                    <label 
                      >最大深度</label
                    >
                    <input
                      v-model.number="currentConfig.subfinder_config.max_depth"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                  <div>
                    <label 
                      >超时（秒）</label
                    >
                    <input
                      v-model.number="currentConfig.subfinder_config.timeout"
                      type="number"
                      min="1"
                      class="border"
                    />
                  </div>
                </div>
              </div>

              <!-- 其他工具配置 -->
              <div
                v-if="['httpx', 'fscan', 'afrog', 'nuclei'].includes(activeTab)"
                
              >
                <div>
                  <label 
                    >线程数</label
                  >
                  <input
                    v-model.number="
                      currentConfig[`${activeTab}_config`].threads
                    "
                    type="number"
                    min="1"
                    class="border"
                  />
                </div>
              </div>
            </div>

            <div >
              <button
                type="button"
                @click="showConfigModal = false"
                class="hover: duration-200"
              >
                取消
              </button>
              <button
                type="submit"
                class="hover: duration-200"
              >
                <i
                  :class="isEditing ? 'ri-save-line' : 'ri-add-line'"
                  class=".5"
                ></i>
                {{ isEditing ? "保存更改" : "创建配置" }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div
      v-if="showDeleteConfirm"
      class="inset-0 duration-300"
      @click="showDeleteConfirm = false"
    >
      <div
        class="border max- duration-300"
        @click.stop
      >
        <div >
          <div >
            <div
              
            >
              <i class="ri-delete-bin-line"></i>
            </div>
            <h3 >确认删除</h3>
          </div>
          <p >
            您确定要删除配置
            <span 
              >"{{ configToDelete?.name }}"</span
            >
            吗？此操作无法撤销。
          </p>
        </div>
        <div
          
        >
          <button
            @click="showDeleteConfirm = false"
            class="hover: duration-200"
          >
            取消
          </button>
          <button
            @click="deleteConfig"
            class="hover: duration-200"
          >
            <i class="ri-delete-bin-line .5"></i>
            确认删除
          </button>
        </div>
      </div>
    </div>

    <!-- 查看详情对话框 -->
    <div
      v-if="showViewDetails"
      class="inset-0 duration-300"
      @click="showViewDetails = false"
    >
      <div
        class="border max- max- duration-300"
        @click.stop
      >
        <div
          
        >
          <h3 >
            <i class="ri-eye-line"></i>
            配置详情
          </h3>
          <button
            @click="showViewDetails = false"
            class="hover: hover: duration-200"
          >
            <i class="ri-close-line"></i>
          </button>
        </div>

        <div >
          <div v-if="configToView" >
            <!-- 基本信息 -->
            <div
              class="border"
            >
              <h4
                
              >
                <i class="ri-information-line"></i>
                基本信息
              </h4>
              <div class="md:">
                <div >
                  <div >配置名称</div>
                  <div >
                    {{ configToView.name }}
                  </div>
                </div>
                <div >
                  <div >状态</div>
                  <div>
                    <span
                      :class="configToView.is_default ? ' ' : ' '"
                      class=".5 border"
                    >
                      {{ configToView.is_default ? "默认配置" : "普通配置" }}
                    </span>
                  </div>
                </div>
                <div >
                  <div >创建时间</div>
                  <div >
                    {{ formatDate(configToView.created_at) }}
                  </div>
                </div>
                <div >
                  <div >更新时间</div>
                  <div >
                    {{ formatDate(configToView.updated_at) }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 各工具配置卡片 -->
            <div class="md:">
              <div
                v-for="tool in [
                  'nmap',
                  'ffuf',
                  'subfinder',
                  'httpx',
                  'fscan',
                  'afrog',
                  'nuclei',
                ]"
                :key="tool"
                :class="configToView[`${tool}_config`].enabled ? '' : ''"
                class="border overflow-"
              >
                <!-- 背景状态指示 -->
                <div
                  v-if="configToView[`${tool}_config`].enabled"
                  class="inset-0"
                ></div>

                <h4
                  
                >
                  <span >
                    <i class="ri-settings-line"></i>
                    {{ tool.charAt(0).toUpperCase() + tool.slice(1) }}
                  </span>
                  <span
                    :class="configToView[`${tool}_config`].enabled ? ' ' : ' '"
                    class=".5 border"
                  >
                    {{
                      configToView[`${tool}_config`].enabled
                        ? "已启用"
                        : "已禁用"
                    }}
                  </span>
                </h4>

                <!-- 通用配置项 -->
                <div
                  v-if="configToView[`${tool}_config`].threads"
                  
                >
                  <div >
                    <div >线程数</div>
                    <div >
                      {{ configToView[`${tool}_config`].threads }}
                    </div>
                  </div>
                </div>

                <!-- Nmap特有配置 -->
                <div
                  v-if="tool === 'nmap'"
                  
                >
                  <div >
                    <div >端口范围</div>
                    <div >
                      {{ configToView.nmap_config.ports || "未设置" }}
                    </div>
                  </div>
                  <div >
                    <div >扫描超时</div>
                    <div >
                      {{ configToView.nmap_config.scan_timeout || "0" }} 秒
                    </div>
                  </div>
                  <div >
                    <div >并发数</div>
                    <div >
                      {{ configToView.nmap_config.concurrency || "0" }}
                    </div>
                  </div>
                </div>

                <!-- Ffuf特有配置 -->
                <div
                  v-if="tool === 'ffuf'"
                  
                >
                  <div >
                    <div >字典文件</div>
                    <div >
                      {{ configToView.ffuf_config.wordlist_path || "未设置" }}
                    </div>
                  </div>
                  <div >
                    <div >
                      <div >扩展名</div>
                      <div >
                        {{ configToView.ffuf_config.extensions || "未设置" }}
                      </div>
                    </div>
                    <div >
                      <div >HTTP状态码</div>
                      <div >
                        {{
                          configToView.ffuf_config.match_http_code || "未设置"
                        }}
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Subfinder特有配置 -->
                <div
                  v-if="tool === 'subfinder'"
                  
                >
                  <div >
                    <div >配置文件</div>
                    <div >
                      {{
                        configToView.subfinder_config.config_path || "未设置"
                      }}
                    </div>
                  </div>
                  <div >
                    <div >
                      <div >最大深度</div>
                      <div >
                        {{ configToView.subfinder_config.max_depth || "0" }}
                      </div>
                    </div>
                    <div >
                      <div >超时</div>
                      <div >
                        {{ configToView.subfinder_config.timeout || "0" }} 秒
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          
        >
          <button
            @click="editConfig(configToView)"
            class="hover: duration-200"
          >
            <i class="ri-edit-line .5"></i>
            编辑此配置
          </button>
          <button
            @click="showViewDetails = false"
            class="hover: duration-200"
          >
            关闭
          </button>
        </div>
      </div>
    </div>

    <FooterPage />

    <PopupNotification
      v-if="showNotification"
      :message="notificationMessage"
      :type="notificationType"
      @close="showNotification = false"
    />
  </div>
</template>

<script>
import { ref, computed, onMounted } from "vue";
import api from "../../api/axiosInstance";
import HeaderPage from "../HeaderPage.vue";
import FooterPage from "../FooterPage.vue";
import PopupNotification from "../Utils/PopupNotification.vue";
import { useNotification } from "../../composables/useNotification";

export default {
  name: "ToolConfig",
  components: {
    HeaderPage,
    FooterPage,
    PopupNotification,
  },
  setup() {
    const toolConfigs = ref([]);
    const toolsInfo = ref(null);
    const loading = ref(false);
    const showConfigModal = ref(false);
    const showDeleteConfirm = ref(false);
    const showViewDetails = ref(false);
    const configToDelete = ref(null);
    const configToView = ref(null);
    const isEditing = ref(false);
    const activeTab = ref("nmap");
    const activeMainTab = ref("configs");
    const configSearchQuery = ref("");

    // 筛选后的配置列表
    const filteredConfigs = computed(() => {
      if (!configSearchQuery.value) return toolConfigs.value;

      const query = configSearchQuery.value.toLowerCase();
      return toolConfigs.value.filter((config) => {
        return config.name.toLowerCase().includes(query);
      });
    });

    const {
      showNotification,
      notificationMessage,
      notificationType,
      showSuccess,
      showError,
    } = useNotification();

    // 默认配置模板
    const defaultConfig = {
      name: "",
      is_default: false,
      nmap_config: {
        enabled: true,
        ports:
          "21,22,23,25,53,80,110,111,135,139,143,443,445,465,587,993,995,1080,1433,1521,3306,3389,5432,5900,6379,8080,8443",
        scan_timeout: 300,
        concurrency: 100,
      },
      ffuf_config: {
        enabled: true,
        wordlist_path:
          "/usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt",
        extensions: "php,asp,aspx,jsp,html,js",
        threads: 50,
        match_http_code: "200,204,301,302,307,401,403",
      },
      subfinder_config: {
        enabled: true,
        config_path: "/etc/subfinder/config.yaml",
        threads: 10,
        max_depth: 2,
        timeout: 60,
      },
      httpx_config: {
        enabled: true,
        threads: 50,
        timeout: 10,
      },
      fscan_config: {
        enabled: true,
        threads: 100,
      },
      afrog_config: {
        enabled: true,
        threads: 50,
      },
      nuclei_config: {
        enabled: true,
        threads: 50,
      },
    };

    const currentConfig = ref({ ...defaultConfig });

    const fetchToolConfigs = async () => {
      loading.value = true;
      try {
        const response = await api.get("/tools/configs");
        if (response.data?.status === "success") {
          toolConfigs.value = response.data.data || [];
          showSuccess("工具配置列表已更新");
        }
      } catch (error) {
        showError(error.response?.data?.message || "获取工具配置失败");
      } finally {
        loading.value = false;
      }
    };

    const fetchToolsStatus = async () => {
      try {
        const response = await api.get("/system/tools");
        if (
          response.data?.status === "success" &&
          response.data?.data?.toolsInfo
        ) {
          toolsInfo.value = response.data.data.toolsInfo;
          showSuccess("工具状态已更新");
        }
      } catch (error) {
        showError(error.response?.data?.message || "获取工具状态失败");
      }
    };

    const openCreateModal = () => {
      isEditing.value = false;
      currentConfig.value = { ...defaultConfig };
      showConfigModal.value = true;
      activeTab.value = "nmap";
    };

    const editConfig = (config) => {
      isEditing.value = true;
      currentConfig.value = JSON.parse(JSON.stringify(config)); // 深拷贝
      showConfigModal.value = true;
      activeTab.value = "nmap";

      // 如果在查看详情时编辑，需要关闭详情窗口
      if (showViewDetails.value) {
        showViewDetails.value = false;
      }
    };

    const saveConfig = async () => {
      try {
        if (isEditing.value) {
          await api.put(
            `/tools/configs/${currentConfig.value.id}`,
            currentConfig.value
          );
          showSuccess("配置已更新");
        } else {
          await api.post("/tools/configs", currentConfig.value);
          showSuccess("配置已创建");
        }
        showConfigModal.value = false;
        fetchToolConfigs();
      } catch (error) {
        showError(error.response?.data?.message || "保存配置失败");
      }
    };

    const confirmDelete = (config) => {
      configToDelete.value = config;
      showDeleteConfirm.value = true;
    };

    const deleteConfig = async () => {
      try {
        await api.delete(`/tools/configs/${configToDelete.value.id}`);
        showSuccess("配置已删除");
        showDeleteConfirm.value = false;
        fetchToolConfigs();
      } catch (error) {
        showError(error.response?.data?.message || "删除配置失败");
      }
    };

    const setAsDefault = async (id) => {
      try {
        await api.put(`/tools/configs/${id}/default`);
        showSuccess("默认配置已设置");
        fetchToolConfigs();
      } catch (error) {
        showError(error.response?.data?.message || "设置默认配置失败");
      }
    };

    const viewConfigDetails = (config) => {
      configToView.value = JSON.parse(JSON.stringify(config)); // 深拷贝
      showViewDetails.value = true;
    };

    const formatDate = (dateString) => {
      if (!dateString) return "未知时间";
      const date = new Date(dateString);
      return new Intl.DateTimeFormat("zh-CN", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
      }).format(date);
    };

    onMounted(() => {
      fetchToolConfigs();
      fetchToolsStatus();
    });

    return {
      toolConfigs,
      filteredConfigs,
      toolsInfo,
      loading,
      showConfigModal,
      showDeleteConfirm,
      showViewDetails,
      currentConfig,
      configToDelete,
      configToView,
      isEditing,
      activeTab,
      activeMainTab,
      configSearchQuery,
      fetchToolConfigs,
      fetchToolsStatus,
      openCreateModal,
      editConfig,
      saveConfig,
      confirmDelete,
      deleteConfig,
      setAsDefault,
      viewConfigDetails,
      formatDate,
      showNotification,
      notificationMessage,
      notificationType,
    };
  },
};
</script>

<style scoped>
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

/* 优化按钮点击效果 */
button:active {
  transform: scale(0.98);
}

/* 隐藏滚动条但保留功能 */
.scrollbar-hidden {
  -ms-overflow-style: none; /* IE and Edge */
  scrollbar-width: none; /* Firefox */
}

.scrollbar-hidden::-webkit-scrollbar {
  display: none; /* Chrome, Safari, Opera */
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
</style>
