package filesystem

import "os"

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}
func CreateDirIfNotExists(path string) error {
	if !dirExists(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
