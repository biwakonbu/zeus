package analysis

import (
	"context"
	"slices"
	"sort"
	"time"
)

// BottleneckType はボトルネックの種類
type BottleneckType string

const (
	BottleneckTypeBlockChain     BottleneckType = "block_chain"     // ブロックチェーン（連鎖的にブロック）
	BottleneckTypeOverdue        BottleneckType = "overdue"         // 期限超過
	BottleneckTypeLongStagnation BottleneckType = "long_stagnation" // 長期停滞
	BottleneckTypeIsolatedEntity BottleneckType = "isolated_entity" // 孤立エンティティ
	BottleneckTypeHighRisk       BottleneckType = "high_risk"       // 高リスク未対応
)

// BottleneckSeverity はボトルネックの深刻度
type BottleneckSeverity string

const (
	SeverityCritical BottleneckSeverity = "critical" // 最も深刻
	SeverityHigh     BottleneckSeverity = "high"     // 高
	SeverityMedium   BottleneckSeverity = "medium"   // 中
	SeverityWarning  BottleneckSeverity = "warning"  // 警告
)

// Bottleneck はボトルネック情報
type Bottleneck struct {
	Type       BottleneckType     `json:"type"`
	Severity   BottleneckSeverity `json:"severity"`
	Entities   []string           `json:"entities"`   // 関連エンティティ ID リスト
	Message    string             `json:"message"`    // 日本語メッセージ
	Impact     string             `json:"impact"`     // 影響説明
	Suggestion string             `json:"suggestion"` // 解決策の提案
}

// BottleneckSummary はボトルネックのサマリー
type BottleneckSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Warning  int `json:"warning"`
}

// BottleneckAnalysis はボトルネック分析結果
type BottleneckAnalysis struct {
	Bottlenecks []Bottleneck      `json:"bottlenecks"`
	Summary     BottleneckSummary `json:"summary"`
}

// RiskInfo はリスク情報
// 使用コンテキスト:
//   - ボトルネック分析: 高リスク（Score >= 6）未対応の検出
//   - アフィニティ分析: Objective/Deliverable との関連付けによるクラスタリング
//
// core.RiskEntity からの変換時に handlers.go で生成される
type RiskInfo struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Probability   string `json:"probability"`    // low, medium, high
	Impact        string `json:"impact"`         // low, medium, high
	Score         int    `json:"score"`          // 計算されたスコア（Probability × Impact）
	Status        string `json:"status"`         // identified, mitigating, mitigated, accepted
	ObjectiveID   string `json:"objective_id"`   // 関連 Objective（Affinity クラスタリング用）
	DeliverableID string `json:"deliverable_id"` // 関連 Deliverable（Affinity クラスタリング用）
}

// QualityInfo は Quality エンティティ情報
// 使用コンテキスト:
//   - アフィニティ分析: Deliverable との関連付けによる品質基準クラスタリング
//
// core.QualityEntity からの変換時に handlers.go で生成される
// Note: QualityEntity には Status フィールドがないため、状態追跡が必要な場合は
// Gates の Pass/Fail や Metrics の達成度から派生させる
type QualityInfo struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	DeliverableID string `json:"deliverable_id"` // 関連 Deliverable（必須）
}

// BottleneckAnalyzerConfig はボトルネック分析の設定
type BottleneckAnalyzerConfig struct {
	StagnationDays int // 停滞とみなす日数（デフォルト: 14）
	OverdueDays    int // 超過日数の警告閾値（デフォルト: 0）
}

// DefaultBottleneckConfig はデフォルト設定
var DefaultBottleneckConfig = BottleneckAnalyzerConfig{
	StagnationDays: 14,
	OverdueDays:    0,
}

// BottleneckAnalyzer はボトルネック分析を行う
type BottleneckAnalyzer struct {
	tasks        []TaskInfo
	objectives   []ObjectiveInfo
	deliverables []DeliverableInfo
	risks        []RiskInfo
	config       BottleneckAnalyzerConfig
	now          time.Time
}

// NewBottleneckAnalyzer は新しい BottleneckAnalyzer を作成
func NewBottleneckAnalyzer(
	tasks []TaskInfo,
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	risks []RiskInfo,
	config *BottleneckAnalyzerConfig,
) *BottleneckAnalyzer {
	cfg := DefaultBottleneckConfig
	if config != nil {
		cfg = *config
	}
	return &BottleneckAnalyzer{
		tasks:        tasks,
		objectives:   objectives,
		deliverables: deliverables,
		risks:        risks,
		config:       cfg,
		now:          time.Now(),
	}
}

