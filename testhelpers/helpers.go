package testhelpers

import (
	"io/ioutil"
	"net"
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

func IsListening(address string) func() bool {
	return func() bool {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}
}
