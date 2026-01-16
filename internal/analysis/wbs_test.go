package analysis

import (
	"context"
	"testing"
)

func TestWBSBuilder_DetectParentCycles_NoCycle(t *testing.T) {
	// 循環なし: A -> B -> C（ルートから子へ）
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: ""},
		{ID: "task-b", Title: "Task B", ParentID: "task-a"},
		{ID: "task-c", Title: "Task C", ParentID: "task-b"},
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %d: %v", len(cycles), cycles)
	}

	// Build も正常に完了することを確認
	tree, err := builder.Build(context.Background())
	if err != nil {
		t.Errorf("Build should succeed without cycles: %v", err)
	}
	if tree == nil {
		t.Error("Build should return non-nil tree")
	}
}

func TestWBSBuilder_DetectParentCycles_TwoNodeCycle(t *testing.T) {
	// 2ノード循環: A -> B -> A
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: "task-b"},
		{ID: "task-b", Title: "Task B", ParentID: "task-a"},
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) == 0 {
		t.Error("expected cycle to be detected")
	}

	// Build はエラーを返すことを確認
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("Build should fail with cycle")
	}
	if err != nil && !contains(err.Error(), "parent cycle detected") {
		t.Errorf("expected 'parent cycle detected' error, got: %v", err)
	}
}

func TestWBSBuilder_DetectParentCycles_ThreeNodeCycle(t *testing.T) {
	// 3ノード循環: A -> B -> C -> A
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: "task-c"},
		{ID: "task-b", Title: "Task B", ParentID: "task-a"},
		{ID: "task-c", Title: "Task C", ParentID: "task-b"},
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) == 0 {
		t.Error("expected cycle to be detected")
	}

	// Build はエラーを返すことを確認
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("Build should fail with cycle")
	}
}

func TestWBSBuilder_DetectParentCycles_SelfReference(t *testing.T) {
	// 自己参照: A -> A
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: "task-a"},
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) == 0 {
		t.Error("expected self-reference cycle to be detected")
	}

	// Build はエラーを返すことを確認
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("Build should fail with self-reference")
	}
}

func TestWBSBuilder_DetectParentCycles_NonExistentParent(t *testing.T) {
	// 存在しない親ID（循環ではない、孤立ノードとして扱う）
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: "non-existent"},
		{ID: "task-b", Title: "Task B", ParentID: "task-a"},
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) != 0 {
		t.Errorf("expected no cycles (non-existent parent is not a cycle), got %d: %v", len(cycles), cycles)
	}

	// Build は成功することを確認
	tree, err := builder.Build(context.Background())
	if err != nil {
		t.Errorf("Build should succeed with non-existent parent: %v", err)
	}
	if tree == nil {
		t.Error("Build should return non-nil tree")
	}
}

func TestWBSBuilder_DetectParentCycles_MixedCycleAndNonCycle(t *testing.T) {
	// 一部に循環あり、他は正常
	tasks := []TaskInfo{
		{ID: "task-a", Title: "Task A", ParentID: ""},          // ルート
		{ID: "task-b", Title: "Task B", ParentID: "task-a"},    // 正常
		{ID: "task-c", Title: "Task C", ParentID: "task-d"},    // 循環
		{ID: "task-d", Title: "Task D", ParentID: "task-c"},    // 循環
	}

	builder := NewWBSBuilder(tasks)
	cycles := builder.DetectParentCycles()

	if len(cycles) == 0 {
		t.Error("expected cycle to be detected in subset")
	}

	// Build はエラーを返すことを確認
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("Build should fail with partial cycle")
	}
}

// contains はエラーメッセージに特定の文字列が含まれるかチェック
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
