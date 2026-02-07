> **履歴資料（非正本）**  
> この文書は履歴資料。現行仕様の正本は `docs/README.md` 参照。

# Zeus 実装ガイド（Go版）

## 1. はじめに

このガイドでは、Zeus システムを Go + Cobra で実装する手順を説明します。

## 2. 技術スタック

| 項目 | 選定 |
|------|------|
| 実装言語 | Go 1.21+ |
| CLIフレームワーク | Cobra |
| データ形式 | YAML（gopkg.in/yaml.v3） |
| AI接続 | Claude Code 経由 |
| 配布形式 | スタンドアロン CLI + Claude Code Plugin |

## 3. プロジェクト構造

### 3.1 Zeus CLI ソースコード構造

```
zeus/
├── cmd/                      # Cobra コマンド
│   ├── root.go               # ルートコマンド
│   ├── init.go               # zeus init
│   ├── status.go             # zeus status
│   ├── add.go                # zeus add
│   ├── list.go               # zeus list
│   ├── suggest.go            # zeus suggest
│   ├── apply.go              # zeus apply
│   ├── approve.go            # zeus approve/reject
│   ├── pending.go            # zeus pending
│   ├── doctor.go             # zeus doctor
│   ├── fix.go                # zeus fix
│   ├── graph.go              # zeus graph（Phase 4）
│   ├── predict.go            # zeus predict（Phase 4）
│   ├── report.go             # zeus report（Phase 4）
│   └── dashboard.go          # zeus dashboard（Phase 5）
│
├── internal/                 # 内部パッケージ
│   ├── core/                 # コア機能
│   │   ├── zeus.go           # メインロジック
│   │   ├── state.go          # 状態管理
│   │   └── approval.go       # 承認管理
│   ├── analysis/             # 分析機能（Phase 4）
│   │   ├── types.go          # 分析用型定義
│   │   └── graph.go          # 依存関係グラフ
│   ├── report/               # レポート生成（Phase 4）
│   │   ├── generator.go      # レポート生成ロジック
│   │   └── templates.go      # 出力テンプレート
│   ├── dashboard/            # Web ダッシュボード（Phase 5）
│   │   ├── server.go         # HTTP サーバー
│   │   ├── handlers.go       # API ハンドラー
│   │   └── static/           # 静的ファイル（embed）
│   ├── yaml/                 # YAML操作
│   │   ├── parser.go
│   │   └── writer.go
│   ├── doctor/               # 診断・修復
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成
│       ├── agents.go         # agent テンプレート生成
│       └── skills.go         # skill テンプレート生成
│
├── templates/                # 埋め込みテンプレート（embed）
│   ├── zeus.yaml.tmpl
│   ├── task.yaml.tmpl
│   ├── state.yaml.tmpl
│   ├── agents/               # agent テンプレート
│   │   ├── zeus-orchestrator.md.tmpl
│   │   ├── zeus-planner.md.tmpl
│   │   └── zeus-reviewer.md.tmpl
│   └── skills/               # skill テンプレート
│       ├── project-scan.md.tmpl
│       ├── task-suggest.md.tmpl
│       └── risk-analysis.md.tmpl
│
├── docs/                     # ドキュメント
│   ├── SYSTEM_DESIGN.md
│   ├── IMPLEMENTATION_GUIDE.md
│   └── OPERATIONS_MANUAL.md
│
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── README.md
```

### 3.2 zeus init 実行後のターゲットリポジトリ構造

```
target-project/               # Zeus を適用するプロジェクト
├── .zeus/                    # Zeus プロジェクト管理
│   ├── zeus.yaml             # メイン設定
│   ├── tasks/                # タスク管理
│   │   ├── active.yaml
│   │   ├── backlog.yaml
│   │   └── _archive/
│   ├── state/                # 状態管理
│   │   ├── current.yaml
│   │   └── snapshots/
│   └── backups/              # 自動バックアップ
│
├── .claude/                  # Claude Code 標準構造
│   ├── agents/               # Zeus 用エージェント
│   │   ├── zeus-orchestrator.md
│   │   ├── zeus-planner.md
│   │   └── zeus-reviewer.md
│   └── skills/               # Zeus 用スキル
│       ├── zeus-suggest/
│       │   └── SKILL.md
│       ├── zeus-risk-analysis/
│       │   └── SKILL.md
│       └── zeus-e2e-tester/
│           └── SKILL.md
│
└── ... (既存のプロジェクトファイル)
```

