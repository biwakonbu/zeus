package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [suggestion-id]",
	Short: "AI提案を適用",
	Long: `zeus suggest で生成された提案を適用します。

提案IDを指定するか、--all フラグですべての提案を適用できます。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runApply,
}

var (
	applyAll    bool
	applyDryRun bool
)

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().BoolVar(&applyAll, "all", false, "すべての保留中の提案を適用")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "実際には適用せずに表示のみ")
}

func runApply(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// 引数検証
	if len(args) > 1 {
		return fmt.Errorf("複数の提案IDは指定できません。1つずつ適用してください")
	}

	if !applyAll && len(args) == 0 {
		return fmt.Errorf("提案IDを指定するか、--all フラグを使用してください")
	}

	if applyAll && len(args) > 0 {
		return fmt.Errorf("--all フラグと提案IDを同時に指定することはできません")
	}

	var suggestionID string
	if len(args) > 0 {
		suggestionID = args[0]
	}

	// 提案を適用
	result, err := zeus.ApplySuggestion(ctx, suggestionID, applyAll, applyDryRun)
	if err != nil {
		return fmt.Errorf("提案適用失敗: %w", err)
	}

	if applyDryRun {
		fmt.Println("\n[DRY-RUN] 実際には適用されません")
	}

	// 結果を表示
	if applyAll {
		fmt.Printf("\n[OK] %d件の提案を適用しました:\n", result.Applied)
		for _, id := range result.AppliedIDs {
			fmt.Printf("  - %s\n", id)
		}
	} else {
		fmt.Printf("\n[OK] 提案 %s を適用しました\n", suggestionID)
		if result.CreatedTaskID != "" {
			fmt.Printf("   新規タスクID: %s\n", result.CreatedTaskID)
		}
	}

	if result.Skipped > 0 {
		fmt.Printf("\n[SKIP] %d件の提案をスキップしました (既に適用済み)\n", result.Skipped)
	}

	if result.Failed > 0 {
		fmt.Printf("\n[FAIL] %d件の提案の適用に失敗しました:\n", result.Failed)
		for _, id := range result.FailedIDs {
			fmt.Printf("  - %s\n", id)
		}
	}

	return nil
}
