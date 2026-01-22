package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var umlCmd = &cobra.Command{
	Use:   "uml",
	Short: "UML図操作",
	Long:  `UML図に関する操作を行います。`,
}

var showUsecaseCmd = &cobra.Command{
	Use:   "show usecase",
	Short: "ユースケース図を表示",
	Long: `ユースケース図を表示します。

出力形式:
  text    - テキスト形式（デフォルト）
  mermaid - Mermaid形式（Markdown埋め込み可能）

オプション:
  --boundary <name>  システム境界名を指定
  --format <type>    出力形式（text|mermaid）
  --output <file>    出力ファイル（省略時は標準出力）

例:
  zeus uml show usecase                            # TEXT形式で標準出力
  zeus uml show usecase --format=mermaid           # Mermaid形式で標準出力
  zeus uml show usecase --boundary "ECサイト" -o uc.md  # システム境界を指定してファイル出力`,
	RunE: runShowUsecase,
}

var (
	umlBoundary string
	umlFormat   string
	umlOutput   string
)

func init() {
	rootCmd.AddCommand(umlCmd)
	umlCmd.AddCommand(showUsecaseCmd)

	showUsecaseCmd.Flags().StringVar(&umlBoundary, "boundary", "", "システム境界名")
	showUsecaseCmd.Flags().StringVarP(&umlFormat, "format", "f", "text", "出力形式 (text|mermaid)")
	showUsecaseCmd.Flags().StringVarP(&umlOutput, "output", "o", "", "出力ファイル（省略時は標準出力）")
}

func runShowUsecase(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// Actor と UseCase を取得
	actors, err := getActors(ctx, zeus)
	if err != nil {
		return fmt.Errorf("アクター取得失敗: %w", err)
	}

	usecases, err := getUseCases(ctx, zeus)
	if err != nil {
		return fmt.Errorf("ユースケース取得失敗: %w", err)
	}

	// データがない場合
	if len(actors) == 0 && len(usecases) == 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Println(cyan("Zeus UseCase Diagram"))
		fmt.Println("============================================================")
		fmt.Println("[INFO] アクターまたはユースケースがありません。")
		fmt.Println("============================================================")
		return nil
	}

	// 形式に応じて出力を生成
	var output string
	switch umlFormat {
	case "text":
		output = formatUsecaseText(actors, usecases, umlBoundary)
	case "mermaid":
		output = formatUsecaseMermaid(actors, usecases, umlBoundary)
	default:
		return fmt.Errorf("不明な出力形式: %s (text, mermaid のいずれかを指定してください)", umlFormat)
	}

	// 出力先に応じて出力
	if umlOutput != "" {
		if err := os.WriteFile(umlOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("ファイル出力失敗: %w", err)
		}
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s ユースケース図を %s に出力しました。\n", green("[SUCCESS]"), umlOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

// getActors はアクター一覧を取得
func getActors(ctx context.Context, zeus *core.Zeus) ([]core.ActorEntity, error) {
	handler, ok := zeus.GetRegistry().Get("actor")
	if !ok {
		return nil, fmt.Errorf("actor ハンドラーが見つかりません")
	}
	actorHandler, ok := handler.(*core.ActorHandler)
	if !ok {
		return nil, fmt.Errorf("actorHandler への型アサーションに失敗しました")
	}

	// FileStore から直接読み込み
	fileStore := zeus.FileStore()
	var actorsFile core.ActorsFile
	if !fileStore.Exists(ctx, "actors.yaml") {
		return []core.ActorEntity{}, nil
	}
	if err := fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		return nil, err
	}
	_ = actorHandler // 将来の拡張用
	return actorsFile.Actors, nil
}

// getUseCases はユースケース一覧を取得
func getUseCases(ctx context.Context, zeus *core.Zeus) ([]core.UseCaseEntity, error) {
	handler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		return nil, fmt.Errorf("usecase ハンドラーが見つかりません")
	}
	usecaseHandler, ok := handler.(*core.UseCaseHandler)
	if !ok {
		return nil, fmt.Errorf("usecaseHandler への型アサーションに失敗しました")
	}
	_ = usecaseHandler // 将来の拡張用

	// FileStore から直接読み込み
	fileStore := zeus.FileStore()
	if !fileStore.Exists(ctx, "usecases") {
		return []core.UseCaseEntity{}, nil
	}

	files, err := fileStore.ListDir(ctx, "usecases")
	if err != nil {
		return nil, err
	}

	usecases := make([]core.UseCaseEntity, 0)
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") && !strings.HasSuffix(file, ".yml") {
			continue
		}
		var usecase core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+file, &usecase); err != nil {
			continue
		}
		usecases = append(usecases, usecase)
	}

	return usecases, nil
}

