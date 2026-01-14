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
	statusCmd.Flags().Bool("detail", false, "詳細表示")
}

func runStatus(cmd *cobra.Command, args []string) error {
	zeus := core.New(".")
	result, err := zeus.Status()
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(cyan("Zeus Project Status"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Project: %s\n", result.Project.Name)
	fmt.Printf("Health:  %s\n", result.State.Health)
	fmt.Println()
	fmt.Println("Tasks Summary:")
	fmt.Printf("  Total:       %d\n", result.State.Summary.TotalTasks)
	fmt.Printf("  Completed:   %d\n", result.State.Summary.Completed)
	fmt.Printf("  In Progress: %d\n", result.State.Summary.InProgress)
	fmt.Printf("  Pending:     %d\n", result.State.Summary.Pending)
	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}
