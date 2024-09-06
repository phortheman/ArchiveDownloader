package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var elements []*Element

var logInfo = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var logError = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

type Element struct {
	Name string
	Size uint // TODO: Need to parse the expected file size for seeing if the file was partially downloaded
	URL  string
	Path string
}

type Options struct {
	URL               string
	Destination       string
	NumWorkers        int
	ExpectedExtension string
}

// Set default options if they weren't set
func (o *Options) SetDefaults() {
	if o.NumWorkers == 0 {
		o.NumWorkers = 4
	}
}

// Check if the file exists (0 byte files don't count)
func fileExists(fileName, destination string) bool {
	path := filepath.Join(destination, fileName)
	file, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return file.Size() > 0
}

// Generate a list of files from the URL and download the files to the destination
func Execute(ctx context.Context, ops Options) int {
	ops.SetDefaults()
	var err error

	// Create the destination absolute path
	ops.Destination, err = filepath.Abs(ops.Destination)
	if err != nil {
		logError.Println(err)
		return 1
	}

	// Create the destination directory if it doesn't exist
	err = os.MkdirAll(ops.Destination, os.ModePerm)
	if err != nil {
		logError.Printf("Error making directory '%s': %v", ops.Destination, err)
		return 1
	}

	// Make sure the URL is setup to handle relative paths to be appended
	if len(ops.URL) > 0 && ops.URL[len(ops.URL)-1] != '/' {
		ops.URL += "/"
	}

	// Parse the HTML from the URL and use the callback function to populate the elements slice
	callback := func(td *html.Node) {
		if td.Type == html.ElementNode && td.Data == "td" {
			for node := td.FirstChild; node != nil; node = node.NextSibling {
				if node.Type == html.ElementNode && node.Data == "a" {
					e := Element{}
					e.Name = extractText(node)
					// Handle edge case
					if strings.TrimSpace(e.Name) == "Go to parent directory" {
						continue
					}
					for _, attr := range node.Attr {
						if attr.Key == "href" {
							e.URL = ops.URL + attr.Val
							break
						}
					}
					if e.URL != "" && e.Name != "" {
						elements = append(elements, &e)
					}
					break
				}
			}
		}
	}
	logInfo.Printf("Fetching page: %s\n", ops.URL)

	// TODO: May need to add select wrapper for this one
	err = getFileNamesAndURLs(ctx, ops.URL, callback)
	if err != nil {
		logInfo.Printf("error: %v\n", err)
		return 1
	}

	// Download files concurrently because I can
	var wg sync.WaitGroup

	// Define a worker task
	worker := func(ctx context.Context, id int, jobs <-chan *Element) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logInfo.Printf("Worker %d: Shutting down due to cancellation\n", id)
				return
			case e, ok := <-jobs:
				if !ok {
					// Channel is closed, worker can exit
					return
				}
				// If we whitelist a specific extension skip any that don't have it
				if ops.ExpectedExtension != "" && !strings.HasSuffix(e.Name, ops.ExpectedExtension) {
					logInfo.Printf("Worker %d:	File name doesn't have expected extension: %s\n", id, e.Name)
					continue
				}

				// If the file exists, skip it
				if fileExists(e.Name, ops.Destination) {
					logInfo.Printf("Worker %d:	File already exists, skipping: %s\n", id, e.Name)
					continue
				}
				log.Printf("Worker %d: Downloading %s\n", id, e.Name)
				err := downloadFile(ctx, e, ops.Destination)
				if err != nil {
					logInfo.Printf("Worker %d: Error: %v\n", id, err)
					continue
				}
				logInfo.Printf("Worker %d:	Finished downloading %s\n", id, e.Name)
			}
		}
	}

	// Create the element channel to feed workers
	elementChan := make(chan *Element)

	// Start the workers
	for i := 1; i <= ops.NumWorkers; i++ {
		wg.Add(1)
		go worker(ctx, i, elementChan)
	}

	// Feed elements to workers
	go func() {
		for _, e := range elements {
			select {
			case <-ctx.Done():
				return
			case elementChan <- e:
			}
		}
		close(elementChan)
	}()

	// Wait for all workers to finish
	wg.Wait()

	return 0
}
