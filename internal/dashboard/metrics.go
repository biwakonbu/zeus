package dashboard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	metricsDir          = ".zeus/metrics"
	metricsMaxBodyBytes = 5 << 20
)

// MetricsPayload はメトリクス送信のペイロード
type MetricsPayload struct {
	SessionID string                 `json:"session_id"`
	Reason    string                 `json:"reason"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	Entries   []json.RawMessage      `json:"entries"`
}

// MetricsRecord は保存用のメトリクスレコード
type MetricsRecord struct {
	SessionID  string                 `json:"session_id"`
	Reason     string                 `json:"reason,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
	Entry      json.RawMessage        `json:"entry"`
	ReceivedAt string                 `json:"received_at"`
}

// MetricsResponse はメトリクス保存結果
type MetricsResponse struct {
	SessionID string `json:"session_id"`
	Saved     int    `json:"saved"`
	Skipped   int    `json:"skipped"`
	Path      string `json:"path"`
}

// handleAPIMetrics はメトリクス保存 API を処理
func (s *Server) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST メソッドのみ許可されています")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, metricsMaxBodyBytes)
	defer r.Body.Close()

	var payload MetricsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "JSON の解析に失敗しました: "+err.Error())
		return
	}

	if len(payload.Entries) == 0 {
		writeError(w, http.StatusBadRequest, "entries が空です")
		return
	}

	sessionID := sanitizeSessionID(payload.SessionID)
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	path, saved, skipped, err := appendMetrics(sessionID, payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MetricsResponse{
		SessionID: sessionID,
		Saved:     saved,
		Skipped:   skipped,
		Path:      path,
	})
}

func appendMetrics(sessionID string, payload MetricsPayload) (string, int, int, error) {
	if err := os.MkdirAll(metricsDir, 0o755); err != nil {
		return "", 0, 0, fmt.Errorf("metrics ディレクトリの作成に失敗: %w", err)
	}

	filename := fmt.Sprintf("dashboard-metrics-%s.jsonl", sessionID)
	path := filepath.Join(metricsDir, filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", 0, 0, fmt.Errorf("metrics ファイルの作成に失敗: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	receivedAt := time.Now().UTC().Format(time.RFC3339Nano)
	saved := 0
	skipped := 0

	for _, entry := range payload.Entries {
		if !json.Valid(entry) {
			skipped++
			continue
		}
		record := MetricsRecord{
			SessionID:  sessionID,
			Reason:     payload.Reason,
			Meta:       payload.Meta,
			Entry:      entry,
			ReceivedAt: receivedAt,
		}
		line, err := json.Marshal(record)
		if err != nil {
			skipped++
			continue
		}
		if _, err := writer.Write(line); err != nil {
			return "", saved, skipped, fmt.Errorf("metrics 書き込みに失敗: %w", err)
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return "", saved, skipped, fmt.Errorf("metrics 書き込みに失敗: %w", err)
		}
		saved++
	}

	return path, saved, skipped, nil
}

func sanitizeSessionID(sessionID string) string {
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func generateSessionID() string {
	ts := time.Now().UTC().Format("20060102-150405")
	nanos := strconv.FormatInt(time.Now().UnixNano(), 10)
	return ts + "-" + nanos
}
