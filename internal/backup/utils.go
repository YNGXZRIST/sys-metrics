package backup

import (
	"os"
	"path"
)

func (bc *BackupConfig) Cleanup() error {
	if bc.Writer != nil {
		_ = bc.Writer.Close()
	}
	if bc.Reader != nil {
		_ = bc.Reader.Close()
	}
	if bc.filePath != path.Join(bc.StoragePath, DefaultFileName) {
		return os.RemoveAll(bc.StoragePath)
	}
	return nil
}
func (bc *BackupConfig) getBackupFilename() string {
	return bc.filePath
}
