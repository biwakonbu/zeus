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

	// 10 概念モデル共通
	addDescription string
	addOwner       string
	addTags        []string

	// Vision 用
	addStatement       string
	addSuccessCriteria []string

	// Deliverable 用
	addObjectiveID         string
	addFormat              string
	addAcceptanceCriteria  []string
)

var addCmd = &cobra.Command{
	Use:   "add <entity> <name>",
	Short: "エンティティを追加",
	Long: `エンティティを追加します。

対応エンティティ:
  task        タスク（既存）
  vision      プロジェクトビジョン
  objective   目標・マイルストーン
  deliverable 成果物

共通オプション:
  --description  説明
  --owner        オーナー
  --tags         タグ（カンマ区切り）

タスク用オプション:
  --parent    親タスクのID（WBS階層構造用）
  --start     開始日（ISO8601形式: 2026-01-17）
  --due       期限日（ISO8601形式: 2026-01-31）
  --progress  進捗率（0-100）
  --wbs       WBSコード（例: 1.2.3）
  --priority  優先度（high, medium, low）
  --assignee  担当者名

Vision 用オプション:
  --statement         ビジョンステートメント
  --success-criteria  成功基準（カンマ区切り）

Objective 用オプション:
  --parent    親 Objective の ID
  --start     開始日
  --due       期限日
  --progress  進捗率（0-100）
  --wbs       WBS コード

Deliverable 用オプション:
  --objective           紐づく Objective の ID
  --format              フォーマット（document, code, design, presentation, other）
  --acceptance-criteria 受入基準（カンマ区切り）

例:
  zeus add task "設計ドキュメント作成"
  zeus add vision "AI駆動PM" --statement "AIと人間が協調するPM"
  zeus add objective "認証システム実装" --wbs 1.1 --due 2026-02-28
  zeus add deliverable "API設計書" --objective obj-001 --format document`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	// 共通フラグ（10 概念モデル）
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "説明")
	addCmd.Flags().StringVar(&addOwner, "owner", "", "オーナー")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "タグ（カンマ区切り）")

	// Phase 6A: タスク用フラグ
	addCmd.Flags().StringVarP(&addParentID, "parent", "p", "", "親タスク/Objective ID")
	addCmd.Flags().StringVar(&addStartDate, "start", "", "開始日（ISO8601形式）")
	addCmd.Flags().StringVar(&addDueDate, "due", "", "期限日（ISO8601形式）")
	addCmd.Flags().IntVar(&addProgress, "progress", 0, "進捗率（0-100）")
	addCmd.Flags().StringVar(&addWBSCode, "wbs", "", "WBSコード（例: 1.2.3）")
	addCmd.Flags().StringVar(&addPriority, "priority", "", "優先度（high, medium, low）")
	addCmd.Flags().StringVar(&addAssignee, "assignee", "", "担当者名")

	// Vision 用フラグ
	addCmd.Flags().StringVar(&addStatement, "statement", "", "ビジョンステートメント")
	addCmd.Flags().StringSliceVar(&addSuccessCriteria, "success-criteria", nil, "成功基準（カンマ区切り）")

	// Deliverable 用フラグ
	addCmd.Flags().StringVar(&addObjectiveID, "objective", "", "紐づく Objective の ID")
	addCmd.Flags().StringVar(&addFormat, "format", "", "フォーマット（document, code, design, presentation, other）")
	addCmd.Flags().StringSliceVar(&addAcceptanceCriteria, "acceptance-criteria", nil, "受入基準（カンマ区切り）")
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	entity := args[0]
	name := args[1]

	zeus := getZeus(cmd)

	// オプションを構築（エンティティタイプに応じて）
	opts := buildAddOptions(entity)

	result, err := zeus.Add(ctx, entity, name, opts...)
	if err != nil {
		return err
	}

	// JSON 出力
	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return printJSONResult(result)
	}

	// テキスト出力
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
func buildAddOptions(entity string) []core.EntityOption {
	var opts []core.EntityOption

	switch entity {
	case "task":
		opts = buildTaskOptions()
	case "vision":
		opts = buildVisionOptions()
	case "objective":
		opts = buildObjectiveOptions()
	case "deliverable":
		opts = buildDeliverableOptions()
	}

	return opts
}

// buildTaskOptions は Task 用オプションを構築
func buildTaskOptions() []core.EntityOption {
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

// buildVisionOptions は Vision 用オプションを構築
func buildVisionOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addStatement != "" {
		opts = append(opts, core.WithVisionStatement(addStatement))
	}
	if len(addSuccessCriteria) > 0 {
		opts = append(opts, core.WithVisionSuccessCriteria(addSuccessCriteria))
	}
	if addOwner != "" {
		opts = append(opts, core.WithVisionOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithVisionTags(addTags))
	}

	return opts
}

// buildObjectiveOptions は Objective 用オプションを構築
func buildObjectiveOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addDescription != "" {
		opts = append(opts, core.WithObjectiveDescription(addDescription))
	}
	if addParentID != "" {
		opts = append(opts, core.WithObjectiveParent(addParentID))
	}
	if addStartDate != "" {
		opts = append(opts, core.WithObjectiveStartDate(addStartDate))
	}
	if addDueDate != "" {
		opts = append(opts, core.WithObjectiveDueDate(addDueDate))
	}
	if addProgress > 0 {
		opts = append(opts, core.WithObjectiveProgress(addProgress))
	}
	if addWBSCode != "" {
		opts = append(opts, core.WithObjectiveWBSCode(addWBSCode))
	}
	if addOwner != "" {
		opts = append(opts, core.WithObjectiveOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithObjectiveTags(addTags))
	}

	return opts
}

// buildDeliverableOptions は Deliverable 用オプションを構築
func buildDeliverableOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addDescription != "" {
		opts = append(opts, core.WithDeliverableDescription(addDescription))
	}
	if addObjectiveID != "" {
		opts = append(opts, core.WithDeliverableObjective(addObjectiveID))
	}
	if addFormat != "" {
		var format core.DeliverableFormat
		switch addFormat {
		case "document":
			format = core.DeliverableFormatDocument
		case "code":
			format = core.DeliverableFormatCode
		case "design":
			format = core.DeliverableFormatDesign
		case "presentation":
			format = core.DeliverableFormatPresentation
		default:
			format = core.DeliverableFormatOther
		}
		opts = append(opts, core.WithDeliverableFormat(format))
	}
	if len(addAcceptanceCriteria) > 0 {
		opts = append(opts, core.WithDeliverableAcceptanceCriteria(addAcceptanceCriteria))
	}
	if addProgress > 0 {
		opts = append(opts, core.WithDeliverableProgress(addProgress))
	}
	if addOwner != "" {
		opts = append(opts, core.WithDeliverableOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithDeliverableTags(addTags))
	}

	return opts
}
