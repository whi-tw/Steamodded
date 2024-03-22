package injector

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type BalatroType int

const (
	BalatroTypeUnknown BalatroType = -1
	BalatroTypeWindows BalatroType = iota
	BalatroTypeLinux   BalatroType = iota
	BalatroTypeMacOS   BalatroType = iota
)

// Embed files into the executable
//
//go:embed steamodded patches
var filesToInject embed.FS

type SteamoddedInjector interface {
	Inject() error
	Cleanup() error
	extractSource() error
	updateGameLua() error
	getVersion() (string, error)
}

func NewInjector(executablePath string) (SteamoddedInjector, error) {
	balatroType, err := whatKindOfBalatroIsThis(executablePath)
	if err != nil {
		return nil, err
	}

	switch balatroType {
	case BalatroTypeWindows:
		log.Debug("Detected Balatro type is Windows")
		return newBalatroWindowsInjector(executablePath)
	case BalatroTypeMacOS:
		log.Debug("Detected Balatro type is MacOS")
		return newBalatroMacOSInjector(executablePath)
	default:
		//goland:noinspection GoErrorStringFormat
		return nil, fmt.Errorf("Could not determine balatro type.")
	}
}

type BalatroInjector struct {
	executablePath string
	extractPath    string
}

func (b BalatroInjector) extractSource() error {
	// should be implemented by the injector for the specific OS
	return fmt.Errorf("extractSource not implemented")
}

func (b BalatroInjector) Inject() error {
	//goland:noinspection GoErrorStringFormat
	return fmt.Errorf("Inject not implemented")
}

func (b BalatroInjector) Cleanup() error {
	err := os.RemoveAll(b.extractPath)
	if err != nil {
		return err
	}
	log.Debugf("Removed extract directory: %s", b.extractPath)
	return nil
}

func (b BalatroInjector) getVersion() (string, error) {
	versionFile, err := os.Open(filepath.Join(b.extractPath, "version.jkr"))
	if err != nil {
		return "", err
	}
	defer versionFile.Close()

	var versions []string
	// loop through the file and append each line to the versions slice
	for {
		var version string
		_, err := fmt.Fscanln(versionFile, &version)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read version file: %w", err)
		}
		versions = append(versions, version)
	}
	return versions[0], nil
}

func (b BalatroInjector) injectLuaFiles() error {
	mainLuaFilePath := filepath.Join(b.extractPath, "main.lua")
	// open the file
	f, err := os.OpenFile(mainLuaFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	injectedFiles := make([]string, 0)
	err = fs.WalkDir(filesToInject, "steamodded", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		sourceFile, err := filesToInject.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		// append the file content to the main.lua file
		_, err = io.Copy(writer, sourceFile)
		if err != nil {
			return err
		}
		_, err = writer.WriteString("\n")
		if err != nil {
			return err
		}

		injectedFiles = append(injectedFiles, path)

		return nil
	})
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (b BalatroInjector) updateGameLua() error {
	//gameLuaFilePath := filepath.Join(b.extractPath, "game.lua")

	version, err := b.getVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	patchPath, err := filesToInject.Open(fmt.Sprintf("patches/%s.patch", version))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Could not find a patch for Balatro version %s. Check if there is a newer version of the injector.", version)
		}
		return fmt.Errorf("failed to open patch file: %w", err)
	}
	defer patchPath.Close()

	files, _, err := gitdiff.Parse(patchPath)
	if err != nil {
		return fmt.Errorf("failed to parse patch file: %w", err)
	}

	gameLua, err := os.Open(filepath.Join(b.extractPath, "game.lua"))
	if err != nil {
		return err
	}
	defer gameLua.Close()

	var output bytes.Buffer
	if err := gitdiff.Apply(&output, gameLua, files[0]); err != nil {
		log.Debugf("Failed to apply patch: %s", err)
		if errors.Is(err, &gitdiff.Conflict{}) {
			return fmt.Errorf("Failed to inject steamodded start code. Is this game already patched?")
		}
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	gameLua, err = os.Create(filepath.Join(b.extractPath, "game.lua"))
	if err != nil {
		return err
	}
	defer gameLua.Close()

	if _, err := io.Copy(gameLua, &output); err != nil {
		return fmt.Errorf("failed to write patched game.lua: %w", err)
	}

	return nil
}
