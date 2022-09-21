package util

import (
	"fmt"
	"os"
	"runtime/debug"
)

// FatalIf exits if the error is not nil
func FatalIf(err error) {
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Fatal error: %s\n", err)
		os.Exit(-1)
	}
}
