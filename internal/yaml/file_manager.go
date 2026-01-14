package yaml

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// セキュリティ関連エラー
var (
	// ErrPathTraversal はディレクトリトラバーサル攻撃を検出
	ErrPathTraversal = errors.New("path traversal detected: access outside base directory is not allowed")
)

// FileManager はファイル操作を管理
type FileManager struct {
	basePath string
	parser   *Parser
	writer   *Writer
}

// NewFileManager は新しい FileManager を作成
func NewFileManager(basePath string) *FileManager {
	// basePath を正規化（シンボリックリンクを解決）
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		absBasePath = basePath
	}
	// シンボリックリンクを解決（ディレクトリが存在する場合）
	if evalPath, err := filepath.EvalSymlinks(absBasePath); err == nil {
		absBasePath = evalPath
	}

	return &FileManager{
		basePath: absBasePath,
		parser:   NewParser(),
		writer:   NewWriter(),
	}
}

// ValidatePath は相対パスがベースパス内に収まるか検証
// ディレクトリトラバーサル攻撃を防止
func (fm *FileManager) ValidatePath(relativePath string) error {
	// 空パスを許可
	if relativePath == "" {
		return nil
	}

	// 絶対パスは不正
	if filepath.IsAbs(relativePath) {
		return ErrPathTraversal
	}

	// パスを正規化
	cleanPath := filepath.Clean(relativePath)

	// ".." で始まるパスは basePath 外へのアクセスを試みている
	if strings.HasPrefix(cleanPath, "..") {
		return ErrPathTraversal
	}

	// フルパスを計算して確認
	fullPath := filepath.Join(fm.basePath, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return ErrPathTraversal
	}

	// basePath 内に収まっているか確認
	// 末尾に区切り文字を追加して、部分一致を防止
	// (例: /base/path と /base/pathextra の区別)
	baseWithSep := fm.basePath + string(filepath.Separator)
	absWithSep := absPath + string(filepath.Separator)

	if !strings.HasPrefix(absWithSep, baseWithSep) && absPath != fm.basePath {
		return ErrPathTraversal
	}

	return nil
}

// ResolvePath は相対パスを絶対パスに変換（検証付き）
func (fm *FileManager) ResolvePath(relativePath string) (string, error) {
	if err := fm.ValidatePath(relativePath); err != nil {
		return "", err
	}
	return filepath.Join(fm.basePath, relativePath), nil
}

// resolvePathUnsafe は内部用の検証なしパス解決（後方互換性）
// 注意: 新規コードでは使用禁止
func (fm *FileManager) resolvePathUnsafe(relativePath string) string {
	return filepath.Join(fm.basePath, relativePath)
}

// Exists はファイルが存在するか確認（Context対応）
func (fm *FileManager) Exists(ctx context.Context, relativePath string) bool {
	if ctx.Err() != nil {
		return false
	}

	path, err := fm.ResolvePath(relativePath)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// ReadYaml は YAML ファイルを読み込む（Context対応）
func (fm *FileManager) ReadYaml(ctx context.Context, relativePath string, v any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := fm.ResolvePath(relativePath)
	if err != nil {
		return err
	}
	return fm.parser.ReadFile(path, v)
}

// WriteYaml は YAML ファイルを書き込む（Context対応）
func (fm *FileManager) WriteYaml(ctx context.Context, relativePath string, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := fm.ResolvePath(relativePath)
	if err != nil {
		return err
	}
	return fm.writer.WriteFile(path, data)
}

// WriteFile はファイルを書き込む（バイナリ対応、Context対応）
func (fm *FileManager) WriteFile(ctx context.Context, relativePath string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fullPath, err := fm.ResolvePath(relativePath)
	if err != nil {
		return err
	}
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

// EnsureDir はディレクトリを作成（Context対応）
func (fm *FileManager) EnsureDir(ctx context.Context, relativePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := fm.ResolvePath(relativePath)
	if err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

// Copy はファイルをコピー（Context対応）
func (fm *FileManager) Copy(ctx context.Context, src, dest string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	srcPath, err := fm.ResolvePath(src)
	if err != nil {
		return err
	}
	destPath, err := fm.ResolvePath(dest)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
}

// Delete はファイルを削除（Context対応）
func (fm *FileManager) Delete(ctx context.Context, relativePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := fm.ResolvePath(relativePath)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// Glob はパターンに一致するファイルを検索（Context対応）
func (fm *FileManager) Glob(ctx context.Context, pattern string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// パターンも検証（基本的なチェック）
	if strings.Contains(pattern, "..") {
		return nil, ErrPathTraversal
	}

	fullPattern := fm.resolvePathUnsafe(pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}

	// basePath からの相対パスに変換し、検証
	relPaths := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, err := filepath.Rel(fm.basePath, match)
		if err != nil {
			continue
		}
		// 結果が basePath 外にならないか確認
		if err := fm.ValidatePath(rel); err != nil {
			continue
		}
		relPaths = append(relPaths, rel)
	}

	return relPaths, nil
}

// ListDir はディレクトリ内のファイルを列挙（Context対応）
func (fm *FileManager) ListDir(ctx context.Context, relativePath string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fullPath, err := fm.ResolvePath(relativePath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// BasePath はベースパスを返す（テスト用）
func (fm *FileManager) BasePath() string {
	return fm.basePath
}
