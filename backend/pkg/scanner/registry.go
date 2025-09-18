package scanner

// ScannerRegistry 扫描工具注册表
type ScannerRegistry struct {
	manager ScanManager
}

// NewScannerRegistry 创建扫描工具注册表
func NewScannerRegistry() *ScannerRegistry {
	return &ScannerRegistry{
		manager: NewScanManager(),
	}
}

// GetManager 获取扫描管理器
func (r *ScannerRegistry) GetManager() ScanManager {
	return r.manager
}

// RegisterScanner 注册单个扫描工具
func (r *ScannerRegistry) RegisterScanner(scanner Scanner) error {
	return r.manager.RegisterScanner(scanner)
}

// GetAvailableTools 获取所有可用工具信息
func (r *ScannerRegistry) GetAvailableTools() map[ScanCategory][]ScannerInfo {
	result := make(map[ScanCategory][]ScannerInfo)

	scanners := r.manager.ListScanners()
	for _, scanner := range scanners {
		category := scanner.GetCategory()
		info := ScannerInfo{
			Name:      scanner.GetName(),
			Category:  category,
			Available: scanner.IsAvailable(),
		}
		result[category] = append(result[category], info)
	}

	return result
}

// ScannerInfo 扫描工具信息
type ScannerInfo struct {
	Name      string       `json:"name"`
	Category  ScanCategory `json:"category"`
	Available bool         `json:"available"`
}

// DefaultPipelines 默认扫描流水线配置
var DefaultPipelines = map[string]ScanPipeline{
	"comprehensive": {
		Name:            "综合扫描",
		Parallel:        false,
		ContinueOnError: true,
		Stages: []ScanStage{
			{
				Name:         "subdomain_discovery",
				ScannerNames: []string{"subfinder"},
				Parallel:     false,
				Options: map[string]string{
					"recursive": "true",
					"sources":   "all",
				},
			},
			{
				Name:         "web_detection",
				ScannerNames: []string{"httpx"},
				Parallel:     false,
				DependsOn:    []string{"subfinder"},
				Options: map[string]string{
					"detect_tech":    "true",
					"extract_title":  "true",
					"show_status":    "true",
					"show_length":    "true",
					"follow_redirect": "true",
				},
			},
			{
				Name:         "port_scan",
				ScannerNames: []string{"nmap"},
				Parallel:     true,
				DependsOn:    []string{"subfinder"},
				Options: map[string]string{
					"scan_type":      "syn",
					"detect_service": "true",
					"ports":          "1-10000",
				},
			},
			{
				Name:         "vulnerability_scan",
				ScannerNames: []string{"nuclei"},
				Parallel:     false,
				DependsOn:    []string{"httpx"},
				Options: map[string]string{
					"severity":        "high,critical",
					"follow_redirect": "true",
				},
			},
			{
				Name:         "directory_scan",
				ScannerNames: []string{"gobuster"},
				Parallel:     false,
				DependsOn:    []string{"httpx"},
				Options: map[string]string{
					"extensions":     "php,html,js,txt,xml,json",
					"status_codes":   "200,204,301,302,307,403",
					"recursive":      "false",
				},
			},
		},
	},
	"quick": {
		Name:            "快速扫描",
		Parallel:        true,
		ContinueOnError: true,
		Stages: []ScanStage{
			{
				Name:         "quick_discovery",
				ScannerNames: []string{"subfinder", "httpx"},
				Parallel:     true,
				Options: map[string]string{
					"timeout": "30s",
				},
			},
			{
				Name:         "critical_vuln_scan",
				ScannerNames: []string{"nuclei"},
				Parallel:     false,
				DependsOn:    []string{"httpx"},
				Options: map[string]string{
					"severity": "critical",
					"tags":     "rce,sqli,xss",
				},
			},
		},
	},
	"deep": {
		Name:            "深度扫描",
		Parallel:        false,
		ContinueOnError: false,
		Stages: []ScanStage{
			{
				Name:         "extensive_subdomain",
				ScannerNames: []string{"subfinder"},
				Parallel:     false,
				Options: map[string]string{
					"recursive": "true",
					"sources":   "all",
					"timeout":   "300s",
				},
			},
			{
				Name:         "full_port_scan",
				ScannerNames: []string{"nmap"},
				Parallel:     false,
				DependsOn:    []string{"subfinder"},
				Options: map[string]string{
					"scan_type":      "syn",
					"detect_service": "true",
					"detect_os":      "true",
					"ports":          "1-65535",
					"scripts":        "vuln,default",
				},
			},
			{
				Name:         "comprehensive_web",
				ScannerNames: []string{"httpx"},
				Parallel:     false,
				DependsOn:    []string{"subfinder"},
				Options: map[string]string{
					"detect_tech":    "true",
					"extract_title":  "true",
					"show_status":    "true",
					"show_server":    "true",
					"show_time":      "true",
					"follow_redirect": "true",
				},
			},
			{
				Name:         "extensive_directory",
				ScannerNames: []string{"gobuster"},
				Parallel:     false,
				DependsOn:    []string{"httpx"},
				Options: map[string]string{
					"extensions":     "php,html,js,txt,xml,json,bak,backup,old,tmp",
					"status_codes":   "200,204,301,302,307,403,500",
					"recursive":      "true",
					"depth":          "3",
				},
			},
			{
				Name:         "comprehensive_vuln",
				ScannerNames: []string{"nuclei"},
				Parallel:     false,
				DependsOn:    []string{"httpx", "gobuster"},
				Options: map[string]string{
					"severity":        "low,medium,high,critical",
					"follow_redirect": "true",
				},
			},
		},
	},
}

// GetPipeline 获取预定义流水线
func GetPipeline(name string) (ScanPipeline, bool) {
	pipeline, exists := DefaultPipelines[name]
	return pipeline, exists
}

// ListPipelines 列出所有可用流水线
func ListPipelines() []string {
	var names []string
	for name := range DefaultPipelines {
		names = append(names, name)
	}
	return names
}