// Analyze はボトルネック分析を実行
func (b *BottleneckAnalyzer) Analyze(ctx context.Context) (*BottleneckAnalysis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &BottleneckAnalysis{
		Bottlenecks: []Bottleneck{},
		Summary:     BottleneckSummary{},
	}

	// 1. ブロックチェーン検出（最も深刻）
	blockChains := b.detectBlockChains()
	for _, chain := range blockChains {
		result.Bottlenecks = append(result.Bottlenecks, chain)
		b.countSeverity(result, chain.Severity)
	}

	// 2. 期限超過検出
	overdues := b.detectOverdues()
	for _, overdue := range overdues {
		result.Bottlenecks = append(result.Bottlenecks, overdue)
		b.countSeverity(result, overdue.Severity)
	}

	// 3. 長期停滞検出
	stagnations := b.detectStagnations()
	for _, stag := range stagnations {
		result.Bottlenecks = append(result.Bottlenecks, stag)
		b.countSeverity(result, stag.Severity)
	}

	// 4. 孤立エンティティ検出
	isolated := b.detectIsolated()
	for _, iso := range isolated {
		result.Bottlenecks = append(result.Bottlenecks, iso)
		b.countSeverity(result, iso.Severity)
	}

	// 5. 高リスク未対応検出
	highRisks := b.detectHighRisks()
	for _, risk := range highRisks {
		result.Bottlenecks = append(result.Bottlenecks, risk)
		b.countSeverity(result, risk.Severity)
	}

	// 深刻度順にソート
	sort.Slice(result.Bottlenecks, func(i, j int) bool {
		return severityOrder(result.Bottlenecks[i].Severity) < severityOrder(result.Bottlenecks[j].Severity)
	})

	return result, nil
}

// detectBlockChains はブロックチェーンを検出
func (b *BottleneckAnalyzer) detectBlockChains() []Bottleneck {
	var bottlenecks []Bottleneck

	// ブロック状態のタスクを収集
	blockedTasks := make(map[string]TaskInfo)
	for _, task := range b.tasks {
		if task.Status == TaskStatusBlocked {
			blockedTasks[task.ID] = task
		}
	}

	// タスクの依存関係を逆引きマップに（このタスクをブロックしているタスク -> このタスクに依存されているタスク）
	dependentsMap := make(map[string][]string)
	for _, task := range b.tasks {
		for _, dep := range task.Dependencies {
			dependentsMap[dep] = append(dependentsMap[dep], task.ID)
		}
	}

	// ブロックチェーンを構築
	visited := make(map[string]bool)
	for taskID, task := range blockedTasks {
		if visited[taskID] {
			continue
		}

		// このタスクから始まるチェーンを構築
		chain := b.buildBlockChain(taskID, blockedTasks, dependentsMap, visited)
		if len(chain) >= 2 {
			// 2つ以上のタスクが連鎖している場合はボトルネック
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:       BottleneckTypeBlockChain,
				Severity:   SeverityCritical,
				Entities:   chain,
				Message:    itoa(len(chain)) + " タスクが連鎖的にブロック",
				Impact:     task.Title + " の完了が遅延",
				Suggestion: chain[len(chain)-1] + " の問題を優先解決",
			})
		}
	}

	return bottlenecks
}

// buildBlockChain はブロックチェーンを再帰的に構築
func (b *BottleneckAnalyzer) buildBlockChain(
	taskID string,
	blockedTasks map[string]TaskInfo,
	dependentsMap map[string][]string,
	visited map[string]bool,
) []string {
	if visited[taskID] {
		return nil
	}
	visited[taskID] = true

	chain := []string{taskID}

	// このタスクに依存しているタスクがブロックされているか確認
	for _, dependent := range dependentsMap[taskID] {
		if _, isBlocked := blockedTasks[dependent]; isBlocked {
			subChain := b.buildBlockChain(dependent, blockedTasks, dependentsMap, visited)
			chain = append(chain, subChain...)
		}
	}

	return chain
}

// detectOverdues は期限超過を検出
func (b *BottleneckAnalyzer) detectOverdues() []Bottleneck {
	var bottlenecks []Bottleneck

	for _, task := range b.tasks {
		if task.Status == TaskStatusCompleted {
			continue // 完了済みは除外
		}
		if task.DueDate == "" {
			continue // 期限なしは除外
		}

		dueDate := b.parseDate(task.DueDate)
		if dueDate == nil {
			continue
		}

		overdueDays := int(b.now.Sub(*dueDate).Hours() / 24)
		if overdueDays > b.config.OverdueDays {
			severity := SeverityHigh
			if overdueDays > 7 {
				severity = SeverityCritical
			} else if overdueDays <= 1 {
				severity = SeverityMedium
			}

			// 影響を受ける上位エンティティを特定
			impact := b.findImpactedParent(task)

			bottlenecks = append(bottlenecks, Bottleneck{
				Type:       BottleneckTypeOverdue,
				Severity:   severity,
				Entities:   []string{task.ID},
				Message:    itoa(overdueDays) + " 日超過",
				Impact:     impact,
				Suggestion: "優先的に対応または期限を見直し",
			})
		}
	}

	return bottlenecks
}

