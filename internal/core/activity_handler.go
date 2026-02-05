package core

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
)

// ActivityHandler はアクティビティエンティティのハンドラー
type ActivityHandler struct {
	fileStore          FileStore
	usecaseHandler     *UseCaseHandler
	deliverableHandler *DeliverableHandler
}

// NewActivityHandler は ActivityHandler を生成
func NewActivityHandler(fs FileStore, usecaseHandler *UseCaseHandler, deliverableHandler *DeliverableHandler, _ *IDCounterManager) *ActivityHandler {
	return &ActivityHandler{
		fileStore:          fs,
		usecaseHandler:     usecaseHandler,
		deliverableHandler: deliverableHandler,
	}
}

// Type はエンティティタイプを返す
func (h *ActivityHandler) Type() string {
	return "activity"
}

// Add はアクティビティを追加
func (h *ActivityHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリを確保
	if err := h.fileStore.EnsureDir(ctx, "activities"); err != nil {
		return nil, fmt.Errorf("failed to ensure activities directory: %w", err)
	}

	// ID を生成
	id, err := h.generateActivityID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate activity ID: %w", err)
	}
	now := Now()

	activity := ActivityEntity{
		ID:     id,
		Title:  name,
		Status: ActivityStatusDraft,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(&activity)
	}

	// 参照整合性チェック: UseCaseID（任意紐付け）
	if activity.UseCaseID != "" && h.usecaseHandler != nil {
		if _, err := h.usecaseHandler.Get(ctx, activity.UseCaseID); err != nil {
			return nil, fmt.Errorf("referenced usecase not found: %s", activity.UseCaseID)
		}
	}

	// 参照整合性チェック: RelatedDeliverables（推奨）
	if len(activity.RelatedDeliverables) > 0 && h.deliverableHandler != nil {
		for _, delID := range activity.RelatedDeliverables {
			if _, err := h.deliverableHandler.Get(ctx, delID); err != nil {
				return nil, fmt.Errorf("referenced deliverable not found in related_deliverables: %s", delID)
			}
		}
	}

	// 参照整合性チェック: Nodes 内の DeliverableIDs（任意）
	if h.deliverableHandler != nil {
		for _, node := range activity.Nodes {
			for _, delID := range node.DeliverableIDs {
				if _, err := h.deliverableHandler.Get(ctx, delID); err != nil {
					return nil, fmt.Errorf("referenced deliverable not found in node %s deliverable_ids: %s", node.ID, delID)
				}
			}
		}
	}

	// 参照整合性チェック: ParentID（任意）
	if activity.ParentID != "" {
		if _, err := h.Get(ctx, activity.ParentID); err != nil {
			return nil, fmt.Errorf("referenced parent activity not found: %s", activity.ParentID)
		}
	}

	// 参照整合性チェック: Dependencies（任意）
	for _, depID := range activity.Dependencies {
		if _, err := h.Get(ctx, depID); err != nil {
			return nil, fmt.Errorf("referenced dependency activity not found: %s", depID)
		}
	}

	// バリデーション
	if err := activity.Validate(); err != nil {
		return nil, err
	}

	// 個別ファイルに保存
	filePath := filepath.Join("activities", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, &activity); err != nil {
		return nil, fmt.Errorf("failed to write activity file: %w", err)
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はアクティビティ一覧を取得
func (h *ActivityHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリが存在しない場合は空リストを返す
	if !h.fileStore.Exists(ctx, "activities") {
		return &ListResult{
			Entity: h.Type(),
			Items:  []ListItem{},
			Total:  0,
		}, nil
	}

	// ディレクトリ内のファイルを列挙
	files, err := h.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return nil, fmt.Errorf("failed to list activities directory: %w", err)
	}

	items := make([]ListItem, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var activity ActivityEntity
		if err := h.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &activity); err != nil {
			continue // 読み込み失敗はスキップ
		}
		// Task に変換（ListResult 互換性のため）
		items = append(items, activity.ToListItem())
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  items,
		Total:  len(items),
	}, nil
}

// Get はアクティビティを取得
func (h *ActivityHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return nil, ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return nil, fmt.Errorf("failed to read activity file: %w", err)
	}

	return &activity, nil
}

