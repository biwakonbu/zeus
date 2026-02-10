package core

import (
	"os"
	"path/filepath"
	"testing"
)

// ===== ValidatePath テスト =====

func TestValidatePath(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 正常なパスの検証
	validPath, err := ValidatePath(tmpDir, "subdir/file.yaml")
	if err != nil {
		t.Fatalf("ValidatePath failed for valid path: %v", err)
	}

	expected := filepath.Join(tmpDir, "subdir", "file.yaml")
	if validPath != expected {
		t.Errorf("expected %q, got %q", expected, validPath)
	}
}

func TestValidatePath_BaseDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// ベースディレクトリ自体も許可される
	validPath, err := ValidatePath(tmpDir, "")
	if err != nil {
		t.Fatalf("ValidatePath failed for base dir: %v", err)
	}

	absBase, _ := filepath.Abs(tmpDir)
	if validPath != absBase {
		t.Errorf("expected %q, got %q", absBase, validPath)
	}
}

func TestValidatePath_PathTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name string
		path string
	}{
		{"simple traversal", "../etc/passwd"},
		{"double traversal", "../../etc/passwd"},
		{"nested traversal", "subdir/../../../etc/passwd"},
		// Note: "/etc/passwd" は filepath.Join により "etc/passwd" として扱われ、
		// 基底ディレクトリ内に結合されるためパストラバーサルにはならない
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePath(tmpDir, tc.path)
			if err == nil {
				t.Errorf("expected error for path traversal: %s", tc.path)
				return
			}

			secErr, ok := err.(*SecurityError)
			if !ok {
				t.Errorf("expected SecurityError, got %T", err)
				return
			}

			if secErr.Type != "path_traversal" {
				t.Errorf("expected error type 'path_traversal', got %q", secErr.Type)
			}
		})
	}
}

func TestValidatePath_NullByte(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name string
		path string
	}{
		{"null at end", "file.yaml\x00"},
		{"null in middle", "file\x00.yaml"},
		{"null at start", "\x00file.yaml"},
		{"multiple nulls", "fi\x00le\x00.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePath(tmpDir, tc.path)
			if err == nil {
				t.Error("expected error for null byte in path")
				return
			}

			secErr, ok := err.(*SecurityError)
			if !ok {
				t.Errorf("expected SecurityError, got %T", err)
				return
			}

			if secErr.Type != "null_byte" {
				t.Errorf("expected error type 'null_byte', got %q", secErr.Type)
			}
		})
	}
}

func TestValidatePath_ControlCharacters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name string
		path string
	}{
		{"newline", "file\n.yaml"},
		{"carriage return", "file\r.yaml"},
		{"tab", "file\t.yaml"},
		{"bell", "file\a.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePath(tmpDir, tc.path)
			if err == nil {
				t.Error("expected error for control character in path")
				return
			}

			secErr, ok := err.(*SecurityError)
			if !ok {
				t.Errorf("expected SecurityError, got %T", err)
				return
			}

			if secErr.Type != "control_char" {
				t.Errorf("expected error type 'control_char', got %q", secErr.Type)
			}
		})
	}
}

// ===== ValidateID テスト =====

func TestValidateID(t *testing.T) {
	testCases := []struct {
		entityType string
		id         string
		wantErr    bool
	}{
		// 正常なケース
		{"task", "task-12345678", false},
		{"vision", "vision-001", false},
		{"objective", "obj-001", false},
		{"consideration", "con-001", false},
		{"decision", "dec-001", false},
		{"problem", "prob-001", false},
		{"risk", "risk-001", false},
		{"assumption", "assum-001", false},
		{"constraint", "const-001", false},
		{"quality", "qual-001", false},
		{"actor", "actor-12345678", false},
		{"usecase", "uc-12345678", false},
		{"subsystem", "sub-12345678", false},
		{"activity", "act-001", false},
		// 無効なケース
		{"task", "task-123", true},           // UUID短すぎ
		{"objective", "obj-1", true},         // 番号短すぎ
		{"objective", "obj-1234", true},      // 番号長すぎ
		{"decision", "dec-abc", true},        // 番号ではない
		{"unknown", "unknown-001", true},     // 不明なエンティティタイプ
		{"task", "objective-001", true},      // 間違ったプレフィックス
		{"objective", "task-12345678", true}, // 間違ったプレフィックス
	}

	for _, tc := range testCases {
		t.Run(tc.entityType+"/"+tc.id, func(t *testing.T) {
			err := ValidateID(tc.entityType, tc.id)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateID(%q, %q) error = %v, wantErr %v", tc.entityType, tc.id, err, tc.wantErr)
			}
		})
	}
}

func TestValidateID_AllEntityTypes(t *testing.T) {
	// 全 15 エンティティタイプをテスト
	entityTests := []struct {
		entityType string
		validID    string
	}{
		{"vision", "vision-001"},
		{"objective", "obj-001"},
		{"consideration", "con-001"},
		{"decision", "dec-001"},
		{"problem", "prob-001"},
		{"risk", "risk-001"},
		{"assumption", "assum-001"},
		{"constraint", "const-001"},
		{"quality", "qual-001"},
		{"task", "task-12345678"},
		{"actor", "actor-12345678"},
		{"usecase", "uc-12345678"},
		{"subsystem", "sub-12345678"},
		{"activity", "act-001"},
	}

	for _, et := range entityTests {
		t.Run(et.entityType, func(t *testing.T) {
			err := ValidateID(et.entityType, et.validID)
			if err != nil {
				t.Errorf("ValidateID(%q, %q) failed: %v", et.entityType, et.validID, err)
			}
		})
	}
}

