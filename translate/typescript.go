package translate

import (
	"fmt"
	"go/ast"
	"strings"
)

type Type struct {
	Name     string
	Inner    *Type
	Array    bool
	Nullable bool
	Optional bool
}

func (t Type) ToTypeScript() (string, error) {
	ret := ": "
	if t.Optional {
		ret = "?" + ret
	}

	if t.Array {
		if t.Inner == nil {
			return "", fmt.Errorf("array with no inner type: %+v", t)
		}
		inner, err := t.Inner.ToTypeScript()
		if err != nil {
			return "", fmt.Errorf("inner type of %+v caused an error: %v", t, err)
		}
		ret += "Array<" + strings.TrimRight(strings.TrimLeft(inner, "?: "), ";") + ">"
	} else {
		ret += t.Name
	}

	if t.Nullable {
		ret += " | null"
	}

	return ret + ";", nil
}

type Property struct {
	Name string
	Doc  *string
	Type string
}

type Structure struct {
	Name   string
	Doc    *string
	Fields []Property
}

var goToTS map[string]string = map[string]string{
	"int":    "number",
	"string": "string",
}

func PropertyFromField(f *ast.Field) Property {
	var doc *string
	if f.Doc != nil {
		doc = new(string)
		*doc = f.Doc.Text()
	}

	var i *ast.Ident
	var star ast.Expr = f.Type
	for i == nil {
		fmt.Printf("%T\n", star)
		switch exp := star.(type) {
		case *ast.Ident:
			i = exp
		case *ast.StarExpr:
			star = exp.X
		case *ast.ArrayType:
			star = exp.Elt
		default:
			panic(fmt.Sprintf("not a type identifier: %s", f.Type))
		}
	}

	ts, ok := goToTS[i.Name]
	if !ok {
		ts = "unknown"
	}

	var name string
	for _, n := range f.Names {
		if len(name) > 0 {
			panic(fmt.Sprintf("multiple names found: %s", f.Names))
		}
		name = n.Name
	}
	if len(name) < 1 {
		panic("field with no name")
	}
	return Property{
		Name: name,
		Doc:  doc,
		Type: ts,
	}
}

func StructureFromType(t *ast.TypeSpec) *Structure {
	s, ok := t.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	name := t.Name.Name
	fields := make([]Property, s.Fields.NumFields())
	for i, field := range s.Fields.List {
		fields[i] = PropertyFromField(field)
	}

	var doc *string
	if t.Doc != nil {
		doc = new(string)
		*doc = t.Doc.Text()
	}

	return &Structure{
		Name:   name,
		Doc:    doc,
		Fields: fields,
	}
}
