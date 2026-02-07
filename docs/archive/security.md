> **履歴資料（非正本）**  
> この文書は履歴資料。現行仕様の正本は `docs/README.md` 参照。

# Zeus セキュリティガイドライン

Zeus のセキュリティ対策と実装ガイドラインを定義する。

## 目次

1. [セキュリティ原則](#セキュリティ原則)
2. [パストラバーサル対策](#パストラバーサル対策)
3. [ID インジェクション対策](#id-インジェクション対策)
4. [入力サニタイズ](#入力サニタイズ)
5. [テンプレートの安全性](#テンプレートの安全性)
6. [監査ログ](#監査ログ)
7. [API セキュリティ](#api-セキュリティ)
8. [ファイル操作のセキュリティ](#ファイル操作のセキュリティ)
9. [セキュリティチェックリスト](#セキュリティチェックリスト)

---

## セキュリティ原則

### 設計方針

1. **多層防御**: 単一の対策に依存せず、複数のセキュリティレイヤーを実装
2. **最小権限**: 必要最小限のアクセス権限のみを付与
3. **デフォルト拒否**: 明示的に許可されていない操作は拒否
4. **入力検証**: すべての外部入力を信頼せず、検証を実施
5. **安全な失敗**: エラー時は安全な状態にフォールバック

### 脅威モデル

| 脅威 | 影響度 | 対策 |
|------|--------|------|
| パストラバーサル | 高 | パス検証、ベースディレクトリ制限 |
| ID インジェクション | 高 | 正規表現による形式チェック |
| YAML インジェクション | 中 | 入力サニタイズ、構造化データのみ許可 |
| テンプレートインジェクション | 中 | データとしてのみ処理、実行禁止 |
| 情報漏洩 | 中 | 監査ログ、エラーメッセージの制限 |

---

## パストラバーサル対策

### 実装ガイドライン

すべてのファイル操作は `ValidatePath` 関数を経由する。

```go
// internal/yaml/security.go

// ValidatePath はパストラバーサル攻撃を防ぐ
func ValidatePath(baseDir, requestedPath string) (string, error) {
    // 1. ベースディレクトリを絶対パスに変換
    absBase, err := filepath.Abs(baseDir)
    if err != nil {
        return "", fmt.Errorf("failed to resolve base directory: %w", err)
    }

    // 2. リクエストパスを結合して絶対パスに変換
    fullPath := filepath.Join(absBase, requestedPath)
    absPath, err := filepath.Abs(fullPath)
    if err != nil {
        return "", fmt.Errorf("failed to resolve path: %w", err)
    }

    // 3. Clean 処理（冗長なセパレータ、. .. を除去）
    cleanPath := filepath.Clean(absPath)

    // 4. ベースディレクトリ配下か確認
    if !strings.HasPrefix(cleanPath, absBase+string(filepath.Separator)) && cleanPath != absBase {
        return "", &SecurityError{
            Type:    "path_traversal",
            Message: "access denied: path is outside base directory",
        }
    }

    // 5. ヌルバイトチェック
    if strings.Contains(requestedPath, "\x00") {
        return "", &SecurityError{
            Type:    "null_byte",
            Message: "access denied: null byte detected in path",
        }
    }

    // 6. 制御文字チェック
    for _, r := range requestedPath {
        if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
            return "", &SecurityError{
                Type:    "control_char",
                Message: "access denied: control character detected in path",
            }
        }
    }

    return cleanPath, nil
}
```

### 検証対象

| 入力ソース | 検証方法 |
|-----------|---------|
| CLI 引数 | `ValidatePath` を適用 |
| API リクエスト | `ValidatePath` を適用 |
| YAML ファイル内のパス参照 | `ValidatePath` を適用 |
| テンプレートファイルパス | `ValidatePath` を適用 |

### テストケース

```go
func TestValidatePath(t *testing.T) {
    baseDir := "/app/.zeus"

    tests := []struct {
        name      string
        path      string
        wantError bool
    }{
        // 正常ケース
        {"valid path", "objectives/obj-001.yaml", false},
        {"nested path", "assumptions/templates/owasp.yaml", false},

        // 攻撃ケース
        {"path traversal ../", "../../../etc/passwd", true},
        {"path traversal encoded", "..%2F..%2F..%2Fetc%2Fpasswd", true},
        {"null byte", "file\x00.yaml", true},
        {"absolute path", "/etc/passwd", true},
        {"backslash traversal", "..\\..\\..\\etc\\passwd", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ValidatePath(baseDir, tt.path)
            if (err != nil) != tt.wantError {
                t.Errorf("ValidatePath() error = %v, wantError %v", err, tt.wantError)
            }
        })
    }
}
```

---

## ID インジェクション対策

### ID 形式の定義

各概念の ID は厳格なパターンに従う。

```go
// internal/core/id_validator.go

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
```

### ID からファイルパス生成

ID を直接ファイルパスに使用せず、必ずバリデーションを経由する。

```go
// GetFilePath は ID からファイルパスを安全に生成する
func (s *Store) GetFilePath(entityType, id string) (string, error) {
    // 1. ID バリデーション
    if err := ValidateID(entityType, id); err != nil {
        return "", err
    }

    // 2. ディレクトリ名を取得（ハードコードされたマッピング）
    dirName, ok := entityDirectories[entityType]
    if !ok {
        return "", fmt.Errorf("unknown entity type: %s", entityType)
    }

    // 3. ファイル名を構築（ID + 固定拡張子）
    filename := fmt.Sprintf("%s.yaml", id)

    // 4. パストラバーサルチェック
    relativePath := filepath.Join(dirName, filename)
    return ValidatePath(s.baseDir, relativePath)
}

// エンティティタイプとディレクトリのマッピング
var entityDirectories = map[string]string{
    "vision":        "",              // ルートに配置
    "objective":     "objectives",
    "deliverable":   "deliverables",
    "consideration": "considerations",
    "decision":      "decisions",
    "problem":       "problems",
    "risk":          "risks",
    "assumption":    "assumptions",
    "constraint":    "",              // constraints.yaml（単一ファイル）
    "quality":       "quality",
}
```

---

## 入力サニタイズ

### サニタイズルール

| フィールド | 最大長 | 許可文字 | 特殊処理 |
|-----------|--------|---------|---------|
| id | 50 | `[a-z0-9-]` | 自動生成のみ |
| title | 200 | 制御文字除去 | Unicode NFC 正規化 |
| description | 5000 | 制御文字除去 | HTML サニタイズ（許可タグのみ） |
| statement | 2000 | 制御文字除去 | Unicode NFC 正規化 |
| rationale | 5000 | 制御文字除去 | HTML サニタイズ |
| owner | 100 | `[a-zA-Z0-9-_]` | - |
| tags | 20/個, 最大10個 | `[a-zA-Z0-9-_]` | - |

### 実装

```go
// internal/core/sanitizer.go

type Sanitizer struct {
    // 許可する HTML タグ（description フィールド用）
    allowedHTMLTags []string

    // フィールドごとの最大長
    maxLengths map[string]int
}

func NewSanitizer() *Sanitizer {
    return &Sanitizer{
        allowedHTMLTags: []string{
            "p", "br", "strong", "em", "b", "i",
            "ul", "ol", "li",
            "a", "code", "pre",
        },
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

    // 3. HTML サニタイズ（description, rationale フィールドのみ）
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

// sanitizeHTML は許可されたタグのみを残す
func (s *Sanitizer) sanitizeHTML(value string) string {
    p := bluemonday.NewPolicy()

    // 許可するタグ
    for _, tag := range s.allowedHTMLTags {
        p.AllowElements(tag)
    }

    // a タグの href 属性のみ許可
    p.AllowAttrs("href").OnElements("a")

    // href の値を検証（http, https のみ）
    p.AllowURLSchemes("http", "https")

    return p.Sanitize(value)
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
```

---

## テンプレートの安全性

### 方針

テンプレートはデータとしてのみ扱い、Go テンプレートとして実行しない。

```go
// internal/template/loader.go

// LoadTemplate はテンプレートファイルを安全に読み込む
func LoadTemplate(baseDir, templateID string) (*Template, error) {
    // 1. テンプレート ID の検証
    if !isValidTemplateID(templateID) {
        return nil, &ValidationError{
            Field:   "template_id",
            Message: "invalid template ID format",
        }
    }

    // 2. パストラバーサルチェック
    relativePath := filepath.Join("assumptions", "templates", templateID+".yaml")
    safePath, err := ValidatePath(baseDir, relativePath)
    if err != nil {
        return nil, err
    }

    // 3. ファイル読み込み
    content, err := os.ReadFile(safePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read template: %w", err)
    }

    // 4. YAML としてパース（テンプレート構文は解釈しない）
    var template Template
    if err := yaml.Unmarshal(content, &template); err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }

    // 5. テンプレート内容の検証
    if err := validateTemplateContent(&template); err != nil {
        return nil, err
    }

    return &template, nil
}

// isValidTemplateID はテンプレート ID の形式を検証する
func isValidTemplateID(id string) bool {
    // 許可: a-z, 0-9, - のみ、最大50文字
    pattern := regexp.MustCompile(`^[a-z0-9\-]{1,50}$`)
    return pattern.MatchString(id)
}

// validateTemplateContent はテンプレート内容の安全性を検証する
func validateTemplateContent(t *Template) error {
    // 危険なパターンをチェック
    dangerousPatterns := []string{
        `{{`,          // Go テンプレート構文
        `${`,          // 変数展開
        `<script`,     // JavaScript
        `javascript:`, // JavaScript URL
    }

    checkFields := []string{t.Name, t.Description}
    for _, item := range t.Items {
        checkFields = append(checkFields, item.Title, item.Description, item.DefaultAssumption)
    }

    for _, field := range checkFields {
        for _, pattern := range dangerousPatterns {
            if strings.Contains(strings.ToLower(field), strings.ToLower(pattern)) {
                return &SecurityError{
                    Type:    "template_injection",
                    Message: fmt.Sprintf("dangerous pattern detected: %s", pattern),
                }
            }
        }
    }

    return nil
}
```

### テンプレート適用

テンプレートの適用は単純な値のコピーのみ行う。

```go
// ApplyTemplate はテンプレートから Assumption を生成する
func ApplyTemplate(template *Template, targetID string, selectedItems []string) ([]*Assumption, error) {
    var assumptions []*Assumption

    for _, itemID := range selectedItems {
        item := template.FindItem(itemID)
        if item == nil {
            continue
        }

        // 値の単純コピー（テンプレート構文の解釈なし）
        assumption := &Assumption{
            ID:               generateID("assumption"),
            Title:            item.Title,
            Description:      item.Description,
            Source:           fmt.Sprintf("template:%s:%s", template.ID, itemID),
            RelatedTo:        []string{targetID},
            ValidationMethod: item.ValidationMethod,
            Status:           "unvalidated",
            Metadata: Metadata{
                CreatedAt: time.Now(),
            },
        }

        assumptions = append(assumptions, assumption)
    }

    return assumptions, nil
}
```

---

## 監査ログ

### ログエントリ構造

```go
// internal/audit/logger.go

type AuditEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    Action      string                 `json:"action"`       // CREATE, READ, UPDATE, DELETE
    EntityType  string                 `json:"entity_type"`
    EntityID    string                 `json:"entity_id"`
    Result      string                 `json:"result"`       // success, failure
    Changes     []AuditChange          `json:"changes,omitempty"`
    ErrorDetail string                 `json:"error,omitempty"`
    ClientInfo  *ClientInfo            `json:"client_info,omitempty"`
}

type AuditChange struct {
    Field    string      `json:"field"`
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}

type ClientInfo struct {
    Source    string `json:"source"`     // cli, api
    UserAgent string `json:"user_agent,omitempty"`
    IP        string `json:"ip,omitempty"`
}
```

### ログ出力

```go
// AuditLogger は監査ログを記録する
type AuditLogger struct {
    writer io.Writer
    mu     sync.Mutex
}

func NewAuditLogger(logPath string) (*AuditLogger, error) {
    // ログディレクトリの作成
    dir := filepath.Dir(logPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create log directory: %w", err)
    }

    // ログファイルを開く（追記モード）
    file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to open log file: %w", err)
    }

    return &AuditLogger{writer: file}, nil
}

// Log は監査エントリを記録する
func (al *AuditLogger) Log(entry *AuditEntry) {
    entry.Timestamp = time.Now()

    al.mu.Lock()
    defer al.mu.Unlock()

    // JSON 形式で出力（1行1エントリ）
    data, err := json.Marshal(entry)
    if err != nil {
        // ログ出力失敗は標準エラーに出力
        fmt.Fprintf(os.Stderr, "audit log error: %v\n", err)
        return
    }

    al.writer.Write(append(data, '\n'))
}

// LogWithContext はコンテキストから情報を取得してログを記録する
func (al *AuditLogger) LogWithContext(ctx context.Context, action, entityType, entityID, result string, changes []AuditChange, err error) {
    entry := &AuditEntry{
        Action:     action,
        EntityType: entityType,
        EntityID:   entityID,
        Result:     result,
        Changes:    changes,
        ClientInfo: getClientInfoFromContext(ctx),
    }

    if err != nil {
        // エラーメッセージは内部情報を含まないようにサニタイズ
        entry.ErrorDetail = sanitizeErrorMessage(err)
    }

    al.Log(entry)
}
```

### ログの保護

```yaml
# ログファイルの権限設定
# .zeus/audit/audit.log: 644（所有者のみ書き込み可能）

# ログローテーション設定（zeus.yaml）
audit:
  enabled: true
  path: ".zeus/audit/audit.log"
  rotation:
    max_size: "10MB"
    max_files: 10
    compress: true
```

---

## API セキュリティ

### 入力検証

API エンドポイントでは、すべての入力パラメータを検証する。

```go
// internal/dashboard/handlers.go

func CreateObjectiveHandler(z *core.Zeus) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // 1. Content-Type チェック
        if r.Header.Get("Content-Type") != "application/json" {
            respondError(w, http.StatusBadRequest, "INVALID_CONTENT_TYPE", "Content-Type must be application/json")
            return
        }

        // 2. リクエストボディのサイズ制限
        r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB

        // 3. JSON パース
        var req CreateObjectiveRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, http.StatusBadRequest, "INVALID_JSON", "Failed to parse request body")
            return
        }

        // 4. バリデーション
        if err := validateCreateObjectiveRequest(&req); err != nil {
            respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
            return
        }

        // 5. サニタイズ
        sanitizer := core.NewSanitizer()
        req.Title, _ = sanitizer.SanitizeString("title", req.Title)
        req.Description, _ = sanitizer.SanitizeString("description", req.Description)

        // 6. 作成処理
        obj, err := z.CreateObjective(ctx, &req)
        if err != nil {
            handleError(w, err)
            return
        }

        respondJSON(w, http.StatusCreated, obj)
    }
}
```

### レート制限

```go
// internal/dashboard/middleware.go

func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(limit)), limit)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.Header().Set("Retry-After", "1")
                respondError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### CORS 設定

```go
// internal/dashboard/server.go

func corsMiddleware(devMode bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if devMode {
                // 開発モードでは localhost からのアクセスを許可
                w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
            } else {
                // 本番モードでは同一オリジンのみ
                w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
            }

            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, If-Match")
            w.Header().Set("Access-Control-Max-Age", "86400")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### エラーレスポンス

内部情報を漏洩しないよう、エラーメッセージを制御する。

```go
// handleError はエラーを適切な HTTP レスポンスに変換する
func handleError(w http.ResponseWriter, err error) {
    switch e := err.(type) {
    case *core.ValidationError:
        respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", e.Message)
    case *core.NotFoundError:
        respondError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
    case *core.ReferenceError:
        respondError(w, http.StatusConflict, "REFERENCE_ERROR", e.Message)
    case *core.SecurityError:
        // セキュリティエラーの詳細は漏洩しない
        respondError(w, http.StatusForbidden, "ACCESS_DENIED", "Access denied")
        // 詳細は監査ログに記録
        auditLogger.Log(&AuditEntry{
            Action:      "SECURITY_VIOLATION",
            Result:      "failure",
            ErrorDetail: e.Error(),
        })
    default:
        // 内部エラーの詳細は漏洩しない
        respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
        // 詳細は監査ログに記録
        auditLogger.Log(&AuditEntry{
            Action:      "INTERNAL_ERROR",
            Result:      "failure",
            ErrorDetail: err.Error(),
        })
    }
}
```

---

## ファイル操作のセキュリティ

### ファイルロック

同時編集による競合を防ぐため、ファイルロックを使用する。

```go
// internal/yaml/filelock.go

type FileLock struct {
    lockFile string
    file     *os.File
}

func NewFileLock(targetPath string) *FileLock {
    return &FileLock{
        lockFile: targetPath + ".lock",
    }
}

// Lock はロックを取得する
func (fl *FileLock) Lock() error {
    // 古いロックファイルをチェック
    if fl.isStale() {
        os.Remove(fl.lockFile)
    }

    // ロックファイルを排他的に作成
    file, err := os.OpenFile(fl.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
    if err != nil {
        if os.IsExist(err) {
            return &LockError{
                Path:    fl.lockFile,
                Message: "file is locked by another process",
            }
        }
        return err
    }

    // PID を書き込み
    fmt.Fprintf(file, "%d", os.Getpid())
    fl.file = file

    return nil
}

// Unlock はロックを解放する
func (fl *FileLock) Unlock() error {
    if fl.file != nil {
        fl.file.Close()
    }
    return os.Remove(fl.lockFile)
}

// isStale はロックが古いかどうかを判定する
func (fl *FileLock) isStale() bool {
    info, err := os.Stat(fl.lockFile)
    if err != nil {
        return false
    }

    // 5分以上古いロックは stale
    return time.Since(info.ModTime()) > 5*time.Minute
}
```

### アトミック書き込み

```go
// internal/yaml/writer.go

// WriteAtomic はファイルをアトミックに書き込む
func WriteAtomic(path string, content []byte) error {
    // 1. 一時ファイルに書き込み
    dir := filepath.Dir(path)
    tempFile, err := os.CreateTemp(dir, ".zeus-*.tmp")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    tempPath := tempFile.Name()

    defer func() {
        // クリーンアップ
        tempFile.Close()
        os.Remove(tempPath)
    }()

    // 2. コンテンツを書き込み
    if _, err := tempFile.Write(content); err != nil {
        return fmt.Errorf("failed to write content: %w", err)
    }

    // 3. 同期
    if err := tempFile.Sync(); err != nil {
        return fmt.Errorf("failed to sync: %w", err)
    }

    // 4. ファイルを閉じる
    if err := tempFile.Close(); err != nil {
        return fmt.Errorf("failed to close: %w", err)
    }

    // 5. リネーム（アトミック操作）
    if err := os.Rename(tempPath, path); err != nil {
        return fmt.Errorf("failed to rename: %w", err)
    }

    return nil
}
```

---

## セキュリティチェックリスト

### 実装時チェック

- [ ] すべてのファイルパスは `ValidatePath` を経由している
- [ ] すべての ID は `ValidateID` で検証している
- [ ] ユーザー入力はサニタイズしている
- [ ] テンプレートはデータとして処理している
- [ ] エラーメッセージに内部情報を含めていない
- [ ] 監査ログを記録している

### コードレビューチェック

- [ ] SQL インジェクション対策（該当する場合）
- [ ] コマンドインジェクション対策
- [ ] YAML インジェクション対策
- [ ] 機密情報のハードコーディングなし
- [ ] 適切なファイル権限設定

### デプロイ前チェック

- [ ] 開発用設定が本番に含まれていない
- [ ] デバッグログが無効化されている
- [ ] CORS 設定が適切
- [ ] ログファイルの権限が適切

---

## セキュリティ更新履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2026-01-18 | 1.0 | 初版作成 |

---

## 関連ドキュメント

- [ドキュメント正本マトリクス](../README.md) - 現行文書の入口
- [システム設計書](../system-design.md) - 現行設計の正本
- [API リファレンス](../api-reference.md) - 現行公開契約

---

*作成日: 2026-01-18*
*バージョン: 1.0*
*再分類日: 2026-02-07（archive移行）*
