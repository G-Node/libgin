package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogs/git-module"
)

const (
	annexident = "/annex/objects"
)

func readAnnexedStubs(archivepath string) (map[string]string, error) {
	archiverc, err := zip.OpenReader(archivepath)
	if err != nil {
		return nil, err
	}

	defer archiverc.Close()

	annexContentMap := make(map[string]string)
	for _, file := range archiverc.File {
		info := file.FileInfo()
		if info.IsDir() {
			continue
		}

		// if it's a file (or a symlink), read the contents and check if it's
		// an annexed object path
		filerc, _ := file.Open()
		data := make([]byte, 1024)
		n, _ := io.ReadFull(filerc, data)
		data = data[:n]
		if strings.Contains(string(data), annexident) {
			_, key := filepath.Split(string(data))
			// git content of unlocked pointer files have newline at the end so
			// we should trim space
			key = strings.TrimSpace(key) // trim newlines and spaces
			annexContentMap[file.Name] = key
		}
	}
	return annexContentMap, nil
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
	annexContent, err := readAnnexedStubs(archivepath)
	if err != nil {
		return err
	}

	for fname, annexkey := range annexContent {
		fmt.Printf("%q -> %q\n", fname, annexkey)
		loc, err := git.NewCommand("annex", "contentlocation", annexkey).RunInDir(path)
		loc = strings.TrimRight(loc, "\n")
		if err != nil {
			fmt.Printf("ERROR: couldn't find content for file %q (%s)\n", fname, annexkey)
			continue
		}

		fmt.Printf("%q will be replaced with %q\n", fname, loc)
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
