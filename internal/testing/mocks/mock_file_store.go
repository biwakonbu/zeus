package mocks

import (
	"context"
	"sync"

	"gopkg.in/yaml.v3"
)

// MockFileStore は FileStore のモック実装
// テスト時に実際のファイルシステムを使用せずにテスト可能にする
type MockFileStore struct {
	mu       sync.RWMutex
	basePath string
	files    map[string][]byte
	errors   map[string]error
}

// NewMockFileStore は新しい MockFileStore を作成
func NewMockFileStore(basePath string) *MockFileStore {
	return &MockFileStore{
		basePath: basePath,
		files:    make(map[string][]byte),
		errors:   make(map[string]error),
	}
}

// SetError は特定のパスにエラーを設定（テスト用）
func (m *MockFileStore) SetError(path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[path] = err
}

// SetFile はファイル内容を設定（テスト用）
func (m *MockFileStore) SetFile(path string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = content
}

// GetFile はファイル内容を取得（テスト用）
func (m *MockFileStore) GetFile(path string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, ok := m.files[path]
	return content, ok
}

// ClearFiles は全ファイルをクリア（テスト用）
func (m *MockFileStore) ClearFiles() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files = make(map[string][]byte)
}

// Exists はファイルが存在するか確認
func (m *MockFileStore) Exists(ctx context.Context, path string) bool {
	if ctx.Err() != nil {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.files[path]
	return ok
}

// ReadYaml は YAML ファイルを読み込む
func (m *MockFileStore) ReadYaml(ctx context.Context, path string, v any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[path]; ok {
		return err
	}

	content, ok := m.files[path]
	if !ok {
		return &FileNotFoundError{Path: path}
	}

	return yaml.Unmarshal(content, v)
}

// WriteYaml は YAML ファイルを書き込む
func (m *MockFileStore) WriteYaml(ctx context.Context, path string, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[path]; ok {
		return err
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	m.files[path] = content
	return nil
}

// EnsureDir はディレクトリを作成
func (m *MockFileStore) EnsureDir(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[path]; ok {
		return err
	}

	// ディレクトリはモックでは特に何もしない
	return nil
}

// Delete はファイルを削除
func (m *MockFileStore) Delete(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[path]; ok {
		return err
	}

	delete(m.files, path)
	return nil
}

// Glob はパターンに一致するファイルを検索
func (m *MockFileStore) Glob(ctx context.Context, pattern string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[pattern]; ok {
		return nil, err
	}

	// 簡易的なパターンマッチング（テスト用）
	var matches []string
	for path := range m.files {
		if matchPattern(pattern, path) {
			matches = append(matches, path)
		}
	}

	return matches, nil
}

// WriteFile はバイナリファイルを書き込む
func (m *MockFileStore) WriteFile(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[path]; ok {
		return err
	}

	m.files[path] = data
	return nil
}

// Copy はファイルをコピー
func (m *MockFileStore) Copy(ctx context.Context, src, dest string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[src]; ok {
		return err
	}

	content, ok := m.files[src]
	if !ok {
		return &FileNotFoundError{Path: src}
	}

	m.files[dest] = make([]byte, len(content))
	copy(m.files[dest], content)
	return nil
}

// ListDir はディレクトリ内のファイルを列挙
func (m *MockFileStore) ListDir(ctx context.Context, path string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[path]; ok {
		return nil, err
	}

	// パス配下のファイルを検索
	var files []string
	prefix := path + "/"
	for filePath := range m.files {
		if len(filePath) > len(prefix) && filePath[:len(prefix)] == prefix {
			// サブディレクトリを除外
			rest := filePath[len(prefix):]
			hasSlash := false
			for _, c := range rest {
				if c == '/' {
					hasSlash = true
					break
				}
			}
			if !hasSlash {
				files = append(files, rest)
			}
		}
	}

	return files, nil
}

// BasePath はベースパスを返す
func (m *MockFileStore) BasePath() string {
	return m.basePath
}

// FileNotFoundError はファイルが見つからないエラー
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return "file not found: " + e.Path
}

// matchPattern は簡易的なパターンマッチング
func matchPattern(pattern, path string) bool {
	// 簡易実装: * のみサポート
	// 例: "tasks/*.yaml" は "tasks/active.yaml" にマッチ
	if len(pattern) == 0 {
		return len(path) == 0
	}

	// * がある場合
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			// * の後のパターンがパスのどこかにマッチするか確認
			rest := pattern[i+1:]
			if len(rest) == 0 {
				return true
			}
			for j := 0; j <= len(path); j++ {
				if matchPattern(rest, path[j:]) {
					return true
				}
			}
			return false
		}
		if i >= len(path) || pattern[i] != path[i] {
			return false
		}
	}

	return len(pattern) == len(path)
}
