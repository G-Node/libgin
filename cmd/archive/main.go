package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogs/git-module"
)

const (
	annexident = "/annex/objects"
)

func addBlob(zipWriter *zip.Writer, blob *git.Blob, fname string, repopath string) {
	var filemode os.FileMode
	filemode |= 0660
	if blob.IsLink() {
		filemode |= os.ModeSymlink
	}
	header := zip.FileHeader{
		Name:     filepath.Join(fname),
		Modified: time.Now(), // TODO: use commit time
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	blob.DataPipeline(stdout, stderr)
	reader := bufio.NewReader(stdout)
	readbuf := make([]byte, 10240)
	// check first 32 bytes for annex path identifier
	n, err := reader.Read(readbuf)
	if strings.Contains(string(readbuf[:32]), annexident) {
		// replace with annexed data
		_, annexkey := filepath.Split(string(readbuf[:n]))
		annexkey = strings.TrimSpace(annexkey) // trim newlines and spaces
		loc, err := git.NewCommand("annex", "contentlocation", annexkey).RunInDir(repopath)
		if err != nil {
			fmt.Printf("ERROR: couldn't find content file %q\n", annexkey)
			return
		}
		loc = strings.TrimRight(loc, "\n")
		loc = filepath.Join(repopath, loc)
		rc, err := os.Open(loc)
		if err != nil {
			fmt.Printf("Failed to read: %s\n", err.Error())
			return
		}
		// copy mode from content file
		rcInfo, _ := rc.Stat()
		filemode = rcInfo.Mode()
		n, err = reader.Read(readbuf)
		reader = bufio.NewReader(rc)
	}
	header.SetMode(filemode)
	writer, _ := zipWriter.CreateHeader(&header)
	for ; n > 0 || err == nil; n, err = reader.Read(readbuf) {
		if err != nil {
			fmt.Printf("ERROR reading %s\n", fname)
			break
		}
		writer.Write(readbuf[:n])
	}
}

func zipTree(zipWriter *zip.Writer, t *git.Tree, prefix string, repopath string) {
	entries, _ := t.ListEntries()
	for _, te := range entries {
		fname := filepath.Join(prefix, te.Name())
		if te.IsDir() {
			subtree, _ := t.SubTree(te.Name())
			zipTree(zipWriter, subtree, fname, repopath)
		} else {
			blob := te.Blob()
			addBlob(zipWriter, blob, fname, repopath)
		}
	}
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

	fname := master.ID.String()[:6] + ".zip"

	// place archive in repository's parent directory
	archivepath, _ := filepath.Abs(filepath.Join(path, "..", fname))

	tree := &master.Tree

	zipfile, _ := os.Create(archivepath)
	zipWriter := zip.NewWriter(zipfile)

	defer zipfile.Close()
	defer zipWriter.Close()

	zipTree(zipWriter, tree, "", path)

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