func TestValidateID_UnknownEntityType(t *testing.T) {
	err := ValidateID("unknown_type", "unknown-001")
	if err == nil {
		t.Error("expected error for unknown entity type")
		return
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
		return
	}

	if valErr.Field != "entity_type" {
		t.Errorf("expected field 'entity_type', got %q", valErr.Field)
	}
}

func TestValidateID_InvalidFormat(t *testing.T) {
	err := ValidateID("objective", "invalid-format")
	if err == nil {
		t.Error("expected error for invalid ID format")
		return
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
		return
	}

	if valErr.Field != "id" {
		t.Errorf("expected field 'id', got %q", valErr.Field)
	}
}

// ===== GetEntityFilePath テスト =====

func TestGetEntityFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		entityType   string
		id           string
		expectedPath string
	}{
		{"objective", "obj-001", "objectives/obj-001.yaml"},
		{"decision", "dec-001", "decisions/dec-001.yaml"},
		{"problem", "prob-001", "problems/prob-001.yaml"},
		{"risk", "risk-001", "risks/risk-001.yaml"},
		{"assumption", "assum-001", "assumptions/assum-001.yaml"},
		{"quality", "qual-001", "quality/qual-001.yaml"},
		{"task", "task-12345678", "tasks/task-12345678.yaml"},
		{"usecase", "uc-12345678", "usecases/uc-12345678.yaml"},
		{"activity", "act-001", "activities/act-001.yaml"},
		// 単一ファイルエンティティ
		{"vision", "vision-001", "vision.yaml"},
		{"constraint", "const-001", "constraints.yaml"},
		{"actor", "actor-12345678", "actors.yaml"},
		{"subsystem", "sub-12345678", "subsystems.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.entityType+"/"+tc.id, func(t *testing.T) {
			result, err := GetEntityFilePath(tmpDir, tc.entityType, tc.id)
			if err != nil {
				t.Fatalf("GetEntityFilePath failed: %v", err)
			}

			expected := filepath.Join(tmpDir, tc.expectedPath)
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	}
}

func TestGetEntityFilePath_InvalidID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = GetEntityFilePath(tmpDir, "objective", "invalid-id")
	if err == nil {
		t.Error("expected error for invalid ID")
	}
}

func TestGetEntityFilePath_UnknownEntityType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = GetEntityFilePath(tmpDir, "unknown", "unknown-001")
	if err == nil {
		t.Error("expected error for unknown entity type")
	}
}

// ===== IsValidEntityType テスト =====

func TestIsValidEntityType(t *testing.T) {
	validTypes := []string{
		"vision", "objective", "consideration",
		"decision", "problem", "risk", "assumption", "constraint",
		"quality", "task", "actor", "usecase", "subsystem", "activity",
	}

	for _, et := range validTypes {
		t.Run(et, func(t *testing.T) {
			if !IsValidEntityType(et) {
				t.Errorf("IsValidEntityType(%q) = false, want true", et)
			}
		})
	}
}

func TestIsValidEntityType_Invalid(t *testing.T) {
	invalidTypes := []string{
		"unknown", "invalid", "", "OBJECTIVE", "Task", // 大文字小文字は区別
	}

	for _, et := range invalidTypes {
		t.Run(et, func(t *testing.T) {
			if IsValidEntityType(et) {
				t.Errorf("IsValidEntityType(%q) = true, want false", et)
			}
		})
	}
}

// ===== GetEntityDirectory テスト =====

func TestGetEntityDirectory(t *testing.T) {
	testCases := []struct {
		entityType string
		wantDir    string
		wantOK     bool
	}{
		{"objective", "objectives", true},
		{"task", "tasks", true},
		{"usecase", "usecases", true},
		// 単一ファイルエンティティ
		{"vision", "", true},
		{"constraint", "", true},
		{"actor", "", true},
		{"subsystem", "", true},
		// 不明なエンティティ
		{"unknown", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.entityType, func(t *testing.T) {
			dir, ok := GetEntityDirectory(tc.entityType)
			if ok != tc.wantOK {
				t.Errorf("GetEntityDirectory(%q) ok = %v, want %v", tc.entityType, ok, tc.wantOK)
			}
			if dir != tc.wantDir {
				t.Errorf("GetEntityDirectory(%q) dir = %q, want %q", tc.entityType, dir, tc.wantDir)
			}
		})
	}
}

// ===== エラー型テスト =====

func TestSecurityErrorMessage(t *testing.T) {
	err := &SecurityError{
		Type:    "path_traversal",
		Message: "access denied: path is outside base directory",
	}

	if err.Error() != "access denied: path is outside base directory" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := &ValidationError{
		Field:   "id",
		Message: "invalid ID format",
	}

	expected := "id: invalid ID format"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
