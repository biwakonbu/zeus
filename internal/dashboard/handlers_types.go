package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
)

// =============================================================================
// 共通型定義
// =============================================================================

// ErrorResponse はエラーレスポンス
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// =============================================================================
// 共通ヘルパー関数
// =============================================================================

// writeJSON は JSON レスポンスを書き込む
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError はエラーレスポンスを書き込む
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// hasYamlSuffix は .yaml または .yml 拡張子を持つかチェック
func hasYamlSuffix(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}

// riskScoreToInt は RiskScore 文字列を数値スコアに変換
func riskScoreToInt(score string) int {
	switch score {
	case "critical":
		return 9
	case "high":
		return 6
	case "medium":
		return 4
	case "low":
		return 2
	default:
		return 0
	}
}

// severityRank は深刻度のランクを返す
func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// riskScoreToSeverity はリスクスコアを深刻度に変換
func riskScoreToSeverity(score string) string {
	switch score {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

// escapeForMermaidDiagram は Mermaid 用にエスケープ
func escapeForMermaidDiagram(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
