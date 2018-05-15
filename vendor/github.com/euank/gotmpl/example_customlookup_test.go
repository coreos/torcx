package gotmpl_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/euank/gotmpl"
)

type fibLookup struct{}

func (fibLookup) Resolve(variable string) (string, bool) {
	if strings.HasPrefix(variable, "fib ") {
		parts := strings.Split(variable, " ")
		if len(parts) != 2 {
			return "", false
		}
		fibNum, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", false
		}
		x, y, i := 1, 0, 0
		for ; i < fibNum; i++ {
			x, y = y, x+y
		}
		return strconv.Itoa(x), true
	}
	return "", false
}

func ExampleTemplate_customLookup() {
	input := `Have a fib sequence! ${fib 1}, ${fib 2}, ${fib 3}, ${fib 4}, ${fib 5}, ${fib 6}, ${fib 7}.. and the 20th is ${fib 20}`
	output, err := gotmpl.TemplateString(input, fibLookup{})
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
	// Output: Have a fib sequence! 0, 1, 1, 2, 3, 5, 8.. and the 20th is 4181
}
