package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/dashboard"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Web ダッシュボードを起動",
	Long: `ブラウザでプロジェクト状態を可視化する Web ダッシュボードを起動します。

ダッシュボードには以下の情報が表示されます:
  - プロジェクト概要と進捗率
  - タスク統計と一覧
  - 依存関係グラフ（Mermaid.js）
  - 予測分析（完了日、リスク、ベロシティ）

デフォルトでブラウザが自動的に開きます。`,
	Example: `  zeus dashboard
  zeus dashboard --port 3000
  zeus dashboard --no-open`,
	RunE: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().IntP("port", "p", 8080, "ポート番号")
	dashboardCmd.Flags().Bool("no-open", false, "ブラウザを自動で開かない")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	port, _ := cmd.Flags().GetInt("port")
	noOpen, _ := cmd.Flags().GetBool("no-open")

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// サーバー作成
	server := dashboard.NewServer(zeus, port)

	// サーバー起動
	fmt.Println(cyan("Zeus Dashboard"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Starting server on port %d...\n", port)

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("サーバーの起動に失敗しました: %w", err)
	}

	url := server.URL()
	fmt.Printf("\nDashboard: %s\n", green(url))
	fmt.Println("\nPress Ctrl+C to stop the server")
	fmt.Println("═══════════════════════════════════════════════════════════")

	// ブラウザを開く
	if !noOpen {
		if err := openBrowser(url); err != nil {
			fmt.Printf("Warning: ブラウザを開けませんでした: %v\n", err)
		}
	}

	// シグナル待機
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\n\nShutting down server...")

	// グレースフルシャットダウン
	shutdownCtx := context.Background()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("サーバーの停止に失敗しました: %w", err)
	}

	fmt.Println("Server stopped")
	return nil
}

// openBrowser はデフォルトブラウザで URL を開く
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
