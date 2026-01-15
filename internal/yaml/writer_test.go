package yaml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	if w == nil {
		t.Error("NewWriter should return non-nil")
	}
}

func TestWriter_Stringify(t *testing.T) {
	w := NewWriter()

	tests := []struct {
		name     string
		data     interface{}
		contains string
		wantErr  bool
	}{
		{
			name:     "simple map",
			data:     map[string]string{"name": "test"},
			contains: "name: test",
			wantErr:  false,
		},
		{
			name: "nested struct",
			data: struct {
				Project struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				} `yaml:"project"`
			}{
				Project: struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				}{
					Name:    "zeus",
					Version: "1.0",
				},
			},
			contains: "name: zeus",
			wantErr:  false,
		},
		{
			name:     "slice",
			data:     []string{"item1", "item2"},
			contains: "- item1",
			wantErr:  false,
		},
		{
			name:     "nil",
			data:     nil,
			contains: "null",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := w.Stringify(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stringify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(string(result), tt.contains) {
				t.Errorf("Stringify() result should contain %q, got %q", tt.contains, string(result))
			}
		})
	}
}

func TestWriter_WriteFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "writer-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewWriter()
	data := map[string]string{
		"name":    "test",
		"version": "1.0",
	}

	testFile := filepath.Join(tmpDir, "test.yaml")
	err = w.WriteFile(testFile, data)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("WriteFile() should create file")
	}

	// ファイル内容を確認
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "name: test") {
		t.Errorf("file content should contain 'name: test', got %q", string(content))
	}
}

func TestWriter_WriteFile_CreateDir(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "writer-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewWriter()
	data := map[string]string{"name": "test"}

	// 存在しないサブディレクトリ内のファイルに書き込み
	testFile := filepath.Join(tmpDir, "subdir", "nested", "test.yaml")
	err = w.WriteFile(testFile, data)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("WriteFile() should create file in nested directory")
	}
}

func TestWriter_Stringify_ComplexStruct(t *testing.T) {
	w := NewWriter()

	type Task struct {
		ID           string   `yaml:"id"`
		Title        string   `yaml:"title"`
		Dependencies []string `yaml:"dependencies"`
	}

	task := Task{
		ID:           "task-1",
		Title:        "Test Task",
		Dependencies: []string{"task-0", "task-2"},
	}

	result, err := w.Stringify(task)
	if err != nil {
		t.Errorf("Stringify() error = %v", err)
	}

	content := string(result)
	if !strings.Contains(content, "id: task-1") {
		t.Errorf("result should contain 'id: task-1', got %q", content)
	}
	if !strings.Contains(content, "title: Test Task") {
		t.Errorf("result should contain 'title: Test Task', got %q", content)
	}
	if !strings.Contains(content, "- task-0") {
		t.Errorf("result should contain '- task-0', got %q", content)
	}
}
