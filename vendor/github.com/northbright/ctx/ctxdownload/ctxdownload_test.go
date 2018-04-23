package ctxdownload_test

import (
	"fmt"
	"os"
	"time"

	"github.com/northbright/ctx/ctxdownload"
	"golang.org/x/net/context"
)

// Run "go test -c && ./ctxdownload.test"
func ExampleDownload() {
	// Download a zip file to test ctxdownload.Download().
	url := "https://github.com/northbright/plants/archive/master.zip"
	outDir := "./download"
	fileName := "" // If file name is empty, it will try to detect file name in response Header in Download().

	// Set total timeout. You'll get "deadline exceeded" error as expected.
	// To make download successful, set it to 300 or longer.
	totalTimeoutSeconds := 20
	totalTimeout := time.Duration(time.Duration(totalTimeoutSeconds) * time.Second)

	// Make a context to carry a deadline(timeout).
	// See http://blog.golang.org/context to for more information.
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	// HTTP request timeout(NOT download timeout).
	requestTimeoutSeconds := 10

	// Make buffer.
	buf := make([]byte, 2*1024*1024)

	// Start download.
	// It will cancel download if cancel() is called in other goroutines.
	// It will stop download if deadline is exceeded(timeout).
	downloadedFileName, err := ctxdownload.Download(ctx, url, outDir, fileName, buf, requestTimeoutSeconds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ctxdownload.Download() err: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "ctxdownload.Download() succeeded.\nDownloaded File: %v\n", downloadedFileName)
	}

	// Output:
}
