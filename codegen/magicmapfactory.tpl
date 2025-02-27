package position

import "fmt"

// Create a implementation for the specified nb of bits in input and output.
func NewMagicMap(in, out int) (mm MagicMap) {

	switch {
{{- range .}}
	case in == {{.IN}} && out == {{.OUT}}:
		return NewMagicMap_{{.IN}}_{{.OUT}}()
{{- end}}
	default:
		// Handle the case when no matching function exists
		panic(fmt.Sprintf("No magic map function exists for size=%d, index=%d", in, out))
	}
}

