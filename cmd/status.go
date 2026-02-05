package cmd

import (
	"context"
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
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(cyan("Zeus Project Status"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Project: %s\n", result.Project.Name)
	fmt.Printf("Health:  %s\n", formatHealth(result.State.Health))
	fmt.Println()

	// Vision 表示
	displayVisionSummary(ctx, zeus, green)

	// Objectives 表示
	displayObjectivesSummary(ctx, zeus, cyan)

	// Deliverables 表示
	displayDeliverablesSummary(ctx, zeus, cyan)

	// Tasks 表示
	fmt.Println("Tasks Summary:")
	fmt.Printf("  Total:       %d\n", result.State.Summary.TotalActivities)
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

// displayVisionSummary は Vision の概要を表示
func displayVisionSummary(ctx context.Context, zeus *core.Zeus, green func(...interface{}) string) {
	visionResult, err := zeus.List(ctx, "vision")
	if err != nil || visionResult.Total == 0 {
		fmt.Println("Vision: (未設定)")
		fmt.Println()
		return
	}

	// Vision を取得
	entity, err := zeus.Get(ctx, "vision", "vision-001")
	if err != nil {
		fmt.Println("Vision: (未設定)")
		fmt.Println()
		return
	}

	vision, ok := entity.(*core.Vision)
	if !ok {
		fmt.Println("Vision: (未設定)")
		fmt.Println()
		return
	}

	fmt.Printf("Vision: %s\n", green(vision.Title))
	if vision.Statement != "" {
		fmt.Printf("  \"%s\"\n", vision.Statement)
	}
	fmt.Println()
}

// displayObjectivesSummary は Objectives の概要を表示
func displayObjectivesSummary(ctx context.Context, zeus *core.Zeus, cyan func(...interface{}) string) {
	objResult, err := zeus.List(ctx, "objective")
	if err != nil {
		return
	}

	if objResult.Total == 0 {
		fmt.Println("Objectives: (なし)")
		fmt.Println()
		return
	}

	fmt.Printf("Objectives: %s\n", cyan(fmt.Sprintf("%d 件", objResult.Total)))
	fmt.Println()
}

// displayDeliverablesSummary は Deliverables の概要を表示
func displayDeliverablesSummary(ctx context.Context, zeus *core.Zeus, cyan func(...interface{}) string) {
	delResult, err := zeus.List(ctx, "deliverable")
	if err != nil {
		return
	}

	if delResult.Total == 0 {
		fmt.Println("Deliverables: (なし)")
		fmt.Println()
		return
	}

	fmt.Printf("Deliverables: %s\n", cyan(fmt.Sprintf("%d 件", delResult.Total)))
	fmt.Println()
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
