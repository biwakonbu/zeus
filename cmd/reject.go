package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var rejectCmd = &cobra.Command{
	Use:   "reject <id>",
	Short: "アイテムを却下",
	Long:  `指定されたIDのアイテムを却下します。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runReject,
}

func init() {
	rootCmd.AddCommand(rejectCmd)
	rejectCmd.Flags().StringP("reason", "r", "", "却下理由")
}

func runReject(cmd *cobra.Command, args []string) error {
	id := args[0]
	reason, _ := cmd.Flags().GetString("reason")

	am := core.NewApprovalManager(".zeus")
	result, err := am.Reject(id, reason)
	if err != nil {
		return err
	}

	if result.Success {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Rejected: %s\n", red("✗"), result.ID)
		if reason != "" {
			fmt.Printf("  Reason: %s\n", reason)
		}
	}

	return nil
}
