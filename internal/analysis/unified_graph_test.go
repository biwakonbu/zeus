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
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "deprecated"},
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

func TestUnifiedGraphBuilder_WithObjectives_Basic(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	result := builder.WithObjectives(objectives)
	if result == nil {
		t.Fatal("WithObjectives returned nil")
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
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
		{ID: "act-002", Title: "Activity 2", Status: "deprecated", UseCaseID: "uc-001"},
	}
	graph := builder.WithUseCases(usecases).WithActivities(activities).Build()

	if graph == nil {
		t.Fatal("Build returned nil")
	}
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
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

	// implements エッジの検証（Activity -> UseCase）
	if len(graph.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(graph.Edges))
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

func TestUnifiedGraphBuilder_Build_WithContributesRelation(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active", ObjectiveID: "obj-001"},
	}
	graph := builder.WithObjectives(objectives).WithUseCases(usecases).Build()

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// UseCase -> Objective の contributes エッジがあるか確認
	foundEdge := false
	for _, edge := range graph.Edges {
		if edge.From == "uc-001" && edge.To == "obj-001" && edge.Relation == RelationContributes {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Error("expected contributes edge from uc-001 to obj-001")
	}
}

func TestUnifiedGraphBuilder_Build_NoCycleForFlatObjectives(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	// Objective は親子関係を持たないため循環依存は発生しない
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
		{ID: "obj-002", Title: "Objective 2", Status: "active"},
		{ID: "obj-003", Title: "Objective 3", Status: "active"},
	}
	graph := builder.WithObjectives(objectives).Build()

	if len(graph.Cycles) != 0 {
		t.Errorf("expected no cycles for flat objectives, got %d", len(graph.Cycles))
	}
}

func TestUnifiedGraphBuilder_Build_IsolatedNodes(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Connected", Status: "active", UseCaseID: "uc-001"},
		{ID: "act-002", Title: "Isolated", Status: "active"}, // 孤立ノード
	}
	graph := builder.WithUseCases(usecases).WithActivities(activities).Build()

	if len(graph.Isolated) != 1 {
		t.Errorf("expected 1 isolated node, got %d", len(graph.Isolated))
	}
	if len(graph.Isolated) > 0 && graph.Isolated[0] != "act-002" {
		t.Errorf("expected isolated node act-002, got %s", graph.Isolated[0])
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
		{ID: "act-002", Title: "Deprecated", Status: "deprecated"},
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
	// obj-001 <- uc-001 <- act-001, obj-001 <- uc-002 (uc-002 は深さ 1 で到達可能)
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active", ObjectiveID: "obj-001"},
		{ID: "uc-002", Title: "UseCase 2", Status: "active", ObjectiveID: "obj-001"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
	}
	// uc-001 から深さ 1 でフォーカス
	filter := NewGraphFilter().WithFocus("uc-001", 1).WithHideUnrelated(true)
	graph := builder.
		WithObjectives(objectives).
		WithUseCases(usecases).
		WithActivities(activities).
		WithFilter(filter).
		Build()

	// uc-001, obj-001, act-001 のみが含まれるはず（深さ1以内）
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes (depth 1 from uc-001), got %d", len(graph.Nodes))
		for id := range graph.Nodes {
			t.Logf("  - %s", id)
		}
	}
}

func TestUnifiedGraph_Stats(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active"},
		{ID: "act-002", Title: "Activity 2", Status: "deprecated"},
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
		t.Errorf("expected 1 deprecated activity, got %d", graph.Stats.CompletedActivities)
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
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
	}
	graph := builder.WithUseCases(usecases).WithActivities(activities).Build()

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
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active"},
	}
	activities := []ActivityInfo{
		{ID: "act-001", Title: "Activity 1", Status: "active", UseCaseID: "uc-001"},
	}
	graph := builder.WithUseCases(usecases).WithActivities(activities).Build()

	mermaid := graph.ToMermaid()
	if mermaid == "" {
		t.Error("ToMermaid returned empty string")
	}
	if !strings.Contains(mermaid, "graph TD") {
		t.Error("ToMermaid output should contain 'graph TD'")
	}
	if !strings.Contains(mermaid, "implements") {
		t.Error("ToMermaid should include relation label 'implements'")
	}
}

func TestUnifiedGraphBuilder_ValidationErrors(t *testing.T) {
	// Activity が UseCase に implements する有効なエッジの検証
	builder := NewUnifiedGraphBuilder()
	builder.
		WithUseCases([]UseCaseInfo{{ID: "uc-001", Title: "UseCase 1", Status: "active"}}).
		WithActivities([]ActivityInfo{
			{
				ID:        "act-001",
				Title:     "Activity 1",
				Status:    "active",
				UseCaseID: "uc-001",
			},
		}).
		Build()

	if len(builder.ValidationErrors()) != 0 {
		t.Fatalf("expected no validation errors, got %v", builder.ValidationErrors())
	}
}

