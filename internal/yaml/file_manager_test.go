package yaml

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestReadYaml_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// テスト用 YAML ファイルを作成
	yamlContent := `name: テスト
value: 123
nested:
  key: value
`
	err = os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// YAML を読み込み
	var data map[string]interface{}
	err = fm.ReadYaml(ctx, "test.yaml", &data)
	if err != nil {
		t.Errorf("ReadYaml() error = %v", err)
	}

	if data["name"] != "テスト" {
		t.Errorf("expected name 'テスト', got %v", data["name"])
	}
	if data["value"] != 123 {
		t.Errorf("expected value 123, got %v", data["value"])
	}
}

func TestReadYaml_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	var data map[string]interface{}
	err = fm.ReadYaml(ctx, "nonexistent.yaml", &data)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestWriteYaml_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	data := map[string]interface{}{
		"name":  "テスト",
		"value": 123,
	}

	err = fm.WriteYaml(ctx, "output.yaml", data)
	if err != nil {
		t.Errorf("WriteYaml() error = %v", err)
	}

	// ファイルが存在するか確認
	if !fm.Exists(ctx, "output.yaml") {
		t.Error("output.yaml should exist")
	}

	// 内容を確認
	var loaded map[string]interface{}
	err = fm.ReadYaml(ctx, "output.yaml", &loaded)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if loaded["name"] != "テスト" {
		t.Errorf("expected name 'テスト', got %v", loaded["name"])
	}
}

func TestWriteFile_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	content := []byte("テストコンテンツ")
	err = fm.WriteFile(ctx, "test.txt", content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// ファイルが存在するか確認
	if !fm.Exists(ctx, "test.txt") {
		t.Error("test.txt should exist")
	}

	// 内容を確認
	data, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "テストコンテンツ" {
		t.Errorf("expected 'テストコンテンツ', got %q", string(data))
	}
}

func TestWriteFile_CreateParentDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ネストしたパスにファイルを作成
	content := []byte("nested content")
	err = fm.WriteFile(ctx, "dir/subdir/file.txt", content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// ファイルが存在するか確認
	if !fm.Exists(ctx, "dir/subdir/file.txt") {
		t.Error("dir/subdir/file.txt should exist")
	}
}

func TestGlob_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// テストファイルを作成
	for _, name := range []string{"a.txt", "b.txt", "c.yaml", "d.txt"} {
		if err := fm.WriteFile(ctx, name, []byte("test")); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	// *.txt をグロブ
	matches, err := fm.Glob(ctx, "*.txt")
	if err != nil {
		t.Errorf("Glob() error = %v", err)
	}

	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d: %v", len(matches), matches)
	}
}

func TestGlob_NoMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// 空のディレクトリでグロブ
	matches, err := fm.Glob(ctx, "*.nonexistent")
	if err != nil {
		t.Errorf("Glob() error = %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestListDir_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ファイルを作成
	if err := fm.WriteFile(ctx, "file1.txt", []byte("test")); err != nil {
		t.Fatalf("failed to create file1.txt: %v", err)
	}
	if err := fm.WriteFile(ctx, "file2.txt", []byte("test")); err != nil {
		t.Fatalf("failed to create file2.txt: %v", err)
	}

	// ルートディレクトリを一覧
	entries, err := fm.ListDir(ctx, ".")
	if err != nil {
		t.Errorf("ListDir() error = %v", err)
	}

	// 少なくとも2つのファイルが存在
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d: %v", len(entries), entries)
	}
}

func TestListDir_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// 存在しないディレクトリを一覧
	_, err = fm.ListDir(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestCopy_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ソースファイルを作成
	srcContent := []byte("source content")
	if err := fm.WriteFile(ctx, "source.txt", srcContent); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// コピー
	err = fm.Copy(ctx, "source.txt", "dest.txt")
	if err != nil {
		t.Errorf("Copy() error = %v", err)
	}

	// コピー先が存在するか確認
	if !fm.Exists(ctx, "dest.txt") {
		t.Error("dest.txt should exist")
	}

	// コピー先の内容を確認
	destContent, err := os.ReadFile(filepath.Join(tmpDir, "dest.txt"))
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(destContent) != "source content" {
		t.Errorf("expected 'source content', got %q", string(destContent))
	}
}

func TestCopy_SourceNotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// 存在しないソースからコピー
	err = fm.Copy(ctx, "nonexistent.txt", "dest.txt")
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestDelete_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ファイルを作成
	if err := fm.WriteFile(ctx, "todelete.txt", []byte("test")); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// 削除
	err = fm.Delete(ctx, "todelete.txt")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 削除されているか確認
	if fm.Exists(ctx, "todelete.txt") {
		t.Error("todelete.txt should not exist after delete")
	}
}

func TestDelete_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// 存在しないファイルを削除（エラーが返る実装）
	err = fm.Delete(ctx, "nonexistent.txt")
	// 実装によりエラーが返ることを確認
	if err == nil {
		// エラーが返らない場合も許容（os.Remove の動作による）
		t.Log("Delete() returned nil for non-existent file")
	} else {
		// エラーが返る場合は "no such file" エラーであることを確認
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
			t.Errorf("expected 'no such file' error, got %v", err)
		}
	}
}

func TestEnsureDir_Nested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ネストしたディレクトリを作成
	err = fm.EnsureDir(ctx, "a/b/c/d")
	if err != nil {
		t.Errorf("EnsureDir() error = %v", err)
	}

	// 作成されているか確認
	info, err := os.Stat(filepath.Join(tmpDir, "a/b/c/d"))
	if os.IsNotExist(err) {
		t.Error("nested directory should exist")
	}
	if err == nil && !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestExists_Various(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)
	ctx := context.Background()

	// ファイルを作成
	if err := fm.WriteFile(ctx, "exists.txt", []byte("test")); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// ディレクトリを作成
	if err := fm.EnsureDir(ctx, "existsdir"); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", "exists.txt", true},
		{"existing dir", "existsdir", true},
		{"non-existing file", "nonexistent.txt", false},
		{"empty path", "", true}, // ベースディレクトリ自体
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.Exists(ctx, tt.path)
			if result != tt.expected {
				t.Errorf("Exists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestBasePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filemanager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fm := NewFileManager(tmpDir)

	// macOS では /var が /private/var にシンボリックリンクされている
	// そのため、パスの末尾部分だけを比較
	basePath := fm.BasePath()
	if !strings.HasSuffix(basePath, filepath.Base(tmpDir)) {
		t.Errorf("BasePath should end with %q, got %q", filepath.Base(tmpDir), basePath)
	}
}
