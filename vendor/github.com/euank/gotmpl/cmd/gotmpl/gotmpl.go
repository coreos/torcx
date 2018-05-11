package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/euank/gotmpl"
)

func main() {
	env := flag.Bool("env", true, "Pull variables from the environment")
	inplace := flag.Bool("inplace", false, "Replace variables in the given file inplace")
	flag.Parse()
	remainingArgs := flag.Args()

	var tmplReader io.Reader
	var outWriter io.WriteCloser = os.Stdout
	defer outWriter.Close()

	if shouldReadStdin() {
		if *inplace {
			log.Fatal("Cannot do inplace replacement of stdin")
		}

		tmplReader = os.Stdin
	} else {
		if len(remainingArgs) == 0 {
			log.Fatal("Must provide an argument of a file to template")
		}
		fileName := remainingArgs[len(remainingArgs)-1]
		lastFile, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("Could not open given file (%v) for templating: %v", fileName, err)
		}
		tmplReader = lastFile
		if *inplace {
			outWriter = &bufferedFileWriter{file: fileName, Buffer: bytes.NewBuffer([]byte{})}
		}
		remainingArgs = remainingArgs[0 : len(remainingArgs)-1]
	}

	resolvers := chainResolver{}

	if *env {
		resolvers = append(resolvers, envLookup{})
	}

	vars := make(map[string]interface{})
	for _, arg := range remainingArgs {
		avar := make(map[string]interface{})
		f, err := os.Open(arg)
		if err != nil {
			log.Fatal("Unable to open file: " + arg)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal("Unable to read file: " + arg)
		}

		err = json.Unmarshal(data, &avar)
		if err != nil {
			log.Fatalf("Invalid json file %v: %v", arg, err)
		}

		for k, v := range avar {
			vars[k] = v
		}
	}

	strVars := make(gotmpl.MapLookup)
	for k, v := range vars {
		if v == nil {
			strVars[k] = ""
		} else {
			strVars[k] = fmt.Sprintf("%v", v)
		}
	}

	resolvers = append(resolvers, strVars)

	if err := gotmpl.Template(tmplReader, outWriter, resolvers); err != nil {
		log.Fatal(err)
	}
}

// bufferedFileWriter is an io.WriteCloser which buffers all 'Write' calls into
// memory, and then flushes them to the provided file on 'Close'.
type bufferedFileWriter struct {
	file string
	*bytes.Buffer
}

func (b *bufferedFileWriter) Close() error {
	return ioutil.WriteFile(b.file, b.Bytes(), 0600)
}

// chainResolver checks for a variable in each element of the chain in order
type chainResolver []gotmpl.Lookup

func (c chainResolver) Resolve(variable string) (string, bool) {
	for _, l := range c {
		if s, ok := l.Resolve(variable); ok {
			return s, ok
		}
	}
	return "", false
}

// envLookup implements a gotmpl.Lookup sources from the environment
type envLookup struct{}

func (envLookup) Resolve(variable string) (string, bool) {
	return os.LookupEnv(variable)
}

// shouldReadStdin determines if stdin should be considered a valid source of data for templating.
func shouldReadStdin() bool {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	return stdinStat.Mode()&os.ModeCharDevice == 0
}
