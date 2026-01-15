package yaml

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileLock(t *testing.T) {
	fl := NewFileLock("/tmp/test")
	if fl == nil {
		t.Error("NewFileLock should return non-nil")
	}
	if fl.path != "/tmp/test.lock" {
		t.Errorf("expected path '/tmp/test.lock', got %q", fl.path)
	}
}

func TestFileLock_Lock_Unlock(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl := NewFileLock(lockPath)

	// ロック取得
	err = fl.Lock()
	if err != nil {
		t.Errorf("Lock() error = %v", err)
	}

	// ロックファイルが存在するか確認
	if _, err := os.Stat(lockPath + ".lock"); os.IsNotExist(err) {
		t.Error("Lock() should create lock file")
	}

	// ロック解放
	err = fl.Unlock()
	if err != nil {
		t.Errorf("Unlock() error = %v", err)
	}
}

func TestFileLock_Unlock_NotLocked(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl := NewFileLock(lockPath)

	// ロックしていない状態で Unlock
	err = fl.Unlock()
	if err != nil {
		t.Errorf("Unlock() should not error when not locked, got %v", err)
	}
}

func TestFileLock_LockWithTimeout_Success(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl := NewFileLock(lockPath)

	// タイムアウト付きでロック取得
	err = fl.LockWithTimeout(5 * time.Second)
	if err != nil {
		t.Errorf("LockWithTimeout() error = %v", err)
	}

	// クリーンアップ
	fl.Unlock()
}

func TestFileLock_TryLock_Success(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl := NewFileLock(lockPath)

	// TryLock でロック取得
	acquired, err := fl.TryLock()
	if err != nil {
		t.Errorf("TryLock() error = %v", err)
	}
	if !acquired {
		t.Error("TryLock() should acquire lock")
	}

	// クリーンアップ
	fl.Unlock()
}

func TestFileLock_TryLock_AlreadyLocked(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl1 := NewFileLock(lockPath)
	fl2 := NewFileLock(lockPath)

	// 最初のロック取得
	err = fl1.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	defer fl1.Unlock()

	// 2つ目のロックは取得できないはず
	acquired, err := fl2.TryLock()
	if err != nil {
		t.Errorf("TryLock() error = %v", err)
	}
	if acquired {
		t.Error("TryLock() should not acquire lock when already locked")
		fl2.Unlock()
	}
}

func TestFileLock_CreateDir(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 存在しないサブディレクトリ内のロック
	lockPath := filepath.Join(tmpDir, "subdir", "nested", "test")
	fl := NewFileLock(lockPath)

	err = fl.Lock()
	if err != nil {
		t.Errorf("Lock() should create directory, error = %v", err)
	}

	fl.Unlock()
}

func TestFileLock_MultipleLockUnlock(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "filelock-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	fl := NewFileLock(lockPath)

	// 複数回のロック/アンロック
	for i := 0; i < 3; i++ {
		err = fl.Lock()
		if err != nil {
			t.Errorf("Lock() iteration %d error = %v", i, err)
		}
		err = fl.Unlock()
		if err != nil {
			t.Errorf("Unlock() iteration %d error = %v", i, err)
		}
	}
}
