package injector

import (
	"fmt"
	"github.com/liamg/magic"
	"io"
	"os"
)

func whatKindOfBalatroIsThis(executablePath string) (BalatroType, error) {
	stat, err := os.Stat(executablePath)
	if err != nil {
		return BalatroTypeUnknown, err
	}

	if stat.IsDir() {
		if _, err := os.Stat(executablePath + "/Contents/Resources/Balatro.love"); err == nil {
			return BalatroTypeMacOS, nil
		}
		return BalatroTypeUnknown, fmt.Errorf("A directory was provided. Please provide the path to the Balatro executable.")
	}
	f, err := os.Open(executablePath)
	if err != nil {
		return BalatroTypeUnknown, err
	}
	defer f.Close()

	var header [512]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return BalatroTypeUnknown, err
	}

	fileType, err := magic.Lookup(header[:])
	if err != nil {
		return BalatroTypeUnknown, err
	}

	if fileType.Extension == "exe" {
		return BalatroTypeWindows, nil
	}

	//goland:noinspection GoErrorStringFormat
	return BalatroTypeUnknown, fmt.Errorf("Could not determine the type of Balatro executable. Ensure you have provided the correct path.")
}
