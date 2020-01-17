package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogs/git-module"
)

func unzip(fname string, dest string) error {
	zr, err := zip.OpenReader(fname)
	if err != nil {
		return err
	}
	defer zr.Close()
	os.MkdirAll(dest, 0777)
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			os.Mkdir(filepath.Join(dest, file.Name), 0777)
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zipped file %q for reading: %s", file.Name, err.Error())
		}
		fw, err := os.Create(filepath.Join(dest, file.Name))
		if err != nil {
			return fmt.Errorf("failed to create file %q during extraction: %s", file.Name, err.Error())
		}
		_, err = io.Copy(fw, fr)
		if err != nil {
			return fmt.Errorf("failed to extract file %q: %s", file.Name, err.Error())
		}
		fr.Close()
		fw.Close()
	}
	return nil
}

// extractTestRepo extracts the zip archive used for testing.
// Returns the git.Repository.
// Uses external (system) unzip command.
func extractTestRepo() (*git.Repository, error) {
	zipfilepath := "../../testdata/testrepo.zip"

	temprepo, err := ioutil.TempDir("", "libgintestrepo")
	if err != nil {
		return nil, err
	}
	unzip(zipfilepath, temprepo)

	return git.OpenRepository(temprepo)
}

func TestZip(t *testing.T) {
	repo, err := extractTestRepo()
	if err != nil {
		t.Fatalf("failed to extract test repository: %s", err.Error())
	}

	defer os.RemoveAll(repo.Path)

	fmt.Println("repository at", repo.Path)
	branches, _ := repo.GetBranches()
	for _, branch := range branches {
		fmt.Println(branch)
	}

	master, err := repo.GetCommit("master")
	if err != nil {
		t.Fatalf("failed to get master branch: %s", err.Error())
	}

	zippath, err := ioutil.TempDir("", "libgintestzip")
	if err != nil {
		t.Fatalf("failed creating directory for zip file: %s", err.Error())
	}

	// defer os.RemoveAll(zippath)

	zipfile := filepath.Join(zippath, "repo.zip")
	writer := NewZipWriter(repo, master)
	if err := writer.Write(zipfile); err != nil {
		t.Fatalf("error creating zip file: %s", err.Error())
	}

	fmt.Println("Zip file created")
	fmt.Println(zippath)
}

func TestTar(t *testing.T) {

}
