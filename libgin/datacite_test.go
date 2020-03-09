package libgin

import (
	"encoding/xml"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func Test_DataCiteMarshal(t *testing.T) {
	example := NewDataCite()
	example.Creators = []Creator{
		Creator{"Achilleas", &NameIdentifier{"0010", "http://orcid.org", "ORCID"}, "University of Bob"},
		Creator{"Bob", &NameIdentifier{"1111", "http://orcid.org", "ORCID"}, "University of University"},
		Creator{"Orphan", nil, ""},
	}

	example.Titles = []string{"This is a sample"}
	example.AddAbstract("This is the abstract")
	example.RightsList = []Rights{Rights{"CC-BY", "http://creativecommons.org/licenses/by/4.0/"}}
	example.Subjects = []string{"One", "Two", "Three"}
	example.AddFunding("DFG, DFG.12345")
	example.AddFunding("EU, EU.12345")
	example.SetResourceType("Dataset")

	example.AddReference(&Reference{ID: "doi:10.1111/example.doi", RefType: "IsDescribedBy", Name: "Manuscript title for reference."})
	example.AddReference(&Reference{ID: "arxiv:10.2222/example.doi", RefType: "IsSupplementTo", Name: "Some other work"})
	example.AddReference(&Reference{ID: "doi:10.3333/example.doi", RefType: "IsReferencedBy", Name: "A work that references this dataset."})

	_, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	// fmt.Println(xml.Header + string(dataciteXML))
}

func Test_DataCiteFromRegInfo(t *testing.T) {
	authors := []Author{
		Author{
			FirstName:   "GivenName1",
			LastName:    "FamilyName1",
			Affiliation: "Affiliation1",
			ID:          "ORCID:0000-0001-2345-6789",
		},
		Author{
			FirstName:   "GivenName2",
			LastName:    "FamilyName2",
			Affiliation: "Affiliation2",
			ID:          "ResearcherID:X-1234-5678",
		},
		Author{
			FirstName: "GivenName3",
			LastName:  "FamilyName3",
		},
	}
	references := []Reference{
		Reference{
			ID:      "doi:10.xxx/zzzz",
			RefType: "IsSupplementTo",
			Name:    "PublicationName1",
		},
		Reference{
			ID:      "arxiv:mmmm.nnnn",
			RefType: "IsDescribedBy",
			Name:    "PublicationName2",
		},
		Reference{
			ID:      "pmid:nnnnnnnn",
			RefType: "IsReferencedBy",
			Name:    "PublicationName3",
		},
	}
	regInfo := &RepositoryYAML{
		Authors:      authors,
		Title:        "This is a sample",
		Description:  "This is the abstract",
		Keywords:     []string{"Neuroscience", "Electrophysiology"},
		License:      &License{"Creative Commons CC0 1.0 Public Domain Dedication", "https://creativecommons.org/publicdomain/zero/1.0/"},
		Funding:      []string{"DFG, DFG.12345", "EU, EU.12345"},
		References:   references,
		ResourceType: "Dataset",
	}
	example := NewDataCiteFromYAML(regInfo)
	dataciteXML, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	fmt.Println(xml.Header + string(dataciteXML))
}

func Test_parseAuthorID(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	randNumStr := func(ndigits int) string {
		maxValue := int64(math.Pow10(ndigits))
		fmtstr := fmt.Sprintf("%%0%dd", ndigits)
		return fmt.Sprintf(fmtstr, rand.Int63n(maxValue))
	}

	if ident := parseAuthorID(""); ident != nil {
		t.Fatal("Empty author ID should return nil")
	}

	validORCIDs := []string{
		// valid, all 0s (different delimiters)
		"orcid.0000-0000-0000-0000",
		"orcid|0000-0000-0000-0000",
		"orcid/0000-0000-0000-0000",
		"orcid:0000-0000-0000-0000",

		// valid, all 0s, X checksum
		"orcid.0000-0000-0000-000X",
		"orcid|0000-0000-0000-000X",
		"orcid/0000-0000-0000-000X",
		"orcid:0000-0000-0000-000X",

		// Valid random
		fmt.Sprintf("orcid.%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid|%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid:%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),

		// Valid random with X checksum
		fmt.Sprintf("orcid.%s-%s-%s-%sX", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(3)),
		fmt.Sprintf("orcid|%s-%s-%s-%sX", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(3)),
		fmt.Sprintf("orcid/%s-%s-%s-%sX", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(3)),
		fmt.Sprintf("orcid:%s-%s-%s-%sX", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(3)),

		// Valid random as URL
		fmt.Sprintf("orcid.org/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid.org/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid.org/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("orcid.org/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),

		// Valid random uppercase
		fmt.Sprintf("ORCID.%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ORCID|%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ORCID/%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ORCID:%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
	}

	for _, id := range validORCIDs {
		ident := parseAuthorID(id)
		if ident == nil {
			t.Fatalf("Author ID parse failed for %q", id)
		}
		if ident.Scheme != "ORCID" {
			t.Fatalf("Author ID parse failed for %q: Invalid Scheme detected %q", id, ident.Scheme)
		}
	}

	invalidORCIDs := []string{
		// Invalid random
		fmt.Sprintf("orcid.%s-%s-%s-%sX", randNumStr(2), randNumStr(2), randNumStr(2), randNumStr(2)),
		fmt.Sprintf("orcid|%s-%s-%s-%sX", randNumStr(4), randNumStr(4), randNumStr(3), randNumStr(4)),
		fmt.Sprintf("orcid/%s-%s-%s-%sX", randNumStr(3), randNumStr(4), randNumStr(4), randNumStr(3)),
		fmt.Sprintf("orcid:%s-%s-%s-%s", randNumStr(3), randNumStr(5), randNumStr(4), randNumStr(4)),
	}

	for _, id := range invalidORCIDs {
		ident := parseAuthorID(id)
		if ident == nil {
			t.Fatalf("Author ID parse failed for %q", id)
		}
		if ident.Scheme != "" {
			t.Fatalf("Invalid author ORCID produced valid response for %q: Detected scheme %q (should be empty)", id, ident.Scheme)
		}
		if ident.ID != id {
			t.Fatalf("Invalid author ORCID produced invalid ID: %q", id)
		}
	}

	validRIDs := []string{
		// valid, all 0s (different delimiters)
		"ResearcherID.A-0000-0000",
		"ResearcherID|B-0000-0000",
		"ResearcherID/C-0000-0000",
		"ResearcherID:D-0000-0000",
		"ResearcherID.E-0000-0000",

		// Valid random
		fmt.Sprintf("ResearcherID.F-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ResearcherID|G-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ResearcherID/H-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("ResearcherID:I-%s-%s", randNumStr(4), randNumStr(4)),

		// Valid random (all lowercase)
		fmt.Sprintf("researcherid.J-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("researcherid|K-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("researcherid/L-%s-%s", randNumStr(4), randNumStr(4)),
		fmt.Sprintf("researcherid:M-%s-%s", randNumStr(4), randNumStr(4)),
	}

	for _, id := range validRIDs {
		ident := parseAuthorID(id)
		if ident == nil {
			t.Fatalf("Author ID parse failed for %q", id)
		}
		if ident.Scheme != "ResercherID" {
			t.Fatalf("Author ID parse failed for %q: Invalid Scheme detected %q", id, ident.Scheme)
		}
	}

	invalidRIDs := []string{
		// Invalid random
		fmt.Sprintf("ResearcherID.A-%s-%s-%s-%sX", randNumStr(2), randNumStr(2), randNumStr(2), randNumStr(2)),
		fmt.Sprintf("ResearcherID|B-%s-%s", randNumStr(4), randNumStr(3)),
		fmt.Sprintf("ResearcherID/F-%s-%s", randNumStr(3), randNumStr(4)),
		fmt.Sprintf("ResearcherID:%s-%s", randNumStr(4), randNumStr(4)),
	}

	for _, id := range invalidRIDs {
		ident := parseAuthorID(id)
		if ident == nil {
			t.Fatalf("Author ID parsing failed for %q", id)
		}
		if ident.Scheme != "" {
			t.Fatalf("Invalid author ResearcherID produced valid response for %q: Detected scheme %q (should be empty)", id, ident.Scheme)
		}
		if ident.ID != id {
			t.Fatalf("Invalid author ResearcherID produced invalid ID: %q", id)
		}
	}

	other := []string{
		"anything else",
		// looks like ResearcherID without prefix
		fmt.Sprintf("A-%s-%s", randNumStr(4), randNumStr(4)),
		// looks like ORCID without prefix
		fmt.Sprintf("%s-%s-%s-%s", randNumStr(4), randNumStr(4), randNumStr(4), randNumStr(4)),
	}

	for _, id := range other {
		ident := parseAuthorID(id)
		if ident == nil {
			t.Fatalf("Author ID parsing failed for %q", id)
		}
		if ident.Scheme != "" {
			t.Fatalf("Unknown author ID produced valid response for %q: Detected scheme %q (should be empty)", id, ident.Scheme)
		}
		if ident.ID != id {
			t.Fatalf("Invalid author ID produced invalid ID: %q", id)
		}
	}

}

func Test_GetArchiveSize(t *testing.T) {
	// URL is earliest archive with the new name format, so wont change.
	// Older archives might be renamed to the new format soon.
	const archiveURL = "https://doi.gin.g-node.org/10.12751/g-node.4bdb22/10.12751_g-node.4bdb22.zip"
	const expSize = 1559190240
	size, err := GetArchiveSize(archiveURL)
	if err != nil {
		t.Fatalf("Failed to retrieve archive size for %q: %v", archiveURL, err)
	}

	if size != expSize {
		t.Fatalf("Incorrect archive size: %d (expected) != %d", expSize, size)
	}
}
