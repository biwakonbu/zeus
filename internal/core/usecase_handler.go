package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// UseCaseHandler はユースケースエンティティのハンドラー
type UseCaseHandler struct {
	fileStore        FileStore
	objectiveHandler *ObjectiveHandler
	actorHandler     *ActorHandler
	idCounterManager *IDCounterManager
}

// NewUseCaseHandler は UseCaseHandler を生成
func NewUseCaseHandler(fs FileStore, objHandler *ObjectiveHandler, actorHandler *ActorHandler, idCounter *IDCounterManager) *UseCaseHandler {
	return &UseCaseHandler{
		fileStore:        fs,
		objectiveHandler: objHandler,
		actorHandler:     actorHandler,
		idCounterManager: idCounter,
	}
}

// Type はエンティティタイプを返す
func (h *UseCaseHandler) Type() string {
	return "usecase"
}

// Add はユースケースを追加
func (h *UseCaseHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// usecases ディレクトリを確保
	if err := h.fileStore.EnsureDir(ctx, "usecases"); err != nil {
		return nil, fmt.Errorf("failed to ensure usecases directory: %w", err)
	}

	// ID を生成
	id := h.generateUseCaseID()
	now := Now()

	usecase := UseCaseEntity{
		ID:     id,
		Title:  name,
		Status: UseCaseStatusDraft,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(&usecase)
	}

	// 参照整合性チェック: ObjectiveID
	if usecase.ObjectiveID != "" && h.objectiveHandler != nil {
		if _, err := h.objectiveHandler.Get(ctx, usecase.ObjectiveID); err != nil {
			return nil, fmt.Errorf("referenced objective not found: %s", usecase.ObjectiveID)
		}
	}

	// 参照整合性チェック: Actor 参照
	if len(usecase.Actors) > 0 && h.actorHandler != nil {
		for _, actorRef := range usecase.Actors {
			if _, err := h.actorHandler.Get(ctx, actorRef.ActorID); err != nil {
				return nil, fmt.Errorf("referenced actor not found: %s", actorRef.ActorID)
			}
		}
	}

	// バリデーション
	if err := usecase.Validate(); err != nil {
		return nil, err
	}

	// 個別ファイルに保存
	filePath := filepath.Join("usecases", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, &usecase); err != nil {
		return nil, fmt.Errorf("failed to write usecase file: %w", err)
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はユースケース一覧を取得
func (h *UseCaseHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// usecases ディレクトリが存在しない場合は空リストを返す
	if !h.fileStore.Exists(ctx, "usecases") {
		return &ListResult{
			Entity: h.Type(),
			Items:  []Task{},
			Total:  0,
		}, nil
	}

	// ディレクトリ内のファイルを列挙
	files, err := h.fileStore.ListDir(ctx, "usecases")
	if err != nil {
		return nil, fmt.Errorf("failed to list usecases directory: %w", err)
	}

	items := make([]Task, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var usecase UseCaseEntity
		if err := h.fileStore.ReadYaml(ctx, filepath.Join("usecases", file), &usecase); err != nil {
			continue // 読み込み失敗はスキップ
		}
		// Task に変換（ListResult 互換性のため）
		items = append(items, Task{
			ID:        usecase.ID,
			Title:     usecase.Title,
			Status:    TaskStatus(usecase.Status),
			CreatedAt: usecase.Metadata.CreatedAt,
			UpdatedAt: usecase.Metadata.UpdatedAt,
		})
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  items,
		Total:  len(items),
	}, nil
}

// Get はユースケースを取得
func (h *UseCaseHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID のセキュリティ検証
	if err := ValidateID("usecase", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("usecases", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return nil, ErrEntityNotFound
	}

	var usecase UseCaseEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &usecase); err != nil {
		return nil, fmt.Errorf("failed to read usecase file: %w", err)
	}

	return &usecase, nil
}

// Update はユースケースを更新
func (h *UseCaseHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("usecase", id); err != nil {
		return err
	}

	filePath := filepath.Join("usecases", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var usecase UseCaseEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &usecase); err != nil {
		return fmt.Errorf("failed to read usecase file: %w", err)
	}

	// 更新データを適用
	if updateMap, ok := update.(map[string]any); ok {
		if title, exists := updateMap["title"].(string); exists {
			usecase.Title = title
		}
		if desc, exists := updateMap["description"].(string); exists {
			usecase.Description = desc
		}
		if status, exists := updateMap["status"].(string); exists {
			usecase.Status = UseCaseStatus(status)
		}
		if objectiveID, exists := updateMap["objective_id"].(string); exists {
			usecase.ObjectiveID = objectiveID
		}
	}
	usecase.Metadata.UpdatedAt = Now()

	// バリデーション
	if err := usecase.Validate(); err != nil {
		return err
	}

	return h.fileStore.WriteYaml(ctx, filePath, &usecase)
}

// Delete はユースケースを削除
func (h *UseCaseHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("usecase", id); err != nil {
		return err
	}

	filePath := filepath.Join("usecases", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	return h.fileStore.Delete(ctx, filePath)
}

// AddRelation はユースケース間の関係を追加
func (h *UseCaseHandler) AddRelation(ctx context.Context, usecaseID string, rel UseCaseRelation) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("usecase", usecaseID); err != nil {
		return err
	}

	filePath := filepath.Join("usecases", usecaseID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var usecase UseCaseEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &usecase); err != nil {
		return fmt.Errorf("failed to read usecase file: %w", err)
	}

	// ターゲット UseCase の存在確認
	if _, err := h.Get(ctx, rel.TargetID); err != nil {
		return fmt.Errorf("target usecase not found: %s", rel.TargetID)
	}

	// 関係を追加
	usecase.Relations = append(usecase.Relations, rel)
	usecase.Metadata.UpdatedAt = Now()

	// バリデーション
	if err := usecase.Validate(); err != nil {
		return err
	}

	return h.fileStore.WriteYaml(ctx, filePath, &usecase)
}

// AddActor はユースケースにアクター参照を追加
func (h *UseCaseHandler) AddActor(ctx context.Context, usecaseID string, actorRef UseCaseActorRef) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("usecase", usecaseID); err != nil {
		return err
	}

	filePath := filepath.Join("usecases", usecaseID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var usecase UseCaseEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &usecase); err != nil {
		return fmt.Errorf("failed to read usecase file: %w", err)
	}

	// Actor の存在確認
	if h.actorHandler != nil {
		if _, err := h.actorHandler.Get(ctx, actorRef.ActorID); err != nil {
			return fmt.Errorf("actor not found: %s", actorRef.ActorID)
		}
	}

	// 重複チェック
	for _, existing := range usecase.Actors {
		if existing.ActorID == actorRef.ActorID {
			return fmt.Errorf("actor already associated: %s", actorRef.ActorID)
		}
	}

	// アクター参照を追加
	usecase.Actors = append(usecase.Actors, actorRef)
	usecase.Metadata.UpdatedAt = Now()

	return h.fileStore.WriteYaml(ctx, filePath, &usecase)
}

// generateUseCaseID はユースケース ID を生成
func (h *UseCaseHandler) generateUseCaseID() string {
	return fmt.Sprintf("uc-%s", uuid.New().String()[:8])
}

// ===== EntityOption 関数群 =====

// WithUseCaseObjective は Objective ID を設定
func WithUseCaseObjective(objID string) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.ObjectiveID = objID
		}
	}
}

// WithUseCaseDescription は説明を設定
func WithUseCaseDescription(desc string) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Description = desc
		}
	}
}

// WithUseCaseActor はアクター参照を追加
func WithUseCaseActor(actorID string, role ActorRole) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Actors = append(u.Actors, UseCaseActorRef{
				ActorID: actorID,
				Role:    role,
			})
		}
	}
}

// WithUseCaseStatus はステータスを設定
func WithUseCaseStatus(status UseCaseStatus) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Status = status
		}
	}
}

// WithUseCaseOwner はオーナーを設定
func WithUseCaseOwner(owner string) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Metadata.Owner = owner
		}
	}
}

// WithUseCaseTags はタグを設定
func WithUseCaseTags(tags []string) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Metadata.Tags = tags
		}
	}
}

// WithUseCaseScenario はシナリオを設定
func WithUseCaseScenario(mainFlow []string) EntityOption {
	return func(v any) {
		if u, ok := v.(*UseCaseEntity); ok {
			u.Scenario = UseCaseScenario{MainFlow: mainFlow}
		}
	}
}
