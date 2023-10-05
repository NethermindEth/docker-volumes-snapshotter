package backuptar

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// BackupWriter is a struct that write files into the backup tar file.
type BackupWriter struct {
	file      *os.File
	tarWriter *tar.Writer
}

// NewBackupWriter creates a new BackupWriter.
func NewBackupWriter(tarPath string) (*BackupWriter, error) {
	tarFile, err := os.OpenFile(tarPath, os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	stats, err := tarFile.Stat()
	if err != nil {
		return nil, err
	}
	if stats.Size() < 2*TarBlockSize {
		return nil, fmt.Errorf("%w: tar file size is less than 2 blocks", ErrPrepareToAppend)
	}

	// Check if the last 1024 bytes are all 0
	d := make([]byte, 1024)
	n, err := tarFile.ReadAt(d, stats.Size()-1024)
	if err != nil {
		return nil, err
	}
	if n != 1024 {
		return nil, fmt.Errorf("%w: read %d bytes instead of 1024", ErrPrepareToAppend, n)
	}
	for _, b := range d {
		if b != 0 {
			return nil, fmt.Errorf("%w: last 1024 bytes are not all 0", ErrPrepareToAppend)
		}
	}
	// Seek last 1024 bytes
	_, err = tarFile.Seek(-1024, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	// Truncate last 1024 bytes
	err = tarFile.Truncate(stats.Size() - 1024)
	if err != nil {
		return nil, err
	}
	return &BackupWriter{
		file:      tarFile,
		tarWriter: tar.NewWriter(tarFile),
	}, nil
}

// AddDir adds a directory into the backup tar file.
func (b *BackupWriter) AddDir(src, dest string) error {
	// walk through every file in the folder
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		fileRelPath, err := filepath.Rel(src, file)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(dest, fileRelPath)

		// write header
		if err := b.tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(b.tarWriter, data); err != nil {
				return err
			}
			err = data.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// AddFile adds a file into the backup tar file.
func (b *BackupWriter) AddFile(src, dest string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("source path is a directory")
	}

	// generate tar header
	header, err := tar.FileInfoHeader(fi, src)
	if err != nil {
		return err
	}

	header.Name = dest

	// write header
	if err := b.tarWriter.WriteHeader(header); err != nil {
		return err
	}

	data, err := os.Open(src)
	if err != nil {
		return err
	}
	if _, err := io.Copy(b.tarWriter, data); err != nil {
		return err
	}
	return data.Close()
}

// Close closes the backup tar file.
func (w *BackupWriter) Close() error {
	err := w.tarWriter.Close()
	if err != nil {
		return err
	}
	return w.file.Close()
}
