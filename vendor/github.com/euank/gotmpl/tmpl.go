// Package gotmpl provides a simple library for stupid-simple templating.
// This templating is limited to variable substitution only. The only special characters are `\` and `$`.
// Each can be escaped with a backslash as `\\` and `\$` respectively.
// A valid variable reference must take the form of `${variable}` where `variable` matches /[a-zA-Z0-9_\-]/.
// Please see the examples and README for more details on usage of this library.
package gotmpl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

// Template parses any variables from a given reader and outputs processed data
// to the given writer. Variables are replaced based on the result of the
// provided lookup.
func Template(r io.Reader, w io.Writer, lookup Lookup) error {
	bufReader := bufio.NewReader(r)
	inTemplate := false
	varName := ""
	for {
		b, err := bufReader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if inTemplate {
			if b == '}' {
				inTemplate = false
				val, ok := lookup.Resolve(varName)
				if !ok {
					return errors.New("Could not resolve variable: " + varName)
				}
				w.Write([]byte(val))
			} else {
				varName += string(b)
			}
			continue
		}

		if b == '\\' {
			nb, err := bufReader.Peek(1)
			if err == io.EOF {
				w.Write([]byte{b})
				break
			} else if err != nil {
				return err
			}
			if nb[0] == byte('\\') {
				// \\ escape
				w.Write([]byte{b})
				bufReader.ReadByte()
				continue
			}
			if nb[0] == '$' {
				// \$ escape
				w.Write([]byte("$"))
				bufReader.ReadByte()
				continue
			}
		}

		if b == '$' {
			nb, err := bufReader.Peek(1)
			if err == io.EOF {
				w.Write([]byte{b})
				break
			} else if err != nil {
				return err
			}
			if nb[0] == '{' {
				inTemplate = true
				varName = ""
				bufReader.ReadByte()
				continue
			}
		}

		w.Write([]byte{b})
	}
	if inTemplate {
		return errors.New("unmatched open '{'")
	}
	return nil
}

// TemplateString is a convenience function to template a given input string in memory
func TemplateString(templateString string, lookup Lookup) (string, error) {
	var out bytes.Buffer
	err := Template(strings.NewReader(templateString), &out, lookup)
	return out.String(), err
}
