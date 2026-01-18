package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var listCmd = &cobra.Command{
	Use:   "list [entity]",
	Short: "エンティティ一覧を表示",
	Long: `エンティティの一覧を表示します。

対応エンティティ:
  task           タスク
  tasks          タスク（複数形）
  vision         ビジョン
  objective(s)   目標
  deliverable(s) 成果物
  consideration(s) 検討事項
  decision(s)    意思決定
  problem(s)     問題
  risk(s)        リスク
  assumption(s)  前提条件
  constraint(s)  制約条件
  quality        品質基準

エンティティを省略すると全タスクを表示します。

例:
  zeus list              # 全タスクを表示
  zeus list tasks        # タスク一覧
  zeus list vision       # ビジョンを表示
  zeus list objectives   # 目標一覧
  zeus list deliverables # 成果物一覧
  zeus list considerations # 検討事項一覧
  zeus list decisions    # 意思決定一覧
  zeus list problems     # 問題一覧
  zeus list risks        # リスク一覧
  zeus list assumptions  # 前提条件一覧
  zeus list constraints  # 制約条件一覧
  zeus list quality      # 品質基準一覧`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("status", "s", "", "ステータスでフィルタ")
}

func runList(cmd *cobra.Command, args []string) error {
	entity := ""
	if len(args) > 0 {
		entity = args[0]
	}

	zeus := getZeus(cmd)

	// エンティティタイプに応じて表示を分岐
	switch entity {
	case "vision":
		return listVision(cmd, zeus)
	case "objective", "objectives":
		return listObjectives(cmd, zeus)
	case "deliverable", "deliverables":
		return listDeliverables(cmd, zeus)
	case "consideration", "considerations":
		return listConsiderations(cmd, zeus)
	case "decision", "decisions":
		return listDecisions(cmd, zeus)
	case "problem", "problems":
		return listProblems(cmd, zeus)
	case "risk", "risks":
		return listRisks(cmd, zeus)
	case "assumption", "assumptions":
		return listAssumptions(cmd, zeus)
	case "constraint", "constraints":
		return listConstraints(cmd, zeus)
	case "quality", "qualities":
		return listQualities(cmd, zeus)
	default:
		// Task（既存の振る舞い）
		return listTasks(cmd, zeus, entity)
	}
}

// listTasks は Task 一覧を表示
func listTasks(cmd *cobra.Command, zeus *core.Zeus, entity string) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, entity)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Tasks"), result.Total)
	fmt.Println("────────────────────────────────────────")

	for _, task := range result.Items {
		statusColor := getStatusColor(task.Status)
		fmt.Printf("[%s] %s - %s\n", statusColor(string(task.Status)), task.ID, task.Title)
	}

	return nil
}

// listVision は Vision を表示
func listVision(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "vision")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if result.Total == 0 {
		fmt.Printf("%s\n", cyan("Vision"))
		fmt.Println("────────────────────────────────────────")
		fmt.Println("ビジョンが設定されていません。")
		fmt.Println("'zeus add vision \"タイトル\"' で作成できます。")
		return nil
	}

	// Vision を取得して詳細表示
	entity, err := zeus.Get(ctx, "vision", "vision-001")
	if err != nil {
		return err
	}

	vision, ok := entity.(*core.Vision)
	if !ok {
		return fmt.Errorf("invalid vision type")
	}

	fmt.Printf("%s\n", cyan("Vision"))
	fmt.Println("════════════════════════════════════════")
	fmt.Printf("ID:        %s\n", vision.ID)
	fmt.Printf("Title:     %s\n", green(vision.Title))
	if vision.Statement != "" {
		fmt.Printf("Statement: %s\n", vision.Statement)
	}
	fmt.Printf("Status:    %s\n", vision.Status)
	if len(vision.SuccessCriteria) > 0 {
		fmt.Println("Success Criteria:")
		for _, c := range vision.SuccessCriteria {
			fmt.Printf("  - %s\n", c)
		}
	}

	return nil
}

// listObjectives は Objective 一覧を表示
func listObjectives(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "objective")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Objectives"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("目標がありません。")
		fmt.Println("'zeus add objective \"タイトル\"' で作成できます。")
		return nil
	}

	// Objective ハンドラーから直接取得
	handler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		return fmt.Errorf("objective handler not found")
	}

	// List を呼び、Total の数だけ個別に Get して表示
	// Note: List は Total を返すが Items は空（Task 互換性のため）
	// TODO: 将来的には List が []Entity を返すよう改善
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		return err
	}

	// objectives ディレクトリから直接取得
	if objHandler, ok := handler.(*core.ObjectiveHandler); ok {
		_ = objHandler // 型確認のみ
	}

	// 簡易表示（Total のみ）
	fmt.Printf("Total: %d objectives\n", listResult.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listDeliverables は Deliverable 一覧を表示
func listDeliverables(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "deliverable")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Deliverables"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("成果物がありません。")
		fmt.Println("'zeus add deliverable \"タイトル\" --objective obj-001' で作成できます。")
		return nil
	}

	// 簡易表示（Total のみ）
	fmt.Printf("Total: %d deliverables\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

func getStatusColor(status core.TaskStatus) func(a ...interface{}) string {
	switch status {
	case core.TaskStatusCompleted:
		return color.New(color.FgGreen).SprintFunc()
	case core.TaskStatusInProgress:
		return color.New(color.FgYellow).SprintFunc()
	case core.TaskStatusBlocked:
		return color.New(color.FgRed).SprintFunc()
	default:
		return color.New(color.FgWhite).SprintFunc()
	}
}

// listConsiderations は Consideration 一覧を表示
func listConsiderations(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "consideration")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Considerations"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("検討事項がありません。")
		fmt.Println("'zeus add consideration \"タイトル\"' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d considerations\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listDecisions は Decision 一覧を表示
func listDecisions(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "decision")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Decisions"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("意思決定がありません。")
		fmt.Println("'zeus add decision \"タイトル\" --consideration con-001 ...' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d decisions\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listProblems は Problem 一覧を表示
func listProblems(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "problem")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Problems"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("問題がありません。")
		fmt.Println("'zeus add problem \"タイトル\" --severity high' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d problems\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listRisks は Risk 一覧を表示
func listRisks(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "risk")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Risks"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("リスクがありません。")
		fmt.Println("'zeus add risk \"タイトル\" --probability medium --impact high' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d risks\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listAssumptions は Assumption 一覧を表示
func listAssumptions(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "assumption")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Assumptions"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("前提条件がありません。")
		fmt.Println("'zeus add assumption \"タイトル\"' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d assumptions\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listConstraints は Constraint 一覧を表示
func listConstraints(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "constraint")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Constraints"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("制約条件がありません。")
		fmt.Println("'zeus add constraint \"タイトル\" --category technical' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d constraints\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}

// listQualities は Quality 一覧を表示
func listQualities(cmd *cobra.Command, zeus *core.Zeus) error {
	ctx := getContext(cmd)
	result, err := zeus.List(ctx, "quality")
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Quality Criteria"), result.Total)
	fmt.Println("────────────────────────────────────────")

	if result.Total == 0 {
		fmt.Println("品質基準がありません。")
		fmt.Println("'zeus add quality \"タイトル\" --deliverable del-001' で作成できます。")
		return nil
	}

	fmt.Printf("Total: %d quality criteria\n", result.Total)
	fmt.Println("\n詳細を見るには 'zeus status' を使用してください。")

	return nil
}
