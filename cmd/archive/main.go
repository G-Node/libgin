package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gogs/git-module"
)

func readAnnexedStubs(archivepath string) ([]string, error) {
	archiverc, err := zip.OpenReader(archivepath)
	if err != nil {
		return nil, err
	}

	defer archiverc.Close()

	for _, file := range archiverc.File {
		info := file.FileInfo()
		if info.IsDir() {
			continue
		}

		// if it's a file (or a symlink), read the contents and check if it's
		// an annexed object path
		filerc, _ := file.Open()
		data := make([]byte, info.Size())
		n, err := io.ReadFull(filerc, data)
		if err != nil {
			fmt.Printf("  Error reading link %q: %s\n", info.Name(), err.Error())
			continue
		}
		fmt.Printf("  Read %d bytes\n", n)
		if n < 1024 {
			fmt.Println(string(data))
		}

		// or if it contains the 'annex/objects' prefix
	}
	return []string{}, nil
}

func ginarchive(path string) error {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return err
	}

	master, err := repo.GetCommit("master")
	if err != nil {
		return err
	}

	// 1. Create git archive
	fname := master.ID.String()[:6] + ".zip"
	archivepath := filepath.Join("/home", "achilleas", "tmp", fname)
	fmt.Printf("Archiving repository at %s to %s\n", path, archivepath)
	if err := master.CreateArchive(archivepath, git.ZIP); err != nil {
		return err
	}

	// 2. Identify annexed files
	stubs, err := readAnnexedStubs(archivepath)
	if err != nil {
		return err
	}

	for _, fname := range stubs {
		fmt.Println(fname)
	}

	// 3. Update git archive with annexed content

	return nil
}

func isDirectory(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}

	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get info for %s: %s", path, err.Error())
	}

	return fileinfo.IsDir()
}

func isRepository(path string) bool {
	_, err := git.NewCommand("rev-parse").RunInDir(path)
	return err == nil
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("Usage: %s <repository location>\n", args[0])
		os.Exit(1)
	}

	path := args[1]
	if !isDirectory(path) {
		fmt.Printf("%s does not appear to be a directory\n", path)
		os.Exit(1)
	}

	if !isRepository(path) {
		fmt.Printf("%s does not appear to be a git repository\n", path)
		os.Exit(1)
	}

	err := ginarchive(path)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
}
