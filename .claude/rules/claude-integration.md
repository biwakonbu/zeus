---
description: Claude Code 連携ファイルの編集ガイドライン。agents/skills 編集時に参照。
paths:
  - ".claude/agents/**"
  - ".claude/skills/**"
---

# Claude Code 連携

## 概要

`zeus init` で生成される Claude Code 連携ファイル群。
更新は `zeus update-claude` コマンドで実行。

## ディレクトリ構成

```
.claude/
├── agents/                   # エージェント定義
│   ├── zeus-orchestrator.md  # 全コマンド一覧・オーケストレーション
│   ├── zeus-planner.md       # WBS・タイムライン設計
│   └── zeus-reviewer.md      # 分析・レビュー
├── skills/                   # スキル定義
│   ├── zeus-e2e-tester/      # E2E テストスキル
│   │   ├── SKILL.md          # スキル定義
│   │   └── resources/        # ゴールデンファイル・シナリオ
│   ├── zeus-project-scan/    # プロジェクト状態取得
│   ├── zeus-task-suggest/    # タスク提案
│   └── zeus-risk-analysis/   # リスク分析
└── rules/                    # ルール定義（このファイル群）
```

## エージェント編集規則

- **フォーマット**: Markdown（frontmatter なし）
- **コマンド一覧**: 実装済みコマンドのみ記載
- **例示**: 具体的な使用例を含める

**自動生成**: `internal/generator/` で生成されるため、手動編集は非推奨

## スキル編集規則

- **SKILL.md**: スキルの入口点、使用方法と前提条件を記載
- **resources/**: スキル実行に必要なリソースファイル

**構造:**
```markdown
# スキル名

## 概要
[スキルの目的と機能]

## 前提条件
[実行に必要な環境・ツール]

## 使用方法
[コマンドや手順]
```

## ゴールデンファイル

- `resources/golden/` に配置
- E2E テストの期待値として使用
- 更新は `scripts/e2e/update-golden.sh` で実行

## 更新フロー

1. Go 実装を変更
2. `zeus update-claude` を実行（agents/skills 自動更新）
3. 必要に応じてゴールデンファイルを更新
4. E2E テストで動作確認
