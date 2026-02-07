---
description: プロジェクトのリスクを分析し、対策を提案するスキル
use_when: |
  Use when user asks about risks, problems, or assumptions.
  Also use when user says "リスク", "問題", "前提条件", "risk", "problem", "assumption".
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の Risk, Problem, Assumption エンティティを活用してリスク要因を特定し、対策を提案します。

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
  --deliverable del-001 \     # 関連 Deliverable（任意）
  -d "リスクの詳細説明"
```

### フィールド

| フィールド | 説明 | 必須 |
|-----------|------|------|
| probability | 発生確率（high/medium/low） | 任意 |
| impact | 影響度（critical/high/medium/low） | 任意 |
| score | 総合スコア（自動計算） | - |
| objective_id | 関連 Objective | 任意 |
| deliverable_id | 関連 Deliverable | 任意 |
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
  --deliverable del-001 \     # 関連 Deliverable
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
  --deliverable del-001 \
  -d "前提条件の説明"
```

### Assumption → Risk 連携
- 未検証の Assumption はリスク要因
- Assumption が崩れた場合の影響分析
- 検証済み Assumption によるリスク軽減

## リスクカテゴリ

### プロジェクトリスク
1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

### 参照整合性リスク
- 循環参照（Objective 階層）
- 孤立エンティティ
- 参照先不明

### 依存関係リスク
- 依存関係のボトルネック
- ブロックされた Activity の累積

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
# 1. リスク登録
zeus add risk "外部API依存" --probability medium --impact high \
  --objective obj-001 -d "外部APIの仕様変更リスク"

# 2. 関連する Problem/Assumption 登録
zeus add problem "API応答遅延" --severity high --objective obj-001
zeus add assumption "APIは99.9%可用" --objective obj-001

# 3. 状況確認
zeus list risks
zeus list problems
zeus list assumptions

# 4. 対策提案取得
zeus suggest --impact high

# 5. 対策適用
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
