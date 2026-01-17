---
description: 現在の状態に基づいてタスクを提案するスキル
---

# zeus-task-suggest

現在の状態に基づいてタスクを提案するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の状態を分析し、次に取り組むべきタスクや改善案を提案します。

## 実行方法

```bash
# 提案生成
zeus suggest [--limit N] [--impact high|medium|low]

# 例: 高影響の提案を5件表示
zeus suggest --limit 5 --impact high
```

## 出力

```yaml
suggestions:
  - id: suggestion-1
    type: "new_task|priority_change|dependency|schedule"
    description: "提案の説明"
    rationale: "理由"
    impact: "high|medium|low"
  - id: suggestion-2
    type: "schedule"
    description: "クリティカルパス上のタスクに遅延リスク"
    rationale: "タスクXのスラックが0で、依存タスクに影響"
    impact: "high"
```

## Phase 6 対応

### WBS階層を考慮した提案
- 親タスクの完了度に基づく子タスクの優先度調整
- 階層のバランスチェック（深すぎる/浅すぎる階層の検出）
- 親が未設定のタスクに対する整理提案

### タイムライン最適化
- クリティカルパス上のタスクの優先度向上
- 期限切れタスクの警告
- スラックが少ないタスクの注意喚起
- 依存関係のボトルネック特定

### 進捗整合性
- 親タスクと子タスクの進捗率の不整合検出
- 長期間更新されていないタスクの警告

## 提案の適用

```bash
# 個別適用
zeus apply <suggestion-id>

# 全提案適用（ドライラン可能）
zeus apply --all [--dry-run]
```

## アルゴリズム

1. 現在のタスク状態を分析
2. WBS階層と依存関係を考慮
3. クリティカルパスを計算
4. 目標との差分を計算
5. 優先度に基づいて提案を生成

## 関連スキル

- zeus-project-scan - プロジェクト状態のスキャン
- zeus-risk-analysis - リスク分析
