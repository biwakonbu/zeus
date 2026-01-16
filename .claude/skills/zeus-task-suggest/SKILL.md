---
description: 現在の状態に基づいてタスクを提案するスキル
---

# zeus-task-suggest

現在の状態に基づいてタスクを提案するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の状態を分析し、次に取り組むべきタスクを提案します。

## 入力

```yaml
context:
  current_tasks: []
  objectives: []
  blockers: []
```

## 出力

```yaml
suggestions:
  - id: suggestion-1
    type: "new_task|priority_change|dependency"
    description: "提案の説明"
    rationale: "理由"
    impact: "high|medium|low"
```

## アルゴリズム

1. 現在のタスク状態を分析
2. 目標との差分を計算
3. 優先度に基づいて提案を生成

## 承認

提案は `zeus pending` で確認し、`zeus approve <id>` で適用します。
