package backup

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const tarBlockSize = 512

var ErrPrepareToAppend = errors.New("tar file is not prepared to append")

type TarFile struct {
	file   *os.File
	writer *tar.Writer
	append bool
}

func NewTarFile(path string, append bool) (*TarFile, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	tarFile := TarFile{
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

func (t *TarFile) AddDir(srcPath, prefix string) error {
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

func (t *TarFile) AddFile(srcPath, destPath string) error {
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

func (t *TarFile) Close() error {
	err := t.writer.Close()
	if err != nil {
		return err
	}
	return t.file.Close()
}

func (t *TarFile) prepareAppend() error {
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
