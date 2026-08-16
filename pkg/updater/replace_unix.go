//go:build !windows

package updater

import "os"

func replaceExecutable(replacementPath, targetPath string) error {
	return os.Rename(replacementPath, targetPath)
}
