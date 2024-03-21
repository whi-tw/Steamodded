package injector

import (
	"fmt"
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
	}
	//goland:noinspection GoErrorStringFormat
	return BalatroTypeUnknown, fmt.Errorf("Could not determine the type of Balatro executable. Ensure you have provided the correct path.")
}
