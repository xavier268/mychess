package main

//go:generate go run magicmap.codegen.go

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// This code will generate the magic map structures for various bit configurations.

func main() {

	// Read the template file
	templfile, err := filepath.Abs("magicmap.tpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}
	tmplData, err := os.ReadFile(templfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading template: %v\n", err)
		os.Exit(1)
	}

	// Parse the template
	tmpl, err := template.New("magicmap").Parse(string(tmplData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	createSourceFile(tmpl, 10, 4)

}

func createSourceFile(tmpl *template.Template, inbit, outbit int) {

	// Execute the template with the configurations
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		IN  int
		OUT int
	}{inbit, outbit})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Write the generated code to magicmap_generated.go
	fname := fmt.Sprintf("magicmap_generated_%d_%d.go", inbit, outbit)
	fname = filepath.Join("..", "position", fname)
	err = os.WriteFile(fname, buf.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", fname, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s successfully!", fname)
}
