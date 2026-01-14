package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
	ctx := getContext(cmd)
	id := args[0]

	zeus := getZeus(cmd)
	result, err := zeus.Approve(ctx, id)
	if err != nil {
		return err
	}

	if result.Success {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Approved: %s\n", green("✓"), result.ID)
	}

	return nil
}
