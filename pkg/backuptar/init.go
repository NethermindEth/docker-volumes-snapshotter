package backuptar

import (
	"errors"
	"os"
)

// TarBlockSize is the size of a tar block following the tar specification
// https://www.gnu.org/software/tar/manual/html_node/Standard.html
const TarBlockSize = 512

// InitBackupTar creates an empty tar file at the given path with the correct
// size fto be reade for append operations. If the file already exists, it is
// truncated and overwritten.
func InitBackupTar(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.Write(make([]byte, 2*TarBlockSize))
	if err != nil {
		return err
	}
	if n != 2*TarBlockSize {
		return errors.New("failed to write empty tar file")
	}
	return nil
}
