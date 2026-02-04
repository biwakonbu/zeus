// SSE (Server-Sent Events) クライアント
import type { SSEEventType } from '$lib/types/api';
import { setConnected, setDisconnected, setConnecting } from '$lib/stores/connection';
import { setStatus } from '$lib/stores/status';
import { setTasks } from '$lib/stores/tasks';

// SSE イベントハンドラー型
type SSEEventHandler = (data: unknown) => void;

// SSE クライアント設定
interface SSEClientOptions {
	url: string;
	reconnectDelay?: number;
	maxReconnectAttempts?: number;
}

// SSE クライアントクラス
export class SSEClient {
	private eventSource: EventSource | null = null;
	private url: string;
	private reconnectDelay: number;
	private maxReconnectAttempts: number;
	private reconnectAttempts: number = 0;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private handlers: Map<SSEEventType, SSEEventHandler[]> = new Map();
	private isManualClose: boolean = false;

	constructor(options: SSEClientOptions) {
		this.url = options.url;
		this.reconnectDelay = options.reconnectDelay ?? 3000;
		this.maxReconnectAttempts = options.maxReconnectAttempts ?? 10;
	}

	// イベントハンドラーを登録
	on(eventType: SSEEventType, handler: SSEEventHandler): void {
		const handlers = this.handlers.get(eventType) ?? [];
		handlers.push(handler);
		this.handlers.set(eventType, handlers);
	}

	// イベントハンドラーを解除
	off(eventType: SSEEventType, handler: SSEEventHandler): void {
		const handlers = this.handlers.get(eventType) ?? [];
		const index = handlers.indexOf(handler);
		if (index !== -1) {
			handlers.splice(index, 1);
			this.handlers.set(eventType, handlers);
		}
	}

	// 接続を開始
	connect(): void {
		if (this.eventSource) {
			return;
		}

		this.isManualClose = false;
		setConnecting();

		try {
			this.eventSource = new EventSource(this.url);

			this.eventSource.onopen = () => {
				console.log('[SSE] Connected');
				setConnected();
				this.reconnectAttempts = 0;
			};

			this.eventSource.onerror = (event) => {
				console.error('[SSE] Error:', event);
				this.handleError();
			};

			// 各イベントタイプのリスナーを登録
			// Note: 'graph', 'prediction' は未使用のため削除済み
			const eventTypes: SSEEventType[] = ['status', 'task', 'approval'];

			eventTypes.forEach((eventType) => {
				this.eventSource!.addEventListener(eventType, (event: MessageEvent) => {
					try {
						const data = JSON.parse(event.data);
						this.dispatchEvent(eventType, data);
					} catch (err) {
						console.error(`[SSE] Failed to parse ${eventType} event:`, err);
					}
				});
			});

			// 接続確立イベント
			this.eventSource.addEventListener('connected', (event: MessageEvent) => {
				console.log('[SSE] Connection confirmed:', event.data);
			});
		} catch (err) {
			console.error('[SSE] Failed to create EventSource:', err);
			this.handleError();
		}
	}

	// 切断
	disconnect(): void {
		this.isManualClose = true;

		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}

		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}

		setDisconnected();
	}

	// イベントをディスパッチ
	private dispatchEvent(eventType: SSEEventType, data: unknown): void {
		const handlers = this.handlers.get(eventType) ?? [];
		handlers.forEach((handler) => {
			try {
				handler(data);
			} catch (err) {
				console.error(`[SSE] Handler error for ${eventType}:`, err);
			}
		});
	}

	// エラーハンドリングと再接続
	private handleError(): void {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}

		setDisconnected();

		if (this.isManualClose) {
			return;
		}

		if (this.reconnectAttempts < this.maxReconnectAttempts) {
			this.reconnectAttempts++;
			const delay = this.reconnectDelay * Math.min(this.reconnectAttempts, 5);

			console.log(`[SSE] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

			this.reconnectTimer = setTimeout(() => {
				this.connect();
			}, delay);
		} else {
			console.error('[SSE] Max reconnect attempts reached, giving up');
			// カスタムイベントを発火してポーリングへのフォールバックを通知
			window.dispatchEvent(new CustomEvent('sse-failed'));
		}
	}

	// 接続状態を取得
	get isConnected(): boolean {
		return this.eventSource?.readyState === EventSource.OPEN;
	}
}

// デフォルト SSE クライアントを作成
export function createSSEClient(): SSEClient {
	const client = new SSEClient({
		url: '/api/events',
		reconnectDelay: 3000,
		maxReconnectAttempts: 10
	});

	// Store と連携するハンドラーを登録
	client.on('status', (data) => {
		setStatus(data as Parameters<typeof setStatus>[0]);
	});

	client.on('task', (data) => {
		setTasks(data as Parameters<typeof setTasks>[0]);
	});

	// Note: graph, prediction ハンドラーは未使用のため削除
	// SSE イベントは引き続き受信されるが、Store 更新は行わない

	return client;
}

// シングルトンインスタンス
let sseClient: SSEClient | null = null;

// SSE クライアントを取得または作成
export function getSSEClient(): SSEClient {
	if (!sseClient) {
		sseClient = createSSEClient();
	}
	return sseClient;
}

// SSE 接続を開始
export function connectSSE(): void {
	getSSEClient().connect();
}

// SSE 接続を切断
export function disconnectSSE(): void {
	if (sseClient) {
		sseClient.disconnect();
		sseClient = null;
	}
}
