package scanner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DefaultScanManager 默认扫描管理器实现
type DefaultScanManager struct {
	scanners map[string]Scanner
	mutex    sync.RWMutex
}

// NewScanManager 创建新的扫描管理器
func NewScanManager() ScanManager {
	return &DefaultScanManager{
		scanners: make(map[string]Scanner),
	}
}

// RegisterScanner 注册扫描工具
func (m *DefaultScanManager) RegisterScanner(scanner Scanner) error {
	if scanner == nil {
		return errors.New("scanner 不能为空")
	}

	name := scanner.GetName()
	if name == "" {
		return errors.New("scanner 名称不能为空")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	if _, exists := m.scanners[name]; exists {
		return fmt.Errorf("scanner '%s' 已经注册", name)
	}

	m.scanners[name] = scanner
	return nil
}

// GetScanner 获取指定扫描工具
func (m *DefaultScanManager) GetScanner(name string) (Scanner, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	scanner, exists := m.scanners[name]
	if !exists {
		return nil, fmt.Errorf("scanner '%s' 未找到", name)
	}

	if !scanner.IsAvailable() {
		return nil, fmt.Errorf("scanner '%s' 不可用", name)
	}

	return scanner, nil
}

// ListScanners 列出所有可用扫描工具
func (m *DefaultScanManager) ListScanners() []Scanner {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []Scanner
	for _, scanner := range m.scanners {
		if scanner.IsAvailable() {
			result = append(result, scanner)
		}
	}

	return result
}

// ListByCategory 按类别列出扫描工具
func (m *DefaultScanManager) ListByCategory(category ScanCategory) []Scanner {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []Scanner
	for _, scanner := range m.scanners {
		if scanner.IsAvailable() && scanner.GetCategory() == category {
			result = append(result, scanner)
		}
	}

	return result
}

// ExecuteScan 执行单个扫描任务
func (m *DefaultScanManager) ExecuteScan(ctx context.Context, config ScanConfig) (*ScanResult, error) {
	// 根据配置选择合适的扫描工具
	var selectedScanner Scanner
	var err error

	// 如果指定了工具名称，直接获取
	if toolName, exists := config.Options["tool"]; exists {
		selectedScanner, err = m.GetScanner(toolName)
		if err != nil {
			return nil, fmt.Errorf("获取指定工具失败: %w", err)
		}
	} else {
		// TODO: 基于目标类型自动选择最佳工具
		return nil, errors.New("未指定扫描工具，自动选择功能暂未实现")
	}

	// 验证配置
	if err := selectedScanner.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 执行扫描
	startTime := time.Now()
	result, err := selectedScanner.Scan(ctx, config)
	if err != nil {
		return &ScanResult{
			ScannerName: selectedScanner.GetName(),
			Category:    selectedScanner.GetCategory(),
			Target:      config.Target,
			StartTime:   startTime,
			EndTime:     time.Now(),
			Duration:    time.Since(startTime),
			Status:      StatusFailed,
			Error:       err.Error(),
		}, err
	}

	// 补充元数据
	result.ScannerName = selectedScanner.GetName()
	result.Category = selectedScanner.GetCategory()
	result.Target = config.Target
	result.StartTime = startTime
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)
	result.Status = StatusCompleted

	return result, nil
}

// ExecutePipeline 执行扫描流水线
func (m *DefaultScanManager) ExecutePipeline(ctx context.Context, pipeline ScanPipeline) ([]ScanResult, error) {
	var allResults []ScanResult

	// 构建阶段依赖图（预留用于循环依赖检测）
	_ = m.buildStageGraph(pipeline.Stages)

	// 执行阶段
	for _, stage := range pipeline.Stages {
		// 检查依赖是否满足
		if !m.areDependenciesMet(stage.DependsOn, allResults) {
			return allResults, fmt.Errorf("阶段 '%s' 的依赖条件未满足", stage.Name)
		}

		// 执行当前阶段
		stageResults, err := m.executeStage(ctx, pipeline, stage, allResults)
		if err != nil {
			if !pipeline.ContinueOnError {
				return allResults, fmt.Errorf("阶段 '%s' 执行失败: %w", stage.Name, err)
			}
			// 记录错误但继续执行
			// TODO: 添加错误日志
		}

		allResults = append(allResults, stageResults...)
	}

	return allResults, nil
}

