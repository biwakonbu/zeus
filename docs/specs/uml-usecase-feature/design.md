# UML ユースケース機能 設計書

## 概要

| 項目 | 値 |
|------|-----|
| バージョン | 1.0.0 |
| ステータス | 実装完了 |

## アーキテクチャ

```
AI / ユーザー
    │ 記述
    ▼
YAML ファイル（Zeus 独自形式）
    │ パース
    ▼
Go バックエンド（zeus uml コマンド / API）
    │ JSON で配信
    ▼
Svelte + Mermaid フロントエンド
    │ 描画
    ▼
インタラクティブな UML ユースケース図
```

## データモデル

### Actor

**ファイル**: `.zeus/actors.yaml`（単一ファイル）

```yaml
actors:
  - id: actor-001
    title: "管理者"
    type: human          # human | system | time | device | external
    description: "システム管理権限を持つユーザー"
    created_at: "2026-01-22T10:00:00+09:00"

  - id: actor-002
    title: "外部認証システム"
    type: system
    description: "OAuth 2.0 プロバイダー"
    created_at: "2026-01-22T10:00:00+09:00"
```

**Go 型定義**:
```go
type Actor struct {
    ID          string    `yaml:"id"`
    Title       string    `yaml:"title"`
    Type        string    `yaml:"type"`
    Description string    `yaml:"description"`
    CreatedAt   time.Time `yaml:"created_at"`
}

type ActorsFile struct {
    Actors []Actor `yaml:"actors"`
}
```

### UseCase

**ファイル**: `.zeus/usecases/uc-NNN.yaml`（個別ファイル）

```yaml
id: uc-001
title: "ユーザー登録"
objective_id: obj-001
description: "新規ユーザーをシステムに登録する"

actors:
  - actor_id: actor-001
    role: primary            # primary | secondary

relations:
  - type: include            # include | extend | generalize
    target_id: uc-002
  - type: extend
    target_id: uc-003
    extension_point: "2段階認証"
    condition: "2段階認証が有効な場合"

scenario:
  main_flow:
    - "管理者がユーザー情報を入力"
    - "システムが入力値を検証"
    - "システムがユーザーを登録"
    - "完了メッセージを表示"

status: draft                # draft | active | deprecated
created_at: "2026-01-22T10:00:00+09:00"
```

**Go 型定義**:
```go
type UseCase struct {
    ID          string            `yaml:"id"`
    Title       string            `yaml:"title"`
    ObjectiveID string            `yaml:"objective_id"`
    Description string            `yaml:"description"`
    Actors      []UseCaseActor    `yaml:"actors"`
    Relations   []UseCaseRelation `yaml:"relations"`
    Scenario    UseCaseScenario   `yaml:"scenario"`
    Status      string            `yaml:"status"`
    CreatedAt   time.Time         `yaml:"created_at"`
}

type UseCaseActor struct {
    ActorID string `yaml:"actor_id"`
    Role    string `yaml:"role"`
}

type UseCaseRelation struct {
    Type           string `yaml:"type"`
    TargetID       string `yaml:"target_id"`
    ExtensionPoint string `yaml:"extension_point,omitempty"`
    Condition      string `yaml:"condition,omitempty"`
}

type UseCaseScenario struct {
    MainFlow []string `yaml:"main_flow"`
}
```

## CLI 設計

### Actor コマンド

```bash
# 追加
zeus add actor "管理者" --type human -d "システム管理権限を持つユーザー"

# 一覧
zeus list actors

# 詳細
zeus show actor-001
```

### UseCase コマンド

```bash
# 追加
zeus add usecase "ユーザー登録" \
  --objective obj-001 \
  --actor actor-001 \
  --actor-role primary \
  -d "新規ユーザーをシステムに登録"

# 一覧
zeus list usecases

# 詳細
zeus show uc-001
```

### UseCase 関係コマンド

```bash
# include 関係
zeus usecase link uc-001 --include uc-002

# extend 関係
zeus usecase link uc-001 --extend uc-003 \
  --condition "2段階認証時" \
  --extension-point "認証方式選択"

# generalize 関係
zeus usecase link uc-001 --generalize uc-004
```

### UML 図表示コマンド

```bash
# TEXT 形式（デフォルト）
zeus uml show usecase

# Mermaid 形式
zeus uml show usecase --format mermaid

# システム境界指定
zeus uml show usecase --boundary "認証システム"

# ファイル出力
zeus uml show usecase --format mermaid -o usecase.md
```

## API 設計

### GET /api/actors

**レスポンス**:
```json
{
  "actors": [
    {
      "id": "actor-001",
      "title": "管理者",
      "type": "human",
      "description": "システム管理権限を持つユーザー"
    }
  ]
}
```

### GET /api/usecases

**レスポンス**:
```json
{
  "usecases": [
    {
      "id": "uc-001",
      "title": "ユーザー登録",
      "objective_id": "obj-001",
      "description": "新規ユーザーをシステムに登録",
      "actors": [
        {"actor_id": "actor-001", "role": "primary"}
      ],
      "relations": [
        {"type": "include", "target_id": "uc-002"}
      ],
      "status": "draft"
    }
  ]
}
```

### GET /api/uml/usecase

**クエリパラメータ**:
- `boundary`: システム境界名（オプション）

**レスポンス**:
```json
{
  "mermaid": "flowchart TB\n  subgraph \"認証システム\"\n    uc1((ユーザー登録))\n  end\n  actor1[/管理者\\]\n  actor1 --> uc1",
  "boundary": "認証システム"
}
```

## ダッシュボード設計

### UseCaseView レイアウト

```
+------------------+------------------------+------------------+
|   Actor 一覧     |    Mermaid 図表示       |    詳細パネル    |
|   （左カラム）    |     （中央カラム）       |   （右カラム）   |
+------------------+------------------------+------------------+
```

### インタラクション

| 操作 | 動作 |
|------|------|
| クリック（Actor/UseCase） | エンティティ詳細パネル表示 |
| ホバー | 関連エンティティハイライト + プレビュー |
| リフレッシュボタン | データ再取得 |

### 技術スタック

- **図表示**: Mermaid.js（ダイナミックインポート）
- **レイアウト**: CSS Grid 3 カラム
- **状態管理**: Svelte 5 runes（$state, $effect）

## 既存モデルとの関係

### Zeus 概念モデルの拡張

| 既存 | 新規 |
|------|------|
| Vision | - |
| Objective | UseCase.objective_id で参照 |
| Deliverable | - |
| Task | - |
| **（新規）Actor** | アクター定義 |
| **（新規）UseCase** | ユースケース定義 |

### 参照整合性

| 参照元 | 参照先 | 種別 |
|--------|--------|------|
| UseCase | Objective | 必須（objective_id） |
| UseCase | Actor | 任意（actors[].actor_id） |
| UseCase | UseCase | 任意（relations[].target_id） |

## 実装フェーズ

### MVP（Phase 1）- 実装完了

| 項目 | 状態 |
|------|------|
| ActorHandler | 実装完了 |
| UseCaseHandler | 実装完了 |
| CLI コマンド | 実装完了 |
| API エンドポイント | 実装完了 |
| UseCaseView | 実装完了 |

### Phase 2（将来拡張）

| 項目 | 内容 |
|------|------|
| シナリオ詳細化 | preconditions, alternative_flows, exception_flows, postconditions |
| Diagram 定義 | `.zeus/diagrams/usecase/ucd-NNN.yaml` |
| PixiJS 描画 | Mermaid から PixiJS への移行（オプション） |

## 関連ドキュメント

- [要件定義](./requirements.md)
- [コードレビュー結果](./REVIEW.md)
