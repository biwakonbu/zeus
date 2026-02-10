package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/biwakonbu/zeus/internal/analysis"
	"github.com/biwakonbu/zeus/internal/core"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// デフォルト設定
const (
	defaultFocusDepth = 3 // --focus 指定時のデフォルト深さ
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "依存関係グラフを表示",
	Long: `依存関係をグラフとして可視化します。

出力形式:
  text    - ASCIIアートでツリー表示（デフォルト）
  dot     - Graphviz DOT形式
  mermaid - Mermaid形式（Markdown埋め込み可能）

モード:
  (デフォルト)   - タスク依存関係グラフ
  --unified     - 統合グラフ（Activity, UseCase, Objective）

フィルタオプション（--unified モード時のみ）:
  --focus <id>  - 指定IDを中心にグラフを表示
  --depth <n>   - フォーカスからの深さ（デフォルト: 無制限）
  --types <t>   - 表示するエンティティタイプ（カンマ区切り: activity,usecase,objective）
  --layers <l>  - 表示するレイヤー（structural）
  --relations <r> - 表示する関係種別（parent,implements,contributes）
  --hide-completed - 完了済み（deprecated）を非表示
  --hide-draft     - ドラフトを非表示

例:
  zeus graph                           # TEXT形式で標準出力
  zeus graph --format=dot              # DOT形式で標準出力
  zeus graph -f mermaid -o deps.md     # Mermaid形式でファイル出力
  zeus graph --unified                 # 統合グラフを表示
  zeus graph --unified --focus act-001 # act-001 を中心に表示
  zeus graph --unified --types activity,usecase       # Activity と UseCase のみ
  zeus graph --unified --layers structural           # 構造層のみ
  zeus graph --unified --relations parent,implements # 関係種別で絞り込み`,
	RunE: runGraph,
}

var (
	graphFormat       string
	graphOutput       string
	graphUnified      bool
	graphFocus        string
	graphDepth        int
	graphTypes        string
	graphLayers       string
	graphRelations    string
	graphHideComplete bool
	graphHideDraft    bool
)

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "text", "出力形式 (text|dot|mermaid)")
	graphCmd.Flags().StringVarP(&graphOutput, "output", "o", "", "出力ファイル（省略時は標準出力）")
	graphCmd.Flags().BoolVar(&graphUnified, "unified", false, "統合グラフ（Activity, UseCase, Objective）を表示")
	graphCmd.Flags().StringVar(&graphFocus, "focus", "", "フォーカスするエンティティID")
	graphCmd.Flags().IntVar(&graphDepth, "depth", 0, "フォーカスからの深さ（0=無制限）")
	graphCmd.Flags().StringVar(&graphTypes, "types", "", "表示するエンティティタイプ（カンマ区切り）")
	graphCmd.Flags().StringVar(&graphLayers, "layers", "", "表示するレイヤー（カンマ区切り: structural）")
	graphCmd.Flags().StringVar(&graphRelations, "relations", "", "表示する関係種別（カンマ区切り）")
	graphCmd.Flags().BoolVar(&graphHideComplete, "hide-completed", false, "完了済みを非表示")
	graphCmd.Flags().BoolVar(&graphHideDraft, "hide-draft", false, "ドラフトを非表示")
}

func runGraph(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// --unified フラグ: 統合グラフを表示
	if graphUnified {
		return runUnifiedGraph(ctx, zeus)
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

// runUnifiedGraph は統合グラフを出力
func runUnifiedGraph(ctx context.Context, zeus *core.Zeus) error {
	// フィルタを構築
	filter := analysis.NewGraphFilter()

	if graphFocus != "" {
		depth := graphDepth
		if depth == 0 {
			depth = defaultFocusDepth
		}
		filter = filter.WithFocus(graphFocus, depth)
	}

	if graphTypes != "" {
		types := parseEntityTypes(graphTypes)
		filter = filter.WithIncludeTypes(types...)
	}
	if graphLayers != "" {
		layers := parseEdgeLayers(graphLayers)
		filter = filter.WithIncludeLayers(layers...)
	}
	if graphRelations != "" {
		relations := parseEdgeRelations(graphRelations)
		filter = filter.WithIncludeRelations(relations...)
	}

	if graphHideComplete {
		filter = filter.WithHideCompleted(true)
	}

	if graphHideDraft {
		filter = filter.WithHideDraft(true)
	}

	// 統合グラフを構築
	graph, err := zeus.BuildUnifiedGraph(ctx, filter)
	if err != nil {
		return fmt.Errorf("統合グラフ構築失敗: %w", err)
	}

	// ノードがない場合
	if graph.Stats.TotalNodes == 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Println(cyan("Zeus Unified Graph"))
		fmt.Println("============================================================")
		fmt.Println("[INFO] エンティティがありません。")
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
		fmt.Printf("[SUCCESS] 統合グラフを %s に出力しました。\n", graphOutput)
	} else {
		fmt.Print(output)
	}

	// 循環依存の警告
	if len(graph.Cycles) > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Println()
		fmt.Println(yellow("[WARNING] 循環依存が検出されました。"))
		for _, cycle := range graph.Cycles {
			fmt.Printf("  %s\n", strings.Join(cycle, " -> "))
		}
	}

	// 統計情報を表示
	fmt.Println()
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Total Nodes: %d\n", graph.Stats.TotalNodes)
	fmt.Printf("  Total Edges: %d\n", graph.Stats.TotalEdges)
	fmt.Printf("  Max Structural Depth: %d\n", graph.Stats.MaxStructuralDepth)
	if graph.Stats.TotalActivities > 0 {
		fmt.Printf("  Activities: %d/%d deprecated\n", graph.Stats.CompletedActivities, graph.Stats.TotalActivities)
	}

	return nil
}

// parseEntityTypes は カンマ区切りの文字列を EntityType 配列に変換
func parseEntityTypes(typesStr string) []analysis.EntityType {
	var types []analysis.EntityType
	for _, t := range strings.Split(typesStr, ",") {
		t = strings.TrimSpace(strings.ToLower(t))
		switch t {
		case "activity":
			types = append(types, analysis.EntityTypeActivity)
		case "usecase":
			types = append(types, analysis.EntityTypeUseCase)
		case "objective":
			types = append(types, analysis.EntityTypeObjective)
		}
	}
	return types
}

// parseEdgeLayers はカンマ区切り文字列を UnifiedEdgeLayer 配列に変換
func parseEdgeLayers(layersStr string) []analysis.UnifiedEdgeLayer {
	var layers []analysis.UnifiedEdgeLayer
	for _, l := range strings.Split(layersStr, ",") {
		l = strings.TrimSpace(strings.ToLower(l))
		switch l {
		case string(analysis.EdgeLayerStructural):
			layers = append(layers, analysis.EdgeLayerStructural)
		}
	}
	return layers
}

// parseEdgeRelations はカンマ区切り文字列を UnifiedEdgeRelation 配列に変換
func parseEdgeRelations(relationsStr string) []analysis.UnifiedEdgeRelation {
	var relations []analysis.UnifiedEdgeRelation
	for _, r := range strings.Split(relationsStr, ",") {
		r = strings.TrimSpace(strings.ToLower(r))
		switch r {
		case string(analysis.RelationParent):
			relations = append(relations, analysis.RelationParent)
		case string(analysis.RelationImplements):
			relations = append(relations, analysis.RelationImplements)
		case string(analysis.RelationContributes):
			relations = append(relations, analysis.RelationContributes)
		}
	}
	return relations
}
