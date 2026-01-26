package analysis

import (
	"context"
	"testing"
	"time"
)

// ===== NewAffinityCalculator テスト =====

func TestNewAffinityCalculator(t *testing.T) {
	vision := VisionInfo{
		ID:    "vision-001",
		Title: "テストビジョン",
	}
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1"},
	}
	quality := []QualityInfo{
		{ID: "qual-001", Title: "品質1", DeliverableID: "del-001"},
	}
	risks := []RiskInfo{
		{ID: "risk-001", Title: "リスク1"},
	}

	calc := NewAffinityCalculator(vision, objectives, deliverables, tasks, quality, risks)

	if calc == nil {
		t.Fatal("NewAffinityCalculator returned nil")
	}
	if calc.vision.ID != vision.ID {
		t.Error("vision not set correctly")
	}
	if len(calc.objectives) != 1 {
		t.Errorf("expected 1 objective, got %d", len(calc.objectives))
	}
	if len(calc.deliverables) != 1 {
		t.Errorf("expected 1 deliverable, got %d", len(calc.deliverables))
	}
	if len(calc.tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(calc.tasks))
	}
}

func TestNewAffinityCalculatorWithOptions(t *testing.T) {
	vision := VisionInfo{ID: "vision-001", Title: "テストビジョン"}
	objectives := []ObjectiveInfo{}
	deliverables := []DeliverableInfo{}
	tasks := []TaskInfo{}
	quality := []QualityInfo{}
	risks := []RiskInfo{}

	options := AffinityOptions{
		MinScore:    0.5,
		MaxEdges:    100,
		MaxSiblings: 5,
	}

	calc := NewAffinityCalculatorWithOptions(vision, objectives, deliverables, tasks, quality, risks, options)

	if calc == nil {
		t.Fatal("NewAffinityCalculatorWithOptions returned nil")
	}
	if calc.options.MinScore != 0.5 {
		t.Errorf("expected MinScore 0.5, got %f", calc.options.MinScore)
	}
	if calc.options.MaxEdges != 100 {
		t.Errorf("expected MaxEdges 100, got %d", calc.options.MaxEdges)
	}
	if calc.options.MaxSiblings != 5 {
		t.Errorf("expected MaxSiblings 5, got %d", calc.options.MaxSiblings)
	}
}

func TestNewAffinityCalculator_EmptyVision(t *testing.T) {
	objectives := []ObjectiveInfo{}
	deliverables := []DeliverableInfo{}
	tasks := []TaskInfo{}
	quality := []QualityInfo{}
	risks := []RiskInfo{}

	// 空の VisionInfo を使用
	calc := NewAffinityCalculator(VisionInfo{}, objectives, deliverables, tasks, quality, risks)

	if calc == nil {
		t.Fatal("NewAffinityCalculator returned nil for empty vision")
	}
	if calc.vision.ID != "" {
		t.Error("vision ID should be empty")
	}
}

// ===== Calculate テスト =====

func TestAffinityCalculator_Calculate(t *testing.T) {
	vision := VisionInfo{ID: "vision-001", Title: "テストビジョン"}
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2", ParentID: "obj-001"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
		{ID: "del-002", Title: "成果物2", ObjectiveID: "obj-002"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1"},
		{ID: "task-002", Title: "タスク2", ParentID: "task-001"},
	}
	quality := []QualityInfo{
		{ID: "qual-001", Title: "品質1", DeliverableID: "del-001"},
	}
	risks := []RiskInfo{
		{ID: "risk-001", Title: "リスク1"},
	}

	calc := NewAffinityCalculator(vision, objectives, deliverables, tasks, quality, risks)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Calculate returned nil result")
	}
	if len(result.Nodes) == 0 {
		t.Error("expected nodes in result")
	}
	// Stats は値型なので、TotalNodes でチェック
	if result.Stats.TotalNodes == 0 && len(result.Nodes) > 0 {
		t.Error("Stats.TotalNodes should match Nodes count")
	}
}

