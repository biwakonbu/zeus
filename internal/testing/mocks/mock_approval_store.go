package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/biwakonbu/zeus/internal/core"
)

// MockApprovalStore は ApprovalStore インターフェースのモック実装
type MockApprovalStore struct {
	// テストデータ
	pendingApprovals []core.PendingApproval

	// エラー注入用
	GetPendingError error
	CreateError     error
	ApproveError    error
	RejectError     error
}

// NewMockApprovalStore は新しいモックを作成
func NewMockApprovalStore() *MockApprovalStore {
	return &MockApprovalStore{
		pendingApprovals: []core.PendingApproval{},
	}
}

// GetPending は承認待ち一覧を返す
func (m *MockApprovalStore) GetPending(ctx context.Context) ([]core.PendingApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.GetPendingError != nil {
		return nil, m.GetPendingError
	}
	pending := []core.PendingApproval{}
	for _, approval := range m.pendingApprovals {
		if approval.Status == core.ApprovalStatusPending {
			pending = append(pending, approval)
		}
	}
	return pending, nil
}

// GetAll は全承認アイテムを返す
func (m *MockApprovalStore) GetAll(ctx context.Context) ([]core.PendingApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m.pendingApprovals, nil
}

// Get は特定の承認アイテムを取得
func (m *MockApprovalStore) Get(ctx context.Context, id string) (*core.PendingApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, approval := range m.pendingApprovals {
		if approval.ID == id {
			return &approval, nil
		}
	}
	return nil, fmt.Errorf("approval not found: %s", id)
}

// DetermineApprovalLevel はアクションに応じた承認レベルを決定
func (m *MockApprovalStore) DetermineApprovalLevel(actionType string, settings *core.Settings) core.ApprovalLevel {
	// テスト用のシンプルな実装
	switch actionType {
	case "task_delete", "config_change":
		return core.ApprovalApprove
	case "task_update":
		return core.ApprovalNotify
	default:
		return core.ApprovalAuto
	}
}

// Create は新しい承認を作成
func (m *MockApprovalStore) Create(ctx context.Context, approvalType, description string, level core.ApprovalLevel, entityID string, payload any) (*core.PendingApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.CreateError != nil {
		return nil, m.CreateError
	}

	approval := &core.PendingApproval{
		ID:          fmt.Sprintf("approval-%d", len(m.pendingApprovals)+1),
		Type:        approvalType,
		EntityID:    entityID,
		Description: description,
		Level:       level,
		Status:      core.ApprovalStatusPending,
		Payload:     payload,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	m.pendingApprovals = append(m.pendingApprovals, *approval)
	return approval, nil
}

// Approve は承認を実行
func (m *MockApprovalStore) Approve(ctx context.Context, id string) (*core.ApprovalResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.ApproveError != nil {
		return nil, m.ApproveError
	}

	for i, approval := range m.pendingApprovals {
		if approval.ID == id {
			m.pendingApprovals[i].Status = core.ApprovalStatusApproved
			m.pendingApprovals[i].UpdatedAt = time.Now().Format(time.RFC3339)
			m.pendingApprovals[i].ApprovedBy = "test-user"

			result := &core.ApprovalResult{
				Success: true,
				ID:      id,
				Status:  core.ApprovalStatusApproved,
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("approval not found: %s", id)
}

// Reject は承認を却下
func (m *MockApprovalStore) Reject(ctx context.Context, id, reason string) (*core.ApprovalResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.RejectError != nil {
		return nil, m.RejectError
	}

	for i, approval := range m.pendingApprovals {
		if approval.ID == id {
			m.pendingApprovals[i].Status = core.ApprovalStatusRejected
			m.pendingApprovals[i].UpdatedAt = time.Now().Format(time.RFC3339)
			m.pendingApprovals[i].RejectedBy = "test-user"
			m.pendingApprovals[i].Reason = reason

			result := &core.ApprovalResult{
				Success: false,
				ID:      id,
				Status:  core.ApprovalStatusRejected,
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("approval not found: %s", id)
}
