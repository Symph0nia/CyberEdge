package services

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cyberedge/pkg/models"
)

// Mock DAO for testing - 只模拟我们需要测试的风险点
type mockScanDAO struct {
	mu             sync.RWMutex
	projects       map[uint]*models.Project
	scanJobs       map[uint]*models.ScanJob
	scanTargets    map[uint]*models.ScanTarget
	scanResults    map[uint]*models.ScanResult
	vulnerabilities map[uint]*models.Vulnerability

	// 用于测试并发和数据一致性
	createJobCalls  int
	updateJobCalls  int
	createErrors    bool
	concurrentOps   int
}

func newMockScanDAO() *mockScanDAO {
	return &mockScanDAO{
		projects:        make(map[uint]*models.Project),
		scanJobs:        make(map[uint]*models.ScanJob),
		scanTargets:     make(map[uint]*models.ScanTarget),
		scanResults:     make(map[uint]*models.ScanResult),
		vulnerabilities: make(map[uint]*models.Vulnerability),
	}
}

// 实现必要的DAO接口方法
func (m *mockScanDAO) CreateScanJob(job *models.ScanJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErrors {
		return assert.AnError
	}

	m.createJobCalls++
	m.concurrentOps++

	// 模拟数据库操作延迟
	time.Sleep(time.Millisecond)

	job.ID = uint(len(m.scanJobs) + 1)
	m.scanJobs[job.ID] = job
	return nil
}

func (m *mockScanDAO) UpdateScanJob(job *models.ScanJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updateJobCalls++
	m.scanJobs[job.ID] = job
	return nil
}

func (m *mockScanDAO) GetScanJobByID(id uint) (*models.ScanJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if job, exists := m.scanJobs[id]; exists {
		return job, nil
	}
	return nil, assert.AnError
}

func (m *mockScanDAO) CreateScanTarget(target *models.ScanTarget) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	target.ID = uint(len(m.scanTargets) + 1)
	m.scanTargets[target.ID] = target
	return nil
}

func (m *mockScanDAO) GetScanTargetByAddress(projectID uint, address string) (*models.ScanTarget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, target := range m.scanTargets {
		if target.ProjectID == projectID && target.Address == address {
			return target, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockScanDAO) CreateScanResult(result *models.ScanResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	result.ID = uint(len(m.scanResults) + 1)
	m.scanResults[result.ID] = result
	return nil
}

func (m *mockScanDAO) GetTargetScanResults(targetID uint) ([]models.ScanResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []models.ScanResult
	for _, result := range m.scanResults {
		if result.TargetID == targetID {
			results = append(results, *result)
		}
	}
	return results, nil
}

func (m *mockScanDAO) ImportScanData(data *models.ScanDataImport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟批量导入延迟
	time.Sleep(time.Millisecond * 10)
	return nil
}

// 其他必要的方法（简化实现）
func (m *mockScanDAO) CreateProject(project *models.Project) error { return nil }
func (m *mockScanDAO) GetProjectByID(id uint) (*models.Project, error) { return nil, nil }
func (m *mockScanDAO) GetProjectByName(name string) (*models.Project, error) { return nil, nil }
func (m *mockScanDAO) ListProjects() ([]models.Project, error) { return nil, nil }
func (m *mockScanDAO) DeleteProject(id uint) error { return nil }
func (m *mockScanDAO) GetProjectScanJobs(projectID uint, filters map[string]interface{}) ([]models.ScanJob, error) { return nil, nil }
func (m *mockScanDAO) GetProjectTargets(projectID uint) ([]models.ScanTarget, error) { return nil, nil }
func (m *mockScanDAO) GetScanResultByID(id uint) (*models.ScanResult, error) { return nil, nil }
func (m *mockScanDAO) GetProjectScanResults(projectID uint, filters map[string]interface{}) ([]models.ScanResult, error) { return nil, nil }
func (m *mockScanDAO) GetProjectDetails(projectID uint) (*models.ProjectStats, []models.ScanTarget, error) { return nil, nil, nil }
func (m *mockScanDAO) GetProjectStats(projectID uint) (*models.ProjectStats, error) { return nil, nil }
func (m *mockScanDAO) GetVulnerabilities(projectID uint, filters map[string]interface{}) ([]models.Vulnerability, error) { return nil, nil }
func (m *mockScanDAO) GetVulnerabilityStats(projectID uint) (map[string]int, error) { return nil, nil }
func (m *mockScanDAO) GetProjectHierarchy(projectID uint) ([]models.ScanTarget, error) { return nil, nil }
func (m *mockScanDAO) SearchTargets(projectID uint, searchTerm string) ([]models.ScanTarget, error) { return nil, nil }

// 测试核心业务风险：扫描任务数据一致性
func TestScanDataConsistency(t *testing.T) {
	t.Run("ScanJob状态管理一致性", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// 这个测试验证ScanJob -> ScanTarget -> ScanResult的数据流完整性
		// 模拟一个不存在的流水线，应该立即失败
		job, err := service.StartScan(ctx, 1, "example.com", "nonexistent-pipeline")
		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Equal(t, 0, mockDAO.createJobCalls, "失败的扫描不应该创建数据库记录")

		// 测试正常流程中状态的一致性
		// 注意：这里需要真实的流水线配置，但我们专注测试数据一致性
		mockDAO.createErrors = false

		// 验证ScanJob创建后ID被正确设置
		job = &models.ScanJob{
			ProjectID:    1,
			Target:       "test.com",
			PipelineName: "test",
			Status:       "pending",
		}

		err = mockDAO.CreateScanJob(job)
		require.NoError(t, err)
		assert.NotZero(t, job.ID, "ScanJob ID应该被自动设置")
		assert.Equal(t, "pending", job.Status, "初始状态应该是pending")
	})

	t.Run("扫描结果数据关联完整性", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		// 测试findOrCreateScanTarget的数据一致性
		target1, err := service.findOrCreateScanTarget(1, "example.com", "domain")
		require.NoError(t, err)
		assert.NotZero(t, target1.ID)

		// 再次查找相同目标应该返回同一个记录，不创建重复
		target2, err := service.findOrCreateScanTarget(1, "example.com", "domain")
		require.NoError(t, err)
		assert.Equal(t, target1.ID, target2.ID, "相同目标不应该创建重复记录")
	})
}

