package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// setupRiskHandlerTest は RiskHandler テストのセットアップを行う
func setupRiskHandlerTest(t *testing.T) (*RiskHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-risk-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/risks", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewRiskHandler(fs, nil, nil, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

// ===== Type() テスト =====

func TestRiskHandler_Type(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	if got := handler.Type(); got != "risk" {
		t.Errorf("Type() = %q, want %q", got, "risk")
	}
}

// ===== Add() テスト =====

func TestRiskHandler_Add(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "技術的負債の蓄積")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if !result.Success {
		t.Error("Add() result.Success = false, want true")
	}
	if result.ID != "risk-001" {
		t.Errorf("Add() result.ID = %q, want %q", result.ID, "risk-001")
	}
	if result.Entity != "risk" {
		t.Errorf("Add() result.Entity = %q, want %q", result.Entity, "risk")
	}

	// 作成されたエンティティを取得して確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Add error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.Title != "技術的負債の蓄積" {
		t.Errorf("Title = %q, want %q", risk.Title, "技術的負債の蓄積")
	}
	if risk.Status != RiskStatusIdentified {
		t.Errorf("Status = %q, want %q", risk.Status, RiskStatusIdentified)
	}
	if risk.Probability != RiskProbabilityMedium {
		t.Errorf("Probability = %q, want %q", risk.Probability, RiskProbabilityMedium)
	}
	if risk.Impact != RiskImpactMedium {
		t.Errorf("Impact = %q, want %q", risk.Impact, RiskImpactMedium)
	}
}

func TestRiskHandler_Add_WithAllOptions(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "サーバーダウンリスク",
		WithRiskProbability(RiskProbabilityHigh),
		WithRiskImpact(RiskImpactCritical),
		WithRiskStatus(RiskStatusMitigating),
		WithRiskDescription("本番サーバーがダウンする可能性"),
		WithRiskTrigger("アクセス急増時"),
		WithRiskMitigation(RiskMitigation{
			Preventive: []string{"スケーリング設定の強化", "監視アラートの設定"},
			Contingent: []string{"手動スケーリング", "緊急対応チームへの連絡"},
		}),
		WithRiskOwner("佐藤"),
		WithRiskReviewDate("2025-02-15"),
	)
	if err != nil {
		t.Fatalf("Add() with options error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.Probability != RiskProbabilityHigh {
		t.Errorf("Probability = %q, want %q", risk.Probability, RiskProbabilityHigh)
	}
	if risk.Impact != RiskImpactCritical {
		t.Errorf("Impact = %q, want %q", risk.Impact, RiskImpactCritical)
	}
	if risk.Status != RiskStatusMitigating {
		t.Errorf("Status = %q, want %q", risk.Status, RiskStatusMitigating)
	}
	if risk.Description != "本番サーバーがダウンする可能性" {
		t.Errorf("Description = %q, want correct value", risk.Description)
	}
	if risk.Trigger != "アクセス急増時" {
		t.Errorf("Trigger = %q, want %q", risk.Trigger, "アクセス急増時")
	}
	if len(risk.Mitigation.Preventive) != 2 {
		t.Errorf("Mitigation.Preventive count = %d, want 2", len(risk.Mitigation.Preventive))
	}
	if len(risk.Mitigation.Contingent) != 2 {
		t.Errorf("Mitigation.Contingent count = %d, want 2", len(risk.Mitigation.Contingent))
	}
	if risk.Owner != "佐藤" {
		t.Errorf("Owner = %q, want %q", risk.Owner, "佐藤")
	}
	if risk.ReviewDate != "2025-02-15" {
		t.Errorf("ReviewDate = %q, want %q", risk.ReviewDate, "2025-02-15")
	}
}

func TestRiskHandler_Add_InvalidInput(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の名前
	_, err := handler.Add(ctx, "")
	if err == nil {
		t.Error("Add() with empty name should return error")
	}
}

// ===== RiskScore 自動計算テスト =====

