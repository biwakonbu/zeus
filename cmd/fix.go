package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/doctor"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "検出された問題を修復",
	RunE:  runFix,
}

func init() {
	rootCmd.AddCommand(fixCmd)
	fixCmd.Flags().Bool("dry-run", false, "実際に修復せず、何が行われるかを表示")
}

func runFix(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	d := doctor.New(".")
	result, err := d.Fix(dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println(color.YellowString("DRY RUN - No changes will be made"))
	}

	for _, fix := range result.Fixes {
		if fix.Executed {
			fmt.Printf("%s %s\n", color.GreenString("✓"), fix.Action)
		} else {
			fmt.Printf("%s %s (would be executed)\n", color.YellowString("○"), fix.Action)
		}
	}

	return nil
}
