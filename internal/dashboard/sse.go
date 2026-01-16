// Package dashboard - SSE (Server-Sent Events) Broadcaster 実装
package dashboard

import (
	"encoding/json"
	"sync"
)

// EventType は SSE イベントの種類
type EventType string

const (
	EventStatus     EventType = "status"
	EventTask       EventType = "task"
	EventApproval   EventType = "approval"
	EventGraph      EventType = "graph"
	EventPrediction EventType = "prediction"
)

// SSEEvent は SSE で送信するイベント
type SSEEvent struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// SSEClient は SSE クライアント接続を表す
type SSEClient struct {
	ID     string
	Events chan SSEEvent
}

// SSEBroadcaster は複数クライアントへの SSE 配信を管理
type SSEBroadcaster struct {
	clients map[string]*SSEClient
	mu      sync.RWMutex
}

// NewSSEBroadcaster は新しい SSEBroadcaster を作成
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[string]*SSEClient),
	}
}

// AddClient はクライアントを追加
func (b *SSEBroadcaster) AddClient(id string) *SSEClient {
	b.mu.Lock()
	defer b.mu.Unlock()

	client := &SSEClient{
		ID:     id,
		Events: make(chan SSEEvent, 10),
	}
	b.clients[id] = client
	return client
}

// RemoveClient はクライアントを削除
func (b *SSEBroadcaster) RemoveClient(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if client, ok := b.clients[id]; ok {
		close(client.Events)
		delete(b.clients, id)
	}
}

// ClientCount はアクティブなクライアント数を返す
func (b *SSEBroadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// Broadcast は全クライアントにイベントを配信
func (b *SSEBroadcaster) Broadcast(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, client := range b.clients {
		select {
		case client.Events <- event:
			// 送信成功
		default:
			// バッファフル - スキップ（クライアントが遅い場合）
		}
	}
}

// BroadcastStatus はステータス更新を配信
func (b *SSEBroadcaster) BroadcastStatus(data interface{}) {
	b.Broadcast(SSEEvent{
		Type: EventStatus,
		Data: data,
	})
}

// BroadcastTask はタスク更新を配信
func (b *SSEBroadcaster) BroadcastTask(data interface{}) {
	b.Broadcast(SSEEvent{
		Type: EventTask,
		Data: data,
	})
}

// BroadcastGraph はグラフ更新を配信
func (b *SSEBroadcaster) BroadcastGraph(data interface{}) {
	b.Broadcast(SSEEvent{
		Type: EventGraph,
		Data: data,
	})
}

// BroadcastPrediction は予測更新を配信
func (b *SSEBroadcaster) BroadcastPrediction(data interface{}) {
	b.Broadcast(SSEEvent{
		Type: EventPrediction,
		Data: data,
	})
}

// FormatSSEMessage は SSE メッセージ形式にフォーマット
func FormatSSEMessage(event SSEEvent) ([]byte, error) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
