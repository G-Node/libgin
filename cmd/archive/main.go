package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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
		writer.Write(readbuf[:n])
	}
}

func addBlobTar(tarWriter *tar.Writer, blob *git.Blob, fname string, repopath string) {
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
		size = rcInfo.Size()
	}
	header.Mode = int64(filemode)
	header.Size = size
	tarWriter.WriteHeader(&header)
	for ; n > 0 || err == nil; n, err = reader.Read(readbuf) {
		tarWriter.Write(readbuf[:n])
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

func tarTree(tarWriter *tar.Writer, t *git.Tree, prefix string, repopath string) {
	entries, _ := t.ListEntries()
	for _, te := range entries {
		fname := filepath.Join(prefix, te.Name())
		if te.IsDir() {
			subtree, _ := t.SubTree(te.Name())
			tarTree(tarWriter, subtree, fname, repopath)
		} else {
			blob := te.Blob()
			addBlobTar(tarWriter, blob, fname, repopath)
		}
	}
}

func ginArchiveZip(path string) error {
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

func ginArchiveTar(path string) error {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return err
	}

	master, err := repo.GetCommit("master")
	if err != nil {
		return err
	}

	fname := master.ID.String()[:6] + ".tar.gz"

	// place archive in repository's parent directory
	archivepath, _ := filepath.Abs(filepath.Join(path, "..", fname))

	tree := &master.Tree

	gzipfile, _ := os.Create(archivepath)
	defer gzipfile.Close()
	gzipWriter := gzip.NewWriter(gzipfile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	tarTree(tarWriter, tree, "", path)

	return nil
}

func gzipTar(srcpath string) error {
	trgtpath := srcpath + ".gz"
	gzipfile, _ := os.Create(trgtpath)
	zipWriter := gzip.NewWriter(gzipfile)
	defer gzipfile.Close()
	defer zipWriter.Close()
	zipWriter.Name = filepath.Base(srcpath)
	zipWriter.ModTime = time.Now()

	srcfile, _ := os.Open(srcpath)
	defer srcfile.Close()
	reader := bufio.NewReader(srcfile)
	n, err := io.Copy(zipWriter, reader)
	if err != nil {
		return err
	}
	fmt.Printf("%d bytes zipped\n", n)
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

	err := ginArchiveZip(path)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	err = ginArchiveTar(path)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
}
