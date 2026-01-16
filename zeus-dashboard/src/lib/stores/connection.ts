import { writable } from 'svelte/store';
import type { ConnectionState } from '$lib/types/api';

// 接続状態を管理するストア
export const connectionState = writable<ConnectionState>('disconnected');

// 接続状態を更新するヘルパー関数
export function setConnected() {
	connectionState.set('connected');
}

export function setDisconnected() {
	connectionState.set('disconnected');
}

export function setConnecting() {
	connectionState.set('connecting');
}
