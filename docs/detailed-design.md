# Zeus 10概念モデル 詳細設計書

## 概要

本書は Zeus の次期バージョン（10概念モデル）の詳細設計を定義する。
ラウンドテーブル議論（Round 1-5）で合意された設計仕様を実装可能なレベルで記述する。

**設計哲学:**
- Zeus は「神の視点」でプロジェクトの構造的全体像を表現する
- タスク管理は外部ツールに委譲し、構造分析に特化
- ファイルベース、人間中心、Git 親和性を維持

---

## 1. 10概念モデル

### 1.1 概念一覧

| 概念 | 役割 | ID形式 | ファイル構造 |
|------|------|--------|-------------|
| Vision | プロジェクトの最上位目標 | `vision-NNN` | 単一ファイル |
| Objective | 達成すべき目標 | `obj-NNN` | 複数ファイル |
| Deliverable | 具体的な成果物 | `del-NNN` | 複数ファイル |
| Consideration | 検討事項 | `con-NNN` | 複数ファイル |
| Decision | 決定記録（イミュータブル） | `dec-NNN` | 複数ファイル |
| Problem | 課題・障害 | `prob-NNN` | 複数ファイル |
| Risk | リスク | `risk-NNN` | 複数ファイル |
| Assumption | 前提条件 | `assum-NNN` | 複数ファイル |
| Constraint | 制約条件 | `const-NNN` | 単一ファイル |
| Quality | 品質基準 | `qual-NNN` | 複数ファイル |

### 1.2 ディレクトリ構造

```
.zeus/
├── zeus.yaml               # 設定
├── vision.yaml             # Vision（単一）
├── constraints.yaml        # Constraint（単一）
├── objectives/
│   ├── obj-001.yaml
│   └── obj-002.yaml
├── deliverables/
│   ├── del-001.yaml
│   └── del-002.yaml
├── considerations/
│   ├── con-001.yaml
│   └── con-002.yaml
├── decisions/
│   ├── dec-001.yaml
│   └── dec-002.yaml
├── problems/
│   ├── prob-001.yaml
│   └── prob-002.yaml
├── risks/
│   ├── risk-001.yaml
│   └── risk-002.yaml
├── assumptions/
│   ├── assum-001.yaml
│   └── templates/
│       ├── owasp-top10.yaml
│       └── non-functional.yaml
├── quality/
│   ├── qual-001.yaml
│   └── qual-002.yaml
├── state/
│   └── current.yaml
└── snapshots/
```

---

## 2. スキーマ定義

### 2.1 Vision

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^vision-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `statement` | string | ビジョン宣言 | 1-2000文字 |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `success_image` | string | "" | 成功イメージ |
| `values` | []string | [] | 優先価値リスト |
| `created_at` | datetime | 作成時刻 | ISO 8601形式 |
| `updated_at` | datetime | 作成時刻 | ISO 8601形式 |

**YAML例（最小構成）:**
```yaml
id: vision-001
title: "プロジェクトの目的"
statement: "達成したい最終目標の宣言"
```

**YAML例（フル構成）:**
```yaml
id: vision-001
title: "AI駆動型プロジェクト管理の実現"
statement: |
  人間とAIが協働し、プロジェクト全体を俯瞰できる
  管理システムを構築する。
success_image: |
  - 検討漏れがゼロ
  - 全決定に根拠がある
values:
  - "人間中心"
  - "透明性"
created_at: "2026-01-15T10:00:00+09:00"
updated_at: "2026-01-18T14:30:00+09:00"
```

### 2.2 Objective

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^obj-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `status` | enum | 状態 | draft, active, completed, on_hold |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `parent_id` | string | "" | 親Objective ID（階層用） |
| `wbs_code` | string | "" | WBS番号 |
| `vision_id` | string | "vision-001" | Vision参照 |
| `description` | string | "" | 詳細説明 |
| `success_criteria` | []SuccessCriterion | [] | 成功基準リスト |
| `progress` | int | 0 | 進捗率 0-100 |
| `owner` | string | "" | 担当者/チーム |
| `priority` | enum | medium | high, medium, low |
| `due_date` | date | nil | 期限 |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