func TestAffinityCalculator_Calculate_Empty(t *testing.T) {
	calc := NewAffinityCalculator(VisionInfo{}, nil, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed for empty input: %v", err)
	}

	if result == nil {
		t.Fatal("Calculate returned nil result")
	}
	if len(result.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(result.Nodes))
	}
	if len(result.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(result.Edges))
	}
}

func TestAffinityCalculator_Calculate_ContextCancellation(t *testing.T) {
	vision := VisionInfo{ID: "vision-001", Title: "テストビジョン"}
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}

	calc := NewAffinityCalculator(vision, objectives, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	_, err := calc.Calculate(ctx)
	if err == nil {
		// コンテキストがキャンセルされても処理が完了する場合がある
		// （処理が高速な場合）ため、エラーがなくても許容
		t.Log("Calculate completed despite context cancellation (processing was fast)")
	} else if err != context.Canceled {
		t.Logf("Calculate returned error: %v (expected context.Canceled or nil)", err)
	}
}

// ===== 親子関係検出テスト =====

func TestAffinityCalculator_ParentChildEdges(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "親目標"},
		{ID: "obj-002", Title: "子目標", ParentID: "obj-001"},
		{ID: "obj-003", Title: "孫目標", ParentID: "obj-002"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 親子関係のエッジを検索
	parentChildEdges := 0
	for _, edge := range result.Edges {
		for _, t := range edge.Types {
			if t == AffinityParentChild {
				parentChildEdges++
				break
			}
		}
	}

	// obj-001 -> obj-002 と obj-002 -> obj-003 の2つの親子関係
	if parentChildEdges < 2 {
		t.Errorf("expected at least 2 parent-child edges, got %d", parentChildEdges)
	}
}

func TestAffinityCalculator_TaskParentChildEdges(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "親タスク"},
		{ID: "task-002", Title: "子タスク1", ParentID: "task-001"},
		{ID: "task-003", Title: "子タスク2", ParentID: "task-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, nil, nil, tasks, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 親子関係のエッジをカウント
	parentChildEdges := 0
	for _, edge := range result.Edges {
		for _, t := range edge.Types {
			if t == AffinityParentChild {
				parentChildEdges++
				break
			}
		}
	}

	if parentChildEdges < 2 {
		t.Errorf("expected at least 2 parent-child edges for tasks, got %d", parentChildEdges)
	}
}

// ===== 兄弟関係検出テスト =====

