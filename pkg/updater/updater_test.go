package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestUpgradeDownloadsVerifiesAndReplacesExecutable(t *testing.T) {
	archive := tarArchive(t, "brf_v0.6.0_linux_amd64/brf", []byte("new binary"))
	archiveName := "brf_v0.6.0_linux_amd64.tar.gz"
	httpClient, requests := releaseClient(t, "v0.6.0", archiveName, archive, checksum(archive))

	target := filepath.Join(t.TempDir(), "brf")
	if err := os.WriteFile(target, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}

	client := New("owner/repo")
	client.apiURL = "https://test.local"
	client.client = httpClient
	client.goos = "linux"
	client.goarch = "amd64"
	client.executablePath = func() (string, error) { return target, nil }

	result, err := client.Upgrade(context.Background(), "0.5.0")
	if err != nil {
		t.Fatalf("expected upgrade to succeed, got %v", err)
	}
	if result.Version != "v0.6.0" || result.AlreadyCurrent {
		t.Fatalf("unexpected result: %#v", result)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new binary" {
		t.Fatalf("expected executable to be replaced, got %q", contents)
	}
	if *requests != 3 {
		t.Fatalf("expected release, checksum, and archive requests; got %d", *requests)
	}
}

func TestUpgradeSkipsDownloadWhenAlreadyCurrent(t *testing.T) {
	httpClient, requests := releaseClient(t, "v0.5.0", "unused", nil, "")

	client := New("owner/repo")
	client.apiURL = "https://test.local"
	client.client = httpClient

	result, err := client.Upgrade(context.Background(), "0.5.0")
	if err != nil {
		t.Fatalf("expected version check to succeed, got %v", err)
	}
	if !result.AlreadyCurrent {
		t.Fatalf("expected current result, got %#v", result)
	}
	if *requests != 1 {
		t.Fatalf("expected only the release request, got %d", *requests)
	}
}

func TestUpgradeRejectsChecksumMismatch(t *testing.T) {
	archive := tarArchive(t, "brf_v0.6.0_linux_amd64/brf", []byte("new binary"))
	archiveName := "brf_v0.6.0_linux_amd64.tar.gz"
	httpClient, _ := releaseClient(t, "v0.6.0", archiveName, archive, fmt.Sprintf("%064d", 0))

	target := filepath.Join(t.TempDir(), "brf")
	if err := os.WriteFile(target, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}

	client := New("owner/repo")
	client.apiURL = "https://test.local"
	client.client = httpClient
	client.goos = "linux"
	client.goarch = "amd64"
	client.executablePath = func() (string, error) { return target, nil }

	if _, err := client.Upgrade(context.Background(), "0.5.0"); err == nil {
		t.Fatal("expected checksum mismatch")
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old binary" {
		t.Fatalf("expected old executable to remain, got %q", contents)
	}
}

func TestExtractWindowsZipBinary(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "brf.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("brf_v0.6.0_windows_amd64/brf.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("windows binary")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(t.TempDir(), "brf.exe")
	if err := extractBinary(archivePath, destination, "windows"); err != nil {
		t.Fatalf("expected zip extraction to succeed, got %v", err)
	}
	contents, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "windows binary" {
		t.Fatalf("unexpected extracted binary: %q", contents)
	}
}

func releaseClient(t *testing.T, tag, archiveName string, archive []byte, archiveChecksum string) (*http.Client, *int) {
	t.Helper()
	requests := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		var body []byte
		status := http.StatusOK
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			checksumName := fmt.Sprintf("brf_%s_checksums.txt", tag)
			body, _ = json.Marshal(release{
				TagName: tag,
				Assets: []asset{
					{Name: archiveName, URL: "https://test.local/archive"},
					{Name: checksumName, URL: "https://test.local/checksums"},
				},
			})
		case "/archive":
			body = archive
		case "/checksums":
			body = []byte(fmt.Sprintf("%s  %s\n", archiveChecksum, archiveName))
		default:
			status = http.StatusNotFound
		}
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})
	return &http.Client{Transport: transport}, &requests
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func tarArchive(t *testing.T, name string, contents []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func checksum(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
