package utils

import (
	"errors"
	"os"
)

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func GetKeyByValue(m map[string]int, value int) string {
	for key, val := range m {
		if val == value {
			return key
		}
	}
	return "" // If value not found, return an empty string or handle it as needed
}
func CheckDirectoryExists(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}

	}
	if file != nil {
		if !file.IsDir() {
			return errors.New(path + " is not a directory")
		}

	} else {
		return errors.New(path + " stat directory error" + path)
	}

	return nil
}
