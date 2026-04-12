package translate

import (
	"fmt"
	"testing"
)

func ExampleType_ToTypeScript() {
	t := Type{
		Inner: &Type{
			Name:     "string",
			Nullable: true,
		},
		Array:    true,
		Nullable: true,
		Optional: true,
	}

	repr, _ := t.ToTypeScript()
	fmt.Println(repr)
	// Output: ?: Array<string | null> | null;
}

func equal[T comparable](t *testing.T, value, expected T) bool {
	if value != expected {
		t.Errorf("expected: '%v', got: '%v'", expected, value)
		return false
	}
	return true
}

func TestType_ToTypeScript_Simple(t *testing.T) {
	typ := Type{
		Name: "string",
	}

	out, err := typ.ToTypeScript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !equal(t, out, ": string;") {
		t.FailNow()
	}
}

func TestType_ToTypeScript_Nullable(t *testing.T) {
	typ := Type{
		Name:     "string",
		Nullable: true,
	}

	out, err := typ.ToTypeScript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !equal(t, out, ": string | null;") {
		t.FailNow()
	}
}

func TestType_ToTypeScript_NestedArrays(t *testing.T) {
	typ := Type{
		Name:  "string",
		Array: true,
		Inner: &Type{
			Name:     "bigint lmao",
			Array:    true,
			Nullable: true,
			Optional: true,
			Inner: &Type{
				Array: true,
				Inner: &Type{
					Name: "number",
				},
			},
		},
	}

	out, err := typ.ToTypeScript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !equal(t, out, ": Array<Array<Array<number>> | null>;") {
		t.FailNow()
	}
}

func TestType_ToTypeScript_NoInnerArrayType(t *testing.T) {
	typ := Type{
		Array: true,
	}

	_, err := typ.ToTypeScript()
	if err == nil {
		t.Fatal("expected an error; didn't get one")
	}
	t.Logf("received expected error: %v", err)
}

func TestType_ToTypeScript_InnerError(t *testing.T) {
	typ := Type{
		Array: true,
		Inner: &Type{
			Array: true,
		},
	}

	_, err := typ.ToTypeScript()
	if err == nil {
		t.Fatal("expected an error; didn't get one")
	}
	t.Logf("received expected error: %v", err)
}

func ExampleProperty_ToTypeScript() {
	p := Property{
		Name: "foo",
		Type: Type{
			Name:     "string",
			Optional: true,
			Nullable: true,
		},
	}

	out, _ := p.ToTypeScript(0)
	fmt.Println(out)
	// Output: foo?: string | null;
}

func TestProperty_ToTypeScript_indentation(t *testing.T) {
	p := Property{
		Name: "foo",
		Type: Type{
			Name:     "string",
			Optional: true,
			Nullable: true,
		},
	}

	SetIndent("  ")
	out, err := p.ToTypeScript(3)
	if err != nil {
		t.Fatalf("failed to convert property to TypeScript: %v", err)
	}
	if !equal(t, out, "      foo?: string | null;") {
		t.FailNow()
	}
}

func ExampleStructure_ToTypeScript() {
	doc := "Foo is an example struct."
	s := Structure{
		Name: "Foo",
		Doc:  &doc,
		Fields: []Property{
			{
				Name: "Bar",
				Type: Type{
					Array:    true,
					Optional: true,
					Inner: &Type{
						Name:     "string",
						Nullable: true,
					},
				},
			},
			{
				Name: "Test",
				Type: Type{
					Name:     "number",
					Nullable: true,
				},
			},
		},
	}

	SetIndent("  ")
	out, _ := s.ToTypeScript(0)
	fmt.Println(out)
	// Output: /**
	//  * Foo is an example struct.
	//  */
	// interface Foo {
	//   Bar?: Array<string | null>;
	//   Test: number | null;
	// }
}
