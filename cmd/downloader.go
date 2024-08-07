package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Function to download a file from a given URL and save it to a specified path
func downloadFile(element *Element, destination string) error {
	// Download the data from the element's URL
	resp, err := http.Get(element.URL)
	if err != nil {
		return fmt.Errorf("error getting file from %s: %v", element.URL, err)
	}
	defer resp.Body.Close()

	// Create the downloadPath of the file that is being downloaded
	downloadPath := filepath.Join(destination, element.Name)
	downloadPath, err = filepath.Abs(downloadPath)
	if err != nil {
		return fmt.Errorf("error creating abs: %s | %v", downloadPath, err)
	}

	// Ensure it is a new file, os.Create will truncate the file which will corrupt it
	err = os.Remove(downloadPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing potential existing file %s: %v", downloadPath, err)
	}

	// Create the file that will be written with the downloaded data
	file, err := os.Create(downloadPath)
	if err != nil {
		return fmt.Errorf("error creating file: %s: %v", downloadPath, err)
	}
	defer file.Close()

	// Copy the data from the HTTP response into the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(file.Name())
		return fmt.Errorf("error copying contents of response to file: %s", downloadPath)
	}

	// Cache the abs of where the file was downloaded for potiental future use
	element.Path = downloadPath

	return nil
}
