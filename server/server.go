package server

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

type Server struct {
	Port    int
	Logger  *log.Logger
	destDir string
}

func (s *Server) ServeTCP() error {
	var err error
	s.destDir, err = os.Getwd()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.transferFile(conn)
	}
}

func (s *Server) transferFile(conn net.Conn) {
	defer conn.Close()
	tarStream := tar.NewReader(conn)
	header, err := tarStream.Next()
	if err != nil {
		s.Logger.Println(err)
		return
	}

	filePath := filepath.Join(s.destDir, header.FileInfo().Name())
	s.Logger.Printf("saving file to %s", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		s.Logger.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, tarStream)
	if err != nil {
		s.Logger.Println(err)
		return
	}
}
