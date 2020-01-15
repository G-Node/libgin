package archive

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/G-Node/libgin/libgin/annex"
	"github.com/gogs/git-module"
)

type ZipArchive struct {
	Repository *git.Repository
	Commit     *git.Commit
	writer     *zip.Writer
}

func NewZipArchive(repo *git.Repository, commit *git.Commit) *ZipArchive {
	return &ZipArchive{Repository: repo, Commit: commit}
}

func (a *ZipArchive) Write(target string) error {
	tree := &a.Commit.Tree

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	a.writer = zip.NewWriter(zipfile)
	defer a.writer.Close()

	a.addTree(tree, "")

	return nil
}

func (a *ZipArchive) addBlob(blob *git.Blob, fname string) {
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
	if strings.Contains(string(readbuf[:32]), annex.IdentStr) {
		// replace with annexed data
		_, annexkey := filepath.Split(string(readbuf[:n]))
		annexkey = strings.TrimSpace(annexkey) // trim newlines and spaces
		loc, err := git.NewCommand("annex", "contentlocation", annexkey).RunInDir(a.Repository.Path)
		if err != nil {
			fmt.Printf("ERROR: couldn't find content file %q\n", annexkey)
			return
		}
		loc = strings.TrimRight(loc, "\n")
		loc = filepath.Join(a.Repository.Path, loc)
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
	writer, _ := a.writer.CreateHeader(&header)
	for ; n > 0 || err == nil; n, err = reader.Read(readbuf) {
		writer.Write(readbuf[:n])
	}
}

func (a *ZipArchive) addTree(tree *git.Tree, path string) {
	entries, _ := tree.ListEntries()
	for _, te := range entries {
		path := filepath.Join(path, te.Name())
		if te.IsDir() {
			subtree, _ := tree.SubTree(te.Name())
			a.addTree(subtree, path)
		} else {
			blob := te.Blob()
			a.addBlob(blob, path)
		}
	}
}