// Update はアクティビティを更新
func (h *ActivityHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// 更新データを適用
	if updateMap, ok := update.(map[string]any); ok {
		if title, exists := updateMap["title"].(string); exists {
			activity.Title = title
		}
		if desc, exists := updateMap["description"].(string); exists {
			activity.Description = desc
		}
		if status, exists := updateMap["status"].(string); exists {
			activity.Status = ActivityStatus(status)
		}
		if usecaseID, exists := updateMap["usecase_id"].(string); exists {
			activity.UseCaseID = usecaseID
		}
		if val, exists := updateMap["related_deliverables"]; exists {
			if relatedDeliverables, ok := val.([]string); ok {
				activity.RelatedDeliverables = relatedDeliverables
			} else if val != nil {
				return fmt.Errorf("related_deliverables must be []string, got %T", val)
			}
		}
		// Task/Activity 統合: 新フィールドの更新
		if val, exists := updateMap["dependencies"]; exists {
			if dependencies, ok := val.([]string); ok {
				activity.Dependencies = dependencies
			} else if val != nil {
				return fmt.Errorf("dependencies must be []string, got %T", val)
			}
		}
		if parentID, exists := updateMap["parent_id"].(string); exists {
			activity.ParentID = parentID
		}
		if estimateHours, exists := updateMap["estimate_hours"].(float64); exists {
			activity.EstimateHours = estimateHours
		}
		if actualHours, exists := updateMap["actual_hours"].(float64); exists {
			activity.ActualHours = actualHours
		}
		if assignee, exists := updateMap["assignee"].(string); exists {
			activity.Assignee = assignee
		}
		if startDate, exists := updateMap["start_date"].(string); exists {
			activity.StartDate = startDate
		}
		if dueDate, exists := updateMap["due_date"].(string); exists {
			activity.DueDate = dueDate
		}
		if priority, exists := updateMap["priority"].(string); exists {
			activity.Priority = ActivityPriority(priority)
		}
		if wbsCode, exists := updateMap["wbs_code"].(string); exists {
			activity.WBSCode = wbsCode
		}
		if progress, exists := updateMap["progress"].(int); exists {
			activity.Progress = progress
		}
		if approvalLevel, exists := updateMap["approval_level"].(string); exists {
			activity.ApprovalLevel = ApprovalLevel(approvalLevel)
		}
	}

	// 参照整合性チェック: UseCaseID（任意紐付け）
	if activity.UseCaseID != "" && h.usecaseHandler != nil {
		if _, err := h.usecaseHandler.Get(ctx, activity.UseCaseID); err != nil {
			return fmt.Errorf("referenced usecase not found: %s", activity.UseCaseID)
		}
	}

	// 参照整合性チェック: RelatedDeliverables（推奨）
	if len(activity.RelatedDeliverables) > 0 && h.deliverableHandler != nil {
		for _, delID := range activity.RelatedDeliverables {
			if _, err := h.deliverableHandler.Get(ctx, delID); err != nil {
				return fmt.Errorf("referenced deliverable not found in related_deliverables: %s", delID)
			}
		}
	}

	// 参照整合性チェック: ParentID（任意）
	if activity.ParentID != "" {
		if _, err := h.Get(ctx, activity.ParentID); err != nil {
			return fmt.Errorf("referenced parent activity not found: %s", activity.ParentID)
		}
	}

	// 参照整合性チェック: Dependencies（任意）
	for _, depID := range activity.Dependencies {
		if _, err := h.Get(ctx, depID); err != nil {
			return fmt.Errorf("referenced dependency activity not found: %s", depID)
		}
	}

	activity.Metadata.UpdatedAt = Now()

	// バリデーション
	if err := activity.Validate(); err != nil {
		return err
	}

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// Delete はアクティビティを削除
func (h *ActivityHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	return h.fileStore.Delete(ctx, filePath)
}

// GetAll は全アクティビティを取得（API用）
func (h *ActivityHandler) GetAll(ctx context.Context) ([]ActivityEntity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリが存在しない場合は空リストを返す
	if !h.fileStore.Exists(ctx, "activities") {
		return []ActivityEntity{}, nil
	}

	// ディレクトリ内のファイルを列挙
	files, err := h.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return nil, fmt.Errorf("failed to list activities directory: %w", err)
	}

	activities := make([]ActivityEntity, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var activity ActivityEntity
		if err := h.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &activity); err != nil {
			continue // 読み込み失敗はスキップ
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// GetAllSimple は SimpleActivity モードのアクティビティのみを取得
func (h *ActivityHandler) GetAllSimple(ctx context.Context) ([]ActivityEntity, error) {
	activities, err := h.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ActivityEntity, 0)
	for _, a := range activities {
		if a.IsSimple() {
			result = append(result, a)
		}
	}
	return result, nil
}

// GetAllFlow は FlowActivity モードのアクティビティのみを取得
func (h *ActivityHandler) GetAllFlow(ctx context.Context) ([]ActivityEntity, error) {
	activities, err := h.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ActivityEntity, 0)
	for _, a := range activities {
		if a.IsFlow() {
			result = append(result, a)
		}
	}
	return result, nil
}

