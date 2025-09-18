package scanner

import (
	"context"
	"testing"
	"time"
)

func TestExecutePipeline_MultipleStageDependencies(t *testing.T) {
	manager := NewScanManager()

	// 注册多个模拟扫描器
	subdomainScanner := &MockScanner{
		name:      "subfinder",
		category:  CategorySubdomain,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			return &ScanResult{
				ScannerName: "subfinder",
				Category:    CategorySubdomain,
				Target:      config.Target,
				Status:      StatusCompleted,
				Data: SubdomainData{
					Subdomains: []SubdomainInfo{
						{Domain: config.Target, Subdomain: "api." + config.Target, Source: "subfinder"},
						{Domain: config.Target, Subdomain: "www." + config.Target, Source: "subfinder"},
					},
				},
			}, nil
		},
	}

	webScanner := &MockScanner{
		name:      "httpx",
		category:  CategoryWebTech,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			return &ScanResult{
				ScannerName: "httpx",
				Category:    CategoryWebTech,
				Target:      config.Target,
				Status:      StatusCompleted,
				Data: WebTechData{
					URL:        "https://" + config.Target,
					StatusCode: 200,
					Title:      "Test Site",
				},
			}, nil
		},
	}

	vulnScanner := &MockScanner{
		name:      "nuclei",
		category:  CategoryVulnerability,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			return &ScanResult{
				ScannerName: "nuclei",
				Category:    CategoryVulnerability,
				Target:      config.Target,
				Status:      StatusCompleted,
				Data: VulnerabilityData{
					Vulnerabilities: []VulnerabilityInfo{
						{Target: config.Target, Title: "Test Vulnerability", Severity: "high"},
					},
				},
			}, nil
		},
	}

	manager.RegisterScanner(subdomainScanner)
	manager.RegisterScanner(webScanner)
	manager.RegisterScanner(vulnScanner)

	pipeline := ScanPipeline{
		Name:            "comprehensive-test",
		ProjectID:       1,
		Target:          "example.com",
		ContinueOnError: false,
		Stages: []ScanStage{
			{
				Name:         "subdomain_discovery",
				ScannerNames: []string{"subfinder"},
				Parallel:     false,
			},
			{
				Name:         "web_detection",
				ScannerNames: []string{"httpx"},
				Parallel:     false,
				DependsOn:    []string{"subfinder"},
			},
			{
				Name:         "vulnerability_scan",
				ScannerNames: []string{"nuclei"},
				Parallel:     false,
				DependsOn:    []string{"httpx"},
			},
		},
	}

	ctx := context.Background()
	results, err := manager.ExecutePipeline(ctx, pipeline)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// 验证执行顺序和结果
	scannerNames := make([]string, len(results))
	for i, result := range results {
		scannerNames[i] = result.ScannerName
		if result.Status != StatusCompleted {
			t.Errorf("Scanner %s failed with status %s", result.ScannerName, result.Status)
		}
	}

	// 验证依赖顺序
	subfinderIndex := findIndex(scannerNames, "subfinder")
	httpxIndex := findIndex(scannerNames, "httpx")
	nucleiIndex := findIndex(scannerNames, "nuclei")

	if subfinderIndex == -1 || httpxIndex == -1 || nucleiIndex == -1 {
		t.Error("Missing expected scanners in results")
	}

	if subfinderIndex >= httpxIndex {
		t.Error("subfinder should execute before httpx")
	}

	if httpxIndex >= nucleiIndex {
		t.Error("httpx should execute before nuclei")
	}
}

func TestExecutePipeline_ParallelStage(t *testing.T) {
	manager := NewScanManager()

	scanner1 := &MockScanner{
		name:      "scanner1",
		category:  CategorySubdomain,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			time.Sleep(100 * time.Millisecond) // 模拟扫描时间
			return &ScanResult{
				ScannerName: "scanner1",
				Category:    CategorySubdomain,
				Target:      config.Target,
				Status:      StatusCompleted,
				Data:        SubdomainData{},
			}, nil
		},
	}

	scanner2 := &MockScanner{
		name:      "scanner2",
		category:  CategoryPort,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			time.Sleep(100 * time.Millisecond) // 模拟扫描时间
			return &ScanResult{
				ScannerName: "scanner2",
				Category:    CategoryPort,
				Target:      config.Target,
				Status:      StatusCompleted,
				Data:        PortData{},
			}, nil
		},
	}

	manager.RegisterScanner(scanner1)
	manager.RegisterScanner(scanner2)

	pipeline := ScanPipeline{
		Name:            "parallel-test",
		ProjectID:       1,
		Target:          "example.com",
		ContinueOnError: false,
		Stages: []ScanStage{
			{
				Name:         "parallel_stage",
				ScannerNames: []string{"scanner1", "scanner2"},
				Parallel:     true,
			},
		},
	}

	ctx := context.Background()
	startTime := time.Now()
	results, err := manager.ExecutePipeline(ctx, pipeline)
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// 并行执行应该比串行快
	// 两个任务各100ms，并行应该接近100ms，串行应该接近200ms
	if duration > 150*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v", duration)
	}
}

