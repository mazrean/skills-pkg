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
	t.Run("success: calculate hash for simple directory", func(t *testing.T) {
		// Setup: Create temporary directory with test files
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result, err := svc.CalculateHash(ctx, tmpDir)

		// Verify
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
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
	})

	t.Run("success: same content produces same hash", func(t *testing.T) {
		// Setup: Create two directories with identical content
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()
		content := []byte("identical content")

		if err := os.WriteFile(filepath.Join(tmpDir1, "file.txt"), content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir2, "file.txt"), content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result1, err1 := svc.CalculateHash(ctx, tmpDir1)
		result2, err2 := svc.CalculateHash(ctx, tmpDir2)

		// Verify
		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got: %v, %v", err1, err2)
		}
		if result1.Value != result2.Value {
			t.Errorf("Expected identical hashes for identical content, got '%s' and '%s'", result1.Value, result2.Value)
		}
	})

	t.Run("success: different content produces different hash", func(t *testing.T) {
		// Setup: Create two directories with different content
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		if err := os.WriteFile(filepath.Join(tmpDir1, "file.txt"), []byte("content 1"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir2, "file.txt"), []byte("content 2"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result1, err1 := svc.CalculateHash(ctx, tmpDir1)
		result2, err2 := svc.CalculateHash(ctx, tmpDir2)

		// Verify
		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got: %v, %v", err1, err2)
		}
		if result1.Value == result2.Value {
			t.Error("Expected different hashes for different content")
		}
	})

	t.Run("success: recursive hash includes subdirectories", func(t *testing.T) {
		// Setup: Create directory with nested structure
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644); err != nil {
			t.Fatalf("Failed to create nested file: %v", err)
		}

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result, err := svc.CalculateHash(ctx, tmpDir)

		// Verify
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result.Value == "" {
			t.Error("Expected non-empty hash for directory with subdirectories")
		}
	})

	t.Run("error: directory does not exist", func(t *testing.T) {
		// Setup: Use non-existent directory path
		nonExistentDir := "/tmp/non-existent-dir-12345"

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result, err := svc.CalculateHash(ctx, nonExistentDir)

		// Verify
		if err == nil {
			t.Fatal("Expected error for non-existent directory, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})

	t.Run("error: path is a file, not a directory", func(t *testing.T) {
		// Setup: Create a file instead of directory
		tmpFile := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Execute
		svc := NewDirhashService()
		ctx := context.Background()
		result, err := svc.CalculateHash(ctx, tmpFile)

		// Verify
		if err == nil {
			t.Fatal("Expected error for file path, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})
}

// TestDirhashService_HashAlgorithm tests the HashAlgorithm method
func TestDirhashService_HashAlgorithm(t *testing.T) {
	svc := NewDirhashService()
	algo := svc.HashAlgorithm()

	if algo != "sha256" {
		t.Errorf("Expected algorithm 'sha256', got '%s'", algo)
	}
}

// TestDirhashService_ImplementsInterface verifies that DirhashService implements HashService
func TestDirhashService_ImplementsInterface(t *testing.T) {
	var _ port.HashService = (*DirhashService)(nil)
}
