package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Zeus プロジェクトを初期化",
	Long:  `プロジェクトディレクトリに .zeus/ フォルダを作成し、Zeus プロジェクトを初期化します。`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("level", "l", "simple", "初期化レベル (simple|standard|advanced)")
}

func runInit(cmd *cobra.Command, args []string) error {
	level, _ := cmd.Flags().GetString("level")

	zeus := core.New(".")
	result, err := zeus.Init(level)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Zeus initialized successfully!\n", green("✓"))
	fmt.Printf("  Level: %s\n", result.Level)
	fmt.Printf("  Path:  %s\n", result.ZeusPath)

	return nil
}
