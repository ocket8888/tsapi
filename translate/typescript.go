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

func docToString(doc *string, indent string) string {
	var ret string
	if doc != nil {
		ret = indent + "/**\n"
		for _, line := range strings.Split(*doc, "\n") {
			ret += indent + " * " + line
		}
		ret += "\n" + indent + " */\n"
	}
	return ret
}

type Property struct {
	Name string
	Doc  *string
	Type Type
}

func (p Property) ToTypeScript(indentLevel uint) (string, error) {
	ident := ""
	for range indentLevel {
		ident += indent
	}

	ret := docToString(p.Doc, ident)
	t, err := p.Type.ToTypeScript()
	if err != nil {
		return "", fmt.Errorf("converting type of field '%s' to string: %w", p.Name, err)
	}
	return ret + ident + p.Name + t, nil
}

type Structure struct {
	Name   string
	Doc    *string
	Fields []Property
}

func (s Structure) ToTypeScript(indentLevel uint) (string, error) {
	ident := ""
	for range indentLevel {
		ident += indent
	}
	var ret strings.Builder
	ret.WriteString(docToString(s.Doc, ident) + ident + "interface " + s.Name + " {")
	for _, field := range s.Fields {
		ret.WriteRune('\n')
		f, err := field.ToTypeScript(indentLevel + 1)
		if err != nil {
			return "", fmt.Errorf("converting field '%s' to string: %w", field.Name, err)
		}
		ret.WriteString(f)
	}
	return ret.String() + ident + "\n}", nil
}

var goToTS map[string]string = map[string]string{
	"int":     "number",
	"int8":    "number",
	"int16":   "number",
	"int32":   "number",
	"int64":   "number",
	"uint":    "number",
	"uint8":   "number",
	"uint16":  "number",
	"uint32":  "number",
	"uint64":  "number",
	"float32": "number",
	"float64": "number",
	"string":  "string",
	"bool":    "boolean",
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
			pt.Name = TypeFromIdent(exp).Name
			break loop
		case *ast.StarExpr:
			pt.Nullable = true
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
