package lc3_test

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/prichrd/lc3/pkg/lc3"
)

func TestReadImageFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "image")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpFile.Name())
	binary.Write(tmpFile, binary.BigEndian, []uint16{0x3000, 0xFFFF, 0xFFFF})
	if err := tmpFile.Close(); err != nil {
		t.Error(err)
	}

	b, err := lc3.ReadImageFile(tmpFile.Name())
	if err != nil {
		t.Errorf("want no error, but got %v", err)
	}

	for i := 0; i < 65536; i++ {
		if i == 0x3000 || i == 0x3001 {
			if b[i] != 0xFFFF {
				t.Errorf("want index %x to be xffff, but got %x", i, b[i])
			}
			continue
		}
		if b[i] != 0 {
			t.Errorf("want index %x to be 0, but got %x", i, b[i])
		}
	}
}
