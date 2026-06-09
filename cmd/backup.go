package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/internal/ops"
	"github.com/spf13/cobra"
)

var backupFlags struct {
	dbFlags
	outputDir         string
	compress          bool
	compression       string
	encrypt           bool
	encryptionPassEnv string
	upload            string
	incremental       bool
	differential      bool
}

var backupCmd = &cobra.Command{
	Use:   "backup [mysql|postgres|mongodb|redis]",
	Short: "Create a database backup",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		positionalType := ""
		if len(args) == 1 {
			positionalType = args[0]
		}
		target, err := resolveTarget(backupFlags.dbFlags, positionalType)
		if err != nil {
			return err
		}

		kind := "full"
		if backupFlags.incremental {
			kind = "incremental"
		}
		if backupFlags.differential {
			kind = "differential"
		}

		passphrase := ""
		if backupFlags.encrypt {
			passphrase = os.Getenv(backupFlags.encryptionPassEnv)
			if passphrase == "" {
				return fmt.Errorf("set --encryption-passphrase-env to an environment variable containing the encryption passphrase")
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Hour)
		defer cancel()

		record, err := ops.Backup(ctx, ops.BackupOptions{
			Target:               target,
			OutputDir:            backupFlags.outputDir,
			Compress:             backupFlags.compress,
			Compression:          backupFlags.compression,
			Encrypt:              backupFlags.encrypt,
			EncryptionPassphrase: passphrase,
			UploadProvider:       backupFlags.upload,
			Kind:                 kind,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Backup successful\nID: %s\nPath: %s\nChecksum: %s\n", record.ID, record.Path, record.ChecksumSHA256)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	addDatabaseFlags(backupCmd, &backupFlags.dbFlags)
	backupCmd.Flags().StringVar(&backupFlags.outputDir, "output-dir", "", "Directory for backup files")
	backupCmd.Flags().BoolVar(&backupFlags.compress, "compress", false, "Compress the backup")
	backupCmd.Flags().StringVar(&backupFlags.compression, "format", "gzip", "Compression format: gzip, zstd, zip")
	backupCmd.Flags().BoolVar(&backupFlags.encrypt, "encrypt", false, "Encrypt the backup with AES-256-GCM")
	backupCmd.Flags().StringVar(&backupFlags.encryptionPassEnv, "encryption-passphrase-env", "SNAPDB_ENCRYPTION_PASSWORD", "Environment variable containing encryption passphrase")
	backupCmd.Flags().StringVar(&backupFlags.upload, "upload", "", "Upload provider: local, s3, gcs, azure")
	backupCmd.Flags().BoolVar(&backupFlags.incremental, "incremental", false, "Mark this backup as incremental")
	backupCmd.Flags().BoolVar(&backupFlags.differential, "differential", false, "Mark this backup as differential")
}
