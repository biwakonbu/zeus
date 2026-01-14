package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	// テスト用ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		// 正常ケース
		{"empty path", "", false},
		{"simple file", "file.txt", false},
		{"nested path", "dir/subdir/file.txt", false},
		{"path with dot", "./file.txt", false},

		// 異常ケース: ディレクトリトラバーサル
		{"parent directory", "../file.txt", true},
		{"nested parent", "dir/../../file.txt", true},
		{"absolute path", "/etc/passwd", true},
		{"deep traversal", "a/b/c/../../../..", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.ValidatePath(tt.path)
			if tt.expectError && err == nil {
				t.Errorf("expected error for path %q, but got nil", tt.path)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for path %q: %v", tt.path, err)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// 正常ケース
	resolved, err := fm.ResolvePath("file.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := filepath.Join(fm.BasePath(), "file.txt")
	if resolved != expected {
		t.Errorf("expected %q, got %q", expected, resolved)
	}

	// 異常ケース
	_, err = fm.ResolvePath("../escape.txt")
	if err != ErrPathTraversal {
		t.Errorf("expected ErrPathTraversal, got %v", err)
	}
}

func TestFileOperationsWithTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// Exists でトラバーサルを防止
	if fm.Exists("../etc/passwd") {
		t.Error("Exists should return false for traversal path")
	}

	// EnsureDir でトラバーサルを防止
	err = fm.EnsureDir("../escape")
	if err != ErrPathTraversal {
		t.Errorf("EnsureDir should return ErrPathTraversal, got %v", err)
	}

	// WriteFile でトラバーサルを防止
	err = fm.WriteFile("../escape.txt", []byte("test"))
	if err != ErrPathTraversal {
		t.Errorf("WriteFile should return ErrPathTraversal, got %v", err)
	}

	// Delete でトラバーサルを防止
	err = fm.Delete("../escape.txt")
	if err != ErrPathTraversal {
		t.Errorf("Delete should return ErrPathTraversal, got %v", err)
	}

	// Copy でトラバーサルを防止
	err = fm.Copy("../src.txt", "dest.txt")
	if err != ErrPathTraversal {
		t.Errorf("Copy (src) should return ErrPathTraversal, got %v", err)
	}
	err = fm.Copy("src.txt", "../dest.txt")
	if err != ErrPathTraversal {
		t.Errorf("Copy (dest) should return ErrPathTraversal, got %v", err)
	}
}

func TestGlobWithTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// パターンに .. を含む場合はエラー
	_, err = fm.Glob("../*.txt")
	if err != ErrPathTraversal {
		t.Errorf("Glob should return ErrPathTraversal for traversal pattern, got %v", err)
	}
}

func TestListDirWithTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// トラバーサルパスはエラー
	_, err = fm.ListDir("../")
	if err != ErrPathTraversal {
		t.Errorf("ListDir should return ErrPathTraversal, got %v", err)
	}
}
