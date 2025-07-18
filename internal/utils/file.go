package utils

import "os"

func IsSymlink(path string) (bool, error) {
	var isSymlink bool
	stat, err := os.Lstat(path)

	if err != nil {
		return isSymlink, err
	}

	isSymlink = (stat.Mode() & os.ModeSymlink) == os.ModeSymlink

	return isSymlink, nil
}
