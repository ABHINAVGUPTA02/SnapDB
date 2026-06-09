package metadata

import (
	"os"
	"testing"
	"time"
)

func TestStoreAddFindDelete(t *testing.T) {
	store := &Store{}
	backup := Backup{
		ID:           "mysql-20260609T020000",
		DatabaseType: "mysql",
		DatabaseName: "appdb",
		Path:         "/tmp/appdb.sql",
		CreatedAt:    time.Now(),
		Status:       "completed",
	}

	store.Add(backup)
	found, _ := store.Find(backup.ID)
	if found == nil {
		t.Fatalf("backup was not found")
	}

	deleted, err := store.Delete(backup.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if deleted.ID != backup.ID {
		t.Fatalf("deleted ID = %q, want %q", deleted.ID, backup.ID)
	}
	if len(store.Backups) != 0 {
		t.Fatalf("store still has backups")
	}
}

func TestChecksumFile(t *testing.T) {
	path := t.TempDir() + "/backup.sql"
	if err := os.WriteFile(path, []byte("snapdb"), 0600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	checksum, err := ChecksumFile(path)
	if err != nil {
		t.Fatalf("ChecksumFile() error = %v", err)
	}
	if len(checksum) != 64 {
		t.Fatalf("checksum length = %d, want 64", len(checksum))
	}
}
