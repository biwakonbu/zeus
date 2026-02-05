package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var updateClaudeCmd = &cobra.Command{
	Use:   "update-claude",
	Short: "Claude Code 連携ファイルを最新テンプレートで再生成",
	Long: `既存の Zeus プロジェクトで .claude/ ディレクトリ内の
エージェントとスキルファイルを最新のテンプレートで再生成します。

zeus init を再実行せずに、Claude Code 連携ファイルのみを更新できます。`,
	RunE: runUpdateClaude,
}

func init() {
	rootCmd.AddCommand(updateClaudeCmd)
}

func runUpdateClaude(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	if err := zeus.UpdateClaudeFiles(ctx); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Claude Code 連携ファイルを更新しました\n", green("✓"))
	fmt.Println("  更新されたファイル:")
	fmt.Println("    .claude/agents/zeus-orchestrator.md")
	fmt.Println("    .claude/agents/zeus-planner.md")
	fmt.Println("    .claude/agents/zeus-reviewer.md")
	fmt.Println("    .claude/skills/zeus-suggest/SKILL.md")
	fmt.Println("    .claude/skills/zeus-risk-analysis/SKILL.md")
	fmt.Println("    .claude/skills/zeus-wbs-design/SKILL.md")
	fmt.Println("    .claude/skills/zeus-e2e-tester/SKILL.md")

	return nil
}
