package analysis

import (
	"context"
	"testing"
	"time"
)

// ===== NewBottleneckAnalyzer テスト =====

func TestNewBottleneckAnalyzer(t *testing.T) {
	tasks := []TaskInfo{{ID: "task-001", Title: "タスク1"}}
	objectives := []ObjectiveInfo{{ID: "obj-001", Title: "目標1"}}
	deliverables := []DeliverableInfo{{ID: "del-001", Title: "成果物1"}}
	risks := []RiskInfo{{ID: "risk-001", Title: "リスク1"}}

	analyzer := NewBottleneckAnalyzer(tasks, objectives, deliverables, risks, nil)

	if analyzer == nil {
		t.Fatal("NewBottleneckAnalyzer returned nil")
	}
	if len(analyzer.tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(analyzer.tasks))
	}
	if len(analyzer.objectives) != 1 {
		t.Errorf("expected 1 objective, got %d", len(analyzer.objectives))
	}
	if len(analyzer.deliverables) != 1 {
		t.Errorf("expected 1 deliverable, got %d", len(analyzer.deliverables))
	}
	if len(analyzer.risks) != 1 {
		t.Errorf("expected 1 risk, got %d", len(analyzer.risks))
	}
}

func TestNewBottleneckAnalyzer_WithConfig(t *testing.T) {
	config := &BottleneckAnalyzerConfig{
		StagnationDays: 30,
		OverdueDays:    5,
	}

	analyzer := NewBottleneckAnalyzer(nil, nil, nil, nil, config)

	if analyzer.config.StagnationDays != 30 {
		t.Errorf("expected StagnationDays 30, got %d", analyzer.config.StagnationDays)
	}
	if analyzer.config.OverdueDays != 5 {
		t.Errorf("expected OverdueDays 5, got %d", analyzer.config.OverdueDays)
	}
}

func TestNewBottleneckAnalyzer_DefaultConfig(t *testing.T) {
	analyzer := NewBottleneckAnalyzer(nil, nil, nil, nil, nil)

	if analyzer.config.StagnationDays != DefaultBottleneckConfig.StagnationDays {
		t.Errorf("expected default StagnationDays %d, got %d",
			DefaultBottleneckConfig.StagnationDays, analyzer.config.StagnationDays)
	}
	if analyzer.config.OverdueDays != DefaultBottleneckConfig.OverdueDays {
		t.Errorf("expected default OverdueDays %d, got %d",
			DefaultBottleneckConfig.OverdueDays, analyzer.config.OverdueDays)
	}
}

// ===== Analyze テスト =====

func TestBottleneckAnalyzer_Analyze(t *testing.T) {
	analyzer := NewBottleneckAnalyzer(nil, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
	if result.Bottlenecks == nil {
		t.Error("Bottlenecks should not be nil")
	}
}

func TestBottleneckAnalyzer_Analyze_ContextCancellation(t *testing.T) {
	analyzer := NewBottleneckAnalyzer(nil, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := analyzer.Analyze(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ===== ブロックチェーン検出テスト =====

func TestBottleneckAnalyzer_DetectBlockChains(t *testing.T) {
	// ブロックチェーンを形成するタスク
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusBlocked},
		{ID: "task-002", Title: "タスク2", Status: TaskStatusBlocked, Dependencies: []string{"task-001"}},
		{ID: "task-003", Title: "タスク3", Status: TaskStatusBlocked, Dependencies: []string{"task-002"}},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// ブロックチェーンが検出されることを確認
	blockChainCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeBlockChain {
			blockChainCount++
		}
	}

	if blockChainCount == 0 {
		t.Error("expected at least one block chain bottleneck")
	}
}

func TestBottleneckAnalyzer_DetectBlockChains_Severity(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusBlocked},
		{ID: "task-002", Title: "タスク2", Status: TaskStatusBlocked, Dependencies: []string{"task-001"}},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeBlockChain {
			if b.Severity != SeverityCritical {
				t.Errorf("expected Critical severity for block chain, got %s", b.Severity)
			}
		}
	}
}

// ===== 期限超過検出テスト =====

func TestBottleneckAnalyzer_DetectOverdues(t *testing.T) {
	// 期限超過のタスク（2日前が期限）
	// Note: タイムゾーンの境界問題を回避するため、-1日ではなく-2日を使用
	// time.Parse は UTC で解析するため、ローカルタイムとの差で -1日が 0日になる場合がある
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusInProgress, DueDate: twoDaysAgo},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	overdueCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeOverdue {
			overdueCount++
		}
	}

	if overdueCount == 0 {
		t.Error("expected at least one overdue bottleneck")
	}
}

