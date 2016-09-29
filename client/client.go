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

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(fileInfo, filepath.Base(filePath))
	if err != nil {
		return err
	}
	if err := tarStream.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tarStream, file); err != nil {
		return err
	}

	return nil
}
