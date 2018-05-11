package gotmpl_test

import (
	"os"
	"strings"

	"github.com/euank/gotmpl"
)

func ExampleTemplate() {
	r := strings.NewReader(`Hello ${name}`)
	gotmpl.Template(r, os.Stdout, gotmpl.MapLookup(map[string]string{"name": "uhh, Jim"}))
	// Output: Hello uhh, Jim
}
