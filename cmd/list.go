package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var listCmd = &cobra.Command{
	Use:   "list [entity]",
	Short: "エンティティ一覧を表示",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("status", "s", "", "ステータスでフィルタ")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	entity := ""
	if len(args) > 0 {
		entity = args[0]
	}

	zeus := getZeus(cmd)
	result, err := zeus.List(ctx, entity)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s (%d items)\n", cyan("Tasks"), result.Total)
	fmt.Println("────────────────────────────────────────")

	for _, task := range result.Items {
		statusColor := getStatusColor(task.Status)
		fmt.Printf("[%s] %s - %s\n", statusColor(string(task.Status)), task.ID, task.Title)
	}

	return nil
}

func getStatusColor(status core.TaskStatus) func(a ...interface{}) string {
	switch status {
	case core.TaskStatusCompleted:
		return color.New(color.FgGreen).SprintFunc()
	case core.TaskStatusInProgress:
		return color.New(color.FgYellow).SprintFunc()
	case core.TaskStatusBlocked:
		return color.New(color.FgRed).SprintFunc()
	default:
		return color.New(color.FgWhite).SprintFunc()
	}
}
