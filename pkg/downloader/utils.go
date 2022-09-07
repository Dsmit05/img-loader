package downloader

import "os"

func CreatePath(nameDir string) error {
	err := os.MkdirAll(nameDir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
