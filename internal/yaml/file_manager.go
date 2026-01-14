package yaml

import (
	"os"
	"path/filepath"
)

// FileManager はファイル操作を管理
type FileManager struct {
	basePath string
	parser   *Parser
	writer   *Writer
}

// NewFileManager は新しい FileManager を作成
func NewFileManager(basePath string) *FileManager {
	return &FileManager{
		basePath: basePath,
		parser:   NewParser(),
		writer:   NewWriter(),
	}
}

// ResolvePath は相対パスを絶対パスに変換
func (fm *FileManager) ResolvePath(relativePath string) string {
	return filepath.Join(fm.basePath, relativePath)
}

// Exists はファイルが存在するか確認
func (fm *FileManager) Exists(relativePath string) bool {
	_, err := os.Stat(fm.ResolvePath(relativePath))
	return err == nil
}

// ReadYaml は YAML ファイルを読み込む
func (fm *FileManager) ReadYaml(relativePath string, v interface{}) error {
	return fm.parser.ReadFile(fm.ResolvePath(relativePath), v)
}

// WriteYaml は YAML ファイルを書き込む
func (fm *FileManager) WriteYaml(relativePath string, data interface{}) error {
	return fm.writer.WriteFile(fm.ResolvePath(relativePath), data)
}

// EnsureDir はディレクトリを作成
func (fm *FileManager) EnsureDir(relativePath string) error {
	return os.MkdirAll(fm.ResolvePath(relativePath), 0755)
}

// Copy はファイルをコピー
func (fm *FileManager) Copy(src, dest string) error {
	data, err := os.ReadFile(fm.ResolvePath(src))
	if err != nil {
		return err
	}
	return os.WriteFile(fm.ResolvePath(dest), data, 0644)
}

// Delete はファイルを削除
func (fm *FileManager) Delete(relativePath string) error {
	return os.Remove(fm.ResolvePath(relativePath))
}
