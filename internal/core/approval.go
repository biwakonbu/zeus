package core

import (
	"path/filepath"
	"time"

	"github.com/biwakonbu/zeus/internal/yaml"
	"github.com/google/uuid"
)

// ApprovalStatus は承認ステータス
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// PendingApproval は承認待ちアイテム
type PendingApproval struct {
	ID          string         `yaml:"id"`
	Type        string         `yaml:"type"` // task_create, task_update, suggestion
	Description string         `yaml:"description"`
	Level       ApprovalLevel  `yaml:"level"`
	Status      ApprovalStatus `yaml:"status"`
	EntityID    string         `yaml:"entity_id,omitempty"`
	Payload     any            `yaml:"payload,omitempty"`
	CreatedAt   string         `yaml:"created_at"`
	UpdatedAt   string         `yaml:"updated_at"`
	ApprovedBy  string         `yaml:"approved_by,omitempty"`
	RejectedBy  string         `yaml:"rejected_by,omitempty"`
	Reason      string         `yaml:"reason,omitempty"`
}

// ApprovalStore は承認ストア
type ApprovalStore struct {
	Approvals []PendingApproval `yaml:"approvals"`
}

// ApprovalResult は承認・却下結果
type ApprovalResult struct {
	Success bool
	ID      string
	Status  ApprovalStatus
}

// ApprovalManager は承認を管理
type ApprovalManager struct {
	zeusPath    string
	fileManager *yaml.FileManager
	lock        *yaml.FileLock
}

// NewApprovalManager は新しい ApprovalManager を作成
func NewApprovalManager(zeusPath string) *ApprovalManager {
	queuePath := filepath.Join(zeusPath, "approvals", "pending", "queue.yaml")
	return &ApprovalManager{
		zeusPath:    zeusPath,
		fileManager: yaml.NewFileManager(zeusPath),
		lock:        yaml.NewFileLock(queuePath),
	}
}

// generateApprovalID はユニークな承認 ID を生成
// UUID v4 を使用して衝突を防止
func (am *ApprovalManager) generateApprovalID() string {
	return "approval-" + uuid.New().String()[:8]
}

// GetPending は承認待ちアイテムを取得
func (am *ApprovalManager) GetPending() ([]PendingApproval, error) {
	var store ApprovalStore
	if err := am.fileManager.ReadYaml("approvals/pending/queue.yaml", &store); err != nil {
		// ファイルが存在しない場合は空のリストを返す
		return []PendingApproval{}, nil
	}

	// ステータスが pending のもののみ返す
	pending := []PendingApproval{}
	for _, a := range store.Approvals {
		if a.Status == ApprovalStatusPending {
			pending = append(pending, a)
		}
	}

	return pending, nil
}

// GetAll は全承認アイテムを取得
func (am *ApprovalManager) GetAll() ([]PendingApproval, error) {
	var store ApprovalStore
	if err := am.fileManager.ReadYaml("approvals/pending/queue.yaml", &store); err != nil {
		return []PendingApproval{}, nil
	}
	return store.Approvals, nil
}

// Get は特定の承認アイテムを取得
func (am *ApprovalManager) Get(id string) (*PendingApproval, error) {
	all, err := am.GetAll()
	if err != nil {
		return nil, err
	}

	for _, a := range all {
		if a.ID == id {
			return &a, nil
		}
	}

	return nil, ErrEntityNotFound
}

