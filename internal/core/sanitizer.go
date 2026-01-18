package core

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Sanitizer は入力値のサニタイズを行う
type Sanitizer struct {
	// フィールドごとの最大長
	maxLengths map[string]int
}

// NewSanitizer は新しい Sanitizer を作成
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		maxLengths: map[string]int{
			"id":          50,
			"title":       200,
			"description": 5000,
			"statement":   2000,
			"rationale":   5000,
			"owner":       100,
		},
	}
}

// SanitizeString はフィールド値をサニタイズする
func (s *Sanitizer) SanitizeString(field, value string) (string, error) {
	// 1. 長さチェック
	if maxLen, ok := s.maxLengths[field]; ok {
		if len(value) > maxLen {
			return "", &ValidationError{
				Field:   field,
				Message: fmt.Sprintf("exceeds maximum length of %d characters", maxLen),
			}
		}
	}

	// 2. 制御文字を除去（改行、タブは許可）
	value = s.removeControlChars(value)

	// 3. HTML タグを除去（シンプルなサニタイズ）
	if field == "description" || field == "rationale" || field == "statement" {
		value = s.sanitizeHTML(value)
	}

	// 4. Unicode NFC 正規化
	value = norm.NFC.String(value)

	return value, nil
}

// removeControlChars は制御文字を除去する（改行、タブは保持）
func (s *Sanitizer) removeControlChars(value string) string {
	return strings.Map(func(r rune) rune {
		// 改行、キャリッジリターン、タブは許可
		if r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		// その他の制御文字は除去
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
}

// sanitizeHTML は HTML エンティティエスケープを行い、プレーンテキストを保持
// XSS 攻撃を防ぐため、全ての HTML メタ文字をエスケープ
func (s *Sanitizer) sanitizeHTML(value string) string {
	// html.EscapeString: <, >, &, ", ' をエスケープ
	// これにより、以下の攻撃を防止：
	// - スクリプトタグ: <script>alert()</script> → &lt;script&gt;...&lt;/script&gt;
	// - イベントハンドラ: onclick="..." → onclick=&quot;...&quot;
	// - ネストしたタグ: <script><script>...</script></script> → 両方エスケープ
	// - SVG インラインスクリプト: <svg onload="..."> → onload=&quot;...&quot;
	return html.EscapeString(value)
}

// SanitizeOwner は owner フィールドをサニタイズする
func (s *Sanitizer) SanitizeOwner(value string) (string, error) {
	// 許可文字パターン
	pattern := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

	if !pattern.MatchString(value) {
		return "", &ValidationError{
			Field:   "owner",
			Message: "contains invalid characters (allowed: a-z, A-Z, 0-9, -, _)",
		}
	}

	if len(value) > 100 {
		return "", &ValidationError{
			Field:   "owner",
			Message: "exceeds maximum length of 100 characters",
		}
	}

	return value, nil
}

// SanitizeTags は tags 配列をサニタイズする
func (s *Sanitizer) SanitizeTags(tags []string) ([]string, error) {
	if len(tags) > 10 {
		return nil, &ValidationError{
			Field:   "tags",
			Message: "exceeds maximum of 10 tags",
		}
	}

	pattern := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	result := make([]string, 0, len(tags))

	for _, tag := range tags {
		if len(tag) > 20 {
			return nil, &ValidationError{
				Field:   "tags",
				Message: fmt.Sprintf("tag '%s' exceeds maximum length of 20 characters", tag),
			}
		}

		if !pattern.MatchString(tag) {
			return nil, &ValidationError{
				Field:   "tags",
				Message: fmt.Sprintf("tag '%s' contains invalid characters", tag),
			}
		}

		result = append(result, strings.ToLower(tag))
	}

	return result, nil
}

// SanitizeID は ID 形式をサニタイズする
func (s *Sanitizer) SanitizeID(entityType, id string) error {
	return ValidateID(entityType, id)
}
