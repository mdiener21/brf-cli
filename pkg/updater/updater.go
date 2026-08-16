package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultAPIURL   = "https://api.github.com"
	maxMetadataSize = 1 << 20
	maxArchiveSize  = 200 << 20
)

type Result struct {
	Version        string
	AlreadyCurrent bool
}

type Updater struct {
	repository     string
	apiURL         string
	client         *http.Client
	goos           string
	goarch         string
	executablePath func() (string, error)
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func New(repository string) *Updater {
	return &Updater{
		repository: repository,
		apiURL:     defaultAPIURL,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		executablePath: os.Executable,
	}
}

func (u *Updater) Upgrade(ctx context.Context, currentVersion string) (Result, error) {
	latest, err := u.latestRelease(ctx)
	if err != nil {
		return Result{}, err
	}

	result := Result{Version: latest.TagName}
	if normalizeVersion(currentVersion) == normalizeVersion(latest.TagName) {
		result.AlreadyCurrent = true
		return result, nil
	}

	archiveName := releaseArchiveName(latest.TagName, u.goos, u.goarch)
	checksumName := fmt.Sprintf("brf_%s_checksums.txt", latest.TagName)
	archiveAsset, ok := findAsset(latest.Assets, archiveName)
	if !ok {
		return Result{}, fmt.Errorf("release %s has no asset for %s/%s (expected %s)", latest.TagName, u.goos, u.goarch, archiveName)
	}
	checksumAsset, ok := findAsset(latest.Assets, checksumName)
	if !ok {
		return Result{}, fmt.Errorf("release %s has no checksum file %s", latest.TagName, checksumName)
	}

	tempDir, err := os.MkdirTemp("", "brf-upgrade-*")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	checksums, err := u.downloadText(ctx, checksumAsset.URL)
	if err != nil {
		return Result{}, fmt.Errorf("download checksums: %w", err)
	}
	expectedChecksum, err := checksumFor(checksums, archiveName)
	if err != nil {
		return Result{}, err
	}

	archivePath := filepath.Join(tempDir, archiveName)
	actualChecksum, err := u.downloadFile(ctx, archiveAsset.URL, archivePath)
	if err != nil {
		return Result{}, fmt.Errorf("download %s: %w", archiveName, err)
	}
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return Result{}, fmt.Errorf("checksum mismatch for %s", archiveName)
	}

	binaryPath := filepath.Join(tempDir, executableName(u.goos))
	if err := extractBinary(archivePath, binaryPath, u.goos); err != nil {
		return Result{}, fmt.Errorf("extract %s: %w", archiveName, err)
	}

	targetPath, err := u.executablePath()
	if err != nil {
		return Result{}, fmt.Errorf("locate current executable: %w", err)
	}
	targetPath, err = filepath.EvalSymlinks(targetPath)
	if err != nil {
		return Result{}, fmt.Errorf("resolve current executable: %w", err)
	}
	if err := installExecutable(binaryPath, targetPath); err != nil {
		return Result{}, fmt.Errorf("replace %s: %w", targetPath, err)
	}

	return result, nil
}

func (u *Updater) latestRelease(ctx context.Context) (release, error) {
	parts := strings.Split(u.repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return release{}, fmt.Errorf("invalid GitHub repository %q", u.repository)
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(u.apiURL, "/"), parts[0], parts[1])
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return release{}, fmt.Errorf("create release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "brf-cli-updater")

	response, err := u.client.Do(req)
	if err != nil {
		return release{}, fmt.Errorf("query latest GitHub release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("query latest GitHub release: GitHub returned %s", response.Status)
	}

	var latest release
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxMetadataSize))
	if err := decoder.Decode(&latest); err != nil {
		return release{}, fmt.Errorf("decode latest GitHub release: %w", err)
	}
	if latest.TagName == "" {
		return release{}, fmt.Errorf("latest GitHub release has no tag")
	}
	return latest, nil
}

func (u *Updater) downloadText(ctx context.Context, url string) (string, error) {
	response, err := u.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, maxMetadataSize+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxMetadataSize {
		return "", fmt.Errorf("response exceeds %d bytes", maxMetadataSize)
	}
	return string(data), nil
}

func (u *Updater) downloadFile(ctx context.Context, url, destination string) (string, error) {
	response, err := u.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, maxArchiveSize+1))
	closeErr := file.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if written > maxArchiveSize {
		return "", fmt.Errorf("archive exceeds %d bytes", maxArchiveSize)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (u *Updater) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "brf-cli-updater")

	response, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("server returned %s", response.Status)
	}
	return response, nil
}

func checksumFor(contents, filename string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || strings.TrimPrefix(fields[1], "*") != filename {
			continue
		}
		if len(fields[0]) != sha256.Size*2 {
			return "", fmt.Errorf("invalid checksum for %s", filename)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return "", fmt.Errorf("invalid checksum for %s", filename)
		}
		return fields[0], nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	return "", fmt.Errorf("checksum file has no entry for %s", filename)
}

func findAsset(assets []asset, name string) (asset, bool) {
	for _, candidate := range assets {
		if candidate.Name == name {
			return candidate, true
		}
	}
	return asset{}, false
}

func releaseArchiveName(tag, goos, goarch string) string {
	extension := ".tar.gz"
	if goos == "windows" {
		extension = ".zip"
	}
	return fmt.Sprintf("brf_%s_%s_%s%s", tag, goos, goarch, extension)
}

func executableName(goos string) string {
	if goos == "windows" {
		return "brf.exe"
	}
	return "brf"
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
