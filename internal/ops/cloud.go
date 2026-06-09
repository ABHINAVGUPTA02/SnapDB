package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
	"github.com/ABHINAVGUPTA02/SnapDB/internal/metadata"
)

func UploadBackup(backup metadata.Backup, provider string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "local", "s3", "gcs", "azure":
	default:
		return "", fmt.Errorf("unsupported upload provider %q", provider)
	}

	dir, err := config.CloudDir(provider)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(dir, filepath.Base(backup.Path))
	if err := copyFile(backup.Path, dst, 0600); err != nil {
		return "", err
	}

	if provider == "local" {
		return dst, nil
	}
	return fmt.Sprintf("%s://snapdb-local/%s", provider, filepath.Base(dst)), nil
}
