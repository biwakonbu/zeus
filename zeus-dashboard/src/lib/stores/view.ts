import { writable, get } from 'svelte/store';
import type { ViewType } from '$lib/viewer';

// 現在のビュー状態を管理するストア
export const currentView = writable<ViewType>('usecase');

// ビュー間遷移時の自動選択用
export interface PendingNavigation {
	view: ViewType;
	entityType?: 'actor' | 'usecase' | 'activity' | 'objective';
	entityId?: string;
}

export const pendingNavigation = writable<PendingNavigation | null>(null);

// エンティティを指定してビューに遷移
export function navigateToEntity(
	view: ViewType,
	entityType: 'actor' | 'usecase' | 'activity' | 'objective',
	entityId: string
): void {
	pendingNavigation.set({ view, entityType, entityId });
	currentView.set(view);
}

// 遷移完了後にクリア
export function clearPendingNavigation(): void {
	pendingNavigation.set(null);
}

// ビューを変更する関数
export function setView(view: ViewType): void {
	currentView.set(view);
}

// UseCase ビュー用の状態
export interface UseCaseViewState {
	// 表示情報
	zoom: number;
	boundary: string;
	actorCount: number;
	usecaseCount: number;
	// リストパネル表示状態
	showListPanel: boolean;
	// コールバック（エンジンへの操作）
	onZoomIn?: () => void;
	onZoomOut?: () => void;
	onZoomReset?: () => void;
	onToggleListPanel?: () => void;
}

const defaultUseCaseViewState: UseCaseViewState = {
	zoom: 1.0,
	boundary: 'System',
	actorCount: 0,
	usecaseCount: 0,
	showListPanel: true
};

export const usecaseViewState = writable<UseCaseViewState>(defaultUseCaseViewState);

// UseCase ビュー状態を更新
export function updateUseCaseViewState(partial: Partial<UseCaseViewState>): void {
	usecaseViewState.update((state) => ({ ...state, ...partial }));
}

// UseCase ビュー状態をリセット
export function resetUseCaseViewState(): void {
	usecaseViewState.set(defaultUseCaseViewState);
}

// UseCase ビュー状態を取得
export function getUseCaseViewState(): UseCaseViewState {
	return get(usecaseViewState);
}

// Graph ビュー用の状態
export interface GraphViewState {
	// 表示情報
	zoom: number;
	nodeCount: number;
	visibleCount: number;
	// 機能トグル状態
	showListPanel: boolean;
	showFilterPanel: boolean;
	showLegend: boolean;
	// 依存関係フィルター
	hasDependencyFilter: boolean;
	dependencyFilterNodeId: string | null;
	// コールバック（エンジンへの操作）
	onZoomIn?: () => void;
	onZoomOut?: () => void;
	onZoomReset?: () => void;
	onToggleListPanel?: () => void;
	onToggleFilterPanel?: () => void;
	onToggleLegend?: () => void;
	onClearDependencyFilter?: () => void;
}

const defaultGraphViewState: GraphViewState = {
	zoom: 1.0,
	nodeCount: 0,
	visibleCount: 0,
	showListPanel: true,
	showFilterPanel: true,
	showLegend: true,
	hasDependencyFilter: false,
	dependencyFilterNodeId: null
};

export const graphViewState = writable<GraphViewState>(defaultGraphViewState);

// Graph ビュー状態を更新
export function updateGraphViewState(partial: Partial<GraphViewState>): void {
	graphViewState.update((state) => ({ ...state, ...partial }));
}

// Graph ビュー状態をリセット
export function resetGraphViewState(): void {
	graphViewState.set(defaultGraphViewState);
}

// Graph ビュー状態を取得
export function getGraphViewState(): GraphViewState {
	return get(graphViewState);
}

// Activity ビュー用の状態
export interface ActivityViewState {
	// 表示情報
	zoom: number;
	activityCount: number;
	// 選択中のアクティビティ
	selectedActivityId: string | null;
	// リストパネル表示状態
	showListPanel: boolean;
	// コールバック（エンジンへの操作）
	onZoomIn?: () => void;
	onZoomOut?: () => void;
	onZoomReset?: () => void;
	onToggleListPanel?: () => void;
}

const defaultActivityViewState: ActivityViewState = {
	zoom: 1.0,
	activityCount: 0,
	selectedActivityId: null,
	showListPanel: true
};

export const activityViewState = writable<ActivityViewState>(defaultActivityViewState);

// Activity ビュー状態を更新
export function updateActivityViewState(partial: Partial<ActivityViewState>): void {
	activityViewState.update((state) => ({ ...state, ...partial }));
}

// Activity ビュー状態をリセット
export function resetActivityViewState(): void {
	activityViewState.set(defaultActivityViewState);
}

// Activity ビュー状態を取得
export function getActivityViewState(): ActivityViewState {
	return get(activityViewState);
}

// Vision ビュー用の状態
export interface VisionViewState {
	objectiveCount: number;
	selectedObjectiveId: string | null;
	showListPanel: boolean;
}

const defaultVisionViewState: VisionViewState = {
	objectiveCount: 0,
	selectedObjectiveId: null,
	showListPanel: true
};

export const visionViewState = writable<VisionViewState>(defaultVisionViewState);

// Vision ビュー状態を更新
export function updateVisionViewState(partial: Partial<VisionViewState>): void {
	visionViewState.update((state) => ({ ...state, ...partial }));
}

// Vision ビュー状態をリセット
export function resetVisionViewState(): void {
	visionViewState.set(defaultVisionViewState);
}
