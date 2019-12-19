package gig

import (
	"testing"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestWalkRef(t *testing.T) {
	rep, err := OpenRepository("tdata/repo1.git")
	if err != nil {
		t.Errorf("Could not open test repository:%v", err)
	}
	commits, err := rep.WalkRef("tag1", func(commitId SHA1) bool {
		return true
	})
	if err != nil {
		t.Errorf("Could not walk master repository:%v", err)
		return
	}
	t.Log("Got commits:%v", commits)
	id_sha, _ := ParseSHA1("f8306602c14ab6f49dae674513b6f6a7748e6f09")
	assert.NotNil(t, commits[id_sha])

}

func TestGetBlobs(t *testing.T) {
	rep, err := OpenRepository("tdata/repo1.git")
	if err != nil {
		t.Errorf("Could not open test repository:%v", err)
	}
	id_commit, _ := ParseSHA1("f8306602c14ab6f49dae674513b6f6a7748e6f09")
	com, err := rep.OpenObject(id_commit)
	if err != nil {
		t.Errorf("Could not open commit:%v", err)
	}
	blobs := make(map[SHA1]*Blob)
	rep.GetBlobsForCommit(com.(*Commit), blobs)
	t.Log("Got Blobs:%v", blobs)
	id_blob, _ := ParseSHA1("02e020cdf53288638ab42fd1529556aeccd3e873")
	assert.NotNil(t, blobs[id_blob])
}
