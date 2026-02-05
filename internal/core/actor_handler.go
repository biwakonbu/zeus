package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ActorHandler はアクターエンティティのハンドラー
type ActorHandler struct {
	fileStore FileStore
}

// NewActorHandler は ActorHandler を生成
func NewActorHandler(fs FileStore) *ActorHandler {
	return &ActorHandler{fileStore: fs}
}

// Type はエンティティタイプを返す
func (h *ActorHandler) Type() string {
	return "actor"
}

// Add はアクターを追加
func (h *ActorHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 既存のアクターファイルを読み込む
	var actorsFile ActorsFile
	if h.fileStore.Exists(ctx, "actors.yaml") {
		if err := h.fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
			return nil, fmt.Errorf("failed to read actors.yaml: %w", err)
		}
	} else {
		actorsFile = ActorsFile{Actors: []ActorEntity{}}
	}

	// ID を生成
	id := h.generateActorID()
	now := Now()

	actor := ActorEntity{
		ID:    id,
		Title: name,
		Type:  ActorTypeHuman,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(&actor)
	}

	// バリデーション
	if err := actor.Validate(); err != nil {
		return nil, err
	}

	// ファイルに追加
	actorsFile.Actors = append(actorsFile.Actors, actor)
	if err := h.fileStore.WriteYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		return nil, fmt.Errorf("failed to write actors.yaml: %w", err)
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はアクター一覧を取得
func (h *ActorHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var actorsFile ActorsFile
	if h.fileStore.Exists(ctx, "actors.yaml") {
		if err := h.fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
			return nil, fmt.Errorf("failed to read actors.yaml: %w", err)
		}
	} else {
		actorsFile = ActorsFile{Actors: []ActorEntity{}}
	}

	// Actor を ListItem に変換
	items := make([]ListItem, 0, len(actorsFile.Actors))
	for _, a := range actorsFile.Actors {
		items = append(items, ListItem{
			ID:        a.ID,
			Title:     a.Title,
			CreatedAt: a.Metadata.CreatedAt,
			UpdatedAt: a.Metadata.UpdatedAt,
		})
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  items,
		Total:  len(actorsFile.Actors),
	}, nil
}

// Get はアクターを取得
func (h *ActorHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID のセキュリティ検証
	if err := ValidateID("actor", id); err != nil {
		return nil, err
	}

	var actorsFile ActorsFile
	if !h.fileStore.Exists(ctx, "actors.yaml") {
		return nil, ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		return nil, fmt.Errorf("failed to read actors.yaml: %w", err)
	}

	for _, actor := range actorsFile.Actors {
		if actor.ID == id {
			return &actor, nil
		}
	}

	return nil, ErrEntityNotFound
}

// Update はアクターを更新
func (h *ActorHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("actor", id); err != nil {
		return err
	}

	var actorsFile ActorsFile
	if !h.fileStore.Exists(ctx, "actors.yaml") {
		return ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		return fmt.Errorf("failed to read actors.yaml: %w", err)
	}

	found := false
	for i := range actorsFile.Actors {
		if actorsFile.Actors[i].ID == id {
			// 更新データを適用
			if updateMap, ok := update.(map[string]any); ok {
				if title, exists := updateMap["title"].(string); exists {
					actorsFile.Actors[i].Title = title
				}
				if actorType, exists := updateMap["type"].(string); exists {
					actorsFile.Actors[i].Type = ActorType(actorType)
				}
				if desc, exists := updateMap["description"].(string); exists {
					actorsFile.Actors[i].Description = desc
				}
			}
			actorsFile.Actors[i].Metadata.UpdatedAt = Now()
			found = true
			break
		}
	}

	if !found {
		return ErrEntityNotFound
	}

	return h.fileStore.WriteYaml(ctx, "actors.yaml", &actorsFile)
}

// Delete はアクターを削除
func (h *ActorHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("actor", id); err != nil {
		return err
	}

	var actorsFile ActorsFile
	if !h.fileStore.Exists(ctx, "actors.yaml") {
		return ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		return fmt.Errorf("failed to read actors.yaml: %w", err)
	}

	// 削除対象を探す
	found := false
	newActors := make([]ActorEntity, 0, len(actorsFile.Actors))
	for _, actor := range actorsFile.Actors {
		if actor.ID == id {
			found = true
			continue
		}
		newActors = append(newActors, actor)
	}

	if !found {
		return ErrEntityNotFound
	}

	actorsFile.Actors = newActors
	return h.fileStore.WriteYaml(ctx, "actors.yaml", &actorsFile)
}

// generateActorID はアクター ID を生成
func (h *ActorHandler) generateActorID() string {
	return fmt.Sprintf("actor-%s", uuid.New().String()[:8])
}

// ===== EntityOption 関数群 =====

// WithActorType はアクタータイプを設定
func WithActorType(t ActorType) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActorEntity); ok {
			a.Type = t
		}
	}
}

// WithActorDescription はアクターの説明を設定
func WithActorDescription(desc string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActorEntity); ok {
			a.Description = desc
		}
	}
}

// WithActorOwner はアクターのオーナーを設定
func WithActorOwner(owner string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActorEntity); ok {
			a.Metadata.Owner = owner
		}
	}
}

// WithActorTags はアクターのタグを設定
func WithActorTags(tags []string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActorEntity); ok {
			a.Metadata.Tags = tags
		}
	}
}
