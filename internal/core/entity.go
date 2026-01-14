package core

import "context"

// EntityHandler はエンティティの CRUD 操作を定義するインターフェース
//
// 新エンティティの追加手順:
//  1. EntityHandler インターフェースを実装
//  2. NewZeus() で EntityRegistry に登録
//  3. cmd/ に対応するコマンドを追加（オプショナル）
//
// 実装例 (IssueHandler):
//
//	type IssueHandler struct {
//	    fileStore core.FileStore
//	}
//
//	func NewIssueHandler(fs core.FileStore) *IssueHandler {
//	    return &IssueHandler{fileStore: fs}
//	}
//
//	func (h *IssueHandler) Type() string {
//	    return "issue"
//	}
//
//	func (h *IssueHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
//	    // Issue 固有のロジック
//	    return &AddResult{Success: true, ID: "issue-1", Entity: "issue"}, nil
//	}
//
//	// ... 他のメソッド実装 ...
//
// 登録例 (NewZeus内):
//
//	registry := NewEntityRegistry()
//	registry.Register(NewTaskHandler(z.fileStore))
//	registry.Register(NewIssueHandler(z.fileStore))
type EntityHandler interface {
	// Type はエンティティタイプを返す（例: "task", "objective", "milestone"）
	Type() string

	// Add はエンティティを追加
	Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error)

	// List はエンティティ一覧を取得
	List(ctx context.Context, filter *ListFilter) (*ListResult, error)

	// Get はエンティティを取得
	Get(ctx context.Context, id string) (any, error)

	// Update はエンティティを更新
	Update(ctx context.Context, id string, update any) error

	// Delete はエンティティを削除
	Delete(ctx context.Context, id string) error
}

// EntityOption はエンティティ作成時のオプション
//
// 使用例:
//
//	result, err := handler.Add(ctx, "Task Name",
//	    WithPriority("high"),
//	    WithAssignee("alice"),
//	)
type EntityOption func(any)

// ListFilter はリストフィルタリング条件
//
// 使用例:
//
//	filter := &ListFilter{
//	    Status: "active",
//	    Limit:  10,
//	    Offset: 0,
//	}
//	result, err := handler.List(ctx, filter)
type ListFilter struct {
	Status string // ステータスでフィルタ（"active", "completed" など）
	Limit  int    // 取得件数上限（0 = 無制限）
	Offset int    // 取得開始位置
}

// EntityRegistry はエンティティハンドラーを管理するレジストリ
//
// 使用例:
//
//	registry := NewEntityRegistry()
//	registry.Register(NewTaskHandler(fileStore))
//
//	handler, ok := registry.Get("task")
//	if !ok {
//	    return fmt.Errorf("unknown entity type: task")
//	}
//	result, err := handler.Add(ctx, "New Task")
type EntityRegistry struct {
	handlers map[string]EntityHandler
}

// NewEntityRegistry は新しい EntityRegistry を作成
func NewEntityRegistry() *EntityRegistry {
	return &EntityRegistry{
		handlers: make(map[string]EntityHandler),
	}
}

// Register はエンティティハンドラーを登録
func (r *EntityRegistry) Register(handler EntityHandler) {
	r.handlers[handler.Type()] = handler
}

// Get はエンティティハンドラーを取得
func (r *EntityRegistry) Get(entityType string) (EntityHandler, bool) {
	h, ok := r.handlers[entityType]
	return h, ok
}

// Types は登録済みエンティティタイプを取得
func (r *EntityRegistry) Types() []string {
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
