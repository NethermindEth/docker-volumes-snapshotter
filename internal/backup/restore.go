package backup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
)

func Restore(c *config.Config) error {
	// Get volumes data
	volumesData, err := GetVolumesData(c.BackupFile, filepath.Join(c.Prefix, volumesDataFileName))
	if err != nil {
		return err
	}
	for _, v := range volumesData {
		// Check target is absolute path
		if !filepath.IsAbs(v.Target) {
			return fmt.Errorf("target of volume %s is not absolute path", v.Id)
		}
		switch v.Type {
		case "dir":
			log.Printf("%-15s : %s", "RESTORING DIR", v.Target)
			// Clear directory
			err := clearDirectory(v.Target)
			if err != nil {
				return err
			}
			// Replace directory with backup data
			err = ExtractDir(c.BackupFile, filepath.Join(c.Prefix, v.Id), v.Target)
			if err != nil {
				return err
			}
		case "file":
			log.Printf("%-15s : %s", "RESTORING FILE", v.Target)
			// Replace file with backup data
			err := ExtractFile(c.BackupFile, filepath.Join(c.Prefix, v.Id), v.Target)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown volume type %s for volume %s", v.Type, v.Id)
		}
	}
	return nil
}

func clearDirectory(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty directory
			return os.MkdirAll(path, 0o755)
		}
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("path %s exists, but is not a directory", path)
	}
	// Remove directory content
	dirItems, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, item := range dirItems {
		itemPath := filepath.Join(path, item.Name())
		log.Printf("%-15s : %s", "removing", itemPath)
		err := os.RemoveAll(itemPath)
		if err != nil {
			return err
		}
	}
	return nil
}
