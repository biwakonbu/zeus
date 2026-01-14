package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "AIによるタスク提案を生成",
	Long: `現在のプロジェクト状態を分析し、AIがタスク提案を生成します。

提案は .zeus/suggestions/ ディレクトリに保存され、
zeus apply コマンドで適用できます。`,
	RunE: runSuggest,
}

var (
	suggestForce  bool
	suggestLimit  int
	suggestImpact string
)

func init() {
	rootCmd.AddCommand(suggestCmd)

	suggestCmd.Flags().BoolVar(&suggestForce, "force", false, "既存の提案を上書き")
	suggestCmd.Flags().IntVar(&suggestLimit, "limit", 5, "生成する提案の最大数")
	suggestCmd.Flags().StringVar(&suggestImpact, "impact", "", "影響度でフィルタ (high, medium, low)")
}

func runSuggest(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// プロジェクト状態を取得
	status, err := zeus.Status(ctx)
	if err != nil {
		return fmt.Errorf("ステータス取得失敗: %w", err)
	}

	// AI提案を生成
	suggestions, err := zeus.GenerateSuggestions(ctx, status, suggestLimit, suggestImpact)
	if err != nil {
		return fmt.Errorf("提案生成失敗: %w", err)
	}

	if len(suggestions) == 0 {
		fmt.Println("[INFO] 提案はありません。プロジェクトは順調です。")
		return nil
	}

	// 提案を表示
	fmt.Printf("\n[SUGGESTIONS] %d件の提案が生成されました:\n\n", len(suggestions))

	for i, suggestion := range suggestions {
		fmt.Printf("%d. [%s] %s\n", i+1, suggestion.Impact, suggestion.Description)
		fmt.Printf("   理由: %s\n", suggestion.Rationale)
		fmt.Printf("   ID: %s\n", suggestion.ID)
		if suggestion.Type == "new_task" && suggestion.TaskData != nil {
			fmt.Printf("   見積: %.1f時間\n", suggestion.TaskData.EstimateHours)
		}
		fmt.Println()
	}

	fmt.Printf("[HINT] 提案を適用するには: zeus apply <suggestion-id>\n")
	fmt.Printf("[HINT] すべて適用するには: zeus apply --all\n")

	return nil
}
