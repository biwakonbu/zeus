package analysis

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// WBSNodeType はノードの種類
type WBSNodeType string

const (
	WBSNodeTypeVision      WBSNodeType = "vision"
	WBSNodeTypeObjective   WBSNodeType = "objective"
	WBSNodeTypeDeliverable WBSNodeType = "deliverable"
	WBSNodeTypeTask        WBSNodeType = "task"
)

// WBSNode はWBS階層のノード
type WBSNode struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Type     WBSNodeType `json:"type"`
	WBSCode  string      `json:"wbs_code"`
	Status   string      `json:"status"`
	Progress int         `json:"progress"`
	Priority string      `json:"priority,omitempty"`
	Assignee string      `json:"assignee,omitempty"`
	Children []*WBSNode  `json:"children,omitempty"`

	// 内部用
	Depth int `json:"depth"`
}

// WBSTree はWBS階層全体
type WBSTree struct {
	Roots    []*WBSNode `json:"roots"`
	MaxDepth int        `json:"max_depth"`
	Stats    WBSStats   `json:"stats"`
}

// WBSStats はWBS統計
type WBSStats struct {
	TotalNodes   int `json:"total_nodes"`
	RootCount    int `json:"root_count"`
	LeafCount    int `json:"leaf_count"`
	MaxDepth     int `json:"max_depth"`
	AvgProgress  int `json:"avg_progress"`
	CompletedPct int `json:"completed_pct"`
}

// WBSBuilder はWBS階層を構築するビルダー
type WBSBuilder struct {
	tasks map[string]*TaskInfo
}

// NewWBSBuilder は新しいWBSBuilderを作成
func NewWBSBuilder(tasks []TaskInfo) *WBSBuilder {
	taskMap := make(map[string]*TaskInfo)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}
	return &WBSBuilder{tasks: taskMap}
}

// Build はParentIDからWBS階層を構築
func (w *WBSBuilder) Build(ctx context.Context) (*WBSTree, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ParentID の循環参照を検出
	if cycles := w.detectParentCycles(); len(cycles) > 0 {
		// 最初の循環をエラーメッセージに含める
		cycle := cycles[0]
		return nil, fmt.Errorf("parent cycle detected: %s", strings.Join(cycle, " -> "))
	}

	// ノードマップを作成
	nodeMap := make(map[string]*WBSNode)
	for id, task := range w.tasks {
		nodeMap[id] = &WBSNode{
			ID:       task.ID,
			Title:    task.Title,
			WBSCode:  task.WBSCode,
			Status:   task.Status,
			Progress: task.Progress,
			Priority: task.Priority,
			Assignee: task.Assignee,
			Children: []*WBSNode{},
			Depth:    0,
		}
	}

	// 親子関係を構築
	roots := []*WBSNode{}
	for id, task := range w.tasks {
		node := nodeMap[id]
		if task.ParentID == "" {
			// ルートノード
			roots = append(roots, node)
		} else if parent, exists := nodeMap[task.ParentID]; exists {
			// 親が存在する場合は子として追加
			parent.Children = append(parent.Children, node)
		} else {
			// 親が存在しない場合はルートとして扱う
			roots = append(roots, node)
		}
	}

	// WBSコードでソート
	sortNodesByWBSCode(roots)

	// 深さを計算
	maxDepth := 0
	for _, root := range roots {
		depth := calculateDepth(root, 0)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// 統計を計算
	stats := w.calculateStats(roots)

	return &WBSTree{
		Roots:    roots,
		MaxDepth: maxDepth,
		Stats:    stats,
	}, nil
}

// sortNodesByWBSCode はノードをWBSコード順にソート
func sortNodesByWBSCode(nodes []*WBSNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return compareWBSCodes(nodes[i].WBSCode, nodes[j].WBSCode) < 0
	})

	for _, node := range nodes {
		if len(node.Children) > 0 {
			sortNodesByWBSCode(node.Children)
		}
	}
}

