package server

import (
	"archive/tar"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

type Server struct {
	Port    int
	DestDir string
	Logger  *log.Logger
}

type acceptedConnection struct {
	conn net.Conn
	err  error
}

func (s *Server) ServeTCP(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	defer listener.Close()

	connChan := make(chan acceptedConnection)

	for {
		go func() {
			conn, err := listener.Accept()
			connChan <- acceptedConnection{conn: conn, err: err}
		}()

		select {
		case connection := <-connChan:
			if connection.err != nil {
				return connection.err
			}
			go s.receiveFiles(connection.conn)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) receiveFiles(conn net.Conn) {
	defer conn.Close()
	tarStream := tar.NewReader(conn)

	for {
		header, err := tarStream.Next()
		if err != nil {
			if err != io.EOF {
				s.Logger.Println(err)
				return
			}
			break
		}

		filePath := filepath.Join(s.DestDir, header.Name)
		if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			s.Logger.Println(err)
			return
		}

		s.Logger.Printf("saving file to %s", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			s.Logger.Println(err)
			return
		}
		defer file.Close()

		checksumWriter := md5.New()
		fileAndChecksum := io.MultiWriter(file, checksumWriter)

		if _, err := io.Copy(fileAndChecksum, tarStream); err != nil {
			s.Logger.Println(err)
			return
		}

		md5Sum := hex.EncodeToString(checksumWriter.Sum(nil))
		expectedMd5Sum := header.Xattrs["md5"]
		if md5Sum != expectedMd5Sum {
			msg := fmt.Sprintf("md5 does not match: expected %s, got %s", expectedMd5Sum, md5Sum)
			s.Logger.Println(msg)
			if _, err := conn.Write([]byte(msg)); err != nil {
				s.Logger.Println(err)
			}
			return
		}
	}

	if _, err := conn.Write([]byte("OK")); err != nil {
		s.Logger.Println(err)
	}
}
