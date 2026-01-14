package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var pendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "承認待ちアイテムを表示",
	Long:  `承認待ちのアイテム一覧を表示します。`,
	RunE:  runPending,
}

func init() {
	rootCmd.AddCommand(pendingCmd)
}

func runPending(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)
	pending, err := zeus.Pending(ctx)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(cyan("Pending Approvals"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	if len(pending) == 0 {
		fmt.Println("No pending approvals.")
		return nil
	}

	for _, item := range pending {
		levelColor := getLevelColor(item.Level)
		fmt.Printf("[%s] %s - %s\n", levelColor(string(item.Level)), yellow(item.ID), item.Description)
		fmt.Printf("    Type: %s | Created: %s\n", item.Type, item.CreatedAt)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Total: %d item(s)\n", len(pending))
	fmt.Println()
	fmt.Println("Use 'zeus approve <id>' to approve or 'zeus reject <id>' to reject.")

	return nil
}

func getLevelColor(level core.ApprovalLevel) func(a ...interface{}) string {
	switch level {
	case core.ApprovalApprove:
		return color.New(color.FgRed).SprintFunc()
	case core.ApprovalNotify:
		return color.New(color.FgYellow).SprintFunc()
	default:
		return color.New(color.FgGreen).SprintFunc()
	}
}
