package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"gopkg.in/yaml.v2"
)

type volumeData struct {
	Id     string `yaml:"id"`
	Type   string `yaml:"type"`
	Target string `yaml:"target"`
}

func Backup(c *config.Config) error {
	tarFile, err := NewTarFile(c.Out, true)
	if err != nil {
		return err
	}

	var volumesData []volumeData

	for _, v := range c.Volumes {
		targetInfo, err := os.Stat(v)
		if err != nil {
			return err
		}

		volumeData := volumeData{
			Id:     volumeId(v),
			Target: v,
		}

		if targetInfo.IsDir() {
			volumeData.Type = "dir"
			err := tarFile.AddDir(v, filepath.Join(c.Prefix, volumeData.Id))
			if err != nil {
				return err
			}
		} else {
			volumeData.Type = "file"
			err := tarFile.AddFile(v, filepath.Join(c.Prefix, volumeData.Id))
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
	err = tarFile.AddFile(dataTemp.Name(), filepath.Join(c.Prefix, "volumes-data.yml"))
	if err != nil {
		return err
	}
	return tarFile.Close()
}

func volumeId(target string) string {
	hash := sha256.Sum256([]byte(target))
	return hex.EncodeToString(hash[:])
}
