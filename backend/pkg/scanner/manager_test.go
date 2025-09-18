package scanner

import (
	"context"
	"testing"
)

// MockScanner 测试用的模拟扫描器
type MockScanner struct {
	name      string
	category  ScanCategory
	available bool
	scanFunc  func(ctx context.Context, config ScanConfig) (*ScanResult, error)
}

func (m *MockScanner) GetName() string {
	return m.name
}

func (m *MockScanner) GetCategory() ScanCategory {
	return m.category
}

func (m *MockScanner) IsAvailable() bool {
	return m.available
}

func (m *MockScanner) Scan(ctx context.Context, config ScanConfig) (*ScanResult, error) {
	if m.scanFunc != nil {
		return m.scanFunc(ctx, config)
	}
	return &ScanResult{
		ScannerName: m.name,
		Category:    m.category,
		Target:      config.Target,
		Status:      StatusCompleted,
		Data:        SubdomainData{Subdomains: []SubdomainInfo{}},
	}, nil
}

func (m *MockScanner) ValidateConfig(config ScanConfig) error {
	if config.Target == "" {
		return ErrInvalidTarget
	}
	if config.ProjectID == 0 {
		return ErrInvalidProjectID
	}
	return nil
}

func TestNewScanManager(t *testing.T) {
	manager := NewScanManager()
	if manager == nil {
		t.Error("Expected manager, got nil")
	}
}

