package utils

import (
	"errors"
	"os"
	"path/filepath"
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
func FindProjectRoot() (string, error) {
	// Start from the current directory and move upwards until we find a main.go file
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "main.go")); !os.IsNotExist(err) {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", os.ErrNotExist
		}
		currentDir = parentDir
	}
}
