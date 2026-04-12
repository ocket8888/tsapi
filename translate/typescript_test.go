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
