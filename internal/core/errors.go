package core

import (
	"errors"
	"fmt"
)

// 基本エラー定義
var (
	// ErrConfigNotFound は設定ファイルが見つからない
	ErrConfigNotFound = errors.New("zeus.yaml not found")
	// ErrYamlSyntax は YAML 構文エラー
	ErrYamlSyntax = errors.New("YAML syntax error")
	// ErrStateMismatch は状態の不整合
	ErrStateMismatch = errors.New("state mismatch")
	// ErrEntityNotFound はエンティティが見つからない
	ErrEntityNotFound = errors.New("entity not found")
	// ErrUnknownEntity は不明なエンティティ
	ErrUnknownEntity = errors.New("unknown entity type")
)

// セキュリティ関連エラー
var (
	// ErrPathTraversal はディレクトリトラバーサル攻撃を検出
	ErrPathTraversal = errors.New("path traversal detected: access outside base directory is not allowed")
	// ErrInvalidPath は不正なパス
	ErrInvalidPath = errors.New("invalid path")
)

// 並行処理関連エラー
var (
	// ErrLockAcquireFailed はロック取得失敗
	ErrLockAcquireFailed = errors.New("failed to acquire file lock")
	// ErrLockTimeout はロックタイムアウト
	ErrLockTimeout = errors.New("lock acquisition timed out")
)

// 承認関連エラー
var (
	// ErrApprovalNotPending は承認待ち状態でない
	ErrApprovalNotPending = errors.New("approval is not in pending state")
)

// ApprovalNotPendingError は承認待ち状態でないエラー（詳細情報付き）
type ApprovalNotPendingError struct {
	ID            string
	CurrentStatus ApprovalStatus
}

func (e *ApprovalNotPendingError) Error() string {
	return fmt.Sprintf("approval %s is not pending (current status: %s)", e.ID, e.CurrentStatus)
}

func (e *ApprovalNotPendingError) Is(target error) bool {
	return target == ErrApprovalNotPending
}

// PathTraversalError はパストラバーサルエラー（詳細情報付き）
type PathTraversalError struct {
	RequestedPath string
	BasePath      string
}

func (e *PathTraversalError) Error() string {
	return fmt.Sprintf("path traversal detected: %s is outside base directory %s", e.RequestedPath, e.BasePath)
}

func (e *PathTraversalError) Is(target error) bool {
	return target == ErrPathTraversal
}
