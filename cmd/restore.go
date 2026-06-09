package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/internal/metadata"
	"github.com/ABHINAVGUPTA02/SnapDB/internal/ops"
	"github.com/spf13/cobra"
)

var restoreFlags struct {
	dbFlags
	encryptionPassEnv string
	rollback          bool
}

var restoreCmd = &cobra.Command{
	Use:   "restore <backup-id-or-file>",
	Short: "Restore a database backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		target, err := resolveTarget(restoreFlags.dbFlags, "")
		if err != nil {
			return err
		}

		if store, err := metadata.Load(); err == nil {
			if record, _ := store.Find(source); record != nil {
				source = record.Path
				if target.Type == "" {
					target.Type = record.DatabaseType
				}
				if target.Database == "" {
					target.Database = record.DatabaseName
				}
			}
		}

		passphrase := os.Getenv(restoreFlags.encryptionPassEnv)
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Hour)
		defer cancel()

		if err := ops.Restore(ctx, ops.RestoreOptions{
			Target:               target,
			Source:               source,
			EncryptionPassphrase: passphrase,
			Rollback:             restoreFlags.rollback,
		}); err != nil {
			return err
		}

		fmt.Println("Restore completed")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	addDatabaseFlags(restoreCmd, &restoreFlags.dbFlags)
	restoreCmd.Flags().StringVar(&restoreFlags.encryptionPassEnv, "encryption-passphrase-env", "SNAPDB_ENCRYPTION_PASSWORD", "Environment variable containing encryption passphrase")
	restoreCmd.Flags().BoolVar(&restoreFlags.rollback, "rollback", false, "Create a pre-restore rollback backup when supported")
}