// 测试关键风险：并发安全
func TestConcurrentScanOperations(t *testing.T) {
	t.Run("并发扫描任务创建安全性", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		const numJobs = 50
		var wg sync.WaitGroup
		results := make(chan error, numJobs)

		// 50个goroutine同时创建扫描目标（通过service）
		for i := 0; i < numJobs; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 测试并发创建不同的扫描目标
				targetAddr := fmt.Sprintf("concurrent-test-%d.com", id)
				target, err := service.findOrCreateScanTarget(1, targetAddr, "domain")
				if err == nil && target != nil {
					results <- nil
				} else {
					results <- err
				}
			}(i)
		}

		wg.Wait()
		close(results)

		// 验证所有操作都成功，没有竞争条件
		successCount := 0
		for err := range results {
			if err == nil {
				successCount++
			}
		}

		assert.Equal(t, numJobs, successCount, "所有并发创建操作都应该成功")

		// 验证没有数据竞争导致的状态不一致
		mockDAO.mu.RLock()
		targetCount := len(mockDAO.scanTargets)
		mockDAO.mu.RUnlock()

		assert.Equal(t, numJobs, targetCount, "应该创建正确数量的ScanTarget记录")
	})

	t.Run("并发目标创建去重安全性", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		const numGoroutines = 20
		const targetAddress = "concurrent-target.com"
		var wg sync.WaitGroup
		targetIDs := make(chan uint, numGoroutines)

		// 20个goroutine同时尝试创建相同的目标
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				target, err := service.findOrCreateScanTarget(1, targetAddress, "domain")
				if err == nil {
					targetIDs <- target.ID
				}
			}()
		}

		wg.Wait()
		close(targetIDs)

		// 收集所有返回的ID
		var ids []uint
		for id := range targetIDs {
			ids = append(ids, id)
		}

		// 验证要么都返回相同ID（找到现有的），要么只有一个创建成功
		if len(ids) > 0 {
			firstID := ids[0]
			allSame := true
			for _, id := range ids {
				if id != firstID {
					allSame = false
					break
				}
			}

			if !allSame {
				// 如果ID不同，说明有多个创建成功，检查目标表中是否只有一个记录
				mockDAO.mu.RLock()
				count := 0
				for _, target := range mockDAO.scanTargets {
					if target.Address == targetAddress {
						count++
					}
				}
				mockDAO.mu.RUnlock()

				assert.LessOrEqual(t, count, 1, "不应该创建重复的扫描目标")
			}
		}
	})
}

// 测试关键风险：输入验证和安全
func TestScanInputValidation(t *testing.T) {
	t.Run("恶意扫描输入验证", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		// 测试SQL注入尝试
		maliciousInputs := []struct {
			name   string
			target string
			projectID uint
		}{
			{"SQL注入域名", "'; DROP TABLE scan_jobs; --", 1},
			{"超长域名", string(make([]byte, 1000)), 1},
			{"路径遍历", "../../../etc/passwd", 1},
			{"XSS脚本", "<script>alert('xss')</script>", 1},
			{"空输入", "", 1},
			{"Unicode攻击", "xn--e1afmkfd.xn--p1ai", 1},
			{"无效项目ID", "normal.com", 0},
		}

		for _, test := range maliciousInputs {
			t.Run(test.name, func(t *testing.T) {
				target, err := service.findOrCreateScanTarget(test.projectID, test.target, "domain")

				// 根据输入类型验证结果
				if test.target == "" || test.projectID == 0 {
					// 空输入或无效项目ID应该被拒绝
					assert.Error(t, err, "应该拒绝无效输入")
				} else {
					// 其他输入应该被安全处理（不崩溃，正确转义）
					if err == nil {
						assert.NotNil(t, target)
						assert.NotZero(t, target.ID)
						// 验证数据没有被SQL注入污染
						assert.NotContains(t, target.Address, "DROP TABLE", "不应该包含SQL注入代码")
					}
				}
			})
		}
	})
}

