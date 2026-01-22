# Design Tokens

WBS Dashboard の統一されたデザイン値を TypeScript と CSS 変数で管理する。

## TypeScript Tokens

```typescript
// zeus-dashboard/src/lib/theme/design-tokens.ts
export const tokens = {
  colors: {
    accent: {
      primary: '#ff9533',
      glow: 'rgba(255, 149, 51, 0.15)'
    },
    status: {
      completed: '#44cc44',
      inProgress: '#4488ff',
      notStarted: '#666666',
      blocked: '#ee4444',
      warning: '#ffcc00'
    },
    background: {
      dark: '#1a1a1a',
      mid: '#2d2d2d',
      elevated: '#3a3a3a'
    },
    border: {
      subtle: '#333333',
      mid: '#4a4a4a',
      strong: '#5a5a5a'
    }
  },
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '24px'
  },
  radius: {
    sm: '2px',
    md: '4px',
    lg: '8px'
  },
  shadows: {
    tooltip: '0 4px 20px rgba(0, 0, 0, 0.5)',
    card: '0 2px 8px rgba(0, 0, 0, 0.3)',
    glow: '0 0 15px rgba(255, 149, 51, 0.3)',
    inset: 'inset 0 1px 0 rgba(255, 255, 255, 0.03)'
  },
  timing: {
    tooltipDelay: 500,
    transitionFast: 100,
    transitionNormal: 200,
    transitionSlow: 300
  },
  tooltip: {
    width: 320,
    height: 220,
    offset: 16
  }
} as const;
```

## CSS Variables

```css
/* zeus-dashboard/src/lib/theme/variables.css */
:root {
  /* Tooltip */
  --tooltip-width: 320px;
  --tooltip-height: 220px;
  --tooltip-delay: 500ms;
  --shadow-tooltip: 0 4px 20px rgba(0, 0, 0, 0.5);

  /* Factorio カラー */
  --factorio-orange: #ff9533;
  --factorio-green: #44cc44;
  --factorio-yellow: #ffcc00;
  --factorio-red: #ee4444;
  --factorio-blue: #4488ff;
  --factorio-bg-dark: #1a1a1a;
  --factorio-bg-mid: #2d2d2d;
  --factorio-border-mid: #4a4a4a;
}
```

## 使用方法

### TypeScript

```typescript
import { tokens } from '$lib/theme/design-tokens';

// タイマー値
setTimeout(() => showTooltip(), tokens.timing.tooltipDelay);

// サイズ
const { width, height } = tokens.tooltip;
```

### CSS

```css
.tooltip {
  width: var(--tooltip-width);
  box-shadow: var(--shadow-tooltip);
}
```

## ヘルパー関数

```typescript
// ステータスから色を取得
export function getStatusColor(status: string): string {
  switch (status) {
    case 'completed': return tokens.colors.status.completed;
    case 'in_progress': return tokens.colors.status.inProgress;
    case 'blocked': return tokens.colors.status.blocked;
    default: return tokens.colors.status.notStarted;
  }
}

// エンティティタイプからアイコン名を取得
export function getEntityIcon(type: EntityType): string {
  switch (type) {
    case 'vision': return 'Target';
    case 'objective': return 'Flag';
    case 'deliverable': return 'Package';
    case 'task': return 'CheckSquare';
    default: return 'Circle';
  }
}
```
