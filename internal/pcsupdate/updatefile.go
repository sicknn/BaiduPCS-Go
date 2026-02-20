package pcsupdate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func update(targetPath string, src io.Reader) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		fmt.Printf("Warning: %s\n", err)
		return nil
	}

	privMode := info.Mode()

	oldPath := filepath.Join(filepath.Dir(targetPath), "old"+filepath.Base(targetPath))

	err = os.Rename(targetPath, oldPath)
	if err != nil {
		return err
	}

	newFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, privMode)
	if err != nil {
		return err
	}

	_, err = io.Copy(newFile, src)
	if err != nil {
		return err
	}

	err = newFile.Close()
	if err != nil {
		fmt.Printf("Warning: failed to close file: %s\n", err)
	}

	err = os.Remove(oldPath)
	if err != nil {
		fmt.Printf("Warning: failed to remove old file: %s\n", err)
	}
	return nil
}
