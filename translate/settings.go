package translate

var indent = "\t"

// Sets the translate package's indentation string. I cant imagine why you
// wouldn't use tabs or spaces, but technically this could be anything.
func SetIndent(i string) {
	indent = i
}
