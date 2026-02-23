package port_test

import (
	"context"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestHashServiceInterface verifies that the HashService interface contract
// can be satisfied by a mock implementation.
// Requirements: 5.1
func TestHashServiceInterface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "interface_contract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that a mock implementation satisfies the interface
			var _ port.HashService = &mockHashService{}
		})
	}
}

// TestHashResultStructure tests HashResult struct fields.
// Requirements: 5.2
func TestHashResultStructure(t *testing.T) {
	tests := []struct {
		name      string
		result    *port.HashResult
		wantValue string
	}{
		{
			name: "valid_hash_result",
			result: &port.HashResult{
				Value: "h1:a1b2c3d4e5f6",
			},
			wantValue: "h1:a1b2c3d4e5f6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Value != tt.wantValue {
				t.Errorf("HashResult.Value = %v, want %v", tt.result.Value, tt.wantValue)
			}
		})
	}
}

// mockHashService is a mock implementation of HashService for testing.
type mockHashService struct{}

func (m *mockHashService) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	return &port.HashResult{
		Value: "h1:mockhash",
	}, nil
}
