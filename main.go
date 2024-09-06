package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/phortheman/ArchiveDownloader/cmd"
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

	// Create the main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		// TODO: Should include this?
		// signal.Notify(c, os.Kill)
		<-c
		cancel()
		// TODO: May need to do os.Exit(1) here?
	}()

	os.Exit(cmd.Execute(
		ctx,
		cmd.Options{
			URL:               url,
			Destination:       destination,
			NumWorkers:        numWorkers,
			ExpectedExtension: extension,
		}))
}