func TestExecutePipeline_ContinueOnError(t *testing.T) {
	manager := NewScanManager()

	successScanner := &MockScanner{
		name:      "success-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	errorScanner := &MockScanner{
		name:      "error-scanner",
		category:  CategoryPort,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			return nil, ErrScanFailed
		},
	}

	manager.RegisterScanner(successScanner)
	manager.RegisterScanner(errorScanner)

	tests := []struct {
		name            string
		continueOnError bool
		expectError     bool
		expectedResults int
	}{
		{
			name:            "Continue on error",
			continueOnError: true,
			expectError:     false,
			expectedResults: 1, // 只有成功的扫描器会返回结果
		},
		{
			name:            "Stop on error",
			continueOnError: false,
			expectError:     true,
			expectedResults: 1, // 第一个成功，第二个失败时停止
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := ScanPipeline{
				Name:            "error-test",
				ProjectID:       1,
				Target:          "example.com",
				ContinueOnError: tt.continueOnError,
				Stages: []ScanStage{
					{
						Name:         "success_stage",
						ScannerNames: []string{"success-scanner"},
						Parallel:     false,
					},
					{
						Name:         "error_stage",
						ScannerNames: []string{"error-scanner"},
						Parallel:     false,
					},
				},
			}

			ctx := context.Background()
			results, err := manager.ExecutePipeline(ctx, pipeline)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if len(results) != tt.expectedResults {
				t.Errorf("Expected %d results, got %d", tt.expectedResults, len(results))
			}
		})
	}
}

func TestExecutePipeline_UnmetDependency(t *testing.T) {
	manager := NewScanManager()

	scanner := &MockScanner{
		name:      "test-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	manager.RegisterScanner(scanner)

	pipeline := ScanPipeline{
		Name:            "dependency-test",
		ProjectID:       1,
		Target:          "example.com",
		ContinueOnError: false,
		Stages: []ScanStage{
			{
				Name:         "dependent_stage",
				ScannerNames: []string{"test-scanner"},
				Parallel:     false,
				DependsOn:    []string{"non-existent-scanner"},
			},
		},
	}

	ctx := context.Background()
	results, err := manager.ExecutePipeline(ctx, pipeline)

	if err == nil {
		t.Error("Expected error for unmet dependency")
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestExecutePipeline_EmptyPipeline(t *testing.T) {
	manager := NewScanManager()

	pipeline := ScanPipeline{
		Name:      "empty-test",
		ProjectID: 1,
		Target:    "example.com",
		Stages:    []ScanStage{},
	}

	ctx := context.Background()
	results, err := manager.ExecutePipeline(ctx, pipeline)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestAreDependenciesMet(t *testing.T) {
	manager := &DefaultScanManager{}

	tests := []struct {
		name         string
		dependencies []string
		results      []ScanResult
		expected     bool
	}{
		{
			name:         "No dependencies",
			dependencies: []string{},
			results:      []ScanResult{},
			expected:     true,
		},
		{
			name:         "Dependencies met",
			dependencies: []string{"scanner1", "scanner2"},
			results: []ScanResult{
				{ScannerName: "scanner1", Status: StatusCompleted},
				{ScannerName: "scanner2", Status: StatusCompleted},
				{ScannerName: "scanner3", Status: StatusCompleted},
			},
			expected: true,
		},
		{
			name:         "Dependencies not met",
			dependencies: []string{"scanner1", "scanner2"},
			results: []ScanResult{
				{ScannerName: "scanner1", Status: StatusCompleted},
			},
			expected: false,
		},
		{
			name:         "Failed dependency",
			dependencies: []string{"scanner1"},
			results: []ScanResult{
				{ScannerName: "scanner1", Status: StatusFailed},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.areDependenciesMet(tt.dependencies, tt.results)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function to find index of string in slice
func findIndex(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}