// compareWBSCodes はWBSコードを比較（"1.2.3" < "1.2.10" < "1.3"）
func compareWBSCodes(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}

	for i := 0; i < minLen; i++ {
		numA := parseWBSPart(partsA[i])
		numB := parseWBSPart(partsB[i])
		if numA != numB {
			return numA - numB
		}
	}

	return len(partsA) - len(partsB)
}

// parseWBSPart はWBSコードの一部を数値に変換
func parseWBSPart(s string) int {
	var num int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		}
	}
	return num
}

// calculateDepth は再帰的に深さを計算
func calculateDepth(node *WBSNode, depth int) int {
	node.Depth = depth
	maxDepth := depth

	for _, child := range node.Children {
		childDepth := calculateDepth(child, depth+1)
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}

	return maxDepth
}

// calculateStats はWBS統計を計算
func (w *WBSBuilder) calculateStats(roots []*WBSNode) WBSStats {
	stats := WBSStats{
		RootCount: len(roots),
	}

	totalProgress := 0
	completedCount := 0
	var countNodes func(node *WBSNode) int
	countNodes = func(node *WBSNode) int {
		count := 1
		isLeaf := len(node.Children) == 0
		if isLeaf {
			stats.LeafCount++
		}

		totalProgress += node.Progress
		if node.Status == TaskStatusCompleted {
			completedCount++
		}

		if node.Depth > stats.MaxDepth {
			stats.MaxDepth = node.Depth
		}

		for _, child := range node.Children {
			count += countNodes(child)
		}
		return count
	}

	for _, root := range roots {
		stats.TotalNodes += countNodes(root)
	}

	if stats.TotalNodes > 0 {
		stats.AvgProgress = totalProgress / stats.TotalNodes
		stats.CompletedPct = (completedCount * 100) / stats.TotalNodes
	}

	return stats
}

// ToText はWBSツリーをテキスト形式で出力
func (tree *WBSTree) ToText() string {
	var sb strings.Builder

	sb.WriteString("Zeus WBS Structure\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	for i, root := range tree.Roots {
		isLast := i == len(tree.Roots)-1
		printWBSNode(&sb, root, "", isLast)
	}

	sb.WriteString("\n" + strings.Repeat("-", 60) + "\n")
	sb.WriteString("Statistics:\n")
	sb.WriteString("  Total tasks:    " + itoa(tree.Stats.TotalNodes) + "\n")
	sb.WriteString("  Root tasks:     " + itoa(tree.Stats.RootCount) + "\n")
	sb.WriteString("  Leaf tasks:     " + itoa(tree.Stats.LeafCount) + "\n")
	sb.WriteString("  Max depth:      " + itoa(tree.Stats.MaxDepth) + "\n")
	sb.WriteString("  Avg progress:   " + itoa(tree.Stats.AvgProgress) + "%\n")
	sb.WriteString("  Completed:      " + itoa(tree.Stats.CompletedPct) + "%\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	return sb.String()
}

// printWBSNode は再帰的にWBSノードを表示
func printWBSNode(sb *strings.Builder, node *WBSNode, prefix string, isLast bool) {
	// ツリー構造の描画
	connector := "├─ "
	if isLast {
		connector = "└─ "
	}
	if prefix == "" {
		connector = ""
	}

	// WBSコードがあれば表示
	wbsPrefix := ""
	if node.WBSCode != "" {
		wbsPrefix = node.WBSCode + " "
	}

	// 進捗バー
	progressBar := createProgressBar(node.Progress, 10)

	// ステータスアイコン
	statusIcon := getStatusIcon(node.Status)

	sb.WriteString(prefix + connector + wbsPrefix + node.Title + " " + progressBar + " " + statusIcon + "\n")

	// 子ノードの処理
	newPrefix := prefix
	if prefix != "" {
		if isLast {
			newPrefix += "   "
		} else {
			newPrefix += "│  "
		}
	}

	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		printWBSNode(sb, child, newPrefix+"  ", isChildLast)
	}
}

