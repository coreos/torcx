package ctxdownload

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/northbright/ctx/ctxcopy"
	"github.com/northbright/httputil"
	"github.com/northbright/pathhelper"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const (
	// defRequestTimeoutSeconds is default HTTP request timeout(seconds). It's NOT download timeout.
	defRequestTimeoutSeconds int = 10
)

// Download downloads the file from HTTP server.
//
//   Params:
//     ctx:
//       Google's Context type which carries deadlines, cacelation signals,
//       and other request-scoped values across API boundaries and between processes.
//       See https://godoc.org/golang.org/x/net/context for more.
//     url:
//       Download URL
//     outDir:
//       Directory to store the downloaded file.
//     fileName:
//       Downloaded file name(base name). If the given file name is empty(""), it'll detect the file name in response header.
//     buf:
//       Buffer(length should >= 0).
//     requestTimeoutSeconds:
//       HTTP request timeout. It's NOT download timeout. Default value(defRequestTimeoutSeconds) is 10 seconds.
//
//   Return:
//     downloadedFileName: Absolute downloaded file path.
func Download(ctx context.Context, url, outDir, fileName string, buf []byte, requestTimeoutSeconds int) (downloadedFileName string, err error) {
	// Get absolute path of out dir.
	// It'll join the directory of current executable and input path if it's relative.
	absOutDir := ""
	if absOutDir, err = pathhelper.GetAbsPath(outDir); err != nil {
		return "", err
	}

	// Make out dir
	if err = os.MkdirAll(absOutDir, 0755); err != nil {
		return "", err
	}

	var (
		newCtx context.Context
		cancel context.CancelFunc
	)

	// Check request timeout
	if requestTimeoutSeconds <= 0 {
		requestTimeoutSeconds = defRequestTimeoutSeconds
	}

	// Derive new context with request timeout.
	reqTimeout := time.Duration(time.Duration(requestTimeoutSeconds) * time.Second)
	if ctx != nil {
		newCtx, cancel = context.WithTimeout(ctx, reqTimeout)
	} else {
		newCtx, cancel = context.WithTimeout(context.Background(), reqTimeout)
	}
	defer cancel()

	// Do HTTP Request by using golang.org/x/net/context/ctxhttp
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Accept Encoding:gzip
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := ctxhttp.Do(newCtx, nil, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read from response.Body or gzip reader
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Try to get file name in response header
	if fileName == "" {
		if fileName, err = httputil.GetFileName(url); err != nil {
			return "", err
		}
	}

	// Get final absolute file name.
	absFileName := path.Join(absOutDir, fileName)

	// Create a new file.
	f, err := os.Create(absFileName)
	if err != nil {
		return "", err
	}

	// Call ctxcopy.Copy() to perform Context-aware copy task.
	err = ctxcopy.Copy(ctx, f, reader, buf)
	f.Close()

	if err != nil {
		os.Remove(absFileName)
		return "", err
	}

	return absFileName, nil
}
