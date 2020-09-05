package lc3

import (
	"encoding/binary"
	"io"
	"math"
	"os"
)

// ReadImageFile reads a file into memory.
func ReadImageFile(filename string) ([65536]uint16, error) {
	mem := [65536]uint16{}

	file, err := os.Open(filename)
	if err != nil {
		return mem, err
	}
	defer file.Close()

	var startingPC uint16
	if err := binary.Read(file, binary.BigEndian, &startingPC); err != nil {
		return mem, err
	}

	for i := startingPC; i < math.MaxUint16; i++ {
		var val uint16
		if err := binary.Read(file, binary.BigEndian, &val); err != nil {
			if err == io.EOF {
				break
			}
			return mem, err
		}
		mem[i] = val
	}

	return mem, err
}
