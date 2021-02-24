package libgin

import (
	"testing"
)

func TestRepoPathToUUID(t *testing.T) {
	mkey := "fabee/efish_locking"
	mval := "6953bbf0087ba444b2d549b759de4a06"

	retval := RepoPathToUUID(mkey)
	if retval != mval {
		t.Fatalf("Expected %s but got %s", mval, retval)
	}

	mkey = "ppfeiffer/cooperative_channels_in_biological_neurons_via_dynamic_clamp"
	mval = "08853efd775dc85d9636bae2043c8f2c"
	retval = RepoPathToUUID(mkey)
	if retval != mval {
		t.Fatalf("Expected %s but got %s", mval, retval)
	}
}