## 4. 依存パッケージ

### 4.1 go.mod

```go
module github.com/biwakonbu/zeus

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/fatih/color v1.16.0
)
```

## 5. コア実装

### 5.1 main.go

```go
package main

import (
    "os"

    "github.com/biwakonbu/zeus/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 5.2 cmd/root.go

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "zeus",
    Short: "AI-driven project management with god's eye view",
    Long: `Zeus は AI によるプロジェクトマネジメントを「神の視点」で
俯瞰するシステムです。上流工程（方針立案からWBS化、タイムライン設計、
仕様作成まで）を支援します。`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // グローバルフラグ
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "詳細出力")
}
```

### 5.3 cmd/init.go

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/biwakonbu/zeus/internal/core"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Zeus プロジェクトを初期化",
    Long:  `プロジェクトディレクトリに .zeus/ と .claude/ を生成します。`,
    RunE: func(cmd *cobra.Command, args []string) error {
        zeus := core.New(".")
        result, err := zeus.Init()
        if err != nil {
            return err
        }

        fmt.Printf("Zeus initialized successfully!\n")
        fmt.Printf("  Path: %s\n", result.ZeusPath)
        fmt.Printf("  Claude integration: %s\n", result.ClaudePath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

### 5.4 internal/core/zeus.go

```go
package core

import (
    "embed"
    "os"
    "path/filepath"
    "time"

    "github.com/biwakonbu/zeus/internal/generator"
    "github.com/biwakonbu/zeus/internal/yaml"
)

//go:embed templates/*
var templates embed.FS

// Zeus はメインのアプリケーション構造体
type Zeus struct {
    ProjectPath string
    ZeusPath    string
    ClaudePath  string
    State       *StateManager
    Approval    *ApprovalManager
}

// InitResult は初期化結果
type InitResult struct {
    Success    bool
    ZeusPath   string
    ClaudePath string
}

// New は新しい Zeus インスタンスを作成
func New(projectPath string) *Zeus {
    return &Zeus{
        ProjectPath: projectPath,
        ZeusPath:    filepath.Join(projectPath, ".zeus"),
        ClaudePath:  filepath.Join(projectPath, ".claude"),
    }
}

// Init はプロジェクトを初期化
func (z *Zeus) Init() (*InitResult, error) {
    // .zeus ディレクトリ構造を作成
    if err := z.createZeusStructure(); err != nil {
        return nil, err
    }

    // zeus.yaml を生成
    config := z.generateInitialConfig()
    if err := yaml.WriteFile(filepath.Join(z.ZeusPath, "zeus.yaml"), config); err != nil {
        return nil, err
    }

    // .claude ディレクトリを作成（agents, skills）
    if err := generator.GenerateClaudeIntegration(z.ClaudePath); err != nil {
        return nil, err
    }

    return &InitResult{
        Success:    true,
        ZeusPath:   z.ZeusPath,
        ClaudePath: z.ClaudePath,
    }, nil
}

func (z *Zeus) createZeusStructure() error {
    dirs := []string{
        ".", "config", "tasks", "tasks/_archive",
        "state", "state/snapshots",
        "entities", "approvals/pending", "approvals/approved", "approvals/rejected",
        "logs", "analytics", "backups",
    }
    for _, dir := range dirs {
        path := filepath.Join(z.ZeusPath, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            return err
        }
    }
    return nil
}

func (z *Zeus) generateInitialConfig() map[string]interface{} {
    return map[string]interface{}{
        "version": "1.0",
        "project": map[string]interface{}{
            "id":          fmt.Sprintf("zeus-%d", time.Now().Unix()),
            "name":        "New Zeus Project",
            "description": "Project managed by Zeus",
            "start_date":  time.Now().Format("2006-01-02"),
        },
        "objectives": []interface{}{},
        "settings": map[string]interface{}{
            "automation_level": "auto",
            "approval_mode":    "default",
            "ai_provider":      "claude-code",
        },
    }
}
```

### 5.5 internal/yaml/parser.go

```go
package yaml

import (
    "os"

    "gopkg.in/yaml.v3"
)

// ReadFile は YAML ファイルを読み込む
func ReadFile(path string, v interface{}) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return yaml.Unmarshal(data, v)
}

// WriteFile は YAML ファイルを書き込む
func WriteFile(path string, v interface{}) error {
    data, err := yaml.Marshal(v)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

## 6. 分析モジュールの実装（Phase 4）

### 6.1 internal/analysis/types.go

分析パッケージは `core` への import cycle を避けるため、独自の型を定義します。

```go
package analysis

// AnalysisActivity は分析用の Activity 型
// core.Activity とは独立して定義し、変換関数で連携
type AnalysisActivity struct {
    ID           string
    Title        string
    Status       string
    Priority     string
    Dependencies []string
    CreatedAt    time.Time
    CompletedAt  *time.Time
}

// ActivityConverter は core.Activity から AnalysisActivity への変換インターフェース
type ActivityConverter interface {
    ToAnalysisActivity() AnalysisActivity
}
```

### 6.2 internal/analysis/graph.go

依存関係グラフの構築と可視化を担当します。

```go
package analysis

import (
    "fmt"
    "strings"
)

// GraphNode はグラフのノード
type GraphNode struct {
    ID       string
    Label    string
    Status   string
    Children []*GraphNode
}

// DependencyGraph は依存関係グラフ
type DependencyGraph struct {
    nodes map[string]*GraphNode
    edges map[string][]string  // 親ID -> 子IDのリスト
}

// NewDependencyGraph は新しいグラフを作成
func NewDependencyGraph() *DependencyGraph {
    return &DependencyGraph{
        nodes: make(map[string]*GraphNode),
        edges: make(map[string][]string),
    }
}

// AddNode はノードを追加
func (g *DependencyGraph) AddNode(id, label, status string) {
    g.nodes[id] = &GraphNode{
        ID:     id,
        Label:  label,
        Status: status,
    }
}

// AddEdge はエッジを追加（from -> to）
func (g *DependencyGraph) AddEdge(from, to string) {
    g.edges[from] = append(g.edges[from], to)
}

// DetectCycles は循環依存を検出
func (g *DependencyGraph) DetectCycles() [][]string {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    var cycles [][]string

    var dfs func(node string, path []string) bool
    dfs = func(node string, path []string) bool {
        visited[node] = true
        recStack[node] = true
        path = append(path, node)

        for _, child := range g.edges[node] {
            if !visited[child] {
                if dfs(child, path) {
                    return true
                }
            } else if recStack[child] {
                // 循環を検出
                cycleStart := -1
                for i, n := range path {
                    if n == child {
                        cycleStart = i
                        break
                    }
                }
                if cycleStart >= 0 {
                    cycle := append(path[cycleStart:], child)
                    cycles = append(cycles, cycle)
                }
            }
        }

        recStack[node] = false
        return false
    }

    for node := range g.nodes {
        if !visited[node] {
            dfs(node, nil)
        }
    }

    return cycles
}

// ToText はテキスト形式で出力
func (g *DependencyGraph) ToText() string {
    var sb strings.Builder
    sb.WriteString("Dependency Graph\n")
    sb.WriteString("================\n\n")

    for id, children := range g.edges {
        node := g.nodes[id]
        for _, childID := range children {
            child := g.nodes[childID]
            sb.WriteString(fmt.Sprintf("%s [%s] --> %s [%s]\n",
                node.Label, node.Status, child.Label, child.Status))
        }
    }

    return sb.String()
}

// ToDOT は Graphviz DOT形式で出力
func (g *DependencyGraph) ToDOT() string {
    var sb strings.Builder
    sb.WriteString("digraph G {\n")
    sb.WriteString("  rankdir=LR;\n")
    sb.WriteString("  node [shape=box];\n\n")

    // ノード定義
    for id, node := range g.nodes {
        color := g.statusColor(node.Status)
        sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\" color=\"%s\"];\n",
            id, node.Label, color))
    }

    sb.WriteString("\n")

    // エッジ定義
    for from, children := range g.edges {
        for _, to := range children {
            sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", from, to))
        }
    }

    sb.WriteString("}\n")
    return sb.String()
}

