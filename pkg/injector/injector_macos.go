package injector

import (
	"archive/zip"
	"github.com/charmbracelet/log"
	"io"
	"os"
	"path/filepath"
)

type BalatroMacOSInjector struct {
	BalatroInjector
}

func newBalatroMacOSInjector(executablePath string) (*BalatroMacOSInjector, error) {
	extractPath, err := os.MkdirTemp("", "steamodded")
	if err != nil {
		return nil, err
	}
	return &BalatroMacOSInjector{
		BalatroInjector: BalatroInjector{
			executablePath: executablePath,
			extractPath:    extractPath,
		},
	}, nil
}

func (b BalatroMacOSInjector) Inject() error {
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
	log.Info("Injection done")
	return nil
}

func (b BalatroMacOSInjector) extractSource() error {
	archive, err := zip.OpenReader(b.executablePath + "/Contents/Resources/Balatro.love")
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		destPath := filepath.Join(b.extractPath, file.Name)
		log.Debugf("Extracting %s", destPath)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Open the file inside the archive
		err = func() error {
			archiveFile, err := file.Open()
			if err != nil {
				return err
			}
			defer archiveFile.Close()

			// Create the file on the filesystem
			err = func() error {
				destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
				if err != nil {
					return err
				}
				defer destFile.Close()

				// Copy the file from the archive to the filesystem
				if _, err := io.Copy(destFile, archiveFile); err != nil {
					return err
				}
				return nil
			}()
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b BalatroMacOSInjector) repackSource() error {
	archive, err := os.Create(b.executablePath + "/Contents/Resources/Balatro.love")
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	err = filepath.Walk(b.extractPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(b.extractPath, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
