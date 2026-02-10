package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// ===== テストセットアップ =====

// setupQualityHandlerTest は QualityHandler テスト用のセットアップを行う
func setupQualityHandlerTest(t *testing.T) (*QualityHandler, *ObjectiveHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-quality-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusPath, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	// quality ディレクトリ作成
	qualityDir := filepath.Join(zeusPath, "quality")
	if err := os.MkdirAll(qualityDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create quality dir: %v", err)
	}

	// objectives ディレクトリ作成
	objectivesDir := filepath.Join(zeusPath, "objectives")
	if err := os.MkdirAll(objectivesDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create objectives dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)

	// 依存ハンドラーを作成
	objHandler := NewObjectiveHandler(fs, nil)
	qualHandler := NewQualityHandler(fs, objHandler, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return qualHandler, objHandler, zeusPath, cleanup
}

// createTestObjective はテスト用の Objective を作成する
func createTestObjective(t *testing.T, objHandler *ObjectiveHandler) string {
	t.Helper()
	ctx := context.Background()

	// Objective を作成
	result, err := objHandler.Add(ctx, "テスト目標")
	if err != nil {
		t.Fatalf("failed to create test objective: %v", err)
	}

	return result.ID
}

// defaultMetrics はテスト用のデフォルトメトリクスを返す（最低 1 つ必要）
func defaultMetrics() []QualityMetric {
	return []QualityMetric{
		{
			ID:      "metric-001",
			Name:    "テストメトリクス",
			Target:  100.0,
			Unit:    "%",
			Current: 50.0,
			Status:  MetricStatusInProgress,
		},
	}
}

// ===== 基本テスト =====

func TestQualityHandler_Type(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	if handler.Type() != "quality" {
		t.Errorf("Type() = %q, want %q", handler.Type(), "quality")
	}
}

func TestQualityHandler_Add(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	metrics := []QualityMetric{
		{
			ID:      "metric-001",
			Name:    "コードカバレッジ",
			Target:  80.0,
			Unit:    "%",
			Current: 65.0,
			Status:  MetricStatusInProgress,
		},
	}

	gates := []QualityGate{
		{
			Name:     "コードレビュー完了",
			Criteria: []string{"全ファイルがレビュー済み", "指摘事項が解決済み"},
			Status:   GateStatusPending,
		},
	}

	result, err := handler.Add(ctx, "品質基準1",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(metrics),
		WithQualityGates(gates),
		WithQualityReviewer("reviewer@example.com"),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Add() Success = false, want true")
	}
	if !strings.HasPrefix(result.ID, "qual-") {
		t.Errorf("Add() ID = %q, expected prefix 'qual-'", result.ID)
	}
	if result.Entity != "quality" {
		t.Errorf("Add() Entity = %q, want %q", result.Entity, "quality")
	}

	// 取得して内容を確認
	got, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)
	if qual.Title != "品質基準1" {
		t.Errorf("Title = %q, want %q", qual.Title, "品質基準1")
	}
	if qual.ObjectiveID != objectiveID {
		t.Errorf("ObjectiveID = %q, want %q", qual.ObjectiveID, objectiveID)
	}
	if qual.Reviewer != "reviewer@example.com" {
		t.Errorf("Reviewer = %q, want %q", qual.Reviewer, "reviewer@example.com")
	}
	if len(qual.Metrics) != 1 {
		t.Errorf("Metrics count = %d, want 1", len(qual.Metrics))
	}
	if len(qual.Gates) != 1 {
		t.Errorf("Gates count = %d, want 1", len(qual.Gates))
	}
}

func TestQualityHandler_Add_ObjectiveRequired(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ObjectiveID なしで追加
	_, err := handler.Add(ctx, "品質基準")
	if err == nil {
		t.Error("Add() without ObjectiveID should fail")
	}
	if err.Error() != "quality objective_id is required" {
		t.Errorf("error = %q, want 'quality objective_id is required'", err.Error())
	}
}

