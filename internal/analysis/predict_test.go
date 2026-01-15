package analysis

import (
	"context"
	"testing"
	"time"
)

func TestNewPredictor(t *testing.T) {
	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
	}

	history := []Snapshot{}
	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted},
	}

	predictor := NewPredictor(state, history, tasks)

	if predictor == nil {
		t.Fatal("NewPredictor returned nil")
	}

	if predictor.currentState != state {
		t.Error("expected currentState to be set")
	}
}

func TestPredictor_PredictCompletion_NoHistory(t *testing.T) {
	ctx := context.Background()

	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
	}

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusInProgress},
	}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	prediction, err := predictor.PredictCompletion(ctx)

	if err != nil {
		t.Fatalf("PredictCompletion returned error: %v", err)
	}

	// 残タスク数を確認（completedでないタスクの数）
	if prediction.RemainingTasks != 2 {
		t.Errorf("expected RemainingTasks=2, got %d", prediction.RemainingTasks)
	}

	// 履歴がないので信頼性は低い
	if prediction.HasSufficientData {
		t.Error("expected HasSufficientData=false without history")
	}

	// デフォルトベロシティで予測される
	if prediction.AverageVelocity != 2.0 {
		t.Errorf("expected AverageVelocity=2.0 (default), got %f", prediction.AverageVelocity)
	}
}

func TestPredictor_PredictCompletion_AllCompleted(t *testing.T) {
	ctx := context.Background()

	state := &ProjectState{
		Health: "Excellent",
		Summary: TaskStats{
			TotalTasks: 3,
			Completed:  3,
			InProgress: 0,
			Pending:    0,
		},
	}

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusCompleted},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusCompleted},
	}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	prediction, err := predictor.PredictCompletion(ctx)

	if err != nil {
		t.Fatalf("PredictCompletion returned error: %v", err)
	}

	// 全完了時は残タスク0
	if prediction.RemainingTasks != 0 {
		t.Errorf("expected RemainingTasks=0, got %d", prediction.RemainingTasks)
	}

	// 信頼度100%
	if prediction.ConfidenceLevel != 100 {
		t.Errorf("expected ConfidenceLevel=100, got %d", prediction.ConfidenceLevel)
	}

	// 予測完了日は今日
	today := time.Now().Format("2006-01-02")
	if prediction.EstimatedDate != today {
		t.Errorf("expected EstimatedDate=%s, got %s", today, prediction.EstimatedDate)
	}
}

func TestPredictor_PredictCompletion_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	predictor := NewPredictor(&ProjectState{}, []Snapshot{}, []TaskInfo{})
	_, err := predictor.PredictCompletion(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestPredictor_PredictRisk_NoFactors(t *testing.T) {
	ctx := context.Background()

	// リスク要因がないケース
	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  8, // 完了率80%（低完了率の閾値以上）
			InProgress: 1, // WIPが少ない
			Pending:    1,
		},
	}

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusCompleted},
	}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	risk, err := predictor.PredictRisk(ctx)

	if err != nil {
		t.Fatalf("PredictRisk returned error: %v", err)
	}

	// リスク要因がないのでLow
	if risk.OverallLevel != RiskLow {
		t.Errorf("expected RiskLow, got %s", risk.OverallLevel)
	}

	if risk.Score != 0 {
		t.Errorf("expected Score=0, got %d", risk.Score)
	}
}

func TestPredictor_PredictRisk_WithBlockedTasks(t *testing.T) {
	ctx := context.Background()

	state := &ProjectState{
		Health: "Fair",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  2,
			InProgress: 3,
			Pending:    5,
		},
	}

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusBlocked},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusBlocked},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending},
	}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	risk, err := predictor.PredictRisk(ctx)

	if err != nil {
		t.Fatalf("PredictRisk returned error: %v", err)
	}

	// ブロックタスクがあるのでリスク要因あり
	foundBlockedFactor := false
	for _, f := range risk.Factors {
		if f.Name == "Blocked Tasks" {
			foundBlockedFactor = true
			break
		}
	}

	if !foundBlockedFactor {
		t.Error("expected 'Blocked Tasks' risk factor")
	}
}

func TestPredictor_PredictRisk_HighWIP(t *testing.T) {
	ctx := context.Background()

	state := &ProjectState{
		Health: "Fair",
		Summary: TaskStats{
			TotalTasks: 20,
			Completed:  10,
			InProgress: 8, // WIP > 5
			Pending:    2,
		},
	}

	tasks := []TaskInfo{}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	risk, err := predictor.PredictRisk(ctx)

	if err != nil {
		t.Fatalf("PredictRisk returned error: %v", err)
	}

	// High WIPのリスク要因
	foundHighWIP := false
	for _, f := range risk.Factors {
		if f.Name == "High WIP" {
			foundHighWIP = true
			break
		}
	}

	if !foundHighWIP {
		t.Error("expected 'High WIP' risk factor")
	}
}

func TestPredictor_PredictRisk_LowCompletionRate(t *testing.T) {
	ctx := context.Background()

	state := &ProjectState{
		Health: "Poor",
		Summary: TaskStats{
			TotalTasks: 100,
			Completed:  10, // 完了率10%（閾値30%未満）
			InProgress: 5,
			Pending:    85,
		},
	}

	tasks := []TaskInfo{}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	risk, err := predictor.PredictRisk(ctx)

	if err != nil {
		t.Fatalf("PredictRisk returned error: %v", err)
	}

	foundLowCompletion := false
	for _, f := range risk.Factors {
		if f.Name == "Low Completion Rate" {
			foundLowCompletion = true
			break
		}
	}

	if !foundLowCompletion {
		t.Error("expected 'Low Completion Rate' risk factor")
	}
}

