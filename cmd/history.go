package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "プロジェクト履歴を表示",
	Long:  `プロジェクト状態の履歴を表示します。`,
	RunE:  runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().IntP("limit", "n", 10, "表示件数")
}

func runHistory(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	limit, _ := cmd.Flags().GetInt("limit")

	zeus := getZeus(cmd)
	history, err := zeus.GetHistory(ctx, limit)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(cyan("Project History"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	if len(history) == 0 {
		fmt.Println("No history found. Create a snapshot with 'zeus snapshot create'.")
		return nil
	}

	for i, s := range history {
		healthColor := getHealthColor(s.State.Health)
		label := ""
		if s.Label != "" {
			label = fmt.Sprintf(" [%s]", s.Label)
		}
		fmt.Printf("%d. %s%s\n", i+1, s.Timestamp, label)
		fmt.Printf("   Health: %s | Tasks: %d (Completed: %d, In Progress: %d, Pending: %d)\n",
			healthColor(string(s.State.Health)),
			s.State.Summary.TotalTasks,
			s.State.Summary.Completed,
			s.State.Summary.InProgress,
			s.State.Summary.Pending)
		if len(s.State.Risks) > 0 {
			fmt.Printf("   Risks: %v\n", s.State.Risks)
		}
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("Use 'zeus snapshot restore <timestamp>' to restore a snapshot.")

	return nil
}

func getHealthColor(health core.HealthStatus) func(a ...interface{}) string {
	switch health {
	case core.HealthGood:
		return color.New(color.FgGreen).SprintFunc()
	case core.HealthFair:
		return color.New(color.FgYellow).SprintFunc()
	case core.HealthPoor:
		return color.New(color.FgRed).SprintFunc()
	default:
		return color.New(color.FgWhite).SprintFunc()
	}
}
