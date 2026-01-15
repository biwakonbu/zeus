package analysis

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Predictor は予測分析を実行
type Predictor struct {
	currentState *ProjectState
	history      []Snapshot
	tasks        []TaskInfo
}

// NewPredictor は新しい Predictor を作成
func NewPredictor(state *ProjectState, history []Snapshot, tasks []TaskInfo) *Predictor {
	return &Predictor{
		currentState: state,
		history:      history,
		tasks:        tasks,
	}
}

// PredictCompletion は完了日を予測
func (p *Predictor) PredictCompletion(ctx context.Context) (*CompletionPrediction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 残タスク数を計算
	remainingTasks := 0
	for _, task := range p.tasks {
		if task.Status != TaskStatusCompleted {
			remainingTasks++
		}
	}

	// ベロシティを計算
	velocity := p.calculateWeeklyVelocity()

	prediction := &CompletionPrediction{
		RemainingTasks:    remainingTasks,
		AverageVelocity:   velocity,
		HasSufficientData: len(p.history) >= 2,
	}

	// 予測完了日を計算
	if velocity > 0 && remainingTasks > 0 {
		weeksToComplete := float64(remainingTasks) / velocity
		daysToComplete := int(math.Ceil(weeksToComplete * 7))
		completionDate := time.Now().AddDate(0, 0, daysToComplete)
		prediction.EstimatedDate = completionDate.Format("2006-01-02")

		// 信頼度と誤差範囲を計算
		prediction.ConfidenceLevel = p.calculateConfidence()
		prediction.MarginDays = p.calculateMargin(daysToComplete)
	} else if remainingTasks == 0 {
		prediction.EstimatedDate = time.Now().Format("2006-01-02")
		prediction.ConfidenceLevel = 100
		prediction.MarginDays = 0
	} else {
		prediction.EstimatedDate = "N/A"
		prediction.ConfidenceLevel = 0
		prediction.MarginDays = 0
	}

	return prediction, nil
}

// PredictRisk はリスクを予測
func (p *Predictor) PredictRisk(ctx context.Context) (*RiskPrediction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	factors := p.analyzeRiskFactors()
	score := p.calculateRiskScore(factors)

	level := RiskLow
	if score >= 70 {
		level = RiskHigh
	} else if score >= 40 {
		level = RiskMedium
	}

	return &RiskPrediction{
		OverallLevel: level,
		Factors:      factors,
		Score:        score,
	}, nil
}

// CalculateVelocity はベロシティを計算
func (p *Predictor) CalculateVelocity(ctx context.Context) (*VelocityReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	report := &VelocityReport{
		DataPoints: len(p.history),
		Trend:      TrendUnknown,
	}

	if len(p.history) < 2 {
		// 履歴が不足している場合は現在の状態のみから計算
		report.WeeklyAverage = 0
		return report, nil
	}

	// 履歴からベロシティを計算
	now := time.Now()

	report.Last7Days = p.countCompletedInPeriod(now.AddDate(0, 0, -7), now)
	report.Last14Days = p.countCompletedInPeriod(now.AddDate(0, 0, -14), now)
	report.Last30Days = p.countCompletedInPeriod(now.AddDate(0, 0, -30), now)

	// 週平均を計算（30日間のデータから）
	if report.Last30Days > 0 {
		report.WeeklyAverage = float64(report.Last30Days) / 4.0
	}

	// トレンドを計算
	report.Trend = p.calculateTrend(report)

	return report, nil
}

// calculateWeeklyVelocity は週単位のベロシティを計算
func (p *Predictor) calculateWeeklyVelocity() float64 {
	if len(p.history) < 2 {
		// 履歴がない場合は、見積もり時間ベースで推定
		// デフォルトで週2タスクと仮定
		return 2.0
	}

	// 最古と最新のスナップショットから計算
	oldest := p.history[len(p.history)-1]
	newest := p.history[0]

	oldestTime, err1 := time.Parse(time.RFC3339, oldest.Timestamp)
	newestTime, err2 := time.Parse(time.RFC3339, newest.Timestamp)

	if err1 != nil || err2 != nil {
		return 2.0
	}

	daysDiff := newestTime.Sub(oldestTime).Hours() / 24
	if daysDiff < 1 {
		return 2.0
	}

	completedDiff := newest.State.Summary.Completed - oldest.State.Summary.Completed
	if completedDiff <= 0 {
		return 0.5 // 最小ベロシティ
	}

	weeksDiff := daysDiff / 7
	if weeksDiff < 1 {
		weeksDiff = 1
	}

	return float64(completedDiff) / weeksDiff
}

// calculateConfidence は予測の信頼度を計算
func (p *Predictor) calculateConfidence() int {
	// 履歴データ量に基づいて信頼度を決定
	dataPoints := len(p.history)

	if dataPoints >= 10 {
		return 85
	} else if dataPoints >= 5 {
		return 70
	} else if dataPoints >= 2 {
		return 50
	}
	return 30
}