**SuccessCriterion 型:**
```yaml
success_criteria:
  - id: "sc-001"
    description: "説明"
    verification: "検証方法"
    status: "not_verified"  # not_verified | verified | failed
```

**YAML例（最小構成）:**
```yaml
id: obj-001
title: "認証システムの構築"
status: active
```

**YAML例（フル構成）:**
```yaml
id: obj-001
title: "セキュアな認証基盤の構築"
parent_id: ""
wbs_code: "1.0"
vision_id: vision-001
description: |
  JWT + OAuth2 ベースの認証システムを構築し、
  全 API エンドポイントを保護する。
success_criteria:
  - id: sc-001
    description: "全保護エンドポイントで認証が動作"
    verification: "E2E テスト"
    status: not_verified
  - id: sc-002
    description: "レスポンス時間 200ms 以下"
    verification: "負荷テスト"
    status: not_verified
progress: 40
status: active
owner: "backend-team"
priority: high
due_date: "2026-02-28"
created_at: "2026-01-15T10:00:00+09:00"
updated_at: "2026-01-18T14:30:00+09:00"
```

### 2.3 Deliverable

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^del-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `objective_id` | string | 紐付くObjective | 存在チェック |
| `status` | enum | 状態 | draft, in_progress, completed, accepted |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `type` | enum | other | code, document, design, data, other |
| `path` | string | "" | リポジトリ内パス |
| `description` | string | "" | 詳細説明 |
| `acceptance_criteria` | []string | [] | 受入基準 |
| `owner` | string | "" | 担当者 |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

**YAML例:**
```yaml
id: del-001
title: "認証モジュール"
objective_id: obj-001
type: code
path: "internal/auth/"
status: in_progress
description: "JWT 発行・検証を担当するモジュール"
acceptance_criteria:
  - "テストカバレッジ 80% 以上"
  - "セキュリティレビュー完了"
owner: "backend-team"
created_at: "2026-01-15T10:00:00+09:00"
updated_at: "2026-01-18T14:30:00+09:00"
```

### 2.4 Consideration

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^con-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `status` | enum | 状態 | open, decided, deferred |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `objective_id` | string | "" | 紐付くObjective |
| `deliverable_id` | string | "" | 紐付くDeliverable |
| `context` | string | "" | 検討背景 |
| `options` | []Option | [] | 選択肢リスト |
| `decision_id` | string | "" | 決定後のDecision参照 |
| `raised_by` | string | "" | 提起者 |
| `due_date` | date | nil | 決定期限 |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

**Option 型:**
```yaml
options:
  - id: "opt-1"
    title: "選択肢名"
    description: "詳細"
    pros: []
    cons: []
```

**YAML例:**
```yaml
id: con-001
title: "リフレッシュトークンの保存方法"
objective_id: obj-001
deliverable_id: del-001
status: open
context: |
  JWT のリフレッシュトークンをどこに保存するかの検討。
  セキュリティと UX のトレードオフがある。
options:
  - id: opt-1
    title: "httpOnly Cookie"
    pros:
      - "XSS 攻撃から保護される"
    cons:
      - "CSRF 対策が必要"
  - id: opt-2
    title: "localStorage"
    pros:
      - "実装が簡単"
    cons:
      - "XSS 攻撃に脆弱"
decision_id: ""
raised_by: "security-team"
due_date: "2026-01-20"
```

### 2.5 Decision

**重要: Decision はイミュータブル（作成後は更新不可）**

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^dec-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `consideration_id` | string | 元のConsideration | 存在チェック |
| `selected` | SelectedOption | 選択されたオプション | 必須 |
| `rationale` | string | 決定理由 | 1-2000文字 |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `rejected` | []RejectedOption | [] | 却下されたオプション |
| `impact` | []string | [] | 決定による影響 |
| `decided_at` | datetime | 作成時刻 | |
| `decided_by` | string | "" | 決定者 |

**YAML例:**
```yaml
id: dec-001
title: "リフレッシュトークンは httpOnly Cookie に保存"
consideration_id: con-001
selected:
  option_id: opt-1
  title: "httpOnly Cookie"
rationale: |
  セキュリティを最優先し、XSS 攻撃のリスクを排除する。
  CSRF 対策は SameSite=Strict と CSRF トークンで対応する。
rejected:
  - option_id: opt-2
    title: "localStorage"
    reason: "XSS 脆弱性が許容できない"
impact:
  - "CSRF 対策の実装が必要"
decided_at: "2026-01-18T15:00:00+09:00"
decided_by: "tech-lead"
```

