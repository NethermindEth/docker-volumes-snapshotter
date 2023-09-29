package backup

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"gopkg.in/yaml.v2"
)

const (
	VolumesDataFileName = "volumes-data.yml"
)

type VolumeData struct {
	Id     string `yaml:"id"`
	Type   string `yaml:"type"`
	Target string `yaml:"target"`
}

// VolumesDataPath returns the path the volumes data file in the tar archive.
func VolumesDataPath(c *config.Config) string {
	return filepath.Join(c.Prefix, VolumesDataFileName)
}

// GetVolumesData returns volumes data from volumesDataPath in the tar archive
// at tarPath. Volume data is stored at the root of the prefix path defined in
// the config file.
func GetVolumesData(tarPath string, volumesDataPath string) ([]VolumeData, error) {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil, nil
			}
			return nil, err
		}
		if header.Name == volumesDataPath {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			var volumesData []VolumeData
			err = yaml.Unmarshal(data, &volumesData)
			if err != nil {
				return nil, err
			}
			return volumesData, nil
		}
	}
}
