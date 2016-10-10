package client_test

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/craigfurman/ezxfer/client"
	"github.com/craigfurman/ezxfer/client/fakes"
	"github.com/craigfurman/ezxfer/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sending files", func() {
	var (
		tempDir  string
		listener net.Listener

		progressBarFactory *fakes.FakeProgressBarFactory
		progressBar        *fakes.FakeProgressBar
		c                  *client.Client
	)

	BeforeEach(func() {
		progressBarFactory = new(fakes.FakeProgressBarFactory)
		progressBar = fakes.NewFakeProgressBar()
		progressBarFactory.NewReturns(progressBar)
		c = &client.Client{ProgressBarFactory: progressBarFactory}

		var err error
		tempDir, err = ioutil.TempDir("", "ezxfer-tests")
		Expect(err).NotTo(HaveOccurred())
		Expect(testhelpers.CreateFile("some content\n", tempDir, "subdirectory", "a_file.txt")).To(Succeed())
		listener, err = net.Listen("tcp", "127.0.0.1:45454")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		listener.Close()
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	It("sends a tar stream", func() {
		errs := make(chan error)

		go func() {
			errs <- c.Send(filepath.Join(tempDir), "127.0.0.1:45454")
		}()

		conn, err := listener.Accept()
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		tarStream := tar.NewReader(conn)
		header, err := tarStream.Next()
		Expect(err).NotTo(HaveOccurred())
		Expect(header.Name).To(Equal("subdirectory/a_file.txt"))
		Expect(header.Xattrs["md5"]).To(Equal("eb9c2bf0eb63f3a7bc0ea37ef18aeba5"))

		content, err := ioutil.ReadAll(tarStream)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("some content\n"))

		_, err = tarStream.Next()
		Expect(err).To(MatchError(io.EOF))

		Expect(<-errs).NotTo(HaveOccurred())

		Expect(progressBarFactory.NewCallCount()).To(Equal(1))
		Expect(progressBarFactory.NewArgsForCall(0)).To(Equal(int64(13)))
		Expect(progressBar.String()).To(Equal("some content\n"))
		Expect(progressBar.FinishCallCount()).To(Equal(1))
	})
})
