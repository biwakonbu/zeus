package analysis

import (
	"strings"
	"testing"
)

func TestNewUnifiedGraphBuilder(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	if builder == nil {
		t.Fatal("NewUnifiedGraphBuilder returned nil")
	}
}

func TestUnifiedGraphBuilder_WithActivities(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", Mode: "simple"},
		{ID: "act-002", Title: "Activity 2", Status: "completed", Mode: "flow"},
	}
	result := builder.WithActivities(activities)
	if result == nil {
		t.Fatal("WithActivities returned nil")
	}
}

func TestUnifiedGraphBuilder_WithUseCases(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	result := builder.WithUseCases(usecases)
	if result == nil {
		t.Fatal("WithUseCases returned nil")
	}
}

func TestUnifiedGraphBuilder_WithDeliverables(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "Deliverable 1", Status: "draft"},
	}
	result := builder.WithDeliverables(deliverables)
	if result == nil {
		t.Fatal("WithDeliverables returned nil")
	}
}

func TestUnifiedGraphBuilder_WithObjectives(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	result := builder.WithObjectives(objectives)
	if result == nil {
		t.Fatal("WithObjectives returned nil")
	}
}

func TestUnifiedGraphBuilder_Build_Empty(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	graph := builder.Build()

	if graph == nil {
		t.Fatal("Build returned nil")
	}
	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(graph.Edges))
	}
}

func TestUnifiedGraphBuilder_Build_WithActivities(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", Mode: "simple"},
		{ID: "act-002", Title: "Activity 2", Status: "completed", Mode: "flow", Dependencies: []string{"act-001"}},
	}
	graph := builder.WithActivities(activities).Build()

	if graph == nil {
		t.Fatal("Build returned nil")
	}
	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// ノードの検証
	node1 := graph.Nodes["act-001"]
	if node1 == nil {
		t.Fatal("act-001 node not found")
	}
	if node1.Type != EntityTypeActivity {
		t.Errorf("expected type activity, got %s", node1.Type)
	}
	if node1.Title != "Activity 1" {
		t.Errorf("expected title 'Activity 1', got %s", node1.Title)
	}

	// 依存関係エッジの検証
	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Edges))
	}
}

func TestUnifiedGraphBuilder_Build_WithUseCaseActivityRelation(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
	}
	graph := builder.WithUseCases(usecases).WithActivities(activities).Build()

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// Activity -> UseCase の関係エッジがあるか確認（Activity が UseCase を implements）
	foundEdge := false
	for _, edge := range graph.Edges {
		if edge.From == "act-001" && edge.To == "uc-001" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Error("expected edge from act-001 to uc-001")
	}
}

func TestUnifiedGraphBuilder_Build_WithDeliverableRelation(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "Deliverable 1", Status: "draft"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", RelatedDeliverables: []string{"del-001"}},
	}
	graph := builder.WithDeliverables(deliverables).WithActivities(activities).Build()

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// Activity -> Deliverable の関係エッジがあるか確認
	foundEdge := false
	for _, edge := range graph.Edges {
		if edge.From == "act-001" && edge.To == "del-001" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Error("expected edge from act-001 to del-001")
	}
}

func TestUnifiedGraphBuilder_Build_CycleDetection(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	// 循環依存: act-001 -> act-002 -> act-003 -> act-001
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", Dependencies: []string{"act-003"}},
		{ID: "act-002", Title: "Activity 2", Status: "active", Dependencies: []string{"act-001"}},
		{ID: "act-003", Title: "Activity 3", Status: "active", Dependencies: []string{"act-002"}},
	}
	graph := builder.WithActivities(activities).Build()

	if len(graph.Cycles) == 0 {
		t.Error("expected cycles to be detected")
	}
}

func TestUnifiedGraphBuilder_Build_IsolatedNodes(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Connected", Status: "active", Dependencies: []string{"act-002"}},
		{ID: "act-002", Title: "Connected", Status: "active"},
		{ID: "act-003", Title: "Isolated", Status: "active"}, // 孤立ノード
	}
	graph := builder.WithActivities(activities).Build()

	if len(graph.Isolated) != 1 {
		t.Errorf("expected 1 isolated node, got %d", len(graph.Isolated))
	}
	if len(graph.Isolated) > 0 && graph.Isolated[0] != "act-003" {
		t.Errorf("expected isolated node act-003, got %s", graph.Isolated[0])
	}
}

