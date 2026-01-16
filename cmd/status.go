package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "プロジェクトの状態を表示",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)
	result, err := zeus.Status(ctx)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(cyan("Zeus Project Status"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Project: %s\n", result.Project.Name)
	fmt.Printf("Health:  %s\n", formatHealth(result.State.Health))
	fmt.Println()
	fmt.Println("Tasks Summary:")
	fmt.Printf("  Total:       %d\n", result.State.Summary.TotalTasks)
	fmt.Printf("  Completed:   %d\n", result.State.Summary.Completed)
	fmt.Printf("  In Progress: %d\n", result.State.Summary.InProgress)
	fmt.Printf("  Pending:     %d\n", result.State.Summary.Pending)

	if result.PendingApprovals > 0 {
		fmt.Println()
		fmt.Printf("Pending Approvals: %s\n", yellow(fmt.Sprintf("%d", result.PendingApprovals)))
	}

	if len(result.State.Risks) > 0 {
		fmt.Println()
		fmt.Println("Risks:")
		for _, risk := range result.State.Risks {
			fmt.Printf("  - %s\n", risk)
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}

func formatHealth(health core.HealthStatus) string {
	switch health {
	case core.HealthGood:
		return color.GreenString(string(health))
	case core.HealthFair:
		return color.YellowString(string(health))
	case core.HealthPoor:
		return color.RedString(string(health))
	default:
		return string(health)
	}
}
