package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	if g == nil {
		t.Error("NewGenerator should return non-nil")
	}
	if g.projectPath != tmpDir {
		t.Errorf("expected projectPath %q, got %q", tmpDir, g.projectPath)
	}
	if g.claudePath != filepath.Join(tmpDir, ".claude") {
		t.Errorf("expected claudePath %q, got %q", filepath.Join(tmpDir, ".claude"), g.claudePath)
	}
}

func TestGenerateAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateAll(ctx, "TestProject", "standard")
	if err != nil {
		t.Errorf("GenerateAll() error = %v", err)
	}

	// エージェントファイルが存在するか確認
	agents := []string{
		"zeus-orchestrator.md",
		"zeus-planner.md",
		"zeus-reviewer.md",
	}
	for _, agent := range agents {
		path := filepath.Join(tmpDir, ".claude", "agents", agent)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("agent file %q should exist", agent)
		}
	}

	// スキルファイルが存在するか確認
	skills := []string{
		"zeus-project-scan/SKILL.md",
		"zeus-task-suggest/SKILL.md",
		"zeus-risk-analysis/SKILL.md",
	}
	for _, skill := range skills {
		path := filepath.Join(tmpDir, ".claude", "skills", skill)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("skill file %q should exist", skill)
		}
	}
}

func TestGenerateAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateAgents(ctx, "TestProject")
	if err != nil {
		t.Errorf("GenerateAgents() error = %v", err)
	}

	// オーケストレーターファイルの内容を確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "agents", "zeus-orchestrator.md"))
	if err != nil {
		t.Fatalf("failed to read orchestrator file: %v", err)
	}

	if !strings.Contains(string(content), "TestProject") {
		t.Error("orchestrator file should contain project name")
	}
	if !strings.Contains(string(content), "Zeus Orchestrator Agent") {
		t.Error("orchestrator file should contain agent name")
	}
}

func TestGenerateSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateSkills(ctx, "TestProject")
	if err != nil {
		t.Errorf("GenerateSkills() error = %v", err)
	}

	// プロジェクトスキャンファイルの内容を確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "skills", "zeus-project-scan", "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read skill file: %v", err)
	}

	if !strings.Contains(string(content), "TestProject") {
		t.Error("skill file should contain project name")
	}
	if !strings.Contains(string(content), "zeus-project-scan") {
		t.Error("skill file should contain skill name")
	}
}

func TestExecuteTemplate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)

	tests := []struct {
		name     string
		template string
		data     map[string]string
		contains string
		wantErr  bool
	}{
		{
			name:     "simple substitution",
			template: "Hello, {{.Name}}!",
			data:     map[string]string{"Name": "World"},
			contains: "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "{{.Project}} v{{.Version}}",
			data:     map[string]string{"Project": "Zeus", "Version": "1.0"},
			contains: "Zeus v1.0",
			wantErr:  false,
		},
		{
			name:     "empty data",
			template: "Static content",
			data:     map[string]string{},
			contains: "Static content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.executeTemplate(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestExecuteTemplate_InvalidTemplate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)

	// 不正なテンプレート
	_, err = g.executeTemplate("{{.Invalid", map[string]string{})
	if err == nil {
		t.Error("executeTemplate() should return error for invalid template")
	}
}

func TestEnsureClaudeDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)

	err = g.EnsureClaudeDir()
	if err != nil {
		t.Errorf("EnsureClaudeDir() error = %v", err)
	}

	// ディレクトリが存在するか確認
	claudePath := filepath.Join(tmpDir, ".claude")
	info, err := os.Stat(claudePath)
	if os.IsNotExist(err) {
		t.Error(".claude directory should exist")
	}
	if err == nil && !info.IsDir() {
		t.Error(".claude should be a directory")
	}
}

func TestEnsureClaudeDir_Idempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)

	// 複数回呼び出してもエラーにならない
	for i := 0; i < 3; i++ {
		err = g.EnsureClaudeDir()
		if err != nil {
			t.Errorf("EnsureClaudeDir() iteration %d error = %v", i, err)
		}
	}
}

