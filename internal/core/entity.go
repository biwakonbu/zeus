package core

import "context"

// EntityHandler はエンティティの操作を定義するインターフェース
// 新しいエンティティタイプを追加する場合は、このインターフェースを実装する
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
type EntityOption func(any)

// ListFilter はリストフィルタ
type ListFilter struct {
	Status string // ステータスでフィルタ
	Limit  int    // 最大件数
	Offset int    // オフセット
}

// EntityRegistry はエンティティハンドラーを管理するレジストリ
// 新しいエンティティハンドラーを登録することで、拡張可能
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
