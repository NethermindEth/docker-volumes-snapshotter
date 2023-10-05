package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/backuptar"
	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"gopkg.in/yaml.v2"
)

func volumeId(target string) string {
	hash := sha256.Sum256([]byte(target))
	return hex.EncodeToString(hash[:])
}

func Backup(c *config.Config) error {
	slog.Info("Starting backup")
	backupWriter, err := backuptar.NewBackupWriter(backuptar.Path)
	if err != nil {
		return err
	}
	defer backupWriter.Close()

	var volumesData []VolumeData
	for _, v := range c.Volumes {
		targetInfo, err := os.Stat(v)
		if err != nil {
			return err
		}
		volumeData := VolumeData{
			Id:     volumeId(v),
			Target: v,
		}
		if targetInfo.IsDir() {
			volumeData.Type = "dir"
			slog.Info("Adding dir to backup", "src", v, "dest", filepath.Join(c.Prefix, volumeData.Id))
			err := backupWriter.AddDir(v, filepath.Join(c.Prefix, volumeData.Id))
			if err != nil {
				return err
			}
		} else {
			volumeData.Type = "file"
			slog.Info("Adding file to backup", "src", v, "dest", filepath.Join(c.Prefix, volumeData.Id))
			err := backupWriter.AddFile(v, filepath.Join(c.Prefix, volumeData.Id))
			if err != nil {
				return err
			}
		}
		volumesData = append(volumesData, volumeData)
	}

	dataTemp, err := os.CreateTemp("/", "volumes-data-*.yml")
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(&volumesData)
	if err != nil {
		return err
	}
	_, err = dataTemp.Write(data)
	if err != nil {
		return err
	}
	err = dataTemp.Close()
	if err != nil {
		return err
	}
	return backupWriter.AddFile(dataTemp.Name(), VolumesDataPath(c))
}
