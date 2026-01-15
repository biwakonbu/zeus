package analysis

import (
	"context"
	"strings"
	"testing"
)

func TestNewGraphBuilder(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusInProgress, Dependencies: []string{"task-1"}},
	}

	builder := NewGraphBuilder(tasks)

	if builder == nil {
		t.Fatal("NewGraphBuilder returned nil")
	}

	if len(builder.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(builder.tasks))
	}
}

func TestGraphBuilder_Build(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusInProgress, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, Dependencies: []string{"task-2"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, err := builder.Build(ctx)

	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	// ノード数を確認
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	// エッジ数を確認（2つの依存関係）
	if len(graph.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(graph.Edges))
	}

	// 統計を確認
	if graph.Stats.TotalNodes != 3 {
		t.Errorf("expected TotalNodes=3, got %d", graph.Stats.TotalNodes)
	}

	if graph.Stats.WithDependencies != 3 {
		t.Errorf("expected WithDependencies=3, got %d", graph.Stats.WithDependencies)
	}

	// 循環がないことを確認
	if len(graph.Cycles) != 0 {
		t.Errorf("expected no cycles, got %d", len(graph.Cycles))
	}
}

func TestGraphBuilder_Build_WithIsolatedNodes(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, Dependencies: []string{}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, Dependencies: []string{}},
	}

	builder := NewGraphBuilder(tasks)
	graph, err := builder.Build(ctx)

	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	// 全てのノードが孤立している
	if len(graph.Isolated) != 3 {
		t.Errorf("expected 3 isolated nodes, got %d", len(graph.Isolated))
	}

	if graph.Stats.IsolatedCount != 3 {
		t.Errorf("expected IsolatedCount=3, got %d", graph.Stats.IsolatedCount)
	}
}

