package seven_zip

import (
	"archive/zip"
	"fmt"
	"github.com/charmbracelet/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type SevenZipHelper interface {
	GetPath() string
	Cleanup() error
}

type SevenZip struct {
	path        string
	extractPath string
}

func (s SevenZip) GetPath() string {
	return s.path
}

func (s SevenZip) Cleanup() error {
	if s.extractPath != "" {
		return os.RemoveAll(s.extractPath)
	}
	log.Debug("Didn't extract 7z, nothing to clean up")
	return nil
}

func NewSevenZip() (SevenZipHelper, error) {
	switch runtime.GOOS {
	case "windows":
		extractPath, err := os.MkdirTemp("", "steamodded-7z")
		if err != nil {
			return nil, err
		}
		path, err := extractEmbeddedSevenZip(extractPath)
		if err != nil {
			return nil, err
		}
		return SevenZipHelper(&SevenZip{path: path, extractPath: extractPath}), nil
	default:
		path, err := locateSevenZip()
		if err != nil {
			return nil, err
		}
		return SevenZipHelper(&SevenZip{path: path}), nil
	}
}

func locateSevenZip() (string, error) {
	fname, err := exec.LookPath("7zz")
	if err != nil {
		return "", fmt.Errorf("7zz not found in PATH. Is it installed? %w", err)
	}
	return filepath.Abs(fname)
}

func extractEmbeddedSevenZip(extractPath string) (string, error) {
	sevenZipArchivePath := filepath.Join(extractPath, "7z-repack.zip")
	f, err := os.Create(sevenZipArchivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(sevenZipBinary); err != nil {
		return "", err
	}

	archive, err := zip.OpenReader(sevenZipArchivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	for _, file := range archive.File {
		destPath := filepath.Join(extractPath, file.Name)
		log.Debugf("Extracting %s", destPath)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return "", err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return "", err
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
			return "", err
		}
	}
	return filepath.Join(extractPath, "7z.exe"), nil
}
