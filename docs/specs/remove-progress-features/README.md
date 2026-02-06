> **履歴資料（非正本）**  
> この文書は履歴資料。現行仕様の正本は `/Users/biwakonbu/github/zeus/docs/README.md` 参照。

# 進捗速度・タスク進行管理機能の削除

## 概要

Zeus v2.0 にて、「進捗速度・タスク進行管理」関連機能を削除しました。
これにより、Zeus は「プロジェクト構造の可視化・AI 支援」に特化したツールになりました。

## 背景

- Zeus の本来の目的は「神の視点」でプロジェクト構造を俯瞰すること
- 進捗管理は Jira、Linear 等の専用ツールが得意とする領域
- コードベースのシンプル化とメンテナンス負荷の軽減

## 削除された機能

### コマンド

| コマンド | 説明 |
|---------|------|
| `zeus predict` | 予測分析（完了日、リスク、ベロシティ） |
| `zeus graph --wbs` | WBS 階層表示 |

### API エンドポイント

| エンドポイント | 説明 |
|--------------|------|
| `/api/predict` | 予測 API |
| `/api/timeline` | タイムライン API |
| `/api/wbs` | WBS 階層 API |
| `/api/downstream` | 影響範囲 API |

### エンティティフィールド

**Activity:**
- `progress` - 進捗率
- `estimate_hours` - 見積もり時間
- `actual_hours` - 実績時間
- `start_date` - 開始日
- `due_date` - 期限日
- `wbs_code` - WBS コード

**Objective:**
- `progress` - 進捗率
- `wbs_code` - WBS コード
- `due_date` - 期限日

**Deliverable:**
- `progress` - 進捗率

## 維持された機能

以下の機能は引き続き利用可能です:

| 機能 | 説明 |
|------|------|
| `status` | 状態管理（pending, in_progress, completed, blocked） |
| `dependencies` | 依存関係 |
| `parent_id` | 階層構造 |
| `priority` | 優先度 |
| `assignee` | 担当者 |
| `zeus graph` | 依存関係グラフ |
| `zeus graph --unified` | 統合グラフ |

## 後方互換性

既存の `.zeus/` ディレクトリ内の YAML ファイルは引き続き読み込み可能です。
削除されたフィールドは単に無視されます。エラーや警告は発生しません。

## マイグレーション

特別な対応は不要です。Zeus を最新版に更新するだけで利用できます。

進捗管理が必要な場合は、以下のツールとの連携を検討してください:
- Jira
- Linear
- Asana
- GitHub Projects

## 関連ドキュメント

- [システム設計書](../../system-design.md)
- [詳細設計書](../../detailed-design.md)
