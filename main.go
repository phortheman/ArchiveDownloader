package main

import (
	"flag"
	"fmt"
	"github.com/phortheman/ArchiveDownloader/cmd"
	"os"
)

var url, destination string
var extension string
var numWorkers int

func init() {
	flag.StringVar(&url, "url", "", "The URL to pull from")
	flag.StringVar(&destination, "dest", "", "The destination to store the files")
	flag.IntVar(&numWorkers, "work", 4, "The number of workers to spawn to download files")
	flag.StringVar(&extension, "ext", "", "Expected extension to be downloading")
}

func main() {
	// Handle flag errors
	flag.Parse()
	if url == "" {
		fmt.Fprintln(os.Stderr, "url is required")
		os.Exit(2)
	}

	if destination == "" {
		fmt.Fprintln(os.Stderr, "destination path is required")
		os.Exit(2)
	}
	os.Exit(cmd.Execute(cmd.Options{
		URL:               url,
		Destination:       destination,
		NumWorkers:        numWorkers,
		ExpectedExtension: extension,
	}))
}