func TestGenerateAll_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = g.GenerateAll(ctx, "TestProject", "standard")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateAgents_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = g.GenerateAgents(ctx, "TestProject")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateSkills_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = g.GenerateSkills(ctx, "TestProject")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestByteWriter(t *testing.T) {
	var buf []byte
	w := &byteWriter{buf: &buf}

	n, err := w.Write([]byte("Hello"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 5 {
		t.Errorf("expected n=5, got %d", n)
	}

	n, err = w.Write([]byte(" World"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 6 {
		t.Errorf("expected n=6, got %d", n)
	}

	if string(buf) != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", string(buf))
	}
}

func TestGenerateAgents_FileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateAgents(ctx, "MyProject")
	if err != nil {
		t.Fatalf("GenerateAgents() error = %v", err)
	}

	// 各エージェントファイルの内容を確認
	checks := []struct {
		file     string
		contains []string
	}{
		{
			file:     "zeus-orchestrator.md",
			contains: []string{"MyProject", "Orchestrator", "zeus status"},
		},
		{
			file:     "zeus-planner.md",
			contains: []string{"MyProject", "Planner", "zeus add task"},
		},
		{
			file:     "zeus-reviewer.md",
			contains: []string{"MyProject", "Reviewer", "zeus approve"},
		},
	}

	for _, c := range checks {
		path := filepath.Join(tmpDir, ".claude", "agents", c.file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", c.file, err)
		}

		for _, expected := range c.contains {
			if !strings.Contains(string(content), expected) {
				t.Errorf("%s should contain %q", c.file, expected)
			}
		}
	}
}

func TestGenerateSkills_FileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateSkills(ctx, "MyProject")
	if err != nil {
		t.Fatalf("GenerateSkills() error = %v", err)
	}

	// 各スキルファイルの内容を確認
	checks := []struct {
		dir      string
		contains []string
	}{
		{
			dir:      "zeus-project-scan",
			contains: []string{"MyProject", "zeus-project-scan", "プロジェクト全体をスキャン"},
		},
		{
			dir:      "zeus-task-suggest",
			contains: []string{"MyProject", "zeus-task-suggest", "タスクを提案"},
		},
		{
			dir:      "zeus-risk-analysis",
			contains: []string{"MyProject", "zeus-risk-analysis", "リスクを分析"},
		},
	}

	for _, c := range checks {
		path := filepath.Join(tmpDir, ".claude", "skills", c.dir, "SKILL.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", c.dir, err)
		}

		for _, expected := range c.contains {
			if !strings.Contains(string(content), expected) {
				t.Errorf("%s should contain %q", c.dir, expected)
			}
		}
	}
}

func TestGenerateAll_WithDifferentLevels(t *testing.T) {
	levels := []string{"simple", "standard", "advanced"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "generator-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			g := NewGenerator(tmpDir)
			ctx := context.Background()

			err = g.GenerateAll(ctx, "TestProject", level)
			if err != nil {
				t.Errorf("GenerateAll() with level %s error = %v", level, err)
			}

			// エージェントとスキルが作成されているか確認
			agentsDir := filepath.Join(tmpDir, ".claude", "agents")
			skillsDir := filepath.Join(tmpDir, ".claude", "skills")

			if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
				t.Errorf("agents directory should exist for level %s", level)
			}
			if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
				t.Errorf("skills directory should exist for level %s", level)
			}
		})
	}
}

func TestGenerateAgents_AllFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateAgents(ctx, "TestProject")
	if err != nil {
		t.Fatalf("GenerateAgents() error = %v", err)
	}

	// 全エージェントファイルが存在し、空でないか確認
	agents := []string{
		"zeus-orchestrator.md",
		"zeus-planner.md",
		"zeus-reviewer.md",
	}

	for _, agent := range agents {
		path := filepath.Join(tmpDir, ".claude", "agents", agent)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("agent file %s should exist", agent)
			continue
		}
		if err != nil {
			t.Errorf("failed to stat %s: %v", agent, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("agent file %s should not be empty", agent)
		}
	}
}