// buildStageGraph 构建阶段依赖图（简化实现）
func (m *DefaultScanManager) buildStageGraph(stages []ScanStage) map[string][]string {
	graph := make(map[string][]string)
	for _, stage := range stages {
		graph[stage.Name] = stage.DependsOn
	}
	return graph
}

// areDependenciesMet 检查依赖是否满足
func (m *DefaultScanManager) areDependenciesMet(dependencies []string, results []ScanResult) bool {
	if len(dependencies) == 0 {
		return true
	}

	completedStages := make(map[string]bool)
	for _, result := range results {
		if result.Status == StatusCompleted {
			completedStages[result.ScannerName] = true
		}
	}

	for _, dep := range dependencies {
		if !completedStages[dep] {
			return false
		}
	}

	return true
}

// executeStage 执行单个阶段
func (m *DefaultScanManager) executeStage(ctx context.Context, pipeline ScanPipeline, stage ScanStage, previousResults []ScanResult) ([]ScanResult, error) {
	if stage.Parallel {
		// 并行执行阶段内的所有工具
		return m.executeStageParallel(ctx, pipeline, stage, previousResults)
	} else {
		// 串行执行阶段内的所有工具
		return m.executeStageSerial(ctx, pipeline, stage, previousResults)
	}
}

// executeStageParallel 并行执行阶段
func (m *DefaultScanManager) executeStageParallel(ctx context.Context, pipeline ScanPipeline, stage ScanStage, previousResults []ScanResult) ([]ScanResult, error) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var stageResults []ScanResult
	var firstError error

	for _, scannerName := range stage.ScannerNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// 构建扫描配置
			config := ScanConfig{
				ProjectID:     pipeline.ProjectID,
				Target:        pipeline.Target,
				Options:       stage.Options,
				Timeout:       30 * time.Minute, // 默认超时
				ParentResults: previousResults,
			}

			// 指定工具名称
			if config.Options == nil {
				config.Options = make(map[string]string)
			}
			config.Options["tool"] = name

			result, err := m.ExecuteScan(ctx, config)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil && firstError == nil {
				firstError = err
			}

			if result != nil {
				stageResults = append(stageResults, *result)
			}
		}(scannerName)
	}

	wg.Wait()
	return stageResults, firstError
}

// executeStageSerial 串行执行阶段
func (m *DefaultScanManager) executeStageSerial(ctx context.Context, pipeline ScanPipeline, stage ScanStage, previousResults []ScanResult) ([]ScanResult, error) {
	var stageResults []ScanResult

	for _, scannerName := range stage.ScannerNames {
		// 构建扫描配置
		config := ScanConfig{
			ProjectID:     pipeline.ProjectID,
			Target:        pipeline.Target,
			Options:       stage.Options,
			Timeout:       30 * time.Minute,
			ParentResults: append(previousResults, stageResults...),
		}

		// 指定工具名称
		if config.Options == nil {
			config.Options = make(map[string]string)
		}
		config.Options["tool"] = scannerName

		result, err := m.ExecuteScan(ctx, config)
		if err != nil {
			return stageResults, err
		}

		if result != nil {
			stageResults = append(stageResults, *result)
		}
	}

	return stageResults, nil
}

// 工具自动选择策略（未来扩展）
func (m *DefaultScanManager) selectBestScanner(category ScanCategory, target string) (Scanner, error) {
	scanners := m.ListByCategory(category)
	if len(scanners) == 0 {
		return nil, fmt.Errorf("没有可用的 %s 类型扫描工具", category)
	}

	// TODO: 实现智能选择逻辑
	// 可以基于目标类型、工具性能、历史成功率等因素选择
	return scanners[0], nil
}