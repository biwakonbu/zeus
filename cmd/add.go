package cmd

import (
	"fmt"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// add コマンドのフラグ
var (
	addParentID  string
	addStartDate string
	addDueDate   string
	addProgress  int
	addWBSCode   string
	addPriority  string
	addAssignee  string
)

var addCmd = &cobra.Command{
	Use:   "add <entity> <name>",
	Short: "エンティティを追加",
	Long: `エンティティを追加します。

タスク追加時のオプション:
  --parent    親タスクのID（WBS階層構造用）
  --start     開始日（ISO8601形式: 2026-01-17）
  --due       期限日（ISO8601形式: 2026-01-31）
  --progress  進捗率（0-100）
  --wbs       WBSコード（例: 1.2.3）
  --priority  優先度（high, medium, low）
  --assignee  担当者名

例:
  zeus add task "設計ドキュメント作成"
  zeus add task "子タスク" --parent task-abc12345
  zeus add task "実装" --start 2026-01-20 --due 2026-01-31 --priority high`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Phase 6A: タスク用フラグ
	addCmd.Flags().StringVarP(&addParentID, "parent", "p", "", "親タスクID")
	addCmd.Flags().StringVar(&addStartDate, "start", "", "開始日（ISO8601形式）")
	addCmd.Flags().StringVar(&addDueDate, "due", "", "期限日（ISO8601形式）")
	addCmd.Flags().IntVar(&addProgress, "progress", 0, "進捗率（0-100）")
	addCmd.Flags().StringVar(&addWBSCode, "wbs", "", "WBSコード（例: 1.2.3）")
	addCmd.Flags().StringVar(&addPriority, "priority", "", "優先度（high, medium, low）")
	addCmd.Flags().StringVar(&addAssignee, "assignee", "", "担当者名")
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	entity := args[0]
	name := args[1]

	zeus := getZeus(cmd)

	// オプションを構築
	opts := buildAddOptions()

	result, err := zeus.Add(ctx, entity, name, opts...)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if result.NeedsApproval {
		// 承認待ちの場合
		fmt.Printf("%s %s '%s' は承認待ちキューに追加されました\n",
			yellow("⏳"), result.Entity, name)
		fmt.Printf("   承認ID: %s\n", result.ApprovalID)
		fmt.Println("   'zeus pending' で確認、'zeus approve <id>' で承認できます")
	} else {
		// 即時追加の場合
		fmt.Printf("%s Added %s: %s (ID: %s)\n",
			green("✓"), result.Entity, name, result.ID)
	}

	return nil
}

// buildAddOptions はフラグからEntityOptionを構築
func buildAddOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addParentID != "" {
		opts = append(opts, core.WithTaskParent(addParentID))
	}
	if addStartDate != "" {
		opts = append(opts, core.WithTaskStartDate(addStartDate))
	}
	if addDueDate != "" {
		opts = append(opts, core.WithTaskDueDate(addDueDate))
	}
	if addProgress > 0 {
		opts = append(opts, core.WithTaskProgress(addProgress))
	}
	if addWBSCode != "" {
		opts = append(opts, core.WithTaskWBSCode(addWBSCode))
	}
	if addPriority != "" {
		var priority core.TaskPriority
		switch addPriority {
		case "high":
			priority = core.PriorityHigh
		case "medium":
			priority = core.PriorityMedium
		case "low":
			priority = core.PriorityLow
		default:
			priority = core.PriorityMedium
		}
		opts = append(opts, core.WithTaskPriority(priority))
	}
	if addAssignee != "" {
		opts = append(opts, core.WithTaskAssignee(addAssignee))
	}

	return opts
}
