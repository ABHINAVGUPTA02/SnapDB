package ops

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func compressFile(path, format string) (string, error) {
	switch format {
	case "", "gzip", "gz":
		return compressGzip(path)
	case "zstd", "zst":
		return compressZstd(path)
	case "zip":
		return compressZip(path)
	default:
		return "", fmt.Errorf("unsupported compression format %q", format)
	}
}

func compressGzip(path string) (string, error) {
	outPath := path + ".gz"
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	if _, err := io.Copy(gz, in); err != nil {
		gz.Close()
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return outPath, os.Remove(path)
}

func compressZstd(path string) (string, error) {
	outPath := path + ".zst"
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	encoder, err := zstd.NewWriter(out)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(encoder, in); err != nil {
		encoder.Close()
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return outPath, os.Remove(path)
}

func compressZip(path string) (string, error) {
	outPath := path + ".zip"
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	archive := zip.NewWriter(out)
	writer, err := archive.Create(filepath.Base(path))
	if err != nil {
		archive.Close()
		return "", err
	}
	if _, err := io.Copy(writer, in); err != nil {
		archive.Close()
		return "", err
	}
	if err := archive.Close(); err != nil {
		return "", err
	}
	return outPath, os.Remove(path)
}

func decompressFile(path string) (string, error) {
	switch {
	case strings.HasSuffix(path, ".gz"):
		return decompressGzip(path)
	case strings.HasSuffix(path, ".zst"):
		return decompressZstd(path)
	case strings.HasSuffix(path, ".zip"):
		return decompressZip(path)
	default:
		return path, nil
	}
}

func decompressGzip(path string) (string, error) {
	outPath := strings.TrimSuffix(path, ".gz")
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	reader, err := gzip.NewReader(in)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", err
	}
	return outPath, nil
}

func decompressZstd(path string) (string, error) {
	outPath := strings.TrimSuffix(path, ".zst")
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	decoder, err := zstd.NewReader(in)
	if err != nil {
		return "", err
	}
	defer decoder.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, decoder); err != nil {
		return "", err
	}
	return outPath, nil
}

func decompressZip(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	if len(reader.File) != 1 {
		return "", fmt.Errorf("zip backup must contain exactly one file")
	}

	archiveFile := reader.File[0]
	if archiveFile.FileInfo().IsDir() {
		return "", fmt.Errorf("zip backup entry is a directory")
	}
	in, err := archiveFile.Open()
	if err != nil {
		return "", err
	}
	defer in.Close()

	outPath := filepath.Join(filepath.Dir(path), filepath.Base(archiveFile.Name))
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return outPath, nil
}
