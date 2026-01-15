package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:   "explain <entity-id>",
	Short: "エンティティの詳細説明を表示",
	Long: `指定されたエンティティ（タスク、プロジェクト等）の詳細説明を生成します。

例:
  zeus explain project          # プロジェクト全体の説明
  zeus explain task-abc123      # 特定タスクの説明
  zeus explain --context        # コンテキスト情報を含む詳細説明`,
	Args: cobra.ExactArgs(1),
	RunE: runExplain,
}

func init() {
	rootCmd.AddCommand(explainCmd)
	explainCmd.Flags().Bool("context", false, "コンテキスト情報を含める")
}

func runExplain(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)
	includeContext, _ := cmd.Flags().GetBool("context")

	entityID := args[0]

	result, err := zeus.Explain(ctx, entityID, includeContext)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Println(cyan("Zeus Explain"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Entity: %s\n", white(result.EntityID))
	fmt.Printf("Type:   %s\n", result.EntityType)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  %s\n", result.Summary)
	fmt.Println()

	if result.Details != "" {
		fmt.Println("Details:")
		fmt.Printf("  %s\n", result.Details)
		fmt.Println()
	}

	if includeContext && len(result.Context) > 0 {
		fmt.Println("Context:")
		for key, value := range result.Context {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("Suggestions:")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}
