package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestDirhashService_CalculateHash tests the CalculateHash method
func TestDirhashService_CalculateHash(t *testing.T) {
	tests := []struct {
		setupFunc func(t *testing.T) string
		checkFunc func(t *testing.T, result *port.HashResult, err error)
		name      string
		wantErr   bool
	}{
		{
			name: "success: calculate hash for simple directory",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.txt")
				if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return tmpDir
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				if result == nil {
					t.Fatal("Expected result, got nil")
				}
				if result.Algorithm != "sha256" {
					t.Errorf("Expected algorithm 'sha256', got '%s'", result.Algorithm)
				}
				if result.Value == "" {
					t.Error("Expected non-empty hash value")
				}
				// Hash should be in Go module format: "h1:<base64>"
				if len(result.Value) < 4 || result.Value[:3] != "h1:" {
					t.Errorf("Expected hash format 'h1:<base64>', got '%s'", result.Value)
				}
			},
		},
		{
			name: "success: same content produces same hash",
			setupFunc: func(t *testing.T) string {
				// This test needs two directories, will be handled specially
				return ""
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				// Special handling for this case
				tmpDir1 := t.TempDir()
				tmpDir2 := t.TempDir()
				content := []byte("identical content")

				if err := os.WriteFile(filepath.Join(tmpDir1, "file.txt"), content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir2, "file.txt"), content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				svc := NewDirhashService()
				ctx := context.Background()
				result1, err1 := svc.CalculateHash(ctx, tmpDir1)
				result2, err2 := svc.CalculateHash(ctx, tmpDir2)

				if err1 != nil || err2 != nil {
					t.Fatalf("Expected no errors, got: %v, %v", err1, err2)
				}
				if result1.Value != result2.Value {
					t.Errorf("Expected identical hashes for identical content, got '%s' and '%s'", result1.Value, result2.Value)
				}
			},
		},
		{
			name: "success: different content produces different hash",
			setupFunc: func(t *testing.T) string {
				// This test needs two directories, will be handled specially
				return ""
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				// Special handling for this case
				tmpDir1 := t.TempDir()
				tmpDir2 := t.TempDir()

				if err := os.WriteFile(filepath.Join(tmpDir1, "file.txt"), []byte("content 1"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir2, "file.txt"), []byte("content 2"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				svc := NewDirhashService()
				ctx := context.Background()
				result1, err1 := svc.CalculateHash(ctx, tmpDir1)
				result2, err2 := svc.CalculateHash(ctx, tmpDir2)

				if err1 != nil || err2 != nil {
					t.Fatalf("Expected no errors, got: %v, %v", err1, err2)
				}
				if result1.Value == result2.Value {
					t.Error("Expected different hashes for different content")
				}
			},
		},
		{
			name: "success: recursive hash includes subdirectories",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatalf("Failed to create subdirectory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644); err != nil {
					t.Fatalf("Failed to create nested file: %v", err)
				}
				return tmpDir
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				if result.Value == "" {
					t.Error("Expected non-empty hash for directory with subdirectories")
				}
			},
		},
		{
			name: "error: directory does not exist",
			setupFunc: func(t *testing.T) string {
				return "/tmp/non-existent-dir-12345"
			},
			wantErr: true,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				if result != nil {
					t.Errorf("Expected nil result on error, got: %v", result)
				}
			},
		},
		{
			name: "error: path is a file, not a directory",
			setupFunc: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "file.txt")
				if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return tmpFile
			},
			wantErr: true,
			checkFunc: func(t *testing.T, result *port.HashResult, err error) {
				if result != nil {
					t.Errorf("Expected nil result on error, got: %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special cases that handle their own execution
			if tt.name == "success: same content produces same hash" || tt.name == "success: different content produces different hash" {
				tt.checkFunc(t, nil, nil)
				return
			}

			dirPath := tt.setupFunc(t)

			svc := NewDirhashService()
			ctx := context.Background()
			result, err := svc.CalculateHash(ctx, dirPath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error: %v, got: %v", tt.wantErr, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

// TestDirhashService_HashAlgorithm tests the HashAlgorithm method
func TestDirhashService_HashAlgorithm(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return sha256",
			want: "sha256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewDirhashService()
			if got := svc.HashAlgorithm(); got != tt.want {
				t.Errorf("Expected algorithm %q, got %q", tt.want, got)
			}
		})
	}
}

// TestDirhashService_ImplementsInterface verifies that DirhashService implements HashService
func TestDirhashService_ImplementsInterface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "DirhashService implements HashService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _ port.HashService = (*DirhashService)(nil)
		})
	}
}
