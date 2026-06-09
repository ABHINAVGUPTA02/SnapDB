package ops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
	"github.com/ABHINAVGUPTA02/SnapDB/internal/metadata"
)

type BackupOptions struct {
	Target               DatabaseTarget
	OutputDir            string
	Compress             bool
	Compression          string
	Encrypt              bool
	EncryptionPassphrase string
	UploadProvider       string
	Kind                 string
}

type RestoreOptions struct {
	Target               DatabaseTarget
	Source               string
	EncryptionPassphrase string
	Rollback             bool
}

func Backup(ctx context.Context, opts BackupOptions) (*metadata.Backup, error) {
	if err := validateTarget(opts.Target); err != nil {
		return nil, err
	}
	if err := config.EnsureDirs(); err != nil {
		return nil, err
	}

	createdAt := time.Now().UTC()
	id := metadata.NewID(opts.Target.Type, createdAt)
	outputDir := opts.OutputDir
	if outputDir == "" {
		defaultDir, err := config.BackupsDir()
		if err != nil {
			return nil, err
		}
		outputDir = defaultDir
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return nil, err
	}

	ext := extensionFor(opts.Target.Type)
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s%s", opts.Target.Database, id, ext))
	if err := runDump(ctx, opts.Target, outputPath); err != nil {
		return nil, err
	}

	finalPath := outputPath
	if opts.Compress {
		format := opts.Compression
		if format == "" {
			format = "gzip"
		}
		compressedPath, err := compressFile(finalPath, format)
		if err != nil {
			return nil, err
		}
		finalPath = compressedPath
		opts.Compression = format
	}
	if opts.Encrypt {
		encryptedPath, err := encryptFile(finalPath, opts.EncryptionPassphrase)
		if err != nil {
			return nil, err
		}
		finalPath = encryptedPath
	}

	checksum, err := metadata.ChecksumFile(finalPath)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(finalPath)
	if err != nil {
		return nil, err
	}

	record := metadata.Backup{
		ID:             id,
		DatabaseType:   opts.Target.Type,
		DatabaseName:   opts.Target.Database,
		Profile:        opts.Target.Profile,
		Path:           finalPath,
		OriginalPath:   outputPath,
		CreatedAt:      createdAt,
		SizeBytes:      stat.Size(),
		ChecksumSHA256: checksum,
		Compressed:     opts.Compress,
		Compression:    opts.Compression,
		Encrypted:      opts.Encrypt,
		Status:         "completed",
		Kind:           opts.Kind,
	}

	if opts.UploadProvider != "" {
		uploaded, err := UploadBackup(record, opts.UploadProvider)
		if err != nil {
			return nil, err
		}
		record.UploadedTo = append(record.UploadedTo, uploaded)
	}

	store, err := metadata.Load()
	if err != nil {
		return nil, err
	}
	store.Add(record)
	if err := metadata.Save(store); err != nil {
		return nil, err
	}

	return &record, nil
}

func Restore(ctx context.Context, opts RestoreOptions) error {
	if err := validateTarget(opts.Target); err != nil {
		return err
	}

	source := opts.Source
	cleanup := make([]string, 0)
	defer func() {
		for _, path := range cleanup {
			_ = os.Remove(path)
		}
	}()

	if strings.HasSuffix(source, ".enc") {
		decrypted, err := decryptFile(source, opts.EncryptionPassphrase)
		if err != nil {
			return err
		}
		source = decrypted
		cleanup = append(cleanup, decrypted)
	}
	decompressed, err := decompressFile(source)
	if err != nil {
		return err
	}
	if decompressed != source {
		source = decompressed
		cleanup = append(cleanup, decompressed)
	}

	return runRestore(ctx, opts.Target, source)
}

func validateTarget(target DatabaseTarget) error {
	if target.Type == "" {
		return errors.New("database type is required")
	}
	if target.Host == "" {
		return errors.New("database host is required")
	}
	if target.Port == "" {
		return errors.New("database port is required")
	}
	if target.User == "" && target.Type != "redis" {
		return errors.New("database user is required")
	}
	if target.Database == "" && target.Type != "redis" {
		return errors.New("database name is required")
	}
	return nil
}

func extensionFor(databaseType string) string {
	switch databaseType {
	case "mongodb", "mongo":
		return ".archive"
	case "redis":
		return ".rdb"
	default:
		return ".sql"
	}
}

func runDump(ctx context.Context, target DatabaseTarget, outputPath string) error {
	switch target.Type {
	case "mysql", "mariadb":
		return runCommandToFile(ctx, outputPath, []string{"MYSQL_PWD=" + target.Password}, "mysqldump", "-h", target.Host, "-P", target.Port, "-u", target.User, target.Database)
	case "postgres", "postgresql":
		return runCommandToFile(ctx, outputPath, []string{"PGPASSWORD=" + target.Password}, "pg_dump", "-h", target.Host, "-p", target.Port, "-U", target.User, "-d", target.Database)
	case "mongodb", "mongo":
		return runCommandToFile(ctx, outputPath, nil, "mongodump", "--host", target.Host, "--port", target.Port, "--username", target.User, "--password", target.Password, "--db", target.Database, "--archive")
	case "redis":
		return runCommandToFile(ctx, outputPath, nil, "redis-cli", "-h", target.Host, "-p", target.Port, "--rdb", "-")
	default:
		return fmt.Errorf("unsupported database type %q", target.Type)
	}
}

func runRestore(ctx context.Context, target DatabaseTarget, source string) error {
	switch target.Type {
	case "mysql", "mariadb":
		return runCommandFromFile(ctx, source, []string{"MYSQL_PWD=" + target.Password}, "mysql", "-h", target.Host, "-P", target.Port, "-u", target.User, target.Database)
	case "postgres", "postgresql":
		return runCommandFromFile(ctx, source, []string{"PGPASSWORD=" + target.Password}, "psql", "-h", target.Host, "-p", target.Port, "-U", target.User, "-d", target.Database)
	case "mongodb", "mongo":
		return runCommand(ctx, nil, "mongorestore", "--host", target.Host, "--port", target.Port, "--username", target.User, "--password", target.Password, "--db", target.Database, "--archive="+source, "--drop")
	case "redis":
		return fmt.Errorf("redis restore requires replacing the Redis RDB file manually; verified source: %s", source)
	default:
		return fmt.Errorf("unsupported database type %q", target.Type)
	}
}

func runCommandToFile(ctx context.Context, outputPath string, extraEnv []string, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but was not found in PATH", name)
	}

	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd.Run()
}

func runCommandFromFile(ctx context.Context, inputPath string, extraEnv []string, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but was not found in PATH", name)
	}

	in, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer in.Close()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = in
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd.Run()
}

func runCommand(ctx context.Context, extraEnv []string, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but was not found in PATH", name)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd.Run()
}
