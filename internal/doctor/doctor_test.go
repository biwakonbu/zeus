package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/core"
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
