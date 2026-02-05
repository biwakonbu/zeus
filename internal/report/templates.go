// Package report はプロジェクトレポートを生成する。
// TEXT, HTML, Markdown 形式のレポート出力に対応。
package report

// TextTemplate はテキスト形式のレポートテンプレート
const TextTemplate = `Zeus Project Report
{{.Separator}}
Generated: {{.Timestamp}}

PROJECT: {{.ProjectName}}
Health:  {{.Health}}

TASK SUMMARY
------------
  Total:       {{.TaskStats.TotalActivities}}
  Completed:   {{.TaskStats.Completed}}
  In Progress: {{.TaskStats.InProgress}}
  Pending:     {{.TaskStats.Pending}}

{{if .Recommendations}}
RECOMMENDATIONS
---------------
{{range .Recommendations}}  - {{.}}
{{end}}{{end}}
{{.Separator}}
`

// HTMLTemplate は HTML 形式のレポートテンプレート
const HTMLTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zeus Project Report - {{.ProjectName}}</title>
    <style>
        :root {
            --primary-color: #2196F3;
            --success-color: #4CAF50;
            --warning-color: #FF9800;
            --danger-color: #F44336;
            --background-color: #f5f5f5;
            --card-background: #ffffff;
            --text-color: #333333;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--background-color);
            color: var(--text-color);
            line-height: 1.6;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header {
            background: linear-gradient(135deg, var(--primary-color), #1976D2);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 20px;
        }
        .header h1 { font-size: 2rem; margin-bottom: 10px; }
        .header .meta { opacity: 0.9; font-size: 0.9rem; }
        .card {
            background: var(--card-background);
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h2 {
            color: var(--primary-color);
            border-bottom: 2px solid var(--primary-color);
            padding-bottom: 10px;
            margin-bottom: 15px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
        }
        .stat-item {
            text-align: center;
            padding: 15px;
            background: var(--background-color);
            border-radius: 8px;
        }
        .stat-item .value { font-size: 2rem; font-weight: bold; color: var(--primary-color); }
        .stat-item .label { font-size: 0.9rem; color: #666; }
        .health-good { color: var(--success-color); }
        .health-fair { color: var(--warning-color); }
        .health-poor { color: var(--danger-color); }
        .progress-bar {
            background: #e0e0e0;
            border-radius: 10px;
            overflow: hidden;
            height: 20px;
            margin: 10px 0;
        }
        .progress-bar .fill {
            height: 100%;
            background: var(--success-color);
            transition: width 0.3s;
        }
        .recommendations li {
            padding: 8px 0;
            border-bottom: 1px solid #eee;
        }
        .recommendations li:last-child { border-bottom: none; }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 0.85rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.ProjectName}}</h1>
            <div class="meta">Generated: {{.Timestamp}} | Health: <span class="health-{{.HealthClass}}">{{.Health}}</span></div>
        </div>

        <div class="card">
            <h2>Task Summary</h2>
            <div class="stats-grid">
                <div class="stat-item">
                    <div class="value">{{.TaskStats.TotalActivities}}</div>
                    <div class="label">Total Tasks</div>
                </div>
                <div class="stat-item">
                    <div class="value" style="color: var(--success-color);">{{.TaskStats.Completed}}</div>
                    <div class="label">Completed</div>
                </div>
                <div class="stat-item">
                    <div class="value" style="color: var(--warning-color);">{{.TaskStats.InProgress}}</div>
                    <div class="label">In Progress</div>
                </div>
                <div class="stat-item">
                    <div class="value">{{.TaskStats.Pending}}</div>
                    <div class="label">Pending</div>
                </div>
            </div>
            {{if gt .TaskStats.TotalActivities 0}}
            <div class="progress-bar">
                <div class="fill" style="width: {{.CompletionPercent}}%;"></div>
            </div>
            <div style="text-align: center; color: #666;">{{.CompletionPercent}}% Complete</div>
            {{end}}
        </div>

        {{if .Recommendations}}
        <div class="card">
            <h2>Recommendations</h2>
            <ul class="recommendations">
                {{range .Recommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}

        <div class="footer">
            Generated by Zeus CLI | <a href="https://github.com/biwakonbu/zeus">Zeus Project</a>
        </div>
    </div>
</body>
</html>
`

// MarkdownTemplate は Markdown 形式のレポートテンプレート
const MarkdownTemplate = `# Zeus Project Report

**Project:** {{.ProjectName}}
**Generated:** {{.Timestamp}}
**Health:** {{.Health}}

---

## Task Summary

| Status | Count | Percentage |
|--------|-------|------------|
| Completed | {{.TaskStats.Completed}} | {{.CompletedPercent}}% |
| In Progress | {{.TaskStats.InProgress}} | {{.InProgressPercent}}% |
| Pending | {{.TaskStats.Pending}} | {{.PendingPercent}}% |
| **Total** | **{{.TaskStats.TotalActivities}}** | **100%** |

{{if .HasGraph}}
## Dependency Graph

{{.GraphMermaid}}
{{end}}

{{if .Recommendations}}
## Recommendations

{{range .Recommendations}}- {{.}}
{{end}}
{{end}}

---

*Generated by [Zeus CLI](https://github.com/biwakonbu/zeus)*
`
