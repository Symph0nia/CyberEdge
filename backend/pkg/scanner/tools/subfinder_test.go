package tools

import (
	"context"
	"testing"
	"time"

	"cyberedge/pkg/scanner"
)

func TestSubfinderScanner_GetName(t *testing.T) {
	scanner := NewSubfinderScanner()
	if scanner.GetName() != "subfinder" {
		t.Errorf("Expected name 'subfinder', got '%s'", scanner.GetName())
	}
}

func TestSubfinderScanner_GetCategory(t *testing.T) {
	s := NewSubfinderScanner()
	if s.GetCategory() != scanner.CategorySubdomain {
		t.Errorf("Expected category '%s', got '%s'", scanner.CategorySubdomain, s.GetCategory())
	}
}

func TestSubfinderScanner_ValidateConfig(t *testing.T) {
	s := NewSubfinderScanner().(*SubfinderScanner)

	tests := []struct {
		name      string
		config    scanner.ScanConfig
		expectErr bool
	}{
		{
			name: "Valid domain",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "example.com",
			},
			expectErr: false,
		},
		{
			name: "Invalid domain - no dot",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "invalid",
			},
			expectErr: true,
		},
		{
			name: "Invalid domain - starts with dot",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    ".example.com",
			},
			expectErr: true,
		},
		{
			name: "Invalid domain - ends with dot",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "example.com.",
			},
			expectErr: true,
		},
		{
			name: "Empty target",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateConfig(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestSubfinderScanner_isValidDomain(t *testing.T) {
	s := &SubfinderScanner{
		BaseScannerTool: NewBaseScannerTool("subfinder", scanner.CategorySubdomain, "subfinder"),
	}

	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "Valid domain",
			domain:   "example.com",
			expected: true,
		},
		{
			name:     "Valid subdomain",
			domain:   "sub.example.com",
			expected: true,
		},
		{
			name:     "Domain with numbers",
			domain:   "test123.example.com",
			expected: true,
		},
		{
			name:     "Domain with hyphen",
			domain:   "test-site.example.com",
			expected: true,
		},
		{
			name:     "Empty domain",
			domain:   "",
			expected: false,
		},
		{
			name:     "Too long domain",
			domain:   string(make([]byte, 300)),
			expected: false,
		},
		{
			name:     "No dot",
			domain:   "invalid",
			expected: false,
		},
		{
			name:     "Starts with dot",
			domain:   ".example.com",
			expected: false,
		},
		{
			name:     "Ends with dot",
			domain:   "example.com.",
			expected: false,
		},
		{
			name:     "Invalid characters",
			domain:   "example@.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.isValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for domain '%s'", tt.expected, result, tt.domain)
			}
		})
	}
}

func TestSubfinderScanner_buildArgs(t *testing.T) {
	s := &SubfinderScanner{
		BaseScannerTool: NewBaseScannerTool("subfinder", scanner.CategorySubdomain, "subfinder"),
	}

	tests := []struct {
		name     string
		target   string
		options  map[string]string
		expected []string
	}{
		{
			name:   "Basic args",
			target: "example.com",
			options: map[string]string{},
			expected: []string{
				"-d", "example.com",
				"-silent",
				"-o", "/dev/stdout",
			},
		},
		{
			name:   "With sources",
			target: "example.com",
			options: map[string]string{
				"sources": "crtsh,virustotal",
			},
			expected: []string{
				"-d", "example.com",
				"-silent",
				"-o", "/dev/stdout",
				"-sources", "crtsh,virustotal",
			},
		},
		{
			name:   "With recursive",
			target: "example.com",
			options: map[string]string{
				"recursive": "true",
			},
			expected: []string{
				"-d", "example.com",
				"-silent",
				"-o", "/dev/stdout",
				"-recursive",
			},
		},
		{
			name:   "With timeout",
			target: "example.com",
			options: map[string]string{
				"timeout": "30",
			},
			expected: []string{
				"-d", "example.com",
				"-silent",
				"-o", "/dev/stdout",
				"-timeout", "30",
			},
		},
		{
			name:   "With config",
			target: "example.com",
			options: map[string]string{
				"config": "/path/to/config",
			},
			expected: []string{
				"-d", "example.com",
				"-silent",
				"-o", "/dev/stdout",
				"-config", "/path/to/config",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.buildArgs(tt.target, tt.options)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}

			// Check each argument
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("Arg %d: expected '%s', got '%s'", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestSubfinderScanner_parseOutput(t *testing.T) {
	s := &SubfinderScanner{
		BaseScannerTool: NewBaseScannerTool("subfinder", scanner.CategorySubdomain, "subfinder"),
	}

	tests := []struct {
		name     string
		output   []byte
		domain   string
		expected int // Number of expected subdomains
	}{
		{
			name: "Basic output",
			output: []byte(`api.example.com
www.example.com
mail.example.com`),
			domain:   "example.com",
			expected: 3,
		},
		{
			name: "Mixed output with invalid lines",
			output: []byte(`api.example.com

invalid.other.com
www.example.com`),
			domain:   "example.com",
			expected: 2,
		},
		{
			name:     "Empty output",
			output:   []byte(""),
			domain:   "example.com",
			expected: 0,
		},
		{
			name: "Output with spaces",
			output: []byte(`  api.example.com
   www.example.com   `),
			domain:   "example.com",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.parseOutput(tt.output, tt.domain)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != tt.expected {
				t.Errorf("Expected %d subdomains, got %d", tt.expected, len(result))
				return
			}

			// Validate each subdomain
			for _, subdomain := range result {
				if subdomain.Domain != tt.domain {
					t.Errorf("Expected domain '%s', got '%s'", tt.domain, subdomain.Domain)
				}
				if subdomain.Source != "subfinder" {
					t.Errorf("Expected source 'subfinder', got '%s'", subdomain.Source)
				}
				if !contains(subdomain.Subdomain, tt.domain) {
					t.Errorf("Subdomain '%s' should contain domain '%s'", subdomain.Subdomain, tt.domain)
				}
			}
		})
	}
}

func TestSubfinderScanner_Scan_MockExecution(t *testing.T) {
	// 这里我们测试扫描逻辑，但不实际执行subfinder命令
	s := &SubfinderScanner{
		BaseScannerTool: NewBaseScannerTool("subfinder", scanner.CategorySubdomain, "echo"),
	}

	config := scanner.ScanConfig{
		ProjectID: 1,
		Target:    "example.com",
		Timeout:   30 * time.Second,
		Options:   map[string]string{},
	}

	ctx := context.Background()
	result, err := s.Scan(ctx, config)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result, got nil")
		return
	}

	// Check if result contains SubdomainData
	if _, ok := result.Data.(scanner.SubdomainData); !ok {
		t.Error("Expected SubdomainData, got different type")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		   len(s) > len(substr) && s[:len(substr)] == substr ||
		   s == substr
}