// calculateMargin は誤差範囲を計算
func (p *Predictor) calculateMargin(estimatedDays int) int {
	confidence := p.calculateConfidence()

	// 信頼度が低いほど誤差範囲が大きい
	marginPercent := float64(100-confidence) / 100.0
	margin := int(float64(estimatedDays) * marginPercent * 0.5)

	if margin < 1 {
		margin = 1
	}
	return margin
}

// analyzeRiskFactors はリスク要因を分析
func (p *Predictor) analyzeRiskFactors() []RiskFactor {
	factors := []RiskFactor{}

	// ブロックされたタスクの分析
	blockedCount := 0
	for _, task := range p.tasks {
		if task.Status == TaskStatusBlocked {
			blockedCount++
		}
	}

	if blockedCount > 0 {
		totalActive := len(p.tasks) - p.currentState.Summary.Completed
		if totalActive > 0 {
			blockedPercent := float64(blockedCount) / float64(totalActive) * 100
			impact := int(blockedPercent / 10)
			if impact < 1 {
				impact = 1
			}
			if impact > 10 {
				impact = 10
			}
			factors = append(factors, RiskFactor{
				Name:        "Blocked Tasks",
				Description: fmt.Sprintf("%d tasks blocked (%.0f%% of active)", blockedCount, blockedPercent),
				Impact:      impact,
			})
		}
	}

	// 完了率の分析
	if p.currentState.Summary.TotalTasks > 0 {
		completionRate := float64(p.currentState.Summary.Completed) / float64(p.currentState.Summary.TotalTasks) * 100
		if completionRate < 30 {
			factors = append(factors, RiskFactor{
				Name:        "Low Completion Rate",
				Description: fmt.Sprintf("Only %.0f%% of tasks completed", completionRate),
				Impact:      7,
			})
		}
	}

	// WIP制限の分析
	if p.currentState.Summary.InProgress > 5 {
		factors = append(factors, RiskFactor{
			Name:        "High WIP",
			Description: fmt.Sprintf("%d tasks in progress (recommended: <=5)", p.currentState.Summary.InProgress),
			Impact:      5,
		})
	}

	// 進捗停滞の分析
	if len(p.history) >= 2 {
		recent := p.history[0]
		previous := p.history[minInt(1, len(p.history)-1)]
		if recent.State.Summary.Completed == previous.State.Summary.Completed {
			factors = append(factors, RiskFactor{
				Name:        "Stalled Progress",
				Description: "No tasks completed since last snapshot",
				Impact:      6,
			})
		}
	}

	return factors
}

// calculateRiskScore はリスクスコアを計算
func (p *Predictor) calculateRiskScore(factors []RiskFactor) int {
	if len(factors) == 0 {
		return 0
	}

	totalImpact := 0
	for _, f := range factors {
		totalImpact += f.Impact
	}

	// 最大スコア100に正規化
	maxPossibleImpact := len(factors) * 10
	score := totalImpact * 100 / maxPossibleImpact

	if score > 100 {
		score = 100
	}
	return score
}

// countCompletedInPeriod は期間内に完了したタスク数を推定
func (p *Predictor) countCompletedInPeriod(start, end time.Time) int {
	// 履歴から期間内のスナップショットを抽出して差分を計算
	var startSnapshot, endSnapshot *Snapshot

	for i := range p.history {
		ts, err := time.Parse(time.RFC3339, p.history[i].Timestamp)
		if err != nil {
			continue
		}

		if ts.Before(start) || ts.Equal(start) {
			if startSnapshot == nil || ts.After(mustParseTime(startSnapshot.Timestamp)) {
				startSnapshot = &p.history[i]
			}
		}
		if ts.Before(end) || ts.Equal(end) {
			if endSnapshot == nil || ts.After(mustParseTime(endSnapshot.Timestamp)) {
				endSnapshot = &p.history[i]
			}
		}
	}

	if startSnapshot == nil || endSnapshot == nil {
		return 0
	}

	diff := endSnapshot.State.Summary.Completed - startSnapshot.State.Summary.Completed
	if diff < 0 {
		diff = 0
	}
	return diff
}

// calculateTrend はベロシティのトレンドを計算
func (p *Predictor) calculateTrend(report *VelocityReport) VelocityTrend {
	if report.Last7Days == 0 && report.Last14Days == 0 && report.Last30Days == 0 {
		return TrendUnknown
	}

	// 直近7日と7-14日を比較
	weekBefore := report.Last14Days - report.Last7Days

	if report.Last7Days > weekBefore+1 {
		return TrendIncreasing
	} else if report.Last7Days < weekBefore-1 {
		return TrendDecreasing
	}
	return TrendStable
}

// mustParseTime はタイムスタンプをパース（エラー時はゼロ値を返す）
func mustParseTime(ts string) time.Time {
	t, _ := time.Parse(time.RFC3339, ts)
	return t
}

// minInt は2つの整数のうち小さい方を返す
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