// Create は新しい承認アイテムを作成（原子的操作）
func (am *ApprovalManager) Create(approvalType, description string, level ApprovalLevel, entityID string, payload any) (*PendingApproval, error) {
	// ロックを取得（タイムアウト: 5秒）
	if err := am.lock.LockWithTimeout(5 * time.Second); err != nil {
		return nil, ErrLockAcquireFailed
	}
	defer am.lock.Unlock()

	all, err := am.GetAll()
	if err != nil {
		return nil, err
	}

	// UUID ベースの ID 生成（衝突防止）
	id := am.generateApprovalID()
	now := Now()

	approval := PendingApproval{
		ID:          id,
		Type:        approvalType,
		Description: description,
		Level:       level,
		Status:      ApprovalStatusPending,
		EntityID:    entityID,
		Payload:     payload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	all = append(all, approval)
	store := ApprovalStore{Approvals: all}

	if err := am.fileManager.EnsureDir("approvals/pending"); err != nil {
		return nil, err
	}
	if err := am.fileManager.WriteYaml("approvals/pending/queue.yaml", &store); err != nil {
		return nil, err
	}

	return &approval, nil
}

// Approve は承認アイテムを承認（原子的操作）
func (am *ApprovalManager) Approve(id string) (*ApprovalResult, error) {
	// ロックを取得（タイムアウト: 5秒）
	if err := am.lock.LockWithTimeout(5 * time.Second); err != nil {
		return nil, ErrLockAcquireFailed
	}
	defer am.lock.Unlock()

	all, err := am.GetAll()
	if err != nil {
		return nil, err
	}

	found := false
	var approvedItem PendingApproval
	for i, a := range all {
		if a.ID == id {
			if a.Status != ApprovalStatusPending {
				return nil, &ApprovalNotPendingError{
					ID:            id,
					CurrentStatus: a.Status,
				}
			}
			all[i].Status = ApprovalStatusApproved
			all[i].UpdatedAt = Now()
			all[i].ApprovedBy = "user"
			approvedItem = all[i]
			found = true
			break
		}
	}

	if !found {
		return nil, ErrEntityNotFound
	}

	// 承認済みファイルに移動
	if err := am.moveToApproved(approvedItem); err != nil {
		return nil, err
	}

	// 元のキューから削除
	remaining := []PendingApproval{}
	for _, a := range all {
		if a.ID != id {
			remaining = append(remaining, a)
		}
	}
	store := ApprovalStore{Approvals: remaining}
	if err := am.fileManager.WriteYaml("approvals/pending/queue.yaml", &store); err != nil {
		return nil, err
	}

	return &ApprovalResult{
		Success: true,
		ID:      id,
		Status:  ApprovalStatusApproved,
	}, nil
}

// Reject は承認アイテムを却下（原子的操作）
func (am *ApprovalManager) Reject(id, reason string) (*ApprovalResult, error) {
	// ロックを取得（タイムアウト: 5秒）
	if err := am.lock.LockWithTimeout(5 * time.Second); err != nil {
		return nil, ErrLockAcquireFailed
	}
	defer am.lock.Unlock()

	all, err := am.GetAll()
	if err != nil {
		return nil, err
	}

	found := false
	var rejectedItem PendingApproval
	for i, a := range all {
		if a.ID == id {
			if a.Status != ApprovalStatusPending {
				return nil, &ApprovalNotPendingError{
					ID:            id,
					CurrentStatus: a.Status,
				}
			}
			all[i].Status = ApprovalStatusRejected
			all[i].UpdatedAt = Now()
			all[i].RejectedBy = "user"
			all[i].Reason = reason
			rejectedItem = all[i]
			found = true
			break
		}
	}

	if !found {
		return nil, ErrEntityNotFound
	}

	// 却下済みファイルに移動
	if err := am.moveToRejected(rejectedItem); err != nil {
		return nil, err
	}

	// 元のキューから削除
	remaining := []PendingApproval{}
	for _, a := range all {
		if a.ID != id {
			remaining = append(remaining, a)
		}
	}
	store := ApprovalStore{Approvals: remaining}
	if err := am.fileManager.WriteYaml("approvals/pending/queue.yaml", &store); err != nil {
		return nil, err
	}

	return &ApprovalResult{
		Success: true,
		ID:      id,
		Status:  ApprovalStatusRejected,
	}, nil
}

// DetermineApprovalLevel はアクションに応じた承認レベルを決定
func (am *ApprovalManager) DetermineApprovalLevel(actionType string, settings *Settings) ApprovalLevel {
	// 承認モードに応じてデフォルトレベルを決定
	switch settings.ApprovalMode {
	case "strict":
		// 厳格モード: ほとんどの操作で承認が必要
		switch actionType {
		case "task_create", "task_update", "suggestion":
			return ApprovalApprove
		default:
			return ApprovalNotify
		}
	case "loose":
		// 緩いモード: 自動承認が多い
		switch actionType {
		case "suggestion":
			return ApprovalNotify
		default:
			return ApprovalAuto
		}
	default:
		// デフォルトモード: バランス型
		switch actionType {
		case "suggestion":
			return ApprovalApprove
		case "task_update":
			return ApprovalNotify
		default:
			return ApprovalAuto
		}
	}
}

// moveToApproved は承認済みファイルに移動
func (am *ApprovalManager) moveToApproved(approval PendingApproval) error {
	if err := am.fileManager.EnsureDir("approvals/approved"); err != nil {
		return err
	}

	filename := approval.ID + ".yaml"
	return am.fileManager.WriteYaml(filepath.Join("approvals/approved", filename), &approval)
}

// moveToRejected は却下済みファイルに移動
func (am *ApprovalManager) moveToRejected(approval PendingApproval) error {
	if err := am.fileManager.EnsureDir("approvals/rejected"); err != nil {
		return err
	}

	filename := approval.ID + ".yaml"
	return am.fileManager.WriteYaml(filepath.Join("approvals/rejected", filename), &approval)
}
