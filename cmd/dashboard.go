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

デフォルトでブラウザが自動的に開きます。

開発モード（--dev）では CORS が有効になり、
Vite Dev Server からの API リクエストを受け付けます。`,
	Example: `  zeus dashboard
  zeus dashboard --port 3000
  zeus dashboard --no-open
  zeus dashboard --dev --port 8080`,
	RunE: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().IntP("port", "p", 8080, "ポート番号")
	dashboardCmd.Flags().Bool("no-open", false, "ブラウザを自動で開かない")
	dashboardCmd.Flags().Bool("dev", false, "開発モード（CORS 有効）")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	port, _ := cmd.Flags().GetInt("port")
	noOpen, _ := cmd.Flags().GetBool("no-open")
	devMode, _ := cmd.Flags().GetBool("dev")

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// サーバー作成（開発モード対応）
	server := dashboard.NewServerWithDevMode(zeus, port, devMode)

	// サーバー起動
	fmt.Println(cyan("Zeus Dashboard"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	if devMode {
		fmt.Printf("Mode: %s (CORS enabled)\n", yellow("Development"))
	} else {
		fmt.Println("Mode: Production")
	}

	fmt.Printf("Starting server on port %d...\n", port)

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("サーバーの起動に失敗しました: %w", err)
	}

	url := server.URL()
	fmt.Printf("\nDashboard: %s\n", green(url))

	if devMode {
		fmt.Println("\nDevelopment Tips:")
		fmt.Println("  1. Run 'npm run dev' in zeus-dashboard/ for HMR")
		fmt.Println("  2. Access http://localhost:5173 for development")
		fmt.Println("  3. API requests will be proxied to this server")
	}

	fmt.Println("\nPress Ctrl+C to stop the server")
	fmt.Println("═══════════════════════════════════════════════════════════")

	// ブラウザを開く（本番モードのみ）
	if !noOpen && !devMode {
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
