package core

import (
	"context"
	"os"
	"testing"
)

func TestNewEntityRegistry(t *testing.T) {
	r := NewEntityRegistry()
	if r == nil {
		t.Error("NewEntityRegistry should return non-nil")
	}
	if r.handlers == nil {
		t.Error("handlers map should be initialized")
	}
}

func TestEntityRegistry_Register(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "entity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	r := NewEntityRegistry()

	// タスクハンドラーを登録
	handler := NewTaskHandler(z.fileStore)
	r.Register(handler)

	// 登録されたか確認
	h, ok := r.Get("task")
	if !ok {
		t.Error("task handler should be registered")
	}
	if h == nil {
		t.Error("retrieved handler should not be nil")
	}
}

func TestEntityRegistry_Get(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "entity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	r := NewEntityRegistry()
	handler := NewTaskHandler(z.fileStore)
	r.Register(handler)

	// 存在するタイプ
	h, ok := r.Get("task")
	if !ok {
		t.Error("should find registered handler")
	}
	if h.Type() != "task" {
		t.Errorf("expected type 'task', got %q", h.Type())
	}

	// 存在しないタイプ
	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("should not find non-registered handler")
	}
}

func TestEntityRegistry_Types(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "entity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	r := NewEntityRegistry()

	// 空の状態
	types := r.Types()
	if len(types) != 0 {
		t.Errorf("expected 0 types, got %d", len(types))
	}

	// ハンドラーを登録
	handler := NewTaskHandler(z.fileStore)
	r.Register(handler)

	types = r.Types()
	if len(types) != 1 {
		t.Errorf("expected 1 type, got %d", len(types))
	}
	if types[0] != "task" {
		t.Errorf("expected type 'task', got %q", types[0])
	}
}

func TestEntityRegistry_RegisterOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "entity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	r := NewEntityRegistry()

	// 同じタイプを2回登録
	handler1 := NewTaskHandler(z.fileStore)
	handler2 := NewTaskHandler(z.fileStore)
	r.Register(handler1)
	r.Register(handler2)

	// 最後に登録したものが取得されるべき
	types := r.Types()
	if len(types) != 1 {
		t.Errorf("expected 1 type after overwrite, got %d", len(types))
	}
}

// MockEntityHandler はテスト用のモックハンドラー
type MockEntityHandler struct {
	entityType string
	addCalled  bool
	listCalled bool
}

func (m *MockEntityHandler) Type() string {
	return m.entityType
}

func (m *MockEntityHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	m.addCalled = true
	return &AddResult{Success: true, ID: "mock-1", Entity: m.entityType}, nil
}

func (m *MockEntityHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	m.listCalled = true
	return &ListResult{Entity: m.entityType, Items: []Task{}, Total: 0}, nil
}

func (m *MockEntityHandler) Get(ctx context.Context, id string) (any, error) {
	return nil, nil
}

func (m *MockEntityHandler) Update(ctx context.Context, id string, update any) error {
	return nil
}

func (m *MockEntityHandler) Delete(ctx context.Context, id string) error {
	return nil
}

func TestEntityRegistry_CustomHandler(t *testing.T) {
	r := NewEntityRegistry()

	// カスタムハンドラーを登録
	mock := &MockEntityHandler{entityType: "custom"}
	r.Register(mock)

	// 取得して使用
	h, ok := r.Get("custom")
	if !ok {
		t.Error("should find custom handler")
	}

	ctx := context.Background()
	_, _ = h.Add(ctx, "test")
	if !mock.addCalled {
		t.Error("Add should have been called")
	}

	_, _ = h.List(ctx, nil)
	if !mock.listCalled {
		t.Error("List should have been called")
	}
}

func TestListFilter(t *testing.T) {
	// ListFilter の構造体テスト
	filter := &ListFilter{
		Status: "active",
		Limit:  10,
		Offset: 5,
	}

	if filter.Status != "active" {
		t.Errorf("expected Status 'active', got %q", filter.Status)
	}
	if filter.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", filter.Limit)
	}
	if filter.Offset != 5 {
		t.Errorf("expected Offset 5, got %d", filter.Offset)
	}
}
