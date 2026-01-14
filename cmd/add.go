package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/biwakonbu/zeus/internal/core"
)

var addCmd = &cobra.Command{
	Use:   "add <entity> <name>",
	Short: "エンティティを追加",
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	entity := args[0]
	name := args[1]

	zeus := core.New(".")
	result, err := zeus.Add(entity, name)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Added %s: %s (ID: %s)\n", green("✓"), result.Entity, name, result.ID)

	return nil
}
