# Drill-Down Mode 仕様

エンティティをダブルクリックした際に遷移する詳細表示モード。

## 仕様

| 項目 | 値 |
|------|-----|
| 遷移方法 | ダブルクリック / Enter キー |
| URL パターン | `/dashboard/entity/{entityId}` |
| 戻り操作 | Escape キー / Back ボタン |
| 状態復元 | スクロール位置、ズームレベル、アクティブビュー |

## 状態管理

```typescript
// zeus-dashboard/src/lib/stores/drillDown.ts

export interface DrillDownState {
  /** 戻り先 URL */
  returnUrl: string;
  /** ビューの状態 */
  viewState: {
    scrollX?: number;
    scrollY?: number;
    zoomLevel?: number;
    activeView?: string;
  };
}

// 遷移前に保存
export function saveDrillDownState(state: DrillDownState): void;

// 戻り時に復元してクリア
export function restoreDrillDownState(): DrillDownState | null;
```

## 遷移フロー

```
1. ユーザーがエンティティをダブルクリック
   ↓
2. saveDrillDownState() で現在状態を保存
   ↓
3. goto(`/dashboard/entity/${entityId}`) で遷移
   ↓
4. 詳細ページを表示
   ↓
5. ユーザーが Escape または Back ボタン
   ↓
6. restoreDrillDownState() で状態を取得
   ↓
7. 元の URL に遷移 + スクロール位置復元
```

## 詳細ページの構成

1. **ヘッダー**
   - Back ボタン
   - Escape ショートカット表示

2. **エンティティヘッダー**
   - タイプバッジ
   - ID
   - タイトル

3. **統計カード**
   - Progress（プログレスバー）
   - Status

4. **説明セクション**
   - Description

5. **関連エンティティ**
   - Related Entities リスト

6. **履歴**
   - History タイムライン

## 使用方法

```svelte
<script>
  import { goto } from '$app/navigation';
  import { saveDrillDownState } from '$lib/stores/drillDown';

  function handleNodeDblClick(nodeId: string, nodeType: string) {
    saveDrillDownState({
      returnUrl: window.location.pathname,
      viewState: {
        scrollX: window.scrollX,
        scrollY: window.scrollY,
        activeView: 'health'
      }
    });

    goto(`/dashboard/entity/${nodeId}`);
  }
</script>
```

## キーボード操作

| キー | 動作 |
|------|------|
| Escape | 元の画面に戻る |
| Tab | フォーカス移動 |

## レスポンシブ対応

- デスクトップ: 2カラム統計カード
- モバイル（768px 以下）: 1カラム、パディング縮小
