package yaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Parser は YAML パーサー
type Parser struct{}

// NewParser は新しい Parser を作成
func NewParser() *Parser {
	return &Parser{}
}

// Parse は YAML 文字列をパース
func (p *Parser) Parse(content []byte, v interface{}) error {
	return yaml.Unmarshal(content, v)
}

// ReadFile は YAML ファイルを読み込む
func (p *Parser) ReadFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return p.Parse(data, v)
}
