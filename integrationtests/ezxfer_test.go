package integrationtests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/craigfurman/ezxfer/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func readFile(nameParts ...string) (string, error) {
	actualContent, err := ioutil.ReadFile(filepath.Join(nameParts...))
	if err != nil {
		return "", err
	}
	return string(actualContent), nil
}

var _ = Describe("transferring files", func() {
	var (
		tempDir       string
		destDir       string
		serverPort    = 45454
		serverProcess *gexec.Session
		sourceFiles   string
		clientStdout  *bytes.Buffer
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "ezxfer-tests")
		Expect(err).NotTo(HaveOccurred())
		destDir = filepath.Join(tempDir, "dest")
		Expect(os.MkdirAll(destDir, 0755)).To(Succeed())
		serverCmd := exec.Command(binPath, fmt.Sprintf("-serveOnPort=%d", serverPort))
		serverCmd.Dir = destDir
		serverProcess, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(testhelpers.IsListening(fmt.Sprintf("localhost:%d", serverPort))).Should(BeTrue())
	})

	JustBeforeEach(func() {
		clientCmd := exec.Command(binPath, "-file", sourceFiles, "-dstHost", "localhost", fmt.Sprintf("-dstPort=%d", serverPort))
		clientStdout = new(bytes.Buffer)
		clientProcess, err := gexec.Start(clientCmd, io.MultiWriter(clientStdout, GinkgoWriter), GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(clientProcess).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		Eventually(serverProcess.Kill()).Should(gexec.Exit())
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Context("when the source is a single file", func() {
		var (
			fileContent = "some content"
			fileName    = "to_copy.txt"
		)

		BeforeEach(func() {
			sourceFiles = filepath.Join(tempDir, fileName)
			Expect(testhelpers.CreateFile(fileContent, sourceFiles)).To(Succeed())
		})

		It("transfers files", func() {
			actualContent, err := ioutil.ReadFile(filepath.Join(destDir, fileName))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actualContent)).To(Equal(fileContent))
		})

		It("shows a progress bar", func() {
			Expect(clientStdout.String()).To(ContainSubstring("100.00%"))
		})
	})

	Context("when the source is a directory, not a file", func() {
		BeforeEach(func() {
			sourceFiles = filepath.Join(tempDir, "some-src")
			Expect(testhelpers.CreateFile("content for a.txt", sourceFiles, "a.txt")).To(Succeed())
			Expect(testhelpers.CreateFile("content for b.txt", sourceFiles, "d1", "b.txt")).To(Succeed())
			Expect(testhelpers.CreateFile("content for c.txt", sourceFiles, "d1", "d2", "c.txt")).To(Succeed())
		})

		It("transfers the directory", func() {
			Expect(readFile(destDir, "a.txt")).To(Equal("content for a.txt"))
			Expect(readFile(destDir, "d1", "b.txt")).To(Equal("content for b.txt"))
			Expect(readFile(destDir, "d1", "d2", "c.txt")).To(Equal("content for c.txt"))
		})
	})
})
