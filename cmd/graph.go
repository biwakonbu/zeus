package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "タスク依存関係のグラフを表示",
	Long: `タスク間の依存関係をグラフとして可視化します。

出力形式:
  text    - ASCIIアートでツリー表示（デフォルト）
  dot     - Graphviz DOT形式
  mermaid - Mermaid形式（Markdown埋め込み可能）

例:
  zeus graph                        # TEXT形式で標準出力
  zeus graph --format=dot           # DOT形式で標準出力
  zeus graph -f mermaid -o deps.md  # Mermaid形式でファイル出力`,
	RunE: runGraph,
}

var (
	graphFormat string
	graphOutput string
)

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "text", "出力形式 (text|dot|mermaid)")
	graphCmd.Flags().StringVarP(&graphOutput, "output", "o", "", "出力ファイル（省略時は標準出力）")
}

func runGraph(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// グラフを構築
	graph, err := zeus.BuildDependencyGraph(ctx)
	if err != nil {
		return fmt.Errorf("グラフ構築失敗: %w", err)
	}

	// タスクがない場合
	if graph.Stats.TotalNodes == 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Println(cyan("Zeus Dependency Graph"))
		fmt.Println("============================================================")
		fmt.Println("[INFO] タスクがありません。")
		fmt.Println("============================================================")
		return nil
	}

	// 形式に応じて出力を生成
	var output string
	switch graphFormat {
	case "text":
		output = graph.ToText()
	case "dot":
		output = graph.ToDot()
	case "mermaid":
		output = graph.ToMermaid()
	default:
		return fmt.Errorf("不明な出力形式: %s (text, dot, mermaid のいずれかを指定してください)", graphFormat)
	}

	// 出力先に応じて出力
	if graphOutput != "" {
		if err := os.WriteFile(graphOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("ファイル出力失敗: %w", err)
		}
		fmt.Printf("[SUCCESS] グラフを %s に出力しました。\n", graphOutput)
	} else {
		fmt.Print(output)
	}

	// 循環依存の警告
	if len(graph.Cycles) > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Println()
		fmt.Println(yellow("[WARNING] 循環依存が検出されました。"))
		fmt.Println("  循環依存はプロジェクトの進行を妨げる可能性があります。")
		fmt.Println("  依存関係を見直すことを推奨します。")
	}

	return nil
}
