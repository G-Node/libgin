package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gogs/git-module"
)

// extractTestRepo extracts the zip archive used for testing.
// Returns the git.Repository.
// Uses external (system) unzip command.
func extractTestRepo() (*git.Repository, error) {
	zipfilepath := "../../testdata/testrepo.zip"

	temprepo, err := ioutil.TempDir("", "libgintestrepo")
	if err != nil {
		return nil, err
	}
	if err := exec.Command("unzip", "-d", temprepo, zipfilepath).Run(); err != nil {
		return nil, err
	}

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
