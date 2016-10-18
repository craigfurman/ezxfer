package client

import (
	"archive/tar"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

const MD5_ATTRIBUTE_KEY = "md5"

type Client struct {
	ProgressBarFactory ProgressBarFactory
}

//go:generate counterfeiter -o fakes/fake_progress_bar_factory.go . ProgressBarFactory
type ProgressBarFactory interface {
	New(fileSize int64) ProgressBar
}

type ProgressBar interface {
	io.Writer
	Finish()
}

func (c *Client) Send(filePath, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	tarStream := tar.NewWriter(conn)
	defer conn.Close()

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := c.sendDir(filePath, tarStream); err != nil {
			return err
		}
	} else {
		if err := c.sendFile(filepath.Dir(filePath), filePath, tarStream); err != nil {
			return err
		}
	}
	if err := tarStream.Close(); err != nil {
		return fmt.Errorf("error closing tar stream: %s", err)
	}

	reply, err := ioutil.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("error reading reply: %s", err)
	}

	if string(reply) == "OK" {
		return nil
	}

	return errors.New(string(reply))
}

func (c *Client) sendFile(basePath string, filePath string, tarStream *tar.Writer) error {
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

	progressBar := c.ProgressBarFactory.New(fileInfo.Size())
	progressTrackingFileReader := io.TeeReader(file, progressBar)
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
	header.Xattrs = map[string]string{MD5_ATTRIBUTE_KEY: md5Checksum}

	if err := tarStream.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tarStream, progressTrackingFileReader); err != nil {
		return err
	}

	return nil
}

func (c *Client) sendDir(filePath string, tarStream *tar.Writer) error {
	return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return c.sendFile(filePath, path, tarStream)
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
