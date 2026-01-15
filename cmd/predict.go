package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var predictCmd = &cobra.Command{
	Use:   "predict [type]",
	Short: "プロジェクトの予測分析を表示",
	Long: `プロジェクトの進捗予測、リスク予測、ベロシティを分析します。

予測タイプ:
  completion - 完了日予測
  risk       - リスク予測
  velocity   - ベロシティ（作業速度）分析
  all        - 全ての予測（デフォルト）

例:
  zeus predict                # 全ての予測を表示
  zeus predict completion     # 完了日予測のみ
  zeus predict risk           # リスク予測のみ
  zeus predict velocity       # ベロシティ分析のみ`,
	RunE: runPredict,
}

func init() {
	rootCmd.AddCommand(predictCmd)
}

func runPredict(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// 予測タイプを決定
	predType := "all"
	if len(args) > 0 {
		predType = args[0]
	}

	// 予測を実行
	result, err := zeus.Predict(ctx, predType)
	if err != nil {
		return fmt.Errorf("予測失敗: %w", err)
	}

	// ヘッダー出力
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println(cyan("Zeus Prediction Analysis"))
	fmt.Println("============================================================")

	// 完了予測
	if result.Completion != nil {
		fmt.Println()
		fmt.Println(cyan("[COMPLETION PREDICTION]"))
		fmt.Printf("  Estimated Completion: %s\n", green(result.Completion.EstimatedDate))
		if result.Completion.MarginDays > 0 {
			fmt.Printf("  Margin:               +/- %d days\n", result.Completion.MarginDays)
		}
		fmt.Printf("  Average Velocity:     %.1f tasks/week\n", result.Completion.AverageVelocity)
		fmt.Printf("  Remaining Tasks:      %d\n", result.Completion.RemainingTasks)
		fmt.Printf("  Confidence:           %d%%\n", result.Completion.ConfidenceLevel)

		if !result.Completion.HasSufficientData {
			fmt.Println()
			fmt.Println(yellow("  [INFO] 履歴データが不足しています。予測精度は限定的です。"))
		}
	}

	// リスク予測
	if result.Risk != nil {
		fmt.Println()
		fmt.Println(cyan("[RISK PREDICTION]"))

		// リスクレベルに応じて色を変更
		riskStr := string(result.Risk.OverallLevel)
		switch result.Risk.OverallLevel {
		case "High":
			riskStr = red(riskStr)
		case "Medium":
			riskStr = yellow(riskStr)
		case "Low":
			riskStr = green(riskStr)
		}

		fmt.Printf("  Overall Risk Level:   %s\n", riskStr)
		fmt.Printf("  Risk Score:           %d/100\n", result.Risk.Score)

		if len(result.Risk.Factors) > 0 {
			fmt.Println()
			fmt.Println("  Risk Factors:")
			for _, factor := range result.Risk.Factors {
				fmt.Printf("    - %s (Impact: %d/10)\n", factor.Name, factor.Impact)
				fmt.Printf("      %s\n", factor.Description)
			}
		}
	}

	// ベロシティ分析
	if result.Velocity != nil {
		fmt.Println()
		fmt.Println(cyan("[VELOCITY ANALYSIS]"))
		fmt.Printf("  Last 7 days:          %d tasks completed\n", result.Velocity.Last7Days)
		fmt.Printf("  Last 14 days:         %d tasks completed\n", result.Velocity.Last14Days)
		fmt.Printf("  Last 30 days:         %d tasks completed\n", result.Velocity.Last30Days)
		fmt.Printf("  Weekly Average:       %.1f tasks\n", result.Velocity.WeeklyAverage)

		// トレンドに応じて色を変更
		trendStr := string(result.Velocity.Trend)
		switch result.Velocity.Trend {
		case "Increasing":
			trendStr = green(trendStr)
		case "Decreasing":
			trendStr = red(trendStr)
		case "Stable":
			trendStr = cyan(trendStr)
		}
		fmt.Printf("  Trend:                %s\n", trendStr)

		if result.Velocity.DataPoints < 5 {
			fmt.Println()
			fmt.Println(yellow("  [INFO] データポイントが少ないため、トレンド分析の精度が限定的です。"))
		}
	}

	fmt.Println()
	fmt.Println("============================================================")

	return nil
}
