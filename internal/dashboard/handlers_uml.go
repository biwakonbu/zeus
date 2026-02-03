package dashboard

import (
	"net/http"
	"strings"

	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// Actor/UseCase API 型定義
// =============================================================================

// ActorItem はアクター API のアイテム
type ActorItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ActorsResponse はアクター一覧 API のレスポンス
type ActorsResponse struct {
	Actors []ActorItem `json:"actors"`
	Total  int         `json:"total"`
}

// UseCaseActorRefItem はユースケースアクター参照 API のアイテム
type UseCaseActorRefItem struct {
	ActorID string `json:"actor_id"`
	Role    string `json:"role"`
}

// UseCaseRelationItem はユースケースリレーション API のアイテム
type UseCaseRelationItem struct {
	Type           string `json:"type"`
	TargetID       string `json:"target_id"`
	Condition      string `json:"condition,omitempty"`
	ExtensionPoint string `json:"extension_point,omitempty"`
}

// AlternativeFlowItem は代替フロー API のアイテム
type AlternativeFlowItem struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Condition string   `json:"condition"`
	Steps     []string `json:"steps"`
	RejoinsAt string   `json:"rejoins_at,omitempty"`
}

// ExceptionFlowItem は例外フロー API のアイテム
type ExceptionFlowItem struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Trigger string   `json:"trigger"`
	Steps   []string `json:"steps"`
	Outcome string   `json:"outcome,omitempty"`
}

// UseCaseScenarioItem はシナリオ API のアイテム
type UseCaseScenarioItem struct {
	Preconditions    []string              `json:"preconditions,omitempty"`
	Trigger          string                `json:"trigger,omitempty"`
	MainFlow         []string              `json:"main_flow,omitempty"`
	AlternativeFlows []AlternativeFlowItem `json:"alternative_flows,omitempty"`
	ExceptionFlows   []ExceptionFlowItem   `json:"exception_flows,omitempty"`
	Postconditions   []string              `json:"postconditions,omitempty"`
}

// UseCaseItem はユースケース API のアイテム
type UseCaseItem struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Status      string                `json:"status"`
	ObjectiveID string                `json:"objective_id,omitempty"`
	SubsystemID string                `json:"subsystem_id,omitempty"`
	Actors      []UseCaseActorRefItem `json:"actors"`
	Relations   []UseCaseRelationItem `json:"relations"`
	Scenario    *UseCaseScenarioItem  `json:"scenario,omitempty"`
}

// UseCasesResponse はユースケース一覧 API のレスポンス
type UseCasesResponse struct {
	UseCases []UseCaseItem `json:"usecases"`
	Total    int           `json:"total"`
}

// UseCaseDiagramResponse はユースケース図 API のレスポンス
type UseCaseDiagramResponse struct {
	Actors   []ActorItem   `json:"actors"`
	UseCases []UseCaseItem `json:"usecases"`
	Boundary string        `json:"boundary"`
	Mermaid  string        `json:"mermaid"`
}

// SubsystemItem はサブシステム API のアイテム
type SubsystemItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SubsystemsResponse はサブシステム一覧 API のレスポンス
type SubsystemsResponse struct {
	Subsystems []SubsystemItem `json:"subsystems"`
	Total      int             `json:"total"`
}

// =============================================================================
// Activity API 型定義
// =============================================================================

// ActivityNodeItem はアクティビティノード API のアイテム
type ActivityNodeItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// ActivityTransitionItem はアクティビティ遷移 API のアイテム
type ActivityTransitionItem struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Guard  string `json:"guard,omitempty"`
}

// ActivityItem はアクティビティ API のアイテム
type ActivityItem struct {
	ID           string                   `json:"id"`
	Title        string                   `json:"title"`
	Description  string                   `json:"description,omitempty"`
	UseCaseID    string                   `json:"usecase_id,omitempty"`
	UseCaseTitle string                   `json:"usecase_title,omitempty"`
	Status       string                   `json:"status"`
	Nodes        []ActivityNodeItem       `json:"nodes"`
	Transitions  []ActivityTransitionItem `json:"transitions"`
	CreatedAt    string                   `json:"created_at"`
	UpdatedAt    string                   `json:"updated_at"`
}

// ActivitiesResponse はアクティビティ一覧 API のレスポンス
type ActivitiesResponse struct {
	Activities []ActivityItem `json:"activities"`
	Total      int            `json:"total"`
}

// ActivityDiagramResponse はアクティビティ図 API のレスポンス
type ActivityDiagramResponse struct {
	Activity *ActivityItem `json:"activity,omitempty"`
	Mermaid  string        `json:"mermaid"`
}

