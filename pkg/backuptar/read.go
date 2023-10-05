package backuptar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractDir extracts the directory srcTarPath from the tar archive at tarPath
// to the filesystem path fsPathTarget.
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

// ExtractFile extracts the file srcTarPath from the tar archive at tarPath to
// the filesystem path fsPathTarget.
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
			return nil
		}
	}
	return ErrFileNotFound
}