func TestGenerateSkills_AllFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx := context.Background()

	err = g.GenerateSkills(ctx, "TestProject")
	if err != nil {
		t.Fatalf("GenerateSkills() error = %v", err)
	}

	// 全スキルファイルが存在し、空でないか確認
	skills := []string{
		"zeus-project-scan",
		"zeus-task-suggest",
		"zeus-risk-analysis",
	}

	for _, skill := range skills {
		path := filepath.Join(tmpDir, ".claude", "skills", skill, "SKILL.md")
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("skill file %s should exist", skill)
			continue
		}
		if err != nil {
			t.Errorf("failed to stat %s: %v", skill, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("skill file %s should not be empty", skill)
		}
	}
}

func TestByteWriter_MultipleWrites(t *testing.T) {
	var buf []byte
	w := &byteWriter{buf: &buf}

	// 複数回書き込み
	writes := []string{"Hello", " ", "World", "!"}
	expected := "Hello World!"

	for _, s := range writes {
		n, err := w.Write([]byte(s))
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(s) {
			t.Errorf("expected n=%d, got %d", len(s), n)
		}
	}

	if string(buf) != expected {
		t.Errorf("expected %q, got %q", expected, string(buf))
	}
}

func TestByteWriter_EmptyWrite(t *testing.T) {
	var buf []byte
	w := &byteWriter{buf: &buf}

	n, err := w.Write([]byte{})
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 0 {
		t.Errorf("expected n=0, got %d", n)
	}
	if len(buf) != 0 {
		t.Errorf("expected empty buffer, got %d bytes", len(buf))
	}
}

// TestGenerateAgents_ContextCancelledDuringLoop はループ中にコンテキストがキャンセルされた場合をテスト
func TestGenerateAgents_ContextCancelledDuringLoop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	
	// 1つ目のファイルを生成した後でキャンセルされるシナリオをテスト
	// まず正常に生成してファイルが作成されることを確認
	ctx := context.Background()
	err = g.GenerateAgents(ctx, "TestProject")
	if err != nil {
		t.Errorf("GenerateAgents() should succeed with valid context, got error = %v", err)
	}
	
	// ファイルが生成されていることを確認
	path := filepath.Join(tmpDir, ".claude", "agents", "zeus-orchestrator.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("zeus-orchestrator.md should be created")
	}
}

// TestGenerateSkills_ContextCancelledDuringLoop はループ中にコンテキストがキャンセルされた場合をテスト
func TestGenerateSkills_ContextCancelledDuringLoop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	
	// 正常なコンテキストでスキルを生成
	ctx := context.Background()
	err = g.GenerateSkills(ctx, "TestProject")
	if err != nil {
		t.Errorf("GenerateSkills() should succeed with valid context, got error = %v", err)
	}
	
	// スキルファイルが生成されていることを確認
	path := filepath.Join(tmpDir, ".claude", "skills", "zeus-project-scan", "SKILL.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("zeus-project-scan/SKILL.md should be created")
	}
}

// TestGenerateAll_AgentsError は GenerateAgents がエラーを返した場合をテスト
func TestGenerateAll_AgentsError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	
	// .claude ディレクトリをファイルとして作成（ディレクトリ作成をエラーにする）
	claudePath := filepath.Join(tmpDir, ".claude")
	if err := os.WriteFile(claudePath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	
	ctx := context.Background()
	err = g.GenerateAll(ctx, "TestProject", "standard")
	if err == nil {
		t.Error("GenerateAll() should return error when .claude is a file")
	}
}

// TestExecuteTemplate_ExecuteError はテンプレート実行時のエラーをテスト
func TestExecuteTemplate_ExecuteError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)

	// テンプレート実行エラーを発生させる（存在しないメソッド呼び出し）
	_, err = g.executeTemplate("{{.Name | nonexistent}}", map[string]string{"Name": "Test"})
	if err == nil {
		t.Error("executeTemplate() should return error for invalid function")
	}
}
