package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"

	"github.com/ocket8888/tsapi/translate"
	"github.com/pborman/getopt/v2"
)

func GetStructNames(contents []byte) ([]translate.Structure, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", string(contents), 0)
	if err != nil {
		return nil, err
	}

	var structs []translate.Structure
	for _, decl := range f.Decls {

		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if s, err := translate.StructureFromType(ts); err != nil {
				return nil, fmt.Errorf("parsing file: %w", err)
			} else if s != nil {
				structs = append(structs, *s)
			}
		}

	}
	return structs, nil
}

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

	bts, err := io.ReadAll(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading '%s': %v\n", inFile.Name(), err)
		os.Exit(-2)
		return
	}

	structs, err := GetStructNames(bts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parsing '%s': %v\n", inFile.Name(), err)
		os.Exit(-3)
		return
	}

	var returnCode uint8 = 0
	for _, s := range structs {
		out, err := s.ToTypeScript(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "converting structure '%s' to a string: %v", s.Name, err)
			// "handles" overflow
			for returnCode == 0 {
				returnCode += 1
			}
			continue
		}
		fmt.Println(out)
		fmt.Println()
	}

	os.Exit(int(returnCode))
}
