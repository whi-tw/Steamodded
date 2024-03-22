package injector

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/whi-tw/steamodded/pkg/seven_zip"
	"os"
	"os/exec"
)

type BalatroWindowsInjector struct {
	BalatroInjector
	sevenZip seven_zip.SevenZipHelper
}

func newBalatroWindowsInjector(executablePath string) (*BalatroWindowsInjector, error) {
	sevenZip, err := seven_zip.NewSevenZip()
	if err != nil {
		return nil, err
	}
	log.Debugf("Using 7z at %s", sevenZip.GetPath())
	extractPath, err := os.MkdirTemp("", "steamodded")
	if err != nil {
		return nil, err
	}
	return &BalatroWindowsInjector{
		BalatroInjector: BalatroInjector{
			executablePath: executablePath,
			extractPath:    extractPath,
		},
		sevenZip: sevenZip,
	}, nil
}

func (b BalatroWindowsInjector) Inject() error {
	log.Info("Injecting mod loader into Balatro executable")
	log.Debugf("Executable path: %s", b.executablePath)
	log.Debugf("Extract Directory: %s", b.extractPath)

	log.Info("Extracting...")
	if err := b.extractSource(); err != nil {
		return err
	}

	log.Info("Extracting done")

	log.Info("Injecting steamodded...")
	if err := b.injectLuaFiles(); err != nil {
		return err
	}
	if err := b.updateGameLua(); err != nil {
		return err
	}
	if err := b.repackSource(); err != nil {
		return err
	}
	if err := b.sevenZip.Cleanup(); err != nil {
		return fmt.Errorf("error cleaning up: %w", err)
	}
	log.Info("Injection done")
	return nil
}

func (b BalatroWindowsInjector) extractSource() error {
	cmd := exec.Command(b.sevenZip.GetPath(), "x", b.executablePath, "-o"+b.extractPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}

func (b BalatroWindowsInjector) repackSource() error {
	cmd := exec.Command(b.sevenZip.GetPath(), "a", "-tzip", b.executablePath, b.extractPath+"/*")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}
