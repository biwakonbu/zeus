// Package e2e はZeus CLIのエンドツーエンドテストを提供する。
// 実際のCLIバイナリを実行し、統合的な動作を検証する。
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// binaryPath はテスト用にビルドされたZeusバイナリのパス
var binaryPath string

// TestMain はテスト実行前にバイナリをビルドし、終了後にクリーンアップする
func TestMain(m *testing.M) {
	// プロジェクトルートを取得
	// tests/e2e からの相対パス
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		println("プロジェクトルート取得失敗:", err.Error())
		os.Exit(1)
	}

	// テスト用バイナリをビルド
	tmpBin := filepath.Join(os.TempDir(), "zeus-e2e-test")
	cmd := exec.Command("go", "build", "-o", tmpBin, projectRoot)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		println("バイナリビルド失敗:", err.Error())
		os.Exit(1)
	}
	binaryPath = tmpBin

	// テスト実行
	code := m.Run()

	// クリーンアップ
	os.Remove(tmpBin)
	os.Exit(code)
}