func TestUnifiedGraphBuilder_WithFilter_IncludeTypes(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	filter := NewGraphFilter().WithIncludeTypes(EntityTypeActivity)
	graph := builder.
		WithActivities(activities).
		WithUseCases(usecases).
		WithFilter(filter).
		Build()

	// Activity のみが含まれるはず
	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node (activity only), got %d", len(graph.Nodes))
	}
	if _, exists := graph.Nodes["act-001"]; !exists {
		t.Error("expected act-001 to be included")
	}
}

func TestUnifiedGraphBuilder_WithFilter_HideCompleted(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Active", Status: "active"},
		{ID: "act-002", Title: "Completed", Status: "completed"},
	}
	filter := NewGraphFilter().WithHideCompleted(true)
	graph := builder.
		WithActivities(activities).
		WithFilter(filter).
		Build()

	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node (active only), got %d", len(graph.Nodes))
	}
	if _, exists := graph.Nodes["act-001"]; !exists {
		t.Error("expected act-001 to be included")
	}
}

func TestUnifiedGraphBuilder_WithFilter_HideDraft(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Active", Status: "active"},
		{ID: "act-002", Title: "Draft", Status: "draft"},
	}
	filter := NewGraphFilter().WithHideDraft(true)
	graph := builder.
		WithActivities(activities).
		WithFilter(filter).
		Build()

	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node (non-draft only), got %d", len(graph.Nodes))
	}
}

func TestUnifiedGraphBuilder_WithFilter_Focus(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	// act-001 -> act-002 -> act-003 -> act-004
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "active", Dependencies: []string{"act-001"}},
		{ID: "act-003", Title: "Activity 3", Status: "active", Dependencies: []string{"act-002"}},
		{ID: "act-004", Title: "Activity 4", Status: "active", Dependencies: []string{"act-003"}},
	}
	// act-002 から深さ 1 でフォーカス
	filter := NewGraphFilter().WithFocus("act-002", 1).WithHideUnrelated(true)
	graph := builder.
		WithActivities(activities).
		WithFilter(filter).
		Build()

	// act-001, act-002, act-003 のみが含まれるはず（深さ1以内）
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes (depth 1 from act-002), got %d", len(graph.Nodes))
		for id := range graph.Nodes {
			t.Logf("  - %s", id)
		}
	}
}

func TestUnifiedGraph_Stats(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "completed"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	graph := builder.WithActivities(activities).WithUseCases(usecases).Build()

	if graph.Stats.TotalNodes != 3 {
		t.Errorf("expected 3 total nodes, got %d", graph.Stats.TotalNodes)
	}
	if graph.Stats.TotalActivities != 2 {
		t.Errorf("expected 2 activities, got %d", graph.Stats.TotalActivities)
	}
	if graph.Stats.CompletedActivities != 1 {
		t.Errorf("expected 1 completed activity, got %d", graph.Stats.CompletedActivities)
	}
}

func TestUnifiedGraph_ToText(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
	}
	graph := builder.WithActivities(activities).Build()

	text := graph.ToText()
	if text == "" {
		t.Error("ToText returned empty string")
	}
	if !strings.Contains(text, "Unified Graph") {
		t.Error("ToText output should contain 'Unified Graph'")
	}
	if !strings.Contains(text, "act-001") {
		t.Error("ToText output should contain node ID")
	}
}

func TestUnifiedGraph_ToDot(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "active", Dependencies: []string{"act-001"}},
	}
	graph := builder.WithActivities(activities).Build()

	dot := graph.ToDot()
	if dot == "" {
		t.Error("ToDot returned empty string")
	}
	if !strings.Contains(dot, "digraph") {
		t.Error("ToDot output should contain 'digraph'")
	}
	if !strings.Contains(dot, "act-001") {
		t.Error("ToDot output should contain node ID")
	}
}

func TestUnifiedGraph_ToMermaid(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "active", Dependencies: []string{"act-001"}},
	}
	graph := builder.WithActivities(activities).Build()

	mermaid := graph.ToMermaid()
	if mermaid == "" {
		t.Error("ToMermaid returned empty string")
	}
	if !strings.Contains(mermaid, "graph TD") {
		t.Error("ToMermaid output should contain 'graph TD'")
	}
}

