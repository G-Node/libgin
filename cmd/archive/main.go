package main

import (
	"fmt"
	"log"
	"os"

	"github.com/G-Node/libgin/libgin/archive"
	"github.com/gogs/git-module"
)

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
	if len(args) != 3 {
		fmt.Printf("Usage: %s <type> <repository location>\n", args[0])
		fmt.Println("  <type> \"tar\" or \"zip\"")
		os.Exit(1)
	}

	archivetype := args[1]
	path := args[2]

	if !isDirectory(path) {
		fmt.Printf("%s does not appear to be a directory\n", path)
		os.Exit(1)
	}

	if !isRepository(path) {
		fmt.Printf("%s does not appear to be a git repository\n", path)
		os.Exit(1)
	}

	repo, err := git.OpenRepository(path)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	master, err := repo.GetCommit("master")
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	var writer archive.Writer
	switch archivetype {
	case "tar":
		writer = archive.NewTarArchive(repo, master)
	case "zip":
		writer = archive.NewZipArchive(repo, master)
	default:
		fmt.Printf("Invalid type %q. Specify either \"tar\" or \"zip\"", archivetype)
		os.Exit(1)
	}

	writer.Write(path)
}
