package cmd

import (
	"fmt"
	"strings"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// add コマンドのフラグ
var (
	addParentID string

	// 10 概念モデル共通
	addDescription string
	addOwner       string
	addTags        []string

	// Vision 用
	addStatement       string
	addSuccessCriteria []string

	// Objective/UseCase 参照用
	addObjectiveID string

	// Consideration 用
	addDueDate string // Consideration 専用

	// Decision 用
	addConsiderationID string
	addSelectedOptID   string
	addSelectedTitle   string
	addRationale       string

	// Problem 用
	addSeverity string

	// Risk 用
	addProbability string
	addImpact      string

	// Constraint 用
	addCategory      string
	addNonNegotiable bool

	// Quality 用
	addMetrics []string

	// Actor 用
	addActorType string

	// UseCase 用
	addActorID       string
	addActorRole     string
	addUseCaseStatus string
	addSubsystemID   string

	// Activity 用（Task/Activity 統合）
	addActivityUseCaseID string
)

var addCmd = &cobra.Command{
	Use:   "add <entity> <name>",
	Short: "エンティティを追加",
	Long: `エンティティを追加します。

対応エンティティ:
  vision        プロジェクトビジョン
  objective     目標・マイルストーン
  consideration 検討事項
  decision      意思決定（イミュータブル）
  problem       問題
  risk          リスク（スコア自動計算）
  assumption    前提条件
  constraint    制約条件
  quality       品質基準
  actor         UML アクター
  usecase       UML ユースケース
  subsystem     UML サブシステム（ユースケース分類）
  activity      アクティビティ（作業単位 + プロセス可視化）

共通オプション:
  --description  説明
  --owner        オーナー
  --tags         タグ（カンマ区切り）

Activity 用オプション:
  --usecase     紐づく UseCase の ID

Vision 用オプション:
  --statement         ビジョンステートメント
  --success-criteria  成功基準（カンマ区切り）

Objective 用オプション:
  --parent    親 Objective の ID

Consideration 用オプション:
  --objective     紐づく Objective の ID
  --due           期限日

Decision 用オプション:
  --consideration     紐づく Consideration の ID（必須）
  --selected-opt-id   選択した Option の ID（必須）
  --selected-title    選択した Option のタイトル（必須）
  --rationale         選択理由（必須）

Problem 用オプション:
  --severity      深刻度（critical, high, medium, low）
  --objective     紐づく Objective の ID

Risk 用オプション:
  --probability   発生確率（high, medium, low）
  --impact        影響度（critical, high, medium, low）
  --objective     紐づく Objective の ID

Assumption 用オプション:
  --objective     紐づく Objective の ID

Constraint 用オプション:
  --category        カテゴリ（technical, business, legal, resource）
  --non-negotiable  交渉不可フラグ

Quality 用オプション:
  --objective     紐づく Objective の ID
  --metric        メトリクス（name:target[:unit] 形式、複数回指定可）

Actor 用オプション:
  --type          アクタータイプ（human, system, time, device, external）

UseCase 用オプション:
  --objective     紐づく Objective の ID（必須）
  --actor         紐づく Actor の ID
  --actor-role    アクターのロール（primary, secondary）
  --status        ステータス（draft, active, deprecated）
  --subsystem     紐づくサブシステムの ID

Subsystem 用オプション:
  --description   説明

例:
  zeus add vision "AI駆動PM" --statement "AIと人間が協調するPM"
  zeus add objective "認証システム実装"
  zeus add consideration "認証方式の選択" --objective obj-001
  zeus add decision "JWT認証を採用" --consideration con-001 --selected-opt-id opt-1 --selected-title "JWT" --rationale "セキュリティと拡張性"
  zeus add problem "パフォーマンス問題" --severity high --objective obj-001
  zeus add risk "外部API依存" --probability medium --impact high
  zeus add assumption "ユーザー数1000人以下" --objective obj-001
  zeus add constraint "外部DB不使用" --category technical --non-negotiable
  zeus add quality "コードカバレッジ" --objective obj-001 --metric "coverage:80:%" --metric "performance:100:ms"
  zeus add actor "管理者" --type human
  zeus add usecase "ログイン" --objective obj-001 --actor actor-001 --actor-role primary --subsystem sub-12345678
  zeus add subsystem "認証システム" --description "ユーザー認証関連のユースケース"
  zeus add activity "API設計" --usecase uc-001`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	// 共通フラグ（10 概念モデル）
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "説明")
	addCmd.Flags().StringVar(&addOwner, "owner", "", "オーナー")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "タグ（カンマ区切り）")

	// 共通フラグ
	addCmd.Flags().StringVarP(&addParentID, "parent", "p", "", "親 Objective の ID")

	// Vision 用フラグ
	addCmd.Flags().StringVar(&addStatement, "statement", "", "ビジョンステートメント")
	addCmd.Flags().StringSliceVar(&addSuccessCriteria, "success-criteria", nil, "成功基準（カンマ区切り）")

	// Objective 参照用フラグ
	addCmd.Flags().StringVar(&addObjectiveID, "objective", "", "紐づく Objective の ID")

	// Consideration 用フラグ
	addCmd.Flags().StringVar(&addDueDate, "due", "", "期限日（Consideration 用）")

	// Decision 用フラグ
	addCmd.Flags().StringVar(&addConsiderationID, "consideration", "", "紐づく Consideration の ID")
	addCmd.Flags().StringVar(&addSelectedOptID, "selected-opt-id", "", "選択した Option の ID")
	addCmd.Flags().StringVar(&addSelectedTitle, "selected-title", "", "選択した Option のタイトル")
	addCmd.Flags().StringVar(&addRationale, "rationale", "", "選択理由")

	// Problem 用フラグ
	addCmd.Flags().StringVar(&addSeverity, "severity", "", "深刻度（critical, high, medium, low）")

	// Risk 用フラグ
	addCmd.Flags().StringVar(&addProbability, "probability", "", "発生確率（high, medium, low）")
	addCmd.Flags().StringVar(&addImpact, "impact", "", "影響度（critical, high, medium, low）")

	// Constraint 用フラグ
	addCmd.Flags().StringVar(&addCategory, "category", "", "カテゴリ（technical, business, legal, resource）")
	addCmd.Flags().BoolVar(&addNonNegotiable, "non-negotiable", false, "交渉不可フラグ")

	// Quality 用フラグ
	addCmd.Flags().StringArrayVar(&addMetrics, "metric", nil, "メトリクス（name:target[:unit] 形式、複数回指定可）")

	// Actor 用フラグ
	addCmd.Flags().StringVar(&addActorType, "type", "", "アクタータイプ（human, system, time, device, external）")

	// UseCase 用フラグ
	addCmd.Flags().StringVar(&addActorID, "actor", "", "紐づく Actor の ID")
	addCmd.Flags().StringVar(&addActorRole, "actor-role", "", "アクターのロール（primary, secondary）")
	addCmd.Flags().StringVar(&addUseCaseStatus, "status", "", "ステータス（draft, active, deprecated）")
	addCmd.Flags().StringVar(&addSubsystemID, "subsystem", "", "紐づくサブシステムの ID")

	// Activity 用フラグ（Task/Activity 統合）
	addCmd.Flags().StringVar(&addActivityUseCaseID, "usecase", "", "紐づく UseCase の ID")
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
	case "vision":
		opts = buildVisionOptions()
	case "objective":
		opts = buildObjectiveOptions()
	case "consideration":
		opts = buildConsiderationOptions()
	case "decision":
		opts = buildDecisionOptions()
	case "problem":
		opts = buildProblemOptions()
	case "risk":
		opts = buildRiskOptions()
	case "assumption":
		opts = buildAssumptionOptions()
	case "constraint":
		opts = buildConstraintOptions()
	case "quality":
		opts = buildQualityOptions()
	case "actor":
		opts = buildActorOptions()
	case "usecase":
		opts = buildUseCaseOptions()
	case "subsystem":
		opts = buildSubsystemOptions()
	case "activity":
		opts = buildActivityOptions()
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
	if addOwner != "" {
		opts = append(opts, core.WithObjectiveOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithObjectiveTags(addTags))
	}

	return opts
}

