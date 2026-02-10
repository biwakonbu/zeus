package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ConsiderationHandler は ConsiderationEntity エンティティのハンドラー
// 個別ファイル (considerations/con-{uuid}.yaml) で管理
type ConsiderationHandler struct {
	fileStore        FileStore
	sanitizer        *Sanitizer
	objectiveHandler *ObjectiveHandler
}

// NewConsiderationHandler は新しい ConsiderationHandler を作成
func NewConsiderationHandler(fs FileStore, objHandler *ObjectiveHandler, _ *IDCounterManager) *ConsiderationHandler {
	return &ConsiderationHandler{
		fileStore:        fs,
		sanitizer:        NewSanitizer(),
		objectiveHandler: objHandler,
	}
}

// Type はエンティティタイプを返す
func (h *ConsiderationHandler) Type() string {
	return "consideration"
}

// Add は Consideration を追加
func (h *ConsiderationHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// サニタイズ
	sanitizedName, err := h.sanitizer.SanitizeString("title", name)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	// UUID 形式の ID を生成
	id := h.generateID()

	now := Now()
	consideration := &ConsiderationEntity{
		ID:     id,
		Title:  sanitizedName,
		Status: ConsiderationStatusOpen,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(consideration)
	}

	// 参照の存在確認
	if consideration.ObjectiveID != "" {
		if err := h.validateObjectiveReference(ctx, consideration.ObjectiveID); err != nil {
			return nil, err
		}
	}
	// バリデーション
	if err := consideration.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("considerations", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, consideration); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Consideration 一覧を取得
func (h *ConsiderationHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	considerations, err := h.getAllConsiderations(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*ConsiderationEntity{}
		for _, con := range considerations {
			if string(con.Status) == filter.Status {
				filtered = append(filtered, con)
			}
		}
		considerations = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(considerations) > filter.Limit {
		considerations = considerations[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []ListItem{},
		Total:  len(considerations),
	}, nil
}

// Get は Consideration を取得
func (h *ConsiderationHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("consideration", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("considerations", id+".yaml")
	var consideration ConsiderationEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &consideration); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &consideration, nil
}

// Update は Consideration を更新
func (h *ConsiderationHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("consideration", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingCon := existing.(*ConsiderationEntity)

	// 更新適用
	if con, ok := update.(*ConsiderationEntity); ok {
		con.ID = id // ID は変更不可
		con.Metadata.CreatedAt = existingCon.Metadata.CreatedAt
		con.Metadata.UpdatedAt = Now()

		// 参照の存在確認
		if con.ObjectiveID != "" && con.ObjectiveID != existingCon.ObjectiveID {
			if err := h.validateObjectiveReference(ctx, con.ObjectiveID); err != nil {
				return err
			}
		}
		// バリデーション
		if err := con.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("considerations", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, con)
	}

	return fmt.Errorf("invalid update type: expected *ConsiderationEntity")
}

// Delete は Consideration を削除
// Decision が参照している Consideration は削除不可（逆参照整合性）
func (h *ConsiderationHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("consideration", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// 逆参照チェック: Decision がこの Consideration を参照しているか確認
	if referencedBy, err := h.checkDecisionReferences(ctx, id); err != nil {
		return fmt.Errorf("failed to check reverse references: %w", err)
	} else if referencedBy != "" {
		return fmt.Errorf("cannot delete consideration %s: referenced by decision %s (decision is immutable)", id, referencedBy)
	}

	// ファイル削除
	filePath := filepath.Join("considerations", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// checkDecisionReferences は指定された Consideration を参照する Decision を検索
// 見つかった場合は Decision ID を返し、見つからない場合は空文字列を返す
func (h *ConsiderationHandler) checkDecisionReferences(ctx context.Context, considerationID string) (string, error) {
	files, err := h.fileStore.ListDir(ctx, "decisions")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // decisions フォルダがなければ参照なし
		}
		return "", err
	}

	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		filePath := filepath.Join("decisions", file)
		var dec DecisionEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &dec); err != nil {
			continue // 読み込みエラーはスキップ
		}

		if dec.ConsiderationID == considerationID {
			return dec.ID, nil // この Decision が参照している
		}
	}

	return "", nil // 参照なし
}

// generateID は UUID 形式の Consideration ID を生成
func (h *ConsiderationHandler) generateID() string {
	return fmt.Sprintf("con-%s", uuid.New().String()[:8])
}

// getAllConsiderations は全 Consideration を取得
func (h *ConsiderationHandler) getAllConsiderations(ctx context.Context) ([]*ConsiderationEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "considerations")
	if err != nil {
		if os.IsNotExist(err) {
			return []*ConsiderationEntity{}, nil
		}
		return nil, err
	}

	var considerations []*ConsiderationEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("consideration", id); err != nil {
			continue
		}

		filePath := filepath.Join("considerations", file)
		var con ConsiderationEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &con); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read consideration file %s: %w", filePath, err)
			}
			continue
		}
		considerations = append(considerations, &con)
	}

	sort.Slice(considerations, func(i, j int) bool {
		return considerations[i].ID < considerations[j].ID
	})

	return considerations, nil
}

// validateObjectiveReference は Objective 参照の存在を確認
func (h *ConsiderationHandler) validateObjectiveReference(ctx context.Context, objectiveID string) error {
	if h.objectiveHandler == nil {
		return nil
	}

	_, err := h.objectiveHandler.Get(ctx, objectiveID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced objective not found: %s", objectiveID)
	}
	return err
}

// SetDecisionID は Decision との紐付けを設定し、ステータスを decided に更新
func (h *ConsiderationHandler) SetDecisionID(ctx context.Context, id, decisionID string) error {
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	con := existing.(*ConsiderationEntity)
	con.DecisionID = decisionID
	con.Status = ConsiderationStatusDecided
	con.Metadata.UpdatedAt = Now()

	filePath := filepath.Join("considerations", id+".yaml")
	return h.fileStore.WriteYaml(ctx, filePath, con)
}

// Consideration オプション関数

// WithConsiderationContext は Consideration のコンテキストを設定
func WithConsiderationContext(context string) EntityOption {
	return func(v any) {
		if con, ok := v.(*ConsiderationEntity); ok {
			con.Context = context
		}
	}
}

// WithConsiderationObjective は Consideration の Objective を設定
func WithConsiderationObjective(objectiveID string) EntityOption {
	return func(v any) {
		if con, ok := v.(*ConsiderationEntity); ok {
			con.ObjectiveID = objectiveID
		}
	}
}

// WithConsiderationOptions は Consideration のオプションを設定
func WithConsiderationOptions(options []ConsiderationOption) EntityOption {
	return func(v any) {
		if con, ok := v.(*ConsiderationEntity); ok {
			con.Options = options
		}
	}
}

// WithConsiderationRaisedBy は Consideration の提起者を設定
func WithConsiderationRaisedBy(raisedBy string) EntityOption {
	return func(v any) {
		if con, ok := v.(*ConsiderationEntity); ok {
			con.RaisedBy = raisedBy
		}
	}
}

// WithConsiderationDueDate は Consideration の期限を設定
func WithConsiderationDueDate(dueDate string) EntityOption {
	return func(v any) {
		if con, ok := v.(*ConsiderationEntity); ok {
			con.DueDate = dueDate
		}
	}
}
