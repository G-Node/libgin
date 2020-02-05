package annex

import (
	"fmt"
	"testing"
)

func Test_NewAFile(t *testing.T) {
	link1 := ".git/annex/objects/QZ/fV/SHA256E-s1000000--d29751f2649b32ff572b5e0a9f541ea660a50f94ff0beedfb0b692b924cc8025.big/SHA256E-s1000000--d29751f2649b32ff572b5e0a9f541ea660a50f94ff0beedfb0b692b924cc8025.big"
	link2 := ".git/annex/objects/Wf/Wp/WORM-s1000000-m1498217233--bigfile.big/WORM-s1000000-m1498217233--bigfile.big"

	af, err := NewAFile("annex", "testdata/fakerepo.git", "bigfile.big", []byte(link1))
	if err != nil {
		t.Logf("%v", err)
		t.Fail()
	}
	if af == nil {
		t.Fatal("Nil Annex file reference without error")
	}
	fmt.Println(af.Filepath)
	fp, err := af.Open()
	if err != nil {
		t.Fatalf("Opening annex file failed: %v", err)
	}
	defer fp.Close()

	af, err = NewAFile("testdata", "testdata/fakerepo.git", "bigfile.big", []byte(link2))
	if err != nil {
		t.Fatalf("annexfile %v, error: %v", af, err)
	}
	fp, err = af.Open()
	if err != nil {
		t.Fatalf("Opening annex file failed: %v", err)
	}
	defer fp.Close()
}
