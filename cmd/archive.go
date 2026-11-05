package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExtractPKG(pluginArchive *string) (string, error) {

	tmpDir, err := os.MkdirTemp("", "caido-pkg-*")
	if err != nil {
		panic(err)
	}

	fmt.Println("Created temporary directory:", tmpDir)

	archive, err := zip.OpenReader(*pluginArchive)
	if err != nil {
		panic(err)
	}

	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(tmpDir, f.Name)
		fmt.Println("Unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(tmpDir)+string(os.PathSeparator)) {
			if err != nil {
				panic(err)
			}
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	// defer os.RemoveAll(tmpDir)
	base := filepath.Base(*pluginArchive)
	pluginName := strings.TrimSuffix(base, filepath.Ext(base))
	pluginPath := fmt.Sprintf("%s/%s", tmpDir, pluginName)
	return pluginPath, nil
}

func BuildPKG(pluginPath string) error {

	err := os.Chdir(pluginPath)
	if err != nil {
		fmt.Println("Error changing directory:", err)
		return err
	}

	cmd := exec.Command("pnpm", "build")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