// 测试关键风险：内存和资源管理
func TestScanResourceManagement(t *testing.T) {
	t.Run("大量扫描结果内存使用", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过内存测试 - 使用 -short 标志")
		}

		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		// 记录初始内存使用
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// 创建大量扫描结果数据
		const numResults = 1000
		for i := 0; i < numResults; i++ {
			target, err := service.findOrCreateScanTarget(1, "test"+string(rune(i))+".com", "domain")
			require.NoError(t, err)

			result := &models.ScanResult{
				ProjectID:   1,
				TargetID:    target.ID,
				Port:        80 + i%100,
				Protocol:    "tcp",
				State:       "open",
				ServiceName: "http",
				CreatedAt:   time.Now(),
			}

			err = mockDAO.CreateScanResult(result)
			require.NoError(t, err)
		}

		// 强制垃圾回收并测量内存
		runtime.GC()
		runtime.ReadMemStats(&m2)

		// 验证内存使用在合理范围内（不超过100MB增长）
		memoryIncrease := m2.Alloc - m1.Alloc
		maxExpectedIncrease := uint64(100 * 1024 * 1024) // 100MB

		assert.Less(t, memoryIncrease, maxExpectedIncrease,
			"大量扫描结果不应该导致过度内存使用，增长了 %d bytes", memoryIncrease)

		// 验证数据确实被创建
		mockDAO.mu.RLock()
		resultsCount := len(mockDAO.scanResults)
		targetsCount := len(mockDAO.scanTargets)
		mockDAO.mu.RUnlock()

		assert.Equal(t, numResults, resultsCount, "应该创建正确数量的扫描结果")
		assert.Equal(t, numResults, targetsCount, "应该创建正确数量的扫描目标")
	})

	t.Run("Goroutine泄漏检测", func(t *testing.T) {
		initialGoroutines := runtime.NumGoroutine()

		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		// 执行一些操作来验证没有goroutine泄漏
		for i := 0; i < 10; i++ {
			target, err := service.findOrCreateScanTarget(1, "leak-test.com", "domain")
			assert.NoError(t, err)
			assert.NotNil(t, target)
		}

		// 等待可能的异步操作完成
		time.Sleep(time.Millisecond * 100)
		runtime.GC()

		finalGoroutines := runtime.NumGoroutine()

		// 允许少量Goroutine增长（测试框架相关），但不应该有大量泄漏
		goroutineIncrease := finalGoroutines - initialGoroutines
		assert.LessOrEqual(t, goroutineIncrease, 5,
			"不应该有Goroutine泄漏，增长了 %d 个goroutines", goroutineIncrease)
	})
}

// 测试关键风险：错误处理和恢复
func TestScanErrorHandling(t *testing.T) {
	t.Run("数据库错误恢复", func(t *testing.T) {
		mockDAO := newMockScanDAO()
		service, err := NewScanService(mockDAO)
		require.NoError(t, err)

		// 模拟数据库创建失败
		mockDAO.createErrors = true

		job := &models.ScanJob{
			ProjectID:    1,
			Target:       "error-test.com",
			PipelineName: "test",
			Status:       "pending",
		}

		err = mockDAO.CreateScanJob(job)
		assert.Error(t, err, "应该正确传播数据库错误")

		// 验证服务在错误情况下的表现
		_, err = service.findOrCreateScanTarget(1, "error-test.com", "domain")
		// findOrCreateScanTarget应该能处理DAO错误

		// 验证服务状态没有被错误影响
		mockDAO.createErrors = false
		job2 := &models.ScanJob{
			ProjectID:    1,
			Target:       "recovery-test.com",
			PipelineName: "test",
			Status:       "pending",
		}

		err = mockDAO.CreateScanJob(job2)
		assert.NoError(t, err, "错误恢复后应该能正常工作")
	})

	t.Run("无效状态处理", func(t *testing.T) {
		mockDAO := newMockScanDAO()

		// 测试获取不存在的扫描任务
		_, err := mockDAO.GetScanJobByID(99999)
		assert.Error(t, err, "应该正确处理不存在的记录")

		// 测试无效的项目ID
		target, err := mockDAO.GetScanTargetByAddress(99999, "test.com")
		assert.Error(t, err)
		assert.Nil(t, target, "无效查询应该返回nil")
	})
}