func TestQualityHandler_Add_InvalidObjectiveReference(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Objective を参照
	_, err := handler.Add(ctx, "品質基準",
		WithQualityObjective("obj-999"),
	)
	if err == nil {
		t.Error("Add() with invalid ObjectiveID should fail")
	}
}

func TestQualityHandler_Add_InvalidInput(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	testCases := []struct {
		name  string
		title string
	}{
		{"empty title", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := handler.Add(ctx, tc.title,
				WithQualityObjective(objectiveID),
			)
			if err == nil {
				t.Error("expected error for invalid input")
			}
		})
	}
}

func TestQualityHandler_List(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	// 3つ追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "品質基準",
			WithQualityObjective(objectiveID),
			WithQualityMetrics(defaultMetrics()),
		)
		if err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() Total = %d, want 3", result.Total)
	}
	if result.Entity != "quality" {
		t.Errorf("List() Entity = %q, want %q", result.Entity, "quality")
	}
}

func TestQualityHandler_List_Empty(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("List() Total = %d, want 0", result.Total)
	}
}

func TestQualityHandler_List_WithLimit(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	// 5つ追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "品質基準",
			WithQualityObjective(objectiveID),
			WithQualityMetrics(defaultMetrics()),
		)
		if err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	result, err := handler.List(ctx, &ListFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() Total = %d, want 3", result.Total)
	}
}

func TestQualityHandler_Get(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	addResult, err := handler.Add(ctx, "テスト品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
		WithQualityReviewer("test-reviewer"),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	got, err := handler.Get(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)
	if qual.ID != addResult.ID {
		t.Errorf("ID = %q, want %q", qual.ID, addResult.ID)
	}
	if qual.Title != "テスト品質基準" {
		t.Errorf("Title = %q, want %q", qual.Title, "テスト品質基準")
	}
	if qual.ObjectiveID != objectiveID {
		t.Errorf("ObjectiveID = %q, want %q", qual.ObjectiveID, objectiveID)
	}
}

func TestQualityHandler_Get_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "qual-999")
	if err != ErrEntityNotFound {
		t.Errorf("Get() error = %v, want ErrEntityNotFound", err)
	}
}

func TestQualityHandler_Get_InvalidID(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	testCases := []string{
		"invalid",
		"qual-1",    // 短すぎ
		"qual-1234", // 長すぎ
		"quality-001",
	}

	for _, id := range testCases {
		t.Run(id, func(t *testing.T) {
			_, err := handler.Get(ctx, id)
			if err == nil {
				t.Error("expected error for invalid ID")
			}
		})
	}
}

