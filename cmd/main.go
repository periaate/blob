package main

import (
	"fmt"
	"log"
	"net/http"

	"blob"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/fsio"
)

func main() {
	clog.SetLogLoggerLevel(clog.LevelDebug)
	args := fsio.Args()
	root := "./blob/"
	if len(args) > 1 {
		root = args[1]
	}

	clog.Debug("root dir", "root", root)

	err := fsio.EnsureDir(root)
	if err != nil {
		clog.Fatal("error creating root dir", "err", err)
	}

	storage := blob.New(root)

	addr := "127.0.0.1:8085"
	clog.Info("starting server", "addr", fmt.Sprintf("http://%s", addr))
	log.Fatal(http.ListenAndServe(addr, storage))
}
