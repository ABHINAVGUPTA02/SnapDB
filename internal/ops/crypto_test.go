package ops

import (
	"os"
	"testing"
)

func TestEncryptDecryptFile(t *testing.T) {
	path := t.TempDir() + "/backup.sql"
	original := []byte("create table snapdb(id int);")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	encrypted, err := encryptFile(path, "correct horse battery staple")
	if err != nil {
		t.Fatalf("encryptFile() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original file still exists after encryption")
	}

	decrypted, err := decryptFile(encrypted, "correct horse battery staple")
	if err != nil {
		t.Fatalf("decryptFile() error = %v", err)
	}
	data, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatalf("read decrypted file: %v", err)
	}
	if string(data) != string(original) {
		t.Fatalf("decrypted data = %q, want %q", data, original)
	}
}
