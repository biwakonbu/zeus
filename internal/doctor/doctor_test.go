package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/biwakonbu/zeus/internal/yaml"
)

func TestNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	d := New(tmpDir)
	if d == nil {
		t.Error("New should return non-nil")
	}
	if d.zeusPath != filepath.Join(tmpDir, ".zeus") {
		t.Errorf("expected zeusPath %q, got %q", filepath.Join(tmpDir, ".zeus"), d.zeusPath)
	}
}

func TestDiagnose_Healthy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Zeus を初期化して健全な状態を作成
	z := core.New(tmpDir)
	ctx := context.Background()
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Doctor で診断
	d := New(tmpDir)
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Errorf("Diagnose() error = %v", err)
	}

	if result.Overall != "healthy" {
		t.Errorf("expected Overall 'healthy', got %q", result.Overall)
	}

	// 全てのチェックが pass
	for _, check := range result.Checks {
		if check.Status != "pass" {
			t.Errorf("check %q should be 'pass', got %q", check.Check, check.Status)
		}
	}
}

func TestDiagnose_Unhealthy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// .zeus ディレクトリのみ作成（設定ファイルなし）
	zeusDir := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusDir, 0755); err != nil {
		t.Fatalf("failed to create .zeus dir: %v", err)
	}

	d := New(tmpDir)
	ctx := context.Background()
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Errorf("Diagnose() error = %v", err)
	}

	if result.Overall != "unhealthy" {
		t.Errorf("expected Overall 'unhealthy', got %q", result.Overall)
	}
}

func TestDiagnose_Degraded(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Zeus を初期化
	z := core.New(tmpDir)
	ctx := context.Background()
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// タスクファイルを削除して degraded 状態を作成
	os.Remove(filepath.Join(tmpDir, ".zeus", "tasks", "active.yaml"))

	d := New(tmpDir)
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Errorf("Diagnose() error = %v", err)
	}

	if result.Overall != "degraded" {
		t.Errorf("expected Overall 'degraded', got %q", result.Overall)
	}
}

func TestFix_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// .zeus ディレクトリのみ作成
	zeusDir := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusDir, 0755); err != nil {
		t.Fatalf("failed to create .zeus dir: %v", err)
	}

	d := New(tmpDir)
	ctx := context.Background()

	// DryRun モードで修復
	result, err := d.Fix(ctx, true)
	if err != nil {
		t.Errorf("Fix() error = %v", err)
	}

	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}

	// DryRun なので実際には修復されていない
	if _, err := os.Stat(filepath.Join(zeusDir, "zeus.yaml")); !os.IsNotExist(err) {
		t.Error("DryRun should not create files")
	}
}

func TestFix_Execute(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// .zeus ディレクトリのみ作成
	zeusDir := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusDir, 0755); err != nil {
		t.Fatalf("failed to create .zeus dir: %v", err)
	}

	d := New(tmpDir)
	ctx := context.Background()

	// 実際に修復を実行
	result, err := d.Fix(ctx, false)
	if err != nil {
		t.Errorf("Fix() error = %v", err)
	}

	if result.DryRun {
		t.Error("expected DryRun to be false")
	}

	// 修復アクションがあったはず
	if len(result.Fixes) == 0 {
		t.Error("expected at least one fix action")
	}

	// 修復後は健全な状態
	diagnosis, _ := d.Diagnose(ctx)
	if diagnosis.Overall != "healthy" {
		t.Errorf("expected Overall 'healthy' after fix, got %q", diagnosis.Overall)
	}
}

