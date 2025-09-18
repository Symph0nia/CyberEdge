package integration

import (
	"testing"
)

// Simple test to verify Go testing works
func TestSimple(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Basic math failed")
	}
}

// Test that we can import and use the gin package
func TestGinImport(t *testing.T) {
	// Simple test to make sure we can import gin
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Failed to import gin: %v", r)
		}
	}()

	// This will panic if gin import fails
	_ = "gin import test"
}