package cmd

import (
	"fmt"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// usecase link コマンドのフラグ
var (
	linkInclude        string
	linkExtend         string
	linkGeneralize     string
	linkCondition      string
	linkExtensionPoint string
)

// usecase add-actor コマンドのフラグ
var (
	usecaseAddActorRole string
)

var usecaseCmd = &cobra.Command{
	Use:   "usecase",
	Short: "ユースケース操作",
	Long:  `ユースケースに関する操作を行います。`,
}

var addActorCmd = &cobra.Command{
	Use:   "add-actor <usecase-id> <actor-id>",
	Short: "ユースケースにアクターを関連付け",
	Long: `ユースケースにアクターを関連付けます。

オプション:
  --role    アクターの役割（primary: 主要アクター, secondary: 副次アクター）

例:
  zeus usecase add-actor uc-setup actor-001
  zeus usecase add-actor uc-setup actor-002 --role secondary`,
	Args: cobra.ExactArgs(2),
	RunE: runAddActor,
}

var linkCmd = &cobra.Command{
	Use:   "link <usecase-id>",
	Short: "ユースケース間の関係を追加",
	Long: `ユースケース間の関係を追加します。

関係タイプ:
  --include       include 関係（対象 UseCase を必ず含む）
  --extend        extend 関係（対象 UseCase を条件付きで拡張）
  --generalize    generalize 関係（対象 UseCase を汎化）

extend オプション:
  --condition         extend の条件
  --extension-point   extend の拡張点

例:
  zeus usecase link uc-setup --include uc-model
  zeus usecase link uc-setup --extend uc-overview --condition "オプション選択時" --extension-point "支払い方法選択"
  zeus usecase link uc-setup --generalize uc-govern`,
	Args: cobra.ExactArgs(1),
	RunE: runLink,
}

func init() {
	rootCmd.AddCommand(usecaseCmd)
	usecaseCmd.AddCommand(linkCmd)
	usecaseCmd.AddCommand(addActorCmd)

	// link コマンドのフラグ
	linkCmd.Flags().StringVar(&linkInclude, "include", "", "include 先 UseCase ID")
	linkCmd.Flags().StringVar(&linkExtend, "extend", "", "extend 先 UseCase ID")
	linkCmd.Flags().StringVar(&linkGeneralize, "generalize", "", "generalize 先 UseCase ID")
	linkCmd.Flags().StringVar(&linkCondition, "condition", "", "extend の条件")
	linkCmd.Flags().StringVar(&linkExtensionPoint, "extension-point", "", "extend の拡張点")

	// add-actor コマンドのフラグ
	addActorCmd.Flags().StringVar(&usecaseAddActorRole, "role", "primary", "アクターの役割 (primary|secondary)")
}

func runAddActor(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	usecaseID := args[0]
	actorID := args[1]

	zeus := getZeus(cmd)

	// Registry から UseCaseHandler を取得
	handler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		return fmt.Errorf("usecase ハンドラーが見つかりません")
	}
	usecaseHandler, ok := handler.(*core.UseCaseHandler)
	if !ok {
		return fmt.Errorf("usecaseHandler への型アサーションに失敗しました")
	}

	// アクター参照を作成
	actorRef := core.UseCaseActorRef{
		ActorID: actorID,
		Role:    core.ActorRole(usecaseAddActorRole),
	}

	// アクターを追加
	if err := usecaseHandler.AddActor(ctx, usecaseID, actorRef); err != nil {
		return err
	}

	// 出力
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Added actor %s to usecase %s (role: %s)\n",
		green("✓"), actorID, usecaseID, usecaseAddActorRole)

	return nil
}

func runLink(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	usecaseID := args[0]

	zeus := getZeus(cmd)

	// Registry から UseCaseHandler を取得
	handler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		return fmt.Errorf("usecase ハンドラーが見つかりません")
	}
	usecaseHandler, ok := handler.(*core.UseCaseHandler)
	if !ok {
		return fmt.Errorf("usecaseHandler への型アサーションに失敗しました")
	}

	// 関係タイプを判定
	var relation core.UseCaseRelation
	var relTypeStr string

	if linkInclude != "" {
		relation = core.UseCaseRelation{
			Type:     core.RelationTypeInclude,
			TargetID: linkInclude,
		}
		relTypeStr = "include"
	} else if linkExtend != "" {
		relation = core.UseCaseRelation{
			Type:           core.RelationTypeExtend,
			TargetID:       linkExtend,
			Condition:      linkCondition,
			ExtensionPoint: linkExtensionPoint,
		}
		relTypeStr = "extend"
	} else if linkGeneralize != "" {
		relation = core.UseCaseRelation{
			Type:     core.RelationTypeGeneralize,
			TargetID: linkGeneralize,
		}
		relTypeStr = "generalize"
	} else {
		return fmt.Errorf("--include, --extend, --generalize のいずれかを指定してください")
	}

	// 関係を追加
	if err := usecaseHandler.AddRelation(ctx, usecaseID, relation); err != nil {
		return err
	}

	// 出力
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Added %s relation: %s -> %s\n",
		green("✓"), relTypeStr, usecaseID, relation.TargetID)

	return nil
}