### 2.6 Problem

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^prob-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `status` | enum | 状態 | open, in_progress, resolved, wont_fix |
| `severity` | enum | 深刻度 | critical, high, medium, low |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `objective_id` | string | "" | 紐付くObjective |
| `deliverable_id` | string | "" | 紐付くDeliverable |
| `description` | string | "" | 詳細説明 |
| `impact` | string | "" | 影響範囲 |
| `root_cause` | string | "" | 根本原因 |
| `potential_solutions` | []Solution | [] | 解決策候補 |
| `reported_by` | string | "" | 報告者 |
| `assigned_to` | string | "" | 担当者 |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

### 2.7 Risk

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^risk-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `status` | enum | 状態 | identified, mitigating, mitigated, occurred, closed |
| `probability` | enum | 発生確率 | high, medium, low |
| `impact` | enum | 影響度 | critical, high, medium, low |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `objective_id` | string | "" | 紐付くObjective |
| `deliverable_id` | string | "" | 紐付くDeliverable |
| `description` | string | "" | 詳細説明 |
| `risk_score` | enum | (自動計算) | critical, high, medium, low |
| `trigger` | string | "" | トリガー条件 |
| `mitigation.preventive` | []string | [] | 予防策 |
| `mitigation.contingent` | []string | [] | 発生時対応 |
| `owner` | string | "" | リスクオーナー |
| `review_date` | date | nil | 次回レビュー日 |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

**risk_score 計算ロジック:**

```
probability × impact    → risk_score
high       × critical  → critical
high       × high      → critical
high       × medium    → high
high       × low       → medium
medium     × critical  → critical
medium     × high      → high
medium     × medium    → medium
medium     × low       → low
low        × critical  → high
low        × high      → medium
low        × medium    → low
low        × low       → low
```

### 2.8 Assumption

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^assum-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `status` | enum | 状態 | assumed, validated, invalidated |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `objective_id` | string | "" | 紐付くObjective |
| `deliverable_id` | string | "" | 紐付くDeliverable |
| `description` | string | "" | 詳細説明 |
| `if_invalid` | string | "" | 無効時の影響 |
| `validation.method` | string | "" | 検証方法 |
| `validation.result` | string | "" | 検証結果 |
| `validation.validated_at` | datetime | nil | 検証日時 |
| `template_source.template_id` | string | "" | テンプレートID |
| `template_source.item_id` | string | "" | テンプレート項目ID |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

### 2.9 Constraint

**構造:** 単一ファイル内に制約リストを格納

**Constraint 項目:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^const-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `category` | enum | 分類 | technical, business, legal, resource |
| `description` | string | 詳細説明 | |
| `source` | string | 制約の出所 | |
| `impact` | []string | 制約による影響 | |
| `non_negotiable` | bool | 交渉不可フラグ | デフォルト: true |

**YAML例:**
```yaml
# .zeus/constraints.yaml
constraints:
  - id: const-001
    title: "外部 DB 不使用"
    category: technical
    description: "Zeus はファイルベースアーキテクチャを採用"
    source: "設計哲学"
    impact:
      - "大規模データの処理に制限"
      - "マルチユーザー同時編集に制約"
    non_negotiable: true

  - id: const-002
    title: "Go 1.21+ 必須"
    category: technical
    description: "ジェネリクス等の最新機能を使用"
    source: "技術選定"
    non_negotiable: false

metadata:
  created_at: "2026-01-15T10:00:00+09:00"
  updated_at: "2026-01-18T14:30:00+09:00"
```

### 2.10 Quality

**必須フィールド:**

| フィールド | 型 | 説明 | バリデーション |
|-----------|-----|------|---------------|
| `id` | string | 一意識別子 | `^qual-[0-9]{3}$` |
| `title` | string | タイトル | 1-200文字 |
| `deliverable_id` | string | 紐付くDeliverable | 存在チェック |
| `metrics` | []Metric | 品質指標 | 1件以上 |

