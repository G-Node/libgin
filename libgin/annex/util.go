package annex

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogs/git-module"
)

const annexDirLetters = "0123456789zqjxkmvwgpfZQJXKMVWGPF"

func IsAnnexFile(blob *git.Blob) bool {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	blob.Pipeline(stdout, stderr)
	reader := bufio.NewReader(stdout)

	// Contents differ if the file is locked (symlink) or unlocked (plain text blob)
	// symlinks contain a relative path to the annex content store (.git/annex/objects)
	// unlocked pointers contain a path rooted at /annex/objects
	// It's important not to look through the entire file for the string '/annex/objects'
	// for a couple of reasons:
	// 1. The file might be very large
	// 2. The file might contain the string for other reasons (e.g., documentation, or this file)

	if blob.IsSymlink() {
		// read the entire contents and check if it's a path inside .git/annex/objects
		readbuf := make([]byte, reader.Size())
		_, err := reader.Read(readbuf)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}
		return strings.Contains(string(readbuf), ".git/annex/objects")
	}

	// check if the first bytes match /annex/objects
	const ident = "/annex/objects"
	readbuf := make([]byte, len(ident))
	_, err := reader.Read(readbuf)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
	}
	return string(readbuf) == ident
}

func Upgrade(dir string) ([]byte, error) {
	cmd := git.NewCommand("annex", "upgrade")
	return cmd.RunInDir(dir)
}

func IsBare(repo *git.Repository) (bool, error) {
	outb, err := git.NewCommand("config", "core.bare").RunInDir(repo.Path())
	out := string(bytes.TrimSpace(outb))
	if err != nil {
		return false, err
	}
	if out == "true" {
		return true, nil
	}
	if out == "false" {
		return false, nil
	}
	return false, fmt.Errorf("unexpected output: %s", out)
}

// ContentLocation returns the location of the content file for a given annex key.
// The returned path is relative to the repository git directory.
func ContentLocation(repo *git.Repository, key string) (string, error) {
	gitdir := repo.Path()
	if bare, err := IsBare(repo); err != nil {
		return "", err
	} else if !bare {
		gitdir = filepath.Join(gitdir, ".git")
	}
	objectstore := filepath.Join(gitdir, "annex", "objects")
	// there are two possible object paths depending on annex version
	// the most common one is the newest, but we should try both anyway
	objectpath := filepath.Join(objectstore, hashdirmixed(key), key)
	if _, err := os.Stat(objectpath); err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("unexpected error occurred while trying to stat %q: %s", objectpath, err)
		}
		// try the other one
		objectpath = filepath.Join(objectstore, hashdirlower(key), key)
		if _, err = os.Stat(objectpath); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("unexpected error occurred while trying to stat %q: %s", objectpath, err)
			}
			return "", fmt.Errorf("failed to find content for key %q: %s", key, err.Error())
		}
	}
	return filepath.Rel(gitdir, objectpath)
}

// hashdirlower is the new method for calculating the location of an annexed
// file's contents based on its key
// See https://git-annex.branchable.com/internals/hashing/ for description
func hashdirlower(key string) string {
	hash := md5.Sum([]byte(key))
	hashstr := fmt.Sprintf("%x", hash)
	return filepath.Join(hashstr[:3], hashstr[3:6], key)
}

// hashdirmixed is the old method for calculating the location of an annexed
// file's contents based on its key
// See https://git-annex.branchable.com/internals/hashing/ for description
func hashdirmixed(key string) string {
	hash := md5.Sum([]byte(key))
	var sum uint64

	sum = 0
	// reverse the first 32bit word of the hash
	firstWord := make([]byte, 4)
	for idx, b := range hash[:4] {
		firstWord[3-idx] = b
	}
	for _, b := range firstWord {
		sum <<= 8
		sum += uint64(b)
	}

	rem := sum
	letters := make([]byte, 4)
	idx := 0
	for rem > 0 && idx < 4 {
		// pull out five bits
		chr := rem & 31
		// save it
		letters[idx] = annexDirLetters[chr]
		// shift the remaining
		rem >>= 6
		idx++
	}

	path := filepath.Join(fmt.Sprintf("%s%s", string(letters[1]), string(letters[0])), fmt.Sprintf("%s%s", string(letters[3]), string(letters[2])), key)
	return path
}
