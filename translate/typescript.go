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

func TypeFromIdent(i *ast.Ident) Type {
	ts, ok := goToTS[i.Name]
	if !ok {
		ts = "unknown"
	}

	return Type{
		Name: ts,
	}
}

func docToString(doc *string) string {
	var ret string
	if doc != nil {
		ret = "/**"
		for _, line := range strings.Split(*doc, "\n") {
			ret += " * " + line
		}
		ret += "\n */\n"
	}
	return ret
}

type Property struct {
	Name string
	Doc  *string
	Type Type
}

func (p Property) ToTypeScript() (string, error) {
	ret := docToString(p.Doc)
	t, err := p.Type.ToTypeScript()
	if err != nil {
		return "", fmt.Errorf("converting type of field '%s' to string: %w", p.Name, err)
	}
	return ret + p.Name + t, nil
}

type Structure struct {
	Name   string
	Doc    *string
	Fields []Property
}

func (s Structure) ToTypeScript() (string, error) {
	ret := docToString(s.Doc) + "interface " + s.Name + "{\n"
	for _, field := range s.Fields {
		f, err := field.ToTypeScript()
		if err != nil {
			return "", fmt.Errorf("converting field '%s' to string: %w", field.Name, err)
		}
		for _, line := range strings.Split(f, "\n") {
			ret += "\t" + line + "\n"
		}
	}
	return ret + "\n}", nil
}

var goToTS map[string]string = map[string]string{
	"int":    "number",
	"string": "string",
	"bool":   "boolean",
}

func PropertyFromField(f *ast.Field) Property {
	var doc *string
	if f.Doc != nil {
		doc = new(string)
		*doc = f.Doc.Text()
	}

	var t Type
	pt := &t
	var star ast.Expr = f.Type
loop:
	for {
		switch exp := star.(type) {
		case *ast.Ident:
			*pt = TypeFromIdent(exp)
			break loop
		case *ast.StarExpr:
			*&pt.Nullable = true
			star = exp.X
		case *ast.ArrayType:
			pt.Array = true
			pt.Inner = new(Type)
			pt = pt.Inner
			star = exp.Elt
		default:
			panic(fmt.Sprintf("not a type identifier: %+v", star))
		}
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
		Type: t,
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
