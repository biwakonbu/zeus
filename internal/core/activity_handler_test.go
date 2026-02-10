package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupActivityHandlerTest(t *testing.T) (*ActivityHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-activity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/activities", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	activityHandler := NewActivityHandler(fs, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return activityHandler, zeusPath, cleanup
}

func TestActivityHandlerType(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	if handler.Type() != "activity" {
		t.Errorf("expected type 'activity', got %q", handler.Type())
	}
}

func TestActivityHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "activity" {
		t.Errorf("expected entity 'activity', got %q", result.Entity)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// リストで確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 activity, got %d", listResult.Total)
	}
}

func TestActivityHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きでアクティビティ追加
	result, err := handler.Add(ctx, "Test Activity with Options",
		WithActivityDescription("This is a test activity"),
		WithActivityStatus(ActivityStatusActive),
		WithActivityOwner("test-user"),
		WithActivityTags([]string{"auth", "flow"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// アクティビティを取得して確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if activity.Description != "This is a test activity" {
		t.Errorf("expected description 'This is a test activity', got %q", activity.Description)
	}

	if activity.Status != ActivityStatusActive {
		t.Errorf("expected status 'active', got %q", activity.Status)
	}

	if activity.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", activity.Metadata.Owner)
	}

	if len(activity.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(activity.Metadata.Tags))
	}
}

func TestActivityHandlerAddAllStatuses(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name   string
		status ActivityStatus
	}{
		{"Draft Activity", ActivityStatusDraft},
		{"Active Activity", ActivityStatusActive},
		{"Deprecated Activity", ActivityStatusDeprecated},
	}

	for _, tt := range tests {
		result, err := handler.Add(ctx, tt.name, WithActivityStatus(tt.status))
		if err != nil {
			t.Fatalf("Add %s failed: %v", tt.name, err)
		}

		activityAny, err := handler.Get(ctx, result.ID)
		if err != nil {
			t.Fatalf("Get %s failed: %v", tt.name, err)
		}

		activity := activityAny.(*ActivityEntity)
		if activity.Status != tt.status {
			t.Errorf("expected status %q, got %q", tt.status, activity.Status)
		}
	}

	// 全アクティビティがリストで確認できる
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 3 {
		t.Errorf("expected 3 activities, got %d", listResult.Total)
	}
}

func TestActivityHandlerList(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数アクティビティを追加
	for range 5 {
		_, err := handler.Add(ctx, "Activity")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// 全リスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 5 {
		t.Errorf("expected 5 activities, got %d", listResult.Total)
	}
}

func TestActivityHandlerListEmpty(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のリスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 activities, got %d", listResult.Total)
	}
}

func TestActivityHandlerGet(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Get Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// アクティビティを取得
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if activity.Title != "Get Test Activity" {
		t.Errorf("expected title 'Get Test Activity', got %q", activity.Title)
	}
}

func TestActivityHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得（有効なフォーマット）
	_, err := handler.Get(ctx, "act-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActivityHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Update Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 更新
	updateData := map[string]any{
		"title":       "Updated Title",
		"status":      "active",
		"description": "Updated description",
	}
	err = handler.Update(ctx, result.ID, updateData)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*ActivityEntity)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != ActivityStatusActive {
		t.Errorf("expected status 'active', got %q", updated.Status)
	}

	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %q", updated.Description)
	}
}

func TestActivityHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新（有効なフォーマット）
	err := handler.Update(ctx, "act-999", map[string]any{"title": "Test"})
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActivityHandlerDelete(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Delete Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 削除
	err = handler.Delete(ctx, result.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 削除されたことを確認
	_, err = handler.Get(ctx, result.ID)
	if err == nil {
		t.Error("expected error for deleted activity")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActivityHandlerDeleteNotFound(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除（有効なフォーマット）
	err := handler.Delete(ctx, "act-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActivityHandlerGetAll(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数アクティビティを追加
	for i := range 3 {
		_, err := handler.Add(ctx, "Activity "+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// GetAll で全件取得
	activities, err := handler.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(activities) != 3 {
		t.Errorf("expected 3 activities, got %d", len(activities))
	}
}

func TestActivityHandlerAddNode(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Test Activity with Nodes")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// ノードを追加
	node := ActivityNode{
		ID:   "node-001",
		Type: ActivityNodeTypeInitial,
	}
	err = handler.AddNode(ctx, result.ID, node)
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if len(activity.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(activity.Nodes))
	}

	if activity.Nodes[0].ID != "node-001" {
		t.Errorf("expected node ID 'node-001', got %q", activity.Nodes[0].ID)
	}

	if activity.Nodes[0].Type != ActivityNodeTypeInitial {
		t.Errorf("expected node type 'initial', got %q", activity.Nodes[0].Type)
	}
}

func TestActivityHandlerAddNodeWithName(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// アクションノードを追加
	node := ActivityNode{
		ID:   "node-002",
		Type: ActivityNodeTypeAction,
		Name: "ユーザーを認証する",
	}
	err = handler.AddNode(ctx, result.ID, node)
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if activity.Nodes[0].Name != "ユーザーを認証する" {
		t.Errorf("expected node name 'ユーザーを認証する', got %q", activity.Nodes[0].Name)
	}
}

func TestActivityHandlerAddNodeDuplicate(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// ノードを追加（1回目）
	node := ActivityNode{
		ID:   "node-001",
		Type: ActivityNodeTypeInitial,
	}
	err = handler.AddNode(ctx, result.ID, node)
	if err != nil {
		t.Fatalf("AddNode (first) failed: %v", err)
	}

	// 同じノードを再度追加（エラーになるはず）
	err = handler.AddNode(ctx, result.ID, node)
	if err == nil {
		t.Error("expected error for duplicate node")
	}
}

func TestActivityHandlerAddNodeAllTypes(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加
	result, err := handler.Add(ctx, "Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 全ノードタイプをテスト
	nodeTypes := []struct {
		id       string
		nodeType ActivityNodeType
		name     string
	}{
		{"node-initial", ActivityNodeTypeInitial, ""},
		{"node-final", ActivityNodeTypeFinal, ""},
		{"node-action", ActivityNodeTypeAction, "アクション"},
		{"node-decision", ActivityNodeTypeDecision, "分岐"},
		{"node-merge", ActivityNodeTypeMerge, "合流"},
		{"node-fork", ActivityNodeTypeFork, "並列分岐"},
		{"node-join", ActivityNodeTypeJoin, "並列合流"},
	}

	for _, nt := range nodeTypes {
		node := ActivityNode{
			ID:   nt.id,
			Type: nt.nodeType,
			Name: nt.name,
		}
		err = handler.AddNode(ctx, result.ID, node)
		if err != nil {
			t.Fatalf("AddNode %s failed: %v", nt.id, err)
		}
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if len(activity.Nodes) != 7 {
		t.Errorf("expected 7 nodes, got %d", len(activity.Nodes))
	}
}

func TestActivityHandlerAddTransition(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加（ノード付き）
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeInitial},
		{ID: "node-002", Type: ActivityNodeTypeAction, Name: "アクション"},
	}
	result, err := handler.Add(ctx, "Test Activity", WithActivityNodes(nodes))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 遷移を追加
	trans := ActivityTransition{
		ID:     "trans-001",
		Source: "node-001",
		Target: "node-002",
	}
	err = handler.AddTransition(ctx, result.ID, trans)
	if err != nil {
		t.Fatalf("AddTransition failed: %v", err)
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if len(activity.Transitions) != 1 {
		t.Errorf("expected 1 transition, got %d", len(activity.Transitions))
	}

	if activity.Transitions[0].ID != "trans-001" {
		t.Errorf("expected transition ID 'trans-001', got %q", activity.Transitions[0].ID)
	}
}

func TestActivityHandlerAddTransitionWithGuard(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加（ノード付き）
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeDecision, Name: "認証成功？"},
		{ID: "node-002", Type: ActivityNodeTypeAction, Name: "ホームへ遷移"},
	}
	result, err := handler.Add(ctx, "Test Activity", WithActivityNodes(nodes))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// ガード条件付き遷移を追加
	trans := ActivityTransition{
		ID:     "trans-001",
		Source: "node-001",
		Target: "node-002",
		Guard:  "[認証成功]",
	}
	err = handler.AddTransition(ctx, result.ID, trans)
	if err != nil {
		t.Fatalf("AddTransition failed: %v", err)
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if activity.Transitions[0].Guard != "[認証成功]" {
		t.Errorf("expected guard '[認証成功]', got %q", activity.Transitions[0].Guard)
	}
}

func TestActivityHandlerAddTransitionDuplicate(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加（ノード付き）
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeInitial},
		{ID: "node-002", Type: ActivityNodeTypeAction, Name: "アクション"},
	}
	result, err := handler.Add(ctx, "Test Activity", WithActivityNodes(nodes))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 遷移を追加（1回目）
	trans := ActivityTransition{
		ID:     "trans-001",
		Source: "node-001",
		Target: "node-002",
	}
	err = handler.AddTransition(ctx, result.ID, trans)
	if err != nil {
		t.Fatalf("AddTransition (first) failed: %v", err)
	}

	// 同じ遷移を再度追加（エラーになるはず）
	err = handler.AddTransition(ctx, result.ID, trans)
	if err == nil {
		t.Error("expected error for duplicate transition")
	}
}

func TestActivityHandlerAddTransitionSourceNotFound(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加（ノード付き）
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeInitial},
	}
	result, err := handler.Add(ctx, "Test Activity", WithActivityNodes(nodes))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 存在しないソースを参照する遷移
	trans := ActivityTransition{
		ID:     "trans-001",
		Source: "node-nonexistent",
		Target: "node-001",
	}
	err = handler.AddTransition(ctx, result.ID, trans)
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestActivityHandlerAddTransitionTargetNotFound(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクティビティ追加（ノード付き）
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeInitial},
	}
	result, err := handler.Add(ctx, "Test Activity", WithActivityNodes(nodes))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 存在しないターゲットを参照する遷移
	trans := ActivityTransition{
		ID:     "trans-001",
		Source: "node-001",
		Target: "node-nonexistent",
	}
	err = handler.AddTransition(ctx, result.ID, trans)
	if err == nil {
		t.Error("expected error for non-existent target")
	}
}

func TestActivityHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add
	_, err := handler.Add(ctx, "Test")
	if err == nil {
		t.Error("Add should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// List
	_, err = handler.List(ctx, nil)
	if err == nil {
		t.Error("List should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Get（有効なフォーマット）
	_, err = handler.Get(ctx, "act-00000000")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update（有効なフォーマット）
	err = handler.Update(ctx, "act-00000000", map[string]any{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete（有効なフォーマット）
	err = handler.Delete(ctx, "act-00000000")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// GetAll
	_, err = handler.GetAll(ctx)
	if err == nil {
		t.Error("GetAll should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateActivityIDFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-activity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zeusPath := tmpDir + "/.zeus"
	fs := yaml.NewFileManager(zeusPath)
	handler := NewActivityHandler(fs, nil)

	ctx := context.Background()

	// ID 生成テスト
	id, err := handler.generateActivityID(ctx)
	if err != nil {
		t.Fatalf("failed to generate activity ID: %v", err)
	}

	// プレフィックスが正しいか
	if len(id) < 4 || id[:4] != "act-" {
		t.Errorf("expected ID to start with 'act-', got %q", id)
	}

	// 長さが正しいか (act- + 8桁UUID = 12文字)
	if len(id) != 12 {
		t.Errorf("expected ID length to be 12, got %d", len(id))
	}
}

func TestActivityEntityValidate(t *testing.T) {
	tests := []struct {
		name     string
		activity ActivityEntity
		wantErr  bool
	}{
		{
			name: "valid activity",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			activity: ActivityEntity{
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "missing title",
			activity: ActivityEntity{
				ID:     "act-001",
				Status: ActivityStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty status (defaults to draft)",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: "",
			},
			wantErr: false,
		},
		{
			name: "valid with nodes",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Nodes: []ActivityNode{
					{ID: "node-001", Type: ActivityNodeTypeInitial},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid node - missing ID",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Nodes: []ActivityNode{
					{ID: "", Type: ActivityNodeTypeInitial},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid node - invalid type",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Nodes: []ActivityNode{
					{ID: "node-001", Type: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid with transitions",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Nodes: []ActivityNode{
					{ID: "node-001", Type: ActivityNodeTypeInitial},
					{ID: "node-002", Type: ActivityNodeTypeAction, Name: "アクション"},
				},
				Transitions: []ActivityTransition{
					{ID: "trans-001", Source: "node-001", Target: "node-002"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid transition - missing ID",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Transitions: []ActivityTransition{
					{ID: "", Source: "node-001", Target: "node-002"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid transition - missing source",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Transitions: []ActivityTransition{
					{ID: "trans-001", Source: "", Target: "node-002"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid transition - missing target",
			activity: ActivityEntity{
				ID:     "act-001",
				Title:  "Test Activity",
				Status: ActivityStatusDraft,
				Transitions: []ActivityTransition{
					{ID: "trans-001", Source: "node-001", Target: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "valid with usecase_id",
			activity: ActivityEntity{
				ID:        "act-001",
				Title:     "Test Activity",
				Status:    ActivityStatusDraft,
				UseCaseID: "uc-12345678",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.activity.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActivityNodeValidate(t *testing.T) {
	tests := []struct {
		name    string
		node    ActivityNode
		wantErr bool
	}{
		{
			name:    "valid initial node",
			node:    ActivityNode{ID: "node-001", Type: ActivityNodeTypeInitial},
			wantErr: false,
		},
		{
			name:    "valid final node",
			node:    ActivityNode{ID: "node-002", Type: ActivityNodeTypeFinal},
			wantErr: false,
		},
		{
			name:    "valid action node with name",
			node:    ActivityNode{ID: "node-003", Type: ActivityNodeTypeAction, Name: "アクション"},
			wantErr: false,
		},
		{
			name:    "valid decision node with name",
			node:    ActivityNode{ID: "node-004", Type: ActivityNodeTypeDecision, Name: "分岐"},
			wantErr: false,
		},
		{
			name:    "valid merge node",
			node:    ActivityNode{ID: "node-005", Type: ActivityNodeTypeMerge},
			wantErr: false,
		},
		{
			name:    "valid fork node",
			node:    ActivityNode{ID: "node-006", Type: ActivityNodeTypeFork},
			wantErr: false,
		},
		{
			name:    "valid join node",
			node:    ActivityNode{ID: "node-007", Type: ActivityNodeTypeJoin},
			wantErr: false,
		},
		{
			name:    "missing ID",
			node:    ActivityNode{ID: "", Type: ActivityNodeTypeInitial},
			wantErr: true,
		},
		{
			name:    "missing type",
			node:    ActivityNode{ID: "node-001", Type: ""},
			wantErr: true,
		},
		{
			name:    "invalid type",
			node:    ActivityNode{ID: "node-001", Type: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActivityTransitionValidate(t *testing.T) {
	tests := []struct {
		name    string
		trans   ActivityTransition
		wantErr bool
	}{
		{
			name:    "valid transition",
			trans:   ActivityTransition{ID: "trans-001", Source: "node-001", Target: "node-002"},
			wantErr: false,
		},
		{
			name:    "valid transition with guard",
			trans:   ActivityTransition{ID: "trans-002", Source: "node-001", Target: "node-002", Guard: "[条件]"},
			wantErr: false,
		},
		{
			name:    "missing ID",
			trans:   ActivityTransition{ID: "", Source: "node-001", Target: "node-002"},
			wantErr: true,
		},
		{
			name:    "missing source",
			trans:   ActivityTransition{ID: "trans-001", Source: "", Target: "node-002"},
			wantErr: true,
		},
		{
			name:    "missing target",
			trans:   ActivityTransition{ID: "trans-001", Source: "node-001", Target: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trans.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActivityHandlerWithNodesAndTransitions(t *testing.T) {
	handler, _, cleanup := setupActivityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ノードと遷移を持つアクティビティを追加
	nodes := []ActivityNode{
		{ID: "node-001", Type: ActivityNodeTypeInitial},
		{ID: "node-002", Type: ActivityNodeTypeAction, Name: "ログイン画面を表示"},
		{ID: "node-003", Type: ActivityNodeTypeDecision, Name: "認証成功？"},
		{ID: "node-004", Type: ActivityNodeTypeAction, Name: "ホーム画面を表示"},
		{ID: "node-005", Type: ActivityNodeTypeAction, Name: "エラー表示"},
		{ID: "node-006", Type: ActivityNodeTypeFinal},
	}
	transitions := []ActivityTransition{
		{ID: "trans-001", Source: "node-001", Target: "node-002"},
		{ID: "trans-002", Source: "node-002", Target: "node-003"},
		{ID: "trans-003", Source: "node-003", Target: "node-004", Guard: "[認証成功]"},
		{ID: "trans-004", Source: "node-003", Target: "node-005", Guard: "[認証失敗]"},
		{ID: "trans-005", Source: "node-004", Target: "node-006"},
		{ID: "trans-006", Source: "node-005", Target: "node-002"},
	}

	result, err := handler.Add(ctx, "ログインフロー",
		WithActivityDescription("ユーザー認証のアクティビティ図"),
		WithActivityStatus(ActivityStatusActive),
		WithActivityNodes(nodes),
		WithActivityTransitions(transitions),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 確認
	activityAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	activity := activityAny.(*ActivityEntity)
	if activity.Title != "ログインフロー" {
		t.Errorf("expected title 'ログインフロー', got %q", activity.Title)
	}

	if len(activity.Nodes) != 6 {
		t.Errorf("expected 6 nodes, got %d", len(activity.Nodes))
	}

	if len(activity.Transitions) != 6 {
		t.Errorf("expected 6 transitions, got %d", len(activity.Transitions))
	}

	// 特定のノードを確認
	foundDecision := false
	for _, node := range activity.Nodes {
		if node.Type == ActivityNodeTypeDecision && node.Name == "認証成功？" {
			foundDecision = true
			break
		}
	}
	if !foundDecision {
		t.Error("expected to find decision node '認証成功？'")
	}

	// ガード条件付き遷移を確認
	foundGuardedTransition := false
	for _, trans := range activity.Transitions {
		if trans.Guard == "[認証成功]" {
			foundGuardedTransition = true
			break
		}
	}
	if !foundGuardedTransition {
		t.Error("expected to find transition with guard '[認証成功]'")
	}
}
