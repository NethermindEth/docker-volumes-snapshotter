package backuptar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackupWriter(t *testing.T) {
	tc := []struct {
		name    string
		tarPath string
		wantErr bool
	}{
		{
			name: "tar file does not exist",
			tarPath: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "test.tar")
			}(t),
			wantErr: true,
		},
		{
			name: "tar doesn't have 2 empty blocks",
			tarPath: func(t *testing.T) string {
				t.Helper()
				tarFile, err := os.CreateTemp(t.TempDir(), "test.tar")
				require.NoError(t, err)
				err = tarFile.Close()
				require.NoError(t, err)
				return tarFile.Name()
			}(t),
			wantErr: true,
		},
		{
			name: "tar has 2 blocks, but not all 0",
			tarPath: func(t *testing.T) string {
				t.Helper()
				tarFile, err := os.CreateTemp(t.TempDir(), "test.tar")
				require.NoError(t, err)
				data := make([]byte, 2*TarBlockSize)
				data[10] = 1
				n, err := tarFile.Write(data)
				require.NoError(t, err)
				require.Equal(t, 2*TarBlockSize, n)
				err = tarFile.Close()
				require.NoError(t, err)
				return tarFile.Name()
			}(t),
			wantErr: true,
		},
		{
			name: "valid tar file",
			tarPath: func(t *testing.T) string {
				t.Helper()
				tarFile, err := os.CreateTemp(t.TempDir(), "test.tar")
				require.NoError(t, err)
				n, err := tarFile.Write(make([]byte, 2*TarBlockSize))
				require.NoError(t, err)
				require.Equal(t, 2*TarBlockSize, n)
				err = tarFile.Close()
				require.NoError(t, err)
				return tarFile.Name()
			}(t),
			wantErr: false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			gotWriter, err := NewBackupWriter(tt.tarPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotWriter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotWriter)
				assert.NotNil(t, gotWriter.file)
				assert.NotNil(t, gotWriter.tarWriter)
			}
		})
	}
}

func TestBackupWriter_AddDir(t *testing.T) {
	data := []byte("test data")
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test")
	err := os.MkdirAll(testDir, 0o755)
	require.NoError(t, err)

	// Create some test files and directories
	files := []string{
		"file1.txt",
		"file2.txt",
		"dir1/file3.txt",
	}
	for _, f := range files {
		fPath := filepath.Join(testDir, f)
		err := os.MkdirAll(filepath.Dir(fPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(fPath, data, 0o644)
		require.NoError(t, err)
	}

	// Create a backup writer for the test tar file
	tarPath := filepath.Join(tmpDir, "test.tar")
	err = os.WriteFile(tarPath, make([]byte, 2*TarBlockSize), 0o644)
	require.NoError(t, err)
	backupWriter, err := NewBackupWriter(tarPath)
	require.NoError(t, err)
	defer func() {
		err := backupWriter.Close()
		require.NoError(t, err)
	}()

	// Add the test directory to the backup tar file
	err = backupWriter.AddDir(testDir, "test")
	require.NoError(t, err)

	// Verify that the files were added to the tar file
	tarFile, err := os.Open(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)

	var tarFiles []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		tarFiles = append(tarFiles, header.Name)
	}

	assert.Equal(t, []string{
		"test",
		"test/dir1",
		"test/dir1/file3.txt",
		"test/file1.txt",
		"test/file2.txt",
	}, tarFiles)
}

func TestBackupWriter_AddFile(t *testing.T) {
	tmpDir := t.TempDir()
	testData := []byte("test data")
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, testData, 0o644)
	require.NoError(t, err)

	// Create a backup writer for the test tar file
	tarPath := filepath.Join(tmpDir, "test.tar")
	err = os.WriteFile(tarPath, make([]byte, 2*TarBlockSize), 0o644)
	require.NoError(t, err)
	backupWriter, err := NewBackupWriter(tarPath)
	require.NoError(t, err)
	defer func() {
		err := backupWriter.Close()
		require.NoError(t, err)
	}()

	// Add the test file to the backup tar file
	err = backupWriter.AddFile(testFile, "test.txt")
	require.NoError(t, err)

	// Verify that the file was added correctly
	tarFile, err := os.Open(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	var tarFiles []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		tarFiles = append(tarFiles, header.Name)
	}

	assert.Equal(t, []string{
		"test.txt",
	}, tarFiles)
}