**任意フィールド:**

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `gates` | []Gate | [] | 品質ゲート |
| `reviewer` | string | "" | レビュアー |
| `created_at` | datetime | 作成時刻 | |
| `updated_at` | datetime | 作成時刻 | |

**Metric 型:**
```yaml
metrics:
  - id: "qm-001"
    name: "指標名"
    target: 80
    unit: "%"
    current: 85
    status: "met"  # met | not_met | in_progress
```

**Gate 型:**
```yaml
gates:
  - name: "ゲート名"
    criteria: ["qm-001"]
    status: "passed"  # passed | failed | pending
```

---

## 3. 参照関係

### 3.1 参照マッピング

```
Vision (1)
   └── Objective (N) [vision_id: 省略可、デフォルト vision-001]
         ├── parent_id: Objective（自己参照、循環禁止）
         ├── Deliverable (N) [objective_id: 必須]
         │     └── Quality (N) [deliverable_id: 必須]
         ├── Consideration (N) [objective_id/deliverable_id: 任意]
         │     └── Decision (1) [consideration_id: 必須]
         ├── Problem (N) [objective_id/deliverable_id: 任意]
         ├── Risk (N) [objective_id/deliverable_id: 任意]
         └── Assumption (N) [objective_id/deliverable_id: 任意]

Constraint: プロジェクト全体（参照なし）
```

### 3.2 参照制約

| 参照 | 種別 | 制約 |
|------|------|------|
| Deliverable → Objective | 必須 | 存在チェック |
| Quality → Deliverable | 必須 | 存在チェック |
| Decision → Consideration | 必須 | 存在チェック、1対1 |
| Objective → Objective | 任意 | 循環禁止、最大深度10 |
| Consideration → Objective/Deliverable | 任意 | 存在チェック |
| Problem → Objective/Deliverable | 任意 | 存在チェック |
| Risk → Objective/Deliverable | 任意 | 存在チェック |
| Assumption → Objective/Deliverable | 任意 | 存在チェック |

### 3.3 循環参照検出

```go
// 循環参照検出アルゴリズム（Objective の親子関係）
func (c *CircularChecker) HasCycle(objectiveID, parentID string) bool {
    visited := make(map[string]bool)
    stack := make(map[string]bool)

    // 仮に更新した場合のグラフで DFS
    current := parentID
    for current != "" {
        if stack[current] {
            return true  // 循環検出
        }
        if visited[current] {
            break
        }
        visited[current] = true
        stack[current] = true

        obj := c.getObjective(current)
        if obj == nil {
            break
        }
        current = obj.ParentID
        stack[obj.ID] = false
    }
    return false
}
```

---

## 4. 状態遷移

### 4.1 Objective

```
                 ┌─────────────────────────────────────┐
                 ▼                                     │
    ┌─────┐    ┌────────┐    ┌───────────┐    ┌───────┴───┐
    │draft│───▶│ active │───▶│ completed │    │  on_hold  │
    └─────┘    └────────┘    └───────────┘    └───────────┘
                 │                 ▲               │  ▲
                 │                 │               │  │
                 └─────────────────┴───────────────┘  │
                                                      │
                         任意の状態から on_hold へ ────┘
```

**遷移ルール:**
- `draft → active`: 明示的な開始
- `active → completed`: 全成功基準が verified
- `active → on_hold`: 一時停止
- `on_hold → active`: 再開
- `completed → active`: 再オープン（稀）

### 4.2 Deliverable

```
    ┌─────┐    ┌─────────────┐    ┌───────────┐    ┌──────────┐
    │draft│───▶│ in_progress │───▶│ completed │───▶│ accepted │
    └─────┘    └─────────────┘    └───────────┘    └──────────┘
                     │                   │
                     │                   │
                     └───────────────────┘
                          差し戻し
```

**遷移ルール:**
- `draft → in_progress`: 作業開始
- `in_progress → completed`: 作業完了
- `completed → accepted`: レビュー承認
- `completed → in_progress`: 差し戻し
- `accepted` は終端状態

### 4.3 Consideration

```
    ┌──────┐                      ┌──────────┐
    │ open │─────────────────────▶│ decided  │
    └──────┘                      └──────────┘
        │                              ▲
        │      ┌──────────┐            │
        └─────▶│ deferred │────────────┘
               └──────────┘
```

