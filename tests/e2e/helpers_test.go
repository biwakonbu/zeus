package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
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
