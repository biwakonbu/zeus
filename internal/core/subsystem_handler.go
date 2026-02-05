package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const subsystemsFileName = "subsystems.yaml"

// SubsystemHandler はサブシステムエンティティのハンドラー
type SubsystemHandler struct {
	fileStore FileStore
}

// NewSubsystemHandler は SubsystemHandler を生成
func NewSubsystemHandler(fs FileStore) *SubsystemHandler {
	return &SubsystemHandler{fileStore: fs}
}

// Type はエンティティタイプを返す
func (h *SubsystemHandler) Type() string {
	return "subsystem"
}

// Add はサブシステムを追加
func (h *SubsystemHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 既存のサブシステムファイルを読み込む
	var subsystemsFile SubsystemsFile
	if h.fileStore.Exists(ctx, subsystemsFileName) {
		if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
		}
	} else {
		subsystemsFile = SubsystemsFile{Subsystems: []SubsystemEntity{}}
	}

	// ID を生成
	id := h.generateSubsystemID()
	now := Now()

	subsystem := SubsystemEntity{
		ID:   id,
		Name: name,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(&subsystem)
	}

	// バリデーション
	if err := subsystem.Validate(); err != nil {
		return nil, err
	}

	// ファイルに追加
	subsystemsFile.Subsystems = append(subsystemsFile.Subsystems, subsystem)
	if err := h.fileStore.WriteYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", subsystemsFileName, err)
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はサブシステム一覧を取得
func (h *SubsystemHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var subsystemsFile SubsystemsFile
	if h.fileStore.Exists(ctx, subsystemsFileName) {
		if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
		}
	} else {
		subsystemsFile = SubsystemsFile{Subsystems: []SubsystemEntity{}}
	}

	// Subsystem を ListItem に変換
	items := make([]ListItem, 0, len(subsystemsFile.Subsystems))
	for _, s := range subsystemsFile.Subsystems {
		items = append(items, ListItem{
			ID:          s.ID,
			Title:       s.Name,
			Description: s.Description,
			CreatedAt:   s.Metadata.CreatedAt,
			UpdatedAt:   s.Metadata.UpdatedAt,
		})
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  items,
		Total:  len(subsystemsFile.Subsystems),
	}, nil
}

// Get はサブシステムを取得
func (h *SubsystemHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID のセキュリティ検証
	if err := ValidateID("subsystem", id); err != nil {
		return nil, err
	}

	var subsystemsFile SubsystemsFile
	if !h.fileStore.Exists(ctx, subsystemsFileName) {
		return nil, ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
	}

	for _, subsystem := range subsystemsFile.Subsystems {
		if subsystem.ID == id {
			return &subsystem, nil
		}
	}

	return nil, ErrEntityNotFound
}

// Update はサブシステムを更新
func (h *SubsystemHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("subsystem", id); err != nil {
		return err
	}

	var subsystemsFile SubsystemsFile
	if !h.fileStore.Exists(ctx, subsystemsFileName) {
		return ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
		return fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
	}

	found := false
	for i := range subsystemsFile.Subsystems {
		if subsystemsFile.Subsystems[i].ID == id {
			// 更新データを適用
			if updateMap, ok := update.(map[string]any); ok {
				if name, exists := updateMap["name"].(string); exists {
					subsystemsFile.Subsystems[i].Name = name
				}
				if desc, exists := updateMap["description"].(string); exists {
					subsystemsFile.Subsystems[i].Description = desc
				}
			}
			subsystemsFile.Subsystems[i].Metadata.UpdatedAt = Now()
			found = true
			break
		}
	}

	if !found {
		return ErrEntityNotFound
	}

	return h.fileStore.WriteYaml(ctx, subsystemsFileName, &subsystemsFile)
}

// Delete はサブシステムを削除
func (h *SubsystemHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("subsystem", id); err != nil {
		return err
	}

	var subsystemsFile SubsystemsFile
	if !h.fileStore.Exists(ctx, subsystemsFileName) {
		return ErrEntityNotFound
	}
	if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
		return fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
	}

	// 削除対象を探す
	found := false
	newSubsystems := make([]SubsystemEntity, 0, len(subsystemsFile.Subsystems))
	for _, subsystem := range subsystemsFile.Subsystems {
		if subsystem.ID == id {
			found = true
			continue
		}
		newSubsystems = append(newSubsystems, subsystem)
	}

	if !found {
		return ErrEntityNotFound
	}

	subsystemsFile.Subsystems = newSubsystems
	return h.fileStore.WriteYaml(ctx, subsystemsFileName, &subsystemsFile)
}

// generateSubsystemID はサブシステム ID を生成
func (h *SubsystemHandler) generateSubsystemID() string {
	return fmt.Sprintf("sub-%s", uuid.New().String()[:8])
}

// ListAll は全サブシステムを取得（内部用）
func (h *SubsystemHandler) ListAll(ctx context.Context) ([]SubsystemEntity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var subsystemsFile SubsystemsFile
	if h.fileStore.Exists(ctx, subsystemsFileName) {
		if err := h.fileStore.ReadYaml(ctx, subsystemsFileName, &subsystemsFile); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", subsystemsFileName, err)
		}
	} else {
		subsystemsFile = SubsystemsFile{Subsystems: []SubsystemEntity{}}
	}

	return subsystemsFile.Subsystems, nil
}

// ===== EntityOption 関数群 =====

// WithSubsystemDescription はサブシステムの説明を設定
func WithSubsystemDescription(desc string) EntityOption {
	return func(v any) {
		if s, ok := v.(*SubsystemEntity); ok {
			s.Description = desc
		}
	}
}

// WithSubsystemOwner はサブシステムのオーナーを設定
func WithSubsystemOwner(owner string) EntityOption {
	return func(v any) {
		if s, ok := v.(*SubsystemEntity); ok {
			s.Metadata.Owner = owner
		}
	}
}

// WithSubsystemTags はサブシステムのタグを設定
func WithSubsystemTags(tags []string) EntityOption {
	return func(v any) {
		if s, ok := v.(*SubsystemEntity); ok {
			s.Metadata.Tags = tags
		}
	}
}
