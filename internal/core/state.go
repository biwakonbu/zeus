package core

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// SnapshotStore はスナップショットストア
type SnapshotStore struct {
	Snapshots []Snapshot `yaml:"snapshots"`
}

// StateManager は状態を管理
type StateManager struct {
	zeusPath  string
	fileStore FileStore
}

// NewStateManager は新しい StateManager を作成
func NewStateManager(zeusPath string, fs FileStore) *StateManager {
	return &StateManager{
		zeusPath:  zeusPath,
		fileStore: fs,
	}
}

// GetCurrentState は現在の状態を取得
func (sm *StateManager) GetCurrentState(ctx context.Context) (*ProjectState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var state ProjectState
	if err := sm.fileStore.ReadYaml(ctx, "state/current.yaml", &state); err != nil {
		return sm.getEmptyState(), nil
	}
	return &state, nil
}

// SaveCurrentState は現在の状態を保存
func (sm *StateManager) SaveCurrentState(ctx context.Context, state *ProjectState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return sm.fileStore.WriteYaml(ctx, "state/current.yaml", state)
}

// CreateSnapshot はスナップショットを作成
func (sm *StateManager) CreateSnapshot(ctx context.Context, label string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 現在の状態を取得
	currentState, err := sm.GetCurrentState(ctx)
	if err != nil {
		return nil, err
	}

	// スナップショット作成
	snapshot := Snapshot{
		Timestamp: Now(),
		Label:     label,
		State:     *currentState,
	}

	// スナップショットを保存
	if err := sm.saveSnapshot(ctx, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// GetHistory はスナップショット履歴を取得
func (sm *StateManager) GetHistory(ctx context.Context, limit int) ([]Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	snapshots, err := sm.getAllSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	// タイムスタンプで降順ソート（新しい順）
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp > snapshots[j].Timestamp
	})

	// 件数制限
	if limit > 0 && len(snapshots) > limit {
		snapshots = snapshots[:limit]
	}

	return snapshots, nil
}

// GetSnapshot は特定のスナップショットを取得
func (sm *StateManager) GetSnapshot(ctx context.Context, timestamp string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	snapshots, err := sm.getAllSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range snapshots {
		if s.Timestamp == timestamp {
			return &s, nil
		}
	}

	return nil, ErrEntityNotFound
}

// RestoreSnapshot はスナップショットから復元
func (sm *StateManager) RestoreSnapshot(ctx context.Context, timestamp string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	snapshot, err := sm.GetSnapshot(ctx, timestamp)
	if err != nil {
		return err
	}

	// 現在の状態を更新
	return sm.SaveCurrentState(ctx, &snapshot.State)
}

// CalculateState はタスクから状態を計算
func (sm *StateManager) CalculateState(tasks []ListItem) *ProjectState {
	stats := TaskStats{
		TotalTasks: len(tasks),
	}

	for _, task := range tasks {
		switch task.Status {
		case ItemStatusCompleted:
			stats.Completed++
		case ItemStatusInProgress:
			stats.InProgress++
		case ItemStatusPending:
			stats.Pending++
		case ItemStatusBlocked:
			// ブロック状態も Pending に含める
			stats.Pending++
		}
	}

	return &ProjectState{
		Timestamp: Now(),
		Summary:   stats,
		Health:    sm.calculateHealth(&stats),
		Risks:     sm.detectRisks(tasks, &stats),
	}
}

// Private methods

func (sm *StateManager) getEmptyState() *ProjectState {
	return &ProjectState{
		Timestamp: Now(),
		Summary:   TaskStats{},
		Health:    HealthUnknown,
		Risks:     []string{},
	}
}

func (sm *StateManager) calculateHealth(stats *TaskStats) HealthStatus {
	if stats.TotalTasks == 0 {
		return HealthUnknown
	}

	progress := float64(stats.Completed) / float64(stats.TotalTasks)
	if progress < 0.3 {
		return HealthPoor
	}
	if progress < 0.7 {
		return HealthFair
	}
	return HealthGood
}

func (sm *StateManager) detectRisks(tasks []ListItem, stats *TaskStats) []string {
	risks := []string{}

	// ブロック状態のタスクが多い
	blockedCount := 0
	for _, task := range tasks {
		if task.Status == ItemStatusBlocked {
			blockedCount++
		}
	}
	if blockedCount > 0 {
		risks = append(risks, fmt.Sprintf("%d task(s) are blocked", blockedCount))
	}

	// 進行中タスクが多すぎる
	if stats.InProgress > 5 {
		risks = append(risks, "Too many tasks in progress (WIP limit exceeded)")
	}

	// 完了率が低い
	if stats.TotalTasks > 0 && float64(stats.Completed)/float64(stats.TotalTasks) < 0.2 {
		risks = append(risks, "Low completion rate")
	}

	return risks
}

func (sm *StateManager) saveSnapshot(ctx context.Context, snapshot *Snapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := sm.fileStore.EnsureDir(ctx, "state/snapshots"); err != nil {
		return err
	}

	// タイムスタンプをファイル名に使用（: を - に置換）
	filename := fmt.Sprintf("snapshot_%s.yaml", sanitizeTimestamp(snapshot.Timestamp))
	return sm.fileStore.WriteYaml(ctx, filepath.Join("state/snapshots", filename), snapshot)
}

func (sm *StateManager) getAllSnapshots(ctx context.Context) ([]Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := sm.fileStore.EnsureDir(ctx, "state/snapshots"); err != nil {
		return nil, err
	}

	// スナップショットディレクトリ内のファイルを列挙
	files, err := sm.fileStore.Glob(ctx, "state/snapshots/snapshot_*.yaml")
	if err != nil {
		return []Snapshot{}, nil
	}

	snapshots := []Snapshot{}
	for _, file := range files {
		var snapshot Snapshot
		if err := sm.fileStore.ReadYaml(ctx, file, &snapshot); err != nil {
			continue
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// sanitizeTimestamp はタイムスタンプをファイル名に使える形式に変換
func sanitizeTimestamp(ts string) string {
	var sb strings.Builder
	sb.Grow(len(ts))
	for _, c := range ts {
		if c == ':' || c == '+' {
			sb.WriteRune('-')
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}