// DetectDependencyCycles は依存関係の循環を検出
// 循環が検出された場合は循環するIDのリストを返す
func (h *ActivityHandler) DetectDependencyCycles(ctx context.Context) ([][]string, error) {
	activities, err := h.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// ID -> Activity のマップを作成
	activityMap := make(map[string]*ActivityEntity)
	for i := range activities {
		activityMap[activities[i].ID] = &activities[i]
	}

	// 訪問状態: 0=未訪問, 1=訪問中, 2=訪問済み
	visited := make(map[string]int)
	var cycles [][]string

	// 現在のパスを追跡
	var path []string

	var dfs func(id string) bool
	dfs = func(id string) bool {
		if visited[id] == 2 {
			return false // 既に完了
		}
		if visited[id] == 1 {
			// 循環検出
			cycleStart := -1
			for i, pid := range path {
				if pid == id {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				cycle = append(cycle, id)
				cycles = append(cycles, cycle)
			}
			return true
		}

		visited[id] = 1
		path = append(path, id)

		activity := activityMap[id]
		if activity != nil {
			for _, depID := range activity.Dependencies {
				dfs(depID)
			}
			// ParentID も依存関係として扱う
			if activity.ParentID != "" {
				dfs(activity.ParentID)
			}
		}

		path = path[:len(path)-1]
		visited[id] = 2
		return false
	}

	// 全アクティビティから DFS を開始
	for id := range activityMap {
		if visited[id] == 0 {
			dfs(id)
		}
	}

	return cycles, nil
}

// GetDependents は指定アクティビティに依存するアクティビティを取得
func (h *ActivityHandler) GetDependents(ctx context.Context, id string) ([]ActivityEntity, error) {
	activities, err := h.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ActivityEntity, 0)
	for _, a := range activities {
		if slices.Contains(a.Dependencies, id) {
			result = append(result, a)
		}
	}
	return result, nil
}

// GetChildren は指定アクティビティの子アクティビティを取得
func (h *ActivityHandler) GetChildren(ctx context.Context, parentID string) ([]ActivityEntity, error) {
	activities, err := h.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ActivityEntity, 0)
	for _, a := range activities {
		if a.ParentID == parentID {
			result = append(result, a)
		}
	}
	return result, nil
}

