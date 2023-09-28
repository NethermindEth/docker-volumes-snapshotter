package backup

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const tarBlockSize = 512

var ErrPrepareToAppend = errors.New("tar file is not prepared to append")

type TarFileWriter struct {
	file   *os.File
	writer *tar.Writer
	append bool
}

func NewTarFileWriter(path string, append bool) (*TarFileWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	tarFile := TarFileWriter{
		file:   file,
		append: append,
	}
	// Prepare for append if necessary
	if append {
		err := tarFile.prepareAppend()
		if err != nil {
			return nil, err
		}
	}
	// Build tar writer
	tarFile.writer = tar.NewWriter(file)
	return &tarFile, nil
}

func (t *TarFileWriter) AddDir(srcPath, prefix string) error {
	log.Printf("Adding dir: \"%s\" into \"%s\"...", srcPath, prefix)
	// walk through every file in the folder
	err := filepath.Walk(srcPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		fileRelPath, err := filepath.Rel(srcPath, file)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(prefix, fileRelPath)

		// write header
		if err := t.writer.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(t.writer, data); err != nil {
				return err
			}
			err = data.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (t *TarFileWriter) AddFile(srcPath, destPath string) error {
	log.Printf("Adding file: \"%s\" into \"%s\"...", srcPath, destPath)
	fi, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("source path is a directory")
	}

	// generate tar header
	header, err := tar.FileInfoHeader(fi, srcPath)
	if err != nil {
		return err
	}

	header.Name = destPath

	// write header
	if err := t.writer.WriteHeader(header); err != nil {
		return err
	}

	data, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(t.writer, data); err != nil {
		return err
	}
	return data.Close()
}

func (t *TarFileWriter) Close() error {
	err := t.writer.Close()
	if err != nil {
		return err
	}
	return t.file.Close()
}

func (t *TarFileWriter) prepareAppend() error {
	stats, err := t.file.Stat()
	if err != nil {
		return err
	}
	if stats.Size() < 2*tarBlockSize {
		return nil
	}

	// Check if the last 1024 bytes are all 0
	d := make([]byte, 1024)
	n, err := t.file.ReadAt(d, stats.Size()-1024)
	if err != nil {
		return err
	}
	if n != 1024 {
		return fmt.Errorf("%w: read %d bytes instead of 1024", ErrPrepareToAppend, n)
	}
	for _, b := range d {
		if b != 0 {
			return fmt.Errorf("%w: last 1024 bytes are not all 0", ErrPrepareToAppend)
		}
	}
	// Seek last 1024 bytes
	_, err = t.file.Seek(-1024, io.SeekEnd)
	return err
}

func GetVolumesData(tarPath string, volumesDataPath string) ([]volumeData, error) {
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
			var volumesData []volumeData
			err = yaml.Unmarshal(data, &volumesData)
			if err != nil {
				return nil, err
			}
			return volumesData, nil
		}
	}
}

func ExtractDir(tarPath, srcTarPath, fsPathTarget string) error {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if strings.HasPrefix(header.Name, srcTarPath) && header.Name != srcTarPath {
			// Build target path from header name
			relPath, err := filepath.Rel(srcTarPath, header.Name)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			targetPath := filepath.Join(fsPathTarget, relPath)

			// Restore item
			switch header.Typeflag {
			case tar.TypeDir:
				err := os.MkdirAll(targetPath, 0o755)
				if err != nil {
					return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
				}
			case tar.TypeReg:
				log.Printf("%-15s : %s", "extract file", targetPath)
				fileDir := filepath.Dir(targetPath)
				err := os.MkdirAll(fileDir, 0o755)
				if err != nil {
					return fmt.Errorf("failed to create directory %s: %w", fileDir, err)
				}
				f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode().Perm())
				if err != nil {
					return err
				}
				n, err := io.Copy(f, tarReader)
				if err != nil {
					return fmt.Errorf("failed to copy file %s: %w", targetPath, err)
				}
				if n != header.Size {
					return fmt.Errorf("failed to copy file %s: copied %d bytes instead of %d", targetPath, n, header.Size)
				}
			default:
				return fmt.Errorf("unexpected typeflag %d for %s", header.Typeflag, header.Name)
			}
		}
	}
}

func ExtractFile(tarPath, srcTarPath, fsPathTarget string) error {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)
	for header, err := tarReader.Next(); err != io.EOF; header, err = tarReader.Next() {
		if err != nil {
			return err
		}
		if header.Name == srcTarPath {
			log.Printf("%-15s : %s", "extract file", fsPathTarget)
			fileDir := filepath.Dir(fsPathTarget)
			err := os.MkdirAll(fileDir, 0o755)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fileDir, err)
			}
			f, err := os.OpenFile(fsPathTarget, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode().Perm())
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", fsPathTarget, err)
			}
			n, err := io.Copy(f, tarReader)
			if err != nil {
				return fmt.Errorf("failed to copy file %s: %w", fsPathTarget, err)
			}
			if n != header.Size {
				return fmt.Errorf("failed to copy file %s: copied %d bytes instead of %d", fsPathTarget, n, header.Size)
			}
		}
	}
	return nil
}
