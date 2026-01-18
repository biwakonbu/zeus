---
description: 現在の状態に基づいてタスクを提案するスキル
---

# zeus-task-suggest

現在の状態に基づいてタスクや改善案を提案するスキル。

## 概要

Zeus プロジェクトの 10概念モデル全体を分析し、次に取り組むべきタスクや改善案を提案します。

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
    type: "new_task|priority_change|dependency|schedule|objective|deliverable|problem|risk"
    description: "提案の説明"
    rationale: "理由"
    impact: "high|medium|low"
  - id: suggestion-2
    type: "problem"
    description: "未解決の Problem に対応するタスクを追加"
    rationale: "prob-001 が high severity で未対応"
    impact: "high"
```

## 10概念モデル対応

### Vision/Objective 関連
- Vision の success_criteria 達成度チェック
- Objective の進捗に基づく優先度調整
- 期限切れ Objective の警告

### Deliverable 関連
- 未着手 Deliverable の着手提案
- 受入基準未達の Deliverable 警告
- Objective との紐付けチェック

### Problem 関連
- 未解決 Problem への対応タスク提案
- severity: high の Problem 優先対応
- 関連 Objective/Deliverable への影響分析

### Risk 関連
- 高スコア Risk の軽減策提案
- 未対応 Risk の監視強化提案
- Assumption 検証によるリスク軽減

### Quality 関連
- 品質基準未達の Deliverable 警告
- メトリクス測定タスク提案

### Constraint 関連
- 制約違反の可能性警告
- non-negotiable 制約のチェック

## WBS階層を考慮した提案
- 親タスク/Objective の完了度に基づく子の優先度調整
- 階層のバランスチェック（深すぎる/浅すぎる階層の検出）
- 親が未設定のタスクに対する整理提案

## タイムライン最適化
- クリティカルパス上のタスクの優先度向上
- 期限切れタスク/Objective の警告
- スラックが少ないタスクの注意喚起
- 依存関係のボトルネック特定

## 提案の適用

```bash
# 個別適用
zeus apply <suggestion-id>

# 全提案適用（ドライラン可能）
zeus apply --all [--dry-run]
```

## アルゴリズム

1. 現在の 10概念モデル全体を分析
2. 参照整合性を確認（`zeus doctor`）
3. Vision/Objective の達成度を評価
4. Problem/Risk の未対応項目を抽出
5. WBS階層と依存関係を考慮
6. クリティカルパスを計算
7. 優先度に基づいて提案を生成

## 関連スキル

- zeus-project-scan - プロジェクト状態のスキャン
- zeus-risk-analysis - リスク分析
