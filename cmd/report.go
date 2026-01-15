package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "プロジェクトレポートを生成",
	Long: `プロジェクトの状態を包括的なレポートとして出力します。

出力形式:
  text     - テキスト形式（デフォルト）
  html     - HTML形式（スタイル付き）
  markdown - Markdown形式（依存関係グラフ付き）

例:
  zeus report                         # TEXT形式で標準出力
  zeus report --format=html           # HTML形式で標準出力
  zeus report -f markdown -o report.md  # Markdown形式でファイル出力
  zeus report -f html -o report.html  # HTML形式でファイル出力`,
	RunE: runReport,
}

var (
	reportFormat string
	reportOutput string
)

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "text", "出力形式 (text|html|markdown)")
	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "出力ファイル（省略時は標準出力）")
}

func runReport(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// 形式を検証
	switch reportFormat {
	case "text", "html", "markdown":
		// OK
	default:
		return fmt.Errorf("不明な出力形式: %s (text, html, markdown のいずれかを指定してください)", reportFormat)
	}

	// レポートを生成
	output, err := zeus.GenerateReport(ctx, reportFormat)
	if err != nil {
		return fmt.Errorf("レポート生成失敗: %w", err)
	}

	// 出力先に応じて出力
	if reportOutput != "" {
		if err := os.WriteFile(reportOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("ファイル出力失敗: %w", err)
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s レポートを %s に出力しました。\n", green("[SUCCESS]"), reportOutput)

		// ファイル形式ごとのヒント
		switch reportFormat {
		case "html":
			fmt.Println("ブラウザで開くと、スタイル付きのレポートを表示できます。")
		case "markdown":
			fmt.Println("Mermaid 対応のビューアで開くと、依存関係グラフを可視化できます。")
		}
	} else {
		fmt.Print(output)
	}

	return nil
}
