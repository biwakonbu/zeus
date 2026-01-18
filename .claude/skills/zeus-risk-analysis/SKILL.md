---
description: プロジェクトのリスクを分析し、対策を提案するスキル
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクトの 10概念モデルを活用してリスク要因を特定し、対策を提案します。Risk エンティティと他の概念との関連を分析します。

## 実行方法

```bash
# リスク一覧表示
zeus list risks

# リスク分析（予測）
zeus predict risk

# 全予測分析（完了日・リスク・ベロシティ）
zeus predict all

# グラフで循環参照・孤立タスク確認
zeus graph

# 参照整合性チェック
zeus doctor
```

## 出力

```yaml
risks:
  - id: risk-001
    name: "リスク名"
    probability: "high|medium|low"
    impact: "critical|high|medium|low"
    score: 12                        # 自動計算（確率×影響度）
    objective_id: "obj-001"          # 関連 Objective
    deliverable_id: ""               # 関連 Deliverable
    status: "identified|mitigating|mitigated|accepted"
    mitigation: "軽減策"

analysis:
  overall_level: "high|medium|low"
  score: 0.65
  factors:
    - name: "Schedule Pressure"
      description: "クリティカルパス上のタスクに遅延"
      impact: 0.7
```

## 10概念モデルとリスク分析

### Risk エンティティ
- `zeus add risk` で登録されたリスクを管理
- probability（確率）× impact（影響度）でスコア自動計算
- Objective/Deliverable との紐付けで影響範囲を把握

### Problem → Risk 連携
- 未解決 Problem がリスク要因に
- severity: high の Problem は高リスク
- Problem 放置期間によるリスク増大

### Assumption → Risk 連携
- 未検証 Assumption がリスク要因に
- Assumption が崩れた場合の影響分析
- 検証済み Assumption によるリスク軽減

### Constraint との関係
- 制約違反リスクの検出
- non-negotiable 制約への影響評価

### Quality との関係
- 品質基準未達のリスク
- メトリクス悪化傾向の早期警告

## リスクカテゴリ

### 従来のリスク
1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

### 10概念モデル固有のリスク

#### 参照整合性リスク
- 循環参照の検出（`zeus doctor`）
- 孤立エンティティの検出
- 参照先不明の検出

#### Vision/Objective リスク
- Vision 達成基準の未達リスク
- Objective 期限超過リスク
- 目標間の整合性リスク

#### Deliverable リスク
- 受入基準未達リスク
- Quality 基準未達リスク
- 依存 Deliverable の遅延リスク

#### WBS階層のリスク
- 階層の深さ不均衡
- 親タスクの進捗と子タスクの不整合
- 孤立タスク（親が削除されて参照切れ）

#### タイムラインのリスク
- クリティカルパス上の遅延
- 依存関係のボトルネック
- スラック不足によるバッファ欠如
- 期限超過タスクの累積

## 対策優先度マトリクス

| 確率 | 影響 | 優先度 | 対応 |
|------|------|--------|------|
| 高 | critical | **最優先** | 即時対応 |
| 高 | high | 優先対応 | 今週中 |
| 中 | high | 優先対応 | 計画的 |
| 中 | medium | 計画的対応 | 監視継続 |
| 低 | low | 監視のみ | 定期確認 |

## リスク管理ワークフロー

```bash
# 1. リスク登録
zeus add risk "リスク名" --probability medium --impact high \
  --objective obj-001 -d "リスクの詳細説明"

# 2. リスク一覧確認
zeus list risks

# 3. 予測分析でリスクスコア確認
zeus predict risk

# 4. 関連する Problem/Assumption 確認
zeus list problems
zeus list assumptions

# 5. 対策提案取得
zeus suggest --impact high

# 6. 対策適用
zeus apply <suggestion-id>
```

## 分析結果の活用

1. `zeus list risks` でリスク一覧確認
2. `zeus predict risk` でリスクスコア分析
3. `zeus doctor` で参照整合性チェック
4. `zeus graph` で依存関係を可視化
5. `zeus dashboard` でリアルタイム監視
6. `zeus suggest` で対策提案を取得
7. `zeus apply <id>` で対策を適用

## 関連

- zeus-project-scan - プロジェクトスキャン
- zeus-task-suggest - タスク提案
