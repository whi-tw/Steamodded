package steamodded

import (
	"github.com/charmbracelet/log"
	"github.com/whi-tw/steamodded/pkg/injector"
	"os"
	"path/filepath"
)

func DoInjection(appPath string) error {
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Errorf("File %s does not exist", absPath)
		return err
	}

	balatroInjector, err := injector.NewInjector(absPath)
	if err != nil {
		return err
	}
	defer balatroInjector.Cleanup()

	err = balatroInjector.Inject()
	if err != nil {
		return err
	}

	log.Infof("Successfully injected mod loader into %s", appPath)
	return nil
}
