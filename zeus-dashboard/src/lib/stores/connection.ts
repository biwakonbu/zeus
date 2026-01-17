import { writable } from 'svelte/store';
import type { ConnectionState } from '$lib/types/api';

// 接続状態を管理するストア
export const connectionState = writable<ConnectionState>('disconnected');

// 接続状態を更新するヘルパー関数
export function setConnected() {
	connectionState.set('connected');
	updateGlobalState('connected');
}

export function setDisconnected() {
	connectionState.set('disconnected');
	updateGlobalState('disconnected');
}

export function setConnecting() {
	connectionState.set('connecting');
	updateGlobalState('connecting');
}

// E2E テスト用: 開発環境でのみグローバルに接続状態を公開
function updateGlobalState(state: ConnectionState) {
	if (typeof window !== 'undefined' && import.meta.env.DEV) {
		(window as Window & { __CONNECTION_STATE__?: ConnectionState }).__CONNECTION_STATE__ = state;
	}
}
