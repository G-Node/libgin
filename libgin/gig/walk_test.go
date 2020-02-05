package gig

import (
	"testing"
)

func TestWalkRef(t *testing.T) {
	rep, err := OpenRepository("tdata/repo1.git")
	if err != nil {
		t.Errorf("Could not open test repository: %s", err.Error())
	}
	commits, err := rep.WalkRef("tag1", func(commitId SHA1) bool {
		return true
	})
	if err != nil {
		t.Errorf("Could not walk master repository: %s", err.Error())
		return
	}
	t.Logf("Got commits: %v", commits)
	id_sha, _ := ParseSHA1("f8306602c14ab6f49dae674513b6f6a7748e6f09")
	if commits[id_sha] == nil {
		t.Fatalf("Expected non-nil value")
	}
}

func TestGetBlobs(t *testing.T) {
	rep, err := OpenRepository("tdata/repo1.git")
	if err != nil {
		t.Errorf("Could not open test repository: %s", err.Error())
	}
	id_commit, _ := ParseSHA1("f8306602c14ab6f49dae674513b6f6a7748e6f09")
	com, err := rep.OpenObject(id_commit)
	if err != nil {
		t.Errorf("Could not open commit: %s", err.Error())
	}
	blobs := make(map[SHA1]*Blob)
	rep.GetBlobsForCommit(com.(*Commit), blobs)
	t.Logf("Got Blobs: %v", blobs)
	id_blob, _ := ParseSHA1("02e020cdf53288638ab42fd1529556aeccd3e873")
	if blobs[id_blob] == nil {
		t.Fatalf("Expected non-nil value")
	}
}
