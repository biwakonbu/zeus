package yaml

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// ロック関連エラー
var (
	// ErrLockTimeout はロックタイムアウトエラー
	ErrLockTimeout = errors.New("lock acquisition timed out")
)

// FileLock はファイルロックを管理
type FileLock struct {
	path string
	file *os.File
	mu   sync.Mutex
}

// NewFileLock は新しい FileLock を作成
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path + ".lock",
	}
}

// Lock はロックを取得（ブロッキング）
func (fl *FileLock) Lock() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// ロックファイルのディレクトリを作成
	dir := filepath.Dir(fl.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	// flock で排他ロックを取得
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	fl.file = file
	return nil
}

// LockWithTimeout はタイムアウト付きでロックを取得
func (fl *FileLock) LockWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- fl.Lock()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return ErrLockTimeout
	}
}

// Unlock はロックを解放
func (fl *FileLock) Unlock() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if fl.file == nil {
		return nil
	}

	// flock を解放
	if err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN); err != nil {
		return err
	}

	err := fl.file.Close()
	fl.file = nil

	// ロックファイルを削除（ベストエフォート）
	os.Remove(fl.path)

	return err
}

// TryLock は非ブロッキングでロック取得を試みる
func (fl *FileLock) TryLock() (bool, error) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// ロックファイルのディレクトリを作成
	dir := filepath.Dir(fl.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}

	file, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false, err
	}

	// 非ブロッキングで排他ロックを試みる
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if err == syscall.EWOULDBLOCK {
			return false, nil
		}
		return false, err
	}

	fl.file = file
	return true, nil
}
