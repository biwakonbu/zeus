# Zeus REST API 仕様書

Zeus 10概念モデルの REST API 仕様を定義する。

## 目次

1. [概要](#概要)
2. [認証](#認証)
3. [共通仕様](#共通仕様)
4. [エンドポイント一覧](#エンドポイント一覧)
5. [リソース別 API](#リソース別-api)
6. [集約エンドポイント](#集約エンドポイント)
7. [SSE イベント](#sse-イベント)
8. [エラーハンドリング](#エラーハンドリング)

---

## 概要

### ベース URL

```
http://localhost:8080/api/v1
```

### API バージョニング

- URL パスにバージョンを含める: `/api/v1/`
- メジャーバージョンのみをパスに含める
- 互換性のない変更時にバージョンを上げる

### コンテンツタイプ

- リクエスト: `application/json`
- レスポンス: `application/json`
- SSE: `text/event-stream`

---

## 認証

### 現在のバージョン

ローカル環境での使用を想定し、認証は不要。

### 将来の拡張

```yaml
# zeus.yaml での認証設定（将来実装）
api:
  auth:
    enabled: true
    method: "api_key"  # api_key | basic | jwt
```

---

## 共通仕様

### リクエストヘッダー

| ヘッダー | 必須 | 説明 |
|---------|------|------|
| Content-Type | Yes | `application/json` |
| Accept | No | `application/json` |
| If-Match | No | 楽観的ロック用 ETag |

### レスポンス構造

```typescript
// 成功時
interface APIResponse<T> {
  data: T;
  meta?: APIMeta;
  timestamp: string;  // ISO 8601
}

// エラー時
interface APIErrorResponse {
  error: APIError;
  timestamp: string;
}

interface APIMeta {
  pagination?: Pagination;
  etag?: string;
}

interface Pagination {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

interface APIError {
  code: string;
  message: string;
  details?: Record<string, any>;
}
```

### ページネーション

一覧取得 API はページネーションをサポート。

**クエリパラメータ:**

| パラメータ | デフォルト | 説明 |
|-----------|-----------|------|
| page | 1 | ページ番号 |
| per_page | 20 | 1ページあたりの件数（最大100） |

**レスポンス例:**

```json
{
  "data": [...],
  "meta": {
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 45,
      "total_pages": 3
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

### フィルタリング

一覧取得 API はクエリパラメータでフィルタリング可能。

```
GET /api/v1/objectives?status=active&owner=team-a
```

### ソート

```
GET /api/v1/objectives?sort=created_at&order=desc
```

| パラメータ | デフォルト | 説明 |
|-----------|-----------|------|
| sort | created_at | ソートフィールド |
| order | asc | asc または desc |

---

## エンドポイント一覧

### 基本 CRUD

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /vision | Vision 取得 |
| PUT | /vision | Vision 更新 |
| GET | /{concept}s | 一覧取得 |
| POST | /{concept}s | 作成 |
| GET | /{concept}s/{id} | 詳細取得 |
| PUT | /{concept}s/{id} | 更新 |
| DELETE | /{concept}s/{id} | 削除 |

**対応する {concept}:**

- objectives
- deliverables
- considerations
- decisions
- problems
- risks
- assumptions
- constraints
- qualities

### 特殊ルール

| 概念 | 制限事項 |
|------|---------|
| Vision | 単一リソース（ID なし）、DELETE 不可 |
| Decision | PUT は 405 Method Not Allowed（イミュータブル） |
| Constraint | 単一ファイル内のリストとして管理 |

### 集約エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /overview | 俯瞰ビュー |
| GET | /coverage | カバレッジスコア |
| GET | /events | SSE イベントストリーム |

### テンプレート API

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /assumptions/templates | テンプレート一覧 |
| GET | /assumptions/templates/{id} | テンプレート詳細 |
| POST | /assumptions/apply | テンプレート適用 |

---

## リソース別 API

### Vision

Vision はプロジェクトに1つのみ存在する単一リソース。

#### Vision 取得

```http
GET /api/v1/vision
```

**レスポンス:**

```json
{
  "data": {
    "id": "vision-001",
    "title": "AI駆動プロジェクト管理の実現",
    "statement": "AIと人間が協調してプロジェクトを成功に導く世界を作る",
    "success_criteria": [
      "90%のプロジェクトが計画通りに完了",
      "意思決定の透明性が100%確保される"
    ],
    "metadata": {
      "created_at": "2026-01-18T10:00:00+09:00",
      "updated_at": "2026-01-18T12:00:00+09:00"
    }
  },
  "meta": {
    "etag": "\"a1b2c3d4\""
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### Vision 更新

```http
PUT /api/v1/vision
Content-Type: application/json
If-Match: "a1b2c3d4"

{
  "title": "AI駆動プロジェクト管理の実現",
  "statement": "更新された Vision ステートメント",
  "success_criteria": ["新しい成功基準"]
}
```

### Objective

#### Objective 一覧取得

```http
GET /api/v1/objectives?status=active&page=1&per_page=20
```

**レスポンス:**

```json
{
  "data": [
    {
      "id": "obj-001",
      "title": "認証システムの実装",
      "status": "active",
      "progress": 45,
      "owner": "team-backend",
      "wbs_code": "1.1",
      "parent_id": null,
      "metadata": {
        "created_at": "2026-01-18T10:00:00+09:00"
      }
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 15,
      "total_pages": 1
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### Objective 作成

```http
POST /api/v1/objectives
Content-Type: application/json

{
  "title": "新機能の実装",
  "description": "ユーザー認証機能を実装する",
  "owner": "team-backend",
  "parent_id": "obj-001",
  "wbs_code": "1.1.1",
  "start_date": "2026-01-20",
  "due_date": "2026-02-15"
}
```

**レスポンス:**

```json
{
  "data": {
    "id": "obj-002",
    "title": "新機能の実装",
    "description": "ユーザー認証機能を実装する",
    "status": "draft",
    "progress": 0,
    "owner": "team-backend",
    "parent_id": "obj-001",
    "wbs_code": "1.1.1",
    "start_date": "2026-01-20",
    "due_date": "2026-02-15",
    "metadata": {
      "created_at": "2026-01-18T15:00:00+09:00"
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### Objective 詳細取得

```http
GET /api/v1/objectives/obj-001
```

#### Objective 更新

```http
PUT /api/v1/objectives/obj-001
Content-Type: application/json

{
  "status": "active",
  "progress": 60
}
```

#### Objective 削除

```http
DELETE /api/v1/objectives/obj-001
```

### Deliverable

#### Deliverable 作成

```http
POST /api/v1/deliverables
Content-Type: application/json

{
  "title": "API 設計書",
  "description": "REST API の詳細設計",
  "objective_id": "obj-001",
  "format": "document",
  "acceptance_criteria": [
    "エンドポイント定義が完了",
    "レスポンス構造が定義済み"
  ]
}
```

### Consideration

#### Consideration 作成

```http
POST /api/v1/considerations
Content-Type: application/json

{
  "title": "認証方式の選択",
  "context": "ユーザー認証をどの方式で実装するか",
  "deadline": "2026-01-25",
  "options": [
    {
      "title": "JWT 方式",
      "pros": ["ステートレス", "スケーラブル"],
      "cons": ["トークン管理が複雑"]
    },
    {
      "title": "セッション方式",
      "pros": ["シンプル", "即座に無効化可能"],
      "cons": ["サーバー側の状態管理"]
    }
  ],
  "stakeholders": ["backend-team", "security-team"],
  "related_to": ["obj-001"]
}
```

### Decision

Decision はイミュータブル。作成後の更新は不可。

#### Decision 作成

```http
POST /api/v1/decisions
Content-Type: application/json

{
  "consideration_id": "con-001",
  "chosen_option": "JWT 方式",
  "rationale": "スケーラビリティを優先するため JWT を選択",
  "decided_by": "architecture-team",
  "consequences": [
    "トークンリフレッシュの仕組みが必要",
    "Redis でトークンブラックリストを管理"
  ],
  "reversibility": "medium"
}
```

#### Decision 更新（エラー）

```http
PUT /api/v1/decisions/dec-001

→ 405 Method Not Allowed
```

```json
{
  "error": {
    "code": "METHOD_NOT_ALLOWED",
    "message": "Decisions are immutable and cannot be updated"
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

### Problem

#### Problem 作成

```http
POST /api/v1/problems
Content-Type: application/json

{
  "title": "認証エンドポイントの応答遅延",
  "description": "ログイン処理に3秒以上かかっている",
  "severity": "high",
  "status": "open",
  "affected_entities": ["del-001", "obj-001"],
  "root_cause": "調査中"
}
```

### Risk

#### Risk 作成

```http
POST /api/v1/risks
Content-Type: application/json

{
  "title": "外部API依存によるサービス停止リスク",
  "description": "認証プロバイダーの障害時に全機能が停止する可能性",
  "probability": "medium",
  "impact": "critical",
  "status": "identified",
  "category": "technical",
  "triggers": ["外部サービスの SLA 違反", "ネットワーク障害"],
  "mitigation": {
    "preventive": ["フォールバック認証の実装"],
    "corrective": ["手動での認証バイパス手順"]
  },
  "affected_entities": ["obj-001"],
  "owner": "infrastructure-team"
}
```

**レスポンス（自動計算されたスコア付き）:**

```json
{
  "data": {
    "id": "risk-001",
    "title": "外部API依存によるサービス停止リスク",
    "probability": "medium",
    "impact": "critical",
    "risk_score": "high",
    "calculated_score": 12,
    "status": "identified",
    ...
  }
}
```

### Assumption

#### Assumption 一覧取得

```http
GET /api/v1/assumptions?status=unvalidated&related_to=obj-001
```

#### Assumption 作成

```http
POST /api/v1/assumptions
Content-Type: application/json

{
  "title": "ユーザーは最新ブラウザを使用",
  "description": "対象ユーザーの90%以上が過去2年以内にリリースされたブラウザを使用",
  "source": "ユーザー調査（2025年Q4）",
  "related_to": ["obj-001"],
  "validation_method": "ブラウザ利用統計の収集",
  "impact_if_wrong": "レガシーブラウザ対応が必要になり工数増加"
}
```

#### テンプレート一覧取得

```http
GET /api/v1/assumptions/templates
```

**レスポンス:**

```json
{
  "data": [
    {
      "id": "owasp-top10",
      "name": "OWASP Top 10 セキュリティリスク",
      "version": "2025",
      "category": "security",
      "items_count": 10
    },
    {
      "id": "nfr-standard",
      "name": "非機能要件チェックリスト",
      "version": "1.0",
      "category": "architecture",
      "items_count": 15
    }
  ],
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### テンプレート適用

```http
POST /api/v1/assumptions/apply
Content-Type: application/json

{
  "template_id": "owasp-top10",
  "target_id": "obj-001",
  "selected_items": ["A01", "A02", "A03"]
}
```

**レスポンス:**

```json
{
  "data": {
    "created": [
      {"id": "assum-001", "title": "A01: アクセス制御の不備", "template_item": "A01"},
      {"id": "assum-002", "title": "A02: 暗号化の失敗", "template_item": "A02"},
      {"id": "assum-003", "title": "A03: インジェクション", "template_item": "A03"}
    ],
    "skipped": [],
    "errors": []
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

### Constraint

Constraint は単一ファイル内のリストとして管理。

#### Constraint 一覧取得

```http
GET /api/v1/constraints
```

**レスポンス:**

```json
{
  "data": [
    {
      "id": "const-001",
      "title": "予算制限",
      "type": "budget",
      "description": "開発予算は500万円以内",
      "value": "5000000",
      "unit": "JPY",
      "flexibility": "hard",
      "source": "経営会議決定"
    },
    {
      "id": "const-002",
      "title": "リリース期限",
      "type": "time",
      "description": "2026年3月末までにリリース",
      "value": "2026-03-31",
      "flexibility": "hard",
      "source": "営業契約"
    }
  ],
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### Constraint 追加

```http
POST /api/v1/constraints
Content-Type: application/json

{
  "title": "チーム規模",
  "type": "resource",
  "description": "開発チームは最大5名",
  "value": "5",
  "unit": "人",
  "flexibility": "soft",
  "source": "人事計画"
}
```

### Quality

#### Quality 作成

```http
POST /api/v1/qualities
Content-Type: application/json

{
  "title": "認証 API パフォーマンス基準",
  "deliverable_id": "del-001",
  "category": "performance",
  "criteria": [
    {
      "metric": "response_time",
      "target": "< 200ms",
      "priority": "must"
    },
    {
      "metric": "throughput",
      "target": "> 1000 req/s",
      "priority": "should"
    }
  ],
  "verification_method": "負荷テスト（k6）",
  "gates": ["code-review", "security-review", "load-test"]
}
```

---

## 集約エンドポイント

### 俯瞰ビュー

プロジェクト全体の状態を一括取得。

```http
GET /api/v1/overview
```

**レスポンス:**

```json
{
  "data": {
    "vision": {
      "id": "vision-001",
      "title": "AI駆動プロジェクト管理の実現",
      "statement": "AIと人間が協調してプロジェクトを成功に導く世界を作る"
    },
    "objectives": {
      "total": 15,
      "by_status": {
        "draft": 2,
        "active": 8,
        "completed": 4,
        "cancelled": 1
      },
      "progress": 62.5,
      "tree": [
        {
          "id": "obj-001",
          "title": "認証システムの実装",
          "wbs_code": "1.1",
          "status": "active",
          "progress": 75,
          "deliverables": [
            {"id": "del-001", "title": "API 設計書", "progress": 100}
          ],
          "children": [
            {"id": "obj-002", "title": "JWT 実装", "wbs_code": "1.1.1", "status": "active", "progress": 50}
          ]
        }
      ]
    },
    "issues": {
      "considerations": {
        "open": 5,
        "decided": 12,
        "deferred": 2,
        "overdue": [
          {"id": "con-003", "title": "デプロイ戦略", "deadline": "2026-01-15"}
        ]
      },
      "risks": {
        "total": 8,
        "by_score": {
          "critical": 1,
          "high": 2,
          "medium": 3,
          "low": 2
        },
        "critical": [
          {"id": "risk-001", "title": "外部API依存リスク", "risk_score": "critical"}
        ]
      },
      "problems": {
        "total": 3,
        "open": 2,
        "high": [
          {"id": "prob-001", "title": "認証エンドポイントの応答遅延", "severity": "high"}
        ]
      }
    },
    "coverage": {
      "overall": 78.5,
      "categories": {
        "objectives_with_deliverables": 90.0,
        "deliverables_with_quality": 75.0,
        "open_considerations": 85.0,
        "risks_mitigated": 62.5
      },
      "warnings": [
        "3つの Objective に Deliverable が未設定",
        "2つの Risk に軽減策が未設定"
      ]
    },
    "metrics": {
      "total_entities": 52,
      "last_updated": "2026-01-18T14:30:00+09:00"
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

### カバレッジスコア

プロジェクトのカバレッジ詳細を取得。

```http
GET /api/v1/coverage
```

**レスポンス:**

```json
{
  "data": {
    "overall": 78.5,
    "categories": {
      "objectives_with_deliverables": {
        "name": "Objective → Deliverable",
        "score": 90.0,
        "total": 10,
        "covered": 9,
        "details": "1つの Objective に Deliverable が未設定"
      },
      "deliverables_with_quality": {
        "name": "Deliverable → Quality",
        "score": 75.0,
        "total": 12,
        "covered": 9,
        "details": "3つの Deliverable に品質基準が未設定"
      },
      "considerations_decided": {
        "name": "Consideration → Decision",
        "score": 85.0,
        "total": 20,
        "covered": 17,
        "details": "3つの Consideration が未決定"
      },
      "risks_mitigated": {
        "name": "Risk with Mitigation",
        "score": 62.5,
        "total": 8,
        "covered": 5,
        "details": "3つの Risk に軽減策が未設定"
      },
      "assumptions_validated": {
        "name": "Assumption Validated",
        "score": 80.0,
        "total": 15,
        "covered": 12,
        "details": "3つの Assumption が未検証"
      }
    },
    "issues": [
      {
        "severity": "warning",
        "category": "deliverables_with_quality",
        "entity_id": "del-005",
        "entity_type": "deliverable",
        "message": "Quality 基準が未設定",
        "fix": "zeus add quality --deliverable del-005"
      },
      {
        "severity": "error",
        "category": "risks_mitigated",
        "entity_id": "risk-002",
        "entity_type": "risk",
        "message": "Critical リスクに軽減策が未設定",
        "fix": "zeus edit risk-002 で mitigation を追加"
      }
    ]
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

---

## SSE イベント

### 接続

```http
GET /api/v1/events
Accept: text/event-stream
```

### イベント種別

| イベント | 説明 | トリガー |
|---------|------|---------|
| entity.created | エンティティ作成 | POST 成功時 |
| entity.updated | エンティティ更新 | PUT 成功時 |
| entity.deleted | エンティティ削除 | DELETE 成功時 |
| status.changed | 状態変更 | status フィールド変更時 |
| progress.changed | 進捗変更 | progress フィールド変更時 |
| coverage.changed | カバレッジ変更 | カバレッジスコア変動時 |
| integrity.issue | 整合性問題検出 | 参照切れ検出時 |

### イベントフォーマット

```
event: entity.updated
data: {"type":"entity.updated","timestamp":"2026-01-18T15:00:00+09:00","data":{"entity_type":"objective","entity_id":"obj-001","changes":[{"field":"progress","old_value":45,"new_value":60}]}}

event: coverage.changed
data: {"type":"coverage.changed","timestamp":"2026-01-18T15:01:00+09:00","data":{"old_overall":78.5,"new_overall":80.0,"changes":{"deliverables_with_quality":2.5}}}
```

### ペイロード構造

#### entity.created / entity.updated

```typescript
interface EntityChangePayload {
  entity_type: string;  // "objective", "deliverable", etc.
  entity_id: string;
  action: "created" | "updated" | "deleted";
  entity?: object;      // created/updated 時のみ
  changes?: Change[];   // updated 時のみ
}

interface Change {
  field: string;
  old_value: any;
  new_value: any;
}
```

#### status.changed

```typescript
interface StatusChangePayload {
  entity_type: string;
  entity_id: string;
  old_status: string;
  new_status: string;
}
```

#### coverage.changed

```typescript
interface CoverageChangePayload {
  old_overall: number;
  new_overall: number;
  changes: Record<string, number>;  // category -> delta
}
```

#### integrity.issue

```typescript
interface IntegrityPayload {
  issues: IntegrityIssue[];
}

interface IntegrityIssue {
  type: "broken_reference" | "orphan_entity" | "cycle_detected";
  entity_type: string;
  entity_id: string;
  field?: string;
  target_id?: string;
  message: string;
}
```

---

## エラーハンドリング

### HTTP ステータスコード

| コード | 説明 |
|--------|------|
| 200 | 成功 |
| 201 | 作成成功 |
| 204 | 削除成功（ボディなし） |
| 400 | リクエスト不正 |
| 404 | リソース未発見 |
| 405 | メソッド不許可（Decision の PUT など） |
| 409 | 競合（ETag 不一致、参照制約違反） |
| 422 | バリデーションエラー |
| 500 | サーバーエラー |

### エラーコード一覧

| コード | 説明 |
|--------|------|
| INVALID_ID_FORMAT | ID 形式が不正 |
| ENTITY_NOT_FOUND | エンティティが存在しない |
| VALIDATION_ERROR | バリデーション失敗 |
| REFERENCE_ERROR | 参照先が存在しない |
| CYCLE_DETECTED | 循環参照検出 |
| ETAG_MISMATCH | 楽観的ロック失敗 |
| METHOD_NOT_ALLOWED | 許可されていない操作 |
| CONSTRAINT_VIOLATION | 制約違反 |
| IMMUTABLE_ENTITY | イミュータブルエンティティの更新試行 |

### エラーレスポンス例

#### バリデーションエラー

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": [
        {"field": "title", "message": "Title is required"},
        {"field": "due_date", "message": "Due date must be in the future"}
      ]
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### 参照エラー

```json
{
  "error": {
    "code": "REFERENCE_ERROR",
    "message": "Referenced entity not found",
    "details": {
      "field": "objective_id",
      "target_type": "objective",
      "target_id": "obj-999"
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

#### 循環参照エラー

```json
{
  "error": {
    "code": "CYCLE_DETECTED",
    "message": "Setting parent_id would create a circular reference",
    "details": {
      "entity_id": "obj-003",
      "parent_id": "obj-001",
      "cycle_path": ["obj-001", "obj-002", "obj-003", "obj-001"]
    }
  },
  "timestamp": "2026-01-18T15:00:00+09:00"
}
```

---

## パフォーマンス目標

| 操作 | 目標時間 | 備考 |
|------|---------|------|
| GET /api/v1/{concept}s | < 100ms | 1000 エンティティ |
| GET /api/v1/{concept}s/{id} | < 10ms | 単一取得 |
| POST /api/v1/{concept}s | < 50ms | 作成 |
| PUT /api/v1/{concept}s/{id} | < 50ms | 更新 |
| GET /api/v1/overview | < 500ms | 集約ビュー |
| GET /api/v1/coverage | < 200ms | カバレッジ計算 |

---

## Go 型定義

```go
// レスポンス構造体
type APIResponse struct {
    Data      interface{} `json:"data,omitempty"`
    Meta      *APIMeta    `json:"meta,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

type APIMeta struct {
    Pagination *Pagination `json:"pagination,omitempty"`
    ETag       string      `json:"etag,omitempty"`
}

type Pagination struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// SSE イベント構造体
type SSEEvent struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

type EntityChangePayload struct {
    EntityType string   `json:"entity_type"`
    EntityID   string   `json:"entity_id"`
    Action     string   `json:"action"`
    Entity     interface{} `json:"entity,omitempty"`
    Changes    []Change `json:"changes,omitempty"`
}

type Change struct {
    Field    string      `json:"field"`
    OldValue interface{} `json:"old_value"`
    NewValue interface{} `json:"new_value"`
}
```

---

## 関連ドキュメント

- [DETAILED_DESIGN.md](./DETAILED_DESIGN.md) - 詳細設計書
- [SECURITY.md](./SECURITY.md) - セキュリティガイドライン

---

*作成日: 2026-01-18*
*バージョン: 1.0*
