package magic

import (
	"encoding/gob"
	"fmt"
	"os"
)

// Signature of the magic format used (prime selection, NBKeys, NBValues) to identify incompatible changes in saved files.
// Uses a very basic hashing function.
func magicFormatSignature(data ...uint64) uint64 {
	var n uint64 = 29
	for i, d := range data {
		n = n*57119 + d*13 + uint64(i)*23 + 7
	}
	return n
}

// Header is always saved first in all data file, to ensure consitency.
var FileHeader = fmt.Sprintf("MAGICMAP FORMAT  %d", magicFormatSignature(NBKeys, NBValues))

// Save MagicMap to specified file. Directory should exist already.
func (m *MagicMap) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	// Save header
	if err = encoder.Encode(FileHeader); err != nil {
		return err
	}
	// Save MagicMap
	if err = encoder.Encode(m); err != nil {
		return err
	}
	fmt.Println("Successfully saved", filename)
	fmt.Println(FileHeader)

	return nil
}

// Load MagicMap from specified file.
func Load(filename string, m *MagicMap) error {

	if m == nil {
		return fmt.Errorf("cannot load from file : MagicMap pointer is nil")
	}

	// Open file.
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// load and check header
	decoder := gob.NewDecoder(file)
	var header string
	if err = decoder.Decode(&header); err != nil {
		return err
	}
	if header != FileHeader {
		return fmt.Errorf("unexpected file header : <%s>", header)
	}
	if err = decoder.Decode(m); err != nil {
		return err
	}
	fmt.Println("Successfully loaded", filename)
	fmt.Println(header)
	return nil
}
