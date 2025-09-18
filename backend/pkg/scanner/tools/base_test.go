package tools

import (
	"context"
	"testing"
	"time"

	"cyberedge/pkg/scanner"
)

func TestBaseScannerTool_GetName(t *testing.T) {
	tool := NewBaseScannerTool("test-tool", scanner.CategorySubdomain, "echo")

	if tool.GetName() != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", tool.GetName())
	}
}

func TestBaseScannerTool_GetCategory(t *testing.T) {
	tool := NewBaseScannerTool("test-tool", scanner.CategoryPort, "echo")

	if tool.GetCategory() != scanner.CategoryPort {
		t.Errorf("Expected category '%s', got '%s'", scanner.CategoryPort, tool.GetCategory())
	}
}

func TestBaseScannerTool_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		cmdPath  string
		expected bool
	}{
		{
			name:     "Available command",
			cmdPath:  "echo",
			expected: true,
		},
		{
			name:     "Unavailable command",
			cmdPath:  "nonexistent-command-12345",
			expected: false,
		},
		{
			name:     "Empty command",
			cmdPath:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewBaseScannerTool("test", scanner.CategorySubdomain, tt.cmdPath)
			if tool.IsAvailable() != tt.expected {
				t.Errorf("Expected availability %v, got %v", tt.expected, tool.IsAvailable())
			}
		})
	}
}

func TestBaseScannerTool_ValidateConfig(t *testing.T) {
	tool := NewBaseScannerTool("test", scanner.CategorySubdomain, "echo")

	tests := []struct {
		name      string
		config    scanner.ScanConfig
		expectErr bool
	}{
		{
			name: "Valid config",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "example.com",
			},
			expectErr: false,
		},
		{
			name: "Empty target",
			config: scanner.ScanConfig{
				ProjectID: 1,
				Target:    "",
			},
			expectErr: true,
		},
		{
			name: "Zero project ID",
			config: scanner.ScanConfig{
				ProjectID: 0,
				Target:    "example.com",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.ValidateConfig(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestBaseScannerTool_SanitizeTarget(t *testing.T) {
	tool := NewBaseScannerTool("test", scanner.CategorySubdomain, "echo")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic domain",
			input:    "Example.COM",
			expected: "example.com",
		},
		{
			name:     "HTTP prefix",
			input:    "http://example.com",
			expected: "example.com",
		},
		{
			name:     "HTTPS prefix",
			input:    "https://example.com",
			expected: "example.com",
		},
		{
			name:     "With path",
			input:    "example.com/path/to/resource",
			expected: "example.com",
		},
		{
			name:     "With protocol and path",
			input:    "https://example.com/path",
			expected: "example.com",
		},
		{
			name:     "Whitespace",
			input:    "  example.com  ",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.SanitizeTarget(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBaseScannerTool_ParseLines(t *testing.T) {
	tool := NewBaseScannerTool("test", scanner.CategorySubdomain, "echo")

	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{
			name:     "Basic lines",
			input:    []byte("line1\nline2\nline3"),
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "Empty lines",
			input:    []byte("line1\n\nline3\n"),
			expected: []string{"line1", "line3"},
		},
		{
			name:     "Comments",
			input:    []byte("line1\n# comment\nline3"),
			expected: []string{"line1", "line3"},
		},
		{
			name:     "Whitespace",
			input:    []byte("  line1  \n\t\nline3\n"),
			expected: []string{"line1", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.ParseLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d: expected '%s', got '%s'", i, tt.expected[i], line)
				}
			}
		})
	}
}

func TestBaseScannerTool_ExecuteCommand(t *testing.T) {
	tool := NewBaseScannerTool("test", scanner.CategorySubdomain, "echo")

	tests := []struct {
		name      string
		args      []string
		timeout   time.Duration
		expectErr bool
	}{
		{
			name:      "Simple echo",
			args:      []string{"hello", "world"},
			timeout:   5 * time.Second,
			expectErr: false,
		},
		{
			name:      "No timeout",
			args:      []string{"test"},
			timeout:   0,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := tool.ExecuteCommand(ctx, tt.args, tt.timeout)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
				return
			}

			if !tt.expectErr && len(output) == 0 {
				t.Error("Expected output, got empty")
			}
		})
	}
}

func TestBaseScannerTool_ExecuteCommand_Timeout(t *testing.T) {
	tool := NewBaseScannerTool("test", scanner.CategorySubdomain, "sleep")

	ctx := context.Background()
	_, err := tool.ExecuteCommand(ctx, []string{"2"}, 100*time.Millisecond)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}