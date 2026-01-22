# Rich Tooltip 仕様

エンティティをホバーした際に表示されるリッチな情報カード。

## 仕様

| 項目 | 値 |
|------|-----|
| サイズ | 320px x 220px（固定） |
| 表示遅延 | 500ms |
| 表示位置 | カーソル右下（画面端で自動調整） |
| アニメーション | fly({ y: 10, duration: 100 }) |

## 表示内容

1. **ヘッダー**: エンティティタイプアイコン + タイプ名
2. **ID**: エンティティ ID（例: OBJ-001）
3. **タイトル**: エンティティ名（2行まで、超過時は省略）
4. **プログレスバー**: 進捗率
5. **ステータス**: ステータス名（色分け）
6. **更新日時**: 最終更新日時（オプション）
7. **ヒント**: "Double-click to view details"

## TypeScript Interface

```typescript
export interface TooltipEntity {
  id: string;
  title: string;
  type: 'vision' | 'objective' | 'deliverable' | 'task';
  status: string;
  progress: number;
  lastUpdate?: string;
}

interface Props {
  visible: boolean;
  entity: TooltipEntity | null;
  position: { x: number; y: number };
}
```

## 位置調整ロジック

```typescript
function calculatePosition(x: number, y: number): { x: number; y: number } {
  const OFFSET = 16;
  let newX = x + OFFSET;
  let newY = y + OFFSET;

  // 右端でフリップ
  if (newX + 320 > window.innerWidth) {
    newX = x - 320 - OFFSET;
  }

  // 下端でフリップ
  if (newY + 220 > window.innerHeight) {
    newY = y - 220 - OFFSET;
  }

  return { x: newX, y: newY };
}
```

## 使用方法

```svelte
<script>
  import { RichTooltip } from '$lib/components/ui';

  let tooltipVisible = $state(false);
  let tooltipEntity = $state(null);
  let tooltipPosition = $state({ x: 0, y: 0 });
</script>

<RichTooltip
  visible={tooltipVisible}
  entity={tooltipEntity}
  position={tooltipPosition}
/>
```

## アクセシビリティ

- `role="tooltip"` 属性
- `aria-live="polite"` で読み上げ対応
- `pointer-events: none` でマウス操作を妨げない