// buildConsiderationOptions は Consideration 用オプションを構築
func buildConsiderationOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addObjectiveID != "" {
		opts = append(opts, core.WithConsiderationObjective(addObjectiveID))
	}
	if addDueDate != "" {
		opts = append(opts, core.WithConsiderationDueDate(addDueDate))
	}
	if addDescription != "" {
		opts = append(opts, core.WithConsiderationContext(addDescription))
	}

	return opts
}

// buildDecisionOptions は Decision 用オプションを構築
func buildDecisionOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addConsiderationID != "" {
		opts = append(opts, core.WithDecisionConsideration(addConsiderationID))
	}
	if addSelectedOptID != "" && addSelectedTitle != "" {
		selected := core.SelectedOption{
			OptionID: addSelectedOptID,
			Title:    addSelectedTitle,
		}
		opts = append(opts, core.WithDecisionSelected(selected))
	}
	if addRationale != "" {
		opts = append(opts, core.WithDecisionRationale(addRationale))
	}
	if addOwner != "" {
		opts = append(opts, core.WithDecisionDecidedBy(addOwner))
	}

	return opts
}

// buildProblemOptions は Problem 用オプションを構築
func buildProblemOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addSeverity != "" {
		var severity core.ProblemSeverity
		switch addSeverity {
		case "critical":
			severity = core.ProblemSeverityCritical
		case "high":
			severity = core.ProblemSeverityHigh
		case "medium":
			severity = core.ProblemSeverityMedium
		case "low":
			severity = core.ProblemSeverityLow
		default:
			severity = core.ProblemSeverityMedium
		}
		opts = append(opts, core.WithProblemSeverity(severity))
	}
	if addObjectiveID != "" {
		opts = append(opts, core.WithProblemObjective(addObjectiveID))
	}
	if addDescription != "" {
		opts = append(opts, core.WithProblemDescription(addDescription))
	}
	if addOwner != "" {
		opts = append(opts, core.WithProblemAssignedTo(addOwner))
	}

	return opts
}

