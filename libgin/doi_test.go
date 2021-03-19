package libgin

import (
	"fmt"
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

func TestAuthor(t *testing.T) {
	lname := "ln"
	fname := "fn"
	aff := "af"
	id := "id"

	var validate string
	var auth Author
	check := func(label string) {
		rend := auth.RenderAuthor()
		if validate != rend {
			t.Fatalf("%s: Expected '%s' but got '%s'", label, validate, rend)
		}
	}

	// Test RenderAuthor; make all expected combinations explicit
	// No omit test
	validate = fmt.Sprintf("%s, %s; %s; %s", lname, fname, aff, id)

	auth = Author{
		FirstName:   fname,
		LastName:    lname,
		Affiliation: aff,
		ID:          id,
	}
	check("No omit")

	// Omit ID test
	validate = fmt.Sprintf("%s, %s; %s", lname, fname, aff)

	auth = Author{
		FirstName:   fname,
		LastName:    lname,
		Affiliation: aff,
	}
	check("Omit ID")

	// Omit affiliation test
	validate = fmt.Sprintf("%s, %s; %s", lname, fname, id)

	auth = Author{
		FirstName: fname,
		LastName:  lname,
		ID:        id,
	}
	check("Omit affiliation")

	// Omit ID and affiliation test
	validate = fmt.Sprintf("%s, %s", lname, fname)

	auth = Author{
		FirstName: fname,
		LastName:  lname,
	}
	check("Omit ID/affiliation")
}