func TestUnifiedGraphBuilder_WithFilter_IncludeLayersAndRelations(t *testing.T) {
	builder := NewUnifiedGraphBuilder()
	graph := builder.
		WithObjectives([]ObjectiveInfo{{ID: "obj-001", Title: "Objective", Status: "active"}}).
		WithUseCases([]UseCaseInfo{{ID: "uc-001", Title: "UseCase", Status: "active", ObjectiveID: "obj-001"}}).
		WithActivities([]ActivityInfo{{
			ID:        "act-001",
			Title:     "Activity",
			Status:    "active",
			UseCaseID: "uc-001",
		}}).
		WithFilter(NewGraphFilter().
			WithIncludeLayers(EdgeLayerStructural).
			WithIncludeRelations(RelationImplements)).
		Build()

	// implements エッジ (act-001 -> uc-001) のみ
	if len(graph.Edges) != 1 {
		t.Fatalf("expected only one structural/implements edge, got %d", len(graph.Edges))
	}
	edge := graph.Edges[0]
	if edge.Layer != EdgeLayerStructural || edge.Relation != RelationImplements {
		t.Fatalf("unexpected edge: %+v", edge)
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
	if graph.Nodes["obj-001"].StructuralDepth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].StructuralDepth)
	}
	// UseCase が深さ 1
	if graph.Nodes["uc-001"].StructuralDepth != 1 {
		t.Errorf("expected uc-001 depth 1, got %d", graph.Nodes["uc-001"].StructuralDepth)
	}
	// Activity が深さ 2
	if graph.Nodes["act-001"].StructuralDepth != 2 {
		t.Errorf("expected act-001 depth 2, got %d", graph.Nodes["act-001"].StructuralDepth)
	}
}

func TestUnifiedGraphBuilder_Build_DepthWithUseCase(t *testing.T) {
	// Objective -> UseCase の階層で深さが正しく計算されるか検証
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "Objective 1", Status: "active"},
	}
	usecases := []UseCaseInfo{
		{ID: "uc-001", Title: "UseCase 1", Status: "active", ObjectiveID: "obj-001"},
	}

	graph := NewUnifiedGraphBuilder().
		WithObjectives(objectives).
		WithUseCases(usecases).
		Build()

	// Objective が深さ 0（ルート）
	if graph.Nodes["obj-001"].StructuralDepth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].StructuralDepth)
	}
	// UseCase が深さ 1
	if graph.Nodes["uc-001"].StructuralDepth != 1 {
		t.Errorf("expected uc-001 depth 1, got %d", graph.Nodes["uc-001"].StructuralDepth)
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
		{ID: "act-002", Title: "Activity 2", Status: "active", UseCaseID: "uc-001"},
		{ID: "act-003", Title: "Activity 3", Status: "active", UseCaseID: "uc-001"},
	}

	graph := NewUnifiedGraphBuilder().
		WithObjectives(objectives).
		WithUseCases(usecases).
		WithActivities(activities).
		Build()

	// Objective が深さ 0
	if graph.Nodes["obj-001"].StructuralDepth != 0 {
		t.Errorf("expected obj-001 depth 0, got %d", graph.Nodes["obj-001"].StructuralDepth)
	}
	// UseCase が深さ 1
	if graph.Nodes["uc-001"].StructuralDepth != 1 {
		t.Errorf("expected uc-001 depth 1, got %d", graph.Nodes["uc-001"].StructuralDepth)
	}
	// Activity 1 が深さ 2（UseCase の子）
	if graph.Nodes["act-001"].StructuralDepth != 2 {
		t.Errorf("expected act-001 depth 2, got %d", graph.Nodes["act-001"].StructuralDepth)
	}
	// Activity 2 も structural depth は 2（UseCase の子）
	if graph.Nodes["act-002"].StructuralDepth != 2 {
		t.Errorf("expected act-002 structural depth 2, got %d", graph.Nodes["act-002"].StructuralDepth)
	}
	// Activity 3 も同様に structural depth は 2
	if graph.Nodes["act-003"].StructuralDepth != 2 {
		t.Errorf("expected act-003 structural depth 2, got %d", graph.Nodes["act-003"].StructuralDepth)
	}
}

func TestGraphFilter_Chaining(t *testing.T) {
	filter := NewGraphFilter().
		WithFocus("act-001", 2).
		WithIncludeTypes(EntityTypeActivity, EntityTypeUseCase).
		WithExcludeTypes(EntityTypeObjective).
		WithIncludeLayers(EdgeLayerStructural).
		WithIncludeRelations(RelationImplements, RelationContributes).
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
	if len(filter.IncludeLayers) != 1 {
		t.Errorf("expected 1 include layer, got %d", len(filter.IncludeLayers))
	}
	if len(filter.IncludeRelations) != 2 {
		t.Errorf("expected 2 include relations, got %d", len(filter.IncludeRelations))
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