// buildRiskOptions は Risk 用オプションを構築
func buildRiskOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addProbability != "" {
		var probability core.RiskProbability
		switch addProbability {
		case "high":
			probability = core.RiskProbabilityHigh
		case "medium":
			probability = core.RiskProbabilityMedium
		case "low":
			probability = core.RiskProbabilityLow
		default:
			probability = core.RiskProbabilityMedium
		}
		opts = append(opts, core.WithRiskProbability(probability))
	}
	if addImpact != "" {
		var impact core.RiskImpact
		switch addImpact {
		case "critical":
			impact = core.RiskImpactCritical
		case "high":
			impact = core.RiskImpactHigh
		case "medium":
			impact = core.RiskImpactMedium
		case "low":
			impact = core.RiskImpactLow
		default:
			impact = core.RiskImpactMedium
		}
		opts = append(opts, core.WithRiskImpact(impact))
	}
	if addObjectiveID != "" {
		opts = append(opts, core.WithRiskObjective(addObjectiveID))
	}
	if addDescription != "" {
		opts = append(opts, core.WithRiskDescription(addDescription))
	}
	if addOwner != "" {
		opts = append(opts, core.WithRiskOwner(addOwner))
	}

	return opts
}

// buildAssumptionOptions は Assumption 用オプションを構築
func buildAssumptionOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addObjectiveID != "" {
		opts = append(opts, core.WithAssumptionObjective(addObjectiveID))
	}
	if addDescription != "" {
		opts = append(opts, core.WithAssumptionDescription(addDescription))
	}

	return opts
}

// buildConstraintOptions は Constraint 用オプションを構築
func buildConstraintOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addCategory != "" {
		var category core.ConstraintCategory
		switch addCategory {
		case "technical":
			category = core.ConstraintCategoryTechnical
		case "business":
			category = core.ConstraintCategoryBusiness
		case "legal":
			category = core.ConstraintCategoryLegal
		case "resource":
			category = core.ConstraintCategoryResource
		default:
			category = core.ConstraintCategoryTechnical
		}
		opts = append(opts, core.WithConstraintCategory(category))
	}
	if addDescription != "" {
		opts = append(opts, core.WithConstraintDescription(addDescription))
	}
	if addNonNegotiable {
		opts = append(opts, core.WithConstraintNonNegotiable(true))
	}

	return opts
}

// buildQualityOptions は Quality 用オプションを構築
func buildQualityOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addObjectiveID != "" {
		opts = append(opts, core.WithQualityObjective(addObjectiveID))
	}

	// メトリクスは name:target[:unit] 形式でパース
	// 例: "coverage:80:%" → Name: "coverage", Target: 80, Unit: "%"
	if len(addMetrics) > 0 {
		metrics := parseMetrics(addMetrics)
		if len(metrics) > 0 {
			opts = append(opts, core.WithQualityMetrics(metrics))
		}
	}

	if addOwner != "" {
		opts = append(opts, core.WithQualityReviewer(addOwner))
	}

	return opts
}

