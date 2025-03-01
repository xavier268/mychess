package magic

import (
	"encoding/binary"
	"os"
)

const Header = ("MAGICMAP FORMAT v1.0")

// Save MagicMap content in binary format to file.
func (m MagicMap) Save(filename string) error {
	// Open file.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	err = binary.Write(file, binary.LittleEndian, Header)
	if err != nil {
		return err
	}
	// Write hard coded dimensions.
	binary.Write(file, binary.LittleEndian, NBKeys)
	binary.Write(file, binary.LittleEndian, NBValues)

	// Write magic map to file.
	err = binary.Write(file, binary.LittleEndian, m)
	if err != nil {
		return err
	}

	return nil
}
