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

type ZipWriter struct {
	Repository *git.Repository
	Commit     *git.Commit
	writer     *zip.Writer
}

func NewZipWriter(repo *git.Repository, commit *git.Commit) *ZipWriter {
	return &ZipWriter{Repository: repo, Commit: commit}
}

func (a *ZipWriter) Write(target string) error {
	tree := a.Commit.Tree

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	a.writer = zip.NewWriter(zipfile)
	defer a.writer.Close()

	return a.addTree(tree, "")
}

func (a *ZipWriter) addBlob(blob *git.Blob, fname string) error {
	var filemode os.FileMode
	filemode |= 0660
	if blob.IsSymlink() {
		filemode |= os.ModeSymlink
	}
	header := zip.FileHeader{
		Name:     fname,
		Modified: time.Now(), // TODO: use commit time
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	blob.Pipeline(stdout, stderr)
	reader := bufio.NewReader(stdout)
	readbuf := make([]byte, 10240)
	if annex.IsAnnexFile(blob) {
		n, err := reader.Read(readbuf)
		if err != nil {
			return fmt.Errorf("failed to read annex pointer blob %q: %s", fname, err.Error())
		}
		// replace with annexed data
		_, annexkey := filepath.Split(string(readbuf[:n]))
		annexkey = strings.TrimSpace(annexkey) // trim newlines and spaces
		loc, err := annex.ContentLocation(a.Repository, annexkey)
		if err != nil {
			return fmt.Errorf("content file not found: %q", annexkey)
		}
		loc = filepath.Join(a.Repository.Path(), loc)
		rc, err := os.Open(loc)
		if err != nil {
			return fmt.Errorf("failed to open content file: %s", err.Error())
		}
		// copy mode from content file
		rcInfo, _ := rc.Stat()
		filemode = rcInfo.Mode()
		// set reader to read from annexed content file
		reader = bufio.NewReader(rc)
	}
	header.SetMode(filemode)
	writer, _ := a.writer.CreateHeader(&header)
	for n, err := reader.Read(readbuf); n > 0 || err == nil; n, err = reader.Read(readbuf) {
		if _, err := writer.Write(readbuf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func (a *ZipWriter) addTree(tree *git.Tree, parent string) error {
	entries, _ := tree.Entries()
	for _, te := range entries {
		path := filepath.Join(parent, te.Name())
		if te.IsTree() {
			if _, err := a.writer.Create(path + "/"); err != nil {
				return err
			}
			subtree, _ := tree.Subtree(te.Name())
			if err := a.addTree(subtree, path); err != nil {
				return err
			}
		} else {
			blob := te.Blob()
			if err := a.addBlob(blob, path); err != nil {
				return err
			}
		}
	}
	return nil
}