func TestBottleneckAnalyzer_DetectOverdues_Severity(t *testing.T) {
	// Note: タイムゾーンの境界問題を回避するため、十分なマージンを持った日数を使用
	// 10日超過（Critical: > 7日）
	tenDaysAgo := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	// 5日超過（High: 2-7日）
	fiveDaysAgo := time.Now().AddDate(0, 0, -5).Format("2006-01-02")

	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusInProgress, DueDate: tenDaysAgo},
		{ID: "task-002", Title: "タスク2", Status: TaskStatusInProgress, DueDate: fiveDaysAgo},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	severities := make(map[BottleneckSeverity]int)
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeOverdue {
			severities[b.Severity]++
		}
	}

	if severities[SeverityCritical] == 0 {
		t.Error("expected at least one Critical severity overdue")
	}
	if severities[SeverityHigh] == 0 {
		t.Error("expected at least one High severity overdue")
	}
}

func TestBottleneckAnalyzer_DetectOverdues_SkipsCompleted(t *testing.T) {
	// 期限超過だが完了済みのタスク
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusCompleted, DueDate: twoDaysAgo},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeOverdue {
			t.Error("completed task should not be detected as overdue")
		}
	}
}

// ===== 長期停滞検出テスト =====

func TestBottleneckAnalyzer_DetectStagnations(t *testing.T) {
	// 15日前に更新されたタスク（デフォルト14日で停滞）
	fifteenDaysAgo := time.Now().AddDate(0, 0, -15).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusInProgress, UpdatedAt: fifteenDaysAgo},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	stagnationCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeLongStagnation {
			stagnationCount++
		}
	}

	if stagnationCount == 0 {
		t.Error("expected at least one stagnation bottleneck")
	}
}

func TestBottleneckAnalyzer_DetectStagnations_ConfigurableDays(t *testing.T) {
	// 10日前に更新されたタスク
	tenDaysAgo := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusInProgress, UpdatedAt: tenDaysAgo},
	}

	// 7日で停滞とみなす設定
	config := &BottleneckAnalyzerConfig{
		StagnationDays: 7,
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, config)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	stagnationCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeLongStagnation {
			stagnationCount++
		}
	}

	if stagnationCount == 0 {
		t.Error("expected stagnation with custom config")
	}
}

// ===== 孤立エンティティ検出テスト =====

func TestBottleneckAnalyzer_DetectIsolated_Deliverable(t *testing.T) {
	// Objective に紐づいていない Deliverable
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: ""},
	}

	analyzer := NewBottleneckAnalyzer(nil, nil, deliverables, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	isolatedCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeIsolatedEntity {
			isolatedCount++
		}
	}

	if isolatedCount == 0 {
		t.Error("expected isolated deliverable to be detected")
	}
}

func TestBottleneckAnalyzer_DetectIsolated_Task(t *testing.T) {
	// 孤立したタスク（親なし、依存関係なし、他からの依存もなし）
	tasks := []TaskInfo{
		{ID: "task-001", Title: "孤立タスク", ParentID: "", Dependencies: nil},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	isolatedCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeIsolatedEntity {
			isolatedCount++
		}
	}

	if isolatedCount == 0 {
		t.Error("expected isolated task to be detected")
	}
}

func TestBottleneckAnalyzer_DetectIsolated_TaskWithDependent(t *testing.T) {
	// 他のタスクから依存されているタスクは孤立ではない
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: "", Dependencies: nil},
		{ID: "task-002", Title: "タスク2", Dependencies: []string{"task-001"}},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeIsolatedEntity && b.Entities[0] == "task-001" {
			t.Error("task with dependent should not be detected as isolated")
		}
	}
}

// ===== 高リスク未対応検出テスト =====

func TestBottleneckAnalyzer_DetectHighRisks(t *testing.T) {
	// 未対応の高リスク（スコア 6 以上）
	risks := []RiskInfo{
		{ID: "risk-001", Title: "高リスク", Status: "identified", Score: 6},
	}

	analyzer := NewBottleneckAnalyzer(nil, nil, nil, risks, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	highRiskCount := 0
	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeHighRisk {
			highRiskCount++
		}
	}

	if highRiskCount == 0 {
		t.Error("expected high risk bottleneck to be detected")
	}
}

