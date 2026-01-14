package core

import "context"

// FileStore はファイル操作の抽象化インターフェース
//
// 実装例:
//   - yaml.FileManager: 実際のファイルシステム操作
//   - mocks.MockFileStore: テスト用のインメモリ実装
//
// 使用例:
//
//	zeus := core.NewZeus(".", core.WithFileStore(myFileStore))
//
// テストでのモック注入例:
//
//	mockFS := mocks.NewMockFileStore()
//	zeus := core.NewZeus(".", core.WithFileStore(mockFS))
//	result, err := zeus.Init(ctx, "simple")
type FileStore interface {
	// Exists はファイルが存在するか確認
	Exists(ctx context.Context, path string) bool

	// ReadYaml は YAML ファイルを読み込む
	ReadYaml(ctx context.Context, path string, v any) error

	// WriteYaml は YAML ファイルを書き込む
	WriteYaml(ctx context.Context, path string, data any) error

	// EnsureDir はディレクトリを作成（存在しない場合）
	EnsureDir(ctx context.Context, path string) error

	// Delete はファイルを削除
	Delete(ctx context.Context, path string) error

	// Glob はパターンに一致するファイルを検索
	Glob(ctx context.Context, pattern string) ([]string, error)

	// WriteFile はバイナリファイルを書き込む
	WriteFile(ctx context.Context, path string, data []byte) error

	// Copy はファイルをコピー
	Copy(ctx context.Context, src, dest string) error

	// ListDir はディレクトリ内のファイルを列挙
	ListDir(ctx context.Context, path string) ([]string, error)

	// BasePath はベースパスを返す
	BasePath() string
}

// StateStore は状態管理の抽象化インターフェース
//
// 実装例:
//   - StateManager: 実際の状態管理実装
//   - mocks.MockStateStore: テスト用のインメモリ実装
//
// 使用例:
//
//	zeus := core.NewZeus(".", core.WithStateStore(myStateStore))
//
// テストでのモック注入例:
//
//	mockState := mocks.NewMockStateStore()
//	zeus := core.NewZeus(".", core.WithStateStore(mockState))
//	result, err := zeus.Status(ctx)
type StateStore interface {
	// GetCurrentState は現在の状態を取得
	GetCurrentState(ctx context.Context) (*ProjectState, error)

	// SaveCurrentState は現在の状態を保存
	SaveCurrentState(ctx context.Context, state *ProjectState) error

	// CreateSnapshot はスナップショットを作成
	CreateSnapshot(ctx context.Context, label string) (*Snapshot, error)

	// GetHistory はスナップショット履歴を取得
	GetHistory(ctx context.Context, limit int) ([]Snapshot, error)

	// GetSnapshot は特定のスナップショットを取得
	GetSnapshot(ctx context.Context, timestamp string) (*Snapshot, error)

	// RestoreSnapshot はスナップショットから復元
	RestoreSnapshot(ctx context.Context, timestamp string) error

	// CalculateState はタスクから状態を計算
	CalculateState(tasks []Task) *ProjectState
}

// ApprovalStore は承認管理の抽象化インターフェース
//
// 実装例:
//   - ApprovalManager: 実際の承認管理実装
//   - mocks.MockApprovalStore: テスト用のインメモリ実装
//
// 使用例:
//
//	zeus := core.NewZeus(".", core.WithApprovalStore(myApprovalStore))
//
// テストでのモック注入例:
//
//	mockApproval := mocks.NewMockApprovalStore()
//	zeus := core.NewZeus(".", core.WithApprovalStore(mockApproval))
//	approvals, err := zeus.Pending(ctx)
type ApprovalStore interface {
	// GetPending は承認待ちアイテムを取得
	GetPending(ctx context.Context) ([]PendingApproval, error)

	// GetAll は全承認アイテムを取得
	GetAll(ctx context.Context) ([]PendingApproval, error)

	// Get は特定の承認アイテムを取得
	Get(ctx context.Context, id string) (*PendingApproval, error)

	// Create は新しい承認アイテムを作成
	Create(ctx context.Context, approvalType, description string, level ApprovalLevel, entityID string, payload any) (*PendingApproval, error)

	// Approve は承認アイテムを承認
	Approve(ctx context.Context, id string) (*ApprovalResult, error)

	// Reject は承認アイテムを却下
	Reject(ctx context.Context, id, reason string) (*ApprovalResult, error)

	// DetermineApprovalLevel はアクションに応じた承認レベルを決定
	DetermineApprovalLevel(actionType string, settings *Settings) ApprovalLevel
}
