package archive

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/G-Node/libgin/libgin/annex"
	"github.com/gogs/git-module"
)

type TarWriter struct {
	Repository *git.Repository
	Commit     *git.Commit
	writer     *tar.Writer
}

func NewTarWriter(repo *git.Repository, commit *git.Commit) *TarWriter {
	return &TarWriter{Repository: repo, Commit: commit}
}

func (a *TarWriter) Write(target string) error {
	tree := &a.Commit.Tree

	gzipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer gzipfile.Close()
	gzipWriter := gzip.NewWriter(gzipfile)
	defer gzipWriter.Close()

	a.writer = tar.NewWriter(gzipWriter)
	defer a.writer.Close()

	a.addTree(tree, "")

	return nil
}

func (a *TarWriter) addBlob(blob *git.Blob, fname string) error {
	var filemode os.FileMode
	filemode |= 0660
	if blob.IsLink() {
		filemode |= os.ModeSymlink
	}
	header := tar.Header{
		Name:    filepath.Join(fname),
		ModTime: time.Now(), // TODO: use commit time
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	blob.DataPipeline(stdout, stderr)
	size := blob.Size()
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
			return fmt.Errorf("content file not found: %q\n", annexkey)
		}
		loc = strings.TrimRight(loc, "\n")
		loc = filepath.Join(a.Repository.Path, loc)
		rc, err := os.Open(loc)
		if err != nil {
			return fmt.Errorf("failed to open content file: %s\n", err.Error())
		}
		// copy mode from content file
		rcInfo, _ := rc.Stat()
		filemode = rcInfo.Mode()
		n, err = reader.Read(readbuf)
		reader = bufio.NewReader(rc)
		size = rcInfo.Size()
	}
	header.Mode = int64(filemode)
	header.Size = size
	a.writer.WriteHeader(&header)
	for ; n > 0 || err == nil; n, err = reader.Read(readbuf) {
		a.writer.Write(readbuf[:n])
	}
	return nil
}

func (a *TarWriter) addTree(tree *git.Tree, path string) error {
	entries, _ := tree.ListEntries()
	for _, te := range entries {
		path := filepath.Join(path, te.Name())
		if te.IsDir() {
			subtree, _ := tree.SubTree(te.Name())
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