// buildActorOptions は Actor 用オプションを構築
func buildActorOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addActorType != "" {
		var actorType core.ActorType
		switch addActorType {
		case "human":
			actorType = core.ActorTypeHuman
		case "system":
			actorType = core.ActorTypeSystem
		case "time":
			actorType = core.ActorTypeTime
		case "device":
			actorType = core.ActorTypeDevice
		case "external":
			actorType = core.ActorTypeExternal
		default:
			actorType = core.ActorTypeHuman
		}
		opts = append(opts, core.WithActorType(actorType))
	}
	if addDescription != "" {
		opts = append(opts, core.WithActorDescription(addDescription))
	}
	if addOwner != "" {
		opts = append(opts, core.WithActorOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithActorTags(addTags))
	}

	return opts
}

// buildUseCaseOptions は UseCase 用オプションを構築
func buildUseCaseOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addObjectiveID != "" {
		opts = append(opts, core.WithUseCaseObjective(addObjectiveID))
	}
	if addDescription != "" {
		opts = append(opts, core.WithUseCaseDescription(addDescription))
	}
	if addActorID != "" {
		var role core.ActorRole
		switch addActorRole {
		case "primary":
			role = core.ActorRolePrimary
		case "secondary":
			role = core.ActorRoleSecondary
		default:
			role = core.ActorRolePrimary
		}
		opts = append(opts, core.WithUseCaseActor(addActorID, role))
	}
	if addUseCaseStatus != "" {
		var status core.UseCaseStatus
		switch addUseCaseStatus {
		case "draft":
			status = core.UseCaseStatusDraft
		case "active":
			status = core.UseCaseStatusActive
		case "deprecated":
			status = core.UseCaseStatusDeprecated
		default:
			status = core.UseCaseStatusDraft
		}
		opts = append(opts, core.WithUseCaseStatus(status))
	}
	if addOwner != "" {
		opts = append(opts, core.WithUseCaseOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithUseCaseTags(addTags))
	}
	if addSubsystemID != "" {
		opts = append(opts, core.WithUseCaseSubsystem(addSubsystemID))
	}

	return opts
}

// buildSubsystemOptions は Subsystem 用オプションを構築
func buildSubsystemOptions() []core.EntityOption {
	var opts []core.EntityOption

	if addDescription != "" {
		opts = append(opts, core.WithSubsystemDescription(addDescription))
	}
	if addOwner != "" {
		opts = append(opts, core.WithSubsystemOwner(addOwner))
	}
	if len(addTags) > 0 {
		opts = append(opts, core.WithSubsystemTags(addTags))
	}

	return opts
}

// buildActivityOptions は Activity 用オプションを構築（Task/Activity 統合）
func buildActivityOptions() []core.EntityOption {
	var opts []core.EntityOption

	// UseCase 参照
	if addActivityUseCaseID != "" {
		opts = append(opts, core.WithActivityUseCase(addActivityUseCaseID))
	}

	// 説明
	if addDescription != "" {
		opts = append(opts, core.WithActivityDescription(addDescription))
	}

	// オーナー
	if addOwner != "" {
		opts = append(opts, core.WithActivityOwner(addOwner))
	}

	// タグ
	if len(addTags) > 0 {
		opts = append(opts, core.WithActivityTags(addTags))
	}

	return opts
}

// parseMetrics はメトリクス文字列を QualityMetric 配列にパース
// フォーマット: name:target[:unit]
func parseMetrics(metricStrs []string) []core.QualityMetric {
	var metrics []core.QualityMetric

	for i, metricStr := range metricStrs {
		parts := strings.Split(metricStr, ":")
		if len(parts) < 2 {
			continue // 不正な形式はスキップ
		}

		name := parts[0]
		var target float64
		if _, err := fmt.Sscanf(parts[1], "%f", &target); err != nil {
			continue // パース失敗はスキップ
		}

		metric := core.QualityMetric{
			ID:     fmt.Sprintf("metric-%d", i+1),
			Name:   name,
			Target: target,
			Status: core.MetricStatusInProgress,
		}

		// オプショナルな unit
		if len(parts) >= 3 {
			metric.Unit = parts[2]
		}

		metrics = append(metrics, metric)
	}

	return metrics
}
