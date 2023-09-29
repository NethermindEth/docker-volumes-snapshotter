package backuptar

const (
	// FileName is the name of the backup file
	FileName = "backup.tar"

	// Path is the path to the backup file inside the container. It should be
	// the target of the volume mount for the backup file.
	Path = "/" + FileName
)