// ToMermaid は Mermaid形式で出力
func (g *DependencyGraph) ToMermaid() string {
    var sb strings.Builder
    sb.WriteString("```mermaid\n")
    sb.WriteString("graph TD\n")

    // ノード定義（ステータスに応じたスタイル）
    for id, node := range g.nodes {
        style := g.mermaidStyle(node.Status)
        sb.WriteString(fmt.Sprintf("  %s[\"%s\"]%s\n", id, node.Label, style))
    }

    sb.WriteString("\n")

    // エッジ定義
    for from, children := range g.edges {
        for _, to := range children {
            sb.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
        }
    }

    sb.WriteString("```\n")
    return sb.String()
}

func (g *DependencyGraph) statusColor(status string) string {
    switch status {
    case "completed":
        return "green"
    case "in_progress":
        return "blue"
    case "pending":
        return "gray"
    default:
        return "black"
    }
}

func (g *DependencyGraph) mermaidStyle(status string) string {
    switch status {
    case "completed":
        return ":::completed"
    case "in_progress":
        return ":::inprogress"
    default:
        return ""
    }
}

// Statistics はグラフの統計情報
type GraphStatistics struct {
    NodeCount     int
    EdgeCount     int
    AvgDeps       float64
    MaxDeps       int
    HasCycles     bool
    CycleCount    int
}

