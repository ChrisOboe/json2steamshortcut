package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/stephen-fox/steamutil/shortcuts"
)

func main() {
	var s []shortcuts.Shortcut
	err := json.NewDecoder(os.Stdin).Decode(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	err = shortcuts.WriteVdfV1(s, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
