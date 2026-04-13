// Package translate provides functionality for translating Go structures into
// TypeScript API types.
package translate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
)

type Package struct {
	Name       string
	Structures map[string]*Structure
}

func (p *Package) addStruct(name string) *Structure {
	if _, ok := p.Structures[name]; !ok {
		p.Structures[name] = &Structure{
			Name: name,
		}
	}
	return p.Structures[name]
}

// Translators are responsible for effecting translation. Zero-value Translators
// will break if you ask them to do anything; construct Translators using
// NewTranslator, please.
type Translator struct {
	// Indent is the string that will be used to indent scopes in the
	// translation.
	Indent string
	// UseTypes will instruct the Translator to use `type` instead of
	// `interface` in its output.
	UseTypes bool
	// Packages is the set of Go packages parsed by the Translator. If nothing
	// has been translated yet, then this will be empty if the Translator was
	// created by NewTranslator but could be nil if you just did it yourself.
	Packages map[string]*Package

	fileset *token.FileSet
}

// NewTranslator constructs a new Translator using the given indentation and a
// setting that determines if `type`s should be used instead of `interface`s.
func NewTranslator(indent string, useTypes bool) Translator {
	return Translator{
		Indent:   indent,
		UseTypes: useTypes,
		Packages: make(map[string]*Package, 1),
		fileset:  token.NewFileSet(),
	}
}

func (t *Translator) addPkg(name string) *Package {
	if _, ok := t.Packages[name]; !ok {
		t.Packages[name] = &Package{
			Name:       name,
			Structures: make(map[string]*Structure, 1),
		}
	}
	return t.Packages[name]
}

func (t *Translator) addStruct(ts *ast.TypeSpec, pkg *Package) error {
	if s, err := StructureFromType(ts); err != nil {
		return fmt.Errorf("parsing file: %w", err)
	} else if s != nil {
		strct := pkg.addStruct(s.Name)
		strct.Doc = s.Doc
		strct.Fields = s.Fields
	}
	return nil
}

// Translate adds the given file to the translator and parses it completely,
// adding to the Translator's Packages.
func (t *Translator) Translate(file io.Reader, filename string) error {
	bts, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read contents: %w", err)
	}

	f, err := parser.ParseFile(t.fileset, filename, string(bts), 0)
	if err != nil {
		return fmt.Errorf("failed to parse file '%s' contents: %w", filename, err)
	}

	pkg := t.addPkg(f.Name.String())

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
			t.addStruct(ts, pkg)
		}

	}

	return nil
}

// TranslateFile opens the given file and translates it. This is identical to
// using Translate but it does the file opening and whatnot for you.
func (t *Translator) TranslateFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", filename, err)
	}

	return t.Translate(f, filename)
}

func (t Translator) Write(file io.Writer) error {
	for _, pkg := range t.Packages {
		for _, s := range pkg.Structures {
			out, err := s.ToTypeScript(0)
			if err != nil {
				return fmt.Errorf("writing output for %s.%s: %w", pkg.Name, s.Name, err)
			}
			fmt.Fprintln(file, out)
			fmt.Fprintln(file)
		}
	}
	return nil
}