// formatUsecaseText はテキスト形式でユースケース図を生成
func formatUsecaseText(actors []core.ActorEntity, usecases []core.UseCaseEntity, boundary string) string {
	var sb strings.Builder

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	sb.WriteString(cyan("Zeus UseCase Diagram"))
	sb.WriteString("\n")
	sb.WriteString("============================================================\n")

	// システム境界
	boundaryName := boundary
	if boundaryName == "" {
		boundaryName = "System"
	}
	fmt.Fprintf(&sb, "\n%s [ %s ]\n", white("System Boundary:"), boundaryName)
	sb.WriteString("------------------------------------------------------------\n")

	// アクター一覧
	fmt.Fprintf(&sb, "\n%s (%d)\n", green("Actors"), len(actors))
	for _, actor := range actors {
		typeIcon := getActorTypeIcon(actor.Type)
		fmt.Fprintf(&sb, "  %s %s [%s] (%s)\n", typeIcon, actor.Title, actor.ID, actor.Type)
	}

	// ユースケース一覧
	fmt.Fprintf(&sb, "\n%s (%d)\n", yellow("UseCases"), len(usecases))
	for _, uc := range usecases {
		statusIcon := getUseCaseStatusIcon(uc.Status)
		fmt.Fprintf(&sb, "  %s (%s) [%s] %s\n", statusIcon, uc.Title, uc.ID, uc.Status)

		// アクター関連
		for _, actorRef := range uc.Actors {
			roleIcon := "→"
			if actorRef.Role == core.ActorRolePrimary {
				roleIcon = "●→"
			}
			fmt.Fprintf(&sb, "      %s %s (%s)\n", roleIcon, actorRef.ActorID, actorRef.Role)
		}

		// リレーション
		for _, rel := range uc.Relations {
			relIcon := getRelationIcon(rel.Type)
			fmt.Fprintf(&sb, "      %s %s %s\n", relIcon, rel.Type, rel.TargetID)
			if rel.Condition != "" {
				fmt.Fprintf(&sb, "          condition: %s\n", rel.Condition)
			}
			if rel.ExtensionPoint != "" {
				fmt.Fprintf(&sb, "          extension-point: %s\n", rel.ExtensionPoint)
			}
		}
	}

	sb.WriteString("\n============================================================\n")
	fmt.Fprintf(&sb, "Total: %d actors, %d usecases\n", len(actors), len(usecases))

	return sb.String()
}

// formatUsecaseMermaid は Mermaid 形式でユースケース図を生成
func formatUsecaseMermaid(actors []core.ActorEntity, usecases []core.UseCaseEntity, boundary string) string {
	var sb strings.Builder

	// Note: Mermaid は標準でユースケース図をサポートしていないため、
	// flowchart で近似的に表現する
	sb.WriteString("```mermaid\n")
	sb.WriteString("flowchart LR\n")

	// システム境界
	boundaryName := boundary
	if boundaryName == "" {
		boundaryName = "System"
	}

	// アクター定義（左側）
	sb.WriteString("    %% Actors\n")
	for _, actor := range actors {
		// Mermaid の ID として使用するため、ハイフンをアンダースコアに置換
		mermaidID := strings.ReplaceAll(actor.ID, "-", "_")
		typeEmoji := getActorTypeEmoji(actor.Type)
		fmt.Fprintf(&sb, "    %s[%s %s]\n", mermaidID, typeEmoji, escapeForMermaid(actor.Title))
	}

	// システム境界サブグラフ
	fmt.Fprintf(&sb, "\n    subgraph %s[%s]\n", "boundary", escapeForMermaid(boundaryName))

	// ユースケース定義
	sb.WriteString("        %% UseCases\n")
	for _, uc := range usecases {
		mermaidID := strings.ReplaceAll(uc.ID, "-", "_")
		fmt.Fprintf(&sb, "        %s((%s))\n", mermaidID, escapeForMermaid(uc.Title))
	}

	sb.WriteString("    end\n")

	// アクターとユースケースの関連
	sb.WriteString("\n    %% Actor-UseCase Relations\n")
	for _, uc := range usecases {
		ucID := strings.ReplaceAll(uc.ID, "-", "_")
		for _, actorRef := range uc.Actors {
			actorID := strings.ReplaceAll(actorRef.ActorID, "-", "_")
			if actorRef.Role == core.ActorRolePrimary {
				fmt.Fprintf(&sb, "    %s ==> %s\n", actorID, ucID)
			} else {
				fmt.Fprintf(&sb, "    %s --> %s\n", actorID, ucID)
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
				fmt.Fprintf(&sb, "    %s -.->|include| %s\n", ucID, targetID)
			case core.RelationTypeExtend:
				label := "extend"
				if rel.Condition != "" {
					label = fmt.Sprintf("extend [%s]", rel.Condition)
				}
				fmt.Fprintf(&sb, "    %s -.->|%s| %s\n", targetID, escapeForMermaid(label), ucID)
			case core.RelationTypeGeneralize:
				fmt.Fprintf(&sb, "    %s -->|generalize| %s\n", ucID, targetID)
			}
		}
	}

	sb.WriteString("```\n")

	return sb.String()
}

// getActorTypeIcon はアクタータイプのアイコンを返す
func getActorTypeIcon(t core.ActorType) string {
	switch t {
	case core.ActorTypeHuman:
		return "[H]"
	case core.ActorTypeSystem:
		return "[S]"
	case core.ActorTypeTime:
		return "[T]"
	case core.ActorTypeDevice:
		return "[D]"
	case core.ActorTypeExternal:
		return "[E]"
	default:
		return "[?]"
	}
}

// getActorTypeEmoji はアクタータイプの絵文字を返す（Mermaid用）
func getActorTypeEmoji(t core.ActorType) string {
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

// getUseCaseStatusIcon はユースケースステータスのアイコンを返す
func getUseCaseStatusIcon(s core.UseCaseStatus) string {
	switch s {
	case core.UseCaseStatusDraft:
		return "[DRAFT]"
	case core.UseCaseStatusActive:
		return "[ACTIVE]"
	case core.UseCaseStatusDeprecated:
		return "[DEPRECATED]"
	default:
		return "[?]"
	}
}

// getRelationIcon はリレーションタイプのアイコンを返す
func getRelationIcon(t core.RelationType) string {
	switch t {
	case core.RelationTypeInclude:
		return "<<include>>"
	case core.RelationTypeExtend:
		return "<<extend>>"
	case core.RelationTypeGeneralize:
		return "<<generalize>>"
	default:
		return "--->"
	}
}

// escapeForMermaid は Mermaid 用にエスケープ
func escapeForMermaid(s string) string {
	// ダブルクォートとその他の特殊文字をエスケープ
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
