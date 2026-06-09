package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/internal/scheduler"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage backup schedules",
}

var scheduleCreateFlags struct {
	every   string
	cron    string
	profile string
	command string
}

var scheduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup schedule definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		if scheduleCreateFlags.every == "" && scheduleCreateFlags.cron == "" {
			return fmt.Errorf("either --every or --cron is required")
		}
		if scheduleCreateFlags.every != "" {
			if _, err := time.ParseDuration(scheduleCreateFlags.every); err != nil {
				return err
			}
		}
		command := scheduleCreateFlags.command
		if command == "" {
			command = "snapdb backup"
			if scheduleCreateFlags.profile != "" {
				command += " --profile " + scheduleCreateFlags.profile
			}
		}
		if !strings.HasPrefix(command, "snapdb ") {
			return fmt.Errorf("scheduled command must start with snapdb")
		}

		store, err := scheduler.Load()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		schedule := scheduler.Schedule{
			ID:        scheduler.NewID(now),
			Profile:   scheduleCreateFlags.profile,
			Every:     scheduleCreateFlags.every,
			Cron:      scheduleCreateFlags.cron,
			CreatedAt: now,
			Command:   command,
		}
		store.Add(schedule)
		if err := scheduler.Save(store); err != nil {
			return err
		}
		fmt.Printf("Created schedule %s\n", schedule.ID)
		return nil
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := scheduler.Load()
		if err != nil {
			return err
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(writer, "ID\tEVERY\tCRON\tPROFILE\tCOMMAND")
		for _, schedule := range store.Schedules {
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", schedule.ID, schedule.Every, schedule.Cron, schedule.Profile, schedule.Command)
		}
		return writer.Flush()
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete <schedule-id>",
	Short: "Delete a backup schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := scheduler.Load()
		if err != nil {
			return err
		}
		deleted, err := store.Delete(args[0])
		if err != nil {
			return err
		}
		if err := scheduler.Save(store); err != nil {
			return err
		}
		fmt.Printf("Deleted schedule %s\n", deleted.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleCreateCmd, scheduleListCmd, scheduleDeleteCmd)

	scheduleCreateCmd.Flags().StringVar(&scheduleCreateFlags.every, "every", "", "Run every duration, such as 24h")
	scheduleCreateCmd.Flags().StringVar(&scheduleCreateFlags.cron, "cron", "", "Cron expression, such as \"0 2 * * *\"")
	scheduleCreateCmd.Flags().StringVar(&scheduleCreateFlags.profile, "profile", "", "Profile to use in the scheduled backup command")
	scheduleCreateCmd.Flags().StringVar(&scheduleCreateFlags.command, "command", "", "Full snapdb command to record")
}
