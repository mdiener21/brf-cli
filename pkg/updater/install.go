package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func installExecutable(sourcePath, targetPath string) error {
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	temp, err := os.CreateTemp(filepath.Dir(targetPath), ".brf-upgrade-*")
	if err != nil {
		return fmt.Errorf("create replacement beside executable: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := io.Copy(temp, source); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(targetInfo.Mode().Perm() | 0500); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	return replaceExecutable(tempPath, targetPath)
}
