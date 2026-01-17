---
description: プロジェクトのリスクを分析し、対策を提案するスキル
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）のリスク要因を特定し、対策を提案します。

## 実行方法

```bash
# リスク分析
zeus predict risk

# 全予測分析（完了日・リスク・ベロシティ）
zeus predict all

# グラフで循環参照・孤立タスク確認
zeus graph
```

## 出力

```yaml
risks:
  - id: risk-1
    category: "schedule|resource|technical|external|wbs|dependency"
    description: "リスクの説明"
    probability: "high|medium|low"
    impact: "high|medium|low"
    mitigation: "対策"
```

## predict コマンド出力例

```json
{
  "risk": {
    "overall_level": "medium",
    "factors": [
      {
        "name": "Schedule Pressure",
        "description": "クリティカルパス上のタスクに遅延",
        "impact": 0.7
      },
      {
        "name": "Dependency Bottleneck",
        "description": "タスクXに5つの依存タスクが集中",
        "impact": 0.5
      }
    ],
    "score": 0.65
  }
}
```

## リスクカテゴリ

### 従来のリスク
1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

### Phase 6 固有のリスク

#### WBS階層のリスク
- **循環参照**: ParentID の循環（自動検出・防止済み）
- **階層の深さ不均衡**: 一部が深すぎる/浅すぎる
- **親タスクの進捗と子タスクの不整合**: 子タスク完了済みだが親が未完了
- **孤立タスク**: 親が削除されて参照切れ

#### タイムラインのリスク
- **クリティカルパス上の遅延**: スラック0のタスクが遅延
- **依存関係のボトルネック**: 特定タスクに依存が集中
- **スラック不足によるバッファ欠如**: 全体的に余裕がない
- **期限超過タスクの累積**: 未対処の遅延タスク

## 対策優先度

| 確率 | 影響 | 優先度 |
|------|------|--------|
| 高 | 高 | **最優先** |
| 高 | 中 | 優先対応 |
| 中 | 高 | 優先対応 |
| 中 | 中 | 計画的対応 |
| 低 | 低 | 監視のみ |

## 分析結果の活用

1. `zeus predict risk` でリスク要因を特定
2. `zeus graph` で依存関係を可視化
3. `zeus dashboard` でリアルタイム監視
4. `zeus suggest` で対策提案を取得
5. `zeus apply <id>` で対策を適用

## 関連

- zeus-project-scan - プロジェクトスキャン
- zeus-task-suggest - タスク提案
