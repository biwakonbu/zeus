package core

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// IDCounters は各エンティティタイプの ID カウンターを管理
// O(1) で次の ID 番号を取得するために使用
type IDCounters struct {
	Counters map[string]int `yaml:"counters"`
}

// IDCounterManager は ID カウンターを管理
type IDCounterManager struct {
	fileStore FileStore
	mu        sync.Mutex
	cache     *IDCounters
}

// NewIDCounterManager は新しい IDCounterManager を作成
func NewIDCounterManager(fs FileStore) *IDCounterManager {
	return &IDCounterManager{
		fileStore: fs,
	}
}

// GetNextID は指定されたエンティティタイプの次の ID 番号を取得し、カウンターをインクリメント
// O(1) で動作（キャッシュがある場合）
func (m *IDCounterManager) GetNextID(ctx context.Context, entityType string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// キャッシュをロード
	if m.cache == nil {
		if err := m.loadCache(ctx); err != nil {
			return 0, err
		}
	}

	// 現在の値を取得してインクリメント
	current := m.cache.Counters[entityType]
	next := current + 1
	m.cache.Counters[entityType] = next

	// ファイルに保存
	if err := m.saveCache(ctx); err != nil {
		// 失敗した場合はカウンターを戻す
		m.cache.Counters[entityType] = current
		return 0, err
	}

	return next, nil
}

// InitializeFromExisting は既存のエンティティから最大 ID を取得してカウンターを初期化
// 初回起動時や移行時に使用
func (m *IDCounterManager) InitializeFromExisting(ctx context.Context, entityType string, maxID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache == nil {
		if err := m.loadCache(ctx); err != nil {
			return err
		}
	}

	// 現在のカウンターよりも大きい場合のみ更新
	if maxID > m.cache.Counters[entityType] {
		m.cache.Counters[entityType] = maxID
		return m.saveCache(ctx)
	}

	return nil
}

// loadCache はファイルからカウンターをロード
func (m *IDCounterManager) loadCache(ctx context.Context) error {
	m.cache = &IDCounters{
		Counters: make(map[string]int),
	}

	err := m.fileStore.ReadYaml(ctx, "id_counters.yaml", m.cache)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルがない場合は空のカウンターで初期化
			return nil
		}
		return fmt.Errorf("failed to load id counters: %w", err)
	}

	// nil マップ対策
	if m.cache.Counters == nil {
		m.cache.Counters = make(map[string]int)
	}

	return nil
}

// saveCache はカウンターをファイルに保存
func (m *IDCounterManager) saveCache(ctx context.Context) error {
	return m.fileStore.WriteYaml(ctx, "id_counters.yaml", m.cache)
}

// InvalidateCache はキャッシュを無効化
func (m *IDCounterManager) InvalidateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = nil
}

// GetCurrentID は指定されたエンティティタイプの現在の ID 番号を取得（インクリメントなし）
func (m *IDCounterManager) GetCurrentID(ctx context.Context, entityType string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache == nil {
		if err := m.loadCache(ctx); err != nil {
			return 0, err
		}
	}

	return m.cache.Counters[entityType], nil
}

// InitializeAllFromScanner は既存エンティティをスキャンしてカウンターを初期化
// scanner は (ctx, entityType) -> maxID を返す関数
func (m *IDCounterManager) InitializeAllFromScanner(ctx context.Context, scanner func(ctx context.Context, entityType string) (int, error), entityTypes []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache == nil {
		if err := m.loadCache(ctx); err != nil {
			return err
		}
	}

	changed := false
	for _, entityType := range entityTypes {
		maxID, err := scanner(ctx, entityType)
		if err != nil {
			continue // スキャン失敗は無視
		}
		if maxID > m.cache.Counters[entityType] {
			m.cache.Counters[entityType] = maxID
			changed = true
		}
	}

	if changed {
		return m.saveCache(ctx)
	}
	return nil
}
