// キーボードナビゲーション Store
// アプリケーション全体のキーボードショートカットを管理

import { writable, derived, get } from 'svelte/store';

export interface KeyboardShortcut {
	key: string;
	modifiers?: ('ctrl' | 'cmd' | 'alt' | 'shift')[];
	description: string;
	category: string;
	action: () => void;
}

interface KeyboardState {
	shortcuts: Map<string, KeyboardShortcut>;
	helpVisible: boolean;
	enabled: boolean;
}

function createKeyboardStore() {
	const { subscribe, update } = writable<KeyboardState>({
		shortcuts: new Map(),
		helpVisible: false,
		enabled: true
	});

	// ショートカットキーを正規化
	function normalizeKey(
		key: string,
		modifiers: ('ctrl' | 'cmd' | 'alt' | 'shift')[] = []
	): string {
		const sortedMods = [...modifiers].sort();
		return [...sortedMods, key.toLowerCase()].join('+');
	}

	// ショートカット登録
	function register(shortcut: KeyboardShortcut): () => void {
		const normalizedKey = normalizeKey(shortcut.key, shortcut.modifiers);

		update((state) => {
			const newShortcuts = new Map(state.shortcuts);
			newShortcuts.set(normalizedKey, shortcut);
			return { ...state, shortcuts: newShortcuts };
		});

		// 登録解除関数を返す
		return () => {
			update((state) => {
				const newShortcuts = new Map(state.shortcuts);
				newShortcuts.delete(normalizedKey);
				return { ...state, shortcuts: newShortcuts };
			});
		};
	}

	// イベントからキーを生成
	function getKeyFromEvent(event: KeyboardEvent): string {
		const modifiers: ('ctrl' | 'cmd' | 'alt' | 'shift')[] = [];

		if (event.ctrlKey) modifiers.push('ctrl');
		if (event.metaKey) modifiers.push('cmd');
		if (event.altKey) modifiers.push('alt');
		if (event.shiftKey) modifiers.push('shift');

		return normalizeKey(event.key, modifiers);
	}

	// キーイベント処理
	function handleKeydown(event: KeyboardEvent): boolean {
		const currentState = get({ subscribe });

		if (!currentState.enabled) return false;

		const key = getKeyFromEvent(event);
		const shortcut = currentState.shortcuts.get(key);

		if (shortcut) {
			event.preventDefault();
			shortcut.action();
			return true;
		}

		return false;
	}

	// ヘルプ表示切り替え
	function toggleHelp(): void {
		update((state) => ({ ...state, helpVisible: !state.helpVisible }));
	}

	function showHelp(): void {
		update((state) => ({ ...state, helpVisible: true }));
	}

	function hideHelp(): void {
		update((state) => ({ ...state, helpVisible: false }));
	}

	// 有効/無効切り替え
	function enable(): void {
		update((state) => ({ ...state, enabled: true }));
	}

	function disable(): void {
		update((state) => ({ ...state, enabled: false }));
	}

	return {
		subscribe,
		register,
		handleKeydown,
		toggleHelp,
		showHelp,
		hideHelp,
		enable,
		disable
	};
}

export const keyboardStore = createKeyboardStore();

// ショートカット一覧を派生ストアとして提供
export const shortcutsList = derived(keyboardStore, ($store) => {
	const shortcuts = Array.from($store.shortcuts.values());

	// カテゴリごとにグループ化
	const grouped = shortcuts.reduce(
		(acc, shortcut) => {
			if (!acc[shortcut.category]) {
				acc[shortcut.category] = [];
			}
			acc[shortcut.category].push(shortcut);
			return acc;
		},
		{} as Record<string, KeyboardShortcut[]>
	);

	return grouped;
});

// 表示用にキーをフォーマット
export function formatShortcutKey(
	key: string,
	modifiers: ('ctrl' | 'cmd' | 'alt' | 'shift')[] = []
): string {
	const isMac =
		typeof navigator !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0;

	const modLabels: Record<string, string> = {
		ctrl: isMac ? '⌃' : 'Ctrl',
		cmd: isMac ? '⌘' : 'Ctrl',
		alt: isMac ? '⌥' : 'Alt',
		shift: isMac ? '⇧' : 'Shift'
	};

	const keyLabel = key.length === 1 ? key.toUpperCase() : key;
	const modParts = modifiers.map((m) => modLabels[m]);

	return [...modParts, keyLabel].join(isMac ? '' : '+');
}
