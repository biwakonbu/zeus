package cmd

import (
	"context"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/spf13/cobra"
)

// contextKey はコンテキストキーの型
type contextKey string

const zeusContextKey contextKey = "zeus"

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

// getZeus はコンテキストからZeusインスタンスを取得（DI対応）
// テスト時はコンテキストにモックを注入可能
func getZeus(cmd *cobra.Command) *core.Zeus {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if z := ctx.Value(zeusContextKey); z != nil {
		return z.(*core.Zeus)
	}
	return core.New(".")
}

// getContext はコマンドからコンテキストを取得
func getContext(cmd *cobra.Command) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// WithZeus はZeusインスタンスをコンテキストに設定（テスト用）
func WithZeus(ctx context.Context, z *core.Zeus) context.Context {
	return context.WithValue(ctx, zeusContextKey, z)
}
