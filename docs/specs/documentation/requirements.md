# ドキュメント整備 要件定義

## 1. 目的

Zeus CLI の利用者向けドキュメントを整備し、ユーザビリティを向上させる。

## 2. 成果物

### 2.1 USER_GUIDE.md（ユーザーガイド）

**対象読者**: Zeus CLI を初めて使うユーザー、日常的に使用するユーザー

**必須セクション**:

1. **はじめに** - Zeus の概要、主な機能、対象ユーザー
2. **インストール** - 前提条件（Go 1.21+）、ビルド方法、動作確認
3. **クイックスタート** - 5分で始める Zeus、基本操作フロー
4. **基本的な使い方** - init, status, add, list, doctor, fix
5. **承認フロー** - 3段階承認レベル、pending, approve, reject
6. **スナップショットと履歴** - snapshot, history
7. **AI機能** - suggest, apply, explain, graph, predict, report
8. **ベストプラクティス** - 推奨ワークフロー、避けるべきこと
9. **トラブルシューティング** - よくある問題と解決方法

### 2.2 API_REFERENCE.md（APIリファレンス）

**対象読者**: コマンドの詳細仕様を知りたいユーザー、スクリプトから利用するユーザー

**必須セクション**:

1. **概要** - コマンド体系、グローバルフラグ、出力形式
2. **コマンドリファレンス** - 全18コマンドの詳細仕様
3. **データ型** - TaskStatus, ApprovalLevel, HealthStatus 等
4. **ファイル形式** - zeus.yaml, task, state, suggestion の構造
5. **エラーコード** - エラー種類と対処方法

## 3. 品質基準

- 全コマンドに使用例を含める
- コードブロックは適切な言語指定をする
- 見出しの階層は3レベルまで
- 日本語で記述
- Markdown 形式

## 4. 関連ドキュメント

- [USER_GUIDE.md](/docs/USER_GUIDE.md) - 成果物
- [API_REFERENCE.md](/docs/API_REFERENCE.md) - 成果物
