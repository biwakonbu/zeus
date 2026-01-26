# 設計書 - UseCase/Actor UI 改善

## 1. コンポーネント設計

### 1.1 コンポーネント階層

```
UseCaseView.svelte
├── UseCaseListPanel.svelte
│   ├── SegmentedTabs.svelte
│   ├── SearchInput.svelte
│   ├── FilterDropdown.svelte
│   └── GroupedList.svelte
├── [Canvas Area]
└── UseCaseViewPanel.svelte
```

### 1.2 新規コンポーネント

#### SegmentedTabs.svelte

タブ切り替え UI コンポーネント。

**Props**:
```typescript
interface Props {
  tabs: Array<{ id: string; label: string; count: number }>;
  activeTab: string;
  onTabChange: (tabId: string) => void;
}
```

**特徴**:
- role="tablist" / role="tab" でアクセシビリティ対応
- 件数をリアルタイム表示
- Factorio 風セグメントボタンスタイル

---

#### SearchInput.svelte

検索入力コンポーネント。

**Props**:
```typescript
interface Props {
  value: string;
  placeholder?: string;
  onInput: (value: string) => void;
  onClear?: () => void;
}
```

**特徴**:
- Search アイコン（左側）
- X クリアボタン（右側、入力時のみ表示）
- aria-label 対応

---

#### FilterDropdown.svelte

ドロップダウン選択コンポーネント。

**Props**:
```typescript
interface Props {
  options: Array<{ id: string; label: string }>;
  selected: string | null;
  placeholder: string;
  onSelect: (id: string | null) => void;
}
```

**特徴**:
- 外部クリックで閉じる
- ESC キーで閉じる
- ChevronDown/ChevronUp アイコン切り替え

---

#### GroupedList.svelte

グループ化リスト表示コンポーネント。

**Props**:
```typescript
type ListItem = (ActorItem & { itemType: 'actor' }) | (UseCaseItem & { itemType: 'usecase' });

interface Props {
  items: ListItem[];
  groupBy: boolean;
  selectedId: string | null;
  actors?: ActorItem[];
  onSelect: (item: ListItem) => void;
}
```

**特徴**:
- グループヘッダーに件数表示
- 選択状態をハイライト
- UseCase に関連 Actor 名をテキスト表示

---

#### UseCaseListPanel.svelte

統合リストパネル。上記コンポーネントを組み合わせて使用。

**Props**:
```typescript
interface Props {
  actors: ActorItem[];
  usecases: UseCaseItem[];
  selectedActorId: string | null;
  selectedUseCaseId: string | null;
  onActorSelect: (actor: ActorItem) => void;
  onUseCaseSelect: (usecase: UseCaseItem) => void;
}
```

**内部状態**:
```typescript
let activeTab = $state<'all' | 'actor' | 'usecase'>('all');
let searchQuery = $state('');
let debouncedQuery = $state('');
let filterActorId = $state<string | null>(null);
```

---

## 2. 状態管理設計

### 2.1 状態の所在

| 状態 | 所在 | 理由 |
|------|------|------|
| activeTab | UseCaseListPanel | パネル内で完結 |
| searchQuery | UseCaseListPanel | パネル内で完結 |
| filterActorId | UseCaseListPanel | パネル内で完結 |
| selectedActorId | UseCaseView | 複数コンポーネントで共有 |
| selectedUseCaseId | UseCaseView | 複数コンポーネントで共有 |
| data | UseCaseView | API データ |

### 2.2 debounce 実装

```typescript
let debounceTimer: ReturnType<typeof setTimeout>;

function handleSearch(value: string) {
  searchQuery = value;
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    debouncedQuery = value;
  }, 250);
}

function handleTabChange(tabId: string) {
  activeTab = tabId;
  searchQuery = '';
  debouncedQuery = '';
  filterActorId = null;
}
```

---

## 3. レイアウト設計

### 3.1 グリッド構成

```css
.usecase-layout {
  display: grid;
  grid-template-columns: 280px 1fr 280px;
  height: 100%;
}
```

### 3.2 パネル内部レイアウト

```
+------------------------------------------+
| [全て (13)] [Actor (3)] [UseCase (10)]   |  SegmentedTabs
+------------------------------------------+
| [Search] 検索...                    [X]  |  SearchInput
+------------------------------------------+
| [関連 Actor: 全て                    v]  |  FilterDropdown
+------------------------------------------+
| Actor (3)                                |  GroupedList
|   [Icon] ユーザー                        |
|   [Icon] 外部システム                    |
| UseCase (10)                             |
|   [Dot] ログイン機能                     |
|         関連: ユーザー, 管理者           |
+------------------------------------------+
```

### 3.3 レスポンシブ対応

**768px 以下**:
```css
@media (max-width: 768px) {
  .usecase-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr auto;
  }

  .list-panel {
    max-height: 200px;
  }
}
```

---

## 4. スタイル設計

### 4.1 CSS 変数使用

```css
/* タブ */
.tab {
  background: var(--bg-secondary);
  border: 1px solid var(--border-metal);
  color: var(--text-secondary);
}

.tab.active {
  background: var(--accent-primary);
  color: var(--bg-primary);
}

/* グループヘッダー */
.group-header {
  background: var(--bg-secondary);
  color: var(--accent-primary);
  font-weight: 600;
}

/* リストアイテム */
.list-item.selected {
  background: var(--accent-primary);
  color: var(--bg-primary);
}
```

### 4.2 アイコンマッピング

| Actor Type | Lucide Icon |
|------------|-------------|
| human | User |
| system | Server |
| time | Clock |
| device | Smartphone |
| external | Globe |

---

## 5. ファイル構成

```
zeus-dashboard/src/lib/viewer/usecase/
├── UseCaseView.svelte
├── UseCaseViewPanel.svelte
├── UseCaseListPanel.svelte
├── index.ts
├── components/
│   ├── SegmentedTabs.svelte
│   ├── SearchInput.svelte
│   ├── FilterDropdown.svelte
│   └── GroupedList.svelte
├── engine/
│   └── UseCaseEngine.ts
└── rendering/
    └── ...
```

---

## 6. 型定義

```typescript
type TabId = 'all' | 'actor' | 'usecase';

interface TabItem {
  id: TabId;
  label: string;
  count: number;
}

interface ActorListItem extends ActorItem {
  itemType: 'actor';
}

interface UseCaseListItem extends UseCaseItem {
  itemType: 'usecase';
}

type ListItem = ActorListItem | UseCaseListItem;
```

---

## 7. テスト観点

| 観点 | 内容 |
|------|------|
| タブ切り替え | 各タブで正しいデータが表示される |
| 検索 | 部分一致検索が機能する |
| フィルタ | 関連 Actor での絞り込みが機能する |
| 選択 | クリックで選択、詳細パネルに反映 |
| レスポンシブ | 768px 以下でレイアウトが切り替わる |
| アクセシビリティ | Tab キーでフォーカス移動、aria 属性が適切 |
