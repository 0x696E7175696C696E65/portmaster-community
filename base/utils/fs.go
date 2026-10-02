package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"runtime"
)

// EnsureDirectory ensures that the given directory exists and that is has the given permissions set.
// If path is a file, it is deleted and a directory created.
func EnsureDirectory(path string, perm FSPermission) error {
	if !perm.IsExecPermission() {
		slog.Warn("utils: setting not executable permission for directory", "dir", path)
	}

	// open path
	f, err := os.Stat(path)
	if err == nil {
		// file exists
		if f.IsDir() {
			// directory exists, check permissions
			// Windows ACLs cannot be inferred from Mode().Perm(). Fail if the
			// requested policy cannot be installed instead of reporting success.
			if runtime.GOOS == "windows" || f.Mode().Perm() != perm.AsUnixPermission() {
				return SetFilePermission(path, perm)
			}
			return nil
		}
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("could not remove file %s to place dir: %w", path, err)
		}
	}
	// file does not exist (or has been deleted)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(path, perm.AsUnixPermission())
		if err != nil {
			return fmt.Errorf("could not create dir %s: %w", path, err)
		}
		// Set permissions.
		return SetFilePermission(path, perm)
	}
	// other error opening path
	return fmt.Errorf("failed to access %s: %w", path, err)
}

// PathExists returns whether the given path (file or dir) exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, fs.ErrExist)
}