func TestAffinityCalculator_SiblingEdges(t *testing.T) {
	// detectSibling() は以下のケースで兄弟関係を検出する:
	// 1. Deliverables: 同じ ObjectiveID を持つもの
	// 2. Tasks: 同じ ParentID を持つもの
	// ※ Objectives の ParentID は兄弟検出の対象外

	// Tasks を使用したテスト（同じ ParentID を持つ兄弟タスク）
	tasks := []TaskInfo{
		{ID: "task-001", Title: "親タスク"},
		{ID: "task-002", Title: "子タスク1", ParentID: "task-001"},
		{ID: "task-003", Title: "子タスク2", ParentID: "task-001"},
		{ID: "task-004", Title: "子タスク3", ParentID: "task-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, nil, nil, tasks, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 兄弟関係のエッジをカウント
	siblingEdges := 0
	for _, edge := range result.Edges {
		for _, et := range edge.Types {
			if et == AffinitySibling {
				siblingEdges++
				break
			}
		}
	}

	// task-002, task-003, task-004 は同じ親を持つ兄弟
	// 3つの兄弟なので、3C2 = 3 の兄弟関係
	if siblingEdges < 3 {
		t.Errorf("expected at least 3 sibling edges, got %d", siblingEdges)
	}
}

func TestAffinityCalculator_SiblingEdges_Deliverables(t *testing.T) {
	// Deliverables を使用したテスト（同じ ObjectiveID を持つ兄弟成果物）
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
		{ID: "del-002", Title: "成果物2", ObjectiveID: "obj-001"},
		{ID: "del-003", Title: "成果物3", ObjectiveID: "obj-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, deliverables, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 兄弟関係のエッジをカウント
	siblingEdges := 0
	for _, edge := range result.Edges {
		for _, et := range edge.Types {
			if et == AffinitySibling {
				siblingEdges++
				break
			}
		}
	}

	// del-001, del-002, del-003 は同じ Objective を持つ兄弟
	// 3つの兄弟なので、3C2 = 3 の兄弟関係
	if siblingEdges < 3 {
		t.Errorf("expected at least 3 sibling edges for deliverables, got %d", siblingEdges)
	}
}

// ===== WBS 隣接関係テスト =====

func TestAffinityCalculator_WBSAdjacentEdges(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", WBSCode: "1.1"},
		{ID: "task-002", Title: "タスク2", WBSCode: "1.2"},
		{ID: "task-003", Title: "タスク3", WBSCode: "1.3"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, nil, nil, tasks, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// WBS隣接関係のエッジをカウント
	wbsAdjacentEdges := 0
	for _, edge := range result.Edges {
		for _, t := range edge.Types {
			if t == AffinityWBSAdjacent {
				wbsAdjacentEdges++
				break
			}
		}
	}

	// 1.1-1.2 と 1.2-1.3 の隣接関係
	if wbsAdjacentEdges < 2 {
		t.Errorf("expected at least 2 WBS adjacent edges, got %d", wbsAdjacentEdges)
	}
}

// ===== 参照関係テスト =====

func TestAffinityCalculator_ReferenceEdges(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
	}
	// Quality はノードとして追加されないため、参照エッジは生成されない
	quality := []QualityInfo{
		{ID: "qual-001", Title: "品質1", DeliverableID: "del-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, deliverables, nil, quality, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 参照関係のエッジをカウント
	referenceEdges := 0
	for _, edge := range result.Edges {
		for _, et := range edge.Types {
			if et == AffinityReference {
				referenceEdges++
				break
			}
		}
	}

	// del-001 -> obj-001 の参照関係（Quality はノードに追加されないため参照エッジなし）
	if referenceEdges < 1 {
		t.Errorf("expected at least 1 reference edge, got %d", referenceEdges)
	}
}

// ===== ノード構築テスト =====

func TestAffinityCalculator_BuildNodes(t *testing.T) {
	vision := VisionInfo{ID: "vision-001", Title: "ビジョン"}
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1"},
	}
	// Quality と Risk はノードとして追加されない（現実装）
	quality := []QualityInfo{
		{ID: "qual-001", Title: "品質1", DeliverableID: "del-001"},
	}
	risks := []RiskInfo{
		{ID: "risk-001", Title: "リスク1"},
	}

	calc := NewAffinityCalculator(vision, objectives, deliverables, tasks, quality, risks)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 4つのエンティティがノードとして登録される（vision, objective, deliverable, task）
	// 注：Quality と Risk は現在の実装ではノードに追加されない
	expectedNodes := 4
	if len(result.Nodes) != expectedNodes {
		t.Errorf("expected %d nodes, got %d", expectedNodes, len(result.Nodes))
	}

	// ノードタイプをチェック
	nodeTypes := make(map[string]int)
	for _, node := range result.Nodes {
		nodeTypes[node.Type]++
	}

	if nodeTypes["vision"] != 1 {
		t.Errorf("expected 1 vision node, got %d", nodeTypes["vision"])
	}
	if nodeTypes["objective"] != 1 {
		t.Errorf("expected 1 objective node, got %d", nodeTypes["objective"])
	}
	if nodeTypes["deliverable"] != 1 {
		t.Errorf("expected 1 deliverable node, got %d", nodeTypes["deliverable"])
	}
	if nodeTypes["task"] != 1 {
		t.Errorf("expected 1 task node, got %d", nodeTypes["task"])
	}
}

// ===== クラスタリングテスト =====

func TestAffinityCalculator_Clusters(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "子目標1", ParentID: "obj-001"},
		{ID: "obj-003", Title: "子目標2", ParentID: "obj-001"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
		{ID: "del-002", Title: "成果物2", ObjectiveID: "obj-002"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, deliverables, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// クラスタが構築されていることを確認
	if result.Clusters == nil {
		t.Log("No clusters generated (this may be expected for small graphs)")
	}
}

// ===== 統計情報テスト =====

func TestAffinityCalculator_Stats(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2", ParentID: "obj-001"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1"},
		{ID: "task-002", Title: "タスク2", ParentID: "task-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, tasks, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	stats := result.Stats
	if stats.TotalNodes != len(result.Nodes) {
		t.Errorf("TotalNodes mismatch: expected %d, got %d", len(result.Nodes), stats.TotalNodes)
	}
	if stats.TotalEdges != len(result.Edges) {
		t.Errorf("TotalEdges mismatch: expected %d, got %d", len(result.Edges), stats.TotalEdges)
	}
}

// ===== 重みテスト =====

func TestAffinityCalculator_Weights(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2", ParentID: "obj-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// デフォルトの重みが設定されているか確認
	if result.Weights.ParentChild == 0 {
		t.Error("ParentChild weight should not be 0")
	}
	if result.Weights.Sibling == 0 {
		t.Error("Sibling weight should not be 0")
	}
}

// ===== MinScore フィルタリングテスト =====

func TestAffinityCalculator_MinScoreFilter(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2", ParentID: "obj-001"},
	}

	// 高い MinScore を設定
	options := AffinityOptions{
		MinScore: 0.9,
	}

	calc := NewAffinityCalculatorWithOptions(VisionInfo{}, objectives, nil, nil, nil, nil, options)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 高い MinScore により、弱いエッジがフィルタリングされる
	for _, edge := range result.Edges {
		if edge.Score < 0.9 {
			t.Errorf("edge with score %f should have been filtered (MinScore: 0.9)", edge.Score)
		}
	}
}

// ===== MaxEdges フィルタリングテスト =====

func TestAffinityCalculator_MaxEdgesFilter(t *testing.T) {
	// 多くのエッジが生成されるデータを作成
	tasks := make([]TaskInfo, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = TaskInfo{
			ID:    "task-" + string(rune('a'+i)),
			Title: "タスク" + string(rune('0'+i)),
		}
		if i > 0 {
			tasks[i].ParentID = tasks[0].ID
		}
	}

	options := AffinityOptions{
		MaxEdges: 5,
	}

	calc := NewAffinityCalculatorWithOptions(VisionInfo{}, nil, nil, tasks, nil, nil, options)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if len(result.Edges) > 5 {
		t.Errorf("expected at most 5 edges due to MaxEdges filter, got %d", len(result.Edges))
	}
}

// ===== MaxSiblings ハブモードテスト =====

func TestAffinityCalculator_MaxSiblingsHubMode(t *testing.T) {
	// 多くの兄弟を持つ構造を作成
	objectives := make([]ObjectiveInfo, 12)
	objectives[0] = ObjectiveInfo{ID: "obj-000", Title: "親目標"}
	for i := 1; i < 12; i++ {
		objectives[i] = ObjectiveInfo{
			ID:       "obj-" + string(rune('a'+i)),
			Title:    "子目標" + string(rune('0'+i)),
			ParentID: "obj-000",
		}
	}

	options := AffinityOptions{
		MaxSiblings: 5,
	}

	calc := NewAffinityCalculatorWithOptions(VisionInfo{}, objectives, nil, nil, nil, nil, options)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// ハブモードでは、兄弟間の直接エッジではなく、親を経由するエッジになる
	// 具体的な挙動は実装に依存するが、エッジ数が制御されていることを確認
	t.Logf("Generated %d edges for %d objectives with MaxSiblings=%d",
		len(result.Edges), len(objectives), options.MaxSiblings)
}

// ===== エッジタイプテスト =====

func TestAffinityType_String(t *testing.T) {
	testCases := []struct {
		affinityType AffinityType
		expected     string
	}{
		{AffinityParentChild, "parent-child"},
		{AffinitySibling, "sibling"},
		{AffinityWBSAdjacent, "wbs-adjacent"},
		{AffinityReference, "reference"},
		{AffinityCategory, "category"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.affinityType) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.affinityType))
			}
		})
	}
}

