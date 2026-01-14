package yaml

import (
	"context"
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
	ctx := context.Background()

	// Exists でトラバーサルを防止
	if fm.Exists(ctx, "../etc/passwd") {
		t.Error("Exists should return false for traversal path")
	}

	// EnsureDir でトラバーサルを防止
	err = fm.EnsureDir(ctx, "../escape")
	if err != ErrPathTraversal {
		t.Errorf("EnsureDir should return ErrPathTraversal, got %v", err)
	}

	// WriteFile でトラバーサルを防止
	err = fm.WriteFile(ctx, "../escape.txt", []byte("test"))
	if err != ErrPathTraversal {
		t.Errorf("WriteFile should return ErrPathTraversal, got %v", err)
	}

	// Delete でトラバーサルを防止
	err = fm.Delete(ctx, "../escape.txt")
	if err != ErrPathTraversal {
		t.Errorf("Delete should return ErrPathTraversal, got %v", err)
	}

	// Copy でトラバーサルを防止
	err = fm.Copy(ctx, "../src.txt", "dest.txt")
	if err != ErrPathTraversal {
		t.Errorf("Copy (src) should return ErrPathTraversal, got %v", err)
	}
	err = fm.Copy(ctx, "src.txt", "../dest.txt")
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
	ctx := context.Background()

	// パターンに .. を含む場合はエラー
	_, err = fm.Glob(ctx, "../*.txt")
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
	ctx := context.Background()

	// トラバーサルパスはエラー
	_, err = fm.ListDir(ctx, "../")
	if err != ErrPathTraversal {
		t.Errorf("ListDir should return ErrPathTraversal, got %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// キャンセル済みコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Exists はキャンセル時 false を返す
	if fm.Exists(ctx, "file.txt") {
		t.Error("Exists should return false for cancelled context")
	}

	// EnsureDir はキャンセル時エラーを返す
	err = fm.EnsureDir(ctx, "dir")
	if err != context.Canceled {
		t.Errorf("EnsureDir should return context.Canceled, got %v", err)
	}

	// WriteFile はキャンセル時エラーを返す
	err = fm.WriteFile(ctx, "file.txt", []byte("test"))
	if err != context.Canceled {
		t.Errorf("WriteFile should return context.Canceled, got %v", err)
	}

	// ReadYaml はキャンセル時エラーを返す
	var data interface{}
	err = fm.ReadYaml(ctx, "file.yaml", &data)
	if err != context.Canceled {
		t.Errorf("ReadYaml should return context.Canceled, got %v", err)
	}

	// WriteYaml はキャンセル時エラーを返す
	err = fm.WriteYaml(ctx, "file.yaml", data)
	if err != context.Canceled {
		t.Errorf("WriteYaml should return context.Canceled, got %v", err)
	}

	// Delete はキャンセル時エラーを返す
	err = fm.Delete(ctx, "file.txt")
	if err != context.Canceled {
		t.Errorf("Delete should return context.Canceled, got %v", err)
	}

	// Copy はキャンセル時エラーを返す
	err = fm.Copy(ctx, "src.txt", "dest.txt")
	if err != context.Canceled {
		t.Errorf("Copy should return context.Canceled, got %v", err)
	}

	// Glob はキャンセル時エラーを返す
	_, err = fm.Glob(ctx, "*.txt")
	if err != context.Canceled {
		t.Errorf("Glob should return context.Canceled, got %v", err)
	}

	// ListDir はキャンセル時エラーを返す
	_, err = fm.ListDir(ctx, ".")
	if err != context.Canceled {
		t.Errorf("ListDir should return context.Canceled, got %v", err)
	}
}