// GetStatistics は統計情報を計算
func (g *DependencyGraph) GetStatistics() GraphStatistics {
    nodeCount := len(g.nodes)
    edgeCount := 0
    maxDeps := 0

    for _, children := range g.edges {
        edgeCount += len(children)
        if len(children) > maxDeps {
            maxDeps = len(children)
        }
    }

    avgDeps := 0.0
    if nodeCount > 0 {
        avgDeps = float64(edgeCount) / float64(nodeCount)
    }

    cycles := g.DetectCycles()

    return GraphStatistics{
        NodeCount:  nodeCount,
        EdgeCount:  edgeCount,
        AvgDeps:    avgDeps,
        MaxDeps:    maxDeps,
        HasCycles:  len(cycles) > 0,
        CycleCount: len(cycles),
    }
}
```

## 7. レポートモジュールの実装（Phase 4）

### 7.1 internal/report/generator.go

```go
package report

import (
    "bytes"
    "context"
    "text/template"

    "github.com/biwakonbu/zeus/internal/analysis"
)

// ZeusConfig はレポート生成に必要な設定情報
type ZeusConfig struct {
    Project ProjectInfo
}

// ProjectInfo はプロジェクト情報
type ProjectInfo struct {
    ID          string
    Name        string
    Description string
    StartDate   string
}

// ProjectState はプロジェクト状態
type ProjectState struct {
    Health  string
    Summary SummaryStats
}

// SummaryStats はサマリー統計（Activity 統計）
type SummaryStats struct {
    TotalActivities int
    Completed       int
    InProgress      int
    Pending         int
}

// Generator はレポートを生成
type Generator struct {
    config   *ZeusConfig
    state    *ProjectState
    analysis *analysis.AnalysisResult
}

// NewGenerator は新しい Generator を作成
func NewGenerator(config *ZeusConfig, state *ProjectState, analysisResult *analysis.AnalysisResult) *Generator {
    return &Generator{
        config:   config,
        state:    state,
        analysis: analysisResult,
    }
}

// ReportData はテンプレートに渡すデータ
type ReportData struct {
    Timestamp     string
    ProjectName   string
    Health        string
    TaskStats     SummaryStats
    HasGraph      bool
    GraphMermaid  string
    Recommendations []string
}