// AddNode はアクティビティにノードを追加
func (h *ActivityHandler) AddNode(ctx context.Context, activityID string, node ActivityNode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", activityID); err != nil {
		return err
	}

	filePath := filepath.Join("activities", activityID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// ノードのバリデーション
	if err := node.Validate(); err != nil {
		return err
	}

	// 参照整合性チェック: DeliverableIDs（任意）
	if len(node.DeliverableIDs) > 0 && h.deliverableHandler != nil {
		for _, delID := range node.DeliverableIDs {
			if _, err := h.deliverableHandler.Get(ctx, delID); err != nil {
				return fmt.Errorf("referenced deliverable not found in node %s deliverable_ids: %s", node.ID, delID)
			}
		}
	}

	// 重複チェック
	for _, existing := range activity.Nodes {
		if existing.ID == node.ID {
			return fmt.Errorf("node already exists: %s", node.ID)
		}
	}

	// ノードを追加
	activity.Nodes = append(activity.Nodes, node)
	activity.Metadata.UpdatedAt = Now()

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// AddTransition はアクティビティに遷移を追加
func (h *ActivityHandler) AddTransition(ctx context.Context, activityID string, trans ActivityTransition) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", activityID); err != nil {
		return err
	}

	filePath := filepath.Join("activities", activityID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// 遷移のバリデーション
	if err := trans.Validate(); err != nil {
		return err
	}

	// ソース/ターゲットのノード存在確認
	nodeIDs := make(map[string]bool)
	for _, node := range activity.Nodes {
		nodeIDs[node.ID] = true
	}
	if !nodeIDs[trans.Source] {
		return fmt.Errorf("transition source not found: %s", trans.Source)
	}
	if !nodeIDs[trans.Target] {
		return fmt.Errorf("transition target not found: %s", trans.Target)
	}

	// 重複チェック
	for _, existing := range activity.Transitions {
		if existing.ID == trans.ID {
			return fmt.Errorf("transition already exists: %s", trans.ID)
		}
	}

	// 遷移を追加
	activity.Transitions = append(activity.Transitions, trans)
	activity.Metadata.UpdatedAt = Now()

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// generateActivityID はアクティビティ ID を生成（UUID 形式）
func (h *ActivityHandler) generateActivityID(_ context.Context) (string, error) {
	return fmt.Sprintf("act-%s", uuid.New().String()[:8]), nil
}

// ===== EntityOption 関数群 =====

// WithActivityUseCase は UseCase ID を設定
func WithActivityUseCase(usecaseID string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.UseCaseID = usecaseID
		}
	}
}

// WithActivityDescription は説明を設定
func WithActivityDescription(desc string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Description = desc
		}
	}
}

// WithActivityStatus はステータスを設定
func WithActivityStatus(status ActivityStatus) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Status = status
		}
	}
}

// WithActivityNodes はノードを設定
func WithActivityNodes(nodes []ActivityNode) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Nodes = nodes
		}
	}
}

// WithActivityTransitions は遷移を設定
func WithActivityTransitions(transitions []ActivityTransition) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Transitions = transitions
		}
	}
}

// WithActivityOwner はオーナーを設定
func WithActivityOwner(owner string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Metadata.Owner = owner
		}
	}
}

// WithActivityTags はタグを設定
func WithActivityTags(tags []string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Metadata.Tags = tags
		}
	}
}

// WithActivityRelatedDeliverables は関連成果物を設定
func WithActivityRelatedDeliverables(deliverableIDs []string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.RelatedDeliverables = deliverableIDs
		}
	}
}

// ===== Task/Activity 統合: 新規 EntityOption 関数群 =====

// WithActivityDependencies は依存関係を設定
func WithActivityDependencies(dependencies []string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Dependencies = dependencies
		}
	}
}

// WithActivityParent は親 Activity ID を設定
func WithActivityParent(parentID string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.ParentID = parentID
		}
	}
}

// WithActivityEstimateHours は見積もり時間を設定
func WithActivityEstimateHours(hours float64) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.EstimateHours = hours
		}
	}
}

// WithActivityActualHours は実績時間を設定
func WithActivityActualHours(hours float64) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.ActualHours = hours
		}
	}
}

// WithActivityAssignee は担当者を設定
func WithActivityAssignee(assignee string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Assignee = assignee
		}
	}
}

// WithActivityStartDate は開始日を設定
func WithActivityStartDate(startDate string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.StartDate = startDate
		}
	}
}

// WithActivityDueDate は期限日を設定
func WithActivityDueDate(dueDate string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.DueDate = dueDate
		}
	}
}

// WithActivityPriority は優先度を設定
func WithActivityPriority(priority ActivityPriority) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Priority = priority
		}
	}
}

// WithActivityWBSCode は WBS コードを設定
func WithActivityWBSCode(wbsCode string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.WBSCode = wbsCode
		}
	}
}

// WithActivityProgress は進捗率を設定
func WithActivityProgress(progress int) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Progress = progress
		}
	}
}

// WithActivityApprovalLevel は承認レベルを設定
func WithActivityApprovalLevel(level ApprovalLevel) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.ApprovalLevel = level
		}
	}
}
