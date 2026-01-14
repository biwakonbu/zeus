package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var approveCmd = &cobra.Command{
	Use:   "approve <id>",
	Short: "アイテムを承認",
	Long:  `指定されたIDのアイテムを承認します。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runApprove,
}

func init() {
	rootCmd.AddCommand(approveCmd)
}

func runApprove(cmd *cobra.Command, args []string) error {
	id := args[0]

	am := core.NewApprovalManager(".zeus")
	result, err := am.Approve(id)
	if err != nil {
		return err
	}

	if result.Success {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Approved: %s\n", green("✓"), result.ID)
	}

	return nil
}
