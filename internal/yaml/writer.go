package yaml

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Writer は YAML ライター
type Writer struct{}

// NewWriter は新しい Writer を作成
func NewWriter() *Writer {
	return &Writer{}
}

// Stringify はデータを YAML 文字列に変換
func (w *Writer) Stringify(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

// WriteFile は YAML ファイルを書き込む
func (w *Writer) WriteFile(path string, data interface{}) error {
	content, err := w.Stringify(data)
	if err != nil {
		return err
	}

	// ディレクトリがなければ作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}
