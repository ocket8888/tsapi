package main

import (
	"fmt"
	"os"

	"github.com/ocket8888/tsapi/translate"
	"github.com/pborman/getopt/v2"
)

func main() {
	var help bool
	getopt.FlagLong(&help, "help", 'h', "print usage information and then exit")
	getopt.ParseV2()
	args := getopt.Args()

	if help {
		getopt.PrintUsage(os.Stdout)
		return
	}

	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "invalid call signature\n")
		getopt.PrintUsage(os.Stderr)
		os.Exit(-1)
		return
	}

	inFile := os.Stdin
	if len(args) > 0 {
		var err error
		if inFile, err = os.Open(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "opening '%s': %v\n", args[0], err)
			os.Exit(-2)
			return
		}
	}

	t := translate.NewTranslator("\t", false)
	t.Translate(inFile, inFile.Name())
	if err := t.Write(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write results: %v\n", err)
		os.Exit(-3)
	}
}
