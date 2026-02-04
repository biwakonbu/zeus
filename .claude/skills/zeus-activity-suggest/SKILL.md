---
description: 現在の状態に基づいて Activity（作業単位）を提案するスキル
---

# zeus-activity-suggest

現在の状態に基づいて Activity（作業単位）や改善案を提案するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の 10概念モデル全体を分析し、次に取り組むべき Activity や改善案を提案します。

## 実行方法

```bash
# 提案生成（デフォルト5件）
zeus suggest

# 件数指定
zeus suggest --limit 10

# 影響度フィルタ
zeus suggest --impact high

# 組み合わせ
zeus suggest --limit 5 --impact high

# 既存提案を上書き
zeus suggest --force
```

## 提案タイプ

| タイプ | 説明 |
|--------|------|
| `new_activity` | 新規 Activity の追加提案 |
| `priority_change` | Activity 優先度の変更提案 |
| `dependency` | 依存関係の追加・修正提案 |
| `risk_mitigation` | リスク軽減策の提案 |

## 出力例

```yaml
suggestions:
  - id: sugg-abc12345
    type: risk_mitigation
    description: "3件のブロックされた Activity を解決する必要があります"
    rationale: "ブロックされた Activity はプロジェクト全体の進行を妨げます"
    impact: high
    status: pending
    created_at: "2026-01-19T10:00:00Z"
```

## 提案の適用

```bash
# 個別適用
zeus apply <suggestion-id>

# 全提案適用
zeus apply --all

# ドライラン（実行せずに確認）
zeus apply --all --dry-run
```

## 10概念モデルとの連携

### Vision/Objective 関連
- Vision の success_criteria 達成度チェック
- Objective の進捗に基づく優先度調整
- 期限切れ Objective の警告

### Deliverable 関連
- 未着手 Deliverable の着手提案
- Objective との紐付けチェック

### Problem 関連
- 未解決 Problem への対応 Activity 提案
- severity: high/critical の Problem 優先対応

### Risk 関連
- 高スコア Risk の軽減策提案
- Assumption 検証によるリスク軽減

### Quality 関連
- 品質基準未達の Deliverable 警告

### Constraint 関連
- 制約違反の可能性警告
- non-negotiable 制約のチェック

## WBS階層を考慮した提案

- 親 Activity/Objective の完了度に基づく子の優先度調整
- 親が未設定の Activity に対する整理提案

## タイムライン最適化

- クリティカルパス上の Activity の優先度向上
- 期限切れ Activity/Objective の警告
- 依存関係のボトルネック特定

## 提案アルゴリズム

1. 現在のプロジェクト状態を取得（`zeus status`）
2. ブロックされた Activity を検出
3. 高リスク項目（Risk, Problem）を分析
4. WBS階層と依存関係を考慮
5. 優先度に基づいて提案を生成

## 保存先

提案は `.zeus/suggestions/active.yaml` に保存されます。

## 関連スキル

- zeus-project-scan - プロジェクト状態のスキャン
- zeus-risk-analysis - リスク分析
