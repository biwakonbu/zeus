---
description: プロジェクトのリスクを分析し、対策を提案するスキル
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）のリスク要因を特定し、対策を提案します。

## 入力

```yaml
project_state:
  tasks: []
  timeline: {}
  resources: []
```

## 出力

```yaml
risks:
  - id: risk-1
    category: "schedule|resource|technical|external"
    description: "リスクの説明"
    probability: "high|medium|low"
    impact: "high|medium|low"
    mitigation: "対策"
```

## リスクカテゴリ

1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

## 対策優先度

- 高確率 x 高影響 = 最優先
- 低確率 x 低影響 = 監視のみ