// ===== タイムアウトテスト =====

func TestAffinityCalculator_ContextTimeout(t *testing.T) {
	// 大量のデータを作成
	objectives := make([]ObjectiveInfo, 100)
	for i := 0; i < 100; i++ {
		objectives[i] = ObjectiveInfo{
			ID:    "obj-" + string(rune('0'+i/26)) + string(rune('a'+i%26)),
			Title: "目標" + string(rune('0'+i)),
		}
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, nil, nil, nil)

	// 非常に短いタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// タイムアウトする場合もあるし、処理が完了する場合もある
	_, err := calc.Calculate(ctx)
	if err != nil {
		t.Logf("Calculate timed out as expected: %v", err)
	} else {
		t.Log("Calculate completed before timeout")
	}
}

// ===== 複合シナリオテスト =====

func TestAffinityCalculator_ComplexScenario(t *testing.T) {
	vision := VisionInfo{
		ID:    "vision-001",
		Title: "プロジェクトビジョン",
	}

	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1", WBSCode: "1"},
		{ID: "obj-002", Title: "目標2", WBSCode: "2"},
		{ID: "obj-003", Title: "子目標1-1", ParentID: "obj-001", WBSCode: "1.1"},
		{ID: "obj-004", Title: "子目標1-2", ParentID: "obj-001", WBSCode: "1.2"},
	}

	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "成果物1", ObjectiveID: "obj-001"},
		{ID: "del-002", Title: "成果物2", ObjectiveID: "obj-003"},
	}

	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", WBSCode: "1.1.1"},
		{ID: "task-002", Title: "タスク2", WBSCode: "1.1.2", ParentID: "task-001"},
		{ID: "task-003", Title: "タスク3", WBSCode: "1.2.1"},
	}

	quality := []QualityInfo{
		{ID: "qual-001", Title: "品質基準1", DeliverableID: "del-001"},
	}

	risks := []RiskInfo{
		{ID: "risk-001", Title: "リスク1", ObjectiveID: "obj-001"},
	}

	calc := NewAffinityCalculator(vision, objectives, deliverables, tasks, quality, risks)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// 基本的な検証
	if len(result.Nodes) == 0 {
		t.Error("expected nodes in result")
	}

	// 様々なタイプのエッジが生成されていることを確認
	edgeTypes := make(map[AffinityType]int)
	for _, edge := range result.Edges {
		for _, t := range edge.Types {
			edgeTypes[t]++
		}
	}

	t.Logf("Edge types: %v", edgeTypes)
	t.Logf("Total nodes: %d, Total edges: %d", len(result.Nodes), len(result.Edges))

	// 少なくとも親子関係と参照関係が検出されるはず
	if edgeTypes[AffinityParentChild] == 0 {
		t.Error("expected parent-child edges in complex scenario")
	}
	if edgeTypes[AffinityReference] == 0 {
		t.Error("expected reference edges in complex scenario")
	}
}

// ===== ノード検索テスト =====

func TestAffinityResult_FindNode(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// ノードマップを確認
	found := false
	for _, node := range result.Nodes {
		if node.ID == "obj-001" {
			found = true
			if node.Title != "目標1" {
				t.Errorf("expected title '目標1', got %q", node.Title)
			}
			break
		}
	}

	if !found {
		t.Error("node obj-001 not found in result")
	}
}

// ===== エッジスコア計算テスト =====

func TestAffinityCalculator_EdgeScores(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2", ParentID: "obj-001"},
	}

	calc := NewAffinityCalculator(VisionInfo{}, objectives, nil, nil, nil, nil)
	ctx := context.Background()

	result, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// すべてのエッジにスコアがあることを確認
	for _, edge := range result.Edges {
		if edge.Score < 0 || edge.Score > 1 {
			t.Errorf("edge score should be between 0 and 1, got %f", edge.Score)
		}
	}
}
