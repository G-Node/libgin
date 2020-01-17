package annex

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/gogs/git-module"
)

func IsAnnexFile(blob *git.Blob) bool {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	blob.DataPipeline(stdout, stderr)
	reader := bufio.NewReader(stdout)

	// Contents differ if the file is locked (symlink) or unlocked (plain text blob)
	// symlinks contain a relative path to the annex content store (.git/annex/objects)
	// unlocked pointers contain a path rooted at /annex/objects
	// It's important not to look through the entire file for the string '/annex/objects'
	// for a couple of reasons:
	// 1. The file might be very large
	// 2. The file might contain the string for other reasons (e.g., documentation, or this file)

	if blob.IsLink() {
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

func Upgrade(dir string) (string, error) {
	cmd := git.NewACommand("upgrade")
	return cmd.RunInDir(dir)
}
