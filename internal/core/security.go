package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// SecurityError はセキュリティ関連のエラー
type SecurityError struct {
	Type    string // エラータイプ（path_traversal, null_byte, control_char, invalid_id）
	Message string
}

func (e *SecurityError) Error() string {
	return e.Message
}

// ValidationError はバリデーション関連のエラー
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// idPatterns は各エンティティタイプの ID パターン
var idPatterns = map[string]*regexp.Regexp{
	"vision":        regexp.MustCompile(`^vision-[0-9]{3}$`),
	"objective":     regexp.MustCompile(`^obj-[0-9]{3}$`),
	"deliverable":   regexp.MustCompile(`^del-[0-9]{3}$`),
	"consideration": regexp.MustCompile(`^con-[0-9]{3}$`),
	"decision":      regexp.MustCompile(`^dec-[0-9]{3}$`),
	"problem":       regexp.MustCompile(`^prob-[0-9]{3}$`),
	"risk":          regexp.MustCompile(`^risk-[0-9]{3}$`),
	"assumption":    regexp.MustCompile(`^assum-[0-9]{3}$`),
	"constraint":    regexp.MustCompile(`^const-[0-9]{3}$`),
	"quality":       regexp.MustCompile(`^qual-[0-9]{3}$`),
	// 既存の Task エンティティ（UUID ベース）
	"task": regexp.MustCompile(`^task-[a-f0-9]{8}$`),
}

// entityDirectories はエンティティタイプとディレクトリのマッピング
var entityDirectories = map[string]string{
	"vision":        "",              // ルートに配置（vision.yaml）
	"objective":     "objectives",    // objectives/obj-NNN.yaml
	"deliverable":   "deliverables",  // deliverables/del-NNN.yaml
	"consideration": "considerations",
	"decision":      "decisions",
	"problem":       "problems",
	"risk":          "risks",
	"assumption":    "assumptions",
	"constraint":    "",               // ルートに配置（constraints.yaml）
	"quality":       "quality",
	"task":          "tasks",          // 既存
}

// ValidatePath はパストラバーサル攻撃を防ぐ
//
// パス検証を実行し、安全なパスを返す。
// 攻撃的なパス（../、ヌルバイト、制御文字など）を検出した場合はエラーを返す。
func ValidatePath(baseDir, requestedPath string) (string, error) {
	// 1. ベースディレクトリを絶対パスに変換
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	// 2. ヌルバイトチェック（早期検出）
	if strings.Contains(requestedPath, "\x00") {
		return "", &SecurityError{
			Type:    "null_byte",
			Message: "access denied: null byte detected in path",
		}
	}

	// 3. 制御文字チェック（改行、タブ、キャリッジリターンは許可しない）
	for _, r := range requestedPath {
		if unicode.IsControl(r) {
			return "", &SecurityError{
				Type:    "control_char",
				Message: "access denied: control character detected in path",
			}
		}
	}

	// 4. リクエストパスを結合して絶対パスに変換
	fullPath := filepath.Join(absBase, requestedPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// 5. Clean 処理（冗長なセパレータ、. .. を除去）
	cleanPath := filepath.Clean(absPath)

	// 6. ベースディレクトリ配下か確認
	// ベースディレクトリ自体も許可
	if cleanPath == absBase {
		return cleanPath, nil
	}

	// ベースディレクトリ + セパレータで始まることを確認
	if !strings.HasPrefix(cleanPath, absBase+string(filepath.Separator)) {
		return "", &SecurityError{
			Type:    "path_traversal",
			Message: "access denied: path is outside base directory",
		}
	}

	return cleanPath, nil
}

// ValidateID はエンティティ ID の形式を検証する
func ValidateID(entityType, id string) error {
	pattern, ok := idPatterns[entityType]
	if !ok {
		return &ValidationError{
			Field:   "entity_type",
			Message: fmt.Sprintf("unknown entity type: %s", entityType),
		}
	}

	if !pattern.MatchString(id) {
		return &ValidationError{
			Field:   "id",
			Message: fmt.Sprintf("invalid ID format: %s (expected pattern: %s)", id, pattern.String()),
		}
	}

	return nil
}

// GetEntityFilePath は ID からファイルパスを安全に生成する
func GetEntityFilePath(baseDir, entityType, id string) (string, error) {
	// 1. ID バリデーション
	if err := ValidateID(entityType, id); err != nil {
		return "", err
	}

	// 2. ディレクトリ名を取得
	dirName, ok := entityDirectories[entityType]
	if !ok {
		return "", fmt.Errorf("unknown entity type: %s", entityType)
	}

	// 3. ファイル名を構築
	var relativePath string
	if dirName == "" {
		// ルートに配置する単一ファイルエンティティ
		if entityType == "vision" {
			relativePath = "vision.yaml"
		} else if entityType == "constraint" {
			relativePath = "constraints.yaml"
		} else {
			relativePath = id + ".yaml"
		}
	} else {
		relativePath = filepath.Join(dirName, id+".yaml")
	}

	// 4. パストラバーサルチェック
	return ValidatePath(baseDir, relativePath)
}

// IsValidEntityType はエンティティタイプが有効かどうかを確認
func IsValidEntityType(entityType string) bool {
	_, ok := idPatterns[entityType]
	return ok
}

// GetEntityDirectory はエンティティタイプのディレクトリを取得
func GetEntityDirectory(entityType string) (string, bool) {
	dir, ok := entityDirectories[entityType]
	return dir, ok
}
