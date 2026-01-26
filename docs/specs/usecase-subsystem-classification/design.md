# 設計書 - USECASE サブシステム分類機能

**バージョン**: 1.0
**公開日**: 2026-01-26

---

## 1. アーキテクチャ概要

```
+---------------------------------------------------------------------+
|                         CLI Layer                                    |
|  zeus add subsystem / zeus list subsystems / zeus add usecase        |
+---------------------------------------------------------------------+
                                |
                                v
+---------------------------------------------------------------------+
|                        Core Layer                                    |
|  SubsystemHandler | UseCaseHandler | SecurityValidator               |
+---------------------------------------------------------------------+
                                |
                                v
+---------------------------------------------------------------------+
|                       Storage Layer                                  |
|  .zeus/subsystems.yaml | .zeus/usecases/*.yaml                       |
+---------------------------------------------------------------------+
                                |
                                v
+---------------------------------------------------------------------+
|                        API Layer                                     |
|  /api/subsystems | /api/usecases                                     |
+---------------------------------------------------------------------+
                                |
                                v
+---------------------------------------------------------------------+
|                      Frontend Layer                                  |
|  UseCaseEngine | SubsystemBoundary | ColorUtils                      |
+---------------------------------------------------------------------+
```

---

## 2. バックエンド設計

### 2.1 型定義

**SubsystemEntity** - サブシステムエンティティ

```go
type SubsystemEntity struct {
    ID          string   `yaml:"id"`
    Name        string   `yaml:"name"`
    Description string   `yaml:"description,omitempty"`
    Metadata    Metadata `yaml:"metadata"`
}
```

**SubsystemsFile** - サブシステム管理ファイル構造

```go
type SubsystemsFile struct {
    Subsystems []SubsystemEntity `yaml:"subsystems"`
}
```

### 2.2 SubsystemHandler

EntityHandler インターフェースに準拠した CRUD 操作を提供:

| メソッド | 説明 |
|---------|------|
| Add | サブシステム追加（ID 自動生成） |
| List | サブシステム一覧取得 |
| Get | サブシステム詳細取得 |
| Update | サブシステム更新 |
| Delete | サブシステム削除 |

### 2.3 Option パターン

```go
// WithSubsystemDescription - 説明設定
func WithSubsystemDescription(desc string) SubsystemOption
```

### 2.4 UseCaseHandler 拡張

```go
// WithUseCaseSubsystem - サブシステム ID 設定
func WithUseCaseSubsystem(subsystemID string) UseCaseOption
```

### 2.5 Doctor 参照整合性チェック

UseCase が参照するサブシステム ID の存在確認を警告レベルで実施:

- 存在しないサブシステム ID を参照: 警告（ReferenceWarning）
- 致命的エラーにはしない（後方互換性）

---

## 3. API 設計

### 3.1 GET /api/subsystems

サブシステム一覧を取得。

**レスポンス**:
```json
{
  "subsystems": [
    {
      "id": "sub-a1b2c3d4",
      "name": "認証サブシステム",
      "description": "ユーザー認証・認可を担当"
    }
  ],
  "total": 1
}
```

### 3.2 GET /api/usecases（拡張）

UseCaseItem に `subsystem_id` フィールドを追加:

```json
{
  "usecases": [
    {
      "id": "uc-12345678",
      "title": "ユーザーログイン",
      "subsystem_id": "sub-a1b2c3d4"
    }
  ]
}
```

---

## 4. フロントエンド設計

### 4.1 型定義

```typescript
interface SubsystemItem {
    id: string;
    name: string;
    description?: string;
}

interface SubsystemsResponse {
    subsystems: SubsystemItem[];
    total: number;
}
```

### 4.2 カラー生成アルゴリズム

DJB2 ハッシュアルゴリズムを使用して、サブシステム ID から一貫したカラーを生成:

1. サブシステム ID を DJB2 ハッシュ
2. ハッシュ値から Hue（色相）を計算（30-330 の範囲）
3. Saturation = 45%, Lightness = 40% で HSL カラーを生成
4. HSL から Hex に変換

**未分類サブシステム**: 固定のグレー色（0x555555）

### 4.3 SubsystemBoundary クラス

PixiJS Container を継承したサブシステム境界描画クラス:

| 要素 | 説明 |
|------|------|
| 影 | ドロップシャドウ効果 |
| 背景 | サブシステムカラーの半透明塗り |
| 外枠 | サブシステムカラーのボーダー |
| タイトルバー | 左上に配置（UML 準拠） |
| タイトルテキスト | サブシステム名 |

### 4.4 レイアウト

サブシステムごとに UseCase をグループ化し、グリッドレイアウトで配置:

1. UseCase を `subsystem_id` でグループ化
2. 各グループの UseCase 数から境界サイズを計算
3. サブシステム境界を順次配置
4. 境界内に UseCase ノードを配置

---

## 5. ファイル構成

### 5.1 バックエンド

| ファイル | 内容 |
|----------|------|
| `internal/core/types.go` | SubsystemEntity, SubsystemsFile 型 |
| `internal/core/subsystem_handler.go` | SubsystemHandler 実装 |
| `internal/core/subsystem_handler_test.go` | テスト |
| `internal/core/usecase_handler.go` | SubsystemID 対応 |
| `internal/core/zeus.go` | SubsystemHandler 登録 |
| `internal/core/security.go` | sub- ID バリデーション |
| `internal/doctor/doctor.go` | 参照整合性チェック |
| `internal/dashboard/handlers.go` | API エンドポイント |
| `cmd/add.go` | subsystem コマンド |
| `cmd/list.go` | subsystems 一覧 |

### 5.2 フロントエンド

| ファイル | 内容 |
|----------|------|
| `zeus-dashboard/src/lib/types/api.ts` | SubsystemItem 型 |
| `zeus-dashboard/src/lib/api/client.ts` | fetchSubsystems 関数 |
| `zeus-dashboard/src/lib/viewer/usecase/utils.ts` | カラー生成関数 |
| `zeus-dashboard/src/lib/viewer/usecase/rendering/SubsystemBoundary.ts` | 境界描画クラス |

---

## 6. データフロー

### 6.1 サブシステム作成フロー

```
[CLI] zeus add subsystem "認証"
    |
    v
[SubsystemHandler.Add]
    +-- ID 生成: sub-a1b2c3d4
    +-- バリデーション
    +-- subsystems.yaml に追記
    |
    v
[出力] Created subsystem: sub-a1b2c3d4
```

### 6.2 UseCase 表示フロー

```
[Frontend] UseCaseView マウント
    |
    v
[API] GET /api/subsystems, GET /api/usecases
    |
    v
[UseCaseEngine]
    +-- サブシステム別グループ化
    +-- レイアウト計算
    +-- SubsystemBoundary 描画
    +-- UseCaseNode をグループ内に配置
    |
    v
[Canvas] PixiJS レンダリング
```

---

## 7. 関連ドキュメント

- [要件定義](./requirements.md)
- [UML ユースケース機能仕様](../uml-usecase-feature/)
