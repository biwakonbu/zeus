package generator

import (
	"context"
	"os"
	"path/filepath"
	"text/template"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// Generator は Claude Code 連携ファイルを生成
type Generator struct {
	projectPath string
	claudePath  string
	fileManager *yaml.FileManager
}

// NewGenerator は新しい Generator を作成
func NewGenerator(projectPath string) *Generator {
	claudePath := filepath.Join(projectPath, ".claude")
	return &Generator{
		projectPath: projectPath,
		claudePath:  claudePath,
		fileManager: yaml.NewFileManager(claudePath),
	}
}

// GenerateAll は全ての Claude Code 連携ファイルを生成（Context対応）
func (g *Generator) GenerateAll(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.GenerateAgents(ctx, projectName); err != nil {
		return err
	}
	if err := g.GenerateSkills(ctx, projectName); err != nil {
		return err
	}
	return nil
}

// GenerateAgents はエージェントファイルを生成（Context対応）
// 埋め込みファイルから読み込んでテンプレート変数を展開
func (g *Generator) GenerateAgents(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.fileManager.EnsureDir(ctx, "agents"); err != nil {
		return err
	}

	agents := []string{
		"zeus-orchestrator.md",
		"zeus-planner.md",
		"zeus-reviewer.md",
	}

	data := map[string]string{
		"ProjectName": projectName,
	}

	for _, agent := range agents {
		if err := ctx.Err(); err != nil {
			return err
		}

		// 埋め込みファイルから読み込み
		content, err := agentFS.ReadFile(filepath.Join("assets", "agents", agent))
		if err != nil {
			return err
		}

		// テンプレート変数を展開
		result, err := g.executeTemplate(string(content), data)
		if err != nil {
			return err
		}

		if err := g.fileManager.WriteFile(ctx, filepath.Join("agents", agent), []byte(result)); err != nil {
			return err
		}
	}

	return nil
}

// GenerateSkills はスキルファイルを生成（Context対応）
// 埋め込みファイルから読み込んでテンプレート変数を展開
func (g *Generator) GenerateSkills(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	skills := []string{
		"zeus-project-scan",
		"zeus-activity-suggest",
		"zeus-risk-analysis",
		"zeus-wbs-design",
	}

	data := map[string]string{
		"ProjectName": projectName,
	}

	for _, skill := range skills {
		if err := ctx.Err(); err != nil {
			return err
		}

		skillDir := filepath.Join("skills", skill)
		if err := g.fileManager.EnsureDir(ctx, skillDir); err != nil {
			return err
		}

		// 埋め込みファイルから読み込み
		content, err := skillFS.ReadFile(filepath.Join("assets", "skills", skill, "SKILL.md"))
		if err != nil {
			return err
		}

		// テンプレート変数を展開
		result, err := g.executeTemplate(string(content), data)
		if err != nil {
			return err
		}

		if err := g.fileManager.WriteFile(ctx, filepath.Join(skillDir, "SKILL.md"), []byte(result)); err != nil {
			return err
		}
	}

	return nil
}

// executeTemplate はテンプレートを実行
func (g *Generator) executeTemplate(tmplContent string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf []byte
	writer := &byteWriter{buf: &buf}
	if err := tmpl.Execute(writer, data); err != nil {
		return "", err
	}

	return string(buf), nil
}

// byteWriter は []byte への Writer
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// EnsureClaudeDir は .claude ディレクトリを確認・作成
func (g *Generator) EnsureClaudeDir() error {
	return os.MkdirAll(g.claudePath, 0755)
}
