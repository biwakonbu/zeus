---
description: WBS 階層設計の知識。プロジェクト構造を L1-L5 の階層で正しく設計するためのガイド。
use_when: |
  Use when user asks about WBS, project structure, or hierarchy design.
  Also use when user says "WBS", "階層", "構造", "work breakdown", "分解".
---

# zeus-wbs-design

WBS（Work Breakdown Structure）の階層設計ガイド。

## 概要

Zeus の 10 概念モデルを使用して、プロジェクト全体を俯瞰可能な WBS 構造として設計するための知識を提供。

## WBS 階層設計

| 階層 | 役割 | Zeus での表現 | 例 |
|------|------|---------------|-----|
| **L1** | ゴール（理想像） | Vision | 「神の視点によるプロジェクト管理」 |
| **L2** | 要件定義 | Objective (親なし) | 「CLI操作性」「AI統合」「可視化」 |
| **L3** | 機能グループ | Objective (子) | 「状態管理機能」「分析機能」 |
| **L4** | 個別機能仕様 | Deliverable | 「status コマンド」「graph コマンド」 |
| **L5** | 作業単位 | Activity | 「バグ修正」「機能追加」（未完了のみ） |

## 階層間の関係

```
L1 Vision (1件・不変)
 └─ L2 Objective (要件・数件)
     └─ L3 Objective (機能グループ・親参照)
         └─ L4 Deliverable (個別仕様・Objective参照)
             └─ L5 Activity (作業・Deliverable または Objective 参照)
```

## 各階層の特性

| 階層 | 変動性 | 完了判定 | 数量目安 |
|------|--------|----------|----------|
| L1 | 不変（プロダクト寿命と同じ） | なし | 1件 |
| L2 | 低（要件変更時のみ） | 子の達成度で自動計算 | 5-10件 |
| L3 | 中（機能追加時） | 子の達成度で自動計算 | 10-30件 |
| L4 | 中〜高（仕様変更時） | 明示的に完了マーク | 30-100件 |
| L5 | 高（日々変動） | 作業完了で消化 | 随時 |

## L2 要件の定義観点

| 観点 | 説明 | 例 |
|------|------|-----|
| **機能要件** | 「〜ができる」 | 「Activity を管理できる」 |
| **非機能要件** | 品質特性 | 「安全である」「高速である」 |
| **ユーザー価値** | 提供価値 | 「俯瞰できる」「AI支援を受けられる」 |

## 10概念モデルとのマッピング

| 10概念 | WBS での役割 | 説明 |
|--------|-------------|------|
| Vision | L1 | プロダクトの理想像 |
| Objective | L2, L3 | 要件・機能グループ |
| Deliverable | L4 | 個別機能仕様・成果物 |
| Activity | L5 | 作業単位（未完了のみ） |
| Consideration | 意思決定プロセス | L2-L4 策定時の検討記録 |
| Decision | 意思決定結果 | 要件・仕様確定の記録 |
| Problem | リスク管理 | L2-L4 に紐付く課題 |
| Risk | リスク管理 | L2-L4 に紐付くリスク |
| Assumption | リスク管理 | L2-L4 の前提条件 |
| Constraint | 全体制約 | 全階層に適用される制約 |
| Quality | 品質基準 | L4 (Deliverable) の品質メトリクス |

## 進捗計算の方針

- **L4 (Deliverable)**: 明示的に progress 設定
- **L3 (Objective子)**: 配下 L4 の加重平均
- **L2 (Objective親)**: 配下 L3 の加重平均

## WBS 構築コマンド

### L2 Objective（要件）登録

```bash
zeus add objective "要件名" --wbs "N.0" --progress <0-100> -d "説明"
```

### L3 Objective（機能グループ）登録

```bash
zeus add objective "機能グループ名" --parent obj-NNN --wbs "N.M" --progress <0-100>
```

### L4 Deliverable（個別仕様）登録

```bash
zeus add deliverable "仕様名" --objective obj-NNN --format code --progress <0-100>
```

### L5 Activity（作業単位）登録

```bash
zeus add activity "作業名" --priority <high|medium|low> --wbs "N.M.L" -d "説明"
```

## WBS 設計原則

1. **L1 は不変**: Vision はプロダクトの理想像であり、変更されない
2. **L2 は要件**: 「〜できる」形式で定義
3. **L3 は機能グループ**: L2 配下に分類される機能群
4. **L4 は個別仕様**: 具体的な機能・成果物
5. **L5 は作業**: 未完了のみ登録（完了したら削除）
6. **完了済みも含める**: WBS は「やったこと」と「やること」の両方を網羅

## 関連スキル

- zeus-suggest - 提案生成
- zeus-risk-analysis - リスク分析
