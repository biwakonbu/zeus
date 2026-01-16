---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクト（New Zeus Project）の構造、タスク、進捗を分析します。

## 入力

なし（カレントディレクトリの .zeus/ を参照）

## 出力

```yaml
project:
  name: "プロジェクト名"
  health: "good|fair|poor"
  tasks:
    total: 10
    completed: 3
    in_progress: 2
    pending: 5
  risks: []
```

## 使用方法

1. `zeus status` コマンドを実行
2. 出力を解析
3. 改善提案を生成

## 関連

- zeus-task-suggest
- zeus-risk-analysis