**遷移ルール:**
- `open → decided`: Decision 作成時に自動遷移
- `open → deferred`: 一時保留
- `deferred → decided`: 後日決定

### 4.4 Decision

- **状態なし**: Decision は作成されたら変更不可（イミュータブル）
- 決定の取り消しは新しい Decision で上書き

### 4.5 Problem

```
    ┌──────┐    ┌─────────────┐    ┌──────────┐
    │ open │───▶│ in_progress │───▶│ resolved │
    └──────┘    └─────────────┘    └──────────┘
        │              │
        │              │           ┌──────────┐
        └──────────────┴──────────▶│ wont_fix │
                                   └──────────┘
```

### 4.6 Risk

```
    ┌────────────┐    ┌────────────┐    ┌───────────┐
    │ identified │───▶│ mitigating │───▶│ mitigated │
    └────────────┘    └────────────┘    └───────────┘
          │                                    │
          │           ┌──────────┐             │
          └──────────▶│ occurred │◀────────────┘
                      └──────────┘
                            │
                            ▼
                      ┌──────────┐
                      │  closed  │
                      └──────────┘
```

### 4.7 Assumption

```
    ┌─────────┐    ┌───────────┐
    │ assumed │───▶│ validated │
    └─────────┘    └───────────┘
         │
         │         ┌─────────────┐
         └────────▶│ invalidated │
                   └─────────────┘
```

---

## 5. Assumption テンプレート

### 5.1 テンプレート構造

```yaml
template:
  id: string              # テンプレートID
  name: string            # 表示名
  version: string         # バージョン
  source: string          # 出典URL
  category: enum          # security | performance | compliance | usability | other

  items:
    - id: string          # 項目ID
      title: string       # タイトル
      description: string # 説明
      default_assumption: string  # デフォルト前提文
      validation_method: string   # 検証方法
      severity: enum      # critical | high | medium | low
      tags: []string      # タグ

metadata:
  created_at: datetime
  updated_at: datetime
  author: string
```

### 5.2 組み込みテンプレート

**OWASP Top 10 (2021):**
- A01: Broken Access Control
- A02: Cryptographic Failures
- A03: Injection
- A04: Insecure Design
- A05: Security Misconfiguration
- A06: Vulnerable and Outdated Components
- A07: Identification and Authentication Failures
- A08: Software and Data Integrity Failures
- A09: Security Logging and Monitoring Failures
- A10: Server-Side Request Forgery (SSRF)

**非機能要件:**
- NFR-PERF-001: レスポンス時間
- NFR-PERF-002: スループット
- NFR-AVAIL-001: 稼働率
- NFR-AVAIL-002: 障害復旧時間
- NFR-SCALE-001: 水平スケーラビリティ
- NFR-DATA-001: データバックアップ
- NFR-OPS-001: 監視・アラート

### 5.3 テンプレート適用フロー

```
1. zeus assumption templates        # テンプレート一覧
2. zeus assumption templates show owasp-top10  # 詳細表示
3. zeus assumption apply owasp-top10 --interactive  # 対話的適用
   または
   zeus assumption apply owasp-top10 --items A01,A02,A03 --objective obj-001
4. 生成された Assumption を必要に応じてカスタマイズ
```

---

## 6. CLI コマンド体系

### 6.1 俯瞰・検証

```bash
zeus overview [--format text|tree|json] [--scope all|objectives|risks|...]
zeus verify coverage [--strict] [--fix]
zeus verify integrity [--fix]
```

### 6.2 移行

```bash
zeus migrate analyze
zeus migrate tasks [--dry-run] [--interactive] [--no-backup]
zeus migrate verify
zeus migrate rollback
```

### 6.3 エンティティ操作

```bash
zeus add objective <title> [--parent <id>] [--wbs <code>]
zeus add deliverable <title> --objective <id>
zeus add consideration <title> [--objective <id>]
zeus add problem <title> --severity <level>
zeus add risk <title> --probability <p> --impact <i>
zeus add assumption <title>
zeus add constraint <title> --category <cat>
```

### 6.4 テンプレート

```bash
zeus assumption templates [show <id>]
zeus assumption apply <template-id> [--items <ids>] [--interactive]
zeus assumption validate <id> --method <m> --result <r>
```

### 6.5 Decision