// detectStagnations は長期停滞を検出
func (b *BottleneckAnalyzer) detectStagnations() []Bottleneck {
	var bottlenecks []Bottleneck

	for _, task := range b.tasks {
		if task.Status == TaskStatusCompleted {
			continue // 完了済みは除外
		}

		// UpdatedAt から停滞日数を計算
		updatedAt := b.parseDate(task.UpdatedAt)
		if updatedAt == nil {
			continue
		}

		stagnationDays := int(b.now.Sub(*updatedAt).Hours() / 24)
		if stagnationDays >= b.config.StagnationDays {
			severity := SeverityMedium
			if stagnationDays > 30 {
				severity = SeverityHigh
			}

			bottlenecks = append(bottlenecks, Bottleneck{
				Type:       BottleneckTypeLongStagnation,
				Severity:   severity,
				Entities:   []string{task.ID},
				Message:    itoa(stagnationDays) + " 日間ステータス変化なし",
				Impact:     task.Title + " が進行していない可能性",
				Suggestion: "状況を確認し、ブロック要因があれば対処",
			})
		}
	}

	return bottlenecks
}

// detectIsolated は孤立エンティティを検出
func (b *BottleneckAnalyzer) detectIsolated() []Bottleneck {
	var bottlenecks []Bottleneck

	// Deliverable の参照マップ
	deliverableToObjective := make(map[string]string)
	for _, del := range b.deliverables {
		deliverableToObjective[del.ID] = del.ObjectiveID
	}

	// 孤立した Deliverable を検出（Objective に紐づいていない）
	for _, del := range b.deliverables {
		if del.ObjectiveID == "" {
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:       BottleneckTypeIsolatedEntity,
				Severity:   SeverityWarning,
				Entities:   []string{del.ID},
				Message:    "Objective に紐づいていない Deliverable",
				Impact:     "成果物の目的が不明確",
				Suggestion: "適切な Objective に紐づけるか、削除を検討",
			})
		}
	}

	// 孤立したタスクを検出（親も依存関係もない）
	for _, task := range b.tasks {
		if task.ParentID == "" && len(task.Dependencies) == 0 {
			// 他のタスクからの依存も確認
			hasDependent := false
			for _, other := range b.tasks {
				if slices.Contains(other.Dependencies, task.ID) {
					hasDependent = true
					break
				}
			}

			if !hasDependent {
				bottlenecks = append(bottlenecks, Bottleneck{
					Type:       BottleneckTypeIsolatedEntity,
					Severity:   SeverityWarning,
					Entities:   []string{task.ID},
					Message:    "孤立したタスク（参照関係なし）",
					Impact:     "タスクの位置づけが不明確",
					Suggestion: "適切な親タスクまたは Deliverable に紐づけ",
				})
			}
		}
	}

	return bottlenecks
}

// detectHighRisks は高リスク未対応を検出
func (b *BottleneckAnalyzer) detectHighRisks() []Bottleneck {
	var bottlenecks []Bottleneck

	for _, risk := range b.risks {
		// 未対応の高リスクを検出
		if risk.Status == "identified" && risk.Score >= 6 {
			severity := SeverityHigh
			if risk.Score >= 9 {
				severity = SeverityCritical
			}

			bottlenecks = append(bottlenecks, Bottleneck{
				Type:       BottleneckTypeHighRisk,
				Severity:   severity,
				Entities:   []string{risk.ID},
				Message:    "高リスクが未対応（スコア: " + itoa(risk.Score) + "）",
				Impact:     risk.Title,
				Suggestion: "軽減策の検討と実施を優先",
			})
		}
	}

	return bottlenecks
}

// findImpactedParent は影響を受ける上位エンティティを特定
func (b *BottleneckAnalyzer) findImpactedParent(task TaskInfo) string {
	// 親タスクを辿って Deliverable または Objective を特定
	if task.ParentID != "" {
		for _, del := range b.deliverables {
			if del.ID == task.ParentID {
				return del.Title + " に影響"
			}
		}
		for _, obj := range b.objectives {
			if obj.ID == task.ParentID {
				return obj.Title + " の達成に影響"
			}
		}
	}
	return "プロジェクト進行に影響"
}

// parseDate は日付文字列をパース
func (b *BottleneckAnalyzer) parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}
	return nil
}

// countSeverity は深刻度をカウント
func (b *BottleneckAnalyzer) countSeverity(result *BottleneckAnalysis, severity BottleneckSeverity) {
	switch severity {
	case SeverityCritical:
		result.Summary.Critical++
	case SeverityHigh:
		result.Summary.High++
	case SeverityMedium:
		result.Summary.Medium++
	case SeverityWarning:
		result.Summary.Warning++
	}
}

// severityOrder は深刻度の順序を返す（ソート用）
func severityOrder(severity BottleneckSeverity) int {
	switch severity {
	case SeverityCritical:
		return 0
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	case SeverityWarning:
		return 3
	default:
		return 4
	}
}

// Note: itoa 関数は wbs.go で定義されているものを使用