func TestGraphBuilder_Build_WithCycles(t *testing.T) {
	ctx := context.Background()

	// 循環依存を含むタスク
	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, Dependencies: []string{"task-3"}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, Dependencies: []string{"task-2"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, err := builder.Build(ctx)

	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	// 循環が検出されることを確認
	if len(graph.Cycles) == 0 {
		t.Error("expected cycles to be detected")
	}

	if graph.Stats.CycleCount == 0 {
		t.Error("expected CycleCount > 0")
	}
}

func TestGraphBuilder_Build_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // キャンセル済み

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, Dependencies: []string{}},
	}

	builder := NewGraphBuilder(tasks)
	_, err := builder.Build(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestDependencyGraph_ToText(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Root Task", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Child Task", Status: TaskStatusPending, Dependencies: []string{"task-1"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	text := graph.ToText()

	// 基本的なヘッダーを含むことを確認
	if !strings.Contains(text, "Zeus Dependency Graph") {
		t.Error("expected header in text output")
	}

	// 統計情報を含むことを確認
	if !strings.Contains(text, "Total tasks:") {
		t.Error("expected stats in text output")
	}
}

func TestDependencyGraph_ToDot(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusInProgress, Dependencies: []string{"task-1"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	dot := graph.ToDot()

	// DOT形式の基本構造を確認
	if !strings.Contains(dot, "digraph ZeusDependencies") {
		t.Error("expected digraph declaration")
	}

	if !strings.Contains(dot, "rankdir=TB") {
		t.Error("expected rankdir declaration")
	}

	// ノード定義を確認
	if !strings.Contains(dot, "task-1") {
		t.Error("expected task-1 node")
	}

	// ステータスに応じた色を確認
	if !strings.Contains(dot, "lightgreen") {
		t.Error("expected lightgreen for completed task")
	}

	if !strings.Contains(dot, "lightyellow") {
		t.Error("expected lightyellow for in_progress task")
	}

	// エッジ定義を確認
	if !strings.Contains(dot, "->") {
		t.Error("expected edge definition")
	}
}

func TestDependencyGraph_ToMermaid(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusBlocked, Dependencies: []string{"task-1"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	mermaid := graph.ToMermaid()

	// Mermaid形式の基本構造を確認
	if !strings.Contains(mermaid, "```mermaid") {
		t.Error("expected mermaid code block start")
	}

	if !strings.Contains(mermaid, "graph TD") {
		t.Error("expected graph TD declaration")
	}

	// IDがサニタイズされていることを確認（ハイフンがアンダースコアに）
	if !strings.Contains(mermaid, "task_1") {
		t.Error("expected sanitized task_1 ID")
	}

	// スタイル定義を確認
	if !strings.Contains(mermaid, "#90EE90") {
		t.Error("expected green style for completed task")
	}

	if !strings.Contains(mermaid, "#F08080") {
		t.Error("expected red style for blocked task")
	}

	// エッジ定義を確認
	if !strings.Contains(mermaid, "-->") {
		t.Error("expected arrow definition")
	}

	if !strings.Contains(mermaid, "```\n") {
		t.Error("expected mermaid code block end")
	}
}

func TestDependencyGraph_DepthCalculation(t *testing.T) {
	ctx := context.Background()

	// 依存関係チェーン: task-4 は task-3 に依存、task-3 は task-2 に依存、task-2 は task-1 に依存
	// グラフの観点では: task-4 -> task-3 -> task-2 -> task-1
	// ルートノード（親がないノード）は task-4（他から参照されていない）
	// 深さは task-4(0) -> task-3(1) -> task-2(2) -> task-1(3)
	tasks := []TaskInfo{
		{ID: "task-1", Title: "Leaf", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Level 2", Status: TaskStatusCompleted, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Level 1", Status: TaskStatusPending, Dependencies: []string{"task-2"}},
		{ID: "task-4", Title: "Root", Status: TaskStatusPending, Dependencies: []string{"task-3"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	// task-4はルートノード（他から参照されていない）なので深さ0
	if graph.Nodes["task-4"].Depth != 0 {
		t.Errorf("expected task-4 depth=0 (root), got %d", graph.Nodes["task-4"].Depth)
	}

	// task-1は最も深い（task-4から3ホップ）
	if graph.Nodes["task-1"].Depth != 3 {
		t.Errorf("expected task-1 depth=3 (leaf), got %d", graph.Nodes["task-1"].Depth)
	}

	if graph.Stats.MaxDepth != 3 {
		t.Errorf("expected MaxDepth=3, got %d", graph.Stats.MaxDepth)
	}
}

func TestDependencyGraph_ParentReferences(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{
		{ID: "task-1", Title: "Parent", Status: TaskStatusCompleted, Dependencies: []string{}},
		{ID: "task-2", Title: "Child 1", Status: TaskStatusPending, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Child 2", Status: TaskStatusPending, Dependencies: []string{"task-1"}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	// task-1は2つのタスク（task-2, task-3）から参照されているので、2つのParentを持つ
	parents := graph.Nodes["task-1"].Parents
	if len(parents) != 2 {
		t.Errorf("expected task-1 to have 2 parents (referencing tasks), got %d", len(parents))
	}

	// task-2はtask-1を依存先として持つ（Children）
	if len(graph.Nodes["task-2"].Children) != 1 {
		t.Errorf("expected task-2 to have 1 child (dependency), got %d", len(graph.Nodes["task-2"].Children))
	}
}

func TestDependencyGraph_EmptyGraph(t *testing.T) {
	ctx := context.Background()

	tasks := []TaskInfo{}

	builder := NewGraphBuilder(tasks)
	graph, err := builder.Build(ctx)

	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}

	if graph.Stats.TotalNodes != 0 {
		t.Errorf("expected TotalNodes=0, got %d", graph.Stats.TotalNodes)
	}
}

func TestDependencyGraph_ToDot_WithQuotes(t *testing.T) {
	ctx := context.Background()

	// タイトルにクォートを含むタスク
	tasks := []TaskInfo{
		{ID: "task-1", Title: `Task with "quotes"`, Status: TaskStatusPending, Dependencies: []string{}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	dot := graph.ToDot()

	// クォートがエスケープされていることを確認
	if !strings.Contains(dot, `\"`) {
		t.Error("expected escaped quotes in DOT output")
	}
}

func TestDependencyGraph_ToMermaid_WithSpecialChars(t *testing.T) {
	ctx := context.Background()

	// ハイフンを含むIDと特殊文字を含むタイトル
	tasks := []TaskInfo{
		{ID: "task-with-hyphens", Title: `Task with "special" chars`, Status: TaskStatusPending, Dependencies: []string{}},
	}

	builder := NewGraphBuilder(tasks)
	graph, _ := builder.Build(ctx)

	mermaid := graph.ToMermaid()

	// ハイフンがアンダースコアに変換されていることを確認
	if !strings.Contains(mermaid, "task_with_hyphens") {
		t.Error("expected hyphens to be replaced with underscores")
	}

	// クォートがシングルクォートに変換されていることを確認
	if !strings.Contains(mermaid, "'") {
		t.Error("expected quotes to be replaced with single quotes")
	}
}
