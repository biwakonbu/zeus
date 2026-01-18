package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/biwakonbu/zeus/internal/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "システムの健全性を診断",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	zeus := getZeus(cmd)

	// IntegrityChecker を作成（ハンドラーがある場合）
	var d *doctor.Doctor
	if zeus != nil {
		checker := createIntegrityChecker(zeus)
		d = doctor.NewWithIntegrity(".", checker)
	} else {
		d = doctor.New(".")
	}

	result, err := d.Diagnose(ctx)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(cyan("Zeus Doctor - System Diagnosis"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	for _, check := range result.Checks {
		var icon string
		switch check.Status {
		case "pass":
			icon = color.GreenString("✓")
		case "warn":
			icon = color.YellowString("⚠")
		case "fail":
			icon = color.RedString("✗")
		}
		fmt.Printf("%s %s: %s\n", icon, check.Check, check.Message)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Overall: %s\n", result.Overall)

	if result.FixableCount > 0 {
		fmt.Printf("\n%d issue(s) can be fixed automatically. Run 'zeus fix' to repair.\n", result.FixableCount)
	}

	return nil
}

// createIntegrityChecker は Zeus インスタンスから IntegrityChecker を作成
func createIntegrityChecker(zeus *core.Zeus) *core.IntegrityChecker {
	registry := zeus.GetRegistry()
	if registry == nil {
		return nil
	}

	// Objective ハンドラーを取得
	objHandler, ok := registry.Get("objective")
	if !ok {
		return nil
	}
	objH, ok := objHandler.(*core.ObjectiveHandler)
	if !ok {
		return nil
	}

	// Deliverable ハンドラーを取得
	delHandler, ok := registry.Get("deliverable")
	if !ok {
		return nil
	}
	delH, ok := delHandler.(*core.DeliverableHandler)
	if !ok {
		return nil
	}

	return core.NewIntegrityChecker(objH, delH)
}
