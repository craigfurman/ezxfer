package client

import (
	"archive/tar"
	"io"
	"net"
	"os"
	"path/filepath"
)

func SendFile(filePath, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	tarStream := tar.NewWriter(conn)
	defer tarStream.Close()

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return handleDirectory(filePath, tarStream)
	}

	return handleSingleFile(filepath.Dir(filePath), filePath, tarStream)
}

func handleSingleFile(basePath string, filePath string, tarStream *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	relativePath, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fileInfo, "I am not needed")
	if err != nil {
		return err
	}
	header.Name = relativePath
	if err := tarStream.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tarStream, file); err != nil {
		return err
	}

	return nil
}

func handleDirectory(filePath string, tarStream *tar.Writer) error {
	return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return handleSingleFile(filePath, path, tarStream)
	})
}
