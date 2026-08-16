//go:build windows

package updater

import (
	"fmt"
	"os"
)

func replaceExecutable(replacementPath, targetPath string) error {
	backupPath := targetPath + ".old"
	_ = os.Remove(backupPath)

	if err := os.Rename(targetPath, backupPath); err != nil {
		return fmt.Errorf("move current executable aside: %w", err)
	}
	if err := os.Rename(replacementPath, targetPath); err != nil {
		if rollbackErr := os.Rename(backupPath, targetPath); rollbackErr != nil {
			return fmt.Errorf("install replacement: %v (rollback failed: %v)", err, rollbackErr)
		}
		return fmt.Errorf("install replacement: %w", err)
	}

	// Windows may keep the running executable locked. A later upgrade removes it.
	_ = os.Remove(backupPath)
	return nil
}
