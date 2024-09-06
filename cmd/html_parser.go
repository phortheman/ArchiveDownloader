package cmd

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

// Parse the HTML page to populate the list of file names and urls
// Callback function handles what to do with with the table data
func getFileNamesAndURLs(ctx context.Context, url string, callback func(*html.Node)) error {
	logInfo.Printf("Fetching page: %s\n", url)
	// Send a GET request to the URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the HTML document
	logInfo.Println("Parsing...")
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}

	table := findFirstElementByClass(doc, "table", "directory-listing-table")
	body := findTableBody(table)
	parseTableRows(body, callback)

	return nil
}

// Find either the tbody or return the table
func findTableBody(table *html.Node) *html.Node {
	var result *html.Node = table
	traverse(table, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			result = n
			return
		}
	})
	return result
}

// Find the first node in the tree with the matching element and class
func findFirstElementByClass(n *html.Node, element, class string) *html.Node {
	var result *html.Node
	traverse(n, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == element {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == class {
					result = n
					return
				}
			}
		}
	})
	return result
}

// Function to parse table rows. Either pass a table or tbody
func parseTableRows(table *html.Node, callback func(n *html.Node)) {
	traverse(table, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				callback(c)
			}
		}
	})
}

// Recursively traverse nodes
func traverse(n *html.Node, f func(*html.Node)) {
	if n != nil {
		f(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c, f)
		}
	}
}

// Function to extract text from a node
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += extractText(c)
	}
	return result
}