func TestScanManager_RegisterScanner(t *testing.T) {
	manager := NewScanManager()

	tests := []struct {
		name      string
		scanner   Scanner
		expectErr bool
	}{
		{
			name: "Valid scanner",
			scanner: &MockScanner{
				name:      "test-scanner",
				category:  CategorySubdomain,
				available: true,
			},
			expectErr: false,
		},
		{
			name:      "Nil scanner",
			scanner:   nil,
			expectErr: true,
		},
		{
			name: "Empty name scanner",
			scanner: &MockScanner{
				name:      "",
				category:  CategorySubdomain,
				available: true,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RegisterScanner(tt.scanner)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestScanManager_RegisterScanner_Duplicate(t *testing.T) {
	manager := NewScanManager()

	scanner1 := &MockScanner{
		name:      "duplicate-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	scanner2 := &MockScanner{
		name:      "duplicate-scanner",
		category:  CategoryPort,
		available: true,
	}

	// First registration should succeed
	err := manager.RegisterScanner(scanner1)
	if err != nil {
		t.Errorf("First registration failed: %v", err)
	}

	// Second registration should fail
	err = manager.RegisterScanner(scanner2)
	if err == nil {
		t.Error("Expected error for duplicate scanner registration")
	}
}

func TestScanManager_GetScanner(t *testing.T) {
	manager := NewScanManager()

	availableScanner := &MockScanner{
		name:      "available-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	unavailableScanner := &MockScanner{
		name:      "unavailable-scanner",
		category:  CategoryPort,
		available: false,
	}

	manager.RegisterScanner(availableScanner)
	manager.RegisterScanner(unavailableScanner)

	tests := []struct {
		name        string
		scannerName string
		expectErr   bool
	}{
		{
			name:        "Get available scanner",
			scannerName: "available-scanner",
			expectErr:   false,
		},
		{
			name:        "Get unavailable scanner",
			scannerName: "unavailable-scanner",
			expectErr:   true,
		},
		{
			name:        "Get non-existent scanner",
			scannerName: "non-existent",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := manager.GetScanner(tt.scannerName)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && scanner == nil {
				t.Error("Expected scanner, got nil")
			}
		})
	}
}

func TestScanManager_ListScanners(t *testing.T) {
	manager := NewScanManager()

	scanner1 := &MockScanner{
		name:      "scanner1",
		category:  CategorySubdomain,
		available: true,
	}

	scanner2 := &MockScanner{
		name:      "scanner2",
		category:  CategoryPort,
		available: true,
	}

	scanner3 := &MockScanner{
		name:      "scanner3",
		category:  CategorySubdomain,
		available: false,
	}

	manager.RegisterScanner(scanner1)
	manager.RegisterScanner(scanner2)
	manager.RegisterScanner(scanner3)

	scanners := manager.ListScanners()

	// Should only return available scanners
	if len(scanners) != 2 {
		t.Errorf("Expected 2 available scanners, got %d", len(scanners))
	}

	names := make(map[string]bool)
	for _, scanner := range scanners {
		names[scanner.GetName()] = true
	}

	if !names["scanner1"] || !names["scanner2"] {
		t.Error("Expected scanner1 and scanner2 in available list")
	}

	if names["scanner3"] {
		t.Error("scanner3 should not be in available list")
	}
}

func TestScanManager_ListByCategory(t *testing.T) {
	manager := NewScanManager()

	subdomainScanner := &MockScanner{
		name:      "subdomain-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	portScanner := &MockScanner{
		name:      "port-scanner",
		category:  CategoryPort,
		available: true,
	}

	manager.RegisterScanner(subdomainScanner)
	manager.RegisterScanner(portScanner)

	subdomainScanners := manager.ListByCategory(CategorySubdomain)
	if len(subdomainScanners) != 1 {
		t.Errorf("Expected 1 subdomain scanner, got %d", len(subdomainScanners))
	}

	if subdomainScanners[0].GetName() != "subdomain-scanner" {
		t.Errorf("Expected 'subdomain-scanner', got '%s'", subdomainScanners[0].GetName())
	}

	portScanners := manager.ListByCategory(CategoryPort)
	if len(portScanners) != 1 {
		t.Errorf("Expected 1 port scanner, got %d", len(portScanners))
	}

	vulnScanners := manager.ListByCategory(CategoryVulnerability)
	if len(vulnScanners) != 0 {
		t.Errorf("Expected 0 vulnerability scanners, got %d", len(vulnScanners))
	}
}

func TestScanManager_ExecuteScan(t *testing.T) {
	manager := NewScanManager()

	mockScanner := &MockScanner{
		name:      "test-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	manager.RegisterScanner(mockScanner)

	tests := []struct {
		name      string
		config    ScanConfig
		expectErr bool
	}{
		{
			name: "Valid scan with tool specified",
			config: ScanConfig{
				ProjectID: 1,
				Target:    "example.com",
				Options: map[string]string{
					"tool": "test-scanner",
				},
			},
			expectErr: false,
		},
		{
			name: "No tool specified",
			config: ScanConfig{
				ProjectID: 1,
				Target:    "example.com",
				Options:   map[string]string{},
			},
			expectErr: true,
		},
		{
			name: "Non-existent tool",
			config: ScanConfig{
				ProjectID: 1,
				Target:    "example.com",
				Options: map[string]string{
					"tool": "non-existent",
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid config",
			config: ScanConfig{
				ProjectID: 0,
				Target:    "",
				Options: map[string]string{
					"tool": "test-scanner",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := manager.ExecuteScan(ctx, tt.config)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr {
				if result == nil {
					t.Error("Expected result, got nil")
				} else {
					if result.Status != StatusCompleted {
						t.Errorf("Expected status %s, got %s", StatusCompleted, result.Status)
					}
					if result.ScannerName != "test-scanner" {
						t.Errorf("Expected scanner name 'test-scanner', got '%s'", result.ScannerName)
					}
				}
			}
		})
	}
}

func TestScanManager_ExecuteScan_ScannerError(t *testing.T) {
	manager := NewScanManager()

	mockScanner := &MockScanner{
		name:      "error-scanner",
		category:  CategorySubdomain,
		available: true,
		scanFunc: func(ctx context.Context, config ScanConfig) (*ScanResult, error) {
			return nil, ErrScanFailed
		},
	}

	manager.RegisterScanner(mockScanner)

	config := ScanConfig{
		ProjectID: 1,
		Target:    "example.com",
		Options: map[string]string{
			"tool": "error-scanner",
		},
	}

	ctx := context.Background()
	result, err := manager.ExecuteScan(ctx, config)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result == nil {
		t.Error("Expected result even on error, got nil")
		return
	}

	if result.Status != StatusFailed {
		t.Errorf("Expected status %s, got %s", StatusFailed, result.Status)
	}

	if result.Error == "" {
		t.Error("Expected error message in result")
	}
}

func TestScanManager_ExecutePipeline_Simple(t *testing.T) {
	manager := NewScanManager()

	mockScanner := &MockScanner{
		name:      "pipeline-scanner",
		category:  CategorySubdomain,
		available: true,
	}

	manager.RegisterScanner(mockScanner)

	pipeline := ScanPipeline{
		Name:            "test-pipeline",
		ProjectID:       1,
		Target:          "example.com",
		ContinueOnError: false,
		Stages: []ScanStage{
			{
				Name:         "stage1",
				ScannerNames: []string{"pipeline-scanner"},
				Parallel:     false,
			},
		},
	}

	ctx := context.Background()
	results, err := manager.ExecutePipeline(ctx, pipeline)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, results[0].Status)
	}
}