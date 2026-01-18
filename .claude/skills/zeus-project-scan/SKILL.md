---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクトの 10概念モデル全体（Vision, Objectives, Deliverables, Tasks, Problems, Risks, Assumptions, Constraints, Quality, Considerations/Decisions）を分析します。

## 入力

なし（カレントディレクトリの .zeus/ を参照）

## 出力

```yaml
project:
  name: "プロジェクト名"
  health: "good|fair|poor"

  # Vision（単一）
  vision:
    title: "ビジョン名"
    statement: "ビジョンステートメント"
    success_criteria: ["基準1", "基準2"]

  # 10概念モデル件数
  entities:
    objectives: 7
    deliverables: 4
    tasks: 3
    problems: 3
    risks: 3
    assumptions: 3
    constraints: 3
    quality: 2
    considerations: 0
    decisions: 0

  # 従来のタスク管理
  tasks:
    total: 10
    completed: 3
    in_progress: 2
    pending: 5

  # 依存関係グラフ
  graph:
    cycles: []            # 循環参照リスト
    isolated: []          # 孤立エンティティリスト
    max_depth: 3          # 依存関係の最大深度

  # 予測分析
  prediction:
    estimated_completion: "2026-03-31"
    risk_level: "medium"
    velocity_trend: "stable|improving|declining"

  # WBS階層
  wbs:
    max_depth: 3
    total_nodes: 15
    orphan_tasks: []

  # タイムライン
  timeline:
    project_start: "2026-01-01"
    project_end: "2026-03-31"
    critical_path_length: 5
    overdue_tasks: []

  # 参照整合性
  integrity:
    status: "healthy|warning|error"
    issues: []
```

## 使用方法

1. `zeus status` コマンドで基本情報取得（Vision, Objectives, Deliverables 含む）
2. `zeus doctor` で参照整合性チェック
3. `zeus list <entity>` で各エンティティ一覧
4. `zeus graph` で依存関係グラフ確認
5. `zeus predict all` で予測分析実行
6. `zeus report` でレポート生成
7. `zeus dashboard` で可視化（推奨）

## コマンド実行例

```bash
# 基本状態の確認（10概念モデル対応）
zeus status

# 参照整合性チェック
zeus doctor

# 各エンティティ一覧
zeus list objectives
zeus list deliverables
zeus list problems
zeus list risks
zeus list assumptions
zeus list constraints
zeus list quality
zeus list considerations
zeus list decisions

# 依存関係グラフ（Mermaid形式）
zeus graph --format mermaid

# 全予測分析
zeus predict all

# レポート生成
zeus report --format markdown

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

## 10概念モデル

| 概念 | 説明 | ファイル |
|------|------|----------|
| Vision | プロジェクトの目指す姿（単一） | `.zeus/vision.yaml` |
| Objective | 達成目標（階層構造可） | `.zeus/objectives/` |
| Deliverable | 成果物定義 | `.zeus/deliverables/` |
| Task | 実行タスク | `.zeus/tasks/` |
| Consideration | 検討事項 | `.zeus/considerations/` |
| Decision | 意思決定（イミュータブル） | `.zeus/decisions/` |
| Problem | 問題報告 | `.zeus/problems/` |
| Risk | リスク管理 | `.zeus/risks/` |
| Assumption | 前提条件 | `.zeus/assumptions/` |
| Constraint | 制約条件 | `.zeus/constraints.yaml` |
| Quality | 品質基準 | `.zeus/quality/` |

## 関連

- zeus-task-suggest
- zeus-risk-analysis
