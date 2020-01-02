package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		if strings.Contains(string(data[:32]), annexident) { // don't check whole contents for annex string
			_, key := filepath.Split(string(data))
			// git content of unlocked pointer files have newline at the end so
			// we should trim space
			key = strings.TrimSpace(key) // trim newlines and spaces
			annexContentMap[file.Name] = key
		}
	}
	return annexContentMap, nil
}

func replaceStubs(archivepath string, repopath string, annexContent map[string]string) error {
	fmt.Println("==== Replacing pointer files with content ====")
	fmt.Println(archivepath)

	archiveloc, archivename := filepath.Split(archivepath)
	tmpfile, err := ioutil.TempFile(archiveloc, fmt.Sprintf(".*_%s", archivename))
	if err != nil {
		fmt.Printf("ERROR: failed to create temporary archive: %s", err.Error())
		return err
	}

	fmt.Printf("Created temporary file %q\n", tmpfile.Name())
	defer tmpfile.Close()

	zipWriter := zip.NewWriter(tmpfile)
	zipReader, err := zip.OpenReader(archivepath)
	if err != nil {
		return err
	}

	zipcp := func(file *zip.File) error {
		fmt.Printf("Copying %s\n", file.Name)
		buf := make([]byte, 10240)
		rc, err := file.Open()
		if err != nil {
			fmt.Printf("Failed to read: %s\n", err.Error())
			return err
		}
		reader := bufio.NewReader(rc)
		fh, err := zip.FileInfoHeader(file.FileInfo())
		if err != nil {
			fmt.Printf("ERROR: failed to create file header: %s\n", err.Error())
			return err
		}
		// set name to zip-relative path
		fh.Name = file.Name
		writer, err := zipWriter.CreateHeader(fh)
		if err != nil {
			fmt.Printf("ERROR: failed to write header: %s\n", err.Error())
			return err
		}
		for rn, err := reader.Read(buf); rn > 0 && err == nil; rn, err = reader.Read(buf) {
			fmt.Printf("Read %d bytes\n", rn)
			wn, werr := writer.Write(buf[:rn])
			if werr != nil {
				fmt.Printf("ERROR: failed to write to new zip file: %s\n", werr.Error())
				return err
			}
			fmt.Printf("Wrote %d bytes\n", wn)
		}
		return nil
	}

	zipreplace := func(file *zip.File, annexkey string) error {
		loc, err := git.NewCommand("annex", "contentlocation", annexkey).RunInDir(repopath)
		if err != nil {
			fmt.Printf("ERROR: couldn't find content file %q\n", annexkey)
			return err
		}
		loc = strings.TrimRight(loc, "\n")
		loc = filepath.Join(repopath, loc)

		fmt.Printf("Replacing %q with %q\n", file.Name, loc)
		buf := make([]byte, 10240)
		rc, err := os.Open(loc)
		if err != nil {
			fmt.Printf("Failed to read: %s\n", err.Error())
			return err
		}
		reader := bufio.NewReader(rc)
		fh, err := zip.FileInfoHeader(file.FileInfo())
		if err != nil {
			fmt.Printf("ERROR: failed to create file header: %s\n", err.Error())
			return err
		}
		// set name to zip-relative path
		fh.Name = file.Name
		// copy mode from content file
		fmt.Printf("Changing mode from %d ", fh.Mode())
		rcInfo, _ := rc.Stat()
		fh.SetMode(rcInfo.Mode())
		fmt.Printf("to %d\n", fh.Mode())
		writer, err := zipWriter.CreateHeader(fh)
		if err != nil {
			fmt.Printf("ERROR: failed to write header: %s\n", err.Error())
			return err
		}
		for rn, err := reader.Read(buf); rn > 0 && err == nil; rn, err = reader.Read(buf) {
			fmt.Printf("Read %d bytes\n", rn)
			wn, werr := writer.Write(buf[:rn])
			if werr != nil {
				fmt.Printf("ERROR: failed to write to new zip file: %s\n", werr.Error())
				return err
			}
			fmt.Printf("Wrote %d bytes\n", wn)
		}
		return nil
	}

	for _, file := range zipReader.File {
		annexkey, ok := annexContent[file.Name]
		if ok {
			zipreplace(file, annexkey)
		} else {
			zipcp(file)
		}
	}

	// TODO: check errors
	zipReader.Close()
	zipWriter.Close()

	// Overwrite original zip with temp
	os.Rename(tmpfile.Name(), archivepath)

	return nil
}

func ginarchiveTwoStep(path string) error {
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
	// place archive in repository's parent directory
	archivepath, _ := filepath.Abs(filepath.Join(path, "..", fname))
	fmt.Printf("Archiving repository at %s to %s\n", path, archivepath)
	if err := master.CreateArchive(archivepath, git.ZIP); err != nil {
		return err
	}

	// 2. Identify annexed files
	annexContent, err := readAnnexedStubs(archivepath)
	if err != nil {
		return err
	}

	// 3. Update git archive with annexed content
	err = replaceStubs(archivepath, path, annexContent)
	if err != nil {
		fmt.Printf("ERROR: failed to update zip archive: %s", err.Error())
	}

	return err
}

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
		fmt.Printf("Replacing %q with %q\n", fname, loc)
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
		fmt.Println(fname)
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

	// 1. Create git archive
	fname := master.ID.String()[:6] + ".zip"
	// place archive in repository's parent directory
	archivepath, _ := filepath.Abs(filepath.Join(path, "..", fname))
	fmt.Printf("Archiving repository at %s to %s\n", path, archivepath)

	// walk tree and zip files
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
