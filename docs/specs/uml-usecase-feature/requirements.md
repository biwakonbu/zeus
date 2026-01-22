# UML ユースケース機能 要件定義

## 概要

| 項目 | 値 |
|------|-----|
| バージョン | 1.0.0 |
| ステータス | 実装完了 |
| 対象モジュール | zeus-core, zeus-dashboard |

## 背景

### 問題提起

WBS Dashboard の 4 ビュー（Health, Timeline, Density, Affinity）はプロジェクト管理の観点からは有用だが、ユーザーが「神の一手」を打つための戦略策定・仮説検証ツールとしては不十分。

### 解決方針

UML ユースケース図をベースにした新しいビューを導入し、以下を実現:

1. UML 標準に準拠した Actor/UseCase の定義
2. Zeus 独自のインタラクティブ機能
3. AI エージェントによる YAML 生成とプレビュー
4. 既存の 10 概念モデル（特に Objective）との連携

## 機能要件

### FR-001: Actor エンティティ

| 項目 | 仕様 |
|------|------|
| ファイル形式 | 単一ファイル（`.zeus/actors.yaml`） |
| ID 形式 | `actor-{uuid[:8]}` |
| type | human, system, time, device, external |

**YAML 形式**:
```yaml
actors:
  - id: actor-001
    title: "管理者"
    type: human
    description: "システム管理権限を持つユーザー"
```

**設計判断**:
- 単一ファイル: Actor は通常 5-10 個程度、個別ファイルは過剰
- `abstract` type は不採用: Actor の継承は過剰な抽象化

### FR-002: UseCase エンティティ

| 項目 | 仕様 |
|------|------|
| ファイル形式 | 個別ファイル（`.zeus/usecases/uc-NNN.yaml`） |
| ID 形式 | `uc-{uuid[:8]}` |
| 必須参照 | `objective_id`（Objective への参照） |

**MVP YAML 形式**:
```yaml
id: uc-001
title: "ユーザー登録"
objective_id: obj-001
description: "新規ユーザーをシステムに登録する"

actors:
  - actor_id: actor-001
    role: primary

relations:
  - type: include
    target_id: uc-002

scenario:
  main_flow:
    - "管理者がユーザー情報を入力"
    - "システムが入力値を検証"
    - "システムがユーザーを登録"
    - "完了メッセージを表示"

status: draft
```

**Phase 2 拡張**:
- preconditions, postconditions
- alternative_flows, exception_flows
- extension_point, condition（extend 関係用）

### FR-003: CLI コマンド

#### Actor 操作

```bash
zeus add actor "アクター名" --type human -d "説明"
zeus list actors
zeus show actor-001
```

#### UseCase 操作

```bash
zeus add usecase "ユースケース名" \
  --objective <obj-id> \
  --actor <actor-id> \
  --actor-role primary \
  -d "説明"

zeus list usecases
zeus show uc-001
```

#### UseCase 関係追加

```bash
zeus usecase link uc-001 --include uc-002
zeus usecase link uc-001 --extend uc-003 --condition "条件" --extension-point "拡張点"
zeus usecase link uc-001 --generalize uc-004
```

#### UML 図表示

```bash
zeus uml show usecase                              # TEXT 形式
zeus uml show usecase --format mermaid             # Mermaid 形式
zeus uml show usecase --boundary "システム名"       # システム境界指定
zeus uml show usecase -o diagram.md                # ファイル出力
```

### FR-004: API エンドポイント

| エンドポイント | メソッド | 説明 |
|---------------|---------|------|
| `/api/actors` | GET | Actor 一覧 |
| `/api/usecases` | GET | UseCase 一覧 |
| `/api/uml/usecase` | GET | ユースケース図（Mermaid 形式） |

### FR-005: ダッシュボード表示

| 機能 | 説明 |
|------|------|
| UseCaseView | UML ユースケース図を表示するビュー |
| レイアウト | 3 カラム（Actor 一覧 / Mermaid 図 / 詳細パネル） |
| インタラクション | クリックで詳細表示、ホバーでハイライト |

## 非機能要件

### NFR-001: UML 準拠

- Actor, UseCase, System Boundary の表現は UML 2.5 仕様に準拠
- 関係（include, extend, generalize）は UML 標準に従う

### NFR-002: 既存モデルとの整合性

- UseCase は必ず Objective を参照（`objective_id` 必須）
- EntityHandler パターンに準拠
- セキュリティ検証（ValidateID, パストラバーサル防止）

### NFR-003: パフォーマンス

- Actor/UseCase の一覧取得: 100ms 以内
- Mermaid 図生成: 500ms 以内（100 UseCase まで）

## 制約事項

1. GUI 編集は不要（AI エージェントが YAML を生成）
2. Diagram 定義ファイルは Phase 2 で導入（MVP は CLI オプションで代替）
3. Actor の継承（abstract type）は不採用

## 関連ドキュメント

- [設計書](./design.md)
- [コードレビュー結果](./REVIEW.md)
- [ラウンド議論](../../.round/20260122-usecase-uml/DISCUSSION.md)
