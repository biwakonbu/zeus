package dashboard

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// ===== NewSSEBroadcaster テスト =====

func TestNewSSEBroadcaster(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	if broadcaster == nil {
		t.Fatal("NewSSEBroadcaster returned nil")
	}
	if broadcaster.clients == nil {
		t.Error("clients map should be initialized")
	}
	if broadcaster.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", broadcaster.ClientCount())
	}
}

// ===== AddClient テスト =====

func TestSSEBroadcaster_AddClient(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	client := broadcaster.AddClient("client-001")

	if client == nil {
		t.Fatal("AddClient returned nil")
	}
	if client.ID != "client-001" {
		t.Errorf("expected client ID 'client-001', got '%s'", client.ID)
	}
	if client.Events == nil {
		t.Error("Events channel should be initialized")
	}
	if broadcaster.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", broadcaster.ClientCount())
	}
}

func TestSSEBroadcaster_AddClient_Multiple(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	broadcaster.AddClient("client-001")
	broadcaster.AddClient("client-002")
	broadcaster.AddClient("client-003")

	if broadcaster.ClientCount() != 3 {
		t.Errorf("expected 3 clients, got %d", broadcaster.ClientCount())
	}
}

func TestSSEBroadcaster_AddClient_SameID(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	broadcaster.AddClient("client-001")
	broadcaster.AddClient("client-001") // 同じ ID で上書き

	// 同じ ID は上書きされるため 1 クライアント
	if broadcaster.ClientCount() != 1 {
		t.Errorf("expected 1 client (same ID overwrites), got %d", broadcaster.ClientCount())
	}
}

// ===== RemoveClient テスト =====

func TestSSEBroadcaster_RemoveClient(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	broadcaster.AddClient("client-001")

	broadcaster.RemoveClient("client-001")

	if broadcaster.ClientCount() != 0 {
		t.Errorf("expected 0 clients after removal, got %d", broadcaster.ClientCount())
	}
}

func TestSSEBroadcaster_RemoveClient_NotFound(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	// 存在しないクライアントの削除（パニックしないこと）
	broadcaster.RemoveClient("nonexistent")

	if broadcaster.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", broadcaster.ClientCount())
	}
}

func TestSSEBroadcaster_RemoveClient_ClosesChannel(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	broadcaster.RemoveClient("client-001")

	// チャネルがクローズされていることを確認
	_, ok := <-client.Events
	if ok {
		t.Error("Events channel should be closed after removal")
	}
}

// ===== ClientCount テスト =====

func TestSSEBroadcaster_ClientCount(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	// 0 クライアント
	if broadcaster.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", broadcaster.ClientCount())
	}

	// 2 クライアント追加
	broadcaster.AddClient("client-001")
	broadcaster.AddClient("client-002")
	if broadcaster.ClientCount() != 2 {
		t.Errorf("expected 2 clients, got %d", broadcaster.ClientCount())
	}

	// 1 クライアント削除
	broadcaster.RemoveClient("client-001")
	if broadcaster.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", broadcaster.ClientCount())
	}
}

// ===== Broadcast テスト =====

