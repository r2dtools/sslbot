package hostmng

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/r2dtools/sslbot/internal/utils"
)

type NginxHostManager struct{}

func (m *NginxHostManager) Enable(configFilePath, enabledConfigRootPath string) (string, error) {
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("config file not found: %s", configFilePath)
	}

	isSymlink, err := utils.IsSymlink(configFilePath)

	if err != nil {
		return "", err
	}

	if isSymlink {
		// if symlink - host already enabled
		return configFilePath, nil
	}

	var enabledConfigFilePath string
	fileName := filepath.Base(configFilePath)
	enabledConfigFilePath = filepath.Join(enabledConfigRootPath, fileName)

	if _, err := os.Lstat(enabledConfigFilePath); errors.Is(err, os.ErrNotExist) {
		err = os.Symlink(configFilePath, enabledConfigFilePath)

		if err != nil {
			return "", err
		}
	}

	return enabledConfigFilePath, nil
}

func (m *NginxHostManager) Disable(enabledConfigFilePath string) error {
	var err error

	if _, err = os.Lstat(enabledConfigFilePath); err == nil {
		if err = os.Remove(enabledConfigFilePath); err != nil {
			return fmt.Errorf("failed to remove config file symlink: %s, err: %v", enabledConfigFilePath, err)
		}
	}

	return err
}