// createProgressBar は進捗バーを生成
func createProgressBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return "[" + bar + "] " + itoa(progress) + "%"
}

// getStatusIcon はステータスに応じたアイコンを返す
func getStatusIcon(status string) string {
	switch status {
	case TaskStatusCompleted:
		return "✓"
	case TaskStatusInProgress:
		return "●"
	case TaskStatusBlocked:
		return "✗"
	default:
		return "○"
	}
}

// itoa はintをstringに変換（シンプル実装）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// GenerateWBSCodes は階層構造に基づいてWBSコードを自動生成
func (tree *WBSTree) GenerateWBSCodes() {
	for i, root := range tree.Roots {
		assignWBSCode(root, itoa(i+1))
	}
}

// assignWBSCode は再帰的にWBSコードを割り当て
func assignWBSCode(node *WBSNode, code string) {
	node.WBSCode = code
	for i, child := range node.Children {
		childCode := code + "." + itoa(i+1)
		assignWBSCode(child, childCode)
	}
}

// detectParentCycles は ParentID による循環参照を検出（DFS アルゴリズム）
// 返り値: 検出された循環のリスト（各循環は参加ノードIDのスライス）
func (w *WBSBuilder) detectParentCycles() [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(nodeID string, path []string) bool
	dfs = func(nodeID string, path []string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true
		path = append(path, nodeID)

		task, exists := w.tasks[nodeID]
		if !exists || task.ParentID == "" {
			recStack[nodeID] = false
			return false
		}

		parentID := task.ParentID

		// 親タスクが存在しない場合はスキップ
		if _, exists := w.tasks[parentID]; !exists {
			recStack[nodeID] = false
			return false
		}

		if !visited[parentID] {
			if dfs(parentID, path) {
				return true
			}
		} else if recStack[parentID] {
			// 循環検出: 現在のパスから循環部分を抽出
			cycleStart := -1
			for i, id := range path {
				if id == parentID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(path[cycleStart:], parentID)
				cycles = append(cycles, cycle)
			}
		}

		recStack[nodeID] = false
		return false
	}

	// 全ノードを起点に DFS を実行
	for nodeID := range w.tasks {
		if !visited[nodeID] {
			dfs(nodeID, []string{})
		}
	}

	return cycles
}

// DetectParentCycles は外部から循環検出を呼び出すためのメソッド
// WBS構築前に検証したい場合に使用
func (w *WBSBuilder) DetectParentCycles() [][]string {
	return w.detectParentCycles()
}

// MultiEntityWBSBuilder は 10概念モデル対応の WBS ビルダー
// Vision → Objective → Deliverable → Task の階層構造を構築
type MultiEntityWBSBuilder struct {
	vision       *VisionInfo
	objectives   map[string]*ObjectiveInfo
	deliverables map[string]*DeliverableInfo
	tasks        map[string]*TaskInfo
}

// NewMultiEntityWBSBuilder は新しい MultiEntityWBSBuilder を作成
func NewMultiEntityWBSBuilder(
	vision *VisionInfo,
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	tasks []TaskInfo,
) *MultiEntityWBSBuilder {
	objMap := make(map[string]*ObjectiveInfo)
	for i := range objectives {
		objMap[objectives[i].ID] = &objectives[i]
	}

	delMap := make(map[string]*DeliverableInfo)
	for i := range deliverables {
		delMap[deliverables[i].ID] = &deliverables[i]
	}

	taskMap := make(map[string]*TaskInfo)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	return &MultiEntityWBSBuilder{
		vision:       vision,
		objectives:   objMap,
		deliverables: delMap,
		tasks:        taskMap,
	}
}

