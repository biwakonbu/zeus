package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

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
	d := doctor.New(".")
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