// =============================================================================
// Actor/UseCase API ハンドラー
// =============================================================================

// handleAPIActors はアクター一覧 API を処理
func (s *Server) handleAPIActors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	var actorsFile core.ActorsFile
	if err := fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		actorsFile = core.ActorsFile{Actors: []core.ActorEntity{}}
	}

	actors := make([]ActorItem, len(actorsFile.Actors))
	for i, a := range actorsFile.Actors {
		actors[i] = ActorItem{
			ID:          a.ID,
			Title:       a.Title,
			Type:        string(a.Type),
			Description: a.Description,
		}
	}

	response := ActorsResponse{
		Actors: actors,
		Total:  len(actors),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIUseCases はユースケース一覧 API を処理
func (s *Server) handleAPIUseCases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// usecases ディレクトリからファイル一覧を取得
	files, err := fileStore.ListDir(ctx, "usecases")
	if err != nil {
		// ディレクトリが存在しない場合は空リストを返す
		response := UseCasesResponse{
			UseCases: []UseCaseItem{},
			Total:    0,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	usecases := make([]UseCaseItem, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var uc core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+file, &uc); err != nil {
			continue
		}

		// アクター参照の変換
		actors := make([]UseCaseActorRefItem, len(uc.Actors))
		for j, ar := range uc.Actors {
			actors[j] = UseCaseActorRefItem{
				ActorID: ar.ActorID,
				Role:    string(ar.Role),
			}
		}

		// リレーションの変換
		relations := make([]UseCaseRelationItem, len(uc.Relations))
		for j, rel := range uc.Relations {
			relations[j] = UseCaseRelationItem{
				Type:           string(rel.Type),
				TargetID:       rel.TargetID,
				Condition:      rel.Condition,
				ExtensionPoint: rel.ExtensionPoint,
			}
		}

		// シナリオの変換
		scenario := convertUseCaseScenario(&uc.Scenario)

		usecases = append(usecases, UseCaseItem{
			ID:          uc.ID,
			Title:       uc.Title,
			Description: uc.Description,
			Status:      string(uc.Status),
			ObjectiveID: uc.ObjectiveID,
			SubsystemID: uc.SubsystemID,
			Actors:      actors,
			Relations:   relations,
			Scenario:    scenario,
		})
	}

	response := UseCasesResponse{
		UseCases: usecases,
		Total:    len(usecases),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIUseCaseDiagram はユースケース図 API を処理
func (s *Server) handleAPIUseCaseDiagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// クエリパラメータからシステム境界名を取得
	boundary := r.URL.Query().Get("boundary")
	if boundary == "" {
		boundary = "System"
	}

	// アクターを取得
	var actorsFile core.ActorsFile
	if err := fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		actorsFile = core.ActorsFile{Actors: []core.ActorEntity{}}
	}

	actors := make([]ActorItem, len(actorsFile.Actors))
	for i, a := range actorsFile.Actors {
		actors[i] = ActorItem{
			ID:          a.ID,
			Title:       a.Title,
			Type:        string(a.Type),
			Description: a.Description,
		}
	}

	// ユースケースを取得
	files, _ := fileStore.ListDir(ctx, "usecases")
	usecases := make([]UseCaseItem, 0)
	ucEntities := make([]core.UseCaseEntity, 0)

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var uc core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+file, &uc); err != nil {
			continue
		}

		ucEntities = append(ucEntities, uc)

		// アクター参照の変換
		ucActors := make([]UseCaseActorRefItem, len(uc.Actors))
		for j, ar := range uc.Actors {
			ucActors[j] = UseCaseActorRefItem{
				ActorID: ar.ActorID,
				Role:    string(ar.Role),
			}
		}

		// リレーションの変換
		relations := make([]UseCaseRelationItem, len(uc.Relations))
		for j, rel := range uc.Relations {
			relations[j] = UseCaseRelationItem{
				Type:           string(rel.Type),
				TargetID:       rel.TargetID,
				Condition:      rel.Condition,
				ExtensionPoint: rel.ExtensionPoint,
			}
		}

		// シナリオの変換
		scenario := convertUseCaseScenario(&uc.Scenario)

		usecases = append(usecases, UseCaseItem{
			ID:          uc.ID,
			Title:       uc.Title,
			Description: uc.Description,
			Status:      string(uc.Status),
			ObjectiveID: uc.ObjectiveID,
			SubsystemID: uc.SubsystemID,
			Actors:      ucActors,
			Relations:   relations,
			Scenario:    scenario,
		})
	}

	// Mermaid 形式でユースケース図を生成
	mermaid := generateUseCaseMermaid(actorsFile.Actors, ucEntities, boundary)

	response := UseCaseDiagramResponse{
		Actors:   actors,
		UseCases: usecases,
		Boundary: boundary,
		Mermaid:  mermaid,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPISubsystems はサブシステム一覧 API を処理
func (s *Server) handleAPISubsystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// subsystems.yaml からサブシステム一覧を取得
	var subsystemsFile core.SubsystemsFile
	if err := fileStore.ReadYaml(ctx, "subsystems.yaml", &subsystemsFile); err != nil {
		// ファイルが存在しない場合は空リストを返す
		subsystemsFile = core.SubsystemsFile{Subsystems: []core.SubsystemEntity{}}
	}

	subsystems := make([]SubsystemItem, len(subsystemsFile.Subsystems))
	for i, sub := range subsystemsFile.Subsystems {
		subsystems[i] = SubsystemItem{
			ID:          sub.ID,
			Name:        sub.Name,
			Description: sub.Description,
		}
	}

	response := SubsystemsResponse{
		Subsystems: subsystems,
		Total:      len(subsystems),
	}

	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// Activity API ハンドラー
// =============================================================================

// handleAPIActivities はアクティビティ一覧 API を処理
func (s *Server) handleAPIActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// activities ディレクトリからファイル一覧を取得
	files, err := fileStore.ListDir(ctx, "activities")
	if err != nil {
		// ディレクトリが存在しない場合は空リストを返す
		response := ActivitiesResponse{
			Activities: []ActivityItem{},
			Total:      0,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	// まずアクティビティを読み込み、使用されているユースケースIDを収集
	actEntities := make([]core.ActivityEntity, 0)
	usecaseIDs := make(map[string]struct{})

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var act core.ActivityEntity
		if err := fileStore.ReadYaml(ctx, "activities/"+file, &act); err != nil {
			continue
		}
		actEntities = append(actEntities, act)
		if act.UseCaseID != "" {
			usecaseIDs[act.UseCaseID] = struct{}{}
		}
	}

	// ユースケースID→タイトルのマップを作成
	usecaseTitles := make(map[string]string)
	if len(usecaseIDs) > 0 {
		ucFiles, err := fileStore.ListDir(ctx, "usecases")
		if err == nil {
			for _, ucFile := range ucFiles {
				if !hasYamlSuffix(ucFile) {
					continue
				}
				var uc core.UseCaseEntity
				if err := fileStore.ReadYaml(ctx, "usecases/"+ucFile, &uc); err != nil {
					continue
				}
				// 使用されているIDのみマップに追加
				if _, needed := usecaseIDs[uc.ID]; needed {
					usecaseTitles[uc.ID] = uc.Title
				}
			}
		}
	}

	// ActivityItem に変換
	activities := make([]ActivityItem, 0, len(actEntities))
	for _, act := range actEntities {
		// ノードの変換
		nodes := make([]ActivityNodeItem, len(act.Nodes))
		for j, n := range act.Nodes {
			nodes[j] = ActivityNodeItem{
				ID:   n.ID,
				Type: string(n.Type),
				Name: n.Name,
			}
		}

		// 遷移の変換
		transitions := make([]ActivityTransitionItem, len(act.Transitions))
		for j, t := range act.Transitions {
			transitions[j] = ActivityTransitionItem{
				ID:     t.ID,
				Source: t.Source,
				Target: t.Target,
				Guard:  t.Guard,
			}
		}

		// ユースケースタイトルを取得
		usecaseTitle := ""
		if act.UseCaseID != "" {
			usecaseTitle = usecaseTitles[act.UseCaseID]
		}

		activities = append(activities, ActivityItem{
			ID:           act.ID,
			Title:        act.Title,
			Description:  act.Description,
			UseCaseID:    act.UseCaseID,
			UseCaseTitle: usecaseTitle,
			Status:       string(act.Status),
			Nodes:        nodes,
			Transitions:  transitions,
			CreatedAt:    act.Metadata.CreatedAt,
			UpdatedAt:    act.Metadata.UpdatedAt,
		})
	}

	response := ActivitiesResponse{
		Activities: activities,
		Total:      len(activities),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIActivityDiagram はアクティビティ図 API を処理
func (s *Server) handleAPIActivityDiagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// クエリパラメータからアクティビティIDを取得（必須）
	activityID := r.URL.Query().Get("id")
	if activityID == "" {
		writeError(w, http.StatusBadRequest, "id パラメータが必要です")
		return
	}

	var response ActivityDiagramResponse

	// 特定のアクティビティを取得
	var act core.ActivityEntity
	if err := fileStore.ReadYaml(ctx, "activities/"+activityID+".yaml", &act); err != nil {
		writeError(w, http.StatusNotFound, "アクティビティが見つかりません: "+activityID)
		return
	}

	// ノードの変換
	nodes := make([]ActivityNodeItem, len(act.Nodes))
	for j, n := range act.Nodes {
		nodes[j] = ActivityNodeItem{
			ID:   n.ID,
			Type: string(n.Type),
			Name: n.Name,
		}
	}

	// 遷移の変換
	transitions := make([]ActivityTransitionItem, len(act.Transitions))
	for j, t := range act.Transitions {
		transitions[j] = ActivityTransitionItem{
			ID:     t.ID,
			Source: t.Source,
			Target: t.Target,
			Guard:  t.Guard,
		}
	}

	// ユースケースタイトルを取得
	usecaseTitle := ""
	if act.UseCaseID != "" {
		var uc core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+act.UseCaseID+".yaml", &uc); err == nil {
			usecaseTitle = uc.Title
		}
	}

	activityItem := &ActivityItem{
		ID:           act.ID,
		Title:        act.Title,
		Description:  act.Description,
		UseCaseID:    act.UseCaseID,
		UseCaseTitle: usecaseTitle,
		Status:       string(act.Status),
		Nodes:        nodes,
		Transitions:  transitions,
		CreatedAt:    act.Metadata.CreatedAt,
		UpdatedAt:    act.Metadata.UpdatedAt,
	}

	response.Activity = activityItem
	response.Mermaid = generateActivityMermaid(&act)

	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// UML ヘルパー関数
// =============================================================================

// convertUseCaseScenario は core.UseCaseScenario を UseCaseScenarioItem に変換
func convertUseCaseScenario(scenario *core.UseCaseScenario) *UseCaseScenarioItem {
	// シナリオが空の場合は nil を返す
	if scenario == nil ||
		(len(scenario.Preconditions) == 0 &&
			scenario.Trigger == "" &&
			len(scenario.MainFlow) == 0 &&
			len(scenario.AlternativeFlows) == 0 &&
			len(scenario.ExceptionFlows) == 0 &&
			len(scenario.Postconditions) == 0) {
		return nil
	}

	// 代替フローの変換
	altFlows := make([]AlternativeFlowItem, len(scenario.AlternativeFlows))
	for i, af := range scenario.AlternativeFlows {
		altFlows[i] = AlternativeFlowItem{
			ID:        af.ID,
			Name:      af.Name,
			Condition: af.Condition,
			Steps:     af.Steps,
			RejoinsAt: af.RejoinsAt,
		}
	}

	// 例外フローの変換
	excFlows := make([]ExceptionFlowItem, len(scenario.ExceptionFlows))
	for i, ef := range scenario.ExceptionFlows {
		excFlows[i] = ExceptionFlowItem{
			ID:      ef.ID,
			Name:    ef.Name,
			Trigger: ef.Trigger,
			Steps:   ef.Steps,
			Outcome: ef.Outcome,
		}
	}

	return &UseCaseScenarioItem{
		Preconditions:    scenario.Preconditions,
		Trigger:          scenario.Trigger,
		MainFlow:         scenario.MainFlow,
		AlternativeFlows: altFlows,
		ExceptionFlows:   excFlows,
		Postconditions:   scenario.Postconditions,
	}
}

// generateUseCaseMermaid は Mermaid 形式でユースケース図を生成
func generateUseCaseMermaid(actors []core.ActorEntity, usecases []core.UseCaseEntity, boundary string) string {
	var sb strings.Builder

	sb.WriteString("flowchart LR\n")

	// アクター定義
	sb.WriteString("    %% Actors\n")
	for _, actor := range actors {
		mermaidID := strings.ReplaceAll(actor.ID, "-", "_")
		typeEmoji := actorTypeEmoji(actor.Type)
		sb.WriteString("    " + mermaidID + "[" + typeEmoji + " " + escapeForMermaidDiagram(actor.Title) + "]\n")
	}

	// システム境界サブグラフ
	sb.WriteString("\n    subgraph boundary[" + escapeForMermaidDiagram(boundary) + "]\n")

	// ユースケース定義
	sb.WriteString("        %% UseCases\n")
	for _, uc := range usecases {
		mermaidID := strings.ReplaceAll(uc.ID, "-", "_")
		sb.WriteString("        " + mermaidID + "((" + escapeForMermaidDiagram(uc.Title) + "))\n")
	}

	sb.WriteString("    end\n")

	// アクターとユースケースの関連
	sb.WriteString("\n    %% Actor-UseCase Relations\n")
	for _, uc := range usecases {
		ucID := strings.ReplaceAll(uc.ID, "-", "_")
		for _, actorRef := range uc.Actors {
			actorID := strings.ReplaceAll(actorRef.ActorID, "-", "_")
			if actorRef.Role == core.ActorRolePrimary {
				sb.WriteString("    " + actorID + " ==> " + ucID + "\n")
			} else {
				sb.WriteString("    " + actorID + " --> " + ucID + "\n")
			}
		}
	}

	// ユースケース間のリレーション
	sb.WriteString("\n    %% UseCase Relations\n")
	for _, uc := range usecases {
		ucID := strings.ReplaceAll(uc.ID, "-", "_")
		for _, rel := range uc.Relations {
			targetID := strings.ReplaceAll(rel.TargetID, "-", "_")
			switch rel.Type {
			case core.RelationTypeInclude:
				sb.WriteString("    " + ucID + " -.->|include| " + targetID + "\n")
			case core.RelationTypeExtend:
				label := "extend"
				if rel.Condition != "" {
					label = "extend [" + rel.Condition + "]"
				}
				sb.WriteString("    " + targetID + " -.->|" + escapeForMermaidDiagram(label) + "| " + ucID + "\n")
			case core.RelationTypeGeneralize:
				sb.WriteString("    " + ucID + " -->|generalize| " + targetID + "\n")
			}
		}
	}

	return sb.String()
}

// actorTypeEmoji はアクタータイプの絵文字を返す
func actorTypeEmoji(t core.ActorType) string {
	switch t {
	case core.ActorTypeHuman:
		return "👤"
	case core.ActorTypeSystem:
		return "🖥️"
	case core.ActorTypeTime:
		return "⏰"
	case core.ActorTypeDevice:
		return "📱"
	case core.ActorTypeExternal:
		return "🌐"
	default:
		return "❓"
	}
}

// generateActivityMermaid は Mermaid 形式でアクティビティ図を生成
func generateActivityMermaid(act *core.ActivityEntity) string {
	var sb strings.Builder

	sb.WriteString("flowchart TD\n")
	sb.WriteString("    %% " + escapeForMermaidDiagram(act.Title) + "\n\n")

	// ノード定義
	sb.WriteString("    %% Nodes\n")
	for _, node := range act.Nodes {
		mermaidID := strings.ReplaceAll(node.ID, "-", "_")
		label := node.Name
		if label == "" {
			label = string(node.Type)
		}

		switch node.Type {
		case core.ActivityNodeTypeInitial:
			// 開始ノード（黒丸）→ 円形
			sb.WriteString("    " + mermaidID + "((●))\n")
		case core.ActivityNodeTypeFinal:
			// 終了ノード（二重丸）→ 二重円形
			sb.WriteString("    " + mermaidID + "(((◉)))\n")
		case core.ActivityNodeTypeAction:
			// アクション（角丸四角形）
			sb.WriteString("    " + mermaidID + "[" + escapeForMermaidDiagram(label) + "]\n")
		case core.ActivityNodeTypeDecision, core.ActivityNodeTypeMerge:
			// 分岐/合流（ひし形）
			sb.WriteString("    " + mermaidID + "{" + escapeForMermaidDiagram(label) + "}\n")
		case core.ActivityNodeTypeFork, core.ActivityNodeTypeJoin:
			// 並列分岐/合流（太い横線）→ サブグラフで表現
			sb.WriteString("    " + mermaidID + "[/━━━/]\n")
		default:
			sb.WriteString("    " + mermaidID + "[" + escapeForMermaidDiagram(label) + "]\n")
		}
	}

	// 遷移定義
	sb.WriteString("\n    %% Transitions\n")
	for _, trans := range act.Transitions {
		sourceID := strings.ReplaceAll(trans.Source, "-", "_")
		targetID := strings.ReplaceAll(trans.Target, "-", "_")

		if trans.Guard != "" {
			// ガード条件付き
			sb.WriteString("    " + sourceID + " -->|" + escapeForMermaidDiagram(trans.Guard) + "| " + targetID + "\n")
		} else {
			sb.WriteString("    " + sourceID + " --> " + targetID + "\n")
		}
	}

	return sb.String()
}
