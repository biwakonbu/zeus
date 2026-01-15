# ドキュメント整備 設計書

## 1. ドキュメント構成

```
docs/
├── SYSTEM_DESIGN.md        # システム設計書
├── IMPLEMENTATION_GUIDE.md # 実装ガイド
├── OPERATIONS_MANUAL.md    # 運用マニュアル
├── USER_GUIDE.md           # ユーザーガイド（新規）
├── API_REFERENCE.md        # APIリファレンス（新規）
└── specs/                  # 仕様書
    └── documentation/      # 本仕様
```

## 2. USER_GUIDE.md 設計

### 構成

9セクション構成で、初心者から上級者までカバー。

### 記述スタイル

- 説明形で記述（例: 「初期化できます」）
- コマンド例は必ず出力例も併記
- 注意点は Note: で明示

## 3. API_REFERENCE.md 設計

### 構成

5セクション構成で、全18コマンドをカバー。

### コマンドリファレンスフォーマット

各コマンドは統一フォーマットで記述:
- 構文
- 説明
- 引数
- オプション
- 使用例
- 出力例
- 終了ステータス

## 4. 相互参照

- USER_GUIDE.md: 概念的な説明と使用例
- API_REFERENCE.md: 詳細仕様と技術情報
- 両ドキュメントは相互補完的な関係