```bash
zeus decision create --consideration <id> --selected <opt> --rationale <text>
```

### 6.6 一覧

```bash
zeus list [objectives|deliverables|considerations|decisions|problems|risks|assumptions|constraints|quality]
```

---

## 7. パフォーマンス目標

| 操作 | 目標時間 | 対象規模 |
|------|---------|---------|
| `zeus list` | < 100ms | 1000 エンティティ |
| `zeus status` | < 200ms | 全概念サマリー |
| `zeus overview` | < 500ms | 詳細ビュー |
| 単一エンティティ取得 | < 10ms | - |
| エンティティ作成 | < 50ms | - |

### 7.1 最適化戦略

- **遅延読み込み**: 一覧表示時はヘッダー（ID, Title, Status）のみパース
- **キャッシュ**: シンプルなインメモリキャッシュ（上限100、ファイル変更時に無効化）
- **インデックス不要**: ディレクトリスキャンで十分な速度

### 7.2 同時編集対応

- **ファイルロック**: `.lock` ファイルによる排他制御
- **stale ロック**: 5分以上古いロックは自動削除
- **Git マージ競合**: `zeus doctor` でコンフリクトマーカーを検出

### 7.3 データ破損対応

- **段階的パース**:
  1. 厳密パース（yaml.UnmarshalStrict）
  2. 緩いパース（yaml.Unmarshal）
  3. 最小限のフィールド抽出（ID, Title のみ）
- **参照エラー処理**:
  - 必須参照が切れている → エラー + 自動修復オプション
  - 任意参照が切れている → 警告 + 参照クリア提案

---

## 8. ダッシュボード連携

### 8.1 視覚表現

| 概念 | 色 | 形状 |
|------|-----|------|
| Vision | #FFD700 (Gold) | hexagon |
| Objective | #4CAF50 (Green) | rounded-rect |
| Deliverable | #2196F3 (Blue) | rect |
| Consideration | #FF9800 (Orange) | diamond |
| Decision | #9C27B0 (Purple) | diamond-filled |
| Problem | #F44336 (Red) | octagon |
| Risk | #FF5722 (Deep Orange) | triangle |
| Assumption | #607D8B (Blue Grey) | ellipse |
| Constraint | #795548 (Brown) | rect-dashed |
| Quality | #00BCD4 (Cyan) | shield |

### 8.2 エッジ種別

| 種別 | スタイル | 用途 |
|------|---------|------|
| hierarchy | solid, #333, 2px | 親子関係 |
| reference | dashed, #999, 1px | 参照関係 |
| resolution | solid, #9C27B0, 1.5px | Consideration → Decision |

---

## 9. 実装フェーズ

### Phase 1: MVP + セキュリティ基盤（Week 1-2）

**Week 1:**
- Vision, Objective, Deliverable の型定義
- CRUD 操作
- 参照整合性チェック
- パストラバーサル対策
- ID バリデーション

**Week 2:**
- Consideration, Decision の型定義
- CRUD 操作
- zeus overview 基本版
- 入力サニタイズ
- 基本的な監査ログ

### Phase 2: 課題・リスク管理 + API（Week 3-4）

**Week 3:**
- Problem, Risk, Assumption の型定義
- CRUD 操作
- Risk スコア自動計算
- REST API 基本エンドポイント

**Week 4:**
- テンプレート機能
- zeus assumption apply
- SSE 実装
- API 集約エンドポイント

### Phase 3: 品質・仕上げ（Week 5-6）

**Week 5:**
- Constraint, Quality の型定義
- CRUD 操作
- zeus verify coverage/integrity
- 整合性チェック

**Week 6:**
- ダッシュボード連携
- ビューワー拡張
- ドキュメント整備
- E2E テスト

### 将来（要望に応じて）

- GitHub Issues 連携（参照ベース）
- Slack 通知
- AI 提案機能
- プラグイン機構

---

## 10. Go 型定義

### 10.1 enum の型付き定数

```go
type ObjectiveStatus string

const (
    ObjectiveStatusDraft     ObjectiveStatus = "draft"
    ObjectiveStatusActive    ObjectiveStatus = "active"
    ObjectiveStatusCompleted ObjectiveStatus = "completed"
    ObjectiveStatusOnHold    ObjectiveStatus = "on_hold"
)
```

