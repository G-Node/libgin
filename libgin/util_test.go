package libgin

import (
	"os"
	"testing"
)

func Test_ReadConfDefault(t *testing.T) {
	oskey := "test"
	val := "tmp"
	defval := "default"

	testval := ReadConfDefault(oskey, defval)
	if testval != defval {
		t.Fatalf("Expected default value %q but got %q", defval, testval)
	}

	os.Setenv(oskey, val)
	testval = ReadConfDefault(oskey, defval)
	if testval != val {
		t.Fatalf("Expected default value %q but got %q", val, testval)
	}
}

func Test_ReadConf(t *testing.T) {
	oskey := "test"
	val := "tmp"
	os.Setenv(oskey, val)

	testval := ReadConf(oskey)
	if val != testval {
		t.Fatal("Could not read environment variable")
	}
}

func Test_GetArchiveSize(t *testing.T) {
	// URL is earliest archive with the new name format, so wont change.
	// Older archives might be renamed to the new format soon.
	const archiveURL = "https://doi.gin.g-node.org/10.12751/g-node.4bdb22/10.12751_g-node.4bdb22.zip"
	const expSize = 1559190240
	size, err := GetArchiveSize(archiveURL)
	if err != nil {
		t.Fatalf("Failed to retrieve archive size for %q: %v", archiveURL, err.Error())
	}

	if size != expSize {
		t.Fatalf("Incorrect archive size: %d (expected) != %d", expSize, size)
	}

	// Check status not ok
	_, err = GetArchiveSize("https://doi.gin.g-node.org/idonotexist")
	if err == nil {
		t.Fatalf("Expected error on invalid URL")
	}

	// Check fail
	_, err = GetArchiveSize("I do not exist")
	if err == nil {
		t.Fatal("Expected error on non URL")
	}
}