func TestCalculateOverall(t *testing.T) {
	d := &Doctor{}

	tests := []struct {
		name     string
		checks   []CheckResult
		expected string
	}{
		{
			name:     "all pass",
			checks:   []CheckResult{{Status: "pass"}, {Status: "pass"}},
			expected: "healthy",
		},
		{
			name:     "one warn",
			checks:   []CheckResult{{Status: "pass"}, {Status: "warn"}},
			expected: "degraded",
		},
		{
			name:     "one fail",
			checks:   []CheckResult{{Status: "pass"}, {Status: "fail"}},
			expected: "unhealthy",
		},
		{
			name:     "fail takes priority over warn",
			checks:   []CheckResult{{Status: "warn"}, {Status: "fail"}},
			expected: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.calculateOverall(tt.checks)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCheckConfigExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// ファイルなし
	d := New(tmpDir)
	check := d.checkConfigExists(ctx)
	if check.Status != "fail" {
		t.Errorf("expected Status 'fail', got %q", check.Status)
	}
	if !check.Fixable {
		t.Error("should be fixable")
	}

	// Zeus を初期化
	z := core.New(tmpDir)
	_, _ = z.Init(ctx)

	// ファイルあり
	check = d.checkConfigExists(ctx)
	if check.Status != "pass" {
		t.Errorf("expected Status 'pass', got %q", check.Status)
	}
}

func TestCheckTasksExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Zeus を初期化
	z := core.New(tmpDir)
	_, _ = z.Init(ctx)

	d := New(tmpDir)

	// ファイルあり
	check := d.checkTasksExists(ctx)
	if check.Status != "pass" {
		t.Errorf("expected Status 'pass', got %q", check.Status)
	}

	// ファイルを削除
	os.Remove(filepath.Join(tmpDir, ".zeus", "tasks", "active.yaml"))

	check = d.checkTasksExists(ctx)
	if check.Status != "warn" {
		t.Errorf("expected Status 'warn', got %q", check.Status)
	}
	if !check.Fixable {
		t.Error("should be fixable")
	}
}

func TestCheckStateExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Zeus を初期化
	z := core.New(tmpDir)
	_, _ = z.Init(ctx)

	d := New(tmpDir)

	// ファイルあり
	check := d.checkStateExists(ctx)
	if check.Status != "pass" {
		t.Errorf("expected Status 'pass', got %q", check.Status)
	}

	// ファイルを削除
	os.Remove(filepath.Join(tmpDir, ".zeus", "state", "current.yaml"))

	check = d.checkStateExists(ctx)
	if check.Status != "warn" {
		t.Errorf("expected Status 'warn', got %q", check.Status)
	}
}

func TestDiagnose_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	d := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = d.Diagnose(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestFix_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	d := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = d.Fix(ctx, false)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestFixableCount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// .zeus ディレクトリのみ作成
	zeusDir := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusDir, 0755); err != nil {
		t.Fatalf("failed to create .zeus dir: %v", err)
	}

	d := New(tmpDir)
	ctx := context.Background()
	result, _ := d.Diagnose(ctx)

	// 設定ファイルがないので fixable があるはず
	if result.FixableCount == 0 {
		t.Error("expected FixableCount > 0")
	}
}

// TASK-016: サブシステム参照チェックのテスト

// setupIntegrityCheckerTest は IntegrityChecker 付き Doctor をセットアップ
func setupIntegrityCheckerTest(t *testing.T) (string, *Doctor, *core.IntegrityChecker, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "doctor-integrity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	zeusPath := filepath.Join(tmpDir, ".zeus")
	if err := os.MkdirAll(zeusPath, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	// 必要なディレクトリを作成（usecases のみ、subsystems はファイルベース）
	if err := os.MkdirAll(filepath.Join(zeusPath, "usecases"), 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create usecases dir: %v", err)
	}

	// FileStore と各ハンドラーを作成
	fs := yaml.NewFileManager(zeusPath)
	usecaseHandler := core.NewUseCaseHandler(fs, nil, nil, nil)
	subsystemHandler := core.NewSubsystemHandler(fs)

	// IntegrityChecker を作成
	checker := core.NewIntegrityChecker(nil, nil)
	checker.SetUseCaseHandler(usecaseHandler)
	checker.SetSubsystemHandler(subsystemHandler)

	// Doctor を作成
	d := NewWithIntegrity(tmpDir, checker)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return zeusPath, d, checker, cleanup
}

func TestCheckIntegrity_SubsystemReference_Valid(t *testing.T) {
	zeusPath, d, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 有効なサブシステムを作成（SubsystemsFile 形式で保存）
	subsystemsFile := &core.SubsystemsFile{
		Subsystems: []core.SubsystemEntity{
			{
				ID:   "sub-12345678",
				Name: "Test Subsystem",
			},
		},
	}
	if err := fs.WriteYaml(ctx, "subsystems.yaml", subsystemsFile); err != nil {
		t.Fatalf("failed to write subsystems.yaml: %v", err)
	}

	// サブシステムを参照するユースケースを作成
	usecase := &core.UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "Test UseCase",
		SubsystemID: "sub-12345678", // 有効な参照
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", usecase); err != nil {
		t.Fatalf("failed to write usecase: %v", err)
	}

	// Zeus 初期化（設定ファイルを作成）
	z := core.New(filepath.Dir(zeusPath))
	_, _ = z.Init(ctx)

	// 診断を実行
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}

	// サブシステム参照チェックが pass であることを確認
	foundSubsystemCheck := false
	for _, check := range result.Checks {
		if check.Check == "subsystem_reference" {
			foundSubsystemCheck = true
			if check.Status != "pass" {
				t.Errorf("expected subsystem_reference check to pass, got %q: %s", check.Status, check.Message)
			}
		}
	}

	if !foundSubsystemCheck {
		t.Error("subsystem_reference check not found")
	}
}

