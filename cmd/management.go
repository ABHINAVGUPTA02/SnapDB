package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/internal/metadata"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List known backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := metadata.Load()
		if err != nil {
			return err
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(writer, "ID\tDATABASE\tTYPE\tCREATED\tSIZE\tSTATUS")
		for _, backup := range store.Backups {
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
				backup.ID,
				backup.DatabaseName,
				backup.DatabaseType,
				backup.CreatedAt.Format("2006-01-02 15:04"),
				formatBytes(backup.SizeBytes),
				backup.Status,
			)
		}
		return writer.Flush()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <backup-id>",
	Short: "Delete a backup and its metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := metadata.Load()
		if err != nil {
			return err
		}
		backup, err := store.Delete(args[0])
		if err != nil {
			return err
		}
		if err := os.Remove(backup.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := metadata.Save(store); err != nil {
			return err
		}
		fmt.Printf("Deleted backup %s\n", backup.ID)
		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show backup storage statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := metadata.Load()
		if err != nil {
			return err
		}
		var total int64
		for _, backup := range store.Backups {
			total += backup.SizeBytes
		}
		fmt.Printf("Total Backups: %d\nStorage Used: %s\n", len(store.Backups), formatBytes(total))
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify <backup-id-or-file>",
	Short: "Verify a backup checksum",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		expected := ""
		store, err := metadata.Load()
		if err != nil {
			return err
		}
		if backup, _ := store.Find(source); backup != nil {
			source = backup.Path
			expected = backup.ChecksumSHA256
		}
		actual, err := metadata.ChecksumFile(source)
		if err != nil {
			return err
		}
		if expected != "" && actual != expected {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
		}
		fmt.Printf("Backup verified\nPath: %s\nChecksum: %s\n", source, actual)
		return nil
	},
}

var cleanupOlderThan string

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete backups older than a duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := time.ParseDuration(cleanupOlderThan)
		if err != nil {
			return err
		}
		cutoff := time.Now().Add(-duration)
		store, err := metadata.Load()
		if err != nil {
			return err
		}

		kept := make([]metadata.Backup, 0, len(store.Backups))
		deleted := 0
		for _, backup := range store.Backups {
			if backup.CreatedAt.Before(cutoff) {
				if err := os.Remove(backup.Path); err != nil && !os.IsNotExist(err) {
					return err
				}
				deleted++
				continue
			}
			kept = append(kept, backup)
		}
		store.Backups = kept
		if err := metadata.Save(store); err != nil {
			return err
		}
		fmt.Printf("Deleted %d old backups\n", deleted)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd, deleteCmd, statsCmd, verifyCmd, cleanupCmd)
	cleanupCmd.Flags().StringVar(&cleanupOlderThan, "older-than", "720h", "Delete backups older than this duration")
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