// Build は 10概念モデルの WBS 階層を構築
func (m *MultiEntityWBSBuilder) Build(ctx context.Context) (*WBSTree, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Vision ノード（L1）を作成
	var visionNode *WBSNode
	if m.vision != nil {
		visionNode = &WBSNode{
			ID:       m.vision.ID,
			Title:    m.vision.Title,
			Type:     WBSNodeTypeVision,
			Status:   m.vision.Status,
			Progress: 0, // Vision は進捗率なし
			Children: []*WBSNode{},
			Depth:    0,
		}
	}

	// Objective ノード（L2/L3）を作成
	objNodes := make(map[string]*WBSNode)
	for id, obj := range m.objectives {
		objNodes[id] = &WBSNode{
			ID:       obj.ID,
			Title:    obj.Title,
			Type:     WBSNodeTypeObjective,
			WBSCode:  obj.WBSCode,
			Status:   obj.Status,
			Progress: obj.Progress,
			Children: []*WBSNode{},
			Depth:    0,
		}
	}

	// Deliverable ノード（L4）を作成
	delNodes := make(map[string]*WBSNode)
	for id, del := range m.deliverables {
		delNodes[id] = &WBSNode{
			ID:       del.ID,
			Title:    del.Title,
			Type:     WBSNodeTypeDeliverable,
			Status:   del.Status,
			Progress: del.Progress,
			Children: []*WBSNode{},
			Depth:    0,
		}
	}

	// Task ノード（L5）を作成
	taskNodes := make(map[string]*WBSNode)
	for id, task := range m.tasks {
		taskNodes[id] = &WBSNode{
			ID:       task.ID,
			Title:    task.Title,
			Type:     WBSNodeTypeTask,
			WBSCode:  task.WBSCode,
			Status:   task.Status,
			Progress: task.Progress,
			Priority: task.Priority,
			Assignee: task.Assignee,
			Children: []*WBSNode{},
			Depth:    0,
		}
	}

	// 親子関係を構築
	// 1. Objective の親子関係（L2 → L3）
	l2Objectives := []*WBSNode{}
	for id, obj := range m.objectives {
		node := objNodes[id]
		if obj.ParentID == "" {
			// L2 (ルート Objective)
			l2Objectives = append(l2Objectives, node)
		} else if parent, exists := objNodes[obj.ParentID]; exists {
			// L3 (子 Objective)
			parent.Children = append(parent.Children, node)
		} else {
			// 親が見つからない場合は L2 として扱う
			l2Objectives = append(l2Objectives, node)
		}
	}

	// 2. Deliverable → Objective（L4 → L3/L2）
	orphanDeliverables := []*WBSNode{}
	for id, del := range m.deliverables {
		node := delNodes[id]
		if parent, exists := objNodes[del.ObjectiveID]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			orphanDeliverables = append(orphanDeliverables, node)
		}
	}

	// 3. Task → Deliverable または Objective（L5 → L4/L3）
	// Note: 現在の実装では Task の親参照は ParentID（タスク間）のみ
	// Deliverable/Objective への参照は別途対応が必要
	orphanTasks := []*WBSNode{}
	for id := range m.tasks {
		node := taskNodes[id]
		orphanTasks = append(orphanTasks, node)
	}

	// 4. Vision → L2 Objectives
	if visionNode != nil {
		visionNode.Children = l2Objectives
	}

	// ルートノードを決定
	roots := []*WBSNode{}
	if visionNode != nil {
		roots = append(roots, visionNode)
	} else {
		roots = l2Objectives
	}

	// 孤立した Deliverables と Tasks を追加
	roots = append(roots, orphanDeliverables...)
	roots = append(roots, orphanTasks...)

	// WBS コードでソート
	sortNodesByWBSCode(roots)

	// 深さを計算
	maxDepth := 0
	for _, root := range roots {
		depth := calculateDepth(root, 0)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// 統計を計算
	stats := m.calculateMultiEntityStats(roots)

	return &WBSTree{
		Roots:    roots,
		MaxDepth: maxDepth,
		Stats:    stats,
	}, nil
}

