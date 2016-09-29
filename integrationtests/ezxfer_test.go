package integrationtests

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("transferring files", func() {
	var (
		tempDir       string
		destDir       string
		serverPort    = 45454
		serverProcess *gexec.Session
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
		Eventually(func() bool {
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", serverPort))
			if err != nil {
				return false
			}
			conn.Close()
			return true
		}).Should(BeTrue())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
		Expect(serverProcess.Kill()).To(gexec.Exit())
	})

	It("transfers files", func() {
		fileContent := "some content"
		fileName := "to_copy.txt"
		filePath := filepath.Join(tempDir, fileName)
		Expect(ioutil.WriteFile(filePath, []byte(fileContent), 0644)).To(Succeed())

		clientCmd := exec.Command(binPath, "-file", filePath, "-dstHost", "localhost", fmt.Sprintf("-dstPort=%d", serverPort))
		clientProcess, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(clientProcess).Should(gexec.Exit(0))

		actualContent, err := ioutil.ReadFile(filepath.Join(destDir, fileName))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(actualContent)).To(Equal(fileContent))
	})
})
