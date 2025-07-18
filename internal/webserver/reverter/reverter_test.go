package reverter

import (
	"os"
	"testing"

	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/unknwon/com"
)

type stubHostManager struct{}

func (m stubHostManager) Enable(configFilePath, originSslConfigFilePath string) (string, error) {
	return "", nil
}

func (m stubHostManager) Disable(configFilePath string) error {
	return nil
}

func TestReverterRollback(t *testing.T) {
	reverter := getReverter(t)
	fileToBackup := "/tmp/fileToRemove"
	createFile(t, fileToBackup)
	err := reverter.BackupConfigs([]string{fileToBackup})
	assert.Nilf(t, err, "could not backup files: %v", err)
	bFileToBackup := reverter.getBackupConfigPath(fileToBackup)
	assert.Equalf(t, true, com.IsExist(bFileToBackup), "backed up file '%s' does not exist", bFileToBackup)
	err = reverter.Rollback()
	assert.Nilf(t, err, "revert error: %v", err)
	assert.Equalf(t, false, com.IsExist(bFileToBackup), "backed up file '%s' steel exists", bFileToBackup)

	reverter.AddConfigToDeletion(fileToBackup)
	assert.Equalf(t, true, com.IsExist(fileToBackup), "file '%s' does not exist", fileToBackup)
	err = reverter.Rollback()
	assert.Nilf(t, err, "revert error: %v", err)
	assert.Equalf(t, false, com.IsExist(fileToBackup), "file '%s' steel exists", fileToBackup)
}

func TestReverterCommit(t *testing.T) {
	reverter := getReverter(t)
	fileToBackup := "/tmp/fileToRemove"
	createFile(t, fileToBackup)
	err := reverter.BackupConfigs([]string{fileToBackup})
	assert.Nilf(t, err, "could not backup files: %v", err)
	bFileToBackup := reverter.getBackupConfigPath(fileToBackup)
	assert.Equalf(t, true, com.IsExist(bFileToBackup), "backed up file '%s' does not exist", bFileToBackup)
	err = reverter.Commit()
	assert.Nilf(t, err, "revert error: %v", err)
	assert.Equalf(t, false, com.IsExist(bFileToBackup), "backed up file '%s' steel exists", bFileToBackup)
	assert.Equalf(t, true, com.IsExist(fileToBackup), "file '%s' does not exist", fileToBackup)
}

func getReverter(t *testing.T) *DefaultReverter {
	reverter := DefaultReverter{
		logger:  &logger.TestLogger{T: t},
		hostMng: stubHostManager{},
	}

	return &reverter
}

func createFile(t *testing.T, path string) {
	err := os.WriteFile(path, []byte("content"), 0644)
	assert.Nilf(t, err, "could not create tmp file: %v", err)
	assert.Equal(t, true, com.IsExist(path), "create file does not exist")
}
