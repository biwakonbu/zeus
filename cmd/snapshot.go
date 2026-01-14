package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "スナップショット管理",
	Long:  `プロジェクト状態のスナップショットを管理します。`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [label]",
	Short: "スナップショットを作成",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSnapshotCreate,
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "スナップショット一覧を表示",
	RunE:  runSnapshotList,
}

var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <timestamp>",
	Short: "スナップショットから復元",
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotRestore,
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)

	snapshotListCmd.Flags().IntP("limit", "n", 10, "表示件数")
}

func runSnapshotCreate(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	label := ""
	if len(args) > 0 {
		label = args[0]
	}

	zeus := getZeus(cmd)
	snapshot, err := zeus.CreateSnapshot(ctx, label)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Snapshot created: %s\n", green("✓"), snapshot.Timestamp)
	if label != "" {
		fmt.Printf("  Label: %s\n", label)
	}

	return nil
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	limit, _ := cmd.Flags().GetInt("limit")

	zeus := getZeus(cmd)
	snapshots, err := zeus.GetHistory(ctx, limit)
	if err != nil {
		return err
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(cyan("Snapshots"))
	fmt.Println("═══════════════════════════════════════════════════════════")

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found.")
		return nil
	}

	for _, s := range snapshots {
		label := ""
		if s.Label != "" {
			label = fmt.Sprintf(" (%s)", s.Label)
		}
		fmt.Printf("  %s%s\n", s.Timestamp, label)
		fmt.Printf("    Health: %s | Tasks: %d | Completed: %d\n",
			s.State.Health, s.State.Summary.TotalTasks, s.State.Summary.Completed)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Total: %d snapshot(s)\n", len(snapshots))

	return nil
}

func runSnapshotRestore(cmd *cobra.Command, args []string) error {
	ctx := getContext(cmd)
	timestamp := args[0]

	zeus := getZeus(cmd)
	if err := zeus.RestoreSnapshot(ctx, timestamp); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Restored from snapshot: %s\n", green("✓"), timestamp)

	return nil
}
