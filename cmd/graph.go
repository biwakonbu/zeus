package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/biwakonbu/zeus/internal/core"
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

オプション:
  --wbs   - 10概念モデル全体のWBS階層を表示
            Vision → Objective → Deliverable → Task の完全な階層構造

例:
  zeus graph                        # TEXT形式で標準出力
  zeus graph --format=dot           # DOT形式で標準出力
  zeus graph -f mermaid -o deps.md  # Mermaid形式でファイル出力
  zeus graph --wbs                  # WBS階層を表示
  zeus graph --wbs -f mermaid       # WBS階層をMermaid形式で表示`,
	RunE: runGraph,
}

var (
	graphFormat string
	graphOutput string
	graphWBS    bool
)

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "text", "出力形式 (text|dot|mermaid)")
	graphCmd.Flags().StringVarP(&graphOutput, "output", "o", "", "出力ファイル（省略時は標準出力）")
	graphCmd.Flags().BoolVar(&graphWBS, "wbs", false, "10概念モデル全体のWBS階層を表示")
}

func runGraph(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// --wbs フラグ: WBS 階層を表示
	if graphWBS {
		return runWBSGraph(ctx, zeus)
	}

	// 通常モード: タスク依存関係グラフを表示
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

// runWBSGraph は WBS 階層グラフを出力
func runWBSGraph(ctx context.Context, zeus *core.Zeus) error {
	wbsTree, err := zeus.BuildWBSGraph(ctx)
	if err != nil {
		return fmt.Errorf("WBS構築失敗: %w", err)
	}

	// ノードがない場合
	if wbsTree.Stats.TotalNodes == 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Println(cyan("Zeus WBS Structure"))
		fmt.Println("============================================================")
		fmt.Println("[INFO] エンティティがありません。")
		fmt.Println("============================================================")
		return nil
	}

	// 形式に応じて出力を生成
	var output string
	switch graphFormat {
	case "text":
		output = wbsTree.ToText()
	case "mermaid":
		output = wbsTree.ToMermaid()
	case "dot":
		// WBS は DOT 形式をサポートしていないため text にフォールバック
		fmt.Println("[INFO] WBS は DOT 形式をサポートしていません。TEXT 形式で出力します。")
		output = wbsTree.ToText()
	default:
		return fmt.Errorf("不明な出力形式: %s (text, mermaid のいずれかを指定してください)", graphFormat)
	}

	// 出力先に応じて出力
	if graphOutput != "" {
		if err := os.WriteFile(graphOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("ファイル出力失敗: %w", err)
		}
		fmt.Printf("[SUCCESS] WBS を %s に出力しました。\n", graphOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}
