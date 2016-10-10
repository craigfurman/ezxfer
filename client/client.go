package client

import (
	"archive/tar"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"os"
	"path/filepath"

	pb "gopkg.in/cheggaaa/pb.v1"
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
		return sendDir(filePath, tarStream)
	}

	return sendFile(filepath.Dir(filePath), filePath, tarStream)
}

func sendFile(basePath string, filePath string, tarStream *tar.Writer) error {
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

	progressBar := pb.New64(fileInfo.Size()).SetUnits(pb.U_BYTES)
	progressBar.Start()
	progressTrackingFileReader := progressBar.NewProxyReader(file)
	defer progressBar.Finish()

	header, err := tar.FileInfoHeader(fileInfo, "What even is this? It seems to make no difference")
	if err != nil {
		return err
	}
	header.Name = relativePath

	md5Checksum, err := checksum(filePath)
	if err != nil {
		return err
	}
	header.Xattrs = map[string]string{"md5": md5Checksum}

	if err := tarStream.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tarStream, progressTrackingFileReader); err != nil {
		return err
	}

	return nil
}

func sendDir(filePath string, tarStream *tar.Writer) error {
	return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return sendFile(filePath, path, tarStream)
	})
}

func checksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	md5Writer := md5.New()

	if _, err := io.Copy(md5Writer, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(md5Writer.Sum(nil)), nil
}