func TestQualityHandler_Update(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	addResult, err := handler.Add(ctx, "初期品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 更新
	updateQual := &QualityEntity{
		ID:          addResult.ID,
		Title:       "更新後品質基準",
		ObjectiveID: objectiveID,
		Reviewer:    "new-reviewer",
		Metrics: []QualityMetric{
			{
				ID:      "metric-001",
				Name:    "テストカバレッジ",
				Target:  90.0,
				Unit:    "%",
				Current: 85.0,
				Status:  MetricStatusMet,
			},
		},
	}

	err = handler.Update(ctx, addResult.ID, updateQual)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// 更新確認
	got, err := handler.Get(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)
	if qual.Title != "更新後品質基準" {
		t.Errorf("Title = %q, want %q", qual.Title, "更新後品質基準")
	}
	if qual.Reviewer != "new-reviewer" {
		t.Errorf("Reviewer = %q, want %q", qual.Reviewer, "new-reviewer")
	}
	if len(qual.Metrics) != 1 {
		t.Errorf("Metrics count = %d, want 1", len(qual.Metrics))
	}
}

func TestQualityHandler_Update_NotFound(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	updateQual := &QualityEntity{
		ID:          "qual-999",
		Title:       "存在しない品質基準",
		ObjectiveID: objectiveID,
	}

	err := handler.Update(ctx, "qual-999", updateQual)
	if err != ErrEntityNotFound {
		t.Errorf("Update() error = %v, want ErrEntityNotFound", err)
	}
}

func TestQualityHandler_Delete(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	addResult, err := handler.Add(ctx, "削除対象品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 削除
	err = handler.Delete(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// 削除確認
	_, err = handler.Get(ctx, addResult.ID)
	if err != ErrEntityNotFound {
		t.Errorf("Get() after Delete() error = %v, want ErrEntityNotFound", err)
	}
}

func TestQualityHandler_Delete_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "qual-999")
	if err != ErrEntityNotFound {
		t.Errorf("Delete() error = %v, want ErrEntityNotFound", err)
	}
}

// ===== GetQualitiesByObjective テスト =====

func TestQualityHandler_GetQualitiesByObjective(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 2つの Objective を作成
	objectiveID1 := createTestObjective(t, objHandler)
	objectiveID2 := createTestObjective(t, objHandler)

	// objectiveID1 に 2つの Quality を紐付け
	_, err := handler.Add(ctx, "品質基準A",
		WithQualityObjective(objectiveID1),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	_, err = handler.Add(ctx, "品質基準B",
		WithQualityObjective(objectiveID1),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// objectiveID2 に 1つの Quality を紐付け
	_, err = handler.Add(ctx, "品質基準C",
		WithQualityObjective(objectiveID2),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// objectiveID1 の Quality を取得
	qualities, err := handler.GetQualitiesByObjective(ctx, objectiveID1)
	if err != nil {
		t.Fatalf("GetQualitiesByObjective() failed: %v", err)
	}

	if len(qualities) != 2 {
		t.Errorf("GetQualitiesByObjective() count = %d, want 2", len(qualities))
	}

	// objectiveID2 の Quality を取得
	qualities, err = handler.GetQualitiesByObjective(ctx, objectiveID2)
	if err != nil {
		t.Fatalf("GetQualitiesByObjective() failed: %v", err)
	}

	if len(qualities) != 1 {
		t.Errorf("GetQualitiesByObjective() count = %d, want 1", len(qualities))
	}
}

func TestQualityHandler_GetQualitiesByObjective_Empty(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	// Quality を追加しない

	qualities, err := handler.GetQualitiesByObjective(ctx, objectiveID)
	if err != nil {
		t.Fatalf("GetQualitiesByObjective() failed: %v", err)
	}

	if len(qualities) != 0 {
		t.Errorf("GetQualitiesByObjective() count = %d, want 0", len(qualities))
	}
}

// ===== UpdateMetric テスト =====

func TestQualityHandler_UpdateMetric(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	metrics := []QualityMetric{
		{
			ID:      "metric-001",
			Name:    "コードカバレッジ",
			Target:  80.0,
			Unit:    "%",
			Current: 65.0,
			Status:  MetricStatusInProgress,
		},
		{
			ID:      "metric-002",
			Name:    "パフォーマンス",
			Target:  100.0,
			Unit:    "ms",
			Current: 150.0,
			Status:  MetricStatusNotMet,
		},
	}

	addResult, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(metrics),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// metric-001 を更新
	err = handler.UpdateMetric(ctx, addResult.ID, "metric-001", 85.0, MetricStatusMet)
	if err != nil {
		t.Fatalf("UpdateMetric() failed: %v", err)
	}

	// 更新確認
	got, err := handler.Get(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)
	var foundMetric *QualityMetric
	for i := range qual.Metrics {
		if qual.Metrics[i].ID == "metric-001" {
			foundMetric = &qual.Metrics[i]
			break
		}
	}

	if foundMetric == nil {
		t.Fatal("metric-001 not found")
	}

	if foundMetric.Current != 85.0 {
		t.Errorf("metric Current = %f, want 85.0", foundMetric.Current)
	}
	if foundMetric.Status != MetricStatusMet {
		t.Errorf("metric Status = %q, want %q", foundMetric.Status, MetricStatusMet)
	}
}

func TestQualityHandler_UpdateMetric_NotFound(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	metrics := []QualityMetric{
		{
			ID:      "metric-001",
			Name:    "テスト",
			Target:  100.0,
			Unit:    "%",
			Current: 50.0,
			Status:  MetricStatusInProgress,
		},
	}

	addResult, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(metrics),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 存在しない metric を更新
	err = handler.UpdateMetric(ctx, addResult.ID, "metric-999", 100.0, MetricStatusMet)
	if err == nil {
		t.Error("UpdateMetric() should fail for non-existent metric")
	}
}

func TestQualityHandler_UpdateMetric_QualityNotFound(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.UpdateMetric(ctx, "qual-999", "metric-001", 100.0, MetricStatusMet)
	if err != ErrEntityNotFound {
		t.Errorf("UpdateMetric() error = %v, want ErrEntityNotFound", err)
	}
}

// ===== UpdateGate テスト =====

func TestQualityHandler_UpdateGate(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	gates := []QualityGate{
		{
			Name:     "コードレビュー",
			Criteria: []string{"全コードがレビュー済み"},
			Status:   GateStatusPending,
		},
		{
			Name:     "セキュリティ監査",
			Criteria: []string{"脆弱性スキャン完了"},
			Status:   GateStatusPending,
		},
	}

	addResult, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
		WithQualityGates(gates),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// コードレビューゲートを更新
	err = handler.UpdateGate(ctx, addResult.ID, "コードレビュー", GateStatusPassed)
	if err != nil {
		t.Fatalf("UpdateGate() failed: %v", err)
	}

	// 更新確認
	got, err := handler.Get(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)
	var foundGate *QualityGate
	for i := range qual.Gates {
		if qual.Gates[i].Name == "コードレビュー" {
			foundGate = &qual.Gates[i]
			break
		}
	}

	if foundGate == nil {
		t.Fatal("コードレビュー gate not found")
	}

	if foundGate.Status != GateStatusPassed {
		t.Errorf("gate Status = %q, want %q", foundGate.Status, GateStatusPassed)
	}
}

func TestQualityHandler_UpdateGate_NotFound(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	gates := []QualityGate{
		{
			Name:     "テストゲート",
			Criteria: []string{"条件1"},
			Status:   GateStatusPending,
		},
	}

	addResult, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
		WithQualityGates(gates),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 存在しない gate を更新
	err = handler.UpdateGate(ctx, addResult.ID, "存在しないゲート", GateStatusPassed)
	if err == nil {
		t.Error("UpdateGate() should fail for non-existent gate")
	}
}

func TestQualityHandler_UpdateGate_QualityNotFound(t *testing.T) {
	handler, _, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.UpdateGate(ctx, "qual-999", "ゲート名", GateStatusPassed)
	if err != ErrEntityNotFound {
		t.Errorf("UpdateGate() error = %v, want ErrEntityNotFound", err)
	}
}

// ===== Context キャンセルテスト =====

func TestQualityHandler_ContextCancellation(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	objectiveID := createTestObjective(t, objHandler)

	// 各操作がコンテキストキャンセルを正しく処理するか確認
	t.Run("Add", func(t *testing.T) {
		_, err := handler.Add(ctx, "テスト",
			WithQualityObjective(objectiveID),
		)
		if err != context.Canceled {
			t.Errorf("Add() error = %v, want context.Canceled", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := handler.List(ctx, nil)
		if err != context.Canceled {
			t.Errorf("List() error = %v, want context.Canceled", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		_, err := handler.Get(ctx, "qual-001")
		if err != context.Canceled {
			t.Errorf("Get() error = %v, want context.Canceled", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		err := handler.Update(ctx, "qual-001", &QualityEntity{
			Title:       "更新",
			ObjectiveID: objectiveID,
		})
		if err != context.Canceled {
			t.Errorf("Update() error = %v, want context.Canceled", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := handler.Delete(ctx, "qual-001")
		if err != context.Canceled {
			t.Errorf("Delete() error = %v, want context.Canceled", err)
		}
	})
}

// ===== ID シーケンステスト =====

func TestQualityHandler_IDSequence(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	seen := make(map[string]bool)
	for i := 0; i < 3; i++ {
		result, err := handler.Add(ctx, "品質基準",
			WithQualityObjective(objectiveID),
			WithQualityMetrics(defaultMetrics()),
		)
		if err != nil {
			t.Fatalf("Add() %d failed: %v", i+1, err)
		}

		if seen[result.ID] {
			t.Errorf("duplicate ID found: %q", result.ID)
		}
		seen[result.ID] = true

		// プレフィックスが正しいことを確認
		if !strings.HasPrefix(result.ID, "qual-") {
			t.Errorf("Add() %d ID = %q, expected prefix 'qual-'", i+1, result.ID)
		}
	}

	// ID 数が正しいことを確認
	if len(seen) != 3 {
		t.Errorf("expected 3 unique IDs, got %d", len(seen))
	}
}

// ===== ファイルパステスト =====

func TestQualityHandler_FilePath(t *testing.T) {
	handler, objHandler, zeusPath, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	result, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	expectedPath := filepath.Join(zeusPath, "quality", result.ID+".yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file not created: %s", expectedPath)
	}
}

// ===== Metadata テスト =====

func TestQualityHandler_Metadata(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	result, err := handler.Add(ctx, "品質基準",
		WithQualityObjective(objectiveID),
		WithQualityMetrics(defaultMetrics()),
	)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	got, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	qual := got.(*QualityEntity)

	// Metadata が正しく設定されているか確認
	if qual.Metadata.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if qual.Metadata.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

// ===== Metrics/Gates バリデーションテスト =====

func TestQualityHandler_MetricStatus(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	testCases := []struct {
		name   string
		status MetricStatus
	}{
		{"met", MetricStatusMet},
		{"not_met", MetricStatusNotMet},
		{"in_progress", MetricStatusInProgress},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := []QualityMetric{
				{
					ID:      "metric-001",
					Name:    "テスト",
					Target:  100.0,
					Unit:    "%",
					Current: 50.0,
					Status:  tc.status,
				},
			}

			result, err := handler.Add(ctx, "品質基準",
				WithQualityObjective(objectiveID),
				WithQualityMetrics(metrics),
			)
			if err != nil {
				t.Fatalf("Add() failed: %v", err)
			}

			got, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}

			qual := got.(*QualityEntity)
			if qual.Metrics[0].Status != tc.status {
				t.Errorf("Metric status = %q, want %q", qual.Metrics[0].Status, tc.status)
			}
		})
	}
}

func TestQualityHandler_GateStatus(t *testing.T) {
	handler, objHandler, _, cleanup := setupQualityHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	objectiveID := createTestObjective(t, objHandler)

	testCases := []struct {
		name   string
		status GateStatus
	}{
		{"passed", GateStatusPassed},
		{"failed", GateStatusFailed},
		{"pending", GateStatusPending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gates := []QualityGate{
				{
					Name:     "テストゲート",
					Criteria: []string{"条件1"},
					Status:   tc.status,
				},
			}

			result, err := handler.Add(ctx, "品質基準",
				WithQualityObjective(objectiveID),
				WithQualityMetrics(defaultMetrics()),
				WithQualityGates(gates),
			)
			if err != nil {
				t.Fatalf("Add() failed: %v", err)
			}

			got, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}

			qual := got.(*QualityEntity)
			if qual.Gates[0].Status != tc.status {
				t.Errorf("Gate status = %q, want %q", qual.Gates[0].Status, tc.status)
			}
		})
	}
}
