package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Error("NewParser should return non-nil")
	}
}

func TestParser_Parse_ValidYAML(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "simple key-value",
			content: "name: test",
			wantErr: false,
		},
		{
			name:    "nested structure",
			content: "project:\n  name: zeus\n  version: 1.0",
			wantErr: false,
		},
		{
			name:    "list",
			content: "items:\n  - item1\n  - item2",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := p.Parse([]byte(tt.content), &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_Parse_InvalidYAML(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "invalid indentation",
			content: "name: test\n  invalid: indent",
		},
		{
			name:    "tabs instead of spaces",
			content: "name:\n\t- item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := p.Parse([]byte(tt.content), &result)
			if err == nil {
				t.Error("Parse() should return error for invalid YAML")
			}
		})
	}
}

func TestParser_ReadFile_Exists(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用ファイルを作成
	content := "name: test\nversion: 1.0"
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	p := NewParser()
	var result struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	err = p.ReadFile(testFile, &result)
	if err != nil {
		t.Errorf("ReadFile() error = %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected name 'test', got %q", result.Name)
	}
	if result.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", result.Version)
	}
}

func TestParser_ReadFile_NotExists(t *testing.T) {
	p := NewParser()
	var result map[string]interface{}

	err := p.ReadFile("/non/existent/file.yaml", &result)
	if err == nil {
		t.Error("ReadFile() should return error for non-existent file")
	}
}

func TestParser_Parse_StructTypes(t *testing.T) {
	p := NewParser()

	type Task struct {
		ID     string `yaml:"id"`
		Title  string `yaml:"title"`
		Status string `yaml:"status"`
	}

	content := "id: task-1\ntitle: Test Task\nstatus: pending"
	var task Task

	err := p.Parse([]byte(content), &task)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if task.ID != "task-1" {
		t.Errorf("expected ID 'task-1', got %q", task.ID)
	}
	if task.Title != "Test Task" {
		t.Errorf("expected Title 'Test Task', got %q", task.Title)
	}
	if task.Status != "pending" {
		t.Errorf("expected Status 'pending', got %q", task.Status)
	}
}
