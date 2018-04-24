# ctxdownload

ctxdownload is a [Golang](http://golang.org) package which provides helper functions for performing context-aware download task.

## [Google's Context](https://godoc.org/golang.org/x/net/context)
*  Context type, which carries deadlines, cancelation signals, and other request-scoped values across API boundaries and between processes.
*  References:
  * <https://godoc.org/golang.org/x/net/context>
  * <http://blog.golang.org/context> 

## Context-aware
*  ctxdownload.Download()'s first parameter is Context type.
  * Download() will stop if context's cancel() function is called in other goroutines.
  * Download() will stop if context's deadline exceeded.

## [Example](./ctxdownload_test.go)

    // Run "go test -c && ./ctxdownload.test"
    func ExampleDownload() {
        // Download a zip file to test ctxdownload.Download().
        url := "https://github.com/northbright/plants/archive/master.zip"
        outDir := "./download"
        fileName := "" // If file name is empty, it will try to detect file name in response Header in Download().

        // Set total timeout. You'll get "deadline exceeded" error as expected.
        // To make download successful, set it to 300 or longer.
        totalTimeoutSeconds := 30
        totalTimeout := time.Duration(time.Duration(totalTimeoutSeconds) * time.Second)

        // Make a context to carry a deadline(timeout).
        // See http://blog.golang.org/context to for more information.
        ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
        defer cancel()

        // HTTP request timeout(NOT download timeout).
        requestTimeoutSeconds := 5

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

## Go Version Requirement
* Go 1.5 and later
* See [Fix the *http.Request has no field or method Cancel Issue](https://github.com/northbright/Notes/blob/master/Golang/http/fix-the-http-request-has-no-field-or-method-cancel-issue.md)

## Documentation
* [API References](https://godoc.org/github.com/northbright/ctx/ctxdownload)

## License
* [MIT License](./LICENSE)

## References
* <https://godoc.org/golang.org/x/net/context>
* <http://blog.golang.org/context>
* <https://github.com/golang/net/tree/master/context>