func TestBottleneckAnalyzer_DetectHighRisks_CriticalSeverity(t *testing.T) {
	// スコア 9 以上は Critical
	risks := []RiskInfo{
		{ID: "risk-001", Title: "最高リスク", Status: "identified", Score: 9},
	}

	analyzer := NewBottleneckAnalyzer(nil, nil, nil, risks, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeHighRisk {
			if b.Severity != SeverityCritical {
				t.Errorf("expected Critical severity for score 9+, got %s", b.Severity)
			}
		}
	}
}

func TestBottleneckAnalyzer_DetectHighRisks_SkipsMitigated(t *testing.T) {
	// 対応中のリスクは検出しない
	risks := []RiskInfo{
		{ID: "risk-001", Title: "対応中リスク", Status: "mitigating", Score: 9},
	}

	analyzer := NewBottleneckAnalyzer(nil, nil, nil, risks, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, b := range result.Bottlenecks {
		if b.Type == BottleneckTypeHighRisk {
			t.Error("mitigating risk should not be detected")
		}
	}
}

// ===== サマリーテスト =====

func TestBottleneckAnalyzer_Summary(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusInProgress, DueDate: yesterday},
	}
	risks := []RiskInfo{
		{ID: "risk-001", Title: "高リスク", Status: "identified", Score: 6},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, nil, risks, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// サマリーが正しく計算されていることを確認
	totalFromSummary := result.Summary.Critical + result.Summary.High + result.Summary.Medium + result.Summary.Warning
	if totalFromSummary != len(result.Bottlenecks) {
		t.Errorf("summary total %d does not match bottlenecks count %d",
			totalFromSummary, len(result.Bottlenecks))
	}
}

// ===== ソートテスト =====

func TestBottleneckAnalyzer_ResultsSortedBySeverity(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	fifteenDaysAgo := time.Now().AddDate(0, 0, -15).Format(time.RFC3339)

	tasks := []TaskInfo{
		// Medium: 停滞
		{ID: "task-001", Title: "停滞タスク", Status: TaskStatusInProgress, UpdatedAt: fifteenDaysAgo},
		// High: 期限超過（3日）
		{ID: "task-002", Title: "期限超過", Status: TaskStatusInProgress,
			DueDate: time.Now().AddDate(0, 0, -3).Format("2006-01-02")},
	}
	// Critical: ブロックチェーン
	blockedTasks := []TaskInfo{
		{ID: "task-003", Title: "ブロック1", Status: TaskStatusBlocked},
		{ID: "task-004", Title: "ブロック2", Status: TaskStatusBlocked, Dependencies: []string{"task-003"}},
	}
	tasks = append(tasks, blockedTasks...)

	// Warning: 孤立 Deliverable
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "孤立成果物", ObjectiveID: ""},
	}

	analyzer := NewBottleneckAnalyzer(tasks, nil, deliverables, nil, nil)
	_ = yesterday // 使用しないがコンパイルエラー回避
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 深刻度順にソートされていることを確認
	prevOrder := -1
	for _, b := range result.Bottlenecks {
		currentOrder := severityOrder(b.Severity)
		if currentOrder < prevOrder {
			t.Errorf("results not sorted by severity: %s came after higher severity", b.Severity)
		}
		prevOrder = currentOrder
	}
}

// ===== ボトルネックタイプテスト =====

func TestBottleneckType_Values(t *testing.T) {
	testCases := []struct {
		bottleneckType BottleneckType
		expected       string
	}{
		{BottleneckTypeBlockChain, "block_chain"},
		{BottleneckTypeOverdue, "overdue"},
		{BottleneckTypeLongStagnation, "long_stagnation"},
		{BottleneckTypeIsolatedEntity, "isolated_entity"},
		{BottleneckTypeHighRisk, "high_risk"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.bottleneckType) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.bottleneckType))
			}
		})
	}
}

func TestBottleneckSeverity_Values(t *testing.T) {
	testCases := []struct {
		severity BottleneckSeverity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeverityHigh, "high"},
		{SeverityMedium, "medium"},
		{SeverityWarning, "warning"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.severity) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.severity))
			}
		})
	}
}

// ===== severityOrder テスト =====

func TestSeverityOrder(t *testing.T) {
	// Critical < High < Medium < Warning の順序
	if severityOrder(SeverityCritical) >= severityOrder(SeverityHigh) {
		t.Error("Critical should have lower order than High")
	}
	if severityOrder(SeverityHigh) >= severityOrder(SeverityMedium) {
		t.Error("High should have lower order than Medium")
	}
	if severityOrder(SeverityMedium) >= severityOrder(SeverityWarning) {
		t.Error("Medium should have lower order than Warning")
	}
}
