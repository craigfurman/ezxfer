package server_test

import (
	"archive/tar"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/craigfurman/ezxfer/client"
	"github.com/craigfurman/ezxfer/server"
	"github.com/craigfurman/ezxfer/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("receiving files", func() {
	const port = 45454

	var (
		address   = fmt.Sprintf("localhost:%d", port)
		tempDir   string
		ctx       context.Context
		canceller context.CancelFunc

		s *server.Server

		serverResult chan error
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "ezxfer-server-unit-tests")
		Expect(err).NotTo(HaveOccurred())

		ctx, canceller = context.WithCancel(context.Background())

		destDir := filepath.Join(tempDir, "dest")
		Expect(os.Mkdir(destDir, 0755)).To(Succeed())

		s = &server.Server{Port: port, DestDir: destDir, Logger: log.New(GinkgoWriter, "[ezxfer server unit tests] ", log.LstdFlags)}
		serverResult = make(chan error)
		go func() {
			serverResult <- s.ServeTCP(ctx)
		}()
		Eventually(testhelpers.IsListening(address)).Should(BeTrue())
	})

	AfterEach(func() {
		canceller()
		Expect(<-serverResult).To(MatchError("context canceled"))
		Expect(os.RemoveAll(tempDir)).To(Succeed())

		Eventually(func() error {
			_, err := net.Dial("tcp", address)
			return err
		}).Should(HaveOccurred())
	})

	testServer := func(md5FromClient, expectedResponse string) {
		conn, err := net.Dial("tcp", address)
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		fileName := "a-file.txt"
		content := "some content\n"
		Expect(testhelpers.CreateFile(content, tempDir, "src", fileName)).To(Succeed())
		srcFileInfo, err := os.Stat(filepath.Join(tempDir, "src", fileName))
		Expect(err).NotTo(HaveOccurred())

		tarWriter := tar.NewWriter(conn)
		header, err := tar.FileInfoHeader(srcFileInfo, "")
		Expect(err).NotTo(HaveOccurred())
		header.Xattrs = map[string]string{client.MD5_ATTRIBUTE_KEY: md5FromClient}
		Expect(tarWriter.WriteHeader(header)).To(Succeed())
		_, err = tarWriter.Write([]byte(content))
		Expect(err).NotTo(HaveOccurred())
		Expect(tarWriter.Close()).To(Succeed())

		transferredFilePath := filepath.Join(tempDir, "dest", fileName)
		Eventually(func() error {
			_, err := os.Stat(transferredFilePath)
			return err
		}).ShouldNot(HaveOccurred())

		Expect(ioutil.ReadFile(transferredFilePath)).To(Equal([]byte(content)))

		Expect(ioutil.ReadAll(conn)).To(Equal([]byte(expectedResponse)))
	}

	It("writes the tar stream to the destination directory and confirms that checksum matches", func() {
		testServer("eb9c2bf0eb63f3a7bc0ea37ef18aeba5", "OK")
	})

	Context("when the md5 does not match", func() {
		It("replies with an error", func() {
			testServer("wrong", "md5 does not match: expected wrong, got eb9c2bf0eb63f3a7bc0ea37ef18aeba5")
		})
	})
})
