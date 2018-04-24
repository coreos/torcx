# ctxcopy

ctxcopy is a [Golang](http://golang.org) package which provides helper functions for performing context-aware copy task.

## [Google's Context](https://godoc.org/golang.org/x/net/context)
*  Context type, which carries deadlines, cancelation signals, and other request-scoped values across API boundaries and between processes.
*  References:
  * <https://godoc.org/golang.org/x/net/context>
  * <http://blog.golang.org/context> 

## Context-aware
*  ctxcopy.Copy()'s first parameter is Context type.
  * Copy() will stop if context's cancel() function is called in other goroutines.
  * Copy() will stop if context's deadline exceeded.

## [Example](./ctxcopy_test.go)

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

## Go Version Requirement
* Go 1.5 and later
* See [Fix the *http.Request has no field or method Cancel Issue](https://github.com/northbright/Notes/blob/master/Golang/http/fix-the-http-request-has-no-field-or-method-cancel-issue.md) 

## Documentation
* [API References](https://godoc.org/github.com/northbright/ctx/ctxcopy)

## License
* [MIT License](./LICENSE)

## References
* <https://godoc.org/golang.org/x/net/context>
* <http://blog.golang.org/context>
* <https://github.com/golang/net/tree/master/context>
