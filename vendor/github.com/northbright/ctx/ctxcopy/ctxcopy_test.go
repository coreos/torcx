package ctxcopy_test

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/northbright/ctx/ctxcopy"
	"github.com/northbright/pathhelper"
	"golang.org/x/net/context"
)

// 1. Run "go get github.com/northbright/pathhelper" to install pathhelper.
// 2. Run "go test -c && ./ctxcopy.test"
func ExampleCopy() {
	// Download a zip from web server to local storage to test Copy().
	url := "https://github.com/northbright/plants/archive/master.zip"
	totalTimeoutSeconds := 10 // to make download successful, set it to 300 or more.
	totalTimeout := time.Duration(time.Duration(totalTimeoutSeconds) * time.Second)

	// Make context to carry a deadline(timeout).
	// See http://blog.golang.org/context for more information.
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	// Get response body for source.
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Get(%v) err: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	// Create a file for destination.
	fileName, _ := pathhelper.GetAbsPath("./1.zip")
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Create(%v) err: %v\n", fileName, err)
		return
	}
	defer f.Sync()
	defer f.Close()

	buf := make([]byte, 2*1024*1024)

	// Copy starts.
	// Copy operation will be canceled if cancel() is called in other goroutine.
	// Copy operation will be stoped if deadline is exceeded(timeout).
	err = ctxcopy.Copy(ctx, f, resp.Body, buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ctxcopy.Copy() err: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "ctxcopy.Copy() succeeded.\n")
	}

	// Output:
}
