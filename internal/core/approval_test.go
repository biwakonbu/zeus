package core

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateApprovalID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	am := NewApprovalManager(tmpDir)

	// ID 生成テスト
	id1 := am.generateApprovalID()
	id2 := am.generateApprovalID()

	// プレフィックスが正しいか
	if !strings.HasPrefix(id1, "approval-") {
		t.Errorf("expected ID to start with 'approval-', got %q", id1)
	}

	// UUID ベースのため、2つの ID が異なるはず
	if id1 == id2 {
		t.Errorf("generated IDs should be unique, but got same: %q", id1)
	}

	// ID の長さが適切か (approval- + 8文字)
	if len(id1) != 17 {
		t.Errorf("expected ID length to be 17, got %d", len(id1))
	}
}

func TestApprovalNotPendingError(t *testing.T) {
	err := &ApprovalNotPendingError{
		ID:            "approval-12345678",
		CurrentStatus: ApprovalStatusApproved,
	}

	// Error() メッセージ検証
	msg := err.Error()
	if !strings.Contains(msg, "approval-12345678") {
		t.Errorf("error message should contain ID, got %q", msg)
	}
	if !strings.Contains(msg, "approved") {
		t.Errorf("error message should contain status, got %q", msg)
	}

	// Is() 検証
	if !err.Is(ErrApprovalNotPending) {
		t.Error("ApprovalNotPendingError should match ErrApprovalNotPending")
	}
}
