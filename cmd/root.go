package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zeus",
	Short: "AI-driven project management with god's eye view",
	Long: `Zeus は AI によるプロジェクトマネジメントを「神の視点」で
俯瞰するシステムです。上流工程（方針立案からWBS化、タイムライン設計、
仕様作成まで）を支援します。`,
}

// Execute はルートコマンドを実行
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "詳細出力")
	rootCmd.PersistentFlags().StringP("format", "f", "text", "出力形式 (text|json)")
}