func TestPredictor_PredictRisk_HighRiskScore(t *testing.T) {
	ctx := context.Background()

	// 複数のリスク要因があるケース
	state := &ProjectState{
		Health: "Poor",
		Summary: TaskStats{
			TotalTasks: 20,
			Completed:  2, // 完了率10%
			InProgress: 8, // WIP > 5
			Pending:    10,
		},
	}

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusBlocked},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusBlocked},
	}

	predictor := NewPredictor(state, []Snapshot{}, tasks)
	risk, err := predictor.PredictRisk(ctx)

	if err != nil {
		t.Fatalf("PredictRisk returned error: %v", err)
	}

	// 複数要因があればMedium以上
	if risk.OverallLevel == RiskLow {
		t.Error("expected risk level to be Medium or High")
	}

	if risk.Score == 0 {
		t.Error("expected Score > 0 with risk factors")
	}
}

func TestPredictor_CalculateVelocity_NoHistory(t *testing.T) {
	ctx := context.Background()

	predictor := NewPredictor(&ProjectState{}, []Snapshot{}, []TaskInfo{})
	velocity, err := predictor.CalculateVelocity(ctx)

	if err != nil {
		t.Fatalf("CalculateVelocity returned error: %v", err)
	}

	// 履歴がない場合
	if velocity.DataPoints != 0 {
		t.Errorf("expected DataPoints=0, got %d", velocity.DataPoints)
	}

	if velocity.Trend != TrendUnknown {
		t.Errorf("expected TrendUnknown, got %s", velocity.Trend)
	}
}

func TestPredictor_CalculateVelocity_WithHistory(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	history := []Snapshot{
		{
			Timestamp: now.Format(time.RFC3339),
			State: ProjectState{
				Summary: TaskStats{Completed: 10},
			},
		},
		{
			Timestamp: now.AddDate(0, 0, -7).Format(time.RFC3339),
			State: ProjectState{
				Summary: TaskStats{Completed: 5},
			},
		},
		{
			Timestamp: now.AddDate(0, 0, -14).Format(time.RFC3339),
			State: ProjectState{
				Summary: TaskStats{Completed: 2},
			},
		},
	}

	predictor := NewPredictor(&ProjectState{}, history, []TaskInfo{})
	velocity, err := predictor.CalculateVelocity(ctx)

	if err != nil {
		t.Fatalf("CalculateVelocity returned error: %v", err)
	}

	if velocity.DataPoints != 3 {
		t.Errorf("expected DataPoints=3, got %d", velocity.DataPoints)
	}
}

func TestPredictor_CalculateVelocity_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	predictor := NewPredictor(&ProjectState{}, []Snapshot{}, []TaskInfo{})
	_, err := predictor.CalculateVelocity(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestPredictor_CalculateConfidence(t *testing.T) {
	tests := []struct {
		name           string
		historyLen     int
		wantConfidence int
	}{
		{"No history", 0, 30},
		{"1 snapshot", 1, 30},
		{"2 snapshots", 2, 50},
		{"5 snapshots", 5, 70},
		{"10 snapshots", 10, 85},
		{"15 snapshots", 15, 85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := make([]Snapshot, tt.historyLen)
			predictor := NewPredictor(&ProjectState{}, history, []TaskInfo{})
			confidence := predictor.calculateConfidence()

			if confidence != tt.wantConfidence {
				t.Errorf("expected confidence=%d, got %d", tt.wantConfidence, confidence)
			}
		})
	}
}

func TestPredictor_CalculateTrend(t *testing.T) {
	predictor := NewPredictor(&ProjectState{}, []Snapshot{}, []TaskInfo{})

	tests := []struct {
		name      string
		report    *VelocityReport
		wantTrend VelocityTrend
	}{
		{
			name:      "All zeros",
			report:    &VelocityReport{Last7Days: 0, Last14Days: 0, Last30Days: 0},
			wantTrend: TrendUnknown,
		},
		{
			name:      "Increasing",
			report:    &VelocityReport{Last7Days: 10, Last14Days: 12, Last30Days: 20},
			wantTrend: TrendIncreasing,
		},
		{
			name:      "Decreasing",
			report:    &VelocityReport{Last7Days: 2, Last14Days: 10, Last30Days: 20},
			wantTrend: TrendDecreasing,
		},
		{
			name:      "Stable",
			report:    &VelocityReport{Last7Days: 5, Last14Days: 10, Last30Days: 20},
			wantTrend: TrendStable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := predictor.calculateTrend(tt.report)
			if trend != tt.wantTrend {
				t.Errorf("expected trend=%s, got %s", tt.wantTrend, trend)
			}
		})
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{3, 3, 3},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		got := minInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMustParseTime(t *testing.T) {
	// 正常なRFC3339形式
	validTime := "2024-01-15T10:30:00Z"
	parsed := mustParseTime(validTime)
	if parsed.IsZero() {
		t.Error("expected valid time, got zero")
	}

	// 無効な形式
	invalidTime := "invalid"
	parsed = mustParseTime(invalidTime)
	if !parsed.IsZero() {
		t.Error("expected zero time for invalid input")
	}
}
