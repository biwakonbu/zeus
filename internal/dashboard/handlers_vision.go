package dashboard

import (
	"net/http"
	"os"

	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// Vision/Objective API 型定義
// =============================================================================

// VisionItem は Vision API のアイテム
type VisionItem struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Statement       string   `json:"statement"`
	SuccessCriteria []string `json:"success_criteria"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// VisionResponse は Vision API のレスポンス
type VisionResponse struct {
	Vision *VisionItem `json:"vision"`
}

// ObjectiveItem は Objective API のアイテム
type ObjectiveItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description,omitempty"`
	Goals        []string `json:"goals,omitempty"`
	Status       string   `json:"status"`
	Owner        string   `json:"owner,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	UseCaseCount int      `json:"usecase_count"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// ObjectivesResponse は Objective 一覧 API のレスポンス
type ObjectivesResponse struct {
	Objectives []ObjectiveItem `json:"objectives"`
	Total      int             `json:"total"`
}

// =============================================================================
// Vision/Objective API ハンドラー
// =============================================================================

// handleAPIVision は Vision 取得 API を処理
func (s *Server) handleAPIVision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	var vision core.Vision
	if err := fileStore.ReadYaml(ctx, "vision.yaml", &vision); err != nil {
		if os.IsNotExist(err) {
			// vision.yaml が存在しない場合は null を返す
			response := VisionResponse{Vision: nil}
			writeJSON(w, http.StatusOK, response)
			return
		}
		// パースエラーや権限エラーなどは 500 を返す
		writeError(w, http.StatusInternalServerError, "vision.yaml の読み込みに失敗: "+err.Error())
		return
	}

	response := VisionResponse{
		Vision: &VisionItem{
			ID:              vision.ID,
			Title:           vision.Title,
			Statement:       vision.Statement,
			SuccessCriteria: vision.SuccessCriteria,
			Status:          string(vision.Status),
			CreatedAt:       vision.Metadata.CreatedAt,
			UpdatedAt:       vision.Metadata.UpdatedAt,
		},
	}

	if response.Vision.SuccessCriteria == nil {
		response.Vision.SuccessCriteria = []string{}
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIObjectives は Objective 一覧 API を処理
func (s *Server) handleAPIObjectives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// objectives ディレクトリからファイル一覧を取得
	files, err := fileStore.ListDir(ctx, "objectives")
	if err != nil {
		response := ObjectivesResponse{
			Objectives: []ObjectiveItem{},
			Total:      0,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	// まず Objective を読み込み
	objEntities := make([]core.ObjectiveEntity, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var obj core.ObjectiveEntity
		if err := fileStore.ReadYaml(ctx, "objectives/"+file, &obj); err != nil {
			continue
		}
		objEntities = append(objEntities, obj)
	}

	// UseCase を読み込み、objective_id ごとにカウント
	usecaseCounts := make(map[string]int)
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
			if uc.ObjectiveID != "" {
				usecaseCounts[uc.ObjectiveID]++
			}
		}
	}

	// ObjectiveItem に変換
	objectives := make([]ObjectiveItem, 0, len(objEntities))
	for _, obj := range objEntities {
		item := ObjectiveItem{
			ID:           obj.ID,
			Title:        obj.Title,
			Description:  obj.Description,
			Goals:        obj.Goals,
			Status:       string(obj.Status),
			Owner:        obj.Owner,
			Tags:         obj.Tags,
			UseCaseCount: usecaseCounts[obj.ID],
			CreatedAt:    obj.Metadata.CreatedAt,
			UpdatedAt:    obj.Metadata.UpdatedAt,
		}
		if item.Goals == nil {
			item.Goals = []string{}
		}
		if item.Tags == nil {
			item.Tags = []string{}
		}
		objectives = append(objectives, item)
	}

	response := ObjectivesResponse{
		Objectives: objectives,
		Total:      len(objectives),
	}

	writeJSON(w, http.StatusOK, response)
}