func TestCheckIntegrity_SubsystemReference_Invalid(t *testing.T) {
	zeusPath, d, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// サブシステムを参照するユースケースを作成（サブシステムは存在しない）
	usecase := &core.UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "Test UseCase",
		SubsystemID: "sub-nonexist", // 無効な参照（存在しないサブシステム）
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", usecase); err != nil {
		t.Fatalf("failed to write usecase: %v", err)
	}

	// Zeus 初期化
	z := core.New(filepath.Dir(zeusPath))
	_, _ = z.Init(ctx)

	// 診断を実行
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}

	// サブシステム参照チェックが warn であることを確認
	foundSubsystemCheck := false
	for _, check := range result.Checks {
		if check.Check == "subsystem_reference" {
			foundSubsystemCheck = true
			if check.Status != "warn" {
				t.Errorf("expected subsystem_reference check to warn, got %q: %s", check.Status, check.Message)
			}
			// メッセージに適切な情報が含まれていることを確認
			if check.Message == "" {
				t.Error("expected warning message to be non-empty")
			}
		}
	}

	if !foundSubsystemCheck {
		t.Error("subsystem_reference check not found")
	}

	// 警告があっても Overall は degraded になることを確認
	// （他のチェックが pass の場合）
	// Note: 実際には zeus.yaml なども必要なので unhealthy or degraded
}

func TestCheckIntegrity_SubsystemReference_NoSubsystem(t *testing.T) {
	zeusPath, d, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// サブシステムを参照しないユースケースを作成
	usecase := &core.UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "Test UseCase",
		SubsystemID: "", // サブシステム未設定（任意フィールド）
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", usecase); err != nil {
		t.Fatalf("failed to write usecase: %v", err)
	}

	// Zeus 初期化
	z := core.New(filepath.Dir(zeusPath))
	_, _ = z.Init(ctx)

	// 診断を実行
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}

	// サブシステム参照チェックが pass であることを確認（未設定は OK）
	foundSubsystemCheck := false
	for _, check := range result.Checks {
		if check.Check == "subsystem_reference" {
			foundSubsystemCheck = true
			if check.Status != "pass" {
				t.Errorf("expected subsystem_reference check to pass when no subsystem_id, got %q: %s", check.Status, check.Message)
			}
		}
	}

	if !foundSubsystemCheck {
		t.Error("subsystem_reference check not found")
	}
}

func TestCheckIntegrity_MultipleUseCases_MixedReferences(t *testing.T) {
	zeusPath, d, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 有効なサブシステムを作成（SubsystemsFile 形式で保存）
	subsystemsFile := &core.SubsystemsFile{
		Subsystems: []core.SubsystemEntity{
			{
				ID:   "sub-12345678",
				Name: "Test Subsystem",
			},
		},
	}
	if err := fs.WriteYaml(ctx, "subsystems.yaml", subsystemsFile); err != nil {
		t.Fatalf("failed to write subsystems.yaml: %v", err)
	}

	// 有効な参照を持つユースケース
	usecase1 := &core.UseCaseEntity{
		ID:          "uc-11111111",
		Title:       "UseCase 1",
		SubsystemID: "sub-12345678", // 有効
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-11111111.yaml", usecase1); err != nil {
		t.Fatalf("failed to write usecase1: %v", err)
	}

	// 無効な参照を持つユースケース
	usecase2 := &core.UseCaseEntity{
		ID:          "uc-22222222",
		Title:       "UseCase 2",
		SubsystemID: "sub-nonexist", // 無効
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-22222222.yaml", usecase2); err != nil {
		t.Fatalf("failed to write usecase2: %v", err)
	}

	// 参照なしのユースケース
	usecase3 := &core.UseCaseEntity{
		ID:          "uc-33333333",
		Title:       "UseCase 3",
		SubsystemID: "", // 未設定
		Status:      core.UseCaseStatusDraft,
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-33333333.yaml", usecase3); err != nil {
		t.Fatalf("failed to write usecase3: %v", err)
	}

	// Zeus 初期化
	z := core.New(filepath.Dir(zeusPath))
	_, _ = z.Init(ctx)

	// 診断を実行
	result, err := d.Diagnose(ctx)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}

	// サブシステム参照チェックの結果を確認
	warnCount := 0
	for _, check := range result.Checks {
		if check.Check == "subsystem_reference" && check.Status == "warn" {
			warnCount++
		}
	}

	// 1つの無効な参照があるので 1 つの警告
	if warnCount != 1 {
		t.Errorf("expected 1 warning, got %d", warnCount)
	}
}

func TestNewWithIntegrity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	checker := core.NewIntegrityChecker(nil, nil)
	d := NewWithIntegrity(tmpDir, checker)

	if d == nil {
		t.Error("NewWithIntegrity should return non-nil")
	}
	if d.integrityChecker == nil {
		t.Error("integrityChecker should be set")
	}
	if d.zeusPath != filepath.Join(tmpDir, ".zeus") {
		t.Errorf("expected zeusPath %q, got %q", filepath.Join(tmpDir, ".zeus"), d.zeusPath)
	}
}
