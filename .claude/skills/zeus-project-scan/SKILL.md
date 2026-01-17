---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の構造、タスク、進捗、依存関係、WBS階層、タイムラインを分析します。

## 入力

なし（カレントディレクトリの .zeus/ を参照）

## 出力

```yaml
project:
  name: "プロジェクト名"
  health: "good|fair|poor"
  tasks:
    total: 10
    completed: 3
    in_progress: 2
    pending: 5
  graph:
    cycles: []            # 循環参照リスト
    isolated: []          # 孤立タスクリスト
    max_depth: 3          # 依存関係の最大深度
  prediction:
    estimated_completion: "2026-03-31"
    risk_level: "medium"
    velocity_trend: "stable|improving|declining"
  wbs:
    max_depth: 3          # WBS階層の最大深度
    total_nodes: 15       # WBSノード総数
    orphan_tasks: []      # 親が見つからないタスク
  timeline:
    project_start: "2026-01-01"
    project_end: "2026-03-31"
    critical_path_length: 5     # クリティカルパス上のタスク数
    overdue_tasks: []           # 期限超過タスク
  risks: []
```

## 使用方法

1. `zeus status` コマンドで基本情報取得
2. `zeus graph` で依存関係グラフ確認
3. `zeus predict all` で予測分析実行
4. `zeus dashboard` で可視化（推奨）
5. 改善提案を生成

## コマンド実行例

```bash
# 基本状態の確認
zeus status

# 依存関係グラフ（Mermaid形式）
zeus graph --format mermaid

# 全予測分析
zeus predict all

# Webダッシュボードで可視化
zeus dashboard
```

## ダッシュボードAPI

スキャン結果をプログラムで取得する場合:

```bash
# ダッシュボード起動後
curl http://localhost:8080/api/status
curl http://localhost:8080/api/tasks
curl http://localhost:8080/api/graph
curl http://localhost:8080/api/predict
curl http://localhost:8080/api/wbs
curl http://localhost:8080/api/timeline
```

## 関連

- zeus-task-suggest
- zeus-risk-analysis