// calculateMultiEntityStats は 10概念モデルの WBS 統計を計算
func (m *MultiEntityWBSBuilder) calculateMultiEntityStats(roots []*WBSNode) WBSStats {
	stats := WBSStats{
		RootCount: len(roots),
	}

	totalProgress := 0
	progressCount := 0
	completedCount := 0

	var countNodes func(node *WBSNode) int
	countNodes = func(node *WBSNode) int {
		count := 1
		isLeaf := len(node.Children) == 0
		if isLeaf {
			stats.LeafCount++
		}

		// Vision 以外は進捗を計算
		if node.Type != WBSNodeTypeVision {
			totalProgress += node.Progress
			progressCount++
		}

		if node.Status == TaskStatusCompleted || node.Progress == 100 {
			completedCount++
		}

		if node.Depth > stats.MaxDepth {
			stats.MaxDepth = node.Depth
		}

		for _, child := range node.Children {
			count += countNodes(child)
		}
		return count
	}

	for _, root := range roots {
		stats.TotalNodes += countNodes(root)
	}

	if progressCount > 0 {
		stats.AvgProgress = totalProgress / progressCount
	}
	if stats.TotalNodes > 0 {
		stats.CompletedPct = (completedCount * 100) / stats.TotalNodes
	}

	return stats
}

// ToMermaid は WBS ツリーを Mermaid 形式で出力
func (tree *WBSTree) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TD\n")

	// ノードとエッジを出力
	for _, root := range tree.Roots {
		printMermaidNode(&sb, root, nil)
	}

	// スタイル定義
	sb.WriteString("\n")
	for _, root := range tree.Roots {
		printMermaidStyles(&sb, root)
	}

	sb.WriteString("```\n")

	return sb.String()
}

// printMermaidNode は再帰的に Mermaid ノードを出力
func printMermaidNode(sb *strings.Builder, node *WBSNode, parent *WBSNode) {
	// ノード ID をサニタイズ
	safeID := strings.ReplaceAll(node.ID, "-", "_")
	label := strings.ReplaceAll(node.Title, "\"", "'")

	// ノードの形状を種類に応じて変更
	var nodeShape string
	switch node.Type {
	case WBSNodeTypeVision:
		nodeShape = "((" + label + "))" // 円形
	case WBSNodeTypeObjective:
		nodeShape = "[" + label + "]" // 四角
	case WBSNodeTypeDeliverable:
		nodeShape = "([" + label + "])" // 角丸四角
	case WBSNodeTypeTask:
		nodeShape = "{{" + label + "}}" // 六角形
	default:
		nodeShape = "[" + label + "]"
	}

	sb.WriteString("    " + safeID + nodeShape + "\n")

	// 親へのエッジ
	if parent != nil {
		parentID := strings.ReplaceAll(parent.ID, "-", "_")
		sb.WriteString("    " + parentID + " --> " + safeID + "\n")
	}

	// 子ノードを再帰処理
	for _, child := range node.Children {
		printMermaidNode(sb, child, node)
	}
}

// printMermaidStyles は Mermaid スタイルを出力
func printMermaidStyles(sb *strings.Builder, node *WBSNode) {
	safeID := strings.ReplaceAll(node.ID, "-", "_")

	// 進捗に応じた色
	var color string
	if node.Progress >= 100 || node.Status == TaskStatusCompleted {
		color = "#90EE90" // 完了: 緑
	} else if node.Progress > 0 {
		color = "#FFFFE0" // 進行中: 黄
	} else {
		color = "#F5F5F5" // 未着手: グレー
	}

	// Vision は特別な色
	if node.Type == WBSNodeTypeVision {
		color = "#87CEEB" // 青
	}

	sb.WriteString("    style " + safeID + " fill:" + color + "\n")

	for _, child := range node.Children {
		printMermaidStyles(sb, child)
	}
}
