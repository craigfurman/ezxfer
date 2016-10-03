package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/craigfurman/ezxfer/client"
	"github.com/craigfurman/ezxfer/server"
)

func main() {
	file := flag.String("file", "", "")
	dstHost := flag.String("dstHost", "", "")
	dstPort := flag.Int("dstPort", 0, "")

	serverPort := flag.Int("serveOnPort", 0, "")
	flag.Parse()

	if *serverPort != 0 {
		logger := createLogger("[ezxfer server] ")
		logger.Println("-serveOnPort is set, starting in server mode")
		srv := server.Server{Port: *serverPort, Logger: logger}
		if err := srv.ServeTCP(); err != nil {
			logger.Println(err)
			os.Exit(1)
		}
	}

	logger := createLogger("[ezxfer] ")
	logger.Printf("will transfer file %s to %s:%d...\n", *file, *dstHost, *dstPort)
	if err := client.SendFile(*file, fmt.Sprintf("%s:%d", *dstHost, *dstPort)); err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	logger.Println("done!")
}

func createLogger(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, log.LstdFlags)
}