func TestRiskHandler_RiskScoreCalculation(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	testCases := []struct {
		name        string
		probability RiskProbability
		impact      RiskImpact
		expected    RiskScore
	}{
		{"High × Critical = Critical", RiskProbabilityHigh, RiskImpactCritical, RiskScoreCritical},
		{"High × High = Critical", RiskProbabilityHigh, RiskImpactHigh, RiskScoreCritical},
		{"High × Medium = High", RiskProbabilityHigh, RiskImpactMedium, RiskScoreHigh},
		{"High × Low = Medium", RiskProbabilityHigh, RiskImpactLow, RiskScoreMedium},
		{"Medium × Critical = Critical", RiskProbabilityMedium, RiskImpactCritical, RiskScoreCritical},
		{"Medium × High = High", RiskProbabilityMedium, RiskImpactHigh, RiskScoreHigh},
		{"Medium × Medium = Medium", RiskProbabilityMedium, RiskImpactMedium, RiskScoreMedium},
		{"Medium × Low = Low", RiskProbabilityMedium, RiskImpactLow, RiskScoreLow},
		{"Low × Critical = High", RiskProbabilityLow, RiskImpactCritical, RiskScoreHigh},
		{"Low × High = Medium", RiskProbabilityLow, RiskImpactHigh, RiskScoreMedium},
		{"Low × Medium = Low", RiskProbabilityLow, RiskImpactMedium, RiskScoreLow},
		{"Low × Low = Low", RiskProbabilityLow, RiskImpactLow, RiskScoreLow},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.Add(ctx, "テストリスク "+tc.name,
				WithRiskProbability(tc.probability),
				WithRiskImpact(tc.impact),
			)
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			risk := entity.(*RiskEntity)
			if risk.RiskScore != tc.expected {
				t.Errorf("RiskScore = %q, want %q (probability=%s, impact=%s)",
					risk.RiskScore, tc.expected, tc.probability, tc.impact)
			}
		})
	}
}

// ===== List() テスト =====

func TestRiskHandler_List(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数の Risk を追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "リスク"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() Total = %d, want 3", result.Total)
	}
}

func TestRiskHandler_List_Empty(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 0 {
		t.Errorf("List() Total = %d, want 0", result.Total)
	}
}

func TestRiskHandler_List_WithLimit(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5つの Risk を追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "リスク"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	result, err := handler.List(ctx, &ListFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List() with limit error = %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() with limit Total = %d, want 3", result.Total)
	}
}

func TestRiskHandler_List_WithStatusFilter(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるステータスの Risk を追加
	_, err := handler.Add(ctx, "特定済みリスク", WithRiskStatus(RiskStatusIdentified))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "軽減済みリスク", WithRiskStatus(RiskStatusMitigated))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "もう一つの特定済みリスク", WithRiskStatus(RiskStatusIdentified))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	result, err := handler.List(ctx, &ListFilter{Status: string(RiskStatusIdentified)})
	if err != nil {
		t.Fatalf("List() with status filter error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("List() with status filter Total = %d, want 2", result.Total)
	}
}

// ===== Get() テスト =====

func TestRiskHandler_Get(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テストリスク")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.ID != result.ID {
		t.Errorf("Get() ID = %q, want %q", risk.ID, result.ID)
	}
	if risk.Title != "テストリスク" {
		t.Errorf("Get() Title = %q, want %q", risk.Title, "テストリスク")
	}
}

func TestRiskHandler_Get_NotFound(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "risk-999")
	if err != ErrEntityNotFound {
		t.Errorf("Get() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestRiskHandler_Get_InvalidID(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("Get() with invalid ID should return error")
	}
}

// ===== Update() テスト =====

