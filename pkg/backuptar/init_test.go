package backuptar

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitBackupTar(t *testing.T) {
	testDir := t.TempDir()
	tarPath := filepath.Join(testDir, "test.tar")

	// Initialize the backup tar file
	err := InitBackupTar(tarPath)
	require.NoError(t, err, "Failed to initialize backup tar file")

	// Verify that the file was created with the correct size
	tarFile, err := os.Open(tarPath)
	require.NoError(t, err, "Failed to open file")
	fileInfo, err := os.Stat(tarPath)
	require.NoError(t, err, "Failed to stat file")
	require.Equal(t, int64(2*TarBlockSize), fileInfo.Size(), "File size is incorrect")

	// Verify that the file can be read
	_, err = tarFile.Seek(0, 0)
	require.NoError(t, err, "Failed to seek to beginning of file")
	data, err := io.ReadAll(tarFile)
	require.NoError(t, err, "Failed to read file")
	require.Len(t, data, 2*TarBlockSize, "File data is incorrect length")
	assert.Equal(t, make([]byte, 2*TarBlockSize), data, "File data is incorrect")
}

func TestInitBackupTar_AlreadyExists(t *testing.T) {
	// Create tar file
	tarFile, err := os.CreateTemp(t.TempDir(), "test.tar")
	require.NoError(t, err, "Failed to open file")
	defer tarFile.Close()
	defer os.Remove(tarFile.Name())

	// Initialize the backup tar file
	err = InitBackupTar(tarFile.Name())
	require.NoError(t, err, "Failed to initialize backup tar file")

	// Verify that the file was created with the correct size
	fileInfo, err := os.Stat(tarFile.Name())
	require.NoError(t, err, "Failed to stat file")
	require.Equal(t, int64(2*TarBlockSize), fileInfo.Size(), "File size is incorrect")

	_, err = tarFile.Seek(0, 0)
	require.NoError(t, err, "Failed to seek to beginning of file")

	data, err := io.ReadAll(tarFile)
	require.NoError(t, err, "Failed to read file")
	require.Len(t, data, 2*TarBlockSize, "File data is incorrect length")
	assert.Equal(t, make([]byte, 2*TarBlockSize), data, "File data is incorrect")
}
