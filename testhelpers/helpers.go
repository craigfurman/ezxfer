package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateFile(content string, nameParts ...string) error {
	fullPath, err := ensureDirExists(nameParts)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullPath, []byte(content), 0644)
}

func ensureDirExists(nameParts []string) (string, error) {
	fullPath := filepath.Join(nameParts...)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}
	return fullPath, nil
}
