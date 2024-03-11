package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"

	"github.com/stephen-fox/steamutil/shortcuts"
)

func hash(s string) int {
	h := fnv.New32()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func main() {
	var s []shortcuts.Shortcut
	err := json.NewDecoder(os.Stdin).Decode(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// generate reproducable id if it's not explicitely set
	for i, shortcut := range s {
		if shortcut.Id == 0 {
			s[i].Id = hash(shortcut.AppName)
		}
	}

	err = shortcuts.WriteVdfV1(s, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