### 10.2 構造体（必須フィールドのみ required）

```go
type Objective struct {
    ID     string          `yaml:"id" json:"id"`
    Title  string          `yaml:"title" json:"title"`
    Status ObjectiveStatus `yaml:"status" json:"status"`

    // 以下は任意（omitempty）
    ParentID    string `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
    Description string `yaml:"description,omitempty" json:"description,omitempty"`
    // ...
}
```

### 10.3 バリデーション

```go
func (o *Objective) Validate() error {
    if o.ID == "" {
        return errors.New("id is required")
    }
    if !regexp.MustCompile(`^obj-\d{3}$`).MatchString(o.ID) {
        return errors.New("invalid id format")
    }
    if o.Title == "" {
        return errors.New("title is required")
    }
    return nil
}
```

---

## 付録: 移行ルール

### Activity と Objective/Deliverable の関係

Activity は実行可能な作業単位として、Objective や Deliverable と関連付けられます。

| Activity の用途 | 関連エンティティ |
|----------------|-----------------|
| 目標達成のための作業 | Objective に関連 |
| 成果物作成の作業 | Deliverable に関連 |
| UseCase の実装作業 | UseCase に関連 |

**Activity ステータス:**
- `pending` - 未着手
- `in_progress` - 作業中
- `completed` - 完了
- `blocked` - ブロック中

---

## 11. Activity 図設計ガイドライン

### 11.1 action name 記載ルール

#### 粒度

| 項目 | 基準 |
|------|------|
| アクション数 | 1 Activity あたり 5-15（7 +/- 2 が理想） |
| 単位 | ユーザーにとって意味のある処理ステップ |
| 判断基準 | エラー時に失敗箇所を特定できる単位 |

#### 命名規則

| 項目 | ルール |
|------|--------|
| 形式 | `<目的語> + <動詞（体言止め）>` |
| 文字数 | 全角 20 文字以内 |
| 抽象度 | 機能レイヤー（実装詳細は避ける） |
| 技術用語 | 一般的なものに限定、Go 固有用語は避ける |

**良い例:**
- `.zeus ディレクトリ作成`
- `zeus.yaml 生成`
- `参照整合性チェック`
- `初期化完了メッセージ表示`

**避けるべき例:**
- `os.MkdirAll() 呼び出し`（実装詳細の露出）
- `処理を行う`（曖昧すぎる）
- `goroutine で並列処理`（Go 固有用語）
- `A と B と C を実行`（複合処理の列挙）

### 11.2 トレーサビリティ

Activity は Deliverable との対応関係を明示することを推奨します。

| レベル | フィールド | 必須性 | 用途 |
|--------|-----------|--------|------|
| Activity | `related_deliverables` | **推奨** | Activity 全体の関連 Deliverable を俯瞰 |
| action | `deliverable_ids` | 任意 | 特定 action と Deliverable の明示的対応 |

**対応関係**: 多対1（複数 Activity/Deliverable が 1 action に対応可）

```yaml
# Activity レベル（推奨）
related_deliverables: [del-001, del-002]

# action レベル（任意）
nodes:
  - id: node-005
    name: ".zeus ディレクトリ作成"
    deliverable_ids: [del-001]
```

### 11.3 チェックリスト

Activity 作成時は以下を確認してください:

**粒度チェック:**
- [ ] 1 Activity あたり 5-15 アクションに収まっているか
- [ ] 各アクションは「エラー時に失敗箇所を特定できる」単位か
- [ ] 細かすぎる分割（関数呼び出し単位）になっていないか

**命名チェック:**
- [ ] `<目的語> + <動詞（体言止め）>` 形式になっているか
- [ ] 全角 20 文字以内か
- [ ] 実装詳細（関数名、メソッド名）を露出していないか

**トレーサビリティチェック:**
- [ ] Activity に `related_deliverables` を設定したか（推奨）
- [ ] UseCase の主要機能に直結する処理が Activity で言及されているか

---

*本書は Zeus 10概念モデル設計のラウンドテーブル議論（Round 1-5）および Activity 図ガイドライン議論（Round 1-2）の成果物である。*
*収束スコア: 100% / 90%*
*最終更新: 2026-02-04*
