package main

import (
	"encoding/json"
	"flag"
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
	var filepath string
	flag.StringVar(&filepath, "file", "-", "the path to the json. when - is used it's read from stdin")

	flag.Parse()

	stream := os.Stdin

	if filepath != "-" {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s: %s\n", filepath, err.Error())
			os.Exit(1)
		}
		defer file.Close()
		stream = file
	}

	var s []shortcuts.Shortcut
	err := json.NewDecoder(stream).Decode(&s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse json: %s\n", err.Error())
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
