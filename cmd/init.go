package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Zeus プロジェクトを初期化",
	Long:  `プロジェクトディレクトリに .zeus/ フォルダを作成し、Zeus プロジェクトを初期化します。`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	// フラグなし（--level 削除済み）
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)

	zeus := getZeus(cmd)
	result, err := zeus.Init(ctx)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Zeus initialized successfully!\n", green("✓"))
	fmt.Printf("  Path:  %s\n", result.ZeusPath)

	return nil
}
