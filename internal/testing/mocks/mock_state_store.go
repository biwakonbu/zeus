package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/biwakonbu/zeus/internal/core"
)

// MockStateStore は StateStore インターフェースのモック実装
type MockStateStore struct {
	// テストデータ
	currentState *core.ProjectState
	snapshots    []core.Snapshot

	// エラー注入用
	GetCurrentStateError  error
	SaveCurrentStateError error
	CreateSnapshotError   error
	GetHistoryError       error
	RestoreSnapshotError  error
}

// NewMockStateStore は新しいモックを作成
func NewMockStateStore() *MockStateStore {
	return &MockStateStore{
		currentState: &core.ProjectState{
			Summary: core.SummaryStats{
				TotalActivities: 0,
				Completed:       0,
				InProgress:      0,
				Pending:         0,
			},
			Health: "good",
		},
		snapshots: []core.Snapshot{},
	}
}

// GetCurrentState は現在の状態を返す
func (m *MockStateStore) GetCurrentState(ctx context.Context) (*core.ProjectState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.GetCurrentStateError != nil {
		return nil, m.GetCurrentStateError
	}
	return m.currentState, nil
}

// SaveCurrentState は状態を保存
func (m *MockStateStore) SaveCurrentState(ctx context.Context, state *core.ProjectState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.SaveCurrentStateError != nil {
		return m.SaveCurrentStateError
	}
	m.currentState = state
	return nil
}

// CreateSnapshot はスナップショットを作成
func (m *MockStateStore) CreateSnapshot(ctx context.Context, label string) (*core.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.CreateSnapshotError != nil {
		return nil, m.CreateSnapshotError
	}

	snapshot := &core.Snapshot{
		Timestamp: time.Now().Format(time.RFC3339),
		Label:     label,
		State:     *m.currentState,
	}
	m.snapshots = append(m.snapshots, *snapshot)
	return snapshot, nil
}

// GetHistory はスナップショット履歴を取得
func (m *MockStateStore) GetHistory(ctx context.Context, limit int) ([]core.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.GetHistoryError != nil {
		return nil, m.GetHistoryError
	}

	if limit <= 0 || limit > len(m.snapshots) {
		return m.snapshots, nil
	}
	return m.snapshots[len(m.snapshots)-limit:], nil
}

// GetSnapshot は特定のスナップショットを取得
func (m *MockStateStore) GetSnapshot(ctx context.Context, timestamp string) (*core.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for _, snapshot := range m.snapshots {
		if snapshot.Timestamp == timestamp {
			return &snapshot, nil
		}
	}
	return nil, fmt.Errorf("snapshot not found: %s", timestamp)
}

// RestoreSnapshot はスナップショットを復元
func (m *MockStateStore) RestoreSnapshot(ctx context.Context, timestamp string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.RestoreSnapshotError != nil {
		return m.RestoreSnapshotError
	}

	for _, snapshot := range m.snapshots {
		if snapshot.Timestamp == timestamp {
			m.currentState = &snapshot.State
			return nil
		}
	}
	return fmt.Errorf("snapshot not found: %s", timestamp)
}

// CalculateState はリスト項目から状態を計算
func (m *MockStateStore) CalculateState(items []core.ListItem) *core.ProjectState {
	stats := core.SummaryStats{
		TotalActivities: len(items),
	}

	for _, item := range items {
		switch item.Status {
		case core.ItemStatusCompleted:
			stats.Completed++
		case core.ItemStatusInProgress:
			stats.InProgress++
		case core.ItemStatusPending:
			stats.Pending++
		}
	}

	health := core.HealthGood
	if stats.TotalActivities > 0 {
		completionRate := float64(stats.Completed) / float64(stats.TotalActivities)
		if completionRate < 0.3 {
			health = core.HealthPoor
		} else if completionRate < 0.7 {
			health = core.HealthFair
		}
	}

	return &core.ProjectState{
		Timestamp: time.Now().Format(time.RFC3339),
		Summary:   stats,
		Health:    health,
		Risks:     []string{},
	}
}
