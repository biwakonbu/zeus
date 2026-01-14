package core

import "errors"

var (
	// ErrConfigNotFound は設定ファイルが見つからない
	ErrConfigNotFound = errors.New("zeus.yaml not found")
	// ErrYamlSyntax は YAML 構文エラー
	ErrYamlSyntax = errors.New("YAML syntax error")
	// ErrStateMismatch は状態の不整合
	ErrStateMismatch = errors.New("state mismatch")
	// ErrEntityNotFound はエンティティが見つからない
	ErrEntityNotFound = errors.New("entity not found")
	// ErrUnknownEntity は不明なエンティティ
	ErrUnknownEntity = errors.New("unknown entity type")
)