func TestUnifiedGraphBuilder_Build_DepthCalculation(t *testing.T) {
	// Objective → UseCase → Activity の階層で深さが正しく計算されるか検証
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active", ObjectiveID: "obj-001"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
	}

	graph := NewUnifiedGraphBuilder().
		WithObjectives(objectives).
		WithUseCases(usecases).
		WithActivities(activities).
		Build()

	// Objective が深さ 0（ルート）
	if graph.Nodes["obj-001"].Depth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].Depth)
	}
	// UseCase が深さ 1
	if graph.Nodes["uc-001"].Depth != 1 {
		t.Errorf("expected uc-001 depth 1, got %d", graph.Nodes["uc-001"].Depth)
	}
	// Activity が深さ 2
	if graph.Nodes["act-001"].Depth != 2 {
		t.Errorf("expected act-001 depth 2, got %d", graph.Nodes["act-001"].Depth)
	}
}

func TestUnifiedGraphBuilder_Build_DepthWithDeliverable(t *testing.T) {
	// Objective → Deliverable の階層で深さが正しく計算されるか検証
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	deliverables := []DeliverableInfo{
		{ID: "del-001", Title: "Deliverable 1", Status: "draft", ObjectiveID: "obj-001"},
	}

	graph := NewUnifiedGraphBuilder().
		WithObjectives(objectives).
		WithDeliverables(deliverables).
		Build()

	// Objective が深さ 0（ルート）
	if graph.Nodes["obj-001"].Depth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].Depth)
	}
	// Deliverable が深さ 1
	if graph.Nodes["del-001"].Depth != 1 {
		t.Errorf("expected del-001 depth 1, got %d", graph.Nodes["del-001"].Depth)
	}
}

func TestUnifiedGraphBuilder_Build_DepthWithMultipleActivities(t *testing.T) {
	// Activity 間の依存関係と UseCase 関連が組み合わさった場合の深さ計算
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active", ObjectiveID: "obj-001"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
		{ID: "act-002", Title: "Activity 2", Status: "active", UseCaseID: "uc-001", Dependencies: []string{"act-001"}},
		{ID: "act-003", Title: "Activity 3", Status: "active", UseCaseID: "uc-001", Dependencies: []string{"act-002"}},
	}

	graph := NewUnifiedGraphBuilder().
		WithObjectives(objectives).
		WithUseCases(usecases).
		WithActivities(activities).
		Build()

	// Objective が深さ 0
	if graph.Nodes["obj-001"].Depth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].Depth)
	}
	// UseCase が深さ 1
	if graph.Nodes["uc-001"].Depth != 1 {
		t.Errorf("expected uc-001 depth 1, got %d", graph.Nodes["uc-001"].Depth)
	}
	// Activity 1 が深さ 2（UseCase の子）
	if graph.Nodes["act-001"].Depth != 2 {
		t.Errorf("expected act-001 depth 2, got %d", graph.Nodes["act-001"].Depth)
	}
	// Activity 2 が深さ 3（Activity 1 に依存）
	if graph.Nodes["act-002"].Depth != 3 {
		t.Errorf("expected act-002 depth 3, got %d", graph.Nodes["act-002"].Depth)
	}
	// Activity 3 が深さ 4（Activity 2 に依存）
	if graph.Nodes["act-003"].Depth != 4 {
		t.Errorf("expected act-003 depth 4, got %d", graph.Nodes["act-003"].Depth)
	}
}

func TestGraphFilter_Chaining(t *testing.T) {
	filter := NewGraphFilter().
		WithFocus("act-001", 2).
		WithIncludeTypes(EntityTypeActivity, EntityTypeUseCase).
		WithExcludeTypes(EntityTypeDeliverable).
		WithHideCompleted(true).
		WithHideDraft(true).
		WithHideUnrelated(true)

	if filter.FocusID != "act-001" {
		t.Errorf("expected focus ID 'act-001', got %s", filter.FocusID)
	}
	if filter.FocusDepth != 2 {
		t.Errorf("expected focus depth 2, got %d", filter.FocusDepth)
	}
	if len(filter.IncludeTypes) != 2 {
		t.Errorf("expected 2 include types, got %d", len(filter.IncludeTypes))
	}
	if len(filter.ExcludeTypes) != 1 {
		t.Errorf("expected 1 exclude type, got %d", len(filter.ExcludeTypes))
	}
	if !filter.HideCompleted {
		t.Error("expected HideCompleted to be true")
	}
	if !filter.HideDraft {
		t.Error("expected HideDraft to be true")
	}
	if !filter.HideUnrelated {
		t.Error("expected HideUnrelated to be true")
	}
}
