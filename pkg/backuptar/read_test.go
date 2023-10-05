package backuptar

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractDir(t *testing.T) {
	testDir := t.TempDir()
	tarFile, err := os.CreateTemp(testDir, "test.tar")
	require.NoError(t, err)
	outDir, err := os.MkdirTemp(testDir, "out")
	require.NoError(t, err)

	// Create a tar archive with some test files
	tarWriter := tar.NewWriter(tarFile)
	fileContents := []byte("test file contents")
	fileNames := []string{"dir1/file1.txt", "dir1/dir2/file2.txt", "dir3/file3.txt"}
	for _, fileName := range fileNames {
		header := &tar.Header{
			Name: fileName,
			Mode: 0o644,
			Size: int64(len(fileContents)),
		}
		err = tarWriter.WriteHeader(header)
		require.NoError(t, err, "Failed to write tar header")
		n, err := tarWriter.Write(fileContents)
		require.NoError(t, err, "Failed to write tar data")
		require.Equal(t, n, len(fileContents), "Failed to write all tar data")
	}
	err = tarWriter.Close()
	require.NoError(t, err)
	err = tarFile.Close()
	require.NoError(t, err)

	// Extract the test files from the tar archive
	err = ExtractDir(tarFile.Name(), "dir1", outDir)
	require.NoError(t, err)

	// Verify that the extracted files exist
	for _, fileName := range []string{"file1.txt", "dir2/file2.txt"} {
		filePath := filepath.Join(outDir, fileName)
		// Verify file exists
		info, err := os.Stat(filePath)
		require.NoError(t, err, "Failed to stat file")
		require.Equal(t, info.Size(), int64(len(fileContents)))
		// Verify content
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, data, fileContents)
	}

	// Verify that other files were not extracted
	for _, fileName := range []string{"file3.txt"} {
		filePath := filepath.Join(outDir, fileName)
		_, err = os.Stat(filePath)
		require.ErrorIs(t, err, os.ErrNotExist)
	}
}

func TestExtractFile(t *testing.T) {
	fileContents := []byte("test file contents")

	tests := []struct {
		name           string
		tarContents    []byte
		srcPath        string
		targetPath     string
		expectedError  error
		expectedExists bool
		expectedData   []byte
	}{
		{
			name: "Extract file from tar archive",
			tarContents: func(t *testing.T) []byte {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				defer tarWriter.Close()
				header := &tar.Header{
					Name: "dir1/file1.txt",
					Mode: 0o644,
					Size: int64(len(fileContents)),
				}
				err := tarWriter.WriteHeader(header)
				require.NoError(t, err)
				_, err = tarWriter.Write(fileContents)
				require.NoError(t, err)
				err = tarWriter.Close()
				require.NoError(t, err)
				return buf.Bytes()
			}(t),
			srcPath:        "dir1/file1.txt",
			targetPath:     "file1.txt",
			expectedError:  nil,
			expectedExists: true,
			expectedData:   fileContents,
		},
		{
			name:           "Tar archive does not exist",
			tarContents:    nil,
			srcPath:        "file.txt",
			targetPath:     "out.txt",
			expectedError:  os.ErrNotExist,
			expectedExists: false,
			expectedData:   nil,
		},
		{
			name: "Source file does not exist",
			tarContents: func(t *testing.T) []byte {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				defer tarWriter.Close()
				header := &tar.Header{
					Name: "dir1/file1.txt",
					Mode: 0o644,
					Size: int64(len(fileContents)),
				}
				err := tarWriter.WriteHeader(header)
				require.NoError(t, err)
				_, err = tarWriter.Write(fileContents)
				require.NoError(t, err)
				err = tarWriter.Close()
				require.NoError(t, err)
				return buf.Bytes()
			}(t),
			srcPath:        "nonexistent.txt",
			targetPath:     "out.txt",
			expectedError:  os.ErrNotExist,
			expectedExists: false,
			expectedData:   nil,
		},
		{
			name: "Target directory does not exist",
			tarContents: func() []byte {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				defer tarWriter.Close()
				header := &tar.Header{
					Name: "dir1/file1.txt",
					Mode: 0o644,
					Size: int64(len(fileContents)),
				}
				err := tarWriter.WriteHeader(header)
				require.NoError(t, err)
				_, err = tarWriter.Write(fileContents)
				require.NoError(t, err)
				err = tarWriter.Close()
				require.NoError(t, err)
				return buf.Bytes()
			}(),
			srcPath:        "dir1/file1.txt",
			targetPath:     "nonexistent/out.txt",
			expectedError:  nil,
			expectedExists: true,
			expectedData:   fileContents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write the tar archive to a file
			if tt.tarContents != nil {
				err := os.WriteFile(filepath.Join(tmpDir, "test.tar"), tt.tarContents, 0o644)
				require.NoError(t, err)
			}

			// Extract the file from the tar archive
			err := ExtractFile(filepath.Join(tmpDir, "test.tar"), tt.srcPath, filepath.Join(tmpDir, tt.targetPath))
			if tt.expectedError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify that the extracted file exists and has the correct contents
			filePath := filepath.Join(tmpDir, tt.targetPath)
			if tt.expectedExists {
				require.FileExists(t, filePath)
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				require.Equal(t, tt.expectedData, data)
			} else {
				require.NoFileExists(t, filePath)
			}
		})
	}
}