func TestRiskHandler_Update(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "更新前のリスク")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 更新データを作成
	updated := &RiskEntity{
		ID:          result.ID,
		Title:       "更新後のリスク",
		Status:      RiskStatusMitigated,
		Probability: RiskProbabilityLow,
		Impact:      RiskImpactLow,
	}

	err = handler.Update(ctx, result.ID, updated)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 更新結果を確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Update error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.Title != "更新後のリスク" {
		t.Errorf("Title after Update = %q, want %q", risk.Title, "更新後のリスク")
	}
	if risk.Status != RiskStatusMitigated {
		t.Errorf("Status after Update = %q, want %q", risk.Status, RiskStatusMitigated)
	}
	// RiskScore は再計算される
	if risk.RiskScore != RiskScoreLow {
		t.Errorf("RiskScore after Update = %q, want %q", risk.RiskScore, RiskScoreLow)
	}
}

func TestRiskHandler_Update_NotFound(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	updated := &RiskEntity{
		ID:    "risk-999",
		Title: "存在しないリスク",
	}

	err := handler.Update(ctx, "risk-999", updated)
	if err != ErrEntityNotFound {
		t.Errorf("Update() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestRiskHandler_Update_InvalidType(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "リスク")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 間違った型で更新
	err = handler.Update(ctx, result.ID, "wrong type")
	if err == nil {
		t.Error("Update() with wrong type should return error")
	}
}

func TestRiskHandler_Update_RiskScoreRecalculation(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Low × Low = Low で作成
	result, err := handler.Add(ctx, "リスク",
		WithRiskProbability(RiskProbabilityLow),
		WithRiskImpact(RiskImpactLow),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.RiskScore != RiskScoreLow {
		t.Errorf("Initial RiskScore = %q, want %q", risk.RiskScore, RiskScoreLow)
	}

	// High × Critical に更新 → RiskScore = Critical に変わるべき
	updated := &RiskEntity{
		ID:          result.ID,
		Title:       risk.Title,
		Status:      risk.Status,
		Probability: RiskProbabilityHigh,
		Impact:      RiskImpactCritical,
	}

	err = handler.Update(ctx, result.ID, updated)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	entity, err = handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Update error = %v", err)
	}

	risk = entity.(*RiskEntity)
	if risk.RiskScore != RiskScoreCritical {
		t.Errorf("RiskScore after Update = %q, want %q", risk.RiskScore, RiskScoreCritical)
	}
}

// ===== Delete() テスト =====

func TestRiskHandler_Delete(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "削除予定のリスク")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = handler.Delete(ctx, result.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 削除後に取得を試みる
	_, err = handler.Get(ctx, result.ID)
	if err != ErrEntityNotFound {
		t.Errorf("Get() after Delete error = %v, want ErrEntityNotFound", err)
	}
}

func TestRiskHandler_Delete_NotFound(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "risk-999")
	if err != ErrEntityNotFound {
		t.Errorf("Delete() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

// ===== GetRisksByScore() テスト =====

func TestRiskHandler_GetRisksByScore(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるスコアの Risk を追加
	_, err := handler.Add(ctx, "クリティカル 1",
		WithRiskProbability(RiskProbabilityHigh),
		WithRiskImpact(RiskImpactCritical),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "低リスク",
		WithRiskProbability(RiskProbabilityLow),
		WithRiskImpact(RiskImpactLow),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "クリティカル 2",
		WithRiskProbability(RiskProbabilityMedium),
		WithRiskImpact(RiskImpactCritical),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Critical スコアを取得
	critical, err := handler.GetRisksByScore(ctx, RiskScoreCritical)
	if err != nil {
		t.Fatalf("GetRisksByScore() error = %v", err)
	}

	if len(critical) != 2 {
		t.Errorf("GetRisksByScore(Critical) count = %d, want 2", len(critical))
	}

	// Low スコアを取得
	low, err := handler.GetRisksByScore(ctx, RiskScoreLow)
	if err != nil {
		t.Fatalf("GetRisksByScore() error = %v", err)
	}

	if len(low) != 1 {
		t.Errorf("GetRisksByScore(Low) count = %d, want 1", len(low))
	}
}

// ===== GetHighRisks() テスト =====

func TestRiskHandler_GetHighRisks(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Critical リスクを追加
	_, err := handler.Add(ctx, "クリティカルリスク",
		WithRiskProbability(RiskProbabilityHigh),
		WithRiskImpact(RiskImpactCritical),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// High リスクを追加
	_, err = handler.Add(ctx, "高リスク",
		WithRiskProbability(RiskProbabilityHigh),
		WithRiskImpact(RiskImpactMedium),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Low リスクを追加
	_, err = handler.Add(ctx, "低リスク",
		WithRiskProbability(RiskProbabilityLow),
		WithRiskImpact(RiskImpactLow),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Critical と High のみ取得
	highRisks, err := handler.GetHighRisks(ctx)
	if err != nil {
		t.Fatalf("GetHighRisks() error = %v", err)
	}

	if len(highRisks) != 2 {
		t.Errorf("GetHighRisks() count = %d, want 2", len(highRisks))
	}

	// 全て Critical または High であることを確認
	for _, risk := range highRisks {
		if risk.RiskScore != RiskScoreCritical && risk.RiskScore != RiskScoreHigh {
			t.Errorf("GetHighRisks() returned risk with score %q, want critical or high", risk.RiskScore)
		}
	}
}

func TestRiskHandler_GetHighRisks_Empty(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Low リスクのみ追加
	_, err := handler.Add(ctx, "低リスク",
		WithRiskProbability(RiskProbabilityLow),
		WithRiskImpact(RiskImpactLow),
	)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	highRisks, err := handler.GetHighRisks(ctx)
	if err != nil {
		t.Fatalf("GetHighRisks() error = %v", err)
	}

	if len(highRisks) != 0 {
		t.Errorf("GetHighRisks() count = %d, want 0", len(highRisks))
	}
}

// ===== Context キャンセルテスト =====

func TestRiskHandler_ContextCancellation(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	// キャンセル済みコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add
	_, err := handler.Add(ctx, "テスト")
	if err == nil {
		t.Error("Add() with cancelled context should return error")
	}

	// List
	_, err = handler.List(ctx, nil)
	if err == nil {
		t.Error("List() with cancelled context should return error")
	}

	// Get
	_, err = handler.Get(ctx, "risk-001")
	if err == nil {
		t.Error("Get() with cancelled context should return error")
	}

	// Update
	err = handler.Update(ctx, "risk-001", &RiskEntity{})
	if err == nil {
		t.Error("Update() with cancelled context should return error")
	}

	// Delete
	err = handler.Delete(ctx, "risk-001")
	if err == nil {
		t.Error("Delete() with cancelled context should return error")
	}
}

// ===== ID 採番テスト =====

func TestRiskHandler_IDSequence(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 連続追加で ID が順番に採番されることを確認
	expectedIDs := []string{"risk-001", "risk-002", "risk-003"}

	for i, expected := range expectedIDs {
		result, err := handler.Add(ctx, "リスク"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() #%d error = %v", i+1, err)
		}
		if result.ID != expected {
			t.Errorf("Add() #%d ID = %q, want %q", i+1, result.ID, expected)
		}
	}
}

// ===== ファイルパステスト =====

func TestRiskHandler_FilePath(t *testing.T) {
	handler, zeusPath, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テストリスク")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// ファイルが正しいパスに作成されたか確認
	expectedPath := filepath.Join(zeusPath, "risks", result.ID+".yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %q does not exist", expectedPath)
	}
}

// ===== メタデータテスト =====

func TestRiskHandler_Metadata(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "メタデータテスト")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	risk := entity.(*RiskEntity)
	if risk.Metadata.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if risk.Metadata.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

// ===== 全ステータステスト =====

func TestRiskHandler_AllStatusLevels(t *testing.T) {
	handler, _, cleanup := setupRiskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	statuses := []RiskStatus{
		RiskStatusIdentified,
		RiskStatusMitigating,
		RiskStatusMitigated,
		RiskStatusOccurred,
		RiskStatusClosed,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			result, err := handler.Add(ctx, "リスク "+string(status), WithRiskStatus(status))
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			risk := entity.(*RiskEntity)
			if risk.Status != status {
				t.Errorf("Status = %q, want %q", risk.Status, status)
			}
		})
	}
}
