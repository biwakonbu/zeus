package core

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateTaskID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// ID 生成テスト
	id1 := z.generateTaskID()
	id2 := z.generateTaskID()

	// プレフィックスが正しいか
	if !strings.HasPrefix(id1, "task-") {
		t.Errorf("expected ID to start with 'task-', got %q", id1)
	}

	// UUID ベースのため、2つの ID が異なるはず
	if id1 == id2 {
		t.Errorf("generated IDs should be unique, but got same: %q", id1)
	}

	// ID の長さが適切か (task- + 8文字)
	if len(id1) != 13 {
		t.Errorf("expected ID length to be 13, got %d", len(id1))
	}
}

func TestGenerateTaskIDUniqueness(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 1000個の ID を生成して重複がないか確認
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := z.generateTaskID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		ids[id] = true
	}
}
