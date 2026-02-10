package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// CommandResult はコマンド実行結果を格納する
type CommandResult struct {
	Stdout   string // 標準出力
	Stderr   string // 標準エラー出力
	ExitCode int    // 終了コード
}

// runCommand はZeusコマンドを実行して結果を返す
// dir: 作業ディレクトリ
// args: コマンド引数（例: "init"）
func runCommand(t *testing.T, dir string, args ...string) CommandResult {
	t.Helper()

	// 30秒タイムアウト
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getBinaryPath(), args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if ctx.Err() == context.DeadlineExceeded {
		t.Logf("コマンドがタイムアウトしました: %v", args)
		exitCode = -1
	} else if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Logf("コマンド実行エラー: %v", err)
		exitCode = -1
	}

	return CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// setupTempDir はテスト用の一時ディレクトリを作成する
func setupTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "zeus-e2e-*")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成失敗: %v", err)
	}
	return dir
}

// cleanupTempDir は一時ディレクトリを削除する
func cleanupTempDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("クリーンアップ警告: %v", err)
	}
}

// assertExitCode は終了コードを検証する
func assertExitCode(t *testing.T, result CommandResult, expected int) {
	t.Helper()
	if result.ExitCode != expected {
		t.Errorf("終了コード不一致: got %d, want %d\nstdout: %s\nstderr: %s",
			result.ExitCode, expected, result.Stdout, result.Stderr)
	}
}

// assertSuccess は成功（終了コード0）を検証する
func assertSuccess(t *testing.T, result CommandResult) {
	t.Helper()
	assertExitCode(t, result, 0)
}

// assertFailure は失敗（終了コード非0）を検証する
func assertFailure(t *testing.T, result CommandResult) {
	t.Helper()
	if result.ExitCode == 0 {
		t.Errorf("コマンドが成功すべきでないのに成功しました\nstdout: %s",
			result.Stdout)
	}
}

// assertOutputContains は標準出力に指定文字列が含まれるか検証する
func assertOutputContains(t *testing.T, result CommandResult, substr string) {
	t.Helper()
	if !strings.Contains(result.Stdout, substr) {
		t.Errorf("出力に %q が含まれていません\nstdout: %s", substr, result.Stdout)
	}
}

// assertOutputNotContains は標準出力に指定文字列が含まれないことを検証する
func assertOutputNotContains(t *testing.T, result CommandResult, substr string) {
	t.Helper()
	if strings.Contains(result.Stdout, substr) {
		t.Errorf("出力に %q が含まれるべきではありません\nstdout: %s", substr, result.Stdout)
	}
}

// assertStderrContains は標準エラー出力に指定文字列が含まれるか検証する
func assertStderrContains(t *testing.T, result CommandResult, substr string) {
	t.Helper()
	if !strings.Contains(result.Stderr, substr) {
		t.Errorf("エラー出力に %q が含まれていません\nstderr: %s", substr, result.Stderr)
	}
}

// assertFileExists はファイルの存在を検証する
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("ファイルが存在しません: %s", path)
		return
	}
	if err != nil {
		t.Errorf("ファイル確認エラー: %v", err)
		return
	}
	if info.IsDir() {
		t.Errorf("ディレクトリではなくファイルが期待されます: %s", path)
	}
}

// assertDirExists はディレクトリの存在を検証する
func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("ディレクトリが存在しません: %s", path)
		return
	}
	if err != nil {
		t.Errorf("ディレクトリ確認エラー: %v", err)
		return
	}
	if !info.IsDir() {
		t.Errorf("ファイルではなくディレクトリが期待されます: %s", path)
	}
}

// assertFileNotExists はファイルが存在しないことを検証する
func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("ファイルが存在すべきではありません: %s", path)
	} else if !os.IsNotExist(err) {
		t.Errorf("ファイル確認エラー: %v", err)
	}
}

// =============================================================================
// 10概念モデル用ヘルパー関数
// =============================================================================

