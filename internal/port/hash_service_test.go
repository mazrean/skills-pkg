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
		name          string
		result        *port.HashResult
		wantAlgorithm string
		wantValue     string
	}{
		{
			name: "valid_hash_result",
			result: &port.HashResult{
				Algorithm: "sha256",
				Value:     "a1b2c3d4e5f6",
			},
			wantAlgorithm: "sha256",
			wantValue:     "a1b2c3d4e5f6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Algorithm != tt.wantAlgorithm {
				t.Errorf("HashResult.Algorithm = %v, want %v", tt.result.Algorithm, tt.wantAlgorithm)
			}
			if tt.result.Value != tt.wantValue {
				t.Errorf("HashResult.Value = %v, want %v", tt.result.Value, tt.wantValue)
			}
		})
	}
}

// TestHashServiceAlgorithm tests that hash algorithm is returned correctly.
// Requirements: 5.1, 5.2
func TestHashServiceAlgorithm(t *testing.T) {
	tests := []struct {
		service      port.HashService
		name         string
		wantNonEmpty bool
	}{
		{
			name:         "mock_hash_service",
			service:      &mockHashService{},
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algo := tt.service.HashAlgorithm()
			if tt.wantNonEmpty && algo == "" {
				t.Error("HashAlgorithm() returned empty string")
			}
		})
	}
}

// mockHashService is a mock implementation of HashService for testing.
type mockHashService struct{}

func (m *mockHashService) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	return &port.HashResult{
		Algorithm: "sha256",
		Value:     "mockhash",
	}, nil
}

func (m *mockHashService) HashAlgorithm() string {
	return "sha256"
}
