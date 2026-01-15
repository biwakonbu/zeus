package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <entity> <name>",
	Short: "エンティティを追加",
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	entity := args[0]
	name := args[1]

	zeus := getZeus(cmd)
	result, err := zeus.Add(ctx, entity, name)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if result.NeedsApproval {
		// 承認待ちの場合
		fmt.Printf("%s %s '%s' は承認待ちキューに追加されました\n",
			yellow("⏳"), result.Entity, name)
		fmt.Printf("   承認ID: %s\n", result.ApprovalID)
		fmt.Println("   'zeus pending' で確認、'zeus approve <id>' で承認できます")
	} else {
		// 即時追加の場合
		fmt.Printf("%s Added %s: %s (ID: %s)\n",
			green("✓"), result.Entity, name, result.ID)
	}

	return nil
}