// extractEntityID はコマンド結果からエンティティIDを抽出する
// 出力例: "Added objective: Test Objective (ID: obj-001)"
// prefix: "obj-", "con-", "dec-", "prob-", "risk-", "assum-", "qual-", "vision-"
func extractEntityID(t *testing.T, result CommandResult, prefix string) string {
	t.Helper()
	re := regexp.MustCompile(`ID: (` + regexp.QuoteMeta(prefix) + `\w+)`)
	matches := re.FindStringSubmatch(result.Stdout)
	if len(matches) < 2 {
		t.Logf("出力からIDを抽出できませんでした（prefix: %s）: %s", prefix, result.Stdout)
		return ""
	}
	return matches[1]
}

// setupBasicProject は基本プロジェクトをセットアップする（init + Vision + Objective）
// 返り値: map[string]string{"vision": visionID, "objective": objID}
func setupBasicProject(t *testing.T, dir string) map[string]string {
	t.Helper()
	ids := make(map[string]string)

	// init
	result := runCommand(t, dir, "init")
	assertSuccess(t, result)

	// Vision 作成
	result = runCommand(t, dir, "add", "vision", "テストビジョン",
		"--statement", "テスト用のプロジェクトビジョン")
	assertSuccess(t, result)
	ids["vision"] = "vision-001"

	// Objective 作成
	result = runCommand(t, dir, "add", "objective", "Phase 1 目標")
	assertSuccess(t, result)
	objID := extractEntityID(t, result, "obj-")
	if objID == "" {
		objID = "obj-001"
	}
	ids["objective"] = objID

	return ids
}

// setupDecisionFlow は検討・決定フローをセットアップする
// 基本プロジェクト（Vision + Objective）の上に Consideration を作成
// 返り値: map に "consideration" を追加（Decision は後から作成する想定）
func setupDecisionFlow(t *testing.T, dir string) map[string]string {
	t.Helper()

	// 基本プロジェクトをセットアップ
	ids := setupBasicProject(t, dir)

	// Consideration 作成
	result := runCommand(t, dir, "add", "consideration", "技術選定",
		"--objective", ids["objective"],
		"--due", "2026-02-15")
	assertSuccess(t, result)
	conID := extractEntityID(t, result, "con-")
	if conID == "" {
		conID = "con-001"
	}
	ids["consideration"] = conID

	return ids
}

// setupFullProject はすべてのエンティティを含むプロジェクトをセットアップする
// Phase 1-3 の各エンティティを作成し、参照整合性テスト用に使用
func setupFullProject(t *testing.T, dir string) map[string]string {
	t.Helper()

	// 基本 + Consideration フローをセットアップ
	ids := setupDecisionFlow(t, dir)

	// Decision 作成
	result := runCommand(t, dir, "add", "decision", "React採用",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "React",
		"--rationale", "コミュニティの大きさとエコシステム")
	assertSuccess(t, result)
	decID := extractEntityID(t, result, "dec-")
	if decID == "" {
		decID = "dec-001"
	}
	ids["decision"] = decID

	// Problem 作成
	result = runCommand(t, dir, "add", "problem", "パフォーマンス問題",
		"--severity", "high",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	probID := extractEntityID(t, result, "prob-")
	if probID == "" {
		probID = "prob-001"
	}
	ids["problem"] = probID

	// Risk 作成
	result = runCommand(t, dir, "add", "risk", "外部API依存",
		"--probability", "medium",
		"--impact", "high",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	riskID := extractEntityID(t, result, "risk-")
	if riskID == "" {
		riskID = "risk-001"
	}
	ids["risk"] = riskID

	// Assumption 作成
	result = runCommand(t, dir, "add", "assumption", "ユーザー数1000人以下",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	assumID := extractEntityID(t, result, "assum-")
	if assumID == "" {
		assumID = "assum-001"
	}
	ids["assumption"] = assumID

	// Constraint 作成
	result = runCommand(t, dir, "add", "constraint", "外部DB不使用",
		"--category", "technical",
		"--non-negotiable")
	assertSuccess(t, result)
	ids["constraint"] = "constraint" // Constraint はグローバルファイルのため固定

	// Quality 作成
	result = runCommand(t, dir, "add", "quality", "コードカバレッジ",
		"--objective", ids["objective"],
		"--metric", "coverage:80:%")
	assertSuccess(t, result)
	qualID := extractEntityID(t, result, "qual-")
	if qualID == "" {
		qualID = "qual-001"
	}
	ids["quality"] = qualID

	return ids
}
