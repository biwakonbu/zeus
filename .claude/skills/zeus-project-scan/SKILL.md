---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の 10概念モデル全体（Vision, Objective, Deliverable, Task, Consideration, Decision, Problem, Risk, Assumption, Constraint, Quality）および Actor/UseCase/Subsystem/Activity を分析します。

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

  # 10概念モデル件数 + Actor/UseCase/Subsystem/Activity
  entities:
    objectives: 7
    deliverables: 4
    tasks: 3
    considerations: 2
    decisions: 1
    problems: 3
    risks: 3
    assumptions: 3
    constraints: 3
    quality: 2
    actors: 5
    usecases: 8
    subsystems: 3
    activities: 3

  # 参照整合性
  integrity:
    status: "healthy|warning|error"
    issues:
      - type: "missing_reference"
        source: "del-001"
        target: "obj-999"
        message: "referenced objective not found"
      - type: "circular_reference"
        entities: ["obj-001", "obj-002", "obj-003", "obj-001"]
        message: "circular parent reference detected"

  # 依存関係グラフ
  graph:
    cycles: []
    isolated: []
    max_depth: 3

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
```

## 基本コマンド

```bash
# プロジェクト全体の状態確認
zeus status

# 参照整合性チェック
zeus doctor

# 問題の自動修復（ドライラン）
zeus fix --dry-run
```

## 10概念モデル一覧取得

```bash
# Vision（単一ファイル）
cat .zeus/vision.yaml

# Objective 一覧
zeus list objectives

# Deliverable 一覧
zeus list deliverables

# Task 一覧
zeus list tasks

# Consideration 一覧（検討事項）
zeus list considerations

# Decision 一覧（意思決定 - イミュータブル）
zeus list decisions

# Problem 一覧
zeus list problems

# Risk 一覧
zeus list risks

# Assumption 一覧
zeus list assumptions

# Constraint 一覧（単一ファイル）
zeus list constraints

# Quality 一覧
zeus list quality

# Actor 一覧
zeus list actors

# UseCase 一覧
zeus list usecases

# Subsystem 一覧
zeus list subsystems

# Activity 一覧
zeus list activities
```

## 分析・可視化

```bash
# 依存関係グラフ（複数形式）
zeus graph --format text
zeus graph --format mermaid
zeus graph --format dot -o graph.dot

# 予測分析
zeus predict completion   # 完了日予測
zeus predict risk         # リスク分析
zeus predict velocity     # ベロシティ分析
zeus predict all          # 全予測

# レポート生成
zeus report --format markdown -o report.md

# Web ダッシュボード
zeus dashboard

# UML ユースケース図
zeus uml show usecase
zeus uml show usecase --format mermaid
zeus uml show usecase --boundary "システム名"

# UML アクティビティ図
zeus uml show activity
zeus uml show activity --id act-001
```

## 10概念モデル詳細

### Phase 1 概念（コア3概念）

| 概念 | 説明 | ファイル |
|------|------|----------|
| Vision | プロジェクトの目指す姿（単一） | `.zeus/vision.yaml` |
| Objective | 達成目標（階層構造可） | `.zeus/objectives/obj-NNN.yaml` |
| Deliverable | 成果物定義 | `.zeus/deliverables/del-NNN.yaml` |

### Phase 2 概念（管理5概念）

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Consideration | 検討事項 | `.zeus/considerations/con-NNN.yaml` | 複数オプション記録 |
| Decision | 意思決定 | `.zeus/decisions/dec-NNN.yaml` | **イミュータブル** |
| Problem | 問題報告 | `.zeus/problems/prob-NNN.yaml` | severity レベル |
| Risk | リスク管理 | `.zeus/risks/risk-NNN.yaml` | スコア自動計算 |
| Assumption | 前提条件 | `.zeus/assumptions/assum-NNN.yaml` | 検証ステータス |

### Phase 3 概念（品質2概念）

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Constraint | 制約条件 | `.zeus/constraints.yaml` | グローバル単一ファイル |
| Quality | 品質基準 | `.zeus/quality/qual-NNN.yaml` | メトリクス・ゲート管理 |

### UML 拡張（Actor/UseCase/Subsystem/Activity）

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Actor | アクター定義 | `.zeus/actors.yaml` | 単一ファイル |
| UseCase | ユースケース定義 | `.zeus/usecases/uc-NNN.yaml` | Objective 参照必須 |
| Subsystem | サブシステム定義 | `.zeus/subsystems.yaml` | 単一ファイル、UseCase グルーピング |
| Activity | アクティビティ図 | `.zeus/activities/act-NNN.yaml` | UseCase 参照任意 |

## 参照整合性チェック

`zeus doctor` で以下の整合性をチェック:

### 必須参照（エラー）
- **Deliverable → Objective**: `objective_id` が必須
- **Decision → Consideration**: `consideration_id` が必須
- **Quality → Deliverable**: `deliverable_id` が必須
- **UseCase → Objective**: `objective_id` が必須

### 任意参照（参照先が存在しない場合はエラー/警告）
- **Objective → Objective**: 親 `parent_id`（循環参照チェックあり）
- **Consideration → Objective/Deliverable/Decision**: 任意の紐付け
- **Problem → Objective/Deliverable**: 関連エンティティ
- **Risk → Objective/Deliverable**: 関連エンティティ
- **Assumption → Objective/Deliverable**: 関連エンティティ
- **UseCase → Actor**: `actors[].actor_id` の参照先確認
- **UseCase → UseCase**: `relations[].target_id` の参照先確認
- **UseCase → Subsystem**: `subsystem_id` の参照先確認（警告レベル）

### 循環参照検出
- Objective の親子階層で循環を検出
- UseCase の relations で循環を検出

## ダッシュボード API

```bash
# ダッシュボード起動後（デフォルト: localhost:8080）
curl http://localhost:8080/api/status
curl http://localhost:8080/api/tasks
curl http://localhost:8080/api/graph
curl http://localhost:8080/api/predict
curl http://localhost:8080/api/wbs
curl http://localhost:8080/api/timeline
curl http://localhost:8080/api/actors
curl http://localhost:8080/api/usecases
curl http://localhost:8080/api/subsystems
curl http://localhost:8080/api/uml/usecase
curl http://localhost:8080/api/activities
curl http://localhost:8080/api/uml/activity?id=act-001
```

## 関連スキル

- zeus-task-suggest - 概念間の関連に基づくタスク提案
- zeus-risk-analysis - Risk/Problem/Assumption の詳細分析
