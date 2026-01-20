// Toast 通知システム Store
// アプリケーション全体で使用可能なトースト通知を管理

import { writable } from 'svelte/store';

export type ToastType = 'info' | 'success' | 'warning' | 'error';

export interface Toast {
	id: string;
	type: ToastType;
	message: string;
	duration: number;
	dismissible: boolean;
}

interface ToastStore {
	toasts: Toast[];
}

function createToastStore() {
	const { subscribe, update } = writable<ToastStore>({ toasts: [] });

	let counter = 0;

	function generateId(): string {
		return `toast-${Date.now()}-${++counter}`;
	}

	function add(
		message: string,
		type: ToastType = 'info',
		options: { duration?: number; dismissible?: boolean } = {}
	): string {
		const id = generateId();
		const duration = options.duration ?? (type === 'error' ? 8000 : 4000);
		const dismissible = options.dismissible ?? true;

		const toast: Toast = {
			id,
			type,
			message,
			duration,
			dismissible
		};

		update((state) => ({
			toasts: [...state.toasts, toast]
		}));

		// 自動削除タイマー
		if (duration > 0) {
			setTimeout(() => {
				remove(id);
			}, duration);
		}

		return id;
	}

	function remove(id: string): void {
		update((state) => ({
			toasts: state.toasts.filter((t) => t.id !== id)
		}));
	}

	function clear(): void {
		update(() => ({ toasts: [] }));
	}

	// 便利メソッド
	function info(message: string, options?: { duration?: number; dismissible?: boolean }): string {
		return add(message, 'info', options);
	}

	function success(
		message: string,
		options?: { duration?: number; dismissible?: boolean }
	): string {
		return add(message, 'success', options);
	}

	function warning(
		message: string,
		options?: { duration?: number; dismissible?: boolean }
	): string {
		return add(message, 'warning', options);
	}

	function error(message: string, options?: { duration?: number; dismissible?: boolean }): string {
		return add(message, 'error', options);
	}

	return {
		subscribe,
		add,
		remove,
		clear,
		info,
		success,
		warning,
		error
	};
}

export const toastStore = createToastStore();
