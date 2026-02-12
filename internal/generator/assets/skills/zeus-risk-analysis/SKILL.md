---
description: プロジェクトのリスクを分析し、対策を提案するスキル
use_when: |
  Use when user asks about risks, problems, or assumptions.
  Also use when user says "リスク", "問題", "前提条件", "risk", "problem", "assumption".
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## Zeus の4層階層モデル

Zeus は以下の4層でプロジェクトを構造化する:

| 層 | エンティティ | 役割 | リスクの観点 |
|---|---|---|---|
| ゴール層 | Vision | 実現するべきゴール | 戦略リスク（方向性の誤り） |
| 目標層 | Objective | 測定可能な成果目標 | 目標リスク（達成不能・期限超過） |
| 抽象層 | UseCase | 本質的な求め | 要件リスク（スコープ不明確・Actor 不足） |
| 具体層 | Activity | 実現手段 | 実装リスク（技術的困難・依存関係） |

Risk/Problem/Assumption は `objective_id` で Objective に紐付けられ、
関連する UseCase/Activity への影響を階層的に分析できる。

## 概要

Zeus プロジェクト（{{.ProjectName}}）の Risk, Problem, Assumption エンティティを活用してリスク要因を特定し、対策を提案します。

## 実行方法

```bash
# Risk エンティティ一覧
zeus list risks

# 参照整合性チェック
zeus doctor

# 依存関係グラフ
zeus graph --format mermaid
```

## Risk エンティティ

### 追加コマンド

```bash
zeus add risk "リスク名" \
  --probability medium \      # 発生確率: high, medium, low
  --impact high \             # 影響度: critical, high, medium, low
  --objective obj-001 \       # 関連 Objective（任意）
  -d "リスクの詳細説明"
```

### フィールド

| フィールド | 説明 | 必須 |
|-----------|------|------|
| probability | 発生確率（high/medium/low） | 任意 |
| impact | 影響度（critical/high/medium/low） | 任意 |
| score | 総合スコア（自動計算） | - |
| objective_id | 関連 Objective | 任意 |
| status | 状態（identified/mitigating/mitigated/accepted） | - |
| mitigation | 軽減策 | 任意 |

### スコア自動計算

probability と impact の組み合わせでスコアが自動計算されます:

| 確率 | critical | high | medium | low |
|------|----------|------|--------|-----|
| high | critical | critical | high | medium |
| medium | critical | high | medium | low |
| low | high | medium | low | low |

## Problem エンティティ

### 追加コマンド

```bash
zeus add problem "問題名" \
  --severity high \           # 深刻度: critical, high, medium, low
  --objective obj-001 \       # 関連 Objective
  -d "問題の詳細"
```

### Problem → Risk 連携
- 未解決の Problem はリスク要因
- severity: high/critical の Problem は高リスク
- Problem 放置期間によるリスク増大

## Assumption エンティティ

### 追加コマンド

```bash
zeus add assumption "前提条件" \
  --objective obj-001 \
  -d "前提条件の説明"
```

### Assumption → Risk 連携
- 未検証の Assumption はリスク要因
- Assumption が崩れた場合の影響分析
- 検証済み Assumption によるリスク軽減

## リスクカテゴリ（層別）

### ゴール層リスク（Vision）
- **戦略リスク**: Vision の方向性が市場・ユーザーのニーズと乖離
- **success_criteria 未達リスク**: 成功基準が曖昧・測定不能

### 目標層リスク（Objective）
- **目標リスク**: Objective が達成不能・期限超過
- **バランスリスク**: Objective 間の優先度・リソース配分の偏り
- **カバレッジリスク**: Objective に UseCase が紐付いていない（目標が空洞化）

### 抽象層リスク（UseCase）
- **要件リスク**: スコープ不明確・Actor 不足
- **関係リスク**: include/extend/generalize の不整合・循環参照

### 具体層リスク（Activity）
- **実装リスク**: 技術的困難・依存関係のボトルネック
- **ブロックリスク**: ブロックされた Activity の累積

### 参照整合性リスク（層横断）
- 循環参照（UseCase 関係）
- 孤立エンティティ（usecase_id/objective_id 未設定）
- 参照先不明

## 対策優先度マトリクス

| 確率 | 影響 | 優先度 | 対応 |
|------|------|--------|------|
| high | critical | **最優先** | 即時対応 |
| high | high | 優先対応 | 今週中 |
| medium | high | 優先対応 | 計画的 |
| medium | medium | 計画的対応 | 監視継続 |
| low | low | 監視のみ | 定期確認 |

## リスク管理ワークフロー

```bash
# 1. 階層コンテキストの確認（どの層のリスクか）
zeus list vision       # ゴール層: Vision の整合性確認
zeus list objectives   # 目標層: Objective の進捗確認

# 2. リスク登録（Objective に紐付けて層を明確化）
zeus add risk "外部API依存" --probability medium --impact high \
  --objective obj-001 -d "外部APIの仕様変更リスク"

# 3. 関連する Problem/Assumption 登録
zeus add problem "API応答遅延" --severity high --objective obj-001
zeus add assumption "APIは99.9%可用" --objective obj-001

# 4. 状況確認
zeus list risks
zeus list problems
zeus list assumptions

# 5. 対策提案取得
zeus suggest --impact high

# 6. 対策適用
zeus apply <suggestion-id>
```

## 分析結果の活用

1. `zeus list risks` でリスク一覧確認
2. `zeus doctor` で参照整合性チェック
3. `zeus graph` で依存関係を可視化
4. `zeus dashboard` でリアルタイム監視
5. `zeus suggest` で対策提案を取得
6. `zeus apply <id>` で対策を適用

## 関連スキル

- zeus-suggest - 提案生成