// GenerateText は TEXT 形式でレポートを生成
func (g *Generator) GenerateText(ctx context.Context) (string, error) {
    if err := ctx.Err(); err != nil {
        return "", err
    }

    data := g.buildReportData()
    tmpl, err := template.New("text").Parse(TextTemplate)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

## 8. ダッシュボードモジュールの実装（Phase 5）

### 8.1 internal/dashboard/server.go

Go 標準ライブラリのみを使用した HTTP サーバー実装です。

```go
package dashboard

import (
    "context"
    "embed"
    "fmt"
    "io/fs"
    "net/http"
    "os/exec"
    "runtime"
    "time"
)

//go:embed static/*
var staticFiles embed.FS

// Server は HTTP サーバー
type Server struct {
    httpServer  *http.Server
    projectPath string
    port        int
}

// NewServer は新しいサーバーを作成
func NewServer(projectPath string, port int) *Server {
    return &Server{
        projectPath: projectPath,
        port:        port,
    }
}

// Start はサーバーを起動
func (s *Server) Start(openBrowser bool) error {
    mux := http.NewServeMux()

    // 静的ファイルの配信
    staticFS, err := fs.Sub(staticFiles, "static")
    if err != nil {
        return err
    }
    mux.Handle("/", http.FileServer(http.FS(staticFS)))

    // API ハンドラーの登録
    h := NewHandlers(s.projectPath)
    mux.HandleFunc("/api/status", h.HandleStatus)
    mux.HandleFunc("/api/activities", h.HandleActivities)
    mux.HandleFunc("/api/graph", h.HandleGraph)
    mux.HandleFunc("/api/predict", h.HandlePredict)

    // ローカルホストのみにバインド（セキュリティ対策）
    addr := fmt.Sprintf("127.0.0.1:%d", s.port)

    s.httpServer = &http.Server{
        Addr:         addr,
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    if openBrowser {
        go func() {
            time.Sleep(500 * time.Millisecond)
            s.openBrowser(fmt.Sprintf("http://localhost:%d", s.port))
        }()
    }

    fmt.Printf("Dashboard running at http://localhost:%d\n", s.port)
    fmt.Println("Press Ctrl+C to stop")

    return s.httpServer.ListenAndServe()
}

// Stop はサーバーを停止
func (s *Server) Stop(ctx context.Context) error {
    if s.httpServer != nil {
        return s.httpServer.Shutdown(ctx)
    }
    return nil
}

func (s *Server) openBrowser(url string) error {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "darwin":
        cmd = "open"
        args = []string{url}
    case "linux":
        cmd = "xdg-open"
        args = []string{url}
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start", url}
    default:
        return fmt.Errorf("unsupported platform")
    }

    return exec.Command(cmd, args...).Start()
}
```

### 8.2 internal/dashboard/handlers.go

```go
package dashboard

import (
    "encoding/json"
    "net/http"

    "github.com/biwakonbu/zeus/internal/analysis"
    "github.com/biwakonbu/zeus/internal/core"
)

// Handlers は API ハンドラー
type Handlers struct {
    projectPath string
}

// NewHandlers は新しいハンドラーを作成
func NewHandlers(projectPath string) *Handlers {
    return &Handlers{projectPath: projectPath}
}

// StatusResponse はステータス API のレスポンス
type StatusResponse struct {
    ProjectName     string  `json:"project_name"`
    Description     string  `json:"description"`
    Progress        float64 `json:"progress"`
    Health          string  `json:"health"`
    TotalActivities int     `json:"total_activities"`
    Completed       int     `json:"completed"`
    InProgress      int     `json:"in_progress"`
    Pending         int     `json:"pending"`
}

// HandleStatus はプロジェクト状態を返す
func (h *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
    zeus := core.New(h.projectPath)
    status, err := zeus.GetStatus(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    resp := StatusResponse{
        ProjectName:     status.ProjectName,
        Description:     status.Description,
        Progress:        status.Progress,
        Health:          status.Health,
        TotalActivities: status.TotalActivities,
        Completed:       status.Completed,
        InProgress:      status.InProgress,
        Pending:         status.Pending,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// HandleActivities はアクティビティ一覧を返す
func (h *Handlers) HandleActivities(w http.ResponseWriter, r *http.Request) {
    zeus := core.New(h.projectPath)
    activities, err := zeus.ListActivities(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(activities)
}

// HandleGraph は依存関係グラフを返す（Mermaid形式）
func (h *Handlers) HandleGraph(w http.ResponseWriter, r *http.Request) {
    zeus := core.New(h.projectPath)
    graph, err := zeus.BuildGraph(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(graph.ToMermaid()))
}

```

### 8.3 internal/dashboard/static/index.html

```html
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zeus Dashboard</title>
    <link rel="stylesheet" href="styles.css">
    <script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
</head>
<body>
    <header>
        <h1>Zeus Dashboard</h1>
        <span id="last-updated"></span>
    </header>

    <main>
        <section id="overview" class="card">
            <h2>Project Overview</h2>
            <div id="project-info"></div>
        </section>

        <section id="stats" class="card">
            <h2>Activity Statistics</h2>
            <div id="activity-stats"></div>
        </section>

        <section id="activities" class="card">
            <h2>Activities</h2>
            <table id="activity-table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Title</th>
                        <th>Status</th>
                        <th>Priority</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </section>

        <section id="graph" class="card">
            <h2>Dependency Graph</h2>
            <div id="mermaid-graph"></div>
        </section>

        <section id="prediction" class="card">
            <h2>Predictions</h2>
            <div id="prediction-info"></div>
        </section>
    </main>

    <script src="app.js"></script>
</body>
</html>
```

### 8.4 internal/dashboard/static/app.js

```javascript
// 自動更新間隔（ミリ秒）
const REFRESH_INTERVAL = 5000;

// 初期化
document.addEventListener('DOMContentLoaded', () => {
    mermaid.initialize({ startOnLoad: false, theme: 'default' });
    refresh();
    setInterval(refresh, REFRESH_INTERVAL);
});

// 全データを更新
async function refresh() {
    try {
        await Promise.all([
            fetchStatus(),
            fetchActivities(),
            fetchGraph()
        ]);
        updateLastUpdated();
    } catch (error) {
        console.error('Failed to refresh:', error);
    }
}

// ステータスを取得
async function fetchStatus() {
    const response = await fetch('/api/status');
    const data = await response.json();

    document.getElementById('project-info').innerHTML = `
        <p><strong>Name:</strong> ${data.project_name}</p>
        <p><strong>Description:</strong> ${data.description}</p>
        <p><strong>Progress:</strong> ${data.progress.toFixed(1)}%</p>
        <p><strong>Health:</strong> <span class="health-${data.health}">${data.health}</span></p>
    `;

    document.getElementById('task-stats').innerHTML = `
        <div class="stat-item">
            <span class="stat-value completed">${data.completed}</span>
            <span class="stat-label">Completed</span>
        </div>
        <div class="stat-item">
            <span class="stat-value in-progress">${data.in_progress}</span>
            <span class="stat-label">In Progress</span>
        </div>
        <div class="stat-item">
            <span class="stat-value pending">${data.pending}</span>
            <span class="stat-label">Pending</span>
        </div>
    `;
}

// アクティビティ一覧を取得
async function fetchActivities() {
    const response = await fetch('/api/activities');
    const activities = await response.json();

    const tbody = document.querySelector('#activity-table tbody');
    tbody.innerHTML = activities.map(activity => `
        <tr class="status-${activity.status}">
            <td>${activity.id}</td>
            <td>${activity.title}</td>
            <td><span class="badge ${activity.status}">${activity.status}</span></td>
            <td>${activity.priority || '-'}</td>
        </tr>
    `).join('');
}

// グラフを取得
async function fetchGraph() {
    const response = await fetch('/api/graph');
    const mermaidCode = await response.text();

    const container = document.getElementById('mermaid-graph');
    container.innerHTML = '';

    const id = 'graph-' + Date.now();
    const { svg } = await mermaid.render(id, mermaidCode);
    container.innerHTML = svg;
}

// 最終更新時刻を更新
function updateLastUpdated() {
    const now = new Date().toLocaleTimeString();
    document.getElementById('last-updated').textContent = `Last updated: ${now}`;
}
```

## 9. ビルドと実行

### 9.1 Makefile

```makefile
.PHONY: build clean test install

BINARY_NAME=zeus
VERSION=1.0.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# 開発用
dev:
	go run . $(ARGS)
```

### 9.2 ビルド確認

```bash
# ビルド
make build

# 実行確認
./zeus --help
./zeus init
./zeus status
./zeus graph
./zeus predict
./zeus report
./zeus dashboard
```

## 10. 実装優先順位

### Phase 1: MVP（最小実行可能プロダクト）- 完了

| コマンド | 優先度 | 説明 |
|---------|--------|------|
| `zeus init` | 高 | プロジェクト初期化（.zeus/ + .claude/ 生成） |
| `zeus status` | 高 | 状態表示 |
| `zeus add activity` | 高 | アクティビティ追加 |
| `zeus list` | 高 | 一覧表示 |
| `zeus doctor` | 中 | 診断 |
| `zeus fix` | 中 | 修復 |

### Phase 2: 承認フロー - 完了

| コマンド | 説明 |
|---------|------|
| `zeus pending` | 承認待ち一覧 |
| `zeus approve` | 承認 |
| `zeus reject` | 却下 |

### Phase 3: AI 統合 - 完了

| コマンド | 説明 |
|---------|------|
| `zeus suggest` | AI 提案 |
| `zeus apply` | 提案適用 |
| `zeus explain` | AI 解説 |

### Phase 4: 分析機能 - 完了

| コマンド | 説明 |
|---------|------|
| `zeus graph` | 依存関係グラフ（text/dot/mermaid） |
| `zeus report` | レポート生成（text/html/markdown） |

### Phase 5: ダッシュボード - 完了

| コマンド | 説明 |
|---------|------|
| `zeus dashboard` | Web ダッシュボード起動 |

### Phase 6: 依存関係強化 - 完了

| 機能 | 説明 |
|------|------|
| 階層構造 | Activity の階層構造管理 |
| 影響範囲可視化 | downstream/upstream 依存の表示 |

### Phase 7: 外部連携 - 計画中

| 機能 | 説明 |
|------|------|
| Git 統合 | コミット履歴との連携 |
| 通知 | Slack/Email 通知 |
| 認証 | ダッシュボード認証 |

## 11. テスト

### 11.1 分析モジュールのテスト

```go
// internal/analysis/graph_test.go
package analysis

import (
    "testing"
)

func TestDependencyGraph_DetectCycles(t *testing.T) {
    g := NewDependencyGraph()
    g.AddNode("A", "Activity A", "pending")
    g.AddNode("B", "Activity B", "pending")
    g.AddNode("C", "Activity C", "pending")

    // A -> B -> C -> A（循環）
    g.AddEdge("A", "B")
    g.AddEdge("B", "C")
    g.AddEdge("C", "A")

    cycles := g.DetectCycles()
    if len(cycles) == 0 {
        t.Error("Expected cycle to be detected")
    }
}

func TestDependencyGraph_ToMermaid(t *testing.T) {
    g := NewDependencyGraph()
    g.AddNode("A", "Activity A", "completed")
    g.AddNode("B", "Activity B", "in_progress")
    g.AddEdge("A", "B")

    output := g.ToMermaid()
    if output == "" {
        t.Error("Expected non-empty mermaid output")
    }
}
```

### 11.2 ダッシュボードのテスト

```go
// internal/dashboard/dashboard_test.go
package dashboard

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandleStatus(t *testing.T) {
    // テスト用の一時ディレクトリでZeusを初期化
    tmpDir := t.TempDir()
    // ... 初期化コード

    h := NewHandlers(tmpDir)

    req := httptest.NewRequest("GET", "/api/status", nil)
    rec := httptest.NewRecorder()

    h.HandleStatus(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

---

*Zeus Implementation Guide (Go版) v1.2*
*作成日: 2026-01-14*
*更新日: 2026-01-17（--level フラグ削除、Phase 6 追加）*
*再分類日: 2026-02-07（archive移行）*
