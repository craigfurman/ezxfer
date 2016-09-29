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
	logger := log.New(os.Stdout, "[ezxfer] ", log.LstdFlags)

	file := flag.String("file", "", "")
	dstHost := flag.String("dstHost", "", "")
	dstPort := flag.Int("dstPort", 0, "")

	serverPort := flag.Int("serveOnPort", 0, "")
	flag.Parse()

	if *serverPort != 0 {
		logger.Println("-serveOnPort is set, starting in server mode")
		srv := server.Server{Port: *serverPort, Logger: logger}
		if err := srv.ServeTCP(); err != nil {
			logger.Println(err)
			os.Exit(1)
		}
	}

	logger.Println("will transfer file...")
	if err := client.SendFile(*file, fmt.Sprintf("%s:%d", *dstHost, *dstPort)); err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	logger.Println("done!")
}