func TestSSEBroadcaster_Broadcast(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	event := SSEEvent{
		Type: EventStatus,
		Data: map[string]string{"message": "test"},
	}

	broadcaster.Broadcast(event)

	// イベントを受信
	select {
	case received := <-client.Events:
		if received.Type != EventStatus {
			t.Errorf("expected event type 'status', got '%s'", received.Type)
		}
		data, ok := received.Data.(map[string]string)
		if !ok {
			t.Fatal("expected data to be map[string]string")
		}
		if data["message"] != "test" {
			t.Errorf("expected message 'test', got '%s'", data["message"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestSSEBroadcaster_Broadcast_MultipleClients(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client1 := broadcaster.AddClient("client-001")
	client2 := broadcaster.AddClient("client-002")
	client3 := broadcaster.AddClient("client-003")

	event := SSEEvent{
		Type: EventGraph,
		Data: "test data",
	}

	broadcaster.Broadcast(event)

	// 全クライアントがイベントを受信
	clients := []*SSEClient{client1, client2, client3}
	for i, client := range clients {
		select {
		case received := <-client.Events:
			if received.Type != EventGraph {
				t.Errorf("client %d: expected event type 'graph', got '%s'", i+1, received.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client %d: timeout waiting for event", i+1)
		}
	}
}

func TestSSEBroadcaster_Broadcast_NoClients(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	event := SSEEvent{
		Type: EventGraph,
		Data: "test",
	}

	// クライアントがいなくてもパニックしない
	broadcaster.Broadcast(event)
}

func TestSSEBroadcaster_Broadcast_BufferFull(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	// バッファ (10) を超えるイベントを送信
	for i := 0; i < 15; i++ {
		event := SSEEvent{
			Type: EventStatus,
			Data: i,
		}
		broadcaster.Broadcast(event)
	}

	// バッファサイズ分（10）は受信できる
	received := 0
	for i := 0; i < 15; i++ {
		select {
		case <-client.Events:
			received++
		default:
			// バッファが空
		}
	}

	// 10 イベントが受信される（バッファサイズ）
	if received != 10 {
		t.Errorf("expected 10 events (buffer size), got %d", received)
	}
}

// ===== BroadcastStatus テスト =====

func TestSSEBroadcaster_BroadcastStatus(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	data := map[string]string{"status": "healthy"}
	broadcaster.BroadcastStatus(data)

	select {
	case received := <-client.Events:
		if received.Type != EventStatus {
			t.Errorf("expected event type 'status', got '%s'", received.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for status event")
	}
}

// ===== BroadcastGraph テスト =====

func TestSSEBroadcaster_BroadcastGraph(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	data := map[string]interface{}{"nodes": 10, "edges": 5}
	broadcaster.BroadcastGraph(data)

	select {
	case received := <-client.Events:
		if received.Type != EventGraph {
			t.Errorf("expected event type 'graph', got '%s'", received.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for graph event")
	}
}

// ===== BroadcastPrediction テスト =====

func TestSSEBroadcaster_BroadcastPrediction(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	data := map[string]interface{}{"completion": 85.5}
	broadcaster.BroadcastPrediction(data)

	select {
	case received := <-client.Events:
		if received.Type != EventPrediction {
			t.Errorf("expected event type 'prediction', got '%s'", received.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for prediction event")
	}
}

// ===== FormatSSEMessage テスト =====

func TestFormatSSEMessage(t *testing.T) {
	event := SSEEvent{
		Type: EventStatus,
		Data: map[string]string{"message": "hello"},
	}

	data, err := FormatSSEMessage(event)
	if err != nil {
		t.Fatalf("FormatSSEMessage failed: %v", err)
	}

	// JSON として解析可能であることを確認
	var result map[string]string
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("expected message 'hello', got '%s'", result["message"])
	}
}

func TestFormatSSEMessage_ComplexData(t *testing.T) {
	event := SSEEvent{
		Type: EventGraph,
		Data: struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Progress int    `json:"progress"`
		}{
			ID:       "task-001",
			Title:    "タスク1",
			Progress: 75,
		},
	}

	data, err := FormatSSEMessage(event)
	if err != nil {
		t.Fatalf("FormatSSEMessage failed: %v", err)
	}

	// JSON として解析可能であることを確認
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["id"] != "task-001" {
		t.Errorf("expected id 'task-001', got '%v'", result["id"])
	}
}

func TestFormatSSEMessage_NilData(t *testing.T) {
	event := SSEEvent{
		Type: EventStatus,
		Data: nil,
	}

	data, err := FormatSSEMessage(event)
	if err != nil {
		t.Fatalf("FormatSSEMessage failed: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("expected 'null', got '%s'", string(data))
	}
}

// ===== EventType 値テスト =====

func TestEventType_Values(t *testing.T) {
	testCases := []struct {
		eventType EventType
		expected  string
	}{
		{EventStatus, "status"},
		{EventApproval, "approval"},
		{EventGraph, "graph"},
		{EventPrediction, "prediction"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.eventType) != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, string(tc.eventType))
			}
		})
	}
}

// ===== 並行アクセステスト =====

func TestSSEBroadcaster_Concurrent_AddRemove(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	var wg sync.WaitGroup

	// 並行でクライアントを追加
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := "client-" + string(rune('0'+id%10))
			broadcaster.AddClient(clientID)
		}(i)
	}

	wg.Wait()

	// クライアント数を確認（重複 ID があるため <= 100）
	count := broadcaster.ClientCount()
	if count == 0 {
		t.Error("expected at least some clients")
	}
	t.Logf("Final client count: %d", count)
}

func TestSSEBroadcaster_Concurrent_Broadcast(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	// 複数クライアントを追加
	clients := make([]*SSEClient, 10)
	for i := 0; i < 10; i++ {
		clients[i] = broadcaster.AddClient("client-" + string(rune('0'+i)))
	}

	var wg sync.WaitGroup

	// 並行でブロードキャスト
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := SSEEvent{
				Type: EventStatus,
				Data: idx,
			}
			broadcaster.Broadcast(event)
		}(i)
	}

	wg.Wait()

	// 各クライアントがイベントを受信したことを確認
	totalReceived := 0
	for _, client := range clients {
		for {
			select {
			case <-client.Events:
				totalReceived++
			default:
				goto nextClient
			}
		}
	nextClient:
	}

	// 少なくとも一部のイベントが受信されていること
	if totalReceived == 0 {
		t.Error("expected at least some events to be received")
	}
	t.Logf("Total events received across all clients: %d", totalReceived)
}

func TestSSEBroadcaster_Concurrent_AddBroadcastRemove(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	var wg sync.WaitGroup

	// 並行でクライアント追加、ブロードキャスト、削除
	for i := 0; i < 50; i++ {
		wg.Add(3)

		// 追加
		go func(id int) {
			defer wg.Done()
			broadcaster.AddClient("client-" + string(rune('a'+id%26)))
		}(i)

		// ブロードキャスト
		go func(idx int) {
			defer wg.Done()
			event := SSEEvent{
				Type: EventGraph,
				Data: idx,
			}
			broadcaster.Broadcast(event)
		}(i)

		// 削除
		go func(id int) {
			defer wg.Done()
			broadcaster.RemoveClient("client-" + string(rune('a'+id%26)))
		}(i)
	}

	wg.Wait()

	// パニックせずに完了することを確認
	t.Log("Concurrent add/broadcast/remove completed without panic")
}

// ===== SSEClient テスト =====

func TestSSEClient_EventsChannel(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	client := broadcaster.AddClient("client-001")

	// チャネルがバッファリングされていることを確認（10）
	for i := 0; i < 10; i++ {
		event := SSEEvent{Type: EventStatus, Data: i}
		select {
		case client.Events <- event:
			// OK
		default:
			t.Errorf("expected buffer capacity for event %d", i)
		}
	}

	// 11 番目はブロックされる（ノンブロッキングで確認）
	select {
	case client.Events <- SSEEvent{Type: EventStatus, Data: 10}:
		t.Error("expected buffer to be full")
	default:
		// OK - バッファがフル
	}
}

// ===== 複合シナリオテスト =====

func TestSSEBroadcaster_ComplexScenario(t *testing.T) {
	broadcaster := NewSSEBroadcaster()

	// 複数クライアントを追加
	client1 := broadcaster.AddClient("web-client")
	client2 := broadcaster.AddClient("mobile-client")

	// 様々なイベントをブロードキャスト
	broadcaster.BroadcastStatus(map[string]string{"status": "active"})
	broadcaster.BroadcastGraph(map[string]int{"nodes": 50, "edges": 75})
	broadcaster.BroadcastPrediction(map[string]float64{"completion": 0.75})

	// 各クライアントが 3 イベントを受信
	for _, client := range []*SSEClient{client1, client2} {
		receivedCount := 0
		for i := 0; i < 3; i++ {
			select {
			case <-client.Events:
				receivedCount++
			case <-time.After(100 * time.Millisecond):
				t.Errorf("client %s: timeout waiting for event %d", client.ID, i+1)
			}
		}
		if receivedCount != 3 {
			t.Errorf("client %s: expected 3 events, got %d", client.ID, receivedCount)
		}
	}

	// 1 クライアントを削除
	broadcaster.RemoveClient("web-client")

	// 残りのクライアントにブロードキャスト
	broadcaster.BroadcastStatus(map[string]string{"status": "updated"})

	// client2 のみが受信
	select {
	case <-client2.Events:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("mobile-client should receive event after web-client removal")
	}

	if broadcaster.ClientCount() != 1 {
		t.Errorf("expected 1 client remaining, got %d", broadcaster.ClientCount())
	}
}
