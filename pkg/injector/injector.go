package injector

import (
	"bufio"
	"crypto/md5"
	"embed"
	"fmt"
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

// Embed the lua files into the executable
//
//go:embed steamodded
var filesToInject embed.FS

type SteamoddedInjector interface {
	Inject() error
	Cleanup() error
	extractSource() error
	updateGameLua() error
	getVersion() ([]string, error)
}

func NewInjector(executablePath string) (SteamoddedInjector, error) {
	balatroType, err := whatKindOfBalatroIsThis(executablePath)
	if err != nil {
		return nil, err
	}

	switch balatroType {
	case BalatroTypeWindows:
		log.Debug("Detected Balatro type is Windows")
	case BalatroTypeMacOS:
		log.Debug("Detected Balatro type is MacOS")
		return newBalatroMacOSInjector(executablePath)
	default:
		//goland:noinspection GoErrorStringFormat
		return nil, fmt.Errorf("Could not determine balatro type.")
	}
	return nil, fmt.Errorf("weird")
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

func (b BalatroInjector) getVersion() ([]string, error) {
	versionFile, err := os.Open(filepath.Join(b.extractPath, "version.jkr"))
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("failed to read version file: %w", err)
		}
		versions = append(versions, version)
	}
	return versions, nil
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
		//destPath := filepath.Join(b.extractPath, path)
		//log.Debugf("Injecting %s", destPath)

		sourceFile, err := filesToInject.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		//// if the file is in a directory, create the directory
		//if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		//	return err
		//}

		// Create the file
		//f, err := os.Create(destPath)
		//if err != nil {
		//	return err
		//}
		//defer f.Close()

		// Copy the file
		//_, err = io.Copy(f, sourceFile)
		//if err != nil {
		//	return err
		//}

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

	//mainLuaFilePath := filepath.Join(b.extractPath, "main.lua")
	//// open the file
	//f, err := os.OpenFile(mainLuaFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//if err != nil {
	//	return err
	//}
	//defer f.Close()
	//
	//// write out 'dofile' statements for each injected file at the end of the main.lua file
	//writer := bufio.NewWriter(f)
	//for _, file := range injectedFiles {
	//	_, err := writer.WriteString(fmt.Sprintf("love.filesystem.load(\"%s\")()\n", file))
	//	if err != nil {
	//		return err
	//	}
	//}
	//return writer.Flush()
}

func (b BalatroInjector) updateGameLua() error {
	gameLuaFilePath := filepath.Join(b.extractPath, "game.lua")

	h := md5.New()

	f, err := os.Open(gameLuaFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	gameLuaHash := fmt.Sprintf("%x", h.Sum(nil))
	log.Debugf("game.lua hash: %s", gameLuaHash)

	// retrieve the injection point from gameLuaInjectionPoints map[string]InjectionPoint
	injectionPoint, ok := gameLuaInjectionPoints[gameLuaHash]
	if !ok {
		return fmt.Errorf("could not find injection point for game.lua with hash %s", gameLuaHash)
	}

	// open the file
	f, err = os.Open(gameLuaFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// read the file line by line
	originalLines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		originalLines = append(originalLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	var indentedStartCode string
	for i := 0; i < injectionPoint.Indentation; i++ {
		indentedStartCode += " "
	}
	indentedStartCode += startCode

	// insert the startCode at the correct line and column
	newLines := originalLines[:injectionPoint.Line-1]
	newLines = append(newLines, indentedStartCode)
	newLines = append(newLines, originalLines[injectionPoint.Line:]...)

	// write the modified originalLines back to the file
	f, err = os.Create(gameLuaFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, line := range